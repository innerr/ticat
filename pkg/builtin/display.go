package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

func DisplayUtf8On(_ core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	env = env.GetLayer(core.EnvLayerSession)
	env.SetBool("display.utf8", true)
	env.SetBool("display.utf8.symbols", true)
	return true
}

func DisplayUtf8Off(_ core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	env = env.GetLayer(core.EnvLayerSession)
	env.SetBool("display.utf8", false)
	env.SetBool("display.utf8.symbols", false)
	return true
}
