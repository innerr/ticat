package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func DumpEnvTree(_ core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	display.PrintTipTitle(cc.Screen, env, "all env key-values:")
	display.DumpEnvTree(cc.Screen, env, 4)
	return true
}

func DumpEnvAbbrs(_ core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	display.PrintTipTitle(cc.Screen, env, "all env key abbrs:")
	display.DumpEnvAbbrs(cc, 4)
	return true
}

func DumpEnvFlattenVals(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	findStrs := getFindStrsFromArgv(argv)
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
	return true
}

func DumpEssentialEnvFlattenVals(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	findStrs := getFindStrsFromArgv(argv)
	screen := display.NewCacheScreen()
	display.DumpEssentialEnvFlattenVals(screen, env, findStrs...)
	if screen.OutputNum() <= 0 {
		if len(findStrs) != 0 {
			display.PrintTipTitle(cc.Screen, env,
				"no matched saved env keys.")
		} else {
			display.PrintTipTitle(cc.Screen, env,
				"no saved env keys.",
				"",
				display.SuggestAddAndSaveEnv(env))
		}
	} else if len(findStrs) == 0 {
		display.PrintTipTitle(cc.Screen, env,
			"essential env key-values: (use command 'e.ls' to show all)")
	} else {
		display.PrintTipTitle(cc.Screen, env,
			"matched essential env key-values: (use command 'e.ls' to show more)")
	}
	screen.WriteTo(cc.Screen)
	if display.TooMuchOutput(env, screen) {
		display.PrintTipTitle(cc.Screen, env,
			"filter env keys by:",
			"",
			display.SuggestFindEnv(env, ""))
	}
	return true
}
