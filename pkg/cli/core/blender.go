package core

import (
	"fmt"
	"strings"
)

type BlenderInvoker interface {
	Invoke(cmd ParsedCmd) (cmds []ParsedCmd, originDeleted bool, originPosDelta int, finished bool)
}

type BlenderInvokerReplace struct {
	cnt  int
	src  ParsedCmd
	dest ParsedCmd
}

func (self BlenderInvokerReplace) Invoke(cmd ParsedCmd) (cmds []ParsedCmd, originDeleted bool, originPosDelta int, finished bool) {
	if self.src.LastCmdNode() != cmd.LastCmdNode() {
		return []ParsedCmd{cmd}, false, 0, false
	}
	if self.cnt > 0 {
		self.cnt -= 1
	}
	return []ParsedCmd{self.dest}, true, -1, self.cnt == 0
}

type BlenderInvokerRemove struct {
	cnt    int
	target ParsedCmd
}

func (self BlenderInvokerRemove) Invoke(cmd ParsedCmd) (cmds []ParsedCmd, originDeleted bool, originPosDelta int, finished bool) {
	if self.target.LastCmdNode() != cmd.LastCmdNode() {
		return []ParsedCmd{cmd}, false, 0, false
	}
	if self.cnt > 0 {
		self.cnt -= 1
	}
	return nil, true, -1, self.cnt == 0
}

type BlenderInvokerInsert struct {
	cnt    int
	target ParsedCmd
	newCmd ParsedCmd
}

func (self BlenderInvokerInsert) Invoke(cmd ParsedCmd) (cmds []ParsedCmd, originDeleted bool, originPosDelta int, finished bool) {
	if self.target.LastCmdNode() != cmd.LastCmdNode() {
		return []ParsedCmd{cmd}, false, 0, false
	}
	if self.cnt > 0 {
		self.cnt -= 1
	}
	return []ParsedCmd{self.newCmd, cmd}, true, 1, self.cnt == 0
}

type BlenderInvokerInsertAfter struct {
	cnt    int
	target ParsedCmd
	newCmd ParsedCmd
}

func (self BlenderInvokerInsertAfter) Invoke(cmd ParsedCmd) (cmds []ParsedCmd, originDeleted bool, originPosDelta int, finished bool) {
	if self.target.LastCmdNode() != cmd.LastCmdNode() {
		return []ParsedCmd{cmd}, false, 0, false
	}
	if self.cnt > 0 {
		self.cnt -= 1
	}
	return []ParsedCmd{cmd, self.newCmd}, false, 0, self.cnt == 0
}

type Blender struct {
	invokers []BlenderInvoker
}

func NewBlender() *Blender {
	return &Blender{
		[]BlenderInvoker{},
	}
}

func (self *Blender) Clone() *Blender {
	invokers := []BlenderInvoker{}
	for _, info := range self.invokers {
		invokers = append(invokers, info)
	}
	return &Blender{
		invokers,
	}
}

func (self *Blender) AddReplace(src ParsedCmd, dest ParsedCmd, cnt int) {
	invoker := &BlenderInvokerReplace{cnt, src, dest}
	self.invokers = append(self.invokers, invoker)
}

func (self *Blender) AddRemove(target ParsedCmd, cnt int) {
	invoker := &BlenderInvokerRemove{cnt, target}
	self.invokers = append(self.invokers, invoker)
}

func (self *Blender) AddInsert(target ParsedCmd, newCmd ParsedCmd, cnt int) {
	invoker := &BlenderInvokerInsert{cnt, target, newCmd}
	self.invokers = append(self.invokers, invoker)
}

func (self *Blender) AddInsertAfter(target ParsedCmd, newCmd ParsedCmd, cnt int) {
	invoker := &BlenderInvokerInsertAfter{cnt, target, newCmd}
	self.invokers = append(self.invokers, invoker)
}

func (self *Blender) Clear() {
	self.invokers = []BlenderInvoker{}
}

func (self *Blender) Invoke(cc *Cli, env *Env, flow *ParsedCmds) (changed bool) {
	result := []ParsedCmd{}

	for i := 0; i < len(flow.Cmds); i++ {
		parsedCmd := flow.Cmds[i]
		// Parsing error will be checked after blender invoked in executor
		//if parsedCmd.ParseResult.Error != nil {
		//	panic(parsedCmd.ParseResult.Error)
		//}
		cmd := parsedCmd.LastCmdNode()
		if cmd == nil {
			result = append(result, parsedCmd)
			continue
		}

		cic := parsedCmd.LastCmd()
		if cmd != nil && cic != nil && cic.IsBlenderCmd() {
			cmdEnv, argv := parsedCmd.ApplyMappingGenEnvAndArgv(
				env.Clone(), cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep)
			_, ok := cic.executeByType(argv, cc, cmdEnv, nil, flow, i, "")
			if !ok {
				panic(fmt.Errorf("[Blender.Invoke] blender '%s' invoke failed", cmd.DisplayPath()))
			}
			continue
		}

		if len(self.invokers) == 0 {
			result = append(result, parsedCmd)
			continue
		}

		originDeleted := false
		originPosDelta := 0
		addedCmdCnt := 0
		for j := 0; j < len(self.invokers); j++ {
			invoker := self.invokers[j]
			cmds, deleted, posDelta, finished := invoker.Invoke(parsedCmd)
			if finished {
				if j == len(self.invokers)-1 {
					self.invokers = self.invokers[0:j]
				} else {
					self.invokers = append(self.invokers[0:j], self.invokers[j+1:]...)
				}
				j += 1
			}

			addedCmdCnt += len(cmds) - 1
			if len(cmds) > 0 {
				result = append(result, cmds...)
			}
			if deleted {
				originDeleted = true
				break
			} else {
				originPosDelta += posDelta
			}
		}

		if originDeleted || originPosDelta != 0 {
			changed = true
		}
		if i == flow.GlobalCmdIdx {
			if originDeleted {
				flow.GlobalCmdIdx = -1
			} else {
				flow.GlobalCmdIdx += originPosDelta
			}
		} else if i < flow.GlobalCmdIdx {
			flow.GlobalCmdIdx += addedCmdCnt
		}
	}

	flow.Cmds = result
	return
}

// TODO: should not in core package
func GetLastStackFrame(env *Env) string {
	stackStr := env.GetRaw("sys.stack")
	if len(stackStr) == 0 {
		panic(fmt.Errorf("[BlenderForestMode] should never happen"))
	}
	listSep := env.GetRaw("strs.list-sep")
	stack := strings.Split(stackStr, listSep)
	return stack[len(stack)-1]
}
