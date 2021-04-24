package builtin

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pingcap/ticat/pkg/cli"
)

func Sleep(argv cli.ArgVals, _ *cli.Cli, env *cli.Env) bool {
	durStr := argv.GetRaw("duration")

	// Default unit is 's'
	_, err := strconv.ParseFloat(durStr, 64)
	if err == nil {
		durStr += "s"
	}

	dur, err := time.ParseDuration(durStr)
	if err != nil {
		fmt.Printf("[ERR] %v\n", err)
		return false
	}
	time.Sleep(dur)
	return true
}

func MockStub(_ cli.ArgVals, _ *cli.Cli, env *cli.Env) bool {
	return true
}

func Dummy(_ cli.ArgVals, _ *cli.Cli, env *cli.Env) bool {
	fmt.Println("Dummy cmd here")
	return true
}
