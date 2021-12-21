package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func DumpCmds(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton()
	return dumpCmds(argv, cc, env, flow, currCmdIdx, dumpArgs)
}

func DumpCmdsMore(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage()
	return dumpCmds(argv, cc, env, flow, currCmdIdx, dumpArgs)
}

func DumpCmdsFull(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs()
	return dumpCmds(argv, cc, env, flow, currCmdIdx, dumpArgs)
}

func DumpCmdsTree(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().NoFlatten()
	return dumpCmds(argv, cc, env, flow, currCmdIdx, dumpArgs)
}

func DumpCmdsTreeMore(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage().NoFlatten()
	return dumpCmds(argv, cc, env, flow, currCmdIdx, dumpArgs)
}

func DumpCmdsTreeFull(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().NoFlatten()
	return dumpCmds(argv, cc, env, flow, currCmdIdx, dumpArgs)
}

func dumpCmds(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int,
	dumpArgs *display.DumpCmdArgs) (int, bool) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)

	cmdPath := argv.GetRaw("cmd-path")
	cmds := cc.Cmds
	if len(cmdPath) != 0 {
		cmds = cmds.GetSubByPath(cmdPath, true)
	}

	//source := argv.GetRaw("source")
	//tag := argv.GetRaw("tag")
	//depth := argv.GetRaw("depth")

	dumpArgs.AddFindStrs(findStrs...)
	display.DumpCmdsWithTips(cmds, cc.Screen, env, dumpArgs, cmdPath, false)
	return currCmdIdx, true
}

/*
func DumpCmdListSimple(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().AddFindStrs(findStrs...)
	display.DumpCmdsWithTips(cc.Cmds, cc.Screen, env, dumpArgs, "", false)
	return currCmdIdx, true
}

func DumpCmdList(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
	dumpArgs := display.NewDumpCmdArgs().AddFindStrs(findStrs...)
	display.DumpCmdsWithTips(cc.Cmds, cc.Screen, env, dumpArgs, "", false)
	return currCmdIdx, true
}

func DumpCmdTree(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath := tailModeCallArg(flow, currCmdIdx, argv, "cmd-path")
	dumpArgs := display.NewDumpCmdArgs().NoFlatten()
	dumpCmdsByPath(cc, env, dumpArgs, cmdPath)
	return currCmdIdx, true
}

func DumpCmdTreeSkeleton(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	recursive := argv.GetBool("recursive")
	cmdPath := tailModeCallArg(flow, currCmdIdx, argv, "cmd-path")
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().NoFlatten()
	if !recursive {
		dumpArgs.NoRecursive()
	}
	dumpCmdsByPath(cc, env, dumpArgs, cmdPath)
	return currCmdIdx, true
}
*/

func DumpCmdsWhoWriteKey(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	key := tailModeCallArg(flow, currCmdIdx, argv, "key")
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetMatchWriteKey(key)
	display.DumpCmdsWithTips(cc.Cmds, cc.Screen, env, dumpArgs, "", false)
	return currCmdIdx, true
}

func DumpCmdUsage(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath := tailModeCallArg(flow, currCmdIdx, argv, "cmd-path")
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage().NoRecursive()
	dumpCmdsByPath(cc, env, dumpArgs, cmdPath)
	return currCmdIdx, true
}

func DumpCmdFull(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath := tailModeCallArg(flow, currCmdIdx, argv, "cmd-path")
	dumpArgs := display.NewDumpCmdArgs().NoRecursive()
	dumpCmdsByPath(cc, env, dumpArgs, cmdPath)
	return currCmdIdx, true
}

func dumpCmdsByPath(cc *core.Cli, env *core.Env, args *display.DumpCmdArgs, path string) {
	if len(path) == 0 && !args.Recursive {
		display.PrintTipTitle(cc.Screen, env,
			"no info about root command. (this should never happen)")
		return
	}
	cmds := cc.Cmds
	if len(path) != 0 {
		cmds = cmds.GetSubByPath(path, true)
	}
	display.DumpCmdsWithTips(cmds, cc.Screen, env, args, path, false)
}
