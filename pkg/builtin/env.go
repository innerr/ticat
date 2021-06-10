package builtin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/utils"
)

func LoadDefaultEnv(env *core.Env) {
	env = env.GetLayer(core.EnvLayerDefault)

	env.Set("sys.bootstrap", "")
	env.SetInt("sys.stack-depth", 0)

	env.SetBool("sys.step-by-step", false)
	env.SetInt("sys.execute-delay-sec", 0)
	env.SetBool("sys.interact", true)

	env.Set("sys.version", "1.0.0")
	env.Set("sys.dev.name", "marsh")

	env.Set("sys.hub.init-repo", "innerr/marsh.ticat")
	setToDefaultVerb(env)
}

func LoadEnvAbbrs(abbrs *core.EnvAbbrs) {
	sys := abbrs.GetOrAddSub("sys")
	sys.GetOrAddSub("bootstrap").AddAbbrs("boot")
	sys.GetOrAddSub("stack-depth").AddAbbrs("stack")
	sys.GetOrAddSub("interact").AddAbbrs("ir", "i", "I")
	sys.GetOrAddSub("step-by-step").AddAbbrs("step")
	sys.GetOrAddSub("delay-execute").AddAbbrs("delay")
	sys.GetOrAddSub("version").AddAbbrs("ver")

	hub := sys.GetOrAddSub("hub")
	hub.GetOrAddSub("init-repo").AddAbbrs("repo")

	disp := abbrs.GetOrAddSub("display").AddAbbrs("disp", "dis", "di")
	disp.GetOrAddSub("width").AddAbbrs("wid", "w", "W")
	disp.GetOrAddSub("style").AddAbbrs("sty", "s", "S")
	utf8 := disp.GetOrAddSub("utf8")
	utf8.AddAbbrs("utf", "u", "U")
	utf8.GetOrAddSub("symbols").AddAbbrs("symbol", "sym", "s", "S")
	disp.GetOrAddSub("executor").AddAbbrs("exe", "exec")
	disp.GetOrAddSub("bootstrap").AddAbbrs("boot")
	disp.GetOrAddSub("one-cmd").AddAbbrs("one", "1")
	disp.GetOrAddSub("max-cmd-cnt").AddAbbrs("cmds")
	disp.GetOrAddSub("env").AddAbbrs("e", "E")

	env := disp.GetOrAddSub("env")
	env.GetOrAddSub("sys").GetOrAddSub("paths").AddAbbrs("path")
	env.GetOrAddSub("default").AddAbbrs("def")
	env.GetOrAddSub("display").AddAbbrs("disp", "dis", "di")

	mod := disp.GetOrAddSub("mod")
	mod.GetOrAddSub("quiet").AddAbbrs("q", "Q")
	mod.GetOrAddSub("realname").AddAbbrs("real", "r", "R")
}

func LoadRuntimeEnv(_ core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	env = env.GetLayer(core.EnvLayerSession)

	path, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("[LoadRuntimeEnv] get abs self-path fail: %v", err))
	}
	data := path + ".data"

	sys := cc.EnvAbbrs.GetOrAddSub("sys")
	paths := sys.GetOrAddSub("paths").AddAbbrs("path", "p", "P")
	env.Set("sys.paths.hub", filepath.Join(data, "hub"))

	env.Set("sys.paths.ticat", path)
	paths.GetOrAddSub("ticat").AddAbbrs("cat")

	env.Set("sys.paths.data", data)
	paths.GetOrAddSub("data").AddAbbrs("dat")

	env.Set("sys.paths.flows", filepath.Join(data, "flows"))
	paths.GetOrAddSub("flows").AddAbbrs("flow")

	env.Set("sys.paths.sessions", filepath.Join(data, "sessions"))
	paths.GetOrAddSub("sessions").AddAbbrs("session", "s", "S")

	return true
}

func LoadLocalEnv(_ core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	kvSep := env.GetRaw("strs.env-kv-sep")
	path := getEnvLocalFilePath(env)
	core.LoadEnvFromFile(env.GetLayer(core.EnvLayerPersisted), path, kvSep)
	env.GetLayer(core.EnvLayerPersisted).DeleteInSelfLayer("sys.stack-depth")
	env.GetLayer(core.EnvLayerPersisted).Deduplicate()
	env.GetLayer(core.EnvLayerSession).Deduplicate()
	return true
}

func SaveEnvToLocal(_ core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	kvSep := env.GetRaw("strs.env-kv-sep")
	path := getEnvLocalFilePath(env)
	core.SaveEnvToFile(env, path, kvSep)
	return true
}

// TODO: support abbrs for arg 'key'
func RemoveEnvValAndSaveToLocal(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	key := argv.GetRaw("key")
	if len(key) == 0 {
		panic(fmt.Errorf("[RemoveEnvValAndSaveToLocal] arg 'key' is empty"))
	}
	env.DeleteEx(key, core.EnvLayerDefault)

	kvSep := env.GetRaw("strs.env-kv-sep")
	path := getEnvLocalFilePath(env)
	core.SaveEnvToFile(env.GetLayer(core.EnvLayerSession), path, kvSep)
	return true
}

func ResetLocalEnv(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	path := getEnvLocalFilePath(env)
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
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
	env.SetBool("display.meow", false)
	env.SetBool("display.executor", true)
	env.SetBool("display.executor.end", false)
	env.SetBool("display.bootstrap", false)
	env.SetBool("display.one-cmd", false)
	env.Set("display.style", "utf8")
	env.SetBool("display.utf8", true)
	env.SetBool("display.utf8.symbols", true)
	env.SetBool("display.env", true)
	env.SetBool("display.env.sys", false)
	env.SetBool("display.env.sys.paths", false)
	env.SetBool("display.env.layer", false)
	env.SetBool("display.env.default", false)
	env.SetBool("display.mod.quiet", false)
	env.SetBool("display.mod.realname", true)
	env.SetBool("display.env.display", false)

	env.SetInt("display.flow.depth", 6)

	env.SetInt("display.max-cmd-cnt", 7)

	row, col := utils.GetTerminalWidth()
	if col > 100 {
		col = 100
	}
	env.SetInt("display.width", col)
	env.SetInt("display.height", row)
}
