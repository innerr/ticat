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
	return dumpCmds(argv, cc, env, flow, currCmdIdx, dumpArgs, "", "cmds.more")
}

func DumpCmdsWithUsage(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage()
	return dumpCmds(argv, cc, env, flow, currCmdIdx, dumpArgs, "cmds", "cmds.full")
}

func DumpCmdsWithDetails(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs()
	return dumpCmds(argv, cc, env, flow, currCmdIdx, dumpArgs, "cmds.more", "")
}

func DumpTailCmdSub(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton()
	return dumpTailCmdSub(argv, cc, env, flow, currCmdIdx, dumpArgs, "", "tail-sub.more")
}

func DumpTailCmdSubWithUsage(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage()
	return dumpTailCmdSub(argv, cc, env, flow, currCmdIdx, dumpArgs, "tail-sub", "tail-sub.full")
}

func DumpTailCmdSubWithDetails(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs()
	return dumpTailCmdSub(argv, cc, env, flow, currCmdIdx, dumpArgs, "tail-sub.more", "")
}

func dumpCmds(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int,
	dumpArgs *display.DumpCmdArgs,
	lessDetailCmd string,
	moreDetailCmd string) (int, bool) {

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
			panic(core.NewCmdError(flow.Cmds[currCmdIdx], "not support mix-search with tag and keywords now"))
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

	display.DumpCmdsWithTips(cmds, cc.Screen, env, dumpArgs, cmdPath, lessDetailCmd, moreDetailCmd)
	return clearFlow(flow)
}

func dumpTailCmdSub(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int,
	dumpArgs *display.DumpCmdArgs,
	lessDetailCmd string,
	moreDetailCmd string) (int, bool) {

	err := flow.FirstErr()
	if err != nil {
		panic(err.Error)
	}

	findStrs := getFindStrsFromArgv(argv)
	if len(findStrs) != 0 {
		dumpArgs.AddFindStrs(findStrs...)
	}

	source := argv.GetRaw("source")
	if len(source) != 0 {
		dumpArgs.SetSource(source)
	}

	if len(argv.GetRaw("depth")) != 0 {
		dumpArgs.SetMaxDepth(argv.GetInt("depth"))
	}

	cmdPath := flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
	cmds := cc.Cmds
	if len(cmdPath) != 0 {
		cmds = cmds.GetSubByPath(cmdPath, true)
	}

	display.DumpCmdsWithTips(cmds, cc.Screen, env, dumpArgs, cmdPath, lessDetailCmd, moreDetailCmd)
	return clearFlow(flow)
}
