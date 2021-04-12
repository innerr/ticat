package main

import (
	"os"
	"strings"

	"github.com/pingcap/ticat/pkg/cli"
)

func main() {
	//preparation := "builtin/env/load/local : builtin/mod/load/local : builtin/greeting/dev"
	preparation := "builtin env load local : builtin mod load local : builtin greeting dev"
	hub := cli.NewHub()
	if !hub.Execute(strings.Split(preparation, " ")) || !hub.Execute(os.Args[1:]) {
		os.Exit(1)
	}
}
