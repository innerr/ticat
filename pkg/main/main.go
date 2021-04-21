package main

import (
	"os"

	"github.com/pingcap/ticat/pkg/cli"
	"github.com/pingcap/ticat/pkg/cli/builtin"
)

func main() {
	preparation := `
		builtin.env.load.local:
		builtin.env.load.runtime:
		builtin.mod.load.local:
	`

	succeeded := cli.NewCli(
		builtin.RegisterBuiltin,
		builtin.LoadBuiltinEnv).Execute(preparation, os.Args[1:]...)
	if !succeeded {
		os.Exit(1)
	}
}
