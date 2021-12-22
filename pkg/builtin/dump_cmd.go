package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func DumpCmdUsage(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath := tailModeCallArg(flow, currCmdIdx, argv, "cmd-path")
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage().NoRecursive()
	dumpCmdByPath(cc, env, dumpArgs, cmdPath)
	return currCmdIdx, true
}

func DumpCmdWithDetails(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath := tailModeCallArg(flow, currCmdIdx, argv, "cmd-path")
	dumpArgs := display.NewDumpCmdArgs().NoRecursive()
	dumpCmdByPath(cc, env, dumpArgs, cmdPath)
	return currCmdIdx, true
}

func DumpTailCmdWithUsage(
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
	dumpCmdByPath(cc, env, dumpArgs, cmdPath)
	return clearFlow(flow)
}

func DumpTailCmdWithDetails(
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
	dumpCmdByPath(cc, env, dumpArgs, cmdPath)
	return clearFlow(flow)
}

func dumpCmdByPath(cc *core.Cli, env *core.Env, args *display.DumpCmdArgs, path string) {
	cmds := cc.Cmds
	if len(path) != 0 {
		cmds = cmds.GetSubByPath(path, true)
	}
	if args.Skeleton {
		// TODO: XXX
		display.PrintTipTitle(cc.Screen, env, "command usage:", "", "(use [cmd.with-full-info] to show all info)")
	} else {
		display.PrintTipTitle(cc.Screen, env, "command details:")
	}
	display.DumpCmds(cmds, cc.Screen, env, args)
}
