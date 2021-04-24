package builtin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pingcap/ticat/pkg/cli"
)

func LoadDefaultEnv(env *cli.Env) {
	env = env.GetLayer(cli.EnvLayerDefault)
	env.Set("sys.stack-depth", "0")
	env.Set("sys.version", "dev")
	env.Set("display", "true")
	env.Set("display.bootstrap", "false")
	env.Set("display.width", "80")
	env.Set("display.one-cmd", "false")
	env.Set("display.max-cmd-cnt", "7")
	env.Set("display.env", "true")
	env.Set("display.env.layer", "false")
	env.Set("display.env.default", "false")
	env.Set("display.mod.quiet", "false")
	env.Set("display.mod.realname", "true")
}

func LoadRuntimeEnv(_ cli.ArgVals, _ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerSession)
	path, err := filepath.Abs(os.Args[0])
	if err != nil {
		panic(fmt.Errorf("[LoadRuntimeEnv] get selfpath's abs fail: %v", err))
	}
	env.Set("sys.paths.ticat", path)
	data := path + ".data"
	env.Set("sys.paths.data", data)
	env.Set("sys.paths.mods", filepath.Join(data, "mods"))
	return true
}

func LoadLocalEnv(_ cli.ArgVals, _ *cli.Cli, env *cli.Env) bool {
	path := filepath.Join(env.Get("sys.paths.data").Raw, "bootstrap.env")
	file, err := os.Open(path)
	if err != nil && !os.IsNotExist(err) {
		panic(fmt.Errorf("[LoadLocalEnv] open local env file '%s' failed: %v", path, err))
	}
	defer file.Close()

	rest, err := cli.EnvInput(env.GetLayer(cli.EnvLayerPersisted), file)
	if err != nil {
		panic(fmt.Errorf("[LoadLocalEnv] read local env file '%s' failed: %v", path, err))
	}
	if len(rest) != 0 {
		panic(fmt.Errorf("[LoadLocalEnv] env file '%s': lines cant' be parsed '%v'", path, rest))
	}
	env.GetLayer(cli.EnvLayerPersisted).DeleteSelf("sys.stack-depth")
	env.GetLayer(cli.EnvLayerSession).Deduplicate()
	return true
}

func SaveEnvToLocal(_ cli.ArgVals, cc *cli.Cli, env *cli.Env) bool {
	path := env.Get("sys.paths.data").Raw
	if len(path) == 0 {
		panic(fmt.Errorf("[SaveEnvToLocal] can't find local data path"))
	}
	path = filepath.Join(path, "bootstrap.env")
	tmp := path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(fmt.Errorf("[SaveEnvToLocal] open local env file '%s' failed: %v", tmp, err))
	}
	defer file.Close()
	err = cli.EnvOutput(env, file)
	if err != nil {
		panic(fmt.Errorf("[SaveEnvToLocal] write local env file '%s' failed: %v", tmp, err))
	}
	err = os.Rename(tmp, path)
	if err != nil {
		panic(fmt.Errorf("[SaveEnvToLocal] rename env file '%s' to '%s' failed: %v", tmp, path, err))
	}
	return true
}

func RemoveEnvValAndSaveToLocal(argv cli.ArgVals, cc *cli.Cli, env *cli.Env) bool {
	key := argv.GetRaw("key")
	if len(key) == 0 {
		panic(fmt.Errorf("[RemoveEnvValAndSaveToLocal] arg 'key' is empty"))
	}
	env.DeleteExt(key, cli.EnvLayerDefault)
	return true
}
