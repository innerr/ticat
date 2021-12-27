package execute

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

type BreakPointAction string

const (
	BPAStepOver = "step over, execute current, stop before next command"
	BPAStepIn   = "step in subflow"
	BPAContinue = "continue"
	BPASkip     = "skip current, stop before next command"
	BPAInteract = "interactive mode"
	BPAQuit     = "quit executing"
)

func tryDelayAndStepByStepAndBreakBefore(cc *core.Cli, env *core.Env, cmd core.ParsedCmd,
	breakByPrev bool, lastCmdInFlow bool, bootstrap bool) BreakPointAction {

	if env.GetBool("sys.breakpoint.status.interact") {
		return BPAContinue
	}

	bpa := tryStepByStepAndBreakBefore(cc, env, cmd, breakByPrev)
	if bpa == BPAContinue {
		if !bootstrap && !cmd.LastCmdNode().IsQuiet() {
			tryDelay(cc, env, "sys.execute-delay-sec")
		}
	} else if bpa == BPAStepIn {
		env.GetLayer(core.EnvLayerSession).SetBool("sys.breakpoint.status.step-in", true)
		bpa = BPAContinue
	} else if bpa == BPAStepOver || bpa == BPASkip {
		if lastCmdInFlow && (cmd.LastCmd() == nil || !cmd.LastCmd().HasSubFlow()) {
			env.GetLayer(core.EnvLayerSession).SetBool("sys.breakpoint.status.step-out", true)
		}
	}
	return bpa
}

func tryStepByStepAndBreakBefore(cc *core.Cli, env *core.Env, cmd core.ParsedCmd, breakByPrev bool) BreakPointAction {
	atBegin := cc.BreakPoints.AtBegin()
	stepByStep := env.GetBool("sys.step-by-step")
	stepIn := env.GetBool("sys.breakpoint.status.step-in")
	stepOut := env.GetBool("sys.breakpoint.status.step-out")
	name := strings.Join(cmd.Path(), cc.Cmds.Strs.PathSep)
	breakBefore := cc.BreakPoints.BreakBefore(name)

	if !atBegin && !breakBefore && !stepByStep && !stepIn && !stepOut && !breakByPrev {
		return BPAContinue
	}

	choices := []string{}
	var reason string

	if cmd.LastCmd() != nil && cmd.LastCmd().HasSubFlow() && !stepByStep {
		choices = append(choices, "t")
	}

	if atBegin {
		cc.BreakPoints.SetAtBegin(false)
		reason = display.ColorTip("break-point: at begin", env)
		choices = append(choices, "s", "d", "c")
	} else if stepByStep {
		reason = display.ColorTip("step-by-step", env)
		choices = append(choices, "c")
	} else if breakBefore {
		reason = display.ColorTip("break-point: before command ", env) + display.ColorCmd("["+name+"]", env)
		choices = append(choices, "s", "d", "c")
	} else if stepIn {
		env.GetLayer(core.EnvLayerSession).Delete("sys.breakpoint.status.step-in")
		reason = display.ColorTip("just stepped in", env)
		choices = append(choices, "s", "d", "c")
	} else if stepOut {
		env.GetLayer(core.EnvLayerSession).Delete("sys.breakpoint.status.step-out")
		reason = display.ColorTip("just stepped out", env)
		choices = append(choices, "s", "d", "c")
	} else if breakByPrev {
		reason = display.ColorTip("previous choice", env)
		choices = append(choices, "s", "d", "c")
	}

	choices = append(choices, "i", "q")

	all := getAllBPAs()
	bpas := BPAs{}
	for _, k := range choices {
		bpas[k] = all[k]
	}
	return readUserBPAChoice(reason, choices, bpas, true, cc, env)
}

func tryDelayAndBreakAfter(cc *core.Cli, env *core.Env, cmd core.ParsedCmd, bootstrap bool) BreakPointAction {
	bpa := tryBreakAfter(cc, env, cmd)
	if bpa == BPAContinue && !bootstrap && !cmd.LastCmdNode().IsQuiet() {
		tryDelay(cc, env, "sys.execute-delay-sec.at-end")
	}
	return bpa
}

func tryBreakAfter(cc *core.Cli, env *core.Env, cmd core.ParsedCmd) BreakPointAction {
	name := strings.Join(cmd.Path(), cc.Cmds.Strs.PathSep)
	if !cc.BreakPoints.BreakAfter(name) {
		return BPAContinue
	}
	reason := display.ColorTip("break-point: after command ", env) + display.ColorCmd("["+name+"]", env)
	return readUserBPAChoice(
		reason,
		[]string{"c", "q"},
		getAllBPAs(),
		true,
		cc,
		env)
}

func tryDelay(cc *core.Cli, env *core.Env, delayKey string) {
	delaySec := env.GetInt(delayKey)
	if delaySec > 0 {
		for i := 0; i < delaySec; i++ {
			time.Sleep(time.Second)
			cc.Screen.Print(".")
		}
		cc.Screen.Print("\n")
	}
}

func readUserBPAChoice(reason string, choices []string, actions BPAs, lowerInput bool,
	cc *core.Cli, env *core.Env) BreakPointAction {

	showTitle := func() {
		cc.Screen.Print(display.ColorTip("[choose]", env) + " paused by '" + reason +
			"', choose action and press enter:\n")
		for _, choice := range choices {
			action := actions[choice]
			cc.Screen.Print(display.ColorWarn(choice, env) + ": " + string(action) + "\n")
		}
	}

	showTitle()

	buf := bufio.NewReader(os.Stdin)
	for {
		lineBytes, err := buf.ReadBytes('\n')
		if err != nil {
			panic(fmt.Errorf("[readFromStdin] read from stdin failed: %v", err))
		}
		if len(lineBytes) == 0 {
			continue
		}
		line := strings.TrimSpace(string(lineBytes))
		if lowerInput {
			line = strings.ToLower(line)
		}
		if action, ok := actions[line]; ok {
			if action == BPAQuit {
				panic(core.NewAbortByUserErr())
			} else if action == BPAInteract {
				interactiveMode(cc, env, "e")
				if env.GetBool("sys.breakpoint.status.interact.leaving") {
					env.GetLayer(core.EnvLayerSession).SetBool("sys.breakpoint.status.interact.leaving", false)
					return BPAContinue
				}
				cc.Screen.Print("\n")
				showTitle()
				continue
			}
			return action
		}
		cc.Screen.Print(display.ColorExplain("(not valid input: "+line+")\n", env))
	}
}

func interactiveMode(cc *core.Cli, env *core.Env, exitStr string) {
	sessionEnv := env.GetLayer(core.EnvLayerSession)
	sessionEnv.SetBool("sys.breakpoint.status.interact", true)

	cc = cc.CopyForInteract()
	buf := bufio.NewReader(os.Stdin)
	for {
		if env.GetBool("sys.breakpoint.status.interact.leaving") {
			break
		}
		selfName := env.GetRaw("strs.self-name")
		cc.Screen.Print("\n" + display.ColorExplain("", env) + display.ColorWarn(exitStr+":", env) +
			display.ColorExplain(" exit interactive mode\n", env))

		cc.Screen.Print(display.ColorTip(selfName+"> ", env))
		lineBytes, err := buf.ReadBytes('\n')
		if err != nil {
			panic(fmt.Errorf("[readFromStdin] read from stdin failed: %v", err))
		}
		if len(lineBytes) == 0 {
			continue
		}
		line := strings.TrimSpace(string(lineBytes))
		if line == exitStr {
			break
		}
		cc.Executor.Execute("(interact)", cc, env, nil, strings.Fields(line)...)
	}

	sessionEnv.GetLayer(core.EnvLayerSession).SetBool("sys.breakpoint.status.interact", false)
}

func getAllBPAs() BPAs {
	return BPAs{
		"c": BPAContinue,
		"s": BPASkip,
		"q": BPAQuit,
		"t": BPAStepIn,
		"d": BPAStepOver,
		"i": BPAInteract,
	}
}

type BPAs map[string]BreakPointAction

type BPAStatus struct {
	BreakAtNext bool
}
