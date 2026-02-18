package builtin

import (
	"github.com/innerr/ticat/pkg/core/model"
)

func SetForestMode(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	if !cc.ForestMode.AtForestTopLvl(env) {
		cc.ForestMode.Push(model.GetLastStackFrame(env))
	}
	return currCmdIdx, nil
}

func BlenderReplaceOnce(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	src, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "src")
	if err != nil {
		return currCmdIdx, err
	}
	dest, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "dest")
	if err != nil {
		return currCmdIdx, err
	}

	srcCmd, _ := cc.ParseCmd(true, src)
	destCmd, _ := cc.ParseCmd(true, model.FlowStrToStrs(dest)...)
	cc.Blender.AddReplace(srcCmd, destCmd, 1)
	return currCmdIdx, nil
}

func BlenderReplaceForAll(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	src, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "src")
	if err != nil {
		return currCmdIdx, err
	}
	dest, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "dest")
	if err != nil {
		return currCmdIdx, err
	}

	srcCmd, _ := cc.ParseCmd(true, src)
	destCmd, _ := cc.ParseCmd(true, model.FlowStrToStrs(dest)...)
	cc.Blender.AddReplace(srcCmd, destCmd, -1)
	return currCmdIdx, nil
}

func BlenderRemoveOnce(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	targetStr, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	if err != nil {
		return currCmdIdx, err
	}
	target, _ := cc.ParseCmd(true, targetStr)
	cc.Blender.AddRemove(target, 1)
	return currCmdIdx, nil
}

func BlenderRemoveForAll(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	targetStr, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	if err != nil {
		return currCmdIdx, err
	}
	target, _ := cc.ParseCmd(true, targetStr)
	cc.Blender.AddRemove(target, -1)
	return currCmdIdx, nil
}

func BlenderInsertOnce(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	targetStr, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	if err != nil {
		return currCmdIdx, err
	}
	newStr, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "new")
	if err != nil {
		return currCmdIdx, err
	}

	target, _ := cc.ParseCmd(true, targetStr)
	newCmd, _ := cc.ParseCmd(true, model.FlowStrToStrs(newStr)...)
	cc.Blender.AddInsert(target, newCmd, 1)
	return currCmdIdx, nil
}

func BlenderInsertForAll(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	targetStr, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	if err != nil {
		return currCmdIdx, err
	}
	newStr, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "new")
	if err != nil {
		return currCmdIdx, err
	}

	target, _ := cc.ParseCmd(true, targetStr)
	newCmd, _ := cc.ParseCmd(true, model.FlowStrToStrs(newStr)...)
	cc.Blender.AddInsert(target, newCmd, -1)
	return currCmdIdx, nil
}

func BlenderInsertAfterOnce(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	targetStr, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	if err != nil {
		return currCmdIdx, err
	}
	newStr, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "new")
	if err != nil {
		return currCmdIdx, err
	}

	target, _ := cc.ParseCmd(true, targetStr)
	newCmd, _ := cc.ParseCmd(true, model.FlowStrToStrs(newStr)...)
	cc.Blender.AddInsertAfter(target, newCmd, 1)
	return currCmdIdx, nil
}

func BlenderInsertAfterForAll(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	targetStr, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "target")
	if err != nil {
		return currCmdIdx, err
	}
	newStr, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "new")
	if err != nil {
		return currCmdIdx, err
	}

	target, _ := cc.ParseCmd(true, targetStr)
	newCmd, _ := cc.ParseCmd(true, model.FlowStrToStrs(newStr)...)
	cc.Blender.AddInsertAfter(target, newCmd, -1)
	return currCmdIdx, nil
}

func BlenderClear(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	cc.Blender.Clear()
	return currCmdIdx, nil
}
