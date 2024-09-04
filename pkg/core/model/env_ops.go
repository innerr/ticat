package model

import (
	"strings"
)

const (
	EnvOpTypeRead     uint = 0x01
	EnvOpTypeWrite    uint = 0x02
	EnvOpTypeMayRead  uint = 0x04
	EnvOpTypeMayWrite uint = 0x08
)

type EnvOps struct {
	orderedNames []string
	ops          map[string][]uint
}

func newEnvOps() EnvOps {
	return EnvOps{nil, map[string][]uint{}}
}

func (self *EnvOps) IsEmpty() bool {
	return len(self.orderedNames) == 0
}

func (self *EnvOps) AddOp(name string, op uint) {
	old, ok := self.ops[name]
	if !ok {
		self.orderedNames = append(self.orderedNames, name)
	}
	self.ops[name] = append(old, op)
}

func (self EnvOps) MatchWriteKey(key string) bool {
	ops, ok := self.ops[key]
	if !ok {
		return false
	}
	for _, op := range ops {
		if (op&EnvOpTypeWrite) > 0 || (op&EnvOpTypeMayWrite) > 0 {
			return true
		}
	}
	return false
}

func (self EnvOps) AllWriteKeys() (keys []string) {
	for key, ops := range self.ops {
		for _, op := range ops {
			if (op&EnvOpTypeWrite) > 0 || (op&EnvOpTypeMayWrite) > 0 {
				keys = append(keys, key)
				break
			}
		}
	}
	return
}

func (self EnvOps) MatchFind(findStr string) bool {
	for _, name := range self.orderedNames {
		if strings.Index(name, findStr) >= 0 {
			return true
		}
		for _, op := range self.Ops(name) {
			if strings.Index(EnvOpStr(op), findStr) >= 0 {
				return true
			}
		}
	}
	return false
}

func (self EnvOps) RawEnvKeys() []string {
	return self.orderedNames
}

func (self EnvOps) Ops(name string) []uint {
	val, _ := self.ops[name]
	return val
}

// TODO: render too many times
func (self EnvOps) RenderedEnvKeys(
	argv ArgVals,
	env *Env,
	cmd *Cmd,
	allowError bool) (renderedKeys []string, origins []string, fullyRendered bool) {

	fullyRendered = true
	for _, name := range self.orderedNames {
		keys, keyFullyRendered := renderTemplateStr(name, "key ops", cmd, argv, env, allowError)
		for _, key := range keys {
			renderedKeys = append(renderedKeys, key)
			origins = append(origins, name)
		}
		fullyRendered = fullyRendered && keyFullyRendered
	}
	return
}

type EnvOpsChecker map[string]envOpsCheckerKeyInfo

type EnvOpsCheckResult struct {
	Key                string
	Cmd                *CmdTree
	CmdDisplayPath     string
	MayWriteCmdsBefore []MayWriteCmd
	ReadMayWrite       bool
	MayReadMayWrite    bool
	MayReadNotExist    bool
	ReadNotExist       bool
	FirstArg2Env       *ParsedCmd
}

type FirstArg2EnvProviders map[string]*ParsedCmd

func (self FirstArg2EnvProviders) Add(matched ParsedCmd) {
	cic := matched.LastCmd()
	arg2env := cic.GetArg2Env()
	keys := arg2env.EnvKeys()
	for _, key := range keys {
		_, ok := self[key]
		if ok {
			continue
		}
		self[key] = &matched
	}
}

func (self FirstArg2EnvProviders) Get(key string) (matched *ParsedCmd) {
	matched, _ = self[key]
	return
}

type EnvOpCmd struct {
	Func   interface{}
	Action func(*EnvOpsChecker, ArgVals, *Env)
}

func (self *EnvOpsChecker) Reset() {
	*self = EnvOpsChecker{}
}

func (self *EnvOpsChecker) RemoveKeyStat(key string) {
	(*self)[key] = envOpsCheckerKeyInfo{}
}

func (self *EnvOpsChecker) SetKeyWritten(key string) {
	old, ok := (*self)[key]
	if !ok {
		old = envOpsCheckerKeyInfo{}
	}
	old.val = old.val | EnvOpTypeWrite
	(*self)[key] = old
}

func (self EnvOpsChecker) OnCallCmd(
	env *Env,
	argv ArgVals,
	matched ParsedCmd,
	pathSep string,
	cmd *Cmd,
	ignoreMaybe bool,
	displayPath string,
	arg2envs FirstArg2EnvProviders) (result []EnvOpsCheckResult) {

	arg2envs.Add(matched)

	ops := cmd.EnvOps()
	keys, origins, _ := ops.RenderedEnvKeys(argv, env, cmd, false)
	for i, key := range keys {
		for _, curr := range ops.Ops(origins[i]) {
			before, _ := self[key]

			if (curr&EnvOpTypeWrite) == 0 && (curr&EnvOpTypeMayWrite) != 0 {
				before.mayWriteCmds = append(before.mayWriteCmds, MayWriteCmd{matched, cmd})
			}
			before.val = before.val | curr
			self[key] = before

			var res EnvOpsCheckResult
			res.Key = key
			res.CmdDisplayPath = displayPath
			res.Cmd = cmd.Owner()
			if (before.val&EnvOpTypeWrite) == 0 &&
				(before.val&EnvOpTypeMayWrite) == 0 {
				if (before.val & EnvOpTypeRead) != 0 {
					res.ReadNotExist = true
				} else if (before.val & EnvOpTypeMayRead) != 0 {
					res.MayReadNotExist = true
				}
			} else if (before.val & EnvOpTypeMayWrite) != 0 {
				if (before.val & EnvOpTypeRead) != 0 {
					res.ReadMayWrite = true
					res.MayWriteCmdsBefore = before.mayWriteCmds
				} else if (before.val & EnvOpTypeMayRead) != 0 {
					res.MayReadMayWrite = true
					res.MayWriteCmdsBefore = before.mayWriteCmds
				}
			}
			var passCheck bool
			if ignoreMaybe {
				passCheck = !res.ReadNotExist
			} else {
				passCheck = !(res.ReadMayWrite || res.MayReadMayWrite ||
					res.MayReadNotExist || res.ReadNotExist)
			}
			if !passCheck && len(env.GetRaw(res.Key)) == 0 {
				res.FirstArg2Env = arg2envs.Get(res.Key)
				result = append(result, res)
			}
		}
	}
	return
}

type envOpsCheckerKeyInfo struct {
	mayWriteCmds []MayWriteCmd
	val          uint
}

type MayWriteCmd struct {
	Matched ParsedCmd
	Cmd     *Cmd
}

func CheckEnvOps(
	cc *Cli,
	flow *ParsedCmds,
	env *Env,
	checker *EnvOpsChecker,
	ignoreMaybe bool,
	envOpCmds []EnvOpCmd,
	result *[]EnvOpsCheckResult) {

	arg2envs := FirstArg2EnvProviders{}
	checkEnvOps(cc, flow, env.Clone(), checker, ignoreMaybe, envOpCmds, result, arg2envs, 0)
}

func checkEnvOps(
	cc *Cli,
	flow *ParsedCmds,
	env *Env,
	checker *EnvOpsChecker,
	ignoreMaybe bool,
	envOpCmds []EnvOpCmd,
	result *[]EnvOpsCheckResult,
	arg2envs FirstArg2EnvProviders,
	depth int) {

	if len(flow.Cmds) == 0 {
		return
	}

	sep := cc.Cmds.Strs.PathSep

	for i, cmd := range flow.Cmds {
		last := cmd.LastCmd()
		if last == nil {
			continue
		}

		displayPath := cmd.DisplayPath(sep, true)

		cmdEnv, argv := cmd.ApplyMappingGenEnvAndArgv(env, cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep, depth+1)

		res := checker.OnCallCmd(cmdEnv, argv, cmd, sep, last, ignoreMaybe, displayPath, arg2envs)
		*result = append(*result, res...)

		TryExeEnvOpCmds(argv, cc, cmdEnv, flow, i, envOpCmds, checker,
			"failed to execute env-op cmd in env-ops checking")

		if !last.HasSubFlow(true) {
			continue
		}

		parsedFlow, flowEnv := renderSubFlowOnChecking(last, cc, argv, cmdEnv)
		checkEnvOps(cc, parsedFlow, flowEnv, checker, ignoreMaybe, envOpCmds, result, arg2envs, depth+1)
	}
}

func TryExeEnvOpCmds(
	argv ArgVals,
	cc *Cli,
	env *Env,
	flow *ParsedCmds,
	currCmdIdx int,
	envOpCmds []EnvOpCmd,
	checker *EnvOpsChecker,
	errString string) {

	cmd := flow.Cmds[currCmdIdx].LastCmd()
	for _, it := range envOpCmds {
		if !cmd.IsTheSameFunc(it.Func) {
			continue
		}
		it.Action(checker, argv, env)
		break
	}
}

// TODO: a bit messy
func renderSubFlowOnChecking(last *Cmd, cc *Cli, argv ArgVals, cmdEnv *Env) (parsedFlow *ParsedCmds, flowEnv *Env) {
	subFlow, _, _ := last.Flow(argv, cc, cmdEnv, false, true)
	parsedFlow = cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
	err := parsedFlow.FirstErr()
	if err != nil {
		panic(err.Error)
	}
	flowEnv = cmdEnv.NewLayer(EnvLayerSubFlow)
	if parsedFlow.GlobalEnv != nil {
		parsedFlow.GlobalEnv.WriteNotArgTo(flowEnv, cc.Cmds.Strs.EnvValDelAllMark)
	}
	return
}

func EnvOpStr(op uint) (str string) {
	switch op {
	case EnvOpTypeWrite:
		str = "write"
	case EnvOpTypeMayWrite:
		str = "may-write"
	case EnvOpTypeRead:
		str = "read"
	case EnvOpTypeMayRead:
		str = "may-read"
	default:
	}
	return
}
