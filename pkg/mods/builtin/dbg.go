package builtin

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/innerr/ticat/pkg/cli/display"
	"github.com/innerr/ticat/pkg/core/model"
)

func DbgEcho(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	msg := argv.GetRaw("message")
	color := argv.GetRaw("color")

	rendered, _ := model.RenderTemplateStrLines([]string{msg}, "echo", flow.Cmds[currCmdIdx].LastCmd(), model.ArgVals{}, env, true)
	str := rendered[0]
	str = display.ColorStrByName(str, color, env)

	cc.Screen.Print(fmt.Sprintf("%v\n", str))
	return currCmdIdx, nil
}

func DbgEchoLn(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	cc.Screen.Print("\n")
	return currCmdIdx, nil
}

func DbgExecBash(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	cmd := exec.Command("bash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return currCmdIdx, model.WrapCmdError(flow.Cmds[currCmdIdx], err)
	}

	return currCmdIdx, nil
}

func DbgWaitSecExecute(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	env.GetLayer(model.EnvLayerSession).SetInt("sys.execute-wait-sec", argv.GetInt("seconds"))
	return currCmdIdx, nil
}

func DbgWaitSecExecuteAtEnd(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	env.GetLayer(model.EnvLayerSession).SetInt("sys.execute-wait-sec.at-end", argv.GetInt("seconds"))
	return currCmdIdx, nil
}

func DbgBreakAtEnd(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	cc.BreakPoints.SetAtEnd(true)
	if 1 == len(flow.Cmds)-1 {
		env.GetLayer(model.EnvLayerSession).SetBool("display.one-cmd", true)
	}
	return currCmdIdx, nil
}

func DbgBreakBefore(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	listSep := env.GetRaw("strs.list-sep")
	cmdList := strings.Split(argv.GetRaw("break-points"), listSep)
	cc.BreakPoints.SetBefores(cc, env, cmdList)
	return currCmdIdx, nil
}

func DbgBreakAfter(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	listSep := env.GetRaw("strs.list-sep")
	cmdList := strings.Split(argv.GetRaw("break-points"), listSep)
	cc.BreakPoints.SetAfters(cc, env, cmdList)
	return currCmdIdx, nil
}

func DbgBreakClean(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	cc.BreakPoints.Clean(cc, env)
	return currCmdIdx, nil
}

func DbgBreakStatus(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

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
		for k := range bps.Befores {
			k = display.ColorCmd("["+k+"]", env)
			cc.Screen.Print(fmt.Sprintf("        %s\n", k))
		}
	}

	if len(bps.Afters) != 0 {
		cc.Screen.Print(fmt.Sprintf(display.ColorTip("    break when one of the commands finishs:\n", env)))
		for k := range bps.Afters {
			k = display.ColorCmd("["+k+"]", env)
			cc.Screen.Print(fmt.Sprintf("        %s\n", k))
		}
	}

	cc.Screen.Print(display.ColorHelp("status:", env) + "\n")
	keys := []string{
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

	return currCmdIdx, nil
}

func DbgInteractLeave(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	env = env.GetLayer(model.EnvLayerSession)
	env.SetBool("sys.interact.leaving", true)
	return currCmdIdx, nil
}

func DbgInteract(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	InteractiveMode(cc, env, "e")
	return currCmdIdx, nil
}

func SysSetExtExecutor(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	ext, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "ext")
	if err != nil {
		return currCmdIdx, err
	}
	exe, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "executor")
	if err != nil {
		return currCmdIdx, err
	}
	env.GetLayer(model.EnvLayerSession).Set("sys.ext.exec."+ext, exe)
	return currCmdIdx, nil
}
