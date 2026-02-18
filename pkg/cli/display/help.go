package display

import (
	"github.com/innerr/ticat/pkg/core/model"
)

func PrintGlobalHelp(cc *model.Cli, env *model.Env) {
	// TODO: use man page instead of help
	PrintSelfHelp(cc.Screen, env)
}

func PrintSelfHelp(screen model.Screen, env *model.Env) {
	pln := func(texts ...string) {
		for _, text := range texts {
			if len(text) == 0 {
				_ = screen.Print("\n")
			} else {
				_ = screen.Print("    " + text + "\n")
			}
		}
	}

	selfName := env.GetRaw("strs.self-name")
	PrintTipTitle(screen, env,
		selfName+": workflow automating in unix-pipe style")

	pln("")
	_ = screen.Print(ColorHelp("usage:\n", env))

	list := []func(*model.Env) []string{
		GlobalSuggestExeCmds,
		GlobalSuggestExeCmdsWithArgs,
		GlobalSuggestShowCmdInfo,
		GlobalSuggestCmdTree,
		GlobalSuggestListCmds,
		GlobalSuggestFindCmds,
		GlobalSuggestHubAdd,
		GlobalSuggestFlowAdd,
		GlobalSuggestDesc,
		GlobalSuggestSessions,
		GlobalSuggestAdvance,
		GlobalSuggestShortcut,
		GlobalSuggestInteract,
	}

	for _, fun := range list {
		pln("")
		pln(fun(env)...)
	}
}
