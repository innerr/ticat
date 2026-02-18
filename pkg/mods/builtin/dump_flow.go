package builtin

import (
	"fmt"

	"github.com/innerr/ticat/pkg/cli/display"
	"github.com/innerr/ticat/pkg/core/model"
)

func DumpFlowAll(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	return dumpFlowAll(argv, cc, env, flow, currCmdIdx, false)
}

func DumpFlowAllSimple(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	return dumpFlowAll(argv, cc, env, flow, currCmdIdx, true)
}

func DumpFlow(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	err := flow.FirstErr()
	if err != nil {
		return currCmdIdx, err.Error
	}

	dumpArgs := display.NewDumpFlowArgs().SetMaxDepth(argv.GetInt("depth")).
		SetMaxTrivial(argv.GetInt("unfold-trivial"))
	display.DumpFlow(cc, env, flow, currCmdIdx+1, dumpArgs, EnvOpCmds())
	printFatalRiskMark(cc, env, flow, currCmdIdx)
	return clearFlow(flow)
}

func DumpFlowSimple(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	err := flow.FirstErr()
	if err != nil {
		return currCmdIdx, err.Error
	}

	dumpArgs := display.NewDumpFlowArgs().SetSimple().SetMaxDepth(argv.GetInt("depth")).
		SetMaxTrivial(argv.GetInt("unfold-trivial"))
	display.DumpFlow(cc, env, flow, currCmdIdx+1, dumpArgs, EnvOpCmds())
	printFatalRiskMark(cc, env, flow, currCmdIdx)
	return clearFlow(flow)
}

func DumpFlowSkeleton(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	err := flow.FirstErr()
	if err != nil {
		return currCmdIdx, err.Error
	}

	dumpArgs := display.NewDumpFlowArgs().SetSkeleton().SetMaxDepth(argv.GetInt("depth")).
		SetMaxTrivial(argv.GetInt("unfold-trivial"))
	display.DumpFlow(cc, env, flow, currCmdIdx+1, dumpArgs, EnvOpCmds())
	printFatalRiskMark(cc, env, flow, currCmdIdx)

	return clearFlow(flow)
}

func printFatalRiskMark(
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) {

	deps := model.Depends{}
	model.CollectDepends(cc, env, flow, currCmdIdx+1, deps, false, EnvOpCmds())
	_, _, missedOsCmds := display.GatherOsCmdsExistingInfo(deps)

	checker := &model.EnvOpsChecker{}
	result := []model.EnvOpsCheckResult{}
	model.CheckEnvOps(cc, flow, env, checker, false, EnvOpCmds(), &result)
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
}

func DumpFlowDepends(
	_ model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	err := flow.FirstErr()
	if err != nil {
		return currCmdIdx, err.Error
	}

	env = env.Clone()
	deps := model.Depends{}
	model.CollectDepends(cc, env, flow, currCmdIdx+1, deps, false, EnvOpCmds())

	if len(deps) != 0 {
		display.DumpDepends(cc.Screen, env, deps)
	} else {
		display.PrintTipTitle(cc.Screen, env, "no depended os commands")
	}
	return clearFlow(flow)
}

func DumpFlowEnvOpsCheckResult(
	_ model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	err := flow.FirstErr()
	if err != nil {
		return currCmdIdx, err.Error
	}

	checker := &model.EnvOpsChecker{}
	result := []model.EnvOpsCheckResult{}
	env = env.Clone()
	model.CheckEnvOps(cc, flow, env, checker, false, EnvOpCmds(), &result)

	if len(result) != 0 {
		cmds := flow.Cmds[currCmdIdx+1:]
		display.DumpEnvOpsCheckResult(cc.Screen, cmds, env, result, cc.Cmds.Strs.PathSep)
	} else {
		display.PrintTipTitle(cc.Screen, env, "all env-ops are satisfied, this command can directly run")
	}

	return clearFlow(flow)
}

func dumpFlowAll(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int,
	simple bool) (int, error) {

	err := flow.FirstErr()
	if err != nil {
		return currCmdIdx, err.Error
	}

	dumpArgs := display.NewDumpFlowArgs().SetMaxDepth(argv.GetInt("depth")).
		SetMaxTrivial(argv.GetInt("unfold-trivial"))
	dumpArgs.Simple = simple
	display.DumpFlow(cc, env, flow, currCmdIdx+1, dumpArgs, EnvOpCmds())

	deps := model.Depends{}
	model.CollectDepends(cc, env, flow, currCmdIdx+1, deps, false, EnvOpCmds())

	if len(deps) != 0 {
		cc.Screen.Print("\n")
		display.DumpDepends(cc.Screen, env, deps)
	}

	checker := &model.EnvOpsChecker{}
	result := []model.EnvOpsCheckResult{}
	model.CheckEnvOps(cc, flow, env, checker, false, EnvOpCmds(), &result)

	if len(result) != 0 {
		cc.Screen.Print("\n")
		display.DumpEnvOpsCheckResult(cc.Screen, flow.Cmds[currCmdIdx+1:], env, result, cc.Cmds.Strs.PathSep)
	}

	return clearFlow(flow)
}
