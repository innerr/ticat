package builtin

import (
	"fmt"
	"strings"
	"time"

	"github.com/innerr/ticat/pkg/cli/display"
	"github.com/innerr/ticat/pkg/core/model"
	"github.com/innerr/ticat/pkg/utils"
)

func Version(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cc.Screen.Print(env.GetRaw("sys.version") + " " + env.GetRaw("sys.dev.name") + "\n")

	sep := cc.Cmds.Strs.ListSep

	init := env.GetRaw("sys.hub.init-repo")
	if len(init) != 0 {
		for _, it := range strings.Split(init, sep) {
			cc.Screen.Print("- init mod repo: " + it + "\n")
		}
	}

	integrated := env.GetRaw("sys.mods.integrated")
	if len(integrated) != 0 {
		for _, it := range strings.Split(integrated, sep) {
			cc.Screen.Print("- integrated: " + it + "\n")
		}
	}

	return currCmdIdx, true
}

func Sleep(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
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
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	key := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "write-to-key")
	env = env.GetLayer(model.EnvLayerSession)
	val := fmt.Sprintf("%d", int(time.Now().Unix()))
	env.Set(key, val)
	cc.Screen.Print(display.ColorKey(key, env) + display.ColorSymbol(" = ", env) + val + "\n")
	return currCmdIdx, true
}

func TimerBegin(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	begin := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "begin-key")
	env = env.GetLayer(model.EnvLayerSession)
	val := fmt.Sprintf("%d", int(time.Now().Unix()))
	env.Set(begin, val)
	cc.Screen.Print(display.ColorKey(begin, env) + display.ColorSymbol(" = ", env) + val + "\n")
	return currCmdIdx, true
}

func TimerElapsed(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	beginKey := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "begin-key")
	begin := env.GetInt(beginKey)
	now := int(time.Now().Unix())

	elapsedKey := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "write-to-key")
	env = env.GetLayer(model.EnvLayerSession)
	elapsed := now - begin
	env.SetInt(elapsedKey, elapsed)
	cc.Screen.Print(display.ColorKey(elapsedKey, env) + display.ColorSymbol(" = ", env) +
		fmt.Sprintf("%d\n", elapsed))
	return currCmdIdx, true
}

func DbgPanic(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	panic(fmt.Errorf("this is a panic test command"))
}

func DbgPanicCmdError(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	panic(model.NewCmdError(flow.Cmds[currCmdIdx], "this is a specified-error-type panic test command"))
}

func DbgError(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	return currCmdIdx, false
}

func Noop(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return currCmdIdx, true
}

func Dummy(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cc.Screen.Print("dummy command here\n")
	return currCmdIdx, true
}

func EnvOpCmds() []model.EnvOpCmd {
	return []model.EnvOpCmd{
		{
			Func: ResetSessionEnv,
			Action: func(checker *model.EnvOpsChecker, argv model.ArgVals, env *model.Env) {
				if checker != nil {
					checker.Reset()
				}
				env.GetLayer(model.EnvLayerSession).Clear(false)
			}},
		{
			Func: ResetLocalEnv,
			Action: func(checker *model.EnvOpsChecker, argv model.ArgVals, env *model.Env) {
				if checker != nil {
					checker.Reset()
				}
				env.GetLayer(model.EnvLayerSession).Clear(false)
			}},
		{
			Func: RemoveEnvValAndSaveToLocal,
			Action: func(checker *model.EnvOpsChecker, argv model.ArgVals, env *model.Env) {
				key := argv.GetRaw("key")
				if checker != nil {
					checker.RemoveKeyStat(key)
				}
				env.DeleteEx(key, model.EnvLayerDefault)
			}},
		{
			Func: MarkTime,
			Action: func(checker *model.EnvOpsChecker, argv model.ArgVals, env *model.Env) {
				key := argv.GetRaw("write-to-key")
				if checker != nil {
					checker.SetKeyWritten(key)
				}
				env.GetLayer(model.EnvLayerSession).
					SetIfEmpty(key, "<dummy-fake-key-for-env-op-check-only-from-MarkTime>")
			}},
		{
			Func: MapEnvKeyValueToAnotherKey,
			Action: func(checker *model.EnvOpsChecker, argv model.ArgVals, env *model.Env) {
				key := argv.GetRaw("dest-key")
				if checker != nil {
					checker.SetKeyWritten(key)
				}
				env.GetLayer(model.EnvLayerSession).
					SetIfEmpty(key, "<dummy-fake-key-for-env-op-check-only-from-MapEnvKeyValueToAnotherKey>")
			}},
		{
			Func: SetEnvKeyValue,
			Action: func(checker *model.EnvOpsChecker, argv model.ArgVals, env *model.Env) {
				key := argv.GetRaw("key")
				if checker != nil {
					checker.SetKeyWritten(key)
				}
				env.GetLayer(model.EnvLayerSession).
					SetIfEmpty(key, "<dummy-fake-key-for-env-op-check-only-from-SetEnvKeyValue>")
			}},
		{
			Func: AddEnvKeyValue,
			Action: func(checker *model.EnvOpsChecker, argv model.ArgVals, env *model.Env) {
				key := argv.GetRaw("key")
				if checker != nil {
					checker.SetKeyWritten(key)
				}
				env.GetLayer(model.EnvLayerSession).
					SetIfEmpty(key, "<dummy-fake-key-for-env-op-check-only-from-AddEnvKeyValue>")
			}},
		{
			Func: RemoveEnvValNotSave,
			Action: func(checker *model.EnvOpsChecker, argv model.ArgVals, env *model.Env) {
				key := argv.GetRaw("key")
				if checker != nil {
					checker.RemoveKeyStat(key)
				}
				env.GetLayer(model.EnvLayerSession).Delete(key)
			}},
		{
			Func: RemoveEnvValAndSaveToLocal,
			Action: func(checker *model.EnvOpsChecker, argv model.ArgVals, env *model.Env) {
				key := argv.GetRaw("key")
				if checker != nil {
					checker.RemoveKeyStat(key)
				}
				env.GetLayer(model.EnvLayerSession).Delete(key)
			}},
		{
			Func:   EnvLoadFromSnapshot,
			Action: opCheckEnvLoadFromSnapshot},
		{
			Func:   EnvLoadNonExistFromSnapshot,
			Action: opCheckEnvLoadFromSnapshot},
	}
}

func opCheckEnvLoadFromSnapshot(checker *model.EnvOpsChecker, argv model.ArgVals, env *model.Env) {
	name := argv.GetRaw("snapshot-name")
	if len(name) == 0 {
		return
	}

	path := getEnvSnapshotPath(env, name)

	sep := env.GetRaw("strs.env-kv-sep")
	delMark := env.GetRaw("strs.env-del-all-mark")

	loaded := model.NewEnv()
	model.LoadEnvFromFile(loaded, path, sep, delMark)

	env = env.GetLayer(model.EnvLayerSession)
	for key := range loaded.FlattenAll() {
		if checker != nil {
			checker.SetKeyWritten(key)
		}
		env.SetIfEmpty(key, "<dummy-fake-key-for-env-op-check-only-from-EnvLoadFromSnapshot>")
	}
}

func Selftest(argv model.ArgVals, cc *model.Cli, env *model.Env) (flow []string, masks []*model.ExecuteMask, ok bool) {
	tag := argv.GetRaw("tag")
	src := argv.GetRaw("match-source")
	filter := argv.GetRaw("filter-source")
	parallel := argv.GetBool("parallel")

	result := []*model.CmdTree{}
	findAllCmdsByTag(tag, src, filter, cc.Cmds, &result)

	if len(result) == 0 {
		/*
			// TODO: use PrintTitleError instead of panic. it's not errors
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
		*/
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

func Repeat(argv model.ArgVals, cc *model.Cli, env *model.Env) (flow []string, masks []*model.ExecuteMask, ok bool) {
	cmd := argv.GetRaw("cmd")
	if len(cmd) == 0 {
		panic(fmt.Errorf("arg 'cmd' is empty"))
	}
	times := argv.GetInt("times")
	if times <= 0 {
		panic(fmt.Errorf("arg 'times' is invalid value '%d'", times))
	}

	trivialMark := env.GetRaw("strs.trivial-mark")
	cmd = trivialMark + cmd
	for i := 0; i < times; i++ {
		flow = append(flow, cmd)
	}
	ok = true
	return
}

func findAllCmdsByTag(tag string, src string, filter string, curr *model.CmdTree, output *[]*model.CmdTree) {
	if curr.MatchTags(tag) &&
		(len(src) == 0 || strings.Index(curr.Source(), src) >= 0) &&
		(len(filter) == 0 || strings.Index(curr.Source(), filter) < 0) {
		*output = append(*output, curr)
	}
	for _, name := range curr.SubNames() {
		findAllCmdsByTag(tag, src, filter, curr.GetSub(name), output)
	}
}
