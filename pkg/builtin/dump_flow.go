package builtin

import (
	"fmt"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func DumpFlowAll(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return dumpFlowAll(argv, cc, env, flow, currCmdIdx, false)
}

func DumpFlowAllSimple(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return dumpFlowAll(argv, cc, env, flow, currCmdIdx, true)
}

func DumpFlow(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	err := flow.FirstErr()
	if err != nil {
		panic(err.Error)
	}

	dumpArgs := display.NewDumpFlowArgs().SetMaxDepth(argv.GetInt("depth")).
		SetMaxTrivial(argv.GetInt("unfold-trivial"))
	display.DumpFlow(cc, env, flow, currCmdIdx+1, dumpArgs, EnvOpCmds())
	return clearFlow(flow)
}

func DumpFlowSimple(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	err := flow.FirstErr()
	if err != nil {
		panic(err.Error)
	}

	dumpArgs := display.NewDumpFlowArgs().SetSimple().SetMaxDepth(argv.GetInt("depth")).
		SetMaxTrivial(argv.GetInt("unfold-trivial"))
	display.DumpFlow(cc, env, flow, currCmdIdx+1, dumpArgs, EnvOpCmds())
	return clearFlow(flow)
}

func DumpFlowSkeleton(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	err := flow.FirstErr()
	if err != nil {
		panic(err.Error)
	}

	dumpArgs := display.NewDumpFlowArgs().SetSkeleton().SetMaxDepth(argv.GetInt("depth")).
		SetMaxTrivial(argv.GetInt("unfold-trivial"))
	display.DumpFlow(cc, env, flow, currCmdIdx+1, dumpArgs, EnvOpCmds())

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

	err := flow.FirstErr()
	if err != nil {
		panic(err.Error)
	}

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

	err := flow.FirstErr()
	if err != nil {
		panic(err.Error)
	}

	checker := &core.EnvOpsChecker{}
	result := []core.EnvOpsCheckResult{}
	env = env.Clone()
	core.CheckEnvOps(cc, flow, env, checker, false, EnvOpCmds(), &result)

	if len(result) != 0 {
		cmds := flow.Cmds[currCmdIdx+1:]
		display.DumpEnvOpsCheckResult(cc.Screen, cmds, env, result, cc.Cmds.Strs.PathSep)
	} else {
		display.PrintTipTitle(cc.Screen, env, "all env-ops are satisfied, this command can directly run")
	}

	return clearFlow(flow)
}

func dumpFlowAll(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int,
	simple bool) (int, bool) {

	err := flow.FirstErr()
	if err != nil {
		panic(err.Error)
	}

	dumpArgs := display.NewDumpFlowArgs().SetMaxDepth(argv.GetInt("depth")).
		SetMaxTrivial(argv.GetInt("unfold-trivial"))
	dumpArgs.Simple = simple
	display.DumpFlow(cc, env, flow, currCmdIdx+1, dumpArgs, EnvOpCmds())

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
