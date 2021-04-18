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
	cmds.AddSub("sleep", "slp", "s", "S").RegCmd(Sleep)

	builtin := cmds.AddSub("builtin", "b", "B")

	env := builtin.AddSub("env")
	envLoad := env.AddSub("load")
	envLoad.AddSub("local").RegCmd(LoadLocalEnv)
	envLoad.AddSub("runtime").RegCmd(LoadRuntimeEnv)

	mod := builtin.AddSub("mod")
	mod.AddSub("load").AddSub("local").RegCmd(LoadLocalMods)
}
