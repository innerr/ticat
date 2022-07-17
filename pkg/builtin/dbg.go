package builtin

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func DbgEcho(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	msg := argv.GetRaw("message")
	color := argv.GetRaw("color")

	rendered, _ := core.RenderTemplateStrLines([]string{msg}, "echo", flow.Cmds[currCmdIdx].LastCmd(), core.ArgVals{}, env, true)
	str := rendered[0]
	str = display.ColorStrByName(str, color, env)

	cc.Screen.Print(fmt.Sprintf("%v\n", str))
	return currCmdIdx, true
}

func DbgEchoLn(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cc.Screen.Print("\n")
	return currCmdIdx, true
}

func DbgExecBash(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	cmd := exec.Command("bash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(core.WrapCmdError(flow.Cmds[currCmdIdx], err))
	}

	return currCmdIdx, true
}

func DbgWaitSecExecute(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	env.GetLayer(core.EnvLayerSession).SetInt("sys.execute-wait-sec", argv.GetInt("seconds"))
	return currCmdIdx, true
}

func DbgWaitSecExecuteAtEnd(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	env.GetLayer(core.EnvLayerSession).SetInt("sys.execute-wait-sec.at-end", argv.GetInt("seconds"))
	return currCmdIdx, true
}

func DbgBreakAtEnd(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cc.BreakPoints.SetAtEnd(true)
	if 1 == len(flow.Cmds)-1 {
		env.GetLayer(core.EnvLayerSession).SetBool("display.one-cmd", true)
	}
	return currCmdIdx, true
}

func DbgBreakBefore(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	listSep := env.GetRaw("strs.list-sep")
	cmdList := strings.Split(argv.GetRaw("break-points"), listSep)
	cc.BreakPoints.SetBefores(cc, env, cmdList)
	return currCmdIdx, true
}

func DbgBreakAfter(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	listSep := env.GetRaw("strs.list-sep")
	cmdList := strings.Split(argv.GetRaw("break-points"), listSep)
	cc.BreakPoints.SetAfters(cc, env, cmdList)
	return currCmdIdx, true
}

func DbgBreakClean(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	cc.BreakPoints.Clean(cc, env)
	return currCmdIdx, true
}

func DbgBreakStatus(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	bps := cc.BreakPoints

	cc.Screen.Print(display.ColorHelp("settings:", env) + "\n")
	if bps.IsEmpty() {
		cc.Screen.Print("    (none)\n")
	}

	if bps.AtEnd {
		cc.Screen.Print(fmt.Sprintf(display.ColorTip("    break at end:\n", env)+"        %v\n", bps.AtEnd))
	}

	if len(bps.Befores) != 0 {
		cc.Screen.Print(fmt.Sprintf(display.ColorTip("    break before one of the commands:\n", env)))
		for k, _ := range bps.Befores {
			k = display.ColorCmd("["+k+"]", env)
			cc.Screen.Print(fmt.Sprintf("        %s\n", k))
		}
	}

	if len(bps.Afters) != 0 {
		cc.Screen.Print(fmt.Sprintf(display.ColorTip("    break when one of the commands finishs:\n", env)))
		for k, _ := range bps.Afters {
			k = display.ColorCmd("["+k+"]", env)
			cc.Screen.Print(fmt.Sprintf("        %s\n", k))
		}
	}

	cc.Screen.Print(display.ColorHelp("status:", env) + "\n")
	keys := []string{
		"sys.step-by-step",
		"sys.breakpoint.here.now",
		"sys.breakpoint.status.step-in",
		"sys.breakpoint.status.step-out",
		"sys.interact.leaving",
		"sys.interact.inside",
	}
	for _, k := range keys {
		cc.Screen.Print(fmt.Sprintf("    %s %s %v\n",
			display.ColorKey(k, env), display.ColorSymbol("=", env), env.GetBool(k)))
	}

	return currCmdIdx, true
}

func DbgInteractLeave(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	env = env.GetLayer(core.EnvLayerSession)
	env.SetBool("sys.interact.leaving", true)
	return currCmdIdx, true
}

func DbgInteract(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	InteractiveMode(cc, env, "e")
	return currCmdIdx, true
}

func SysSetExtExecutor(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	ext := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "ext")
	exe := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "executor")
	env.GetLayer(core.EnvLayerSession).Set("sys.ext.exec."+ext, exe)
	return currCmdIdx, true
}
