package builtin

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func DbgEcho(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

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
