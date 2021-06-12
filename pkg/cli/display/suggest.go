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
	}
}

func SuggestListCmds(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" -", indent) + "- list all commands",
		padR(selfName+" +", indent) + "- list all commands with details",
	}
}

func SuggestFindCmds(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" str1 str2 :-", indent) + "- search commands",
		padR(selfName+" str1 str2 :+", indent) + "- search commands with details",
	}
}

func SuggestFindRepoTag(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	prefix := selfName + " @ready repo-name str1 :"
	explain := "- search commands in repo"
	return []string{
		padR(prefix+"-", indent) + explain,
		padR(prefix+"+", indent) + explain + ", with details",
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
		padR(selfName+" h.init", indent) + "- add a default git repo.",
		padR(selfName+" h.+ innerr/tidb."+selfName, indent) + "- add a git repo,",
		padR("", indent+2) + "could use https address.",
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
		padR(selfName+" h :-", indent) + explain,
		padR(selfName+" h :+", indent) + explain + ", with details",
	}
}

func SuggestFlowAdd(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	return []string{
		padR(selfName+" dbg.echo hi : slp 1s : f.+ xx", indent) +
			"- create a flow 'xx' by 'f.+' for convenient",
		padR(selfName+" xx", indent) + "- execute command 'xx'",
	}
}

func SuggestDesc(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	explain := "- show what 'xx' will do without executing it"
	return []string{
		padR(selfName+" xx :-", indent) + explain,
		padR(selfName+" xx :+", indent) + explain + ", with details",
	}
}

func SuggestFindConfigFlows(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	tagOutOfTheBox := env.GetRaw("strs.tag-out-of-the-box")
	tagProvider := env.GetRaw("strs.tag-provider")
	prefix := selfName + " " + tagOutOfTheBox + " " + tagProvider + " :"
	explain := "- find configuring flows"
	return []string{
		padR(prefix+"-", indent) + explain,
		padR(prefix+"-", indent) + explain + ", with details",
	}
}

func SuggestFindProvider(env *core.Env) []string {
	selfName, indent := getSuggestArgs(env)
	prefix := selfName + " key-name write :"
	explain := "- find modules will write this key"
	return []string{
		padR(prefix+"-", indent) + explain,
		padR(prefix+"+", indent) + explain + ", with details",
	}
}

func getSuggestArgs(env *core.Env) (selfName string, explainIndent int) {
	selfName = env.GetRaw("strs.self-name")
	explainIndent = env.GetInt("display.hint.indent.2rd")
	return
}
