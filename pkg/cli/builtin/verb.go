package builtin

import (
	"github.com/pingcap/ticat/pkg/cli"
)

func SetQuietMode(_ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerSession)
	env.Set("runtime.display", "false")
	return true
}

func SetVerbMode(_ *cli.Cli, env *cli.Env) bool {
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

func IncreaseVerb(_ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerSession)

	if !env.Get("runtime.display").GetBool() {
		env.Set("runtime.display", "true")
		return true
	}

	if !env.Get("runtime.display.env").GetBool() {
		env.Set("runtime.display.env", "true")
		return true
	}

	if !env.Get("runtime.display.env.layer").GetBool() {
		env.Set("runtime.display.env.layer", "true")
		return true
	}

	if !env.Get("runtime.display.env.default").GetBool() {
		env.Set("runtime.display.env.default", "true")
		env.Set("runtime.display.max-cmd-cnt", "10")
		return true
	}

	if !env.Get("runtime.display.env.runtime.sys").GetBool() ||
		!env.Get("runtime.display.mod.builtin").GetBool() {
		env.Set("runtime.display.env.runtime.sys", "true")
		env.Set("runtime.display.mod.builtin", "true")
		env.Set("runtime.display.max-cmd-cnt", "14")
		return true
	}

	if !env.Get("runtime.display.mod.realname").GetBool() {
		env.Set("runtime.display.mod.realname", "true")
		env.Set("runtime.display.max-cmd-cnt", "9999")
		return true
	}

	return true
}

func DecreaseVerb(_ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerSession)

	if env.Get("runtime.display.mod.realname").GetBool() {
		env.Set("runtime.display.mod.realname", "false")
		env.Set("runtime.display.max-cmd-cnt", "14")
		return true
	}

	if env.Get("runtime.display.env.runtime.sys").GetBool() ||
		env.Get("runtime.display.mod.builtin").GetBool() {
		env.Set("runtime.display.env.runtime.sys", "false")
		env.Set("runtime.display.mod.builtin", "false")
		env.Set("runtime.display.max-cmd-cnt", "10")
		return true
	}

	if env.Get("runtime.display.env.default").GetBool() {
		env.Set("runtime.display.env.default", "false")
		env.Set("runtime.display.max-cmd-cnt", "7")
		return true
	}

	if env.Get("runtime.display.env.layer").GetBool() {
		env.Set("runtime.display.env.layer", "false")
		return true
	}

	if env.Get("runtime.display.env").GetBool() {
		env.Set("runtime.display.env", "false")
		return true
	}

	if env.Get("runtime.display").GetBool() {
		env.Set("runtime.display", "false")
		return true
	}

	return true
}
