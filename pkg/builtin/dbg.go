package builtin

import (
	"fmt"
	"strings"

	"github.com/mattn/go-tty"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func DbgEcho(argv core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	cc.Screen.Print(fmt.Sprintf("echo msg: '%s'\n", argv.GetRaw("message")))
	return true
}

func DbgReadFromTty(_ core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	tty, err := tty.Open()
	if err != nil {
		panic(fmt.Errorf("failed to open tty: %s", err))
	}
	cc.Screen.Print("Enter: ")
	text, err := tty.ReadString()
	if err != nil {
		panic(fmt.Errorf("failed to read from tty: %s", err))
	}
	cc.Screen.Print("Got: ")
	cc.Screen.Print(strings.TrimSpace(text) + "\n")
	return true
}
