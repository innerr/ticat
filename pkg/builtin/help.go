package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func GlobalHelpMoreInfo(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return DumpFlowAllSimple(argv, cc, env, flow, currCmdIdx)
}

func GlobalHelpLessInfo(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return DumpFlowSkeleton(argv, cc, env, flow, currCmdIdx)
}

func DumpTailCmdInfo(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath := flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
	dumpArgs := display.NewDumpCmdArgs().NoRecursive()
	dumpCmdsByPath(cc, env, dumpArgs, cmdPath)
	return clearFlow(flow)
}

func DumpTailCmdUsage(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath := flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage().NoRecursive()
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

func FindByTags(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
	if len(findStrs) == 0 {
		display.ListTags(cc.Cmds, cc.Screen, env)
		return clearFlow(flow)
	}

	dumpArgs := display.NewDumpCmdArgs().AddFindStrs(findStrs...).SetFindByTags().SetSkeleton()
	display.DumpCmds(cc.Cmds, cc.Screen, env, dumpArgs)
	return clearFlow(flow)
}

func GlobalHelp(_ core.ArgVals, cc *core.Cli, env *core.Env, _ []core.ParsedCmd) bool {
	display.PrintGlobalHelp(cc, env)
	return true
}

func SelfHelp(_ core.ArgVals, cc *core.Cli, env *core.Env, _ []core.ParsedCmd) bool {
	display.PrintSelfHelp(cc.Screen, env)
	return true
}

// This is a mess, rewrite
/*
func GlobalHelpMoreInfo(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return globalHelpLessMoreInfo(argv, cc, env, flow, currCmdIdx, false)
}

func GlobalHelpLessInfo(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return globalHelpLessMoreInfo(argv, cc, env, flow, currCmdIdx, true)
}

func globalHelpLessMoreInfo(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int,
	skeleton bool) (int, bool) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)

	for _, cmd := range flow.Cmds {
		if cmd.ParseResult.Error == nil {
			continue
		}
		findStrs = append(cmd.ParseResult.Input, findStrs...)
		cmdPathStr := ""
		cic := cc.Cmds
		//	if !cmd.IsEmpty() {
		//		cic = cmd.Last().Matched.Cmd
		//		cmdPathStr = cmd.DisplayPath(cc.Cmds.Strs.PathSep, true)
		//	}
		return dumpMoreLessFindResult(flow, currCmdIdx, cc.Screen, env, cmdPathStr, cic, skeleton, findStrs...)
	}

	if len(flow.Cmds) >= 2 {
		cmdPathStr := flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
		cmd := cc.Cmds.GetSub(strings.Split(cmdPathStr, cc.Cmds.Strs.PathSep)...)
		if cmd == nil {
			panic("[globalHelpLessMoreInfo] should never happen")
		}
		cmdPathStr = flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, true)
		if len(findStrs) != 0 {
			return dumpMoreLessFindResult(flow, currCmdIdx, cc.Screen, env, cmdPathStr, cmd, skeleton, findStrs...)
		}
		if len(flow.Cmds) > 2 ||
			cmd.Cmd() != nil && cmd.Cmd().Type() == core.CmdTypeFlow {
			if skeleton {
				return DumpFlowSkeleton(argv, cc, env, flow, currCmdIdx)
			} else {
				return DumpFlowAllSimple(argv, cc, env, flow, currCmdIdx)
			}
		}
		input := flow.Last().ParseResult.Input
		if len(input) > 1 {
			return dumpMoreLessFindResult(flow, currCmdIdx, cc.Screen, env, "", cc.Cmds, skeleton, input...)
		}
		return dumpMoreLessFindResult(flow, currCmdIdx, cc.Screen, env, cmdPathStr, cmd, skeleton)
	}

	return dumpMoreLessFindResult(flow, currCmdIdx, cc.Screen, env, "", cc.Cmds, skeleton, findStrs...)
}

func dumpMoreLessFindResult(
	flow *core.ParsedCmds,
	currCmdIdx int,
	screen core.Screen,
	env *core.Env,
	cmdPathStr string,
	cmd *core.CmdTree,
	skeleton bool,
	findStrs ...string) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().AddFindStrs(findStrs...)
	dumpArgs.Skeleton = skeleton
	display.DumpCmdsWithTips(cmd, screen, env, dumpArgs, cmdPathStr, true)
	return clearFlow(flow)
}
*/
