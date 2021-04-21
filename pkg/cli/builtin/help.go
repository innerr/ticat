package builtin

import (
	"fmt"

	"github.com/pingcap/ticat/pkg/cli"
)

func GlobalHelp(_ cli.ArgVals, _ *cli.Cli, env *cli.Env, cmds []cli.ParsedCmd,
	currCmdIdx int) ([]cli.ParsedCmd, int, bool) {

	fmt.Println("TODO: global help")
	return nil, 0, true
}
