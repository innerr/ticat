package cli

import (
	"fmt"
	"os"
	"time"
)

func RegisterBuiltins(cmds *CmdTree) {
	cmds.AddSub("help", "h", "HELP", "H", "?").SetQuietPowerCmd(GlobalHelp)
	cmds.AddSub("verbose", "verb", "v", "V").SetCmd(SetVerbMode)
	cmds.AddSub("quiet", "q", "Q").SetQuietCmd(SetQuietMode)
	cmds.AddSub("dummy", "D").SetCmd(Dummy)
	cmds.AddSub("sleep", "slp").SetCmd(Sleep)

	builtin := cmds.AddSub("builtin")

	env := builtin.AddSub("env")
	envLoad := env.AddSub("load")
	envLoad.AddSub("local").SetCmd(LoadLocalEnv)
	envLoad.AddSub("runtime").SetCmd(LoadRuntimeEnv)

	mod := builtin.AddSub("mod")
	mod.AddSub("load").AddSub("local").SetCmd(LoadLocalMods)
}

func LoadBuiltinEnv(env *Env) {
	env = env.GetLayer(EnvLayerDefault)
	env.Set("runtime.version", "dev")
	env.Set("runtime.stack-depth", "0")
	env.Set("runtime.display", "true")
	env.Set("runtime.display.width", "80")
	env.Set("runtime.display.max-cmd-cnt", "7")
	env.Set("runtime.display.env", "true")
	env.Set("runtime.display.env.layer", "false")
	env.Set("runtime.display.env.default", "false")
	env.Set("runtime.display.env.runtime", "false")
	env.Set("runtime.display.mod.builtin", "false")
	env.Set("runtime.display.mod.realname", "false")
}

func SetQuietMode(cli *Cli, env *Env) bool {
	env.Set("runtime.display", "false")
	return true
}

func SetVerbMode(cli *Cli, env *Env) bool {
	env = env.GetLayer(EnvLayerSession)
	env.Set("runtime.display", "true")
	env.Set("runtime.display.max-cmd-cnt", "9999")
	env.Set("runtime.display.env", "true")
	env.Set("runtime.display.env.layer", "true")
	env.Set("runtime.display.env.default", "true")
	env.Set("runtime.display.env.runtime.sys", "true")
	env.Set("runtime.display.mod.builtin", "true")
	env.Set("runtime.display.mod.realname", "true")
	return true
}

func GlobalHelp(cli *Cli, env *Env, cmds []ParsedCmd, currCmdIdx int) (modified []ParsedCmd, succeeded bool) {
	fmt.Println("TODO: global help")
	return nil, true
}

func LoadRuntimeEnv(cli *Cli, env *Env) bool {
	env.Set("runtime.sys.ticap-path", os.Args[0])
	return true
}

func LoadLocalEnv(cli *Cli, env *Env) bool {
	//fmt.Println("TODO: load local env")
	return true
}

func LoadLocalMods(cli *Cli, env *Env) bool {
	//fmt.Println("TODO: load local mods")
	return true
}

func Sleep(cli *Cli, env *Env) bool {
	dur, err := time.ParseDuration(env.Get("sleep").Raw)
	if err != nil {
		fmt.Printf("[ERR] %v\n", err)
		return false
	}
	time.Sleep(dur)
	return true
}

func Dummy(cli *Cli, env *Env) bool {
	fmt.Println("Dummy cmd here")
	return true
}
