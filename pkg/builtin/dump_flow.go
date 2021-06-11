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

func SetDumpFlowDepth(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	depth := argv.GetInt("depth")
	env.GetLayer(core.EnvLayerSession).SetInt("display.flow.depth", depth)
	return true
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
