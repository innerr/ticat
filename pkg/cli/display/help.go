package display

import (
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func PrintGlobalHelp(cc *core.Cli, env *core.Env) {
	if len(cc.Helps.Sections) == 0 {
		PrintSelfHelp(cc.Screen, env)
		return
	}

	width := env.GetInt("display.width")

	pln := func(line string) {
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
	cc.Screen.Print("\n")
	selfName := cc.Cmds.Strs.SelfName
	cc.Screen.Print("(use command 'help." + selfName + "' to show " + selfName + "'s self usage)\n")
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
	screen.Print("usage:\n")

	list := []func(*core.Env) []string{
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
