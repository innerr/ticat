package builtin

import (
	"fmt"
	"strings"
	"time"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
	"github.com/pingcap/ticat/pkg/utils"
)

func Sleep(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	durStr := utils.NormalizeDurStr(argv.GetRaw("duration"))
	dur, err := time.ParseDuration(durStr)
	if err != nil {
		panic(fmt.Errorf("[Sleep] time string '%s' parse failed: %v\n", durStr, err))
	}
	secs := int(dur.Seconds())
	if secs == 0 {
		return currCmdIdx, true
	}

	for i := 0; i < secs; i++ {
		if i%60 == 0 && i != 0 && i+1 != secs {
			cc.Screen.Print("\n")
		}
		if i%60 == 0 && i+1 != secs {
			cc.Screen.Print(".zzZZ ")
		}
		cc.Screen.Print(".")
		time.Sleep(time.Second)
	}
	cc.Screen.Print("\n")
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
	panic(core.NewCmdError(flow.Cmds[currCmdIdx], "this is a specified-error-type panic test command"))
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

	return currCmdIdx, true
}

func Dummy(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cc.Screen.Print("dummy command here\n")
	return currCmdIdx, true
}

func EnvOpCmds() []core.EnvOpCmd {
	return []core.EnvOpCmd{
		core.EnvOpCmd{
			ResetSessionEnv,
			func(checker *core.EnvOpsChecker, argv core.ArgVals, env *core.Env) {
				if checker != nil {
					checker.Reset()
				}
				env.GetLayer(core.EnvLayerSession).Clear(false)
			}},
		core.EnvOpCmd{
			ResetLocalEnv,
			func(checker *core.EnvOpsChecker, argv core.ArgVals, env *core.Env) {
				if checker != nil {
					checker.Reset()
				}
				env.GetLayer(core.EnvLayerSession).Clear(false)
			}},
		core.EnvOpCmd{RemoveEnvValAndSaveToLocal,
			func(checker *core.EnvOpsChecker, argv core.ArgVals, env *core.Env) {
				key := argv.GetRaw("key")
				if checker != nil {
					checker.RemoveKeyStat(key)
				}
				env.DeleteEx(key, core.EnvLayerDefault)
			}},
		core.EnvOpCmd{MarkTime,
			func(checker *core.EnvOpsChecker, argv core.ArgVals, env *core.Env) {
				key := argv.GetRaw("write-to-key")
				if checker != nil {
					checker.SetKeyWritten(key)
				}
				env.GetLayer(core.EnvLayerSession).
					Set(key, "<dummy-fake-key-for-env-op-check-only-from-MarkTime>")
			}},
		core.EnvOpCmd{MapEnvKeyValueToAnotherKey,
			func(checker *core.EnvOpsChecker, argv core.ArgVals, env *core.Env) {
				key := argv.GetRaw("dest-key")
				if checker != nil {
					checker.SetKeyWritten(key)
				}
				env.GetLayer(core.EnvLayerSession).
					Set(key, "<dummy-fake-key-for-env-op-check-only-from-MapEnvKeyValueToAnotherKey>")
			}},
		core.EnvOpCmd{SetEnvKeyValue,
			func(checker *core.EnvOpsChecker, argv core.ArgVals, env *core.Env) {
				key := argv.GetRaw("key")
				if checker != nil {
					checker.SetKeyWritten(key)
				}
				env.GetLayer(core.EnvLayerSession).
					Set(key, "<dummy-fake-key-for-env-op-check-only-from-SetEnvKeyValue>")
			}},
		core.EnvOpCmd{AddEnvKeyValue,
			func(checker *core.EnvOpsChecker, argv core.ArgVals, env *core.Env) {
				key := argv.GetRaw("key")
				if checker != nil {
					checker.SetKeyWritten(key)
				}
				env.GetLayer(core.EnvLayerSession).
					Set(key, "<dummy-fake-key-for-env-op-check-only-from-AddEnvKeyValue>")
			}},
		core.EnvOpCmd{RemoveEnvValNotSave,
			func(checker *core.EnvOpsChecker, argv core.ArgVals, env *core.Env) {
				key := argv.GetRaw("key")
				if checker != nil {
					checker.RemoveKeyStat(key)
				}
				env.GetLayer(core.EnvLayerSession).Delete(key)
			}},
		core.EnvOpCmd{RemoveEnvValAndSaveToLocal,
			func(checker *core.EnvOpsChecker, argv core.ArgVals, env *core.Env) {
				key := argv.GetRaw("key")
				if checker != nil {
					checker.RemoveKeyStat(key)
				}
				env.GetLayer(core.EnvLayerSession).Delete(key)
			}},
	}
}

func Selftest(argv core.ArgVals, cc *core.Cli, env *core.Env) (flow []string, masks []*core.ExecuteMask, ok bool) {
	tag := argv.GetRaw("tag")
	src := argv.GetRaw("match-source")
	filter := argv.GetRaw("filter-source")
	parallel := argv.GetBool("parallel")

	result := []*core.CmdTree{}
	findAllCmdsByTag(tag, src, filter, cc.Cmds, &result)

	if len(result) == 0 {
		if len(src) == 0 {
			if len(filter) == 0 {
				panic(fmt.Errorf("no selftest command with tag '%s'", tag))
			} else {
				panic(fmt.Errorf("no selftest command with tag '%s' and source not match '%s'", tag, filter))
			}
		} else {
			if len(filter) == 0 {
				panic(fmt.Errorf("no selftest command with tag '%s' and source match '%s'", tag, src))
			} else {
				panic(fmt.Errorf("no selftest command with tag '%s' and source match '%s' but not match '%s'", tag, src, filter))
			}
		}
		return
	}
	ok = true
	if len(result) != 1 && !parallel {
		flow = append(flow, "flow.forest-mode")
	}
	trivialMark := env.GetRaw("strs.trivial-mark")
	for _, it := range result {
		cmd := trivialMark + it.DisplayPath()
		if parallel {
			cmd += " %delay=0"
		}
		flow = append(flow, cmd)
	}
	if parallel {
		flow = append(flow, "background.wait")
	}
	return
}

func Repeat(argv core.ArgVals, cc *core.Cli, env *core.Env) (flow []string, masks []*core.ExecuteMask, ok bool) {
	cmd := argv.GetRaw("cmd")
	if len(cmd) == 0 {
		panic(fmt.Errorf("arg 'cmd' is empty"))
	}
	times := argv.GetInt("times")
	if times <= 0 {
		panic(fmt.Errorf("arg 'times' is invalid value '%d'", times))
	}

	for i := 0; i < times; i++ {
		flow = append(flow, cmd)
	}
	ok = true
	return
}

func findAllCmdsByTag(tag string, src string, filter string, curr *core.CmdTree, output *[]*core.CmdTree) {
	if curr.MatchTags(tag) &&
		(len(src) == 0 || strings.Index(curr.Source(), src) >= 0) &&
		(len(filter) == 0 || strings.Index(curr.Source(), filter) < 0) {
		*output = append(*output, curr)
	}
	for _, name := range curr.SubNames() {
		findAllCmdsByTag(tag, src, filter, curr.GetSub(name), output)
	}
}
