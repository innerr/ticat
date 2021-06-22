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
	Cmd                *CmdTree
	CmdDisplayPath     string
	Key                string
	MayWriteCmdsBefore []MayWriteCmd
	ReadMayWrite       bool
	MayReadMayWrite    bool
	MayReadNotExist    bool
	ReadNotExist       bool
}

func (self EnvOpsChecker) OnCallCmd(
	env *Env,
	matched ParsedCmd,
	pathSep string,
	cmd *Cmd,
	ignoreMaybe bool,
	displayPath string) (result []EnvOpsCheckResult) {

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
			res.CmdDisplayPath = displayPath
			res.Cmd = cmd.Owner()
			res.Key = key
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
		cmdEnv, _ := cmd.GenEnvAndArgv(env, cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep)
		res := checker.OnCallCmd(cmdEnv, cmd, sep, last, ignoreMaybe, displayPath)

		*result = append(*result, res...)

		if last.Type() == CmdTypeFlow {
			subFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, last.Flow(cmdEnv)...)
			if subFlow.GlobalEnv != nil {
				env = env.GetOrNewLayer(EnvLayerTmp)
				subFlow.GlobalEnv.WriteNotArgTo(env, cc.Cmds.Strs.EnvValDelAllMark)
			}
			CheckEnvOps(cc, subFlow, env, checker, ignoreMaybe, result)
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
