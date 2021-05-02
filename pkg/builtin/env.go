package builtin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func LoadDefaultEnv(env *core.Env) {
	env = env.GetLayer(core.EnvLayerDefault)

	env.Set("sys.version", "1.0.0")
	env.Set("sys.dev.name", "kitty")

	env.SetInt("sys.stack-depth", 0)

	env.SetBool("display.executor", true)
	env.SetBool("display.bootstrap", false)
	env.SetBool("display.one-cmd", false)
	env.SetBool("display.env", true)
	env.SetBool("display.env.sys", false)
	env.SetBool("display.env.layer", false)
	env.SetBool("display.env.default", false)
	env.SetBool("display.mod.quiet", false)
	env.SetBool("display.mod.realname", true)

	env.SetInt("display.width", 80)
	env.SetInt("display.max-cmd-cnt", 7)
}

func LoadRuntimeEnv(_ core.ArgVals, _ *core.Cli, env *core.Env) bool {
	env = env.GetLayer(core.EnvLayerSession)

	path, err := filepath.Abs(os.Args[0])
	if err != nil {
		panic(fmt.Errorf("[LoadRuntimeEnv] get abs self-path '%s' fail: %v",
			os.Args[0], err))
	}

	env.Set("sys.paths.ticat", path)
	data := path + ".data"
	env.Set("sys.paths.data", data)
	env.Set("sys.paths.mods", filepath.Join(data, "mods"))
	return true
}

func LoadLocalEnv(_ core.ArgVals, _ *core.Cli, env *core.Env) bool {
	protoEnvMark := env.Get("strs.proto-env-mark").Raw
	protoSep := env.Get("strs.proto-sep").Raw
	path := getEnvLocalFilePath(env)
	file, err := os.Open(path)
	if err != nil && !os.IsNotExist(err) {
		panic(fmt.Errorf("[LoadLocalEnv] open local env file '%s' failed: %v",
			path, err))
	}
	defer file.Close()

	rest, err := core.EnvInput(env.GetLayer(core.EnvLayerPersisted),
		file, protoEnvMark, protoSep)
	if err != nil {
		panic(fmt.Errorf("[LoadLocalEnv] read local env file '%s' failed: %v",
			path, err))
	}
	if len(rest) != 0 {
		panic(fmt.Errorf("[LoadLocalEnv] env file '%s': lines cant' be parsed '%v'",
			path, rest))
	}
	env.GetLayer(core.EnvLayerPersisted).DeleteSelf("sys.stack-depth")
	env.GetLayer(core.EnvLayerSession).Deduplicate()
	return true
}

func SaveEnvToLocal(_ core.ArgVals, cc *core.Cli, env *core.Env) bool {
	protoEnvMark := env.Get("strs.proto-env-mark").Raw
	protoSep := env.Get("strs.proto-sep").Raw

	path := getEnvLocalFilePath(env)
	tmp := path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(fmt.Errorf("[SaveEnvToLocal] open local env file '%s' failed: %v", tmp, err))
	}
	defer file.Close()

	err = core.EnvOutput(env, file, protoEnvMark, protoSep)
	if err != nil {
		panic(fmt.Errorf("[SaveEnvToLocal] write local env file '%s' failed: %v", tmp, err))
	}
	err = os.Rename(tmp, path)
	if err != nil {
		panic(fmt.Errorf("[SaveEnvToLocal] rename env file '%s' to '%s' failed: %v",
			tmp, path, err))
	}
	return true
}

func RemoveEnvValAndSaveToLocal(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	key := argv.GetRaw("key")
	if len(key) == 0 {
		panic(fmt.Errorf("[RemoveEnvValAndSaveToLocal] arg 'key' is empty"))
	}
	env.DeleteEx(key, core.EnvLayerDefault)
	return true
}

func getEnvLocalFilePath(env *core.Env) string {
	path := env.Get("sys.paths.data").Raw
	file := env.Get("strs.env-file-name").Raw
	if len(path) == 0 || len(file) == 0 {
		panic(fmt.Errorf("[getEnvLocalFilePath] can't find local data path"))
	}
	return filepath.Join(path, file)
}
