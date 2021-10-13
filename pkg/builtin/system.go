package builtin

import (
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

	cmdStr := argv.GetRaw("command")
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return currCmdIdx, false
	}
	return currCmdIdx, true
}
