package builtin

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func EnvSaveToSnapshot(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	kvSep := env.GetRaw("strs.env-kv-sep")

	cmd := flow.Cmds[currCmdIdx]
	name := getAndCheckArg(argv, cmd, "snapshot-name")
	path := getEnvSnapshotPath(env, cmd, name)

	overwrite := argv.GetBool("overwrite")
	if !overwrite && fileExists(path) {
		panic(core.NewCmdError(cmd, "env snapshot '"+name+"' already exists"))
	}

	core.SaveEnvToFile(env, path, kvSep, true)
	display.PrintTipTitle(cc.Screen, env,
		"session env are saved to snapshot '"+name+"', could be use by:",
		"",
		display.SuggestLoadEnvSnapshot(env))

	return currCmdIdx, true
}

func EnvRemoveSnapshot(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	cmd := flow.Cmds[currCmdIdx]
	name := getAndCheckArg(argv, cmd, "snapshot-name")
	path := getEnvSnapshotPath(env, cmd, name)

	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			display.PrintTipTitle(cc.Screen, env,
				fmt.Sprintf("env snapshot '%s' not exists\n", name))
		} else {
			panic(core.NewCmdError(cmd,
				fmt.Sprintf("remove env snapshot file '%s' failed: %v", path, err)))
		}
	} else {
		display.PrintTipTitle(cc.Screen, env,
			"env snapshot '"+name+"' is removed")
	}

	return currCmdIdx, true
}

func EnvListSnapshots(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	cmd := flow.Cmds[currCmdIdx]
	ext := env.GetRaw("strs.env-snapshot-ext")
	if len(ext) == 0 {
		panic(core.NewCmdError(cmd, "env value 'strs.env-snapshot-ext' is empty"))
	}

	root := getEnvSnapshotDir(env, cmd)
	var names []string

	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if path == root {
			return nil
		}
		if !strings.HasSuffix(path, ext) {
			return nil
		}
		name := path[len(root)+1:]
		name = name[0 : len(name)-len(ext)]
		names = append(names, name)
		return nil
	})

	if len(names) > 0 {
		var title string
		if len(names) == 1 {
			title = "has 1 saved env snapshot."
		} else {
			title = fmt.Sprintf("have %v saved env snapshots.", len(names))
		}
		display.PrintTipTitle(cc.Screen, env,
			title,
			"",
			"could be use by:",
			"",
			display.SuggestLoadEnvSnapshot(env))
		for _, name := range names {
			fmt.Println(name)
		}
	}

	return currCmdIdx, true
}

func EnvLoadFromSnapshot(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	cmd := flow.Cmds[currCmdIdx]
	name := getAndCheckArg(argv, cmd, "snapshot-name")
	path := getEnvSnapshotPath(env, cmd, name)

	sep := cc.Cmds.Strs.EnvKeyValSep
	delMark := cc.Cmds.Strs.EnvValDelAllMark

	loaded := core.NewEnv()
	core.LoadEnvFromFile(loaded, path, sep, delMark)
	loaded.WriteCurrLayerTo(env.GetLayer(core.EnvLayerSession))

	return currCmdIdx, true
}

func EnvLoadNonExistFromSnapshot(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	cmd := flow.Cmds[currCmdIdx]
	name := getAndCheckArg(argv, cmd, "snapshot-name")
	path := getEnvSnapshotPath(env, cmd, name)

	sep := cc.Cmds.Strs.EnvKeyValSep
	delMark := cc.Cmds.Strs.EnvValDelAllMark

	loaded := core.NewEnv()
	core.LoadEnvFromFile(loaded, path, sep, delMark)

	envSession := env.GetLayer(core.EnvLayerSession)
	for k, v := range loaded.FlattenAll() {
		if !env.Has(k) {
			envSession.Set(k, v)
		}
	}

	return currCmdIdx, true
}

func getEnvSnapshotDir(env *core.Env, cmd core.ParsedCmd) string {
	dir := env.GetRaw("sys.paths.env.snapshot")
	if len(dir) == 0 {
		panic(core.NewCmdError(cmd, "env value 'sys.paths.env.snapshot' is empty"))
	}
	os.MkdirAll(dir, os.ModePerm)
	return dir
}

func getEnvSnapshotPath(env *core.Env, cmd core.ParsedCmd, name string) string {
	ext := env.GetRaw("strs.env-snapshot-ext")
	if len(ext) == 0 {
		panic(core.NewCmdError(cmd, "env value 'strs.env-snapshot-ext' is empty"))
	}
	dir := getEnvSnapshotDir(env, cmd)
	return filepath.Join(dir, name) + ext
}
