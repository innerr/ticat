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

	PrintTipTitle(screen, env,
		"use "+env.GetRaw("strs.self-name")+" to automate workflow in unix-pipe style",
		"", "cheat sheet:")

	pln("")
	pln(SuggestStrsExeCmds()...)
	pln("")
	pln(SuggestStrsExeCmdsWithArgs()...)
	pln("")
	pln(SuggestStrsListCmds()...)
	pln("")
	pln(SuggestStrsFindCmds()...)
	pln("")
	pln(SuggestStrsHubAdd()...)
	pln("")
	pln(SuggestStrsFlowAdd()...)
	pln("")
	pln(SuggestStrsDesc()...)
}

func SuggestStrsExeCmds() []string {
	return []string{
		"ticat cmd1 : cmd2 : cmd3              - execute commands one by one,",
		"                                        like unix-pipe, use ':' instead of '|'",
	}
}

func SuggestStrsExeCmdsWithArgs() []string {
	return []string{
		"ticat dbg.echo msg=hi : slp 1s        - an example of executing commands,",
		"                                        'dbg.echo' is a command name",
	}
}

func SuggestStrsListCmds() []string {
	return []string{
		"ticat -                               - list all commands",
		"ticat +                               - list all commands with details",
	}
}

func SuggestStrsFindCmds() []string {
	return []string{
		"ticat str1 str2 :-                    - search commands",
		"ticat str1 str2 :+                    - search commands with details",
	}
}

func SuggestStrsHubAdd() []string {
	return []string{
		"ticat h.init                          - get more commands by adding a default git repo",
		"ticat h.+ innerr/tidb.ticat           - get more commands by adding a git repo,",
		"                                        could use https address like:",
		"                                        'https://github.com/innerr/tidb.ticat'",
	}
}

func SuggestStrsFlowAdd() []string {
	return []string{
		"ticat dbg.echo hi : slp 1s : f.+ xx   - create a flow 'xx' by 'f.+' for convenient",
		"ticat xx                              - execute command 'xx'",
	}
}

func SuggestStrsDesc() []string {
	return []string{
		"ticat xx :-                           - show what 'xx' will do without executing it",
		"ticat xx :+                           - show what 'xx' will do without executing it, with details",
	}
}
