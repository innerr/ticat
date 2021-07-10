package builtin

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func Sleep(argv core.ArgVals, _ *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
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
	fmt.Printf(".zzZZ ")
	secs := int(dur.Seconds())
	for i := 0; i < secs; i++ {
		fmt.Printf(".")
		time.Sleep(time.Second)
	}
	fmt.Printf(" *\\O/*\n")
	return true
}

func Time(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	key := argv.GetRaw("write-to-key")
	env = env.GetLayer(core.EnvLayerSession)
	env.SetInt(key, int(time.Now().Unix()))
	return true
}

func MockStub(_ core.ArgVals, _ *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	return true
}

func Dummy(_ core.ArgVals, cc *core.Cli, _ *core.Env, _ core.ParsedCmd) bool {
	cc.Screen.Print("dummy command here\n")
	return true
}

func QuietDummy(_ core.ArgVals, cc *core.Cli, _ *core.Env, _ core.ParsedCmd) bool {
	cc.Screen.Print("quiet dummy command here\n")
	return true
}

func PowerDummy(
	_ core.ArgVals,
	cc *core.Cli,
	_ *core.Env,
	_ *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cc.Screen.Print("power dummy command here\n")
	return currCmdIdx, true
}

func PriorityPowerDummy(
	_ core.ArgVals,
	cc *core.Cli,
	_ *core.Env,
	_ *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cc.Screen.Print("priority power dummy command here\n")
	return currCmdIdx, true
}
