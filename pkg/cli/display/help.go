package display

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

func PrintTipTitle(screen core.Screen, env *core.Env, msgs ...string) {
	if len(msgs) == 0 {
		return
	}
	tip := "<TIP>"
	tipLen := len(tip)
	if env.GetBool("display.utf8.symbols") {
		tip = "💡"
		tipLen = 2
	}
	buf := NewCacheScreen()
	buf.PrintEx(tip+" "+msgs[0], len(msgs[0])+1+tipLen)
	msgs = msgs[1:]
	for _, msg := range msgs {
		buf.Print("   " + msg)
	}
	PrintFramedLines(screen, env, buf)
}

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
		"use "+selfName+" to automate workflow in unix-pipe style",
		"", "cheat sheet:")

	pln("")
	pln(SuggestStrsExeCmds(selfName)...)
	pln("")
	pln(SuggestStrsExeCmdsWithArgs(selfName)...)
	pln("")
	pln(SuggestStrsListCmds(selfName)...)
	pln("")
	pln(SuggestStrsFindCmds(selfName)...)
	pln("")
	pln(SuggestStrsHubAdd(selfName)...)
	pln("")
	pln(SuggestStrsFlowAdd(selfName)...)
	pln("")
	pln(SuggestStrsDesc(selfName)...)
}

func SuggestStrsExeCmds(selfName string) []string {
	return []string{
		selfName + " cmd1 : cmd2 : cmd3              - execute commands one by one,",
		"                                        like unix-pipe, use ':' instead of '|'",
	}
}

func SuggestStrsExeCmdsWithArgs(selfName string) []string {
	return []string{
		selfName + " dbg.echo msg=hi : slp 1s        - an example of executing commands,",
		"                                        'dbg.echo' is a command name",
	}
}

func SuggestStrsListCmds(selfName string) []string {
	return []string{
		selfName + " -                               - list all commands",
		selfName + " +                               - list all commands with details",
	}
}

func SuggestStrsFindCmds(selfName string) []string {
	return []string{
		selfName + " str1 str2 :-                    - search commands",
		selfName + " str1 str2 :+                    - search commands with details",
	}
}

func SuggestStrsHubAdd(selfName string) []string {
	return []string{
		selfName + " h.init                          - get more commands by adding a default git repo",
		selfName + " h.+ innerr/tidb." + selfName + "           - get more commands by adding a git repo,",
		"                                        could use https address like:",
		"                                        'https://github.com/innerr/tidb." + selfName + "'",
	}
}

func SuggestStrsFlowAdd(selfName string) []string {
	return []string{
		selfName + " dbg.echo hi : slp 1s : f.+ xx   - create a flow 'xx' by 'f.+' for convenient",
		selfName + " xx                              - execute command 'xx'",
	}
}

func SuggestStrsDesc(selfName string) []string {
	return []string{
		selfName + " xx :-                           - show what 'xx' will do without executing it",
		selfName + " xx :+                           - show what 'xx' will do without executing it, with details",
	}
}
