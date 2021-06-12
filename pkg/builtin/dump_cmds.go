package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func DumpCmdNoRecursive(argv core.ArgVals, cc *core.Cli, env *core.Env, cmd core.ParsedCmd) bool {
	cmdPath := argv.GetRaw("cmd-path")
	if len(cmdPath) == 0 {
		return DumpCmdListSimple(argv, cc, env, cmd)
	}
	dumpArgs := display.NewDumpCmdArgs().NoFlatten().NoRecursive()
	display.DumpCmdsByPath(cc, dumpArgs, cmdPath)
	return true
}

func DumpCmdTree(argv core.ArgVals, cc *core.Cli, _ *core.Env, _ core.ParsedCmd) bool {
	dumpArgs := display.NewDumpCmdArgs().NoFlatten()
	display.DumpCmdsByPath(cc, dumpArgs, argv.GetRaw("cmd-path"))
	return true
}

func DumpCmdTreeSkeleton(argv core.ArgVals, cc *core.Cli, _ *core.Env, _ core.ParsedCmd) bool {
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().NoFlatten()
	display.DumpCmdsByPath(cc, dumpArgs, argv.GetRaw("cmd-path"))
	return true
}

func DumpCmdListSimple(argv core.ArgVals, cc *core.Cli, _ *core.Env, _ core.ParsedCmd) bool {
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().AddFindStrs(getFindStrsFromArgv(argv)...)
	display.DumpCmdsByPath(cc, dumpArgs, argv.GetRaw("cmd-path"))
	return true
}

func DumpCmds(argv core.ArgVals, cc *core.Cli, _ *core.Env, _ core.ParsedCmd) bool {
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().AddFindStrs(getFindStrsFromArgv(argv)...)
	display.DumpCmds(cc.Cmds, cc.Screen, dumpArgs)
	return true
}
