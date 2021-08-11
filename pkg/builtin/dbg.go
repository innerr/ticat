package builtin

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func DbgEcho(argv core.ArgVals, cc *core.Cli, _ *core.Env, flow []core.ParsedCmd) bool {
	for _, input := range gatherInputsFromFlow(flow[1:]) {
		cc.Screen.Print(fmt.Sprintf("[flow] %v\n", input))
	}
	cc.Screen.Print(fmt.Sprintf("%v\n", argv.GetRaw("message")))
	return true
}

func DbgExecBash(_ core.ArgVals, cc *core.Cli, _ *core.Env, _ []core.ParsedCmd) bool {
	cmd := exec.Command("bash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	return true
}

func DbgDelayExecute(argv core.ArgVals, cc *core.Cli, env *core.Env, _ []core.ParsedCmd) bool {
	env.GetLayer(core.EnvLayerSession).SetInt("sys.execute-delay-sec", argv.GetInt("seconds"))
	return true
}
