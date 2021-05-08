package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

func RegisterCmds(cmds *core.CmdTree) {
	RegisterExecutorCmds(cmds)
	RegisterEnvCmds(cmds)
	RegisterVerbCmds(cmds)
	RegisterTrivialCmds(cmds)
	RegisterFlowCmds(cmds)
	RegisterBuiltinCmds(cmds.AddSub("builtin", "b", "B"))
}

func RegisterExecutorCmds(cmds *core.CmdTree) {
	cmds.AddSub("help", "h", "H", "?").
		RegPowerCmd(GlobalHelp,
			"get help").
		SetQuiet().
		SetPriority().
		AddArg("1st-str", "", "1", "find", "str", "s", "S").
		AddArg("2rd-str", "", "2").
		AddArg("3th-str", "", "3")
	cmds.AddSub("search", "find", "fnd", "s", "S").
		RegCmd(FindAny,
			"find anything with given string").
		AddArg("1st-str", "", "1", "find", "str", "s", "S").
		AddArg("2rd-str", "", "2").
		AddArg("3th-str", "", "3")
	cmds.AddSub("desc", "d", "D").
		RegPowerCmd(DbgDumpFlow,
			"desc the flow about to execute").
		SetQuiet().
		SetPriority()
	mod := cmds.AddSub("cmds", "cmd", "mod", "mods", "m", "M", "c", "C")
	mod.AddSub("tree", "t", "T").
		RegCmd(DbgDumpCmdTree,
			"list builtin and loaded cmds")
	mod.AddSub("list", "ls", "l", "flatten", "flat", "f", "F").
		RegCmd(DbgDumpCmds,
			"list builtin and loaded cmds").
		AddArg("1st-str", "", "1", "find", "str", "s", "S").
		AddArg("2rd-str", "", "2").
		AddArg("3th-str", "", "3")
}

func RegisterFlowCmds(cmds *core.CmdTree) {
	flow := cmds.AddSub("flow", "fl", "f", "F")
	flow.AddSub("save", "persist", "s", "S").
		RegPowerCmd(SaveFlow,
			"save current cmds as a flow").
		SetQuiet().
		SetPriority().
		AddArg("to-cmd-path", "", "path", "p", "P")
	flow.AddSub("remove", "rm", "delete", "del").
		RegCmd(RemoveFlow,
			"remove a saved flow").
		AddArg("cmd-path", "", "path", "p", "P")
}

func RegisterEnvCmds(cmds *core.CmdTree) {
	env := cmds.AddSub("env", "e", "E")
	env.AddSub("tree", "t", "T").
		RegCmd(DbgDumpEnv,
			"list all env layers and KVs in tree format")
	env.AddSub("abbrs", "abbr", "a", "A").
		RegCmd(DbgDumpEnvAbbrs,
			"list env tree and abbrs")
	env.AddSub("list", "ls", "flatten", "flat", "f", "F").
		RegCmd(DbgDumpEnvFlattenVals,
			"list env values in flatten format").
		AddArg("1st-str", "", "1", "find", "str", "s", "S").
		AddArg("2rd-str", "", "2").
		AddArg("3th-str", "", "3")

	env.AddSub("save", "persist", "s", "S").
		RegCmd(SaveEnvToLocal,
			"save session env changes to local").
		SetQuiet()
	env.AddSub("remove-and-save", "remove", "rm", "delete", "del").
		RegCmd(RemoveEnvValAndSaveToLocal,
			"remove specific env KV and save changes to local").
		AddArg("key", "", "k", "K")
	env.AddSub("reset-and-save", "reset").
		RegCmd(ResetLocalEnv,
			"reset all local saved env KVs")
}

func RegisterVerbCmds(cmds *core.CmdTree) {
	cmds.AddSub("quiet", "q", "Q").
		RegCmd(SetQuietMode,
			"change into quiet mode").
		SetQuiet()
	verbose := cmds.AddSub("verbose", "verb", "v", "V")

	verbose.RegCmd(SetVerbMode,
		"change into verbose mode").
		SetQuiet()
	verbose.AddSub("default", "def", "d", "D").
		RegCmd(SetToDefaultVerb,
			"set to default verbose mode").
		SetQuiet()
	verbose.AddSub("increase", "inc", "v+", "+").
		RegCmd(IncreaseVerb,
			"increase verbose").
		SetQuiet().
		AddArg("volume", "1", "vol", "v", "V")
	verbose.AddSub("decrease", "dec", "v-", "-").
		RegCmd(DecreaseVerb,
			"decrease verbose").
		SetQuiet().
		AddArg("volume", "1", "vol", "v", "V")
}

func RegisterBuiltinCmds(cmds *core.CmdTree) {
	env := cmds.AddSub("env", "e", "E")
	envLoad := env.AddSub("load", "l", "L")
	envLoad.AddSub("local", "l", "L").
		RegCmd(LoadLocalEnv,
			"load env KVs from local").
		SetQuiet()
	envLoad.AddSub("runtime", "rt", "r", "R").
		RegCmd(LoadRuntimeEnv,
			"setup runtime env KVs").
		SetQuiet()
	envLoad.AddSub("stdin", "s", "S").
		RegCmd(LoadStdinEnv,
			"load env KVs from stdin").
		SetQuiet()
	envLoad.AddSub("abbrs", "abbr", "a", "A").
		RegCmd(LoadEnvAbbrs,
			"setup runtime env abbrs").
		SetQuiet()

	mod := cmds.AddSub("mod", "mods", "m", "M")
	modLoad := mod.AddSub("load", "l", "L")
	modLoad.AddSub("local", "l", "L").
		RegCmd(LoadLocalMods,
			"load mods from local").
		SetQuiet()
	modLoad.AddSub("flows", "flows", "f", "F").
		RegCmd(LoadLocalFlows,
			"load flows from local")
	modLoad.AddSub("ext-exec", "ext", "e", "E").
		RegCmd(SetExtExec,
			"load default setting of how to run a executable file by ext name")
}

func RegisterTrivialCmds(cmds *core.CmdTree) {
	dummy := cmds.AddSub("dummy", "dmy", "dm")
	dummy.RegCmd(Dummy,
		"dummy cmd for testing")
	dummy.AddSub("quiet", "q", "Q").
		RegCmd(QuietDummy,
			"quiet dummy cmd for testing").
		SetQuiet()
	dummy.AddSub("power", "p", "P").
		RegPowerCmd(PowerDummy,
			"power dummy cmd for testing")
	dummy.AddSub("priority", "prior", "prio", "pri").
		RegPowerCmd(PriorityPowerDummy,
			"power dummy cmd for testing").
		SetPriority()
	cmds.AddSub("sleep", "slp").
		RegCmd(Sleep,
			"sleep for specific duration").
		AddArg("duration", "1s", "dur", "d", "D")
}
