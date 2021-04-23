package main

import (
	"os"

	"github.com/pingcap/ticat/pkg/cli"
	"github.com/pingcap/ticat/pkg/cli/builtin"
)

func main() {
	env := cli.GenEnvFromStdin()
	succeeded := cli.NewCli(
		builtin.RegisterBuiltin,
		builtin.LoadBuiltinEnv).Execute(env, os.Args[1:]...)
	if !succeeded {
		os.Exit(1)
	}
}
