package core

import (
	"fmt"
	"strings"
)

type BlenderInvoker interface {
	Invoke(cmd ParsedCmd) (cmds []ParsedCmd, originDeleted bool, originPosDelta int)
}

type BlenderInvokerReplace struct {
	cnt  int
	src  ParsedCmd
	dest ParsedCmd
}

func (self BlenderInvokerReplace) Invoke(cmd ParsedCmd) (cmds []ParsedCmd, originDeleted bool, originPosDelta int) {
	if self.src.LastCmdNode() != cmd.LastCmdNode() {
		return []ParsedCmd{cmd}, false, 0
	}
	return []ParsedCmd{self.dest}, true, -1
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

func (self *Blender) Invoke(cc *Cli, env *Env, flow *ParsedCmds) (changed bool) {
	result := []ParsedCmd{}

	for i := 0; i < len(flow.Cmds); i++ {
		parsedCmd := flow.Cmds[i]
		if parsedCmd.ParseResult.Error != nil {
			panic(parsedCmd.ParseResult.Error)
		}
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
		for _, invoker := range self.invokers {
			cmds, deleted, posDelta := invoker.Invoke(parsedCmd)
			if len(cmds) > 0 {
				result = append(result, cmds...)
			}
			if deleted {
				originDeleted = true
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
