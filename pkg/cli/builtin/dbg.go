package builtin

import (
	"github.com/pingcap/ticat/pkg/cli"
	"github.com/pingcap/ticat/pkg/cli/core"
)

func DbgDumpCmds(_ core.ArgVals, cc *core.Cli, env *core.Env, cmds []core.ParsedCmd,
	currCmdIdx int) ([]core.ParsedCmd, int, bool) {

	cli.DumpCmdsEx(cc.Screen, env, cmds, cc.Cmds.Strs.PathSep)
	return nil, 0, true
}

func DbgDumpEnv(_ core.ArgVals, cc *core.Cli, env *core.Env) bool {
	cli.DumpEnv(cc.Screen, env)
	return true
}

func DbgDumpMods(argv core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	cli.DumpMods(cc)
	return true
}
