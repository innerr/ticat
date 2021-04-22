package builtin

import (
	"github.com/pingcap/ticat/pkg/cli"
)

func DbgDumpCmds(_ cli.ArgVals, cc *cli.Cli, env *cli.Env, cmds []cli.ParsedCmd,
	currCmdIdx int) ([]cli.ParsedCmd, int, bool) {
	cli.DumpCmdsEx(cc.Screen, env, cmds, cc.Parser.CmdPathSep())
	return nil, 0, true
}

func DbgDumpEnv(_ cli.ArgVals, cc *cli.Cli, env *cli.Env) bool {
	cli.DumpEnv(env)
	return true
}
