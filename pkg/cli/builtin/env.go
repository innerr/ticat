package builtin

import (
	"os"

	"github.com/pingcap/ticat/pkg/cli"
)

func LoadBuiltinEnv(env *cli.Env) {
	env = env.GetLayer(cli.EnvLayerDefault)
	env.Set("runtime.version", "dev")
	env.Set("runtime.stack-depth", "0")
	env.Set("runtime.display", "true")
	env.Set("runtime.display.width", "80")
	env.Set("runtime.display.one-cmd", "false")
	env.Set("runtime.display.max-cmd-cnt", "7")
	env.Set("runtime.display.env", "true")
	env.Set("runtime.display.env.layer", "false")
	env.Set("runtime.display.env.default", "false")
	env.Set("runtime.display.env.runtime", "false")
	env.Set("runtime.display.mod.builtin", "false")
	env.Set("runtime.display.mod.realname", "false")
}

func LoadRuntimeEnv(_ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerSession)
	env.Set("runtime.sys.ticap-path", os.Args[0])
	return true
}

func LoadLocalEnv(_ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerPersisted)
	return true
}
