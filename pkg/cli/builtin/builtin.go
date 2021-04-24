package builtin

import (
	"github.com/pingcap/ticat/pkg/cli"
)

func RegisterBuiltinMods(cmds *cli.CmdTree) {
	cmds.AddSub("help", "h", "H", "?").RegQuietPowerCmd(GlobalHelp)
	cmds.AddSub("mod", "mods").RegCmd(DbgDumpMods).AddArg("alias", "off", "a", "A")
	cmds.AddSub("desc", "d", "D").RegQuietPowerCmd(DbgDumpCmds)
	cmds.AddSub("save", "persist", "s", "S").RegQuietCmd(SaveEnvToLocal)
	env := cmds.AddSub("env", "e", "E")
	env.RegCmd(DbgDumpEnv)
	env.AddSub("remove-and-save", "remove", "rm", "delete", "del").
		RegQuietCmd(RemoveEnvValAndSaveToLocal).AddArg("key", "", "k", "K")

	cmds.AddSub("verbose", "verb", "v", "V").RegQuietCmd(SetVerbMode)
	cmds.AddSub("verb+", "v+").RegQuietCmd(IncreaseVerb).AddArg("volume", "1", "vol", "v", "V")
	cmds.AddSub("verb-", "v-").RegQuietCmd(DecreaseVerb).AddArg("volume", "1", "vol", "v", "V")
	cmds.AddSub("quiet", "q", "Q").RegQuietCmd(SetQuietMode)

	cmds.AddSub("dummy", "dmy", "dm").RegCmd(Dummy)
	cmds.AddSub("sleep", "slp").RegCmd(Sleep).AddArg("duration", "1s", "dur", "d", "D")

	builtin := cmds.AddSub("builtin", "b", "B")

	builtinEnv := builtin.AddSub("env", "E")
	envLoad := builtinEnv.AddSub("load", "L")
	envLoad.AddSub("local", "l", "L").RegQuietCmd(LoadLocalEnv)
	envLoad.AddSub("runtime", "rt", "r", "R").RegQuietCmd(LoadRuntimeEnv)

	mod := builtin.AddSub("mod", "mods", "M")
	mod.AddSub("load", "L").AddSub("local", "l", "L").RegQuietCmd(LoadLocalMods)
}
