package builtin

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func DbgEcho(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	arg := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "message")
	rendered, _ := core.RenderTemplateStrLines([]string{arg}, "echo", flow.Cmds[currCmdIdx].LastCmd(), core.ArgVals{}, env, true)
	cc.Screen.Print(fmt.Sprintf("%v\n", rendered[0]))
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
