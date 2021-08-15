package builtin

import (
	"fmt"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func DumpCmdListSimple(argv core.ArgVals, cc *core.Cli, env *core.Env, _ []core.ParsedCmd) bool {
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().AddFindStrs(getFindStrsFromArgv(argv)...)
	display.DumpCmdsWithTips(cc.Cmds, cc.Screen, env, dumpArgs, "", false)
	return true
}

func DumpCmdList(argv core.ArgVals, cc *core.Cli, env *core.Env, _ []core.ParsedCmd) bool {
	dumpArgs := display.NewDumpCmdArgs().AddFindStrs(getFindStrsFromArgv(argv)...)
	display.DumpCmdsWithTips(cc.Cmds, cc.Screen, env, dumpArgs, "", false)
	return true
}

func DumpCmdNoRecursive(argv core.ArgVals, cc *core.Cli, env *core.Env, flow []core.ParsedCmd) bool {
	cmdPath := argv.GetRaw("cmd-path")
	if len(cmdPath) == 0 {
		return DumpCmdListSimple(argv, cc, env, flow)
	}
	dumpArgs := display.NewDumpCmdArgs().NoRecursive()
	dumpCmdsByPath(cc, env, dumpArgs, cmdPath)
	return true
}

func DumpCmdTree(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	dumpArgs := display.NewDumpCmdArgs().NoFlatten()
	dumpCmdsByPath(cc, env, dumpArgs, argv.GetRaw("cmd-path"))
	return currCmdIdx, true
}

func DumpCmdTreeSkeleton(argv core.ArgVals, cc *core.Cli, env *core.Env, _ []core.ParsedCmd) bool {
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().NoFlatten()
	dumpCmdsByPath(cc, env, dumpArgs, argv.GetRaw("cmd-path"))
	return true
}

func DumpCmdsWhoWriteKey(argv core.ArgVals, cc *core.Cli, env *core.Env, flow []core.ParsedCmd) bool {
	cmd := flow[0]
	key := argv.GetRaw("key")
	if len(key) == 0 {
		panic(core.NewCmdError(cmd, "missed arg 'key'"))
	}

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetMatchWriteKey(key)
	display.DumpCmdsWithTips(cc.Cmds, cc.Screen, env, dumpArgs, "", false)
	return true
}

func dumpCmdsByPath(cc *core.Cli, env *core.Env, args *display.DumpCmdArgs, path string) {
	if len(path) == 0 && !args.Recursive {
		display.PrintTipTitle(cc.Screen, env,
			"no info about root command. (this should never happen)")
		return
	}
	cmds := cc.Cmds
	if len(path) != 0 {
		cmds = cmds.GetSub(strings.Split(path, cc.Cmds.Strs.PathSep)...)
		if cmds == nil {
			panic(fmt.Errorf("can't find sub cmd tree by path '%s'", path))
		}
	}
	display.DumpCmdsWithTips(cmds, cc.Screen, env, args, path, false)
}
