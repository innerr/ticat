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
	currCmdIdx int) (int, error) {

	cmdPath, err := tailModeCallArg(flow, currCmdIdx, argv, "cmd-path")
	if err != nil {
		return currCmdIdx, err
	}
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage().NoRecursive()
	dumpCmdByPath(cc, env, dumpArgs, cmdPath, "cmd.full")
	return currCmdIdx, nil
}

func DumpCmdWithDetails(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	cmdPath, err := tailModeCallArg(flow, currCmdIdx, argv, "cmd-path")
	if err != nil {
		return currCmdIdx, err
	}
	dumpArgs := display.NewDumpCmdArgs().NoRecursive()
	dumpCmdByPath(cc, env, dumpArgs, cmdPath, "")
	return currCmdIdx, nil
}

func DumpTailCmdWithUsage(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	err := flow.FirstErr()
	if err != nil {
		return currCmdIdx, err.Error
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
	currCmdIdx int) (int, error) {

	err := flow.FirstErr()
	if err != nil {
		return currCmdIdx, err.Error
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
