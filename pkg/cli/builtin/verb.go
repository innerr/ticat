package builtin

import (
	"github.com/pingcap/ticat/pkg/cli"
)

func SetQuietMode(_ cli.ArgVals, _ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerSession)
	env.Set("display", "false")
	return true
}

func SetVerbMode(_ cli.ArgVals, _ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerSession)
	env.Set("display", "true")
	env.Set("display.env", "true")
	env.Set("display.env.layer", "true")
	env.Set("display.env.default", "true")
	env.Set("display.env.sys", "true")
	env.Set("display.mod.quiet", "true")
	env.Set("display.mod.realname", "true")
	env.Set("display.max-cmd-cnt", "9999")
	return true
}

func IncreaseVerb(argv cli.ArgVals, _ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerSession)

	volume := argv.GetInt("volume")
	if volume <= 0 {
		return true
	}

	if !env.SetBool("display", true) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 8)
	env.SetBool("display.mod.realname", true)
	env.SetBool("display.one-cmd", true)
	if volume <= 0 {
		return true
	}

	if !env.SetBool("display.env", true) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 9)
	if volume <= 0 {
		return true
	}

	if !env.SetBool("display.env.layer", true) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 10)
	if volume <= 0 {
		return true
	}

	if !env.SetBool("display.env.default", true) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 11)
	if volume <= 0 {
		return true
	}

	if !env.SetBool("display.env.sys", true) ||
		!env.SetBool("display.mod.quiet", true) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 12)
	if volume <= 0 {
		return true
	}

	env.SetInt("display.max-cmd-cnt", 9999)
	return true
}

func DecreaseVerb(argv cli.ArgVals, _ *cli.Cli, env *cli.Env) bool {
	env = env.GetLayer(cli.EnvLayerSession)

	volume := argv.GetInt("volume")
	if volume <= 0 {
		return true
	}

	env.SetInt("display.max-cmd-cnt", 12)

	if env.SetBool("display.env.sys", false) ||
		env.SetBool("display.mod.quiet", false) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 11)
	if volume <= 0 {
		return true
	}

	if env.SetBool("display.env.default", false) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 10)
	if volume <= 0 {
		return true
	}

	if env.SetBool("display.env.layer", false) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 9)
	if volume <= 0 {
		return true
	}

	if env.SetBool("display.env", false) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 8)
	if volume <= 0 {
		return true
	}

	if env.SetBool("display", false) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 7)

	env.SetBool("display.one-cmd", false)
	env.SetBool("display.mod.realname", false)

	return true
}
