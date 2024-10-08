package builtin

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/innerr/ticat/pkg/cli/display"
	"github.com/innerr/ticat/pkg/core/model"
	"github.com/innerr/ticat/pkg/utils"
)

func LoadDefaultEnv(env *model.Env, info *model.EnvKeysInfo) {
	env = env.GetLayer(model.EnvLayerDefault)

	env.Set("sys.bootstrap", "")
	env.SetInt("sys.stack-depth", 0)

	env.SetBool("sys.bg.wait", true)

	env.SetBool("sys.panic.recover", true)
	env.SetInt("sys.execute-wait-sec", 0)
	env.SetBool("sys.confirm.ask", true)

	env.Set("sys.version", "1.5.1")
	env.Set("sys.dev.name", "jungle")
	env.Set("sys.mods.integrated", "builtin")

	env.SetBool("sys.env.use-cmd-abbrs", false)

	// 100 days
	env.SetDur("sys.sessions.keep-status-duration", "2400h")

	env.Set("sys.hub.init-repo", "ticat-mods/marsh")
	env.Set("sys.self.repo", "https://github.com/innerr/ticat")

	row, col := utils.GetTerminalWidth(50, 100)
	env.SetInt("display.width.max", col)
	col = adjustDisplayWidth(col)
	env.SetInt("display.width", col)
	env.SetInt("display.height", row)

	env.SetBool("display.completion.hidden", false)
	env.SetBool("display.completion.abbr", true)
	env.SetBool("display.completion.shortcut", false)

	env.Set("display.example-https-repo", "https://github.com/ticat-mods/tidb")

	env.SetInt("display.hint.indent.2rd", 41)

	env.Set("display.utf8.symbols.tip", "💡 ")
	info.GetOrAdd("display.utf8.symbols.tip").DisplayLen = 3
	env.SetInt("display.utf8.symbols.tip.len", 3)
	env.Set("display.utf8.symbols.err", "⛔ ")
	info.GetOrAdd("display.utf8.symbols.err").DisplayLen = 3
	env.SetInt("display.utf8.symbols.err.len", 3)

	setToDefaultVerb(env)
}

func LoadEnvAbbrs(abbrs *model.EnvAbbrs) {
	sys := abbrs.GetOrAddSub("sys")
	sys.GetOrAddSub("bootstrap").AddAbbrs("boot")
	sys.GetOrAddSub("interact").AddAbbrs("ir", "i")
	sys.GetOrAddSub("wait-execute").AddAbbrs("wait-exec", "wait-exe")
	sys.GetOrAddSub("version").AddAbbrs("ver")

	hub := sys.GetOrAddSub("hub")
	hub.GetOrAddSub("init-repo").AddAbbrs("repo")

	disp := abbrs.GetOrAddSub("display").AddAbbrs("disp", "dis", "di")
	disp.GetOrAddSub("width").AddAbbrs("wid", "w")
	disp.GetOrAddSub("style").AddAbbrs("sty", "s")
	disp.GetOrAddSub("color").AddAbbrs("colour", "clr")
	utf8 := disp.GetOrAddSub("utf8")
	utf8.AddAbbrs("utf", "u")
	utf8.GetOrAddSub("symbols").AddAbbrs("symbol", "sym", "s")
	disp.GetOrAddSub("executor").AddAbbrs("exe", "exec")
	disp.GetOrAddSub("bootstrap").AddAbbrs("boot")
	disp.GetOrAddSub("one-cmd").AddAbbrs("one", "1")
	disp.GetOrAddSub("max-cmd-cnt").AddAbbrs("cmds")
	disp.GetOrAddSub("env").AddAbbrs("e")

	env := disp.GetOrAddSub("env")
	env.GetOrAddSub("sys").GetOrAddSub("paths").AddAbbrs("path")
	env.GetOrAddSub("default").AddAbbrs("def")
	env.GetOrAddSub("display").AddAbbrs("disp", "dis", "di")

	mod := disp.GetOrAddSub("mod")
	mod.GetOrAddSub("quiet").AddAbbrs("q")
	input := mod.GetOrAddSub("input-name").AddAbbrs("input")
	input.GetOrAddSub("with-realname").AddAbbrs("realname", "real", "r")
}

func LoadRuntimeEnv(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	env = env.GetLayer(model.EnvLayerSession)

	pwd, err := os.Getwd()
	if err != nil {
		panic(model.NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("get current work dir fail: %v", err)))
	}
	env.Set("sys.paths.work-dir", pwd)

	path, err := os.Executable()
	if err != nil {
		panic(model.NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("get abs self-path fail: %v", err)))
	}
	pathWithoutLinks, err := filepath.EvalSymlinks(path)
	if err == nil {
		path = pathWithoutLinks
	}
	data := path + ".data"

	sys := cc.EnvAbbrs.GetOrAddSub("sys")
	paths := sys.GetOrAddSub("paths").AddAbbrs("path", "p", "P")
	env.Set("sys.paths.hub", filepath.Join(data, "hub"))

	env.Set("sys.paths.ticat", path)
	paths.GetOrAddSub("ticat").AddAbbrs("cat")

	env.Set("sys.paths.cache", filepath.Join(data, "cache"))

	env.Set("sys.paths.env.snapshot", filepath.Join(data, "env"))

	env.Set("sys.paths.data", data)
	paths.GetOrAddSub("data").AddAbbrs("dat")

	env.Set("sys.paths.data.shared", filepath.Join(data, "shared"))

	env.Set("sys.paths.flows", filepath.Join(data, "flows"))
	paths.GetOrAddSub("flows").AddAbbrs("flow")

	env.Set("sys.paths.sessions", filepath.Join(data, "sessions"))
	paths.GetOrAddSub("sessions").AddAbbrs("session", "s", "S")

	ip := utils.IpId()
	env.Set("sys.session.id.ip", ip)

	return currCmdIdx, true
}

func LoadLocalEnv(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	kvSep := cc.Cmds.Strs.EnvKeyValSep
	delMark := cc.Cmds.Strs.EnvValDelAllMark

	path := getEnvLocalFilePath(env, flow.Cmds[currCmdIdx])
	model.LoadEnvFromFile(env.GetLayer(model.EnvLayerPersisted), path, kvSep, delMark)
	env.GetLayer(model.EnvLayerPersisted).DeleteInSelfLayer("sys.stack-depth")
	env.GetLayer(model.EnvLayerPersisted).Deduplicate()
	env.GetLayer(model.EnvLayerSession).Deduplicate()

	if !env.Has("display.color") {
		env.GetLayer(model.EnvLayerDefault).SetBool("display.color", !utils.StdoutIsPipe())
	}

	return currCmdIdx, true
}

func SaveEnvToLocal(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	kvSep := env.GetRaw("strs.env-kv-sep")
	path := getEnvLocalFilePath(env, flow.Cmds[currCmdIdx])
	model.SaveEnvToFile(env, path, kvSep, true)
	display.PrintTipTitle(cc.Screen, env,
		"changes of env are saved, could be listed by:",
		"",
		display.SuggestListEnv(env))
	return currCmdIdx, true
}

func RemoveEnvValNotSave(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return removeEnvVal(argv, cc, env, flow, currCmdIdx, false)
}

func RemoveEnvValHavePrefixNotSave(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	prefix := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "prefix")

	keyVals := env.FlattenAll()
	for key, _ := range keyVals {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		keyStr := display.ColorKey("'"+key+"'", env)
		env.DeleteEx(key, model.EnvLayerDefault)
		cc.Screen.Print(keyStr + " deleted\n")
	}
	return currCmdIdx, true
}

func RemoveEnvValAndSaveToLocal(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return removeEnvVal(argv, cc, env, flow, currCmdIdx, true)
}

// TODO: support abbrs for arg 'key'
func removeEnvVal(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int,
	saveToLocal bool) (int, bool) {

	tailMode := flow.TailModeCall && flow.Cmds[currCmdIdx].TailMode

	var keys []string
	if tailMode {
		keys = append(keys, tailModeGetInput(flow, currCmdIdx, false)...)
	}
	key := argv.GetRaw("key")
	if len(key) != 0 {
		keys = append(keys, key)
	} else if !tailMode {
		panic(model.NewCmdError(flow.Cmds[currCmdIdx], "arg 'key' is empty"))
	}
	if len(keys) == 0 {
		panic(model.NewCmdError(flow.Cmds[currCmdIdx], "no specified key to remove"))
	}

	deleted := 0
	for _, key := range keys {
		keyStr := display.ColorKey("'"+key+"'", env)
		if env.Has(key) {
			env.DeleteEx(key, model.EnvLayerDefault)
			deleted += 1
			cc.Screen.Print(keyStr + " deleted\n")
		} else {
			cc.Screen.Print(keyStr + " not exist, skipped deleting\n")
		}
	}

	if deleted != 0 && saveToLocal {
		kvSep := env.GetRaw("strs.env-kv-sep")
		path := getEnvLocalFilePath(env, flow.Cmds[currCmdIdx])
		model.SaveEnvToFile(env.GetLayer(model.EnvLayerSession), path, kvSep, true)
		display.PrintTipTitle(cc.Screen, env, "changes of env are saved")
	}
	return currCmdIdx, true
}

func ResetLocalEnv(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	path := getEnvLocalFilePath(env, flow.Cmds[currCmdIdx])
	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			display.PrintTipTitle(cc.Screen, env, "there is no saved env changes, nothing to do")
			return currCmdIdx, true
		} else {
			panic(model.NewCmdError(flow.Cmds[currCmdIdx],
				fmt.Sprintf("remove env file '%s' failed: %v", path, err)))
		}
	}
	display.PrintTipTitle(cc.Screen, env, "all saved env changes are removed")
	return currCmdIdx, true
}

func ResetSessionEnv(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	excludes := map[string]string{}
	excludeStr := argv.GetRaw("exclude-keys")
	if len(excludeStr) != 0 {
		listSep := env.GetRaw("strs.list-sep")
		for _, key := range strings.Split(excludeStr, listSep) {
			excludes[key] = env.GetRaw(key)
		}
	}

	sessionEnv := env.GetLayer(model.EnvLayerSession)
	sessionEnv.Clear(false)
	for key, val := range excludes {
		sessionEnv.Set(key, val)
	}

	title := "all session env values are removed"
	if len(excludes) != 0 {
		plural := ""
		if len(excludes) > 1 {
			plural = "s"
		}
		title += fmt.Sprintf(", exclude %d key%s by demand", len(excludes), plural)
	}
	display.PrintTipTitle(cc.Screen, env, title)
	return currCmdIdx, true
}

func EnvAssertEqual(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	key := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "key")
	val := argv.GetRaw("value")
	envVal := env.GetRaw(key)

	if val != envVal {
		panic(fmt.Errorf("assert env '%s' = '%s' failed, is '%s'", key, val, envVal))
	}
	cc.Screen.Print(display.KeyValueDisplayStr(key, val, env) + "\n")
	return currCmdIdx, true
}

func EnvAssertNotExists(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	key := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "key")

	if env.Has(key) {
		panic(fmt.Errorf("assert key '%s' not in env failed", key))
	}
	return currCmdIdx, true
}

func MapEnvKeyValueToAnotherKey(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	src := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "src-key")
	dest := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "dest-key")

	env = env.GetLayer(model.EnvLayerSession)
	value := env.GetRaw(src)
	env.Set(dest, value)

	cc.Screen.Print(display.KeyValueDisplayStr(dest, value, env) + "\n")
	return currCmdIdx, true
}

func SetEnvKeyValue(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	key := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "key")
	value := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "value")

	env = env.GetLayer(model.EnvLayerSession)
	env.Set(key, value)

	cc.Screen.Print(display.KeyValueDisplayStr(key, value, env) + "\n")
	return currCmdIdx, true
}

func UpdateEnvKeyValue(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	key := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "key")
	value := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "value")

	env = env.GetLayer(model.EnvLayerSession)
	_, ok := env.GetEx(key)
	if !ok {
		cc.Screen.Print(display.ColorError(fmt.Sprintf("env key '%s' not exists\n", key), env))
		return currCmdIdx, false
	}
	env.Set(key, value)

	cc.Screen.Print(display.KeyValueDisplayStr(key, value, env) + "\n")
	return currCmdIdx, true
}

func AddEnvKeyValue(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	key := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "key")
	value := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "value")

	env = env.GetLayer(model.EnvLayerSession)
	_, ok := env.GetEx(key)
	if ok {
		cc.Screen.Print(display.ColorError(fmt.Sprintf("env key '%s' already exists\n", key), env))
		return currCmdIdx, false
	}
	env.Set(key, value)

	cc.Screen.Print(display.KeyValueDisplayStr(key, value, env) + "\n")
	return currCmdIdx, true
}

func DisplayEnvVal(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	key := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "key")
	cc.Screen.Print(env.GetRaw(key) + "\n")
	return currCmdIdx, true
}

func getEnvLocalFilePath(env *model.Env, cmd model.ParsedCmd) string {
	path := env.GetRaw("sys.paths.data")
	file := env.GetRaw("strs.env-file-name")
	if len(path) == 0 || len(file) == 0 {
		panic(model.NewCmdError(cmd, fmt.Sprintf("can't find local data path")))
	}
	return filepath.Join(path, file)
}

func setToDefaultVerb(env *model.Env) {
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
	env.SetBool("display.env.display", false)
	env.SetInt("display.max-cmd-cnt", 14)

	env.SetBool("display.sensitive", false)

	env.SetBool("display.tip", true)
}

func adjustDisplayWidth(col int) int {
	if col > 100 {
		col = int(math.Max(100+float64(col-100)*4/5-1, float64(col-24)))
	} else {
		col -= 1
	}
	return col
}
