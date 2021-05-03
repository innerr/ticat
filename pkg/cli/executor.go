package cli

import (
	"time"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

type Executor struct {
}

func (self *Executor) Execute(cc *core.Cli, bootstrap string, input ...string) bool {
	if !self.execute(cc, true, bootstrap) {
		return false
	}

	extra := cc.GlobalEnv.Get("bootstrap").Raw
	if len(extra) != 0 && !self.execute(cc, true, extra) {
		return false
	}

	cc.GlobalEnv.PlusInt("sys.stack-depth", 1)
	if !self.execute(cc, false, input...) {
		return false
	}
	cc.GlobalEnv.PlusInt("sys.stack-depth", -1)
	return true
}

func (self *Executor) execute(cc *core.Cli, isBootstrap bool, input ...string) bool {
	if len(input) == 0 {
		return true
	}
	useCmdsAbbrs(cc.EnvAbbrs, cc.Cmds)
	flow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, input...)
	filterEmptyCmdsAndReorderByPriority(flow)
	return self.executeFlow(cc, isBootstrap, flow)
}

func (self *Executor) executeFlow(cc *core.Cli, isBootstrap bool, flow *core.ParsedCmds) bool {
	env := cc.GlobalEnv.GetLayer(core.EnvLayerSession)
	if flow.GlobalEnv != nil {
		flow.GlobalEnv.WriteNotArgTo(env, cc.Cmds.Strs.EnvValDelMark, cc.Cmds.Strs.EnvValDelAllMark)
	}
	for i := 0; i < len(flow.Cmds); i++ {
		var newCmds []core.ParsedCmd
		var succeeded bool
		cmd := flow.Cmds[i]
		newCmds, i, succeeded = self.executeCmd(cc, isBootstrap, cmd, env, flow.Cmds, i)
		if !succeeded {
			return false
		}
		flow.Cmds = newCmds
	}
	return true
}

func (self *Executor) executeCmd(
	cc *core.Cli,
	isBootstrap bool,
	cmd core.ParsedCmd,
	env *core.Env,
	flow []core.ParsedCmd,
	currCmdIdx int) (newCmds []core.ParsedCmd, newCurrCmdIdx int, succeeded bool) {

	// The env modifications from input will be popped out after a command is executed
	// (TODO) But if a mod modified the env, the modifications stay in session level
	cmdEnv := cmd.GenEnv(env, cc.Cmds.Strs.EnvValDelMark, cc.Cmds.Strs.EnvValDelAllMark)
	argv := cmdEnv.GetArgv(cmd.Path(), cc.Cmds.Strs.PathSep, cmd.Args())

	stackLines := display.PrintCmdStack(isBootstrap, cc.Screen, cmd, cmdEnv, flow, currCmdIdx, cc.Cmds.Strs)
	display.RenderCmdStack(stackLines, cmdEnv, cc.Screen)

	last := cmd[len(cmd)-1].Cmd.Cmd
	start := time.Now()
	if last != nil {
		newCmds, newCurrCmdIdx, succeeded = last.Execute(argv,
			cc, cmdEnv, flow, currCmdIdx)
	} else {
		newCmds, newCurrCmdIdx, succeeded = flow, currCmdIdx, false
	}
	elapsed := time.Now().Sub(start)

	resultLines := display.PrintCmdResult(isBootstrap, cc.Screen, cmd,
		cmdEnv, succeeded, elapsed, flow, currCmdIdx, cc.Cmds.Strs)
	display.RenderCmdResult(resultLines, cmdEnv, cc.Screen)
	return
}

func useCmdsAbbrs(abbrs *core.EnvAbbrs, cmds *core.CmdTree) {
	if cmds == nil {
		return
	}
	for _, subName := range cmds.SubNames() {
		subAbbrs := cmds.SubAbbrs(subName)
		subEnv := abbrs.GetSub(subName)
		if subEnv == nil {
			subEnv = abbrs.AddSub(subName, subAbbrs...)
		} else {
			abbrs.AddSubAbbrs(subName, subAbbrs...)
		}
		subTree := cmds.GetSub(subName)
		useCmdsAbbrs(subEnv, subTree)
	}
}

// 1. remove the cmds only have cmd-level env definication but have no executable
// 2. move priority cmds to the head
func filterEmptyCmdsAndReorderByPriority(flow *core.ParsedCmds) {
	var unfiltered []core.ParsedCmd
	var priorities []core.ParsedCmd
	for _, cmd := range flow.Cmds {
		if cmd.TotallyEmpty() {
			continue
		}
		if cmd.IsPriority() {
			priorities = append(priorities, cmd)
		} else {
			unfiltered = append(unfiltered, cmd)
		}
	}
	flow.Cmds = append(priorities, unfiltered...)
}
