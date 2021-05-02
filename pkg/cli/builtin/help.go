package builtin

import (
	"fmt"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func GlobalHelp(_ core.ArgVals, _ *core.Cli, env *core.Env, cmds []core.ParsedCmd,
	currCmdIdx int) ([]core.ParsedCmd, int, bool) {

	fmt.Println("TODO: global help")
	return nil, 0, true
}
