package display

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

func SuggestExeCmds(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" cmd1 : cmd2 : cmd3", indent) + "- execute commands one by one,",
		padR("", indent+2) + "similar to unix-pipe, use ':' instead of '|'",
	}
}

func SuggestExeCmdsWithArgs(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" dbg.echo msg=hi : sleep 1s", indent) +
			"- an example of executing commands,",
		padR("", indent+2) + "'dbg.echo' is a command name, 'msg' is an arg",
		padR(selfName+" dbg.echo :=", indent) + "- show usage of the command",
		padR(selfName+" dbg.echo :==", indent) + "- show full details of the command",
	}
}

func SuggestListCmds(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" /", indent) + "- list all commands",
		padR(selfName+" //", indent) + "- list all commands with usage",
	}
}

func SuggestFindCmdsMore(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" str1 str2 ://", indent) + "- search commands with usage",
	}
}

func SuggestAddAndSaveEnv(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	res := []string{
		padR(selfName+" {k1=v2 k2=v2} cmd1 : cmd2", indent) + "- set/change env key-values, one time only",
		"",
		padR(selfName+" {k1=v2 k2=v2} e.save", indent) + "- save the changed key-values",
		padR(selfName+" cmd1 : cmd2", indent) + "- execute with saved key-values",
		"",
	}
	return append(res, SuggestListEnv(env)...)
}

func SuggestListEnv(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" e", indent) + "- list changed env key-values, not include system's",
		padR(selfName+" e.ls", indent) + "- list all key-values, include system's",
	}
}

func SuggestFindEnv(env *core.Env, subCmd string) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" e"+subCmd+" str1 str2", indent) + "- search env by strings",
		padR(selfName+" str1 str2 ::e"+subCmd, indent) + "- same as above, tail mode",
	}
}

func SuggestFindCmdsLess(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" cmd :~", indent) + "- search commands in the branch of 'cmd'",
		padR(selfName+" cmd :~~", indent) + "- search commands in the branch of 'cmd', with usage",
		padR(selfName+" str1 str2 :/", indent) + "- search commands",
	}
}

func SuggestFindCmds(env *core.Env) []string {
	return append(SuggestFindCmdsLess(env), SuggestFindCmdsMore(env)...)
}

func SuggestFilterRepoInMove(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" h.mv str1 str2", indent) + "- filter repos by strings then move flows to it",
		padR(selfName+" str1 str2 ::h.mv", indent) + "- same as above, in tail-mode",
	}
}

func SuggestFindCmdsInRepo(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" repo-name str1 str2 ::cmds", indent) + "- search commands in repo",
		"",
		padR(selfName+" cmd  ::=", indent) + "- show command usage",
		padR(selfName+" cmd  ::==", indent) + "- command details",
	}
}

func SuggestHubAdd(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	exampleRepo := env.GetRaw("display.example-https-repo")
	return []string{
		padR(selfName+" h.init", indent) +
			"- get more commands by adding a default git repo",
		padR(selfName+" h.+ innerr/tidb."+selfName, indent) +
			"- get more commands by adding a git repo,",
		padR("", indent+2) + "could use https address like:",
		padR("", indent+2) + "'" + exampleRepo + "'",
	}
}

func SuggestHubAddShort(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" h.init", indent) + "- add the default git repo.",
		padR(selfName+" h.+ innerr/tidb."+selfName, indent) + "- add a git repo,",
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

func SuggestHubBranch(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	explain := "- branch 'hub' usage"
	return []string{
		padR(selfName+" h :~", indent) + explain,
		padR(selfName+" h :~~", indent) + explain + ", with details",
	}
}

func SuggestFlowAdd(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" dbg.echo hi : slp 1s : f.+ foo", indent) +
			"- create a flow 'foo' by 'f.+' for convenient",
		padR(selfName+" foo", indent) + "- execute flow command 'foo'",
	}
}

func SuggestDesc(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	explain := "- show what 'foo' will do without executing it"
	return []string{
		padR(selfName+" foo :-", indent) + explain,
		padR(selfName+" foo :+", indent) + explain + ", with details",
	}
}

// TODO: no use by now
/*
func SuggestFindConfigFlows(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	tagOutOfTheBox := env.GetRaw("strs.tag-out-of-the-box")
	tagProvider := env.GetRaw("strs.tag-provider")
	prefix := selfName + " " + tagOutOfTheBox + " " + tagProvider + " :"
	explain := "- find configuring flows"
	return []string{
		padR(prefix+"-", indent) + explain,
		padR(prefix+"+", indent) + explain + ", with details",
	}
}
*/

func SuggestFindProvider(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	//prefix := selfName + " key-name write :"
	explain := "- find modules which will write this key"
	return []string{
		padR(selfName+" env.who-write key-name", indent) + explain,
		padR(selfName+" e.ww          key-name", indent) + "- same as above",
		//padR(prefix+"/", indent) + explain,
		//padR(prefix+"//", indent) + explain + ", with usage",
	}
}

func SuggestFlowsFilter(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" f str1 str2", indent) + "- find flows matched these strings",
		padR(selfName+" str1 str2 ::f", indent) + "- same as above, in tail-mode",
	}
}

func SuggestTailInfo(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" cmd :==", indent) + "- show command's detail info",
	}
}

func getSuggestArgs(env *core.Env) (selfName string, explainIndent int) {
	selfName = env.GetRaw("strs.self-name")
	explainIndent = env.GetInt("display.hint.indent.2rd")
	return
}
