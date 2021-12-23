package display

import (
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/utils"
)

func PrintGlobalHelp(cc *core.Cli, env *core.Env) {
	if len(cc.Helps.Sections) == 0 {
		PrintSelfHelp(cc.Screen, env)
		return
	}

	// TODO: with color output this is not right, disable it by setting to a very big value
	_, width := utils.GetTerminalWidth()
	width = 4096

	pln := func(line string) {
		line = DecodeColor(line, env)
		if len(line) <= width {
			cc.Screen.Print(line + "\n")
			return
		}

		indent := 0
		for _, char := range line {
			if char != ' ' && char != '\t' {
				break
			}
			indent += 1
		}
		prefix := line[:indent]

		printWithPrefix := func(printed bool, line string) {
			if printed {
				cc.Screen.Print(prefix)
				cc.Screen.Print(strings.TrimLeft(line, " \t") + "\n")
			} else {
				cc.Screen.Print(line + "\n")
			}
		}

		printed := false
		for {
			if len(line) > width {
				printWithPrefix(printed, line[:width])
				line = line[width:]
				printed = true
				continue
			} else {
				printWithPrefix(printed, line)
			}
			break
		}
	}

	for _, help := range cc.Helps.Sections {
		PrintTipTitle(cc.Screen, env, help.Title)
		for _, line := range help.Text {
			pln(line)
		}
	}
}

func PrintSelfHelp(screen core.Screen, env *core.Env) {
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
	screen.Print(ColorHelp("usage:\n", env))

	list := []func(*core.Env) []string{
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
	}

	for _, fun := range list {
		pln("")
		pln(fun(env)...)
	}
}
