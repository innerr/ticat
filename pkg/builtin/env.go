package builtin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func LoadDefaultEnv(env *core.Env) {
	env = env.GetLayer(core.EnvLayerDefault)
	env.Set("sys.bootstrap", "")
	env.Set("sys.version", "1.0.0")
	env.Set("sys.dev.name", "marsh")
	env.Set("sys.hub.init-repo", "innerr/marsh.ticat")
	env.SetInt("sys.stack-depth", 0)
	setToDefaultVerb(env)
}

func LoadEnvAbbrs(_ core.ArgVals, cc *core.Cli, env *core.Env) bool {
	display := cc.EnvAbbrs.GetOrAddSub("display").AddAbbrs("dis", "disp")
	display.GetOrAddSub("width").AddAbbrs("wid", "w", "W")
	display.GetOrAddSub("style").AddAbbrs("sty", "s", "S")
	return true
}

func LoadRuntimeEnv(_ core.ArgVals, _ *core.Cli, env *core.Env) bool {
	env = env.GetLayer(core.EnvLayerSession)

	path, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("[LoadRuntimeEnv] get abs self-path fail: %v", err))
	}

	env.SetIfEmpty("session", "")
	env.Set("sys.paths.ticat", path)
	data := path + ".data"
	env.Set("sys.paths.data", data)
	env.Set("sys.paths.hub", filepath.Join(data, "hub"))
	env.Set("sys.paths.flows", filepath.Join(data, "flows"))
	env.Set("sys.paths.sessions", filepath.Join(data, "sessions"))
	return true
}

/* TODO: remove
// Interacting methods between ticat and mods:
//   1. mod.stdin(as mod's input args) -> mod.stderr(as mods's return)
//   2. (recursively) calling ticat inside a mod -> ticat.stdin(pass the env from mod to ticat)
//
// The stdin-env could be very useful for customized mods-loader or env-loader
//   1. those loaders will be loaded from 'bootstrap' string above
//   2. put a string val with key 'bootstrap' to env could launch it as an extra bootstrap
func LoadStdinEnv(_ core.ArgVals, _ *core.Cli, env *core.Env) bool {
	kvSep := env.GetRaw("strs.env-kv-sep")
	stdinEnv := genEnvFromStdin(kvSep)
	if stdinEnv != nil {
		env.GetLayer(core.EnvLayerSession).Merge(stdinEnv)
	}
	return true
}
*/

func LoadLocalEnv(_ core.ArgVals, _ *core.Cli, env *core.Env) bool {
	kvSep := env.GetRaw("strs.env-kv-sep")
	path := getEnvLocalFilePath(env)
	core.LoadEnvFromFile(env.GetLayer(core.EnvLayerPersisted), path, kvSep)
	env.GetLayer(core.EnvLayerPersisted).DeleteSelf("sys.stack-depth")
	env.GetLayer(core.EnvLayerSession).Deduplicate()
	return true
}

func SaveEnvToLocal(_ core.ArgVals, cc *core.Cli, env *core.Env) bool {
	kvSep := env.GetRaw("strs.env-kv-sep")
	path := getEnvLocalFilePath(env)
	core.SaveEnvToFile(env, path, kvSep)
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

func ResetLocalEnv(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	path := getEnvLocalFilePath(env)
	err := os.Remove(path)
	if err != nil {
		panic(fmt.Errorf("[ResetLocalEnv] remove env file '%s' failed: %v", path, err))
	}
	return true
}

func getEnvLocalFilePath(env *core.Env) string {
	path := env.GetRaw("sys.paths.data")
	file := env.GetRaw("strs.env-file-name")
	if len(path) == 0 || len(file) == 0 {
		panic(fmt.Errorf("[getEnvLocalFilePath] can't find local data path"))
	}
	return filepath.Join(path, file)
}

func setToDefaultVerb(env *core.Env) {
	env.SetBool("display.executor", true)
	env.SetBool("display.executor.end", false)
	env.SetBool("display.bootstrap", false)
	env.SetBool("display.one-cmd", false)
	env.Set("display.style", "utf8")
	env.SetBool("display.utf8", false)
	env.SetBool("display.env", true)
	env.SetBool("display.env.sys", false)
	env.SetBool("display.env.layer", false)
	env.SetBool("display.env.default", false)
	env.SetBool("display.mod.quiet", false)
	env.SetBool("display.mod.realname", true)

	env.SetInt("display.width", 80)
	env.SetInt("display.max-cmd-cnt", 7)
}

/* TODO: remove
func genEnvFromStdin(kvSep string) *core.Env {
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(fmt.Errorf("[GenEnvFromStdin] get stdin stat failed %v", err))
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil
	}
	env := core.NewEnv()
	rest, err := core.EnvInput(env, os.Stdin, kvSep)
	if err != nil {
		panic(fmt.Errorf("[GenEnvFromStdin] parse stdin failed %v", err))
	}
	if len(rest) != 0 {
		panic(fmt.Errorf("[GenEnvFromStdin] lines cant' be parsed '%v'", rest))
	}
	return env
}
*/
