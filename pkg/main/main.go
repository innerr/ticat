package main

import (
	"os"

	"github.com/pingcap/ticat/pkg/cli"
	"github.com/pingcap/ticat/pkg/cli/builtin"
	"github.com/pingcap/ticat/pkg/cli/parser"
)

func main() {
	// For more detail, in termial execute:
	// $> ticat desc: the-bootstrap-string
	bootstrap := "B.E.L.R:B.E.L.L:B.M.L.L"

	env := cli.GenEnvFromStdin()
	succeeded := cli.NewCli(
		builtin.RegisterBuiltinMods,
		builtin.LoadDefaultEnv,
		parser.NewParser()).Execute(bootstrap, env, os.Args[1:]...)
	if !succeeded {
		os.Exit(1)
	}
}
