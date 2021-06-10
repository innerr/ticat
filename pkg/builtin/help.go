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
		return dumpMoreLessFindResult(flow, cc.Screen, env, "", cc.Cmds, false, findStrs...)
	}

	if len(flow.Cmds) >= 2 {
		cmdPathStr := flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
		cmd := cc.Cmds.GetSub(strings.Split(cmdPathStr, cc.Cmds.Strs.PathSep)...)
		if cmd == nil {
			panic("[GlobalHelpLessInfo] should never happen")
		}
		if len(findStrs) != 0 {
			return dumpMoreLessFindResult(flow, cc.Screen, env, cmdPathStr, cmd, false, findStrs...)
		}
		if len(flow.Cmds) > 2 ||
			cmd.Cmd() != nil && cmd.Cmd().Type() == core.CmdTypeFlow {
			return DumpFlowAllSimple(argv, cc, env, flow, currCmdIdx)
		}
		return dumpMoreLessFindResult(flow, cc.Screen, env, "", cmd, false)
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
		return dumpMoreLessFindResult(flow, cc.Screen, env, "", cc.Cmds, true, findStrs...)
	}

	if len(flow.Cmds) >= 2 {
		cmdPathStr := flow.Cmds[1].DisplayPath(cc.Cmds.Strs.PathSep, false)
		cmd := cc.Cmds.GetSub(strings.Split(cmdPathStr, cc.Cmds.Strs.PathSep)...)
		if cmd == nil {
			panic("[GlobalHelpLessInfo] should never happen")
		}
		if len(findStrs) != 0 {
			return dumpMoreLessFindResult(flow, cc.Screen, env, cmdPathStr, cmd, true, findStrs...)
		}
		if len(flow.Cmds) > 2 ||
			cmd.Cmd() != nil && cmd.Cmd().Type() == core.CmdTypeFlow {
			return DumpFlowSkeleton(argv, cc, env, flow, currCmdIdx)
		}
		return dumpMoreLessFindResult(flow, cc.Screen, env, "", cmd, true)
	}

	return dumpMoreLessFindResult(flow, cc.Screen, env, "", cc.Cmds, true, findStrs...)
}

func FindAny(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	findStrs := getFindStrsFromArgv(argv)
	if len(findStrs) == 0 {
		return true
	}
	display.DumpEnvFlattenVals(cc.Screen, env, findStrs...)
	display.DumpCmds(cc, false, 4, true, true, "", findStrs...)
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

	findStr := strings.Join(findStrs, " ")

	prt := func(text string) {
		display.PrintTipTitle(screen, env, text)
	}

	printer := display.NewCacheScreen()
	display.DumpAllCmds(cmd, printer, skeleton, 4, true, true, findStrs...)
	if len(findStrs) != 0 {
		tip := "search "
		matchStr := " commands matched '" + findStr + "'"
		if len(cmdPathStr) != 0 {
			if printer.OutputNum() > 0 {
				prt(tip + "branch '" + cmdPathStr + "', found" + matchStr + ":")
			} else {
				prt(tip + "branch '" + cmdPathStr + "', no" + matchStr + ".")
			}
		} else {
			if printer.OutputNum() > 0 {
				prt(tip + "and found" + matchStr)
			} else {
				prt(tip + "but no" + matchStr)
			}
		}
	} else {
		if len(cmdPathStr) != 0 {
			if printer.OutputNum() > 0 {
				prt("branch '" + cmdPathStr + "' has commands:")
			} else {
				prt("branch '" + cmdPathStr + "' has no commands. (this should never happen)")
			}
		} else {
			selfName := env.GetRaw("strs.self-name")
			if printer.OutputNum() > 0 {
				printer.WriteTo(screen)
				tips := display.NewTipBoxPrinter(screen, env, false)
				tips.Prints(selfName+" commands are listed above.", "", "find commands by keywords:", "")
				tips.Prints(display.SuggestStrsFindCmds(selfName)...)
				tips.Finish()
				return clearFlow(flow)
			} else {
				prt(selfName + " has no commands. (this should never happen)")
			}
		}
	}
	printer.WriteTo(screen)

	return clearFlow(flow)
}

func clearFlow(flow *core.ParsedCmds) (int, bool) {
	flow.Cmds = nil
	return 0, true
}
