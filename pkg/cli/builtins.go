package cli

import (
)

func RegisterBuiltins(cmds *CmdTree) {
	cmds.AddSub("help", "h", "HELP", "H", "?").SetCmd(GlobalHelp)

	builtin := cmds.AddSub("builtin")
	builtin.AddSub("greeting").AddSub("dev").SetCmd(GreetingDev)

	env := builtin.AddSub("env")
	env.AddSub("load").AddSub("local").SetCmd(LoadLocalEnv)

	mod := builtin.AddSub("mod")
	mod.AddSub("load").AddSub("local").SetCmd(LoadLocalMods)
}

func GlobalHelp(ctx *Context, argv []string) {
	print("TODO: global help")
}

func LoadLocalEnv(ctx *Context, argv []string) {
	print("TODO: load local env")
}

func LoadLocalMods(ctx *Context, argv []string) {
	print("TODO: load local mods")
}

func GreetingDev(ctx *Context, argv []string) {
	print("Hello World")
}
