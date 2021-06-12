package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func DumpEnv(_ core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	display.DumpEnv(cc.Screen, env, 4)
	return true
}

func DumpEnvAbbrs(_ core.ArgVals, cc *core.Cli, _ *core.Env, _ core.ParsedCmd) bool {
	display.DumpEnvAbbrs(cc, 4)
	return true
}

func DumpEssentialEnvFlattenVals(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	display.DumpEssentialEnvFlattenVals(cc.Screen, env, getFindStrsFromArgv(argv)...)
	return true
}

func DumpEnvFlattenVals(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	display.DumpEnvFlattenVals(cc.Screen, env, getFindStrsFromArgv(argv)...)
	return true
}
