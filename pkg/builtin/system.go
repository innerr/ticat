package builtin

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func ExecCmds(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	cmdStr := argv.GetRaw("command")
	if cmdStr == "" {
		panic(core.NewCmdError(flow.Cmds[currCmdIdx],
			"can't execute null os command"))
	}
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(core.NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("execute os command '%s' failed: %s", cmdStr, err.Error())))
	}
	return currCmdIdx, true
}
