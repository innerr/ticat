package builtin

import (
	"fmt"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func SetDumpFlowDepth(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	key := "display.flow.depth"
	if len(argv.GetRaw("depth")) != 0 {
		depth := argv.GetInt("depth")
		env.GetLayer(core.EnvLayerSession).SetInt("display.flow.depth", depth)
	}
	cc.Screen.Print(display.ColorKey(key, env) + display.ColorSymbol(" = ", env) + env.GetRaw(key) + "\n")
	return currCmdIdx, true
}

func DumpFlowAll(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return dumpFlowAll(cc, env, flow, currCmdIdx, false)
}

func DumpFlowAllSimple(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return dumpFlowAll(cc, env, flow, currCmdIdx, true)
}

func DumpFlow(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpFlowArgs()
	display.DumpFlow(cc, env, flow.GlobalEnv, flow.Cmds[currCmdIdx+1:], dumpArgs)
	return clearFlow(flow)
}

func DumpFlowSimple(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpFlowArgs().SetSimple()
	display.DumpFlow(cc, env, flow.GlobalEnv, flow.Cmds[currCmdIdx+1:], dumpArgs)
	return clearFlow(flow)
}

func DumpFlowSkeleton(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpFlowArgs().SetSkeleton()
	display.DumpFlow(cc, env, flow.GlobalEnv, flow.Cmds[currCmdIdx+1:], dumpArgs)

	deps := core.Depends{}
	core.CollectDepends(cc, env, flow, currCmdIdx+1, deps, false, EnvOpCmds())
	_, _, missedOsCmds := display.GatherOsCmdsExistingInfo(deps)

	checker := &core.EnvOpsChecker{}
	result := []core.EnvOpsCheckResult{}
	core.CheckEnvOps(cc, flow, env, checker, false, EnvOpCmds(), &result)
	fatals, risks, _ := display.AggEnvOpsCheckResult(result)

	adds := func(n int, s string) string {
		if n > 1 {
			return s + "s"
		}
		return s
	}

	errs := missedOsCmds + len(fatals.Result)
	if errs > 0 {
		errStr := adds(errs, "fatal")
		cc.Screen.Print(display.ColorError(fmt.Sprintf("(%s:%d)", errStr, errs), env) + "\n")
	} else if len(risks.Result) > 0 {
		riskStr := adds(errs, "risk")
		cc.Screen.Print(display.ColorWarn(fmt.Sprintf("(%s:%d)", riskStr, len(risks.Result)), env) + "\n")
	}

	return clearFlow(flow)
}

func DumpFlowDepends(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	env = env.Clone()
	deps := core.Depends{}
	core.CollectDepends(cc, env, flow, currCmdIdx+1, deps, false, EnvOpCmds())

	if len(deps) != 0 {
		display.DumpDepends(cc.Screen, env, deps)
	} else {
		display.PrintTipTitle(cc.Screen, env, "no depended os commands")
	}
	return clearFlow(flow)
}

func DumpFlowEnvOpsCheckResult(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	checker := &core.EnvOpsChecker{}
	result := []core.EnvOpsCheckResult{}
	env = env.Clone()
	core.CheckEnvOps(cc, flow, env, checker, false, EnvOpCmds(), &result)

	if len(result) != 0 {
		cmds := flow.Cmds[currCmdIdx+1:]
		display.DumpEnvOpsCheckResult(cc.Screen, cmds, env, result, cc.Cmds.Strs.PathSep)
	} else {
		display.PrintTipTitle(cc.Screen, env, "all env-ops are satisfied, can directly run")
	}

	return clearFlow(flow)
}

func dumpFlowAll(
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int,
	simple bool) (int, bool) {

	dumpArgs := display.NewDumpFlowArgs()
	dumpArgs.Simple = simple
	display.DumpFlow(cc, env, flow.GlobalEnv, flow.Cmds[currCmdIdx+1:], dumpArgs)

	deps := core.Depends{}
	core.CollectDepends(cc, env, flow, currCmdIdx+1, deps, false, EnvOpCmds())

	if len(deps) != 0 {
		cc.Screen.Print("\n")
		display.DumpDepends(cc.Screen, env, deps)
	}

	checker := &core.EnvOpsChecker{}
	result := []core.EnvOpsCheckResult{}
	core.CheckEnvOps(cc, flow, env, checker, false, EnvOpCmds(), &result)

	if len(result) != 0 {
		cc.Screen.Print("\n")
		display.DumpEnvOpsCheckResult(cc.Screen, flow.Cmds[currCmdIdx+1:], env, result, cc.Cmds.Strs.PathSep)
	}

	return clearFlow(flow)
}
