package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func DumpFlow(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	display.DumpFlow(cc, env, flow.Cmds[currCmdIdx+1:], cc.Cmds.Strs.PathSep, 4)

	checker := &core.EnvOpsChecker{}
	result := []core.EnvOpsCheckResult{}
	core.CheckEnvOps(cc, flow, env, checker, false, &result)

	if len(result) != 0 {
		cc.Screen.Print("\n")
		display.DumpEnvOpsCheckResult(cc.Screen, result, cc.Cmds.Strs.PathSep)
	}

	flow.Cmds = nil
	return 0, true
}

func DumpEnv(_ core.ArgVals, cc *core.Cli, env *core.Env) bool {
	display.DumpEnv(cc.Screen, env, 4)
	return true
}

func DumpCmdTree(argv core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	display.DumpCmds(cc, false, 4, false, argv.GetRaw("path"))
	return true
}

func DumpCmdTreeSimple(argv core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	display.DumpCmds(cc, true, 4, false, argv.GetRaw("path"))
	return true
}

func DumpCmds(argv core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	display.DumpCmds(cc, false, 4, true, "", getFindStrsFromArgv(argv)...)
	return true
}

func DumpEnvAbbrs(_ core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	display.DumpEnvAbbrs(cc, 4)
	return true
}

func DumpEnvFlattenVals(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	display.DumpEnvFlattenVals(cc.Screen, env, getFindStrsFromArgv(argv)...)
	return true
}

func getFindStrsFromArgv(argv core.ArgVals) (findStrs []string) {
	names := []string{
		"1st-str",
		"2nd-str",
		"3rd-str",
		"4th-str",
		"5th-str",
		"6th-str",
	}
	for _, name := range names {
		val := argv.GetRaw(name)
		if len(val) != 0 {
			findStrs = append(findStrs, val)
		}
	}
	return
}
