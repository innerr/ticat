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
	BPAStepIn   = "step in subflow"
	BPAContinue = "continue"
	BPASkip     = "skip current, stop before next command"
	BPAQuit     = "quit executing"
)

func tryDelayAndStepByStepAndBreakBefore(cc *core.Cli, env *core.Env, cmd core.ParsedCmd, breakByPrev bool) BreakPointAction {
	bpa := tryStepByStepAndBreakBefore(cc, env, cmd, breakByPrev)
	if bpa == BPAContinue {
		tryDelay(cc, env)
	} else if bpa == BPAStepIn {
		env.GetLayer(core.EnvLayerSession).SetBool("sys.step-in-subflow", true)
		bpa = BPAContinue
	}
	return bpa
}

func tryStepByStepAndBreakBefore(cc *core.Cli, env *core.Env, cmd core.ParsedCmd, breakByPrev bool) BreakPointAction {
	stepByStep := env.GetBool("sys.step-by-step")
	stepIn := env.GetBool("sys.step-in-subflow")
	name := strings.Join(cmd.Path(), cc.Cmds.Strs.PathSep)
	breakBefore := cc.BreakPoints.BreakBefore(name)

	if !breakBefore && !stepByStep && !stepIn && !breakByPrev {
		return BPAContinue
	}

	var reason string
	var choices []string
	if stepByStep {
		reason = display.ColorTip("step-by-step", env)
		choices = []string{"c"}
	} else if breakBefore {
		reason = display.ColorTip("break-point: before command ", env) + display.ColorCmd("["+name+"]", env)
		choices = []string{"c", "s"}
	} else if stepIn {
		env.GetLayer(core.EnvLayerSession).Delete("sys.step-in-subflow")
		reason = display.ColorTip("step in subflow", env)
		choices = []string{"c", "s"}
	} else if breakByPrev {
		reason = display.ColorTip("previous choice", env)
		choices = []string{"c", "s"}
	}

	if cmd.LastCmd() != nil && cmd.LastCmd().HasSubFlow() {
		choices = append(choices, "t")
	}
	choices = append(choices, "q")

	bpas := BPAs{
		"c": BPAContinue,
		"s": BPASkip,
		"q": BPAQuit,
		"t": BPAStepIn,
	}
	return readUserBPAChoice(reason, choices, bpas, true, cc.Screen, env)
}

func tryDelay(cc *core.Cli, env *core.Env) {
	delaySec := env.GetInt("sys.execute-delay-sec")
	if delaySec > 0 {
		for i := 0; i < delaySec; i++ {
			time.Sleep(time.Second)
			cc.Screen.Print(".")
		}
		cc.Screen.Print("\n")
	}
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
		BPAs{
			"c": BPAContinue,
			"q": BPAQuit,
		},
		true,
		cc.Screen,
		env)
}

func readUserBPAChoice(reason string, choices []string, actions BPAs, lowerInput bool,
	screen core.Screen, env *core.Env) BreakPointAction {

	screen.Print(display.ColorTip("[choose]", env) + " paused by '" + reason +
		"', choose action and press enter:\n")
	for _, choice := range choices {
		action := actions[choice]
		screen.Print(display.ColorWarn(choice, env) + ": " + string(action) + "\n")
	}

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
				panic(fmt.Errorf("aborted by user"))
			}
			return action
		}
		screen.Print(display.ColorExplain("(not valid input: "+line+")\n", env))
	}
}

type BPAs map[string]BreakPointAction

type BPAStatus struct {
	BreakAtNext bool
}
