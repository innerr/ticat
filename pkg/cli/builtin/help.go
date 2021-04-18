package builtin

import (
	"fmt"

	"github.com/pingcap/ticat/pkg/cli"
)

func GlobalHelp(_ *cli.Cli, env *cli.Env, cmds []cli.ParsedCmd,
	currCmdIdx int) (modified []cli.ParsedCmd, succeeded bool) {

	fmt.Println("TODO: global help")
	return nil, true
}
