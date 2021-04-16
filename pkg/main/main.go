package main

import (
	"os"

	"github.com/pingcap/ticat/pkg/cli"
)

func main() {
	preparation := "builtin/env/load/local : builtin/mod/load/local : builtin/greeting/dev"

	succeeded := cli.Execute(preparation, os.Args[1:]...)
	if !succeeded {
		os.Exit(1)
	}
}
