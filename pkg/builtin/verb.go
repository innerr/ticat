package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

func SetQuietMode(argv core.ArgVals, cc *core.Cli, env *core.Env, _ []core.ParsedCmd) bool {
	env = env.GetLayer(core.EnvLayerSession)
	env.SetBool("display.executor", false)
	env.SetBool("display.env", false)
	env.SetBool("display.env.layer", false)
	env.SetBool("display.env.default", false)
	env.SetBool("display.env.display", false)
	env.SetBool("display.env.sys", false)
	env.SetBool("display.mod.quiet", false)
	env.SetBool("display.mod.realname", false)
	env.SetInt("display.max-cmd-cnt", 14)
	return true
}

func SetVerbMode(_ core.ArgVals, _ *core.Cli, env *core.Env, _ []core.ParsedCmd) bool {
	env = env.GetLayer(core.EnvLayerSession)
	env.SetBool("display.executor", true)
	env.SetBool("display.env", true)
	env.SetBool("display.env.layer", true)
	env.SetBool("display.env.default", true)
	env.SetBool("display.env.display", true)
	env.SetBool("display.env.sys", true)
	env.SetBool("display.mod.quiet", true)
	env.SetBool("display.mod.realname", true)
	env.SetInt("display.max-cmd-cnt", 9999)
	env.SetBool("display.executor.end", true)
	return true
}

func IncreaseVerb(argv core.ArgVals, _ *core.Cli, env *core.Env, _ []core.ParsedCmd) bool {
	env = env.GetLayer(core.EnvLayerSession)

	volume := argv.GetInt("volume")
	if volume <= 0 {
		return true
	}

	if !env.SetBool("display.executor", true) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 15)
	env.SetBool("display.mod.realname", true)
	env.SetBool("display.one-cmd", true)
	if volume <= 0 {
		return true
	}

	if !env.SetBool("display.env", true) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 16)
	if volume <= 0 {
		return true
	}

	if !env.SetBool("display.env.layer", true) {
		volume -= 1
	}
	env.SetBool("display.executor.end", true)
	env.SetInt("display.max-cmd-cnt", 17)
	if volume <= 0 {
		return true
	}

	if !env.SetBool("display.env.default", true) {
		volume -= 1
	}
	env.SetBool("display.env.display", true)
	env.SetInt("display.max-cmd-cnt", 18)
	if volume <= 0 {
		return true
	}

	if !env.SetBool("display.mod.quiet", true) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 19)
	if volume <= 0 {
		return true
	}

	if !env.SetBool("display.env.sys", true) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 20)
	if volume <= 0 {
		return true
	}

	env.SetInt("display.max-cmd-cnt", 9999)
	return true
}

func DecreaseVerb(argv core.ArgVals, _ *core.Cli, env *core.Env, _ []core.ParsedCmd) bool {
	env = env.GetLayer(core.EnvLayerSession)

	volume := argv.GetInt("volume")
	if volume <= 0 {
		return true
	}

	env.SetInt("display.max-cmd-cnt", 20)

	if env.SetBool("display.env.sys", false) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 19)
	if volume <= 0 {
		return true
	}

	if env.SetBool("display.mod.quiet", false) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 18)
	if volume <= 0 {
		return true
	}

	if env.SetBool("display.env.default", false) {
		volume -= 1
	}
	env.SetBool("display.env.display", false)
	env.SetInt("display.max-cmd-cnt", 17)
	if volume <= 0 {
		return true
	}

	if env.SetBool("display.env.layer", false) {
		volume -= 1
	}
	env.SetBool("display.executor.end", false)
	env.SetInt("display.max-cmd-cnt", 16)
	if volume <= 0 {
		return true
	}

	if env.SetBool("display.env", false) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 15)
	if volume <= 0 {
		return true
	}

	if env.SetBool("display.executor", false) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 14)

	env.SetBool("display.one-cmd", false)
	env.SetBool("display.mod.realname", false)

	return true
}

func SetToDefaultVerb(_ core.ArgVals, _ *core.Cli, env *core.Env, _ []core.ParsedCmd) bool {
	env = env.GetLayer(core.EnvLayerSession)
	setToDefaultVerb(env)
	return true
}
