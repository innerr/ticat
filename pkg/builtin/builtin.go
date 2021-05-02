package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

func RegisterBuiltinCmds(cmds *core.CmdTree) {
	cmds.AddSub("help", "h", "H", "?").
		RegPowerCmd(GlobalHelp, "TODO").SetQuiet()
	cmds.AddSub("cmds", "cmd", "mod", "mods").
		RegCmd(DbgDumpCmds,
			"list builtin and loaded cmds")
	cmds.AddSub("desc", "d", "D").
		RegPowerCmd(DbgDumpFlow,
			"desc the flow about to execute").SetQuiet()
	cmds.AddSub("save", "persist", "s", "S").
		RegCmd(SaveEnvToLocal,
			"save session env changes to local").SetQuiet()
	env := cmds.AddSub("env", "e", "E")
	env.RegCmd(DbgDumpEnv, "list all env KVs")
	env.AddSub("remove-and-save", "remove", "rm", "delete", "del").
		RegCmd(RemoveEnvValAndSaveToLocal,
			"remove specific env KV and save changes to local").
		SetQuiet().AddArg("key", "", "k", "K")

	cmds.AddSub("verbose", "verb", "v", "V").
		RegCmd(SetVerbMode,
			"change into verbose mode").SetQuiet()
	cmds.AddSub("verb+", "v+").
		RegCmd(IncreaseVerb,
			"increase verbose").SetQuiet().
		AddArg("volume", "1", "vol", "v", "V")
	cmds.AddSub("verb-", "v-").
		RegCmd(DecreaseVerb,
			"decrease verbose").SetQuiet().
		AddArg("volume", "1", "vol", "v", "V")
	cmds.AddSub("quiet", "q", "Q").
		RegCmd(SetQuietMode,
			"change into quiet mode").SetQuiet()

	cmds.AddSub("dummy", "dmy", "dm").
		RegCmd(Dummy,
			"dummy cmd for testing")
	cmds.AddSub("sleep", "slp").
		RegCmd(Sleep,
			"sleep for specific duration").
		AddArg("duration", "1s", "dur", "d", "D")

	builtin := cmds.AddSub("builtin", "b", "B")

	builtinEnv := builtin.AddSub("env", "e", "E")
	envLoad := builtinEnv.AddSub("load", "l", "L")
	envLoad.AddSub("local", "l", "L").
		RegCmd(LoadLocalEnv,
			"load env KVs from local").SetQuiet()
	envLoad.AddSub("runtime", "rt", "r", "R").
		RegCmd(LoadRuntimeEnv,
			"setup runtime env KVs").SetQuiet()

	mod := builtin.AddSub("mod", "mods", "m", "M")
	mod.AddSub("load", "L").AddSub("local", "l", "L").
		RegCmd(LoadLocalMods,
			"load mods from local").SetQuiet()
}
