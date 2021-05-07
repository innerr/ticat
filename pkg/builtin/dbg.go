package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func DbgDumpFlow(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	display.DumpFlow(cc, env, flow.Cmds[currCmdIdx+1:], cc.Cmds.Strs.PathSep, 4)
	flow.Cmds = nil
	return 0, true
}

func DbgDumpEnv(_ core.ArgVals, cc *core.Cli, env *core.Env) bool {
	display.DumpEnv(cc.Screen, env, 4)
	return true
}

func DbgDumpCmdTree(_ core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	display.DumpCmds(cc, 4, false, "")
	return true
}

func DbgDumpCmds(argv core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	display.DumpCmds(cc, 4, true, getFindStrsFromArgv(argv)...)
	return true
}

func DbgDumpEnvAbbrs(_ core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	display.DumpEnvAbbrs(cc, 4)
	return true
}

func DbgDumpEnvFlattenVals(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	display.DumpEnvFlattenVals(cc.Screen, env, getFindStrsFromArgv(argv)...)
	return true
}

func getFindStrsFromArgv(argv core.ArgVals) (findStrs []string) {
	str1 := argv.GetRaw("1st-str")
	if len(str1) != 0 {
		findStrs = append(findStrs, str1)
	}
	str2 := argv.GetRaw("2rd-str")
	if len(str2) != 0 {
		findStrs = append(findStrs, str2)
	}
	str3 := argv.GetRaw("3th-str")
	if len(str3) != 0 {
		findStrs = append(findStrs, str3)
	}
	return
}
