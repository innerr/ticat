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
		"                                        similar to unix-pipe, use ':' instead of '|'",
	}
}

func SuggestStrsExeCmdsWithArgs(selfName string) []string {
	return []string{
		selfName + " dbg.echo msg=hi : sleep 1s      - an example of executing commands,",
		"                                        'dbg.echo' is a command name, 'msg' is an arg",
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

func SuggestStrsFindRepoTag(selfName string) []string {
	return []string{
		selfName + " @ready repo-name str1 :-        - search commands in repo",
		selfName + " @ready repo-name str1 :+        - search commands in repo, with details",
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

func SuggestStrsHubAddShort(selfName string) []string {
	return []string{
		selfName + " h.init                          - add a default git repo.",
		selfName + " h.+ innerr/tidb." + selfName + "           - add a git repo,",
		"                                        could use https address.",
	}
}

func SuggestStrsEnvSetting(selfName string) []string {
	return []string{
		selfName + " {k=v} : env                     - set 'k=v', then display it",
		selfName + " {k=v} dummy : env               - set 'k=v', then display it",
		selfName + " dummy : {k=v} env               - set 'k=v', then display it",
	}
}

func SuggestStrsHubBranch(selfName string) []string {
	return []string{
		selfName + " h :-                            - branch 'hub' usage",
		selfName + " h :+                            - branch 'hub' usage, with details",
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

func SuggestStrsFindConfigFlows(selfName string) []string {
	return []string{
		selfName + " @ready @provider :-             - find configuring flows",
		selfName + " @ready @provider :+             - find configuring flows, with details",
	}
}

func SuggestStrsFindProvider(selfName string) []string {
	return []string{
		selfName + " key-name write :-               - find modules will write this key",
		selfName + " key-name write :+               - find modules will write this key, with details",
	}
}
