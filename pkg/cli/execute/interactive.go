package execute

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pingcap/ticat/pkg/builtin"
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

type BreakPointAction string

const (
	BPAStepOver   = "step over, execute current, pause before next command"
	BPAStepToNext = "step over, pause before next command"
	BPAStepIn     = "step into subflow"
	BPAContinue   = "continue"
	BPASkip       = "skip current, pause before next command"
	BPAInteract   = "interactive mode"
	BPAQuit       = "quit executing"
)

func tryWaitSecAndBreakBefore(cc *core.Cli, env *core.Env, cmd core.ParsedCmd, mask *core.ExecuteMask,
	/*breakByPrev bool, */ lastCmdInFlow bool, bootstrap bool, showStack func()) BreakPointAction {

	if env.GetBool("sys.interact.inside") {
		return BPAContinue
	}

	if cmd.IsQuiet() && !env.GetBool("sys.breakpoint.here.now") {
		return BPAContinue
	}

	bpa := tryBreakBefore(cc, env, cmd, mask /*breakByPrev, */, showStack)
	if bpa == BPAContinue {
		if !bootstrap && cmd.LastCmdNode() != nil && !cmd.LastCmdNode().IsQuiet() {
			tryWaitSec(cc, env, "sys.execute-wait-sec")
		}
	} else if bpa == BPAStepIn {
		env.GetLayer(core.EnvLayerSession).SetBool("sys.breakpoint.status.step-in", true)
		bpa = BPAContinue
	} else if bpa == BPAStepOver || bpa == BPASkip {
		//if lastCmdInFlow && (cmd.LastCmd() == nil || !cmd.LastCmd().HasSubFlow(false)) {
		if lastCmdInFlow {
			env.GetLayer(core.EnvLayerSession).SetBool("sys.breakpoint.status.step-out", true)
		}
	}
	return bpa
}

func tryBreakBefore(cc *core.Cli, env *core.Env, cmd core.ParsedCmd, mask *core.ExecuteMask,
	/*breakByPrev bool, */ showStack func()) BreakPointAction {

	stepIn := env.GetBool("sys.breakpoint.status.step-in")
	stepOut := env.GetBool("sys.breakpoint.status.step-out")
	name := strings.Join(cmd.Path(), cc.Cmds.Strs.PathSep)

	breakHereKey := "sys.breakpoint.here.now"
	breakBefore := cc.BreakPoints.BreakBefore(name) || env.GetBool(breakHereKey)
	env.GetLayer(core.EnvLayerSession).Delete(breakHereKey)

	breakByPrev := env.GetBool("sys.breakpoint.at-next")
	if !breakBefore && !stepIn && !stepOut && !breakByPrev {
		return BPAContinue
	}

	choices := []string{}
	var reason string

	if cmd.LastCmd() != nil && cmd.LastCmd().HasSubFlow(false) && (mask == nil || mask.SubFlow != nil) {
		choices = append(choices, "t")
	}

	if breakBefore {
		reason = display.ColorTip("break-point: before command ", env) + display.ColorCmd("["+name+"]", env)
		choices = append(choices, "s", "d", "c")
	} else if stepIn {
		env.GetLayer(core.EnvLayerSession).Delete("sys.breakpoint.status.step-in")
		reason = display.ColorTip("just stepped in", env)
		choices = append(choices, "s", "d", "c")
	} else if stepOut {
		if env.GetBool("sys.breakpoint.status.step-out") {
			env.GetLayer(core.EnvLayerSession).Delete("sys.breakpoint.status.step-out")
		}
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
	return readUserBPAChoice(reason, choices, bpas, true, cc, env, showStack)
}

func tryWaitSecAndBreakAfter(cc *core.Cli, env *core.Env, cmd core.ParsedCmd, bootstrap bool, lastCmdInFlow bool, showStack func()) BreakPointAction {
	bpa := tryBreakAfter(cc, env, cmd, showStack)
	if bpa == BPAStepOver {
		if lastCmdInFlow && (cmd.LastCmd() == nil || !cmd.LastCmd().HasSubFlow(false)) {
			env.GetLayer(core.EnvLayerSession).SetBool("sys.breakpoint.status.step-out", true)
		}
	} else if bpa == BPAContinue && !bootstrap && cmd.LastCmdNode() != nil && !cmd.LastCmdNode().IsQuiet() {
		tryWaitSec(cc, env, "sys.execute-wait-sec.at-end")
	}
	return bpa
}

func tryBreakAfter(cc *core.Cli, env *core.Env, cmd core.ParsedCmd, showStack func()) BreakPointAction {
	name := strings.Join(cmd.Path(), cc.Cmds.Strs.PathSep)
	if !cc.BreakPoints.BreakAfter(name) {
		return BPAContinue
	}
	reason := display.ColorTip("break-point: after command ", env) + display.ColorCmd("["+name+"]", env)
	bpas := getAllBPAs()
	// Use BPAStepToNext instead of BPAStepOverf or display
	bpas["d"] = BPAStepToNext
	bpa := readUserBPAChoice(reason, []string{"d", "c", "i", "q"}, bpas, true, cc, env, showStack)
	if bpa == BPAStepToNext {
		bpa = BPAStepOver
	}
	return bpa
}

func tryBreakAtEnd(cc *core.Cli, env *core.Env) {
	breakHereKey := "sys.breakpoint.here.now"
	breakHere := env.GetBool(breakHereKey)
	if !cc.BreakPoints.BreakAtEnd() && !breakHere {
		return
	}
	env.GetLayer(core.EnvLayerSession).Delete(breakHereKey)

	showEOF := func() {
		cc.Screen.Print(display.ColorExplain("(end of flow)\n", env))
	}
	showEOF()

	reason := display.ColorTip("break-point: at main-thread end", env)
	bpa := readUserBPAChoice(
		reason,
		[]string{"c", "i", "q"},
		getAllBPAs(),
		true,
		cc,
		env,
		showEOF)
	if bpa != BPAContinue {
		panic(fmt.Errorf("[tryBreakAtEnd] should never happen"))
	}
}

func tryWaitSec(cc *core.Cli, env *core.Env, waitSecKey string) {
	waitSec := env.GetInt(waitSecKey)
	if waitSec > 0 {
		for i := 0; i < waitSec; i++ {
			time.Sleep(time.Second)
			cc.Screen.Print(".")
		}
		cc.Screen.Print("\n")
	}
}

func tryBreakInsideFileNFlow(cc *core.Cli, env *core.Env, cmd *core.Cmd /*breakByPrev bool, */, showStack func()) (shouldExec bool) {
	breakByPrev := env.GetBool("sys.breakpoint.at-next")

	if !breakByPrev && !env.GetBool("sys.breakpoint.status.step-out") {
		return true
	}

	name := cmd.Owner().DisplayPath()
	reason := display.ColorCmd("["+name+"]", env) + display.ColorTip(": pre-executing subflow done, before executing", env)
	bpa := readUserBPAChoice(
		reason,
		[]string{"s", "d", "c", "i", "q"},
		getAllBPAs(),
		true,
		cc,
		env,
		showStack)
	if bpa == BPAContinue {
		env.GetLayer(core.EnvLayerSession).Delete("sys.breakpoint.status.step-out")
		return true
	}
	if bpa == BPAStepOver {
		return true
	}
	return false
}

func clearBreakPointStatusInEnv(env *core.Env) {
	env = env.GetLayer(core.EnvLayerSession)
	env.Delete("sys.interact.leaving")
	env.Delete("sys.breakpoint.status.step-in")
	env.Delete("sys.breakpoint.status.step-out")
}

func readUserBPAChoice(reason string, choices []string, actions BPAs, lowerInput bool,
	cc *core.Cli, env *core.Env, showStack func()) BreakPointAction {

	showTitle := func() {
		cc.Screen.Print(display.ColorTip("[actions]", env) + " paused by '" + reason +
			"', choose one and press enter:\n")
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
				cc.Screen.Print("\n")
				builtin.InteractiveMode(cc, env, "e")
				if env.GetBool("sys.interact.leaving") {
					env.GetLayer(core.EnvLayerSession).Delete("sys.interact.leaving")
					return BPAContinue
				}
				cc.Screen.Print("\n")
				if showStack != nil {
					showStack()
				}
				showTitle()
				continue
			} else if action == BPAContinue {
				env.GetLayer(core.EnvLayerSession).SetBool("sys.breakpoint.at-next", false)
			}
			return action
		}
		cc.Screen.Print(display.ColorExplain("(not valid input: "+line+")\n", env))
	}
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
