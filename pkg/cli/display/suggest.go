package display

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

func GlobalSuggestExeCmds(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" cmd1 : cmd2 : cmd3", indent) + "- execute commands one by one,",
		padR("", indent+2) + "similar to unix-pipe, just use ':' instead of '|'",
	}
}

func GlobalSuggestExeCmdsWithArgs(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" echo msg=hi : sleep 1s", indent) +
			"- an example of executing commands",
	}
}

func GlobalSuggestShowCmdInfo(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" cmd      echo", indent) + "- show usage of command 'echo'",
		padR(selfName+" cmd.full echo", indent) + "- show full details of command 'echo'",
	}
}

func GlobalSuggestCmdTree(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" cmds.tree foo", indent) + "- display commands in the branch of 'foo'",
		padR(selfName+" cmds.tree foo.bar", indent) + "  then browse the tree level by level",
	}
}

func GlobalSuggestListCmds(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" cmds path=foo src=git-addr", indent) + "- display commands from 'git-addr' in the branch of 'foo'",
		padR(selfName+" cmds      foo     git-addr", indent) + "  same as above",
		padR(selfName+" cmds.full foo     git-addr", indent) + "  same as above, with full details",
	}
}

func GlobalSuggestFindCmds(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" find      str1 str2", indent) + "- search commands",
		padR(selfName+" find.more str1 str2", indent) + "  and display more info",
	}
}

func GlobalSuggestHubAdd(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" hub.init", indent) +
			"- get more commands by adding a default git repo",
		padR(selfName+" hub.add innerr/tidb."+selfName, indent) +
			"- get more commands by adding a git repo,",
		padR("", indent+2) + "could be a full git address",
	}
}

func GlobalSuggestFlowAdd(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" echo hi : sleep 1s :flow.save foo", indent) +
			"- create a flow 'foo' by 'flow.save' for convenient",
		padR(selfName+" foo", indent) + "- execute command 'foo'",
	}
}

func GlobalSuggestDesc(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" foo :desc", indent) + "- show the plan graphic of 'foo' without executing it",
		padR(selfName+" foo :desc.more", indent) + "  same as above, with more details",
	}
}

func GlobalSuggestSessions(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" sessions.last", indent) + "- show the status of last execution",
		padR(selfName+" sessions.last.desc", indent) + "- show the executed graphic of last session",
		padR(selfName+" sessions str1 str2", indent) + "- search executed sessions",
	}
}

func GlobalSuggestAdvance(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" cmds.tree desc", indent) + "- explore the branch 'desc'",
		padR("", indent) + "  a command set for describe commands before executing",
		padR(selfName+" cmds.tree cmds", indent) + "- a command set to locate commands by branch, source",
		padR(selfName+" cmds.tree find", indent) + "- a command set to search commands by strings",
		padR(selfName+" cmds.tree flow", indent) + "- a command set to manage saved flows",
		padR(selfName+" cmds.tree hub", indent) + "- a command set to manage command source(repo or dir)",
		padR(selfName+" cmds.tree sessions", indent) + "- a command set to manage executed sessions",
	}
}

func GlobalSuggestShortcut(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" find shortcut", indent) + "- show shortcut commands",
	}
}

func SuggestHubBranch(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	explain := "- branch 'hub' usage"
	return []string{
		padR(selfName+" cmds hub", indent) + explain,
		padR(selfName+" cmds.full hub", indent) + explain + ", with details",
	}
}

func SuggestFindCmdsMore(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" find.more str1 str2", indent) + "- search commands and display more info",
		padR(selfName+" find.full  str1 str2", indent) + "- search commands and display full details",
	}
}

func SuggestAddAndSaveEnv(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	res := []string{
		padR(selfName+" {k1=v2 k2=v2} cmd1 : cmd2", indent) + "- set/change env key-values, one time only",
		"",
		padR(selfName+" {k1=v2 k2=v2} env.save", indent) + "- save the changed key-values",
		padR(selfName+" cmd1 : cmd2", indent) + "- execute with saved key-values",
		"",
	}
	return append(res, SuggestListEnv(env)...)
}

func SuggestListEnv(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" env", indent) + "- list changed env key-values, not include system's",
		padR(selfName+" env.ls", indent) + "- list all key-values, include system's",
	}
}

func SuggestFindEnv(env *core.Env, subCmd string) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" env"+subCmd+" str1 str2", indent) + "- search env by strings",
	}
}

func SuggestFindCmdsLess(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" find str1 str2", indent) + "- search commands",
	}
}

func SuggestFilterRepoInMove(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" hub.move str1 str2", indent) + "- filter repos to only one result by strings, then move saved flows to it",
	}
}

func SuggestFindCmdsInRepo(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" cmds src=repo-name", indent) + "- search commands by source(repo or dir)",
		"",
		padR(selfName+" cmd cmd1", indent) + "- show command usage",
		padR(selfName+" cmd.full cmd1", indent) + "- command details",
	}
}

func SuggestHubAddShort(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" hub.init", indent) + "- add the default git repo.",
		padR(selfName+" hub.add innerr/tidb."+selfName, indent) + "- add a git repo,",
		padR("", indent+2) + "use https by default.",
	}
}

func SuggestEnvSetting(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	explain := "- set 'k=v', then display it"
	return []string{
		padR(selfName+" {k=v} : env", indent) + explain,
		padR(selfName+" {k=v} dummy : env", indent) + explain,
		padR(selfName+" dummy : {k=v} env", indent) + explain,
	}
}

func SuggestFindProvider(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	explain := "- find modules which will write this key"
	return []string{
		padR(selfName+" env.who-write key-name", indent) + explain,
	}
}

func SuggestFlowsFilter(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" flow str1 str2", indent) + "- find flows matched these strings",
	}
}

func SuggestTailInfo(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" cmd cmd1", indent) + "- show command's detail info",
	}
}

func SuggestFindCmds(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" find      str1 str2", indent) + "- search commands",
		padR(selfName+" find.more str1 str2", indent) + "  and display more info",
	}
}

func SuggestFlowAdd(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" echo hi : sleep 1s :flow.save foo", indent) +
			"- create a flow 'foo' by 'flow.save' for convenient",
		padR(selfName+" foo", indent) + "- execute command 'foo'",
	}
}

func getSuggestArgs(env *core.Env) (selfName string, explainIndent int) {
	selfName = env.GetRaw("strs.self-name")
	explainIndent = env.GetInt("display.hint.indent.2rd")
	return
}
