package builtin

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func getAndCheckArg(argv core.ArgVals, env *core.Env, cmd core.ParsedCmd, arg string) string {
	val := argv.GetRaw(arg)
	if len(val) == 0 {
		panic(core.NewCmdError(cmd, "arg '"+arg+"' is empty"))
	}
	return val
}

func isOsCmdExists(cmd string) bool {
	path, err := exec.LookPath(cmd)
	return err == nil && len(path) > 0
}

func osRemoveDir(path string, cmd core.ParsedCmd) {
	path = strings.TrimSpace(path)
	if len(path) <= 1 {
		panic(core.WrapCmdError(cmd, fmt.Errorf("removing path '%v', looks not right", path)))
	}
	err := os.RemoveAll(path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		panic(core.WrapCmdError(cmd, fmt.Errorf("remove repo '%s' failed: %v", path, err)))
	}
}
