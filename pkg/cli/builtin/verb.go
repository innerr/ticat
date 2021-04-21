package builtin

import (
	"github.com/pingcap/ticat/pkg/cli"
)

func SetQuietMode(_ cli.ArgVals, _ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerSession)
	env.Set("runtime.display", "false")
	return true
}

func SetVerbMode(_ cli.ArgVals, _ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerSession)
	env.Set("runtime.display", "true")
	env.Set("runtime.display.env", "true")
	env.Set("runtime.display.env.layer", "true")
	env.Set("runtime.display.env.default", "true")
	env.Set("runtime.display.env.runtime.sys", "true")
	env.Set("runtime.display.mod.builtin", "true")
	env.Set("runtime.display.mod.realname", "true")
	env.Set("runtime.display.max-cmd-cnt", "9999")
	return true
}

func IncreaseVerb(argv cli.ArgVals, _ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerSession)

	volume := argv.GetInt("volume")
	if volume <= 0 {
		return true
	}

	if !env.Get("runtime.display").GetBool() {
		env.Set("runtime.display", "true")
		volume -= 1
	}
	env.Set("runtime.display.mod.realname", "true")
	env.Set("runtime.display.one-cmd", "true")
	if volume <= 0 {
		return true
	}

	if !env.Get("runtime.display.env").GetBool() {
		env.Set("runtime.display.env", "true")
		volume -= 1
	}
	if volume <= 0 {
		return true
	}

	if !env.Get("runtime.display.env.layer").GetBool() {
		env.Set("runtime.display.env.layer", "true")
		volume -= 1
	}
	if volume <= 0 {
		return true
	}

	if !env.Get("runtime.display.env.default").GetBool() {
		env.Set("runtime.display.env.default", "true")
		env.Set("runtime.display.max-cmd-cnt", "10")
		volume -= 1
	}
	if volume <= 0 {
		return true
	}

	if !env.Get("runtime.display.env.runtime.sys").GetBool() ||
		!env.Get("runtime.display.mod.builtin").GetBool() {
		env.Set("runtime.display.env.runtime.sys", "true")
		env.Set("runtime.display.mod.builtin", "true")
		env.Set("runtime.display.max-cmd-cnt", "14")
		volume -= 1
	}
	if volume <= 0 {
		return true
	}

	env.Set("runtime.display.max-cmd-cnt", "9999")

	return true
}

func DecreaseVerb(argv cli.ArgVals, _ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerSession)

	volume := argv.GetInt("volume")
	if volume <= 0 {
		return true
	}

	if env.Get("runtime.display.env.runtime.sys").GetBool() ||
		env.Get("runtime.display.mod.builtin").GetBool() {
		env.Set("runtime.display.env.runtime.sys", "false")
		env.Set("runtime.display.mod.builtin", "false")
		env.Set("runtime.display.max-cmd-cnt", "10")
		volume -= 1
	}
	env.Set("runtime.display.max-cmd-cnt", "14")
	if volume <= 0 {
		return true
	}

	if env.Get("runtime.display.env.default").GetBool() {
		env.Set("runtime.display.env.default", "false")
		env.Set("runtime.display.max-cmd-cnt", "7")
		volume -= 1
	}
	if volume <= 0 {
		return true
	}

	if env.Get("runtime.display.env.layer").GetBool() {
		env.Set("runtime.display.env.layer", "false")
		volume -= 1
	}
	if volume <= 0 {
		return true
	}

	if env.Get("runtime.display.env").GetBool() {
		env.Set("runtime.display.env", "false")
		volume -= 1
	}
	if volume <= 0 {
		return true
	}

	if env.Get("runtime.display").GetBool() {
		env.Set("runtime.display", "false")
		volume -= 1
	}

	env.Set("runtime.display.one-cmd", "false")
	env.Set("runtime.display.mod.realname", "false")

	return true
}
