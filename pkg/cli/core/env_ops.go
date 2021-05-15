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
	ignoreMaybe bool) (result []EnvOpsCheckResult) {

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
	matched ParsedCmd
	cmd     *Cmd
}
