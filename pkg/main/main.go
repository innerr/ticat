package main

import (
	"os"

	"github.com/pingcap/ticat/pkg/cli"
	"github.com/pingcap/ticat/pkg/cli/builtin"
)

func main() {
	// For more detail, in termial execute:
	// $> ticat desc: the-bootstrap-string
	bootstrap := "B.E.L.L:B.E.L.R:B.M.L.L"

	env := cli.GenEnvFromStdin()
	succeeded := cli.NewCli(
		builtin.RegisterBuiltin,
		builtin.LoadBuiltinEnv).Execute(bootstrap, env, os.Args[1:]...)
	if !succeeded {
		os.Exit(1)
	}
}
