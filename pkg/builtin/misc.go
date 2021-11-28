package builtin

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func Sleep(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	durStr := argv.GetRaw("duration")

	// Default unit is 's'
	_, err := strconv.ParseFloat(durStr, 64)
	if err == nil {
		durStr += "s"
	}

	dur, err := time.ParseDuration(durStr)
	if err != nil {
		fmt.Printf("[Sleep] time string '%s' parse failed: %v\n", durStr, err)
		return currCmdIdx, false
	}
	secs := int(dur.Seconds())
	for i := 0; i < secs; i++ {
		if i%60 == 0 && i != 0 && i+1 != secs {
			fmt.Printf("\n")
		}
		if i%60 == 0 && i+1 != secs {
			fmt.Printf(".zzZZ ")
		}
		fmt.Printf(".")
		time.Sleep(time.Second)
	}
	fmt.Printf("\n")
	return currCmdIdx, true
}

func MarkTime(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	key := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "write-to-key")
	env = env.GetLayer(core.EnvLayerSession)
	val := fmt.Sprintf("%d", int(time.Now().Unix()))
	env.Set(key, val)
	cc.Screen.Print(display.ColorKey(key, env) + display.ColorSymbol(" = ", env) + val + "\n")
	return currCmdIdx, true
}

func TimerBegin(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	begin := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "begin-key")
	env = env.GetLayer(core.EnvLayerSession)
	val := fmt.Sprintf("%d", int(time.Now().Unix()))
	env.Set(begin, val)
	cc.Screen.Print(display.ColorKey(begin, env) + display.ColorSymbol(" = ", env) + val + "\n")
	return currCmdIdx, true
}

func TimerElapsed(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	beginKey := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "begin-key")
	begin := env.GetInt(beginKey)
	now := int(time.Now().Unix())

	elapsedKey := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "write-to-key")
	env = env.GetLayer(core.EnvLayerSession)
	elapsed := now - begin
	env.SetInt(elapsedKey, elapsed)
	cc.Screen.Print(display.ColorKey(elapsedKey, env) + display.ColorSymbol(" = ", env) +
		fmt.Sprintf("%d\n", elapsed))
	return currCmdIdx, true
}

func DbgPanic(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	panic(fmt.Errorf("this is a panic test command"))
}

func DbgPanicCmdError(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	panic(core.NewCmdError(flow.Cmds[currCmdIdx], "this is a specified-panic test command"))
}

func DbgError(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	return currCmdIdx, false
}

func Noop(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	return currCmdIdx, true
}

func Dummy(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	cc.Screen.Print("dummy command here\n")
	return currCmdIdx, true
}

func EnvOpCmds() []core.EnvOpCmd {
	return []core.EnvOpCmd{
		core.EnvOpCmd{
			ResetSessionEnv,
			func(checker *core.EnvOpsChecker, argv core.ArgVals) {
				checker.Reset()
			}},
		core.EnvOpCmd{
			ResetLocalEnv,
			func(checker *core.EnvOpsChecker, argv core.ArgVals) {
				checker.Reset()
			}},
		core.EnvOpCmd{RemoveEnvValAndSaveToLocal,
			func(checker *core.EnvOpsChecker, argv core.ArgVals) {
				checker.RemoveKeyStat(argv.GetRaw("key"))
			}},
	}
}
