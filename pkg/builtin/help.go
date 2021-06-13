package builtin

import (
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func GlobalHelpMoreInfo(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	findStrs := getFindStrsFromArgv(argv)

	for _, cmd := range flow.Cmds {
		if cmd.ParseError.Error == nil {
			continue
		}
		findStrs = append(cmd.ParseError.Input, findStrs...)
		cmdPathStr := ""
		cic := cc.Cmds
		if !cmd.IsEmpty() {
			cic = cmd.Last().Matched.Cmd
			cmdPathStr = cmd.DisplayPath(cc.Cmds.Strs.PathSep, true)
		}
		return dumpMoreLessFindResult(flow, cc.Screen, env, cmdPathStr, cic, false, findStrs...)
	}

	if len(flow.Cmds) >= 2 {
		cmdPathStr := flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
		cmd := cc.Cmds.GetSub(strings.Split(cmdPathStr, cc.Cmds.Strs.PathSep)...)
		if cmd == nil {
			panic("[GlobalHelpMoreInfo] should never happen")
		}
		cmdPathStr = flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, true)
		if len(findStrs) != 0 {
			return dumpMoreLessFindResult(flow, cc.Screen, env, cmdPathStr, cmd, false, findStrs...)
		}
		if len(flow.Cmds) > 2 ||
			cmd.Cmd() != nil && cmd.Cmd().Type() == core.CmdTypeFlow {
			return DumpFlowAllSimple(argv, cc, env, flow, currCmdIdx)
		}
		return dumpMoreLessFindResult(flow, cc.Screen, env, cmdPathStr, cmd, false)
	}

	return dumpMoreLessFindResult(flow, cc.Screen, env, "", cc.Cmds, false, findStrs...)
}

func GlobalHelpLessInfo(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	findStrs := getFindStrsFromArgv(argv)

	for _, cmd := range flow.Cmds {
		if cmd.ParseError.Error == nil {
			continue
		}
		findStrs = append(cmd.ParseError.Input, findStrs...)
		cmdPathStr := ""
		cic := cc.Cmds
		if !cmd.IsEmpty() {
			cic = cmd.Last().Matched.Cmd
			cmdPathStr = cmd.DisplayPath(cc.Cmds.Strs.PathSep, true)
		}
		return dumpMoreLessFindResult(flow, cc.Screen, env, cmdPathStr, cic, true, findStrs...)
	}

	if len(flow.Cmds) >= 2 {
		cmdPathStr := flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
		cmd := cc.Cmds.GetSub(strings.Split(cmdPathStr, cc.Cmds.Strs.PathSep)...)
		if cmd == nil {
			panic("[GlobalHelpLessInfo] should never happen")
		}
		cmdPathStr = flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, true)
		if len(findStrs) != 0 {
			return dumpMoreLessFindResult(flow, cc.Screen, env, cmdPathStr, cmd, true, findStrs...)
		}
		if len(flow.Cmds) > 2 ||
			cmd.Cmd() != nil && cmd.Cmd().Type() == core.CmdTypeFlow {
			return DumpFlowSkeleton(argv, cc, env, flow, currCmdIdx)
		}
		return dumpMoreLessFindResult(flow, cc.Screen, env, cmdPathStr, cmd, true)
	}

	return dumpMoreLessFindResult(flow, cc.Screen, env, "", cc.Cmds, true, findStrs...)
}

func DumpTellTailCmd(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	if len(flow.Cmds) < 2 {
		return clearFlow(flow)
	}
	cmdPath := flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
	dumpArgs := display.NewDumpCmdArgs().NoFlatten().NoRecursive()
	dumpCmdsByPath(cc, env, dumpArgs, cmdPath)
	return clearFlow(flow)
}

func FindAny(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	findStrs := getFindStrsFromArgv(argv)
	if len(findStrs) == 0 {
		return true
	}
	display.DumpEnvFlattenVals(cc.Screen, env, findStrs...)
	dumpArgs := display.NewDumpCmdArgs().AddFindStrs(findStrs...)
	display.DumpCmdsWithTips(cc.Cmds, cc.Screen, env, dumpArgs, "", false)
	return true
}

func GlobalHelp(_ core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	display.PrintGlobalHelp(cc.Screen, env)
	return true
}

func dumpMoreLessFindResult(
	flow *core.ParsedCmds,
	screen core.Screen,
	env *core.Env,
	cmdPathStr string,
	cmd *core.CmdTree,
	skeleton bool,
	findStrs ...string) (int, bool) {

	printer := display.NewCacheScreen()
	dumpArgs := display.NewDumpCmdArgs().AddFindStrs(findStrs...)
	dumpArgs.Skeleton = skeleton
	display.DumpCmdsWithTips(cmd, printer, env, dumpArgs, cmdPathStr, true)
	printer.WriteTo(screen)
	return clearFlow(flow)
}
