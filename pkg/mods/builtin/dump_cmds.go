package builtin

import (
	"github.com/innerr/ticat/pkg/cli/display"
	"github.com/innerr/ticat/pkg/core/model"
)

func DumpCmds(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton()
	return dumpCmds(argv, cc, env, flow, currCmdIdx, dumpArgs, "", "cmds.more")
}

func DumpCmdsWithUsage(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage()
	return dumpCmds(argv, cc, env, flow, currCmdIdx, dumpArgs, "cmds", "cmds.full")
}

func DumpCmdsWithDetails(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs()
	return dumpCmds(argv, cc, env, flow, currCmdIdx, dumpArgs, "cmds.more", "")
}

func DumpTailCmdSub(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton()
	return dumpTailCmdSub(argv, cc, env, flow, currCmdIdx, dumpArgs, "", "tail-sub.more")
}

func DumpTailCmdSubWithUsage(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage()
	return dumpTailCmdSub(argv, cc, env, flow, currCmdIdx, dumpArgs, "tail-sub", "tail-sub.full")
}

func DumpTailCmdSubWithDetails(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs()
	return dumpTailCmdSub(argv, cc, env, flow, currCmdIdx, dumpArgs, "tail-sub.more", "")
}

func dumpCmds(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
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
			panic(model.NewCmdError(flow.Cmds[currCmdIdx], "not support mix-search with tag and keywords now"))
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
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
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
