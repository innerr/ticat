package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func DbgDumpFlow(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	cmds []core.ParsedCmd,
	currCmdIdx int) ([]core.ParsedCmd, int, bool) {

	display.DumpFlow(cc, env, cmds, cc.Cmds.Strs.PathSep, 4)
	return nil, 0, true
}

func DbgDumpEnv(_ core.ArgVals, cc *core.Cli, env *core.Env) bool {
	display.DumpEnv(cc.Screen, env, 4)
	return true
}

func DbgDumpCmds(argv core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	display.DumpCmds(cc, 4)
	return true
}
