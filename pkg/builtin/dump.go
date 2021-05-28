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

	display.DumpFlow(cc, env, flow.Cmds[currCmdIdx+1:],
		cc.Cmds.Strs.PathSep, 4, false, false)
	flow.Cmds = nil
	return 0, true
}

func DumpFlowSimple(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	display.DumpFlow(cc, env, flow.Cmds[currCmdIdx+1:],
		cc.Cmds.Strs.PathSep, 4, true, false)
	flow.Cmds = nil
	return 0, true
}

func DumpFlowSkeleton(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	display.DumpFlow(cc, env, flow.Cmds[currCmdIdx+1:],
		cc.Cmds.Strs.PathSep, 4, true, true)
	flow.Cmds = nil
	return 0, true
}

func DumpFlowDepends(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	deps := display.Depends{}
	display.CollectDepends(cc, flow.Cmds[currCmdIdx+1:], deps)

	if len(deps) != 0 {
		display.DumpDepends(cc, env, deps)
	} else {
		cc.Screen.Print("no depended os commands\n")
	}
	flow.Cmds = nil
	return 0, true
}

func DumpFlowEnvOpsCheckResult(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	checker := &core.EnvOpsChecker{}
	result := []core.EnvOpsCheckResult{}
	core.CheckEnvOps(cc, flow, env, checker, false, &result)

	if len(result) != 0 {
		display.DumpEnvOpsCheckResult(cc.Screen, env, result, cc.Cmds.Strs.PathSep)
	} else {
		cc.Screen.Print("all env-ops are satisfied, can directly run\n")
	}

	flow.Cmds = nil
	return 0, true
}

func DumpFlowAllSimple(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return dumpFlowAll(cc, env, flow, currCmdIdx, true)
}

func DumpFlowAll(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return dumpFlowAll(cc, env, flow, currCmdIdx, false)
}

func dumpFlowAll(
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int,
	simple bool) (int, bool) {

	cmds := flow.Cmds[currCmdIdx+1:]
	display.DumpFlow(cc, env, cmds, cc.Cmds.Strs.PathSep, 4, simple, false)

	deps := display.Depends{}
	display.CollectDepends(cc, flow.Cmds[currCmdIdx+1:], deps)

	if len(deps) != 0 {
		cc.Screen.Print("\n")
		display.DumpDepends(cc, env, deps)
	}

	checker := &core.EnvOpsChecker{}
	result := []core.EnvOpsCheckResult{}
	core.CheckEnvOps(cc, flow, env, checker, false, &result)

	if len(result) != 0 {
		cc.Screen.Print("\n")
		display.DumpEnvOpsCheckResult(cc.Screen, env, result, cc.Cmds.Strs.PathSep)
	}

	flow.Cmds = nil
	return 0, true
}

func DumpEnv(_ core.ArgVals, cc *core.Cli, env *core.Env) bool {
	display.DumpEnv(cc.Screen, env, 4)
	return true
}

func DumpCmdNoRecursive(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	display.DumpCmds(cc, false, 4, false, false, argv.GetRaw("cmd-path"))
	return true
}

func DumpCmdTree(argv core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	display.DumpCmds(cc, false, 4, false, true, argv.GetRaw("cmd-path"))
	return true
}

func DumpCmdTreeSkeleton(argv core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	display.DumpCmds(cc, true, 4, false, true, argv.GetRaw("cmd-path"))
	return true
}

func DumpCmdListSimple(argv core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	display.DumpCmds(cc, true, 4, true, true, argv.GetRaw("cmd-path"),
		getFindStrsFromArgv(argv)...)
	return true
}

func DumpCmds(argv core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	display.DumpCmds(cc, false, 4, true, true, "", getFindStrsFromArgv(argv)...)
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
