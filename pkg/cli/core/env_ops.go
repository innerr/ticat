package core

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
		if (op|EnvOpTypeWrite) > 0 || (op|EnvOpTypeMayWrite) > 0 {
			return true
		}
	}
	return false
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

func (self EnvOps) EnvKeys() []string {
	return self.orderedNames
}

func (self EnvOps) Ops(name string) []uint {
	val, _ := self.ops[name]
	return val
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

func (self EnvOpsChecker) OnCallCmd(
	env *Env,
	matched ParsedCmd,
	pathSep string,
	cmd *Cmd,
	ignoreMaybe bool,
	displayPath string,
	arg2envs FirstArg2EnvProviders) (result []EnvOpsCheckResult) {

	arg2envs.Add(matched)

	ops := cmd.EnvOps()
	for _, key := range ops.EnvKeys() {
		for _, curr := range ops.Ops(key) {
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
	result *[]EnvOpsCheckResult) {

	arg2envs := FirstArg2EnvProviders{}
	checkEnvOps(cc, flow, env, checker, ignoreMaybe, result, arg2envs)
}

func checkEnvOps(
	cc *Cli,
	flow *ParsedCmds,
	env *Env,
	checker *EnvOpsChecker,
	ignoreMaybe bool,
	result *[]EnvOpsCheckResult,
	arg2envs FirstArg2EnvProviders) {

	if len(flow.Cmds) == 0 {
		return
	}

	sep := cc.Cmds.Strs.PathSep

	for _, cmd := range flow.Cmds {
		last := cmd.LastCmd()
		if last == nil {
			continue
		}
		displayPath := cmd.DisplayPath(sep, true)
		cmdEnv, argv := cmd.ApplyMappingGenEnvAndArgv(env, cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep)
		res := checker.OnCallCmd(cmdEnv, cmd, sep, last, ignoreMaybe, displayPath, arg2envs)

		*result = append(*result, res...)

		if last.Type() == CmdTypeFlow {
			subFlow, _ := last.Flow(argv, cmdEnv, false)
			parsedFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
			err := parsedFlow.FirstErr()
			if err != nil {
				panic(err.Error)
			}
			if parsedFlow.GlobalEnv != nil {
				env = env.GetOrNewLayer(EnvLayerTmp)
				parsedFlow.GlobalEnv.WriteNotArgTo(env, cc.Cmds.Strs.EnvValDelAllMark)
			}
			checkEnvOps(cc, parsedFlow, env, checker, ignoreMaybe, result, arg2envs)
		}
	}
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
