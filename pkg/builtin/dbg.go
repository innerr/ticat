package builtin

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	//"github.com/pingcap/ticat/pkg/cli/display"
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

func DbgBreakAtBegin(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	cc.BreakPoints.SetAtBegin(true)
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
