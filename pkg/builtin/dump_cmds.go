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

func DumpTailCmdSub(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton()
	return dumpTailCmdSub(argv, cc, env, flow, currCmdIdx, dumpArgs)
}

func DumpTailCmdSubWithUsage(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage()
	return dumpTailCmdSub(argv, cc, env, flow, currCmdIdx, dumpArgs)
}

func DumpTailCmdSubWithDetails(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs()
	return dumpTailCmdSub(argv, cc, env, flow, currCmdIdx, dumpArgs)
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

	display.DumpCmdsWithTips(cmds, cc.Screen, env, dumpArgs, cmdPath)
	return clearFlow(flow)
}

func dumpTailCmdSub(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int,
	dumpArgs *display.DumpCmdArgs) (int, bool) {

	err := flow.FirstErr()
	if err != nil {
		panic(err.Error)
	}
	cmdPath := flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
	dumpCmdByPath(cc, env, dumpArgs, cmdPath)
	return clearFlow(flow)
}
