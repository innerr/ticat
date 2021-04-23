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
	} else {
		env.PlusInt("runtime.display.max-cmd-cnt", volume)
	}

	if !env.SetBool("runtime.display", true) {
		volume -= 1
	}
	env.SetBool("runtime.display.mod.realname", true)
	env.SetBool("runtime.display.one-cmd", true)
	if volume <= 0 {
		return true
	}

	if !env.SetBool("runtime.display.env", true) {
		volume -= 1
	}
	if volume <= 0 {
		return true
	}

	if !env.SetBool("runtime.display.env.layer", true) {
		volume -= 1
	}
	if volume <= 0 {
		return true
	}

	if !env.SetBool("runtime.display.env.default", true) {
		volume -= 1
	}
	if volume <= 0 {
		return true
	}

	if !env.SetBool("runtime.display.env.runtime.sys", true) ||
		!env.SetBool("runtime.display.mod.builtin", true) {
		volume -= 1
	}
	if volume <= 0 {
		return true
	}

	env.SetInt("runtime.display.max-cmd-cnt", 9999)
	return true
}

func DecreaseVerb(argv cli.ArgVals, _ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerSession)

	volume := argv.GetInt("volume")
	if volume <= 0 {
		return true
	}

	maxCmdCnt := env.GetInt("runtime.display.max-cmd-cnt")
	if maxCmdCnt > 12 {
		env.SetInt("runtime.display.max-cmd-cnt", 12)
	}
	env.PlusInt("runtime.display.max-cmd-cnt", -volume)

	if env.SetBool("runtime.display.env.runtime.sys", false) ||
		env.SetBool("runtime.display.mod.builtin", false) {
		volume -= 1
	}
	if volume <= 0 {
		return true
	}

	if env.SetBool("runtime.display.env.default", false) {
		volume -= 1
	}
	if volume <= 0 {
		return true
	}

	if env.SetBool("runtime.display.env.layer", false) {
		volume -= 1
	}
	if volume <= 0 {
		return true
	}

	if env.SetBool("runtime.display.env", false) {
		volume -= 1
	}
	if volume <= 0 {
		return true
	}

	if env.SetBool("runtime.display", false) {
		volume -= 1
	}

	env.SetBool("runtime.display.one-cmd", false)
	env.SetBool("runtime.display.mod.realname", false)

	return true
}
