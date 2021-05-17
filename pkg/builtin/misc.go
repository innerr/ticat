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
		fmt.Printf("[Sleep] time string '%s' parse failed: %v\n", durStr, err)
		return false
	}
	time.Sleep(dur)
	return true
}

func MockStub(_ core.ArgVals, _ *core.Cli, env *core.Env) bool {
	return true
}

func Dummy(_ core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	cc.Screen.Print("dummy cmd here\n")
	return true
}

func QuietDummy(_ core.ArgVals, cc *core.Cli, _ *core.Env) bool {
	cc.Screen.Print("quiet dummy cmd here\n")
	return true
}

func PowerDummy(
	_ core.ArgVals,
	cc *core.Cli,
	_ *core.Env,
	_ *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cc.Screen.Print("power dummy cmd here\n")
	return currCmdIdx, true
}

func PriorityPowerDummy(
	_ core.ArgVals,
	cc *core.Cli,
	_ *core.Env,
	_ *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cc.Screen.Print("priority power dummy cmd here\n")
	return currCmdIdx, true
}
