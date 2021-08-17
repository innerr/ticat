package builtin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
	"github.com/pingcap/ticat/pkg/utils"
)

func LoadDefaultEnv(env *core.Env) {
	env = env.GetLayer(core.EnvLayerDefault)

	env.Set("sys.bootstrap", "")
	env.SetInt("sys.stack-depth", 0)

	env.SetBool("sys.step-by-step", false)
	env.SetBool("sys.panic.recover", true)
	env.SetInt("sys.execute-delay-sec", 0)
	env.SetBool("sys.interact", true)

	env.Set("sys.version", "1.0.0")
	env.Set("sys.dev.name", "marsh")

	env.SetBool("sys.env.use-cmd-abbrs", false)

	env.Set("sys.hub.init-repo", "innerr/marsh.ticat")

	row, col := utils.GetTerminalWidth()
	if col > 100 {
		col = 100
	}
	env.SetInt("display.width", col)
	env.SetInt("display.height", row)

	env.Set("display.example-https-repo", "https://github.com/innerr/tidb.ticat")

	env.SetInt("display.hint.indent.2rd", 38)

	env.Set("display.utf8.symbols.tip", "💡 ")
	env.SetInt("display.utf8.symbols.tip.len", 3)
	env.Set("display.utf8.symbols.err", "⛔ ")
	env.SetInt("display.utf8.symbols.err.len", 3)

	setToDefaultVerb(env)
}

func LoadEnvAbbrs(abbrs *core.EnvAbbrs) {
	sys := abbrs.GetOrAddSub("sys")
	sys.GetOrAddSub("bootstrap").AddAbbrs("boot")
	sys.GetOrAddSub("interact").AddAbbrs("ir", "i", "I")
	sys.GetOrAddSub("step-by-step").AddAbbrs("step")
	sys.GetOrAddSub("delay-execute").AddAbbrs("delay")
	sys.GetOrAddSub("version").AddAbbrs("ver")

	hub := sys.GetOrAddSub("hub")
	hub.GetOrAddSub("init-repo").AddAbbrs("repo")

	disp := abbrs.GetOrAddSub("display").AddAbbrs("disp", "dis", "di")
	disp.GetOrAddSub("width").AddAbbrs("wid", "w", "W")
	disp.GetOrAddSub("style").AddAbbrs("sty", "s", "S")
	disp.GetOrAddSub("color").AddAbbrs("colour", "clr")
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

func LoadRuntimeEnv(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	env = env.GetLayer(core.EnvLayerSession)

	path, err := os.Executable()
	if err != nil {
		panic(core.NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("get abs self-path fail: %v", err)))
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

	return currCmdIdx, true
}

func LoadLocalEnv(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	kvSep := env.GetRaw("strs.env-kv-sep")
	path := getEnvLocalFilePath(env, flow.Cmds[currCmdIdx])
	core.LoadEnvFromFile(env.GetLayer(core.EnvLayerPersisted), path, kvSep)
	env.GetLayer(core.EnvLayerPersisted).DeleteInSelfLayer("sys.stack-depth")
	env.GetLayer(core.EnvLayerPersisted).Deduplicate()
	env.GetLayer(core.EnvLayerSession).Deduplicate()

	if !env.Has("display.color") {
		env.GetLayer(core.EnvLayerSession).SetBool("display.color", !utils.StdoutIsPipe())
	}

	return currCmdIdx, true
}

func SaveEnvToLocal(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	kvSep := env.GetRaw("strs.env-kv-sep")
	path := getEnvLocalFilePath(env, flow.Cmds[currCmdIdx])
	core.SaveEnvToFile(env, path, kvSep)
	display.PrintTipTitle(cc.Screen, env,
		"changes of env are saved, could be listed by:",
		"",
		display.SuggestListEnv(env))
	return currCmdIdx, true
}

// TODO: support abbrs for arg 'key'
func RemoveEnvValAndSaveToLocal(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	tailMode := flow.TailModeCall && flow.Cmds[currCmdIdx].TailMode

	var keys []string
	if tailMode {
		keys = append(keys, tailModeGetInput(flow, currCmdIdx, false)...)
	}
	key := argv.GetRaw("key")
	if len(key) != 0 {
		keys = append(keys, key)
	} else if !tailMode {
		panic(core.NewCmdError(flow.Cmds[currCmdIdx], "arg 'key' is empty"))
	}
	if len(keys) == 0 {
		panic(core.NewCmdError(flow.Cmds[currCmdIdx], "no specified key to remove"))
	}

	deleted := 0
	for _, key := range keys {
		keyStr := display.ColorKey("'"+key+"'", env)
		if env.Has(key) {
			env.DeleteEx(key, core.EnvLayerDefault)
			deleted += 1
			cc.Screen.Print(keyStr + " deleted\n")
		} else {
			cc.Screen.Print(keyStr + " not exist, skipped deleting\n")
		}
	}

	if deleted != 0 {
		kvSep := env.GetRaw("strs.env-kv-sep")
		path := getEnvLocalFilePath(env, flow.Cmds[currCmdIdx])
		core.SaveEnvToFile(env.GetLayer(core.EnvLayerSession), path, kvSep)
		display.PrintTipTitle(cc.Screen, env, "changes of env are saved")
	}
	return currCmdIdx, true
}

func ResetLocalEnv(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	path := getEnvLocalFilePath(env, flow.Cmds[currCmdIdx])
	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			display.PrintTipTitle(cc.Screen, env, "there is no saved env changes, nothing to do")
			return currCmdIdx, true
		} else {
			panic(core.NewCmdError(flow.Cmds[currCmdIdx],
				fmt.Sprintf("remove env file '%s' failed: %v", path, err)))
		}
	}
	display.PrintTipTitle(cc.Screen, env, "all saved env changes are removed")
	return currCmdIdx, true
}

func ResetSessionEnv(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	env.GetLayer(core.EnvLayerSession).Clear(false)
	display.PrintTipTitle(cc.Screen, env, "all session env values are removed")
	return currCmdIdx, true
}

func getEnvLocalFilePath(env *core.Env, cmd core.ParsedCmd) string {
	path := env.GetRaw("sys.paths.data")
	file := env.GetRaw("strs.env-file-name")
	if len(path) == 0 || len(file) == 0 {
		panic(core.NewCmdError(cmd, fmt.Sprintf("can't find local data path")))
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
	env.SetBool("display.stack", true)
	env.SetBool("display.env", true)
	env.SetBool("display.env.sys", false)
	env.SetBool("display.env.sys.paths", false)
	env.SetBool("display.env.layer", false)
	env.SetBool("display.env.default", false)
	env.SetBool("display.mod.quiet", false)
	env.SetBool("display.mod.realname", true)
	env.SetBool("display.env.display", false)
	env.SetInt("display.flow.depth", 6)
	env.SetInt("display.max-cmd-cnt", 14)
}
