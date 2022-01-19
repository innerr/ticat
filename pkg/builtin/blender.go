package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

func SetForestMode(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	if !cc.ForestMode.AtForestTopLvl(env) {
		cc.ForestMode.Push(core.GetLastStackFrame(env))
	}
	return currCmdIdx, true
}

func BlenderReplace(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	src := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "src")
	dest := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "dest")

	srcCmd, _ := cc.ParseCmd(true, src)
	destCmd, _ := cc.ParseCmd(true, core.FlowStrToStrs(dest)...)
	cc.Blender.AddReplace(srcCmd, destCmd, -1)
	return currCmdIdx, true
}
