package cli

func RegisterBuiltins(cmds *CmdTree) {
	cmds.AddSub("help", "h", "HELP", "H", "?").SetCmd(GlobalHelp)

	builtin := cmds.AddSub("builtin")
	builtin.AddSub("greeting").AddSub("dev").SetCmd(GreetingDev)

	env := builtin.AddSub("env")
	env.AddSub("load").AddSub("local").SetCmd(LoadLocalEnv)

	mod := builtin.AddSub("mod")
	mod.AddSub("load").AddSub("local").SetCmd(LoadLocalMods)
}

func GlobalHelp(cli *Cli, env *Env) bool {
	print("TODO: global help")
	return true
}

func LoadLocalEnv(cli *Cli, env *Env) bool {
	println("TODO: load local env")
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
