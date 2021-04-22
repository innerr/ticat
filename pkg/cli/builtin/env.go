package builtin

import (
	"fmt"
	"os"
	"path/filepath"

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
	env.Set("runtime.display.mod.realname", "true")
}

func LoadRuntimeEnv(_ cli.ArgVals, _ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerSession)
	path, err := filepath.Abs(os.Args[0])
	if err != nil {
		panic(fmt.Errorf("[LoadRuntimeEnv] get selfpath's abs fail: %v", err))
	}
	env.Set("runtime.sys.paths.ticat", path)
	data := path + ".data"
	env.Set("runtime.sys.paths.data", data)
	env.Set("runtime.sys.paths.mods", filepath.Join(data, "mods"))
	return true
}

func LoadLocalEnv(_ cli.ArgVals, _ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerPersisted)
	return true
}
