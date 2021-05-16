package builtin

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func DbgEcho(argv core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	cc.Screen.Print(fmt.Sprintf("echo msg: '%s'\n", argv.GetRaw("message")))
	return true
}

func DbgExecBash(_ core.ArgVals, cc *core.Cli, _ *core.Env) bool {
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

func DbgStepOn(_ core.ArgVals, cc *core.Cli, env *core.Env) bool {
	env.GetLayer(core.EnvLayerSession).SetBool("sys.step-by-step", true)
	return true
}

func DbgStepOff(_ core.ArgVals, cc *core.Cli, env *core.Env) bool {
	env.GetLayer(core.EnvLayerSession).SetBool("sys.step-by-step", false)
	return true
}
