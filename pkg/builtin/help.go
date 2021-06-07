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

	if len(flow.Cmds) >= 2 {
		cmdPathStr := flow.Cmds[1].DisplayPath(cc.Cmds.Strs.PathSep, false)
		cmd := cc.Cmds.GetSub(strings.Split(cmdPathStr, cc.Cmds.Strs.PathSep)...)
		if cmd == nil {
			return clearFlow(flow)
		}
		findStrs := getFindStrsFromArgv(argv)
		if len(findStrs) != 0 {
			display.DumpCmds(cc, false, 4, true, true, cmdPathStr, findStrs...)
			return clearFlow(flow)
		}
		if len(flow.Cmds) > 2 ||
			cmd.Cmd() != nil && cmd.Cmd().Type() == core.CmdTypeFlow {
			return DumpFlowAllSimple(argv, cc, env, flow, currCmdIdx)
		}
		if cmd.HasSub() && cmd.Cmd() == nil {
			display.DumpCmds(cc, true, 4, false, true, cmdPathStr)
		} else {
			display.DumpCmds(cc, false, 4, false, false, cmdPathStr)
		}
		return clearFlow(flow)
	}

	if len(argv.GetRaw("1st-str")) != 0 {
		findStrs := getFindStrsFromArgv(argv)
		if len(findStrs) != 0 {
			display.DumpCmds(cc, false, 4, true, true, "", findStrs...)
		}
		return clearFlow(flow)
	}

	display.PrintGlobalHelp(cc)
	return clearFlow(flow)
}

func GlobalHelpLessInfo(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	if len(flow.Cmds) >= 2 {
		cmdPathStr := flow.Cmds[1].DisplayPath(cc.Cmds.Strs.PathSep, false)
		cmd := cc.Cmds.GetSub(strings.Split(cmdPathStr, cc.Cmds.Strs.PathSep)...)
		if cmd == nil {
			return clearFlow(flow)
		}
		findStrs := getFindStrsFromArgv(argv)
		if len(findStrs) != 0 {
			display.DumpCmds(cc, true, 4, true, true, cmdPathStr, findStrs...)
			return clearFlow(flow)
		}
		if len(flow.Cmds) > 2 ||
			cmd.Cmd() != nil && cmd.Cmd().Type() == core.CmdTypeFlow {
			return DumpFlowSkeleton(argv, cc, env, flow, currCmdIdx)
		}
		if cmd.HasSub() {
			display.DumpCmds(cc, true, 4, false, true, cmdPathStr)
		} else {
			display.DumpCmds(cc, false, 4, false, false, cmdPathStr)
		}
		return clearFlow(flow)
	}

	if len(argv.GetRaw("1st-str")) != 0 {
		findStrs := getFindStrsFromArgv(argv)
		if len(findStrs) != 0 {
			display.DumpCmds(cc, true, 4, true, true, "", findStrs...)
		}
		return clearFlow(flow)
	}

	display.PrintGlobalHelp(cc)
	return clearFlow(flow)
}

func FindAny(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	findStrs := getFindStrsFromArgv(argv)
	if len(findStrs) == 0 {
		return true
	}
	display.DumpEnvFlattenVals(cc.Screen, env, findStrs...)
	display.DumpCmds(cc, false, 4, true, true, "", findStrs...)
	return true
}

func GlobalHelp(_ core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	display.PrintGlobalHelp(cc)
	return true
}

func clearFlow(flow *core.ParsedCmds) (int, bool) {
	flow.Cmds = nil
	return 0, true
}
