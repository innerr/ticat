package builtin

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func Sleep(argv core.ArgVals, _ *core.Cli, env *core.Env) bool {
	durStr := argv.GetRaw("duration")

	// Default unit is 's'
	_, err := strconv.ParseFloat(durStr, 64)
	if err == nil {
		durStr += "s"
	}

	dur, err := time.ParseDuration(durStr)
	if err != nil {
		fmt.Printf("[ERR] time string '%s' parse failed: %v\n", durStr, err)
		return false
	}
	time.Sleep(dur)
	return true
}

func MockStub(_ core.ArgVals, _ *core.Cli, env *core.Env) bool {
	return true
}

func Dummy(_ core.ArgVals, _ *core.Cli, env *core.Env) bool {
	fmt.Println("Dummy cmd here")
	return true
}
