package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func DumpEnvTree(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	// TODO: this can't be colorize, because it share codes with executor display(with frame, color will break len(str))
	assertNotTailMode(flow, currCmdIdx, flow.TailMode)
	display.PrintTipTitle(cc.Screen, env, "all env key-values:")
	display.DumpEnvTree(cc.Screen, env, 4)
	return currCmdIdx, true
}

func DumpEnvAbbrs(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx, flow.TailMode)
	display.PrintTipTitle(cc.Screen, env, "all env key abbrs:")
	display.DumpEnvAbbrs(cc, env, 4)
	return currCmdIdx, true
}

func DumpEnvFlattenVals(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	findStrs := getFindStrsFromArgv(argv)
	if flow.TailMode {
		findStrs = append(findStrs, gatherInputsFromFlow(flow, currCmdIdx)...)
	}

	screen := display.NewCacheScreen()
	display.DumpEnvFlattenVals(screen, env, findStrs...)
	if screen.OutputNum() <= 0 {
		display.PrintTipTitle(cc.Screen, env, "no matched env keys.")
	} else if len(findStrs) == 0 {
		display.PrintTipTitle(cc.Screen, env, "all env key-values:")
	} else {
		display.PrintTipTitle(cc.Screen, env, "all matched env key-values:")
	}
	screen.WriteTo(cc.Screen)
	if display.TooMuchOutput(env, screen) {
		display.PrintTipTitle(cc.Screen, env,
			"filter env keys by:",
			"",
			display.SuggestFindEnv(env, ".ls"))
	}
	return currCmdIdx, true
}

func DumpEssentialEnvFlattenVals(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	findStrs := getFindStrsFromArgv(argv)
	if flow.TailMode {
		findStrs = append(findStrs, gatherInputsFromFlow(flow, currCmdIdx)...)
	}

	screen := display.NewCacheScreen()
	display.DumpEssentialEnvFlattenVals(screen, env, findStrs...)
	if screen.OutputNum() <= 0 {
		if len(findStrs) != 0 {
			display.PrintTipTitle(cc.Screen, env,
				"no matched changed env key-values.",
				"(system's are not included)",
				"",
				display.SuggestListEnv(env))
		} else {
			display.PrintTipTitle(cc.Screen, env,
				"no changed env key-values found.",
				"(system's are not included)",
				"",
				"env usage:",
				"",
				display.SuggestAddAndSaveEnv(env))
		}
	} else if len(findStrs) == 0 {
		display.PrintTipTitle(cc.Screen, env,
			"essential env key-values: (use command 'e.ls' to show all)")
	} else {
		display.PrintTipTitle(cc.Screen, env,
			"matched essential env key-values: (use command 'e.ls' to show all)")
	}
	screen.WriteTo(cc.Screen)
	if display.TooMuchOutput(env, screen) {
		display.PrintTipTitle(cc.Screen, env,
			"filter env keys by:",
			"",
			display.SuggestFindEnv(env, ""))
	}
	return currCmdIdx, true
}
