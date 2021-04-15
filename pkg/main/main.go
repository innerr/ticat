package main

import (
	"os"

	"github.com/pingcap/ticat/pkg/hub"
)

func main() {
	preparation := "builtin/env/load/local : builtin/mod/load/local : builtin/greeting/dev"

	succeeded := hub.Execute(preparation, os.Args[1:])
	if !succeeded {
		os.Exit(1)
	}
}
