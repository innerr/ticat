package builtin

import (
	"github.com/pingcap/ticat/pkg/cli"
)

func RegisterBuiltin(cmds *cli.CmdTree) {
	cmds.AddSub("help", "h", "H", "?").RegQuietPowerCmd(GlobalHelp)
	cmds.AddSub("verbose", "verb", "v", "V").RegQuietCmd(SetVerbMode)
	cmds.AddSub("verb+", "v+", "V+").RegQuietCmd(IncreaseVerb).AddArg("volume", "1", "vol", "v", "V")
	cmds.AddSub("verb-", "v-", "V-").RegQuietCmd(DecreaseVerb).AddArg("volume", "1", "vol", "v", "V")
	cmds.AddSub("quiet", "q", "Q").RegQuietCmd(SetQuietMode)

	cmds.AddSub("desc").RegQuietPowerCmd(DbgDumpCmds)
	cmds.AddSub("dummy", "d", "D").RegCmd(Dummy)
	cmds.AddSub("sleep", "slp", "s", "S").RegCmd(Sleep).AddArg("duration", "1s", "dur", "d", "D")

	builtin := cmds.AddSub("builtin", "b", "B")

	env := builtin.AddSub("env", "E")
	envLoad := env.AddSub("load", "L")
	envLoad.AddSub("local", "l", "L").RegQuietCmd(LoadLocalEnv)
	envLoad.AddSub("runtime", "rt", "r", "R").RegQuietCmd(LoadRuntimeEnv)

	mod := builtin.AddSub("mod", "mods", "M")
	mod.AddSub("load", "L").AddSub("local", "l", "L").RegQuietCmd(LoadLocalMods)

	dbg := cmds.AddSub("dump")
	dbg.AddSub("cmd", "cmds").RegQuietPowerCmd(DbgDumpCmds)
	dbg.AddSub("env").RegCmd(DbgDumpEnv)
	dbg.AddSub("mod", "mods").RegCmd(DbgDumpMods).AddArg("alias", "off", "a", "A")

	// Nodes without executables, could provide a convenient way to define env values
	runtime := cmds.AddSub("runtime", "rt")
	display := runtime.AddSub("display", "d", "D")
	display.AddSub("env", "e", "E")
	display.AddSub("mod", "mods", "m", "M")
}
