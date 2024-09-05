package builtin

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/innerr/ticat/pkg/core/model"
)

func ExecCmds(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	cmdStr := argv.GetRaw("command")
	if cmdStr == "" {
		panic(model.NewCmdError(flow.Cmds[currCmdIdx],
			"can't execute null os command"))
	}
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(model.NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("execute os command '%s' failed: %s", cmdStr, err.Error())))
	}
	return currCmdIdx, true
}
