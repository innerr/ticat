package cli

import (
	"os"
)

func RegisterBuiltins(cmds *CmdTree) {
	cmds.AddSub("help", "h", "HELP", "H", "?").SetCmd(GlobalHelp)

	builtin := cmds.AddSub("builtin")
	builtin.AddSub("greeting").AddSub("dev").SetCmd(GreetingDev)

	env := builtin.AddSub("env")
	envLoad := env.AddSub("load")
	envLoad.AddSub("local").SetCmd(LoadLocalEnv)
	envLoad.AddSub("runtime").SetCmd(LoadRuntimeEnv)
	envLoad.AddSub("builtin").SetCmd(LoadBuiltinEnv)

	mod := builtin.AddSub("mod")
	mod.AddSub("load").AddSub("local").SetCmd(LoadLocalMods)
}

func GlobalHelp(cli *Cli, env *Env) bool {
	println("TODO: global help")
	return true
}

func LoadLocalEnv(cli *Cli, env *Env) bool {
	println("TODO: load local env")
	return true
}

func LoadBuiltinEnv(cli *Cli, env *Env) bool {
	env.Set("runtime.version", "dev")
	env.Set("runtime.stack-depth", "0")
	env.Set("runtime.display.executor.width", "80")
	env.Set("runtime.display.executor.max-cmd-cnt", "7")
	return true
}

func LoadRuntimeEnv(cli *Cli, env *Env) bool {
	env.Set("runtime.ticap-path", os.Args[0])
	return true
}

func LoadLocalMods(cli *Cli, env *Env) bool {
	println("TODO: load local mods")
	return true
}

func GreetingDev(cli *Cli, env *Env) bool {
	println("Hello World")
	return true
}
