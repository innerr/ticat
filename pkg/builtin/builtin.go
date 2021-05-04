package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

func RegisterCmds(cmds *core.CmdTree) {
	RegisterExecutorCmds(cmds)
	RegisterEnvOpCmds(cmds)
	RegisterVerbCmds(cmds)
	RegisterTrivialCmds(cmds)
	RegisterBuiltinCmds(cmds.AddSub("builtin", "b", "B"))
}

func RegisterExecutorCmds(cmds *core.CmdTree) {
	cmds.AddSub("help", "h", "H", "?").
		RegPowerCmd(GlobalHelp, "TODO").SetQuiet()
	cmds.AddSub("desc", "d", "D").
		RegPowerCmd(DbgDumpFlow,
			"desc the flow about to execute").SetQuiet().SetPriority()
	cmds.AddSub("cmds", "cmd", "mod", "mods").
		RegCmd(DbgDumpCmds,
			"list builtin and loaded cmds")
}

func RegisterEnvOpCmds(cmds *core.CmdTree) {
	cmds.AddSub("save", "persist", "s", "S").
		RegCmd(SaveEnvToLocal,
			"save session env changes to local").SetQuiet()
	env := cmds.AddSub("env", "e", "E")
	env.RegCmd(DbgDumpEnv,
		"list all env KVs")
	env.AddSub("abbrs", "a", "A").
		RegCmd(DbgDumpEnvAbbrs,
			"list env tree and abbrs")
	env.AddSub("flatten", "flat", "f", "F").
		RegCmd(DbgDumpEnvFlattenVals,
			"list env values in flatten format")
	env.AddSub("remove-and-save", "remove", "rm", "delete", "del").
		RegCmd(RemoveEnvValAndSaveToLocal,
			"remove specific env KV and save changes to local").
		SetQuiet().AddArg("key", "", "k", "K")
}

func RegisterVerbCmds(cmds *core.CmdTree) {
	verbose := cmds.AddSub("verbose", "verb", "v", "V")
	verbose.RegCmd(SetVerbMode,
		"change into verbose mode").SetQuiet()
	verbose.AddSub("default", "def", "d", "D").
		RegCmd(SetToDefaultVerb,
			"set to default verbose mode").SetQuiet()
	verbose.AddSub("increase", "inc", "v+", "+").
		RegCmd(IncreaseVerb,
			"increase verbose").SetQuiet().
		AddArg("volume", "1", "vol", "v", "V")
	verbose.AddSub("decrease", "dec", "v-", "-").
		RegCmd(DecreaseVerb,
			"decrease verbose").SetQuiet().
		AddArg("volume", "1", "vol", "v", "V")
	cmds.AddSub("quiet", "q", "Q").
		RegCmd(SetQuietMode,
			"change into quiet mode").SetQuiet()
}

func RegisterBuiltinCmds(cmds *core.CmdTree) {
	env := cmds.AddSub("env", "e", "E")
	envLoad := env.AddSub("load", "l", "L")
	envLoad.AddSub("local", "l", "L").
		RegCmd(LoadLocalEnv,
			"load env KVs from local").SetQuiet()
	envLoad.AddSub("runtime", "rt", "r", "R").
		RegCmd(LoadRuntimeEnv,
			"setup runtime env KVs").SetQuiet()
	envLoad.AddSub("stdin", "s", "S").
		RegCmd(LoadStdinEnv,
			"load env KVs from stdin").SetQuiet()
	envLoad.AddSub("abbrs", "abbr", "a", "A").
		RegCmd(LoadEnvAbbrs,
			"setup runtime env abbrs").SetQuiet()
	mod := cmds.AddSub("mod", "mods", "m", "M")
	mod.AddSub("load", "l", "L").AddSub("local", "l", "L").
		RegCmd(LoadLocalMods,
			"load mods from local").SetQuiet()
}

func RegisterTrivialCmds(cmds *core.CmdTree) {
	cmds.AddSub("dummy", "dmy", "dm").
		RegCmd(Dummy,
			"dummy cmd for testing")
	cmds.AddSub("sleep", "slp").
		RegCmd(Sleep,
			"sleep for specific duration").
		AddArg("duration", "1s", "dur", "d", "D")
}
