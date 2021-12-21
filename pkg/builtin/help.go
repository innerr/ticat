package builtin

import (
	"fmt"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func DumpTailCmdUsage(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath := argv.GetRaw("cmd-path")
	if len(cmdPath) == 0 {
		cmdPath = flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
	} else {
		cmdPath = cc.ParseCmd(cmdPath, true)
	}
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage().NoRecursive()
	dumpCmdsByPath(cc, env, dumpArgs, cmdPath)
	return clearFlow(flow)
}

func DumpTailCmdInfo(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath := argv.GetRaw("cmd-path")
	if len(cmdPath) == 0 {
		cmdPath = flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
	} else {
		cmdPath = cc.ParseCmd(cmdPath, true)
	}
	dumpArgs := display.NewDumpCmdArgs().NoRecursive()
	dumpCmdsByPath(cc, env, dumpArgs, cmdPath)
	return clearFlow(flow)
}

func DumpTailCmdSub(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath := flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton()
	dumpCmdsByPath(cc, env, dumpArgs, cmdPath)
	return clearFlow(flow)
}

func DumpTailCmdSubUsage(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath := flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage()
	dumpCmdsByPath(cc, env, dumpArgs, cmdPath)
	return clearFlow(flow)
}

func DumpTailCmdSubDetails(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath := flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
	dumpArgs := display.NewDumpCmdArgs()
	dumpCmdsByPath(cc, env, dumpArgs, cmdPath)
	return clearFlow(flow)
}

/*
func GlobalFindCmd(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().AddFindStrs(findStrs...)
	display.DumpCmdsWithTips(cc.Cmds, cc.Screen, env, dumpArgs, "", true)
	return clearFlow(flow)
}

func GlobalFindCmdWithUsage(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().AddFindStrs(findStrs...).SetShowUsage()
	display.DumpCmdsWithTips(cc.Cmds, cc.Screen, env, dumpArgs, "", true)
	return clearFlow(flow)
}

func GlobalFindCmdWithDetail(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
	dumpArgs := display.NewDumpCmdArgs().AddFindStrs(findStrs...)
	display.DumpCmdsWithTips(cc.Cmds, cc.Screen, env, dumpArgs, "", true)
	return clearFlow(flow)
}
*/

func ListTags(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	display.ListTags(cc.Cmds, cc.Screen, env)
	return currCmdIdx, true
}

func FindByTags(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
	if len(findStrs) == 0 {
		display.ListTags(cc.Cmds, cc.Screen, env)
		return currCmdIdx, true
	}

	dumpArgs := display.NewDumpCmdArgs().AddFindStrs(findStrs...).SetFindByTags().SetSkeleton()
	display.DumpCmds(cc.Cmds, cc.Screen, env, dumpArgs)
	return currCmdIdx, true
}

func GlobalHelp(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	target := argv.GetRaw("target")
	if len(target) != 0 {
		cmdPath := cc.ParseCmd(target, false)
		if len(cmdPath) == 0 {
			display.PrintErrTitle(cc.Screen, env, fmt.Sprintf("'%s' is not a valid command", target))
		} else {
			dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage().NoRecursive()
			dumpCmdsByPath(cc, env, dumpArgs, cmdPath)
		}
	} else {
		display.PrintGlobalHelp(cc, env)
	}
	return currCmdIdx, true
}

func SelfHelp(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	display.PrintSelfHelp(cc.Screen, env)
	return currCmdIdx, true
}
