package display

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

func PrintGlobalHelp(screen core.Screen, env *core.Env) {
	pln := func(texts ...string) {
		for _, text := range texts {
			if len(text) == 0 {
				screen.Print("\n")
			} else {
				screen.Print("    " + text + "\n")
			}
		}
	}

	selfName := env.GetRaw("strs.self-name")

	PrintTipTitle(screen, env,
		selfName+": workflow automating in unix-pipe style")

	pln("")
	screen.Print("usage:\n")

	list := []func(*core.Env)[]string {
		SuggestExeCmds,
		SuggestExeCmdsWithArgs,
		SuggestListCmds,
		SuggestFindCmds,
		SuggestHubAdd,
		SuggestFlowAdd,
		SuggestDesc,
	}

	for _, fun := range list {
		pln("")
		pln(fun(env)...)
	}
}
