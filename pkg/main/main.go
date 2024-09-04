package main

import (
	"os"

	"github.com/pingcap/ticat/pkg/ticat"
)

func main() {
	ticat := ticat.NewTiCat()
	ticat.RunCli(os.Args[1:]...)
}
