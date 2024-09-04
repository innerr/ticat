package builtin

import (
	"github.com/pingcap/ticat/pkg/core/model"
)

func SetForestMode(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	if !cc.ForestMode.AtForestTopLvl(env) {
		cc.ForestMode.Push(model.GetLastStackFrame(env))
	}
	return currCmdIdx, true
}

func BlenderReplaceOnce(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	src := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "src")
	dest := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "dest")

	srcCmd, _ := cc.ParseCmd(true, src)
	destCmd, _ := cc.ParseCmd(true, model.FlowStrToStrs(dest)...)
	cc.Blender.AddReplace(srcCmd, destCmd, 1)
	return currCmdIdx, true
}

func BlenderReplaceForAll(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	src := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "src")
	dest := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "dest")

	srcCmd, _ := cc.ParseCmd(true, src)
	destCmd, _ := cc.ParseCmd(true, model.FlowStrToStrs(dest)...)
	cc.Blender.AddReplace(srcCmd, destCmd, -1)
	return currCmdIdx, true
}

func BlenderRemoveOnce(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	targetStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	target, _ := cc.ParseCmd(true, targetStr)
	cc.Blender.AddRemove(target, 1)
	return currCmdIdx, true
}

func BlenderRemoveForAll(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	targetStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	target, _ := cc.ParseCmd(true, targetStr)
	cc.Blender.AddRemove(target, -1)
	return currCmdIdx, true
}

func BlenderInsertOnce(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	targetStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	newStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "new")

	target, _ := cc.ParseCmd(true, targetStr)
	newCmd, _ := cc.ParseCmd(true, model.FlowStrToStrs(newStr)...)
	cc.Blender.AddInsert(target, newCmd, 1)
	return currCmdIdx, true
}

func BlenderInsertForAll(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	targetStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	newStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "new")

	target, _ := cc.ParseCmd(true, targetStr)
	newCmd, _ := cc.ParseCmd(true, model.FlowStrToStrs(newStr)...)
	cc.Blender.AddInsert(target, newCmd, -1)
	return currCmdIdx, true
}

func BlenderInsertAfterOnce(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	targetStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	newStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "new")

	target, _ := cc.ParseCmd(true, targetStr)
	newCmd, _ := cc.ParseCmd(true, model.FlowStrToStrs(newStr)...)
	cc.Blender.AddInsertAfter(target, newCmd, 1)
	return currCmdIdx, true
}

func BlenderInsertAfterForAll(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	targetStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	newStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "new")

	target, _ := cc.ParseCmd(true, targetStr)
	newCmd, _ := cc.ParseCmd(true, model.FlowStrToStrs(newStr)...)
	cc.Blender.AddInsertAfter(target, newCmd, -1)
	return currCmdIdx, true
}

func BlenderClear(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	cc.Blender.Clear()
	return currCmdIdx, true
}
