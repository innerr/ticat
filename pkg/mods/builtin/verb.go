package builtin

import (
	"github.com/innerr/ticat/pkg/core/model"
)

func SetQuietMode(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	env = env.GetLayer(model.EnvLayerSession)
	env.SetBool("display.executor", false)
	env.SetBool("display.env", false)
	env.SetBool("display.env.layer", false)
	env.SetBool("display.env.default", false)
	env.SetBool("display.env.display", false)
	env.SetBool("display.env.sys", false)
	env.SetBool("display.mod.quiet", false)
	env.SetBool("display.mod.realname", false)
	env.SetInt("display.max-cmd-cnt", 14)
	return currCmdIdx, nil
}

func SetVerbMode(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	env = env.GetLayer(model.EnvLayerSession)
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
	return currCmdIdx, nil
}

func IncreaseVerb(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	env = env.GetLayer(model.EnvLayerSession)

	volume := argv.GetInt("volume")
	if volume <= 0 {
		return currCmdIdx, nil
	}

	if !env.SetBool("display.executor", true) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 15)
	env.SetBool("display.mod.realname", true)
	env.SetBool("display.one-cmd", true)
	if volume <= 0 {
		return currCmdIdx, nil
	}

	if !env.SetBool("display.env", true) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 16)
	if volume <= 0 {
		return currCmdIdx, nil
	}

	if !env.SetBool("display.env.layer", true) {
		volume -= 1
	}
	env.SetBool("display.executor.end", true)
	env.SetInt("display.max-cmd-cnt", 17)
	if volume <= 0 {
		return currCmdIdx, nil
	}

	if !env.SetBool("display.env.default", true) {
		volume -= 1
	}
	env.SetBool("display.env.display", true)
	env.SetInt("display.max-cmd-cnt", 18)
	if volume <= 0 {
		return currCmdIdx, nil
	}

	if !env.SetBool("display.mod.quiet", true) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 19)
	if volume <= 0 {
		return currCmdIdx, nil
	}

	if !env.SetBool("display.env.sys", true) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 20)
	if volume <= 0 {
		return currCmdIdx, nil
	}

	env.SetInt("display.max-cmd-cnt", 9999)
	return currCmdIdx, nil
}

func DecreaseVerb(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	env = env.GetLayer(model.EnvLayerSession)

	volume := argv.GetInt("volume")
	if volume <= 0 {
		return currCmdIdx, nil
	}

	env.SetInt("display.max-cmd-cnt", 20)

	if env.SetBool("display.env.sys", false) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 19)
	if volume <= 0 {
		return currCmdIdx, nil
	}

	if env.SetBool("display.mod.quiet", false) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 18)
	if volume <= 0 {
		return currCmdIdx, nil
	}

	if env.SetBool("display.env.default", false) {
		volume -= 1
	}
	env.SetBool("display.env.display", false)
	env.SetInt("display.max-cmd-cnt", 17)
	if volume <= 0 {
		return currCmdIdx, nil
	}

	if env.SetBool("display.env.layer", false) {
		volume -= 1
	}
	env.SetBool("display.executor.end", false)
	env.SetInt("display.max-cmd-cnt", 16)
	if volume <= 0 {
		return currCmdIdx, nil
	}

	if env.SetBool("display.env", false) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 15)
	if volume <= 0 {
		return currCmdIdx, nil
	}

	if env.SetBool("display.executor", false) {
		volume -= 1
	}
	env.SetInt("display.max-cmd-cnt", 14)

	env.SetBool("display.one-cmd", false)
	env.SetBool("display.mod.realname", false)

	return currCmdIdx, nil
}

func SetToDefaultVerb(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	env = env.GetLayer(model.EnvLayerSession)
	setToDefaultVerb(env)
	return currCmdIdx, nil
}
