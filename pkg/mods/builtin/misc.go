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
	currCmdIdx int) (int, error) {

	_ = cc.Screen.Print(env.GetRaw("sys.version") + " " + env.GetRaw("sys.dev.name") + "\n")

	sep := cc.Cmds.Strs.ListSep

	init := env.GetRaw("sys.hub.init-repo")
	if len(init) != 0 {
		for _, it := range strings.Split(init, sep) {
			_ = cc.Screen.Print("- init mod repo: " + it + "\n")
		}
	}

	integrated := env.GetRaw("sys.mods.integrated")
	if len(integrated) != 0 {
		for _, it := range strings.Split(integrated, sep) {
			_ = cc.Screen.Print("- integrated: " + it + "\n")
		}
	}

	return currCmdIdx, nil
}

func Sleep(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	durStr := utils.NormalizeDurStr(argv.GetRaw("duration"))
	dur, err := time.ParseDuration(durStr)
	if err != nil {
		return currCmdIdx, fmt.Errorf("[Sleep] time string '%s' parse failed: %v", durStr, err)
	}
	secs := int(dur.Seconds())
	if secs == 0 {
		return currCmdIdx, nil
	}

	for i := 0; i < secs; i++ {
		if i%60 == 0 && i != 0 && i+1 != secs {
			_ = cc.Screen.Print("\n")
		}
		if i%60 == 0 && i+1 != secs {
			_ = cc.Screen.Print(".zzZZ ")
		}
		_ = cc.Screen.Print(".")
		time.Sleep(time.Second)
	}
	_ = cc.Screen.Print("\n")
	return currCmdIdx, nil
}

func MarkTime(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	key, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "write-to-key")
	if err != nil {
		return currCmdIdx, err
	}
	env = env.GetLayer(model.EnvLayerSession)
	val := fmt.Sprintf("%d", int(time.Now().Unix()))
	env.Set(key, val)
	_ = cc.Screen.Print(display.ColorKey(key, env) + display.ColorSymbol(" = ", env) + val + "\n")
	return currCmdIdx, nil
}

func TimerBegin(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	begin, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "begin-key")
	if err != nil {
		return currCmdIdx, err
	}
	env = env.GetLayer(model.EnvLayerSession)
	val := fmt.Sprintf("%d", int(time.Now().Unix()))
	env.Set(begin, val)
	_ = cc.Screen.Print(display.ColorKey(begin, env) + display.ColorSymbol(" = ", env) + val + "\n")
	return currCmdIdx, nil
}

func TimerElapsed(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	beginKey, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "begin-key")
	if err != nil {
		return currCmdIdx, err
	}
	begin := env.GetInt(beginKey)
	now := int(time.Now().Unix())

	elapsedKey, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "write-to-key")
	if err != nil {
		return currCmdIdx, err
	}
	env = env.GetLayer(model.EnvLayerSession)
	elapsed := now - begin
	env.SetInt(elapsedKey, elapsed)
	_ = cc.Screen.Print(display.ColorKey(elapsedKey, env) + display.ColorSymbol(" = ", env) +
		fmt.Sprintf("%d\n", elapsed))
	return currCmdIdx, nil
}

func DbgPanic(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	return currCmdIdx, fmt.Errorf("this is a panic test command")
}

func DbgPanicCmdError(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	return currCmdIdx, model.NewCmdError(flow.Cmds[currCmdIdx], "this is a specified-error-type panic test command")
}

func DbgError(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	return currCmdIdx, fmt.Errorf("debug error")
}

func Noop(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	return currCmdIdx, nil
}

func Dummy(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	_ = cc.Screen.Print("dummy command here\n")
	return currCmdIdx, nil
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
	_ = model.LoadEnvFromFile(loaded, path, sep, delMark)

	env = env.GetLayer(model.EnvLayerSession)
	for key := range loaded.FlattenAll() {
		if checker != nil {
			checker.SetKeyWritten(key)
		}
		env.SetIfEmpty(key, "<dummy-fake-key-for-env-op-check-only-from-EnvLoadFromSnapshot>")
	}
}

func Selftest(argv model.ArgVals, cc *model.Cli, env *model.Env) (flow []string, masks []*model.ExecuteMask, err error) {
	tag := argv.GetRaw("tag")
	src := argv.GetRaw("match-source")
	filter := argv.GetRaw("filter-source")
	parallel := argv.GetBool("parallel")

	result := []*model.CmdTree{}
	findAllCmdsByTag(tag, src, filter, cc.Cmds, &result)

	if len(result) == 0 {
		return
	}

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

func Repeat(argv model.ArgVals, cc *model.Cli, env *model.Env) (flow []string, masks []*model.ExecuteMask, err error) {
	cmd := argv.GetRaw("cmd")
	if len(cmd) == 0 {
		return nil, nil, fmt.Errorf("arg 'cmd' is empty")
	}
	times := argv.GetInt("times")
	if times <= 0 {
		return nil, nil, fmt.Errorf("arg 'times' is invalid value '%d'", times)
	}

	trivialMark := env.GetRaw("strs.trivial-mark")
	cmd = trivialMark + cmd
	for i := 0; i < times; i++ {
		flow = append(flow, cmd)
	}
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
