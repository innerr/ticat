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

func BlenderReplaceOnce(
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
	cc.Blender.AddReplace(srcCmd, destCmd, 1)
	return currCmdIdx, true
}

func BlenderReplaceForAll(
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

func BlenderRemoveOnce(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	targetStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	target, _ := cc.ParseCmd(true, targetStr)
	cc.Blender.AddRemove(target, 1)
	return currCmdIdx, true
}

func BlenderRemoveForAll(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	targetStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	target, _ := cc.ParseCmd(true, targetStr)
	cc.Blender.AddRemove(target, -1)
	return currCmdIdx, true
}

func BlenderInsertOnce(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	targetStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	newStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "new")

	target, _ := cc.ParseCmd(true, targetStr)
	newCmd, _ := cc.ParseCmd(true, core.FlowStrToStrs(newStr)...)
	cc.Blender.AddInsert(target, newCmd, 1)
	return currCmdIdx, true
}

func BlenderInsertForAll(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	targetStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	newStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "new")

	target, _ := cc.ParseCmd(true, targetStr)
	newCmd, _ := cc.ParseCmd(true, core.FlowStrToStrs(newStr)...)
	cc.Blender.AddInsert(target, newCmd, -1)
	return currCmdIdx, true
}

func BlenderInsertAfterOnce(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	targetStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	newStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "new")

	target, _ := cc.ParseCmd(true, targetStr)
	newCmd, _ := cc.ParseCmd(true, core.FlowStrToStrs(newStr)...)
	cc.Blender.AddInsertAfter(target, newCmd, 1)
	return currCmdIdx, true
}

func BlenderInsertAfterForAll(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	targetStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	newStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "new")

	target, _ := cc.ParseCmd(true, targetStr)
	newCmd, _ := cc.ParseCmd(true, core.FlowStrToStrs(newStr)...)
	cc.Blender.AddInsertAfter(target, newCmd, -1)
	return currCmdIdx, true
}

func BlenderClear(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	cc.Blender.Clear()
	return currCmdIdx, true
}
