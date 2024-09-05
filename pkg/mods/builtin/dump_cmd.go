package builtin

import (
	"github.com/innerr/ticat/pkg/cli/display"
	"github.com/innerr/ticat/pkg/core/model"
)

func DumpCmdUsage(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath := tailModeCallArg(flow, currCmdIdx, argv, "cmd-path")
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage().NoRecursive()
	dumpCmdByPath(cc, env, dumpArgs, cmdPath, "cmd.full")
	return currCmdIdx, true
}

func DumpCmdWithDetails(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath := tailModeCallArg(flow, currCmdIdx, argv, "cmd-path")
	dumpArgs := display.NewDumpCmdArgs().NoRecursive()
	dumpCmdByPath(cc, env, dumpArgs, cmdPath, "")
	return currCmdIdx, true
}

func DumpTailCmdWithUsage(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	err := flow.FirstErr()
	if err != nil {
		panic(err.Error)
	}

	cmdPath := argv.GetRaw("cmd-path")
	if len(cmdPath) == 0 {
		cmdPath = flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
	} else {
		cmdPath = cc.NormalizeCmd(true, cmdPath)
	}
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage().NoRecursive()
	dumpCmdByPath(cc, env, dumpArgs, cmdPath, "==")
	return clearFlow(flow)
}

func DumpTailCmdWithDetails(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	err := flow.FirstErr()
	if err != nil {
		panic(err.Error)
	}

	cmdPath := argv.GetRaw("cmd-path")
	if len(cmdPath) == 0 {
		cmdPath = flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
	} else {
		cmdPath = cc.NormalizeCmd(true, cmdPath)
	}
	dumpArgs := display.NewDumpCmdArgs().NoRecursive()
	dumpCmdByPath(cc, env, dumpArgs, cmdPath, "")
	return clearFlow(flow)
}

func dumpCmdByPath(cc *model.Cli, env *model.Env, args *display.DumpCmdArgs, path string, fullDetailCmd string) {
	cmds := cc.Cmds
	if len(path) != 0 {
		cmds = cmds.GetSubByPath(path, true)
	}
	if args.Skeleton {
		if len(fullDetailCmd) != 0 {
			display.PrintTipTitle(cc.Screen, env,
				"command usage: (use '"+fullDetailCmd+"' for full details)")
		} else {
			display.PrintTipTitle(cc.Screen, env, "command usage:")
		}
	} else {
		display.PrintTipTitle(cc.Screen, env, "full command details:")
	}
	display.DumpCmds(cmds, cc.Screen, env, args)
}
