package builtin

import (
	"fmt"

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

func DumpCmdsWithUsage(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage()
	return dumpCmds(argv, cc, env, flow, currCmdIdx, dumpArgs)
}

func DumpCmdsWithDetails(
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

	cmdPath := ""
	cmds := cc.Cmds
	if len(argv.GetRaw("cmd-path")) != 0 {
		cmdPath = tailModeCallArg(flow, currCmdIdx, argv, "cmd-path")
		cmds = cmds.GetSubByPath(cmdPath, true)
	}

	depth := 0
	if len(argv.GetRaw("depth")) != 0 {
		depth = argv.GetInt("depth")
		dumpArgs.SetMaxDepth(depth)
	}

	buf := display.NewCacheScreen()
	allShown := display.DumpCmds(cmds, buf, env, dumpArgs)

	text := ""
	if len(cmdPath) == 0 {
		text = "the tree of all commands:"
	} else {
		text = "the tree branch of '" + cmdPath + "'"
	}
	if !allShown {
		text += fmt.Sprintf(", some are not showed by arg depth='%d'", depth)
	}

	display.PrintTipTitle(cc.Screen, env, text)
	buf.WriteTo(cc.Screen)
	return clearFlow(flow)
}

func dumpCmds(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int,
	dumpArgs *display.DumpCmdArgs) (int, bool) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
	if len(findStrs) != 0 {
		dumpArgs.AddFindStrs(findStrs...)
	}

	cmdPath := argv.GetRaw("cmd-path")
	cmds := cc.Cmds
	if len(cmdPath) != 0 {
		cmds = cmds.GetSubByPath(cmdPath, true)
	}

	tag := argv.GetRaw("tag")
	if len(tag) != 0 {
		if len(findStrs) != 0 {
			panic(core.NewCmdError(flow.Cmds[currCmdIdx], "not support mix-search with tag and find-strs now"))
		}
		dumpArgs.AddFindStrs(tag).SetFindByTags()
	}

	source := argv.GetRaw("source")
	if len(source) != 0 {
		dumpArgs.SetSource(source)
	}

	if len(argv.GetRaw("depth")) != 0 {
		dumpArgs.SetMaxDepth(argv.GetInt("depth"))
	}

	display.DumpCmdsWithTips(cmds, cc.Screen, env, dumpArgs, cmdPath, false)
	return clearFlow(flow)
}

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

func DumpCmdWithDetails(
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
