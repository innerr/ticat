package main

import (
	"os"

	"github.com/innerr/ticat/pkg/ticat"
)

func main() {
	ticat := ticat.NewTiCat()
	ticat.RunCli(os.Args[1:]...)
}
