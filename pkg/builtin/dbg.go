package builtin

import (
	"bufio"
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

	arg := argv.GetRaw("message")
	if len(arg) == 0 {
		cc.Screen.Print("(arg 'message' is empty)")
		return currCmdIdx, true
	}
	cc.Screen.Print(fmt.Sprintf("%v\n", arg))
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

func DbgDelayExecute(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	env.GetLayer(core.EnvLayerSession).SetInt("sys.execute-delay-sec", argv.GetInt("seconds"))
	return currCmdIdx, true
}

func DbgDelayExecuteAtEnd(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	env.GetLayer(core.EnvLayerSession).SetInt("sys.execute-delay-sec.at-end", argv.GetInt("seconds"))
	return currCmdIdx, true
}

func DbgBreakAtBegin(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	cc.BreakPoints.SetAtBegin(true)
	if 1 == len(flow.Cmds)-1 {
		env.GetLayer(core.EnvLayerSession).SetBool("display.one-cmd", true)
	}
	return currCmdIdx, true
}

func DbgBreakAtEnd(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
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

	InteractiveMode(cc.CopyForInteract(), env, "e")
	return currCmdIdx, true
}

func InteractiveMode(cc *core.Cli, env *core.Env, exitStr string) {
	selfName := env.GetRaw("strs.self-name")
	cc.Screen.Print(display.ColorExplain("", env) + display.ColorWarn(exitStr, env) +
		display.ColorExplain(": exit interactive mode\n", env))

	sessionEnv := env.GetLayer(core.EnvLayerSession)
	sessionEnv.SetBool("sys.interact.inside", true)

	cc = cc.CopyForInteract()
	buf := bufio.NewReader(os.Stdin)
	for {
		if env.GetBool("sys.interact.leaving") {
			break
		}
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
		executorSafeExecute("(interact)", cc, env, nil, core.FlowStrToStrs(line)...)
	}

	sessionEnv.GetLayer(core.EnvLayerSession).Delete("sys.interact.inside")
}

func executorSafeExecute(caller string, cc *core.Cli, env *core.Env, masks []*core.ExecuteMask, input ...string) {
	env = env.GetLayer(core.EnvLayerSession)
	stackDepth := env.GetRaw("sys.stack-depth")
	stack := env.GetRaw("sys.stack")

	defer func() {
		env.Set("sys.stack-depth", stackDepth)
		env.Set("sys.stack", stack)

		if !env.GetBool("sys.panic.recover") {
			return
		}
		if r := recover(); r != nil {
			display.PrintError(cc, env, r.(error))
		}
	}()

	cc.Executor.Execute(caller, false, cc, env, masks, input...)
}
