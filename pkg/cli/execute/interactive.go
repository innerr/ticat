package execute

import (
	"strings"
	"time"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
	"github.com/pingcap/ticat/pkg/utils"
)

type BreakPointAction string

const (
	BPAStepOver      = "step over current command"
	BPAStepInSubFlow = "step in subflow"
	BPAContinue      = "continue to next stop"
	BPAIgnore        = "skip current command"
)

func tryBreakBefore(cc *core.Cli, env *core.Env, cmd core.ParsedCmd) bool {
	name := strings.Join(cmd.Path(), cc.Cmds.Strs.PathSep)
	if !cc.BreakPoints.BreakBefore(name) {
		return true
	}
	cc.Screen.Print(display.ColorTip("[confirm]", env) + " paused by break-point: before command " +
		display.ColorCmd("["+name+"]", env) + ", type " +
		display.ColorWarn("'y'", env) + " and press enter:\n")
	return utils.UserConfirm()
}

func tryBreakAfter(cc *core.Cli, env *core.Env, cmd core.ParsedCmd) bool {
	name := strings.Join(cmd.Path(), cc.Cmds.Strs.PathSep)
	if !cc.BreakPoints.BreakAfter(name) {
		return true
	}
	cc.Screen.Print(display.ColorTip("[confirm]", env) + " paused by break-point: post command " +
		display.ColorCmd("["+name+"]", env) + ", type " +
		display.ColorWarn("'y'", env) + " and press enter:\n")
	return utils.UserConfirm()
}

func tryDelayAndStepByStep(cc *core.Cli, env *core.Env) bool {
	delaySec := env.GetInt("sys.execute-delay-sec")
	if delaySec > 0 {
		for i := 0; i < delaySec; i++ {
			time.Sleep(time.Second)
			cc.Screen.Print(".")
		}
		cc.Screen.Print("\n")
	}
	if env.GetBool("sys.step-by-step") {
		cc.Screen.Print(display.ColorTip("[confirm]", env) + " type " +
			display.ColorWarn("'y'", env) + " and press enter:\n")
		return utils.UserConfirm()
	}
	return true
}
