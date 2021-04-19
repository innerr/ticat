package builtin

import (
	"github.com/pingcap/ticat/pkg/cli"
)

func RegisterBuiltin(cmds *cli.CmdTree) {
	cmds.AddSub("help", "h", "H", "?").RegQuietPowerCmd(GlobalHelp)
	cmds.AddSub("verbose", "verb", "v", "V").RegQuietCmd(SetVerbMode)
	cmds.AddSub("verb+", "v+", "V+").RegQuietCmd(IncreaseVerb)
	cmds.AddSub("verb-", "v-", "V-").RegQuietCmd(DecreaseVerb)
	cmds.AddSub("quiet", "q", "Q").RegQuietCmd(SetQuietMode)

	cmds.AddSub("dummy", "d", "D").RegCmd(Dummy)
	cmds.AddSub("sleep", "slp", "s", "S").RegCmd(Sleep).AddArg("duration", "dur")

	// Nodes without executables, could provide a convenient way to define env values
	runtime := cmds.AddSub("runtime", "rt")
	display := runtime.AddSub("display", "d", "D")
	display.AddSub("env", "e", "E")
	display.AddSub("mod", "mods", "m", "M")

	builtin := cmds.AddSub("builtin", "b", "B")

	env := builtin.AddSub("env")
	envLoad := env.AddSub("load")
	envLoad.AddSub("local", "l", "L").RegCmd(LoadLocalEnv)
	envLoad.AddSub("runtime", "rt", "r", "R").RegCmd(LoadRuntimeEnv)

	mod := builtin.AddSub("mod")
	mod.AddSub("load").AddSub("local", "l", "L").RegCmd(LoadLocalMods)
}