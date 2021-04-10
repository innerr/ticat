package main

import (
	"os"
	"strings"

	"github.com/pingcap/ticat/pkg/cli"
)

func main() {
	preparation := "builtin/env/load/local : builtin/mod/load/local : builtin/greeting/dev"
	executor := cli.NewExecutor()
	if !executor.Execute(strings.split(preparation)) || !executor.Execute(strings.split(os.Args[1:]) {
		os.Exit(1)
	}
}
