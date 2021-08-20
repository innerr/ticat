package builtin

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
	meta "github.com/pingcap/ticat/pkg/proto/hub_meta"
	"github.com/pingcap/ticat/pkg/utils"
)

func LoadModsFromHub(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	metaExt := env.GetRaw("strs.meta-ext")
	flowExt := env.GetRaw("strs.flow-ext")
	helpExt := env.GetRaw("strs.help-ext")
	abbrsSep := env.GetRaw("strs.abbrs-sep")
	envPathSep := env.GetRaw("strs.env-path-sep")
	reposFileName := env.GetRaw("strs.repos-file-name")

	metaPath := getReposInfoPath(env, flow.Cmds[currCmdIdx])
	fieldSep := env.GetRaw("strs.proto-sep")

	panicRecover := env.GetBool("sys.panic.recover")

	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	for _, info := range infos {
		if info.OnOff != "on" {
			continue
		}
		source := info.Addr
		if len(source) == 0 {
			source = info.Path
		}
		loadLocalMods(cc, info.Path, reposFileName, metaExt, flowExt, helpExt,
			abbrsSep, envPathSep, source, panicRecover)
	}
	return currCmdIdx, true
}

func AddGitRepoToHub(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	addr := tailModeCallArg(flow, currCmdIdx, argv, "git-address")
	addRepoToHub(addr, argv, cc.Screen, env, flow.Cmds[currCmdIdx])
	showHubFindTip(cc.Screen, env)
	return currCmdIdx, true
}

func AddDefaultGitRepoToHub(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	addr := env.GetRaw("sys.hub.init-repo")
	if len(addr) == 0 {
		panic(core.NewCmdError(flow.Cmds[currCmdIdx],
			"cant't get init-repo address from env, 'sys.hub.init-repo' is empty"))
	}
	addRepoToHub(addr, argv, cc.Screen, env, flow.Cmds[currCmdIdx])
	showHubFindTip(cc.Screen, env)
	return currCmdIdx, true
}

func ListHub(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmd := flow.Cmds[currCmdIdx]
	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)

	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	screen := display.NewCacheScreen()

	listHub(screen, env, infos, findStrs...)
	if screen.OutputNum() <= 0 {
		if len(findStrs) == 0 {
			display.PrintTipTitle(cc.Screen, env,
				"'hub' manages all added git repos, now it's empty.",
				"",
				"add more git repos to get more avaialable commands:",
				"",
				display.SuggestHubAddShort(env))
		} else {
			var displayFindStrs []string
			for _, str := range findStrs {
				displayFindStrs = append(displayFindStrs, "'"+str+"'")
			}
			display.PrintTipTitle(cc.Screen, env,
				"can't find any repo/dir in hub by these keywords:",
				displayFindStrs)
		}
	} else {
		display.PrintTipTitle(cc.Screen, env, "repo list in hub:")
		screen.WriteTo(cc.Screen)
		cmdName := cmd.DisplayPath(cc.Cmds.Strs.PathSep, true)
		display.PrintTipTitle(cc.Screen, env,
			"command branch '"+cmdName+"' manages the repos in local disk.",
			"",
			"to see more usage:",
			"",
			display.SuggestHubBranch(env))
	}
	return currCmdIdx, true
}

func RemoveAllFromHub(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	cmd := flow.Cmds[currCmdIdx]

	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)

	for _, info := range infos {
		if !info.IsLocal() {
			osRemoveDir(info.Path, cmd)
		}
		cc.Screen.Print(fmt.Sprintf("%s%s\n", repoDisplayName(info, env), purgedStr(env, info.IsLocal())))
		printInfoProps(cc.Screen, env, info)
	}

	err := os.Remove(metaPath)
	if err != nil {
		if os.IsNotExist(err) && len(infos) == 0 {
			display.PrintTipTitle(cc.Screen, env,
				"hub is empty.",
				"",
				"add more git repos to get more avaialable commands:",
				"",
				display.SuggestHubAddShort(env))
			return currCmdIdx, true
		}
		panic(core.NewCmdError(cmd, fmt.Sprintf("remove '%s' failed: %v", metaPath, err)))
	}

	display.PrintTipTitle(cc.Screen, env,
		"hub now is empty.",
		"",
		"add more git repos to get more avaialable commands:",
		"",
		display.SuggestHubAddShort(env))

	return currCmdIdx, true
}

func PurgeAllInactiveReposFromHub(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	purgeInactiveRepoFromHub("", cc, env, flow.Cmds[currCmdIdx])
	return currCmdIdx, true
}

func PurgeInactiveRepoFromHub(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	findStr := tailModeCallArg(flow, currCmdIdx, argv, "find-str")
	purgeInactiveRepoFromHub(findStr, cc, env, flow.Cmds[currCmdIdx])
	return currCmdIdx, true
}

func CheckGitRepoStatus(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	panic("TODO")
}

func UpdateHub(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	cmd := flow.Cmds[currCmdIdx]
	metaPath := getReposInfoPath(env, cmd)
	listFileName := env.GetRaw("strs.repos-file-name")
	repoExt := env.GetRaw("strs.mods-repo-ext")

	path := getHubPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	oldInfos, oldList := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	finisheds := map[string]bool{}
	for _, info := range oldInfos {
		if info.OnOff != "on" {
			finisheds[info.Addr] = true
		}
	}

	selfName := env.GetRaw("strs.self-name")
	var infos []meta.RepoInfo

	for _, info := range oldInfos {
		if len(info.Addr) == 0 {
			continue
		}
		_, addrs, helpStrs := meta.UpdateRepoAndSubRepos(
			cc.Screen, env, finisheds, path, info.Addr, repoExt, listFileName, selfName, cmd)
		for i, addr := range addrs {
			if oldList[addr] {
				continue
			}
			repoPath := meta.GetRepoPath(path, addr)
			infos = append(infos, meta.RepoInfo{addr, info.Addr, repoPath, helpStrs[i], "on"})
		}
	}

	infos = append(oldInfos, infos...)
	if len(infos) != len(oldInfos) {
		meta.WriteReposInfoFile(metaPath, infos, fieldSep)
	}

	display.PrintTipTitle(cc.Screen, env, fmt.Sprintf(
		"local dir could also add to %s, use command 'h.add.local'",
		env.GetRaw("strs.self-name")))
	return currCmdIdx, true
}

func EnableRepoInHub(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmd := flow.Cmds[currCmdIdx]
	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	findStr := tailModeCallArg(flow, currCmdIdx, argv, "find-str")

	extracted, rest := meta.ExtractAddrFromList(infos, findStr)
	checkFoundRepos(env, cmd, extracted, findStr, true)

	var count int
	for i, info := range extracted {
		if info.OnOff == "on" {
			continue
		}
		count += 1
		cc.Screen.Print(fmt.Sprintf("%s%s\n", repoDisplayName(info, env), enabledStr(env, true)))
		printInfoProps(cc.Screen, env, info)
		info.OnOff = "on"
		extracted[i] = info
	}

	meta.WriteReposInfoFile(metaPath, append(rest, extracted...), fieldSep)

	if count > 0 {
		display.PrintTipTitle(cc.Screen, env,
			"add a disabled repo manually will enable it")
	} else {
		display.PrintTipTitle(cc.Screen, env,
			"no disabled repo matched find string '"+findStr+"'")
	}
	return currCmdIdx, true
}

func DisableRepoInHub(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmd := flow.Cmds[currCmdIdx]
	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	findStr := tailModeCallArg(flow, currCmdIdx, argv, "find-str")

	extracted, rest := meta.ExtractAddrFromList(infos, findStr)
	checkFoundRepos(env, cmd, extracted, findStr, false)

	var count int
	for i, info := range extracted {
		if info.OnOff == "on" {
			cc.Screen.Print(fmt.Sprintf("%s%s\n", repoDisplayName(info, env), disabledStr(env)))
			printInfoProps(cc.Screen, env, info)
			if info.OnOff != "disabled" {
				count += 1
			}
			info.OnOff = "disabled"
			extracted[i] = info
		}
	}

	meta.WriteReposInfoFile(metaPath, append(rest, extracted...), fieldSep)

	if count > 0 {
		display.PrintTipTitle(cc.Screen, env,
			"need two steps to remove a repo or unlink a dir: disable, purge")
	} else {
		display.PrintTipTitle(cc.Screen, env,
			"no enabled repo matched find string '"+findStr+"'")
	}
	return currCmdIdx, true
}

func AddLocalDirToHub(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmd := flow.Cmds[currCmdIdx]
	path := tailModeCallArg(flow, currCmdIdx, argv, "path")

	stat, err := os.Stat(path)
	if err != nil {
		panic(core.NewCmdError(cmd, fmt.Sprintf("access path '%v' failed: %v", path, err)))
	}
	if !stat.IsDir() {
		panic(core.NewCmdError(cmd, fmt.Sprintf("path '%v' is not dir", path)))
	}

	path, err = filepath.Abs(path)
	if err != nil {
		panic(core.NewCmdError(cmd, fmt.Sprintf("get abs path of '%v' failed: %v", path, err)))
	}

	screen := display.NewCacheScreen()

	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	found := false
	for i, info := range infos {
		if info.Path == path {
			if info.OnOff == "on" {
				screen.Print(fmt.Sprintf("%s (exists)\n", repoDisplayName(info, env)))
				printInfoProps(screen, env, info)
				display.PrintTipTitle(cc.Screen, env,
					"local dir already in hub, nothing to do")
				return currCmdIdx, true
			}
			info.OnOff = "on"
			infos[i] = info
			screen.Print(fmt.Sprintf("%s%s\n", repoDisplayName(info, env), enabledStr(env, true)))
			printInfoProps(screen, env, info)
			found = true
			break
		}
	}

	if !found {
		listFileName := env.GetRaw("strs.repos-file-name")
		listFilePath := filepath.Join(path, listFileName)
		helpStr, _, _ := meta.ReadRepoListFromFile(env.GetRaw("strs.self-name"), listFilePath)
		info := meta.RepoInfo{"", "<local>", path, helpStr, "on"}
		infos = append(infos, info)
		screen.Print(fmt.Sprintf("%s\n", repoDisplayName(info, env)))
		printInfoProps(screen, env, info)
	}
	meta.WriteReposInfoFile(metaPath, infos, fieldSep)

	if !found {
		display.PrintTipTitle(cc.Screen, env,
			"local dir added to hub")
	} else {
		display.PrintTipTitle(cc.Screen, env,
			"local dir re-enabled")
	}
	screen.WriteTo(cc.Screen)

	// TODO: load mods now?
	return currCmdIdx, true
}

func MoveSavedFlowsToLocalDir(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	// TODO: do accurate matching before search
	cmd := flow.Cmds[currCmdIdx]
	path := argv.GetRaw("path")
	if len(path) == 0 {
		args := tailModeGetInput(flow, currCmdIdx, false)
		if len(args) > 1 {
			panic(core.NewCmdError(cmd, "too many input of arg 'path' in tail-mode"))
		} else if len(args) == 1 {
			path = args[0]
		}
	}

	if len(path) != 0 {
		stat, err := os.Stat(path)
		if err != nil && !os.IsNotExist(err) {
			panic(core.NewCmdError(cmd, fmt.Sprintf("access path '%v' failed: %v", path, err)))
		}

		if !os.IsNotExist(err) {
			if !stat.IsDir() {
				panic(core.NewCmdError(cmd, fmt.Sprintf("path '%v' exists but is not a dir", path)))
			}
			moveSavedFlowsToLocalDir(path, cc, env, cmd)
			display.PrintTipTitle(cc.Screen, env,
				"all saved flow moved to '"+path+"'.")
			return currCmdIdx, true
		}
	}

	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)

	var locals []meta.RepoInfo
	for _, info := range infos {
		if len(info.Addr) != 0 {
			continue
		}
		if strings.Index(info.Path, path) >= 0 {
			locals = append(locals, info)
		}
	}

	if len(locals) > 1 {
		var actives []meta.RepoInfo
		for _, info := range locals {
			if info.OnOff == "on" {
				actives = append(actives, info)
			}
		}
		locals = actives
	}

	if len(locals) == 0 {
		if len(path) != 0 {
			display.PrintErrTitle(cc.Screen, env,
				fmt.Sprintf("can't find matched dir by string '%s'.", path),
				"notice this command only search added local dirs, not repos")
		} else {
			display.PrintErrTitle(cc.Screen, env,
				"no local dirs to move flows into.",
				"notice this command only search added local dirs, not repos")
		}
		return currCmdIdx, false
	}
	if len(locals) > 1 {
		if len(path) != 0 {
			display.PrintErrTitle(cc.Screen, env,
				"can't determine which dir by string '"+path+"'.",
				"only could move to the one and only matched dir.",
				"", "current matcheds:")
		} else {
			display.PrintErrTitle(cc.Screen, env,
				"can't determine which dir, only could move to the one and only matched dir.",
				"",
				"add more find-str to filter it:",
				"",
				display.SuggestFilterRepoInMove(env),
				"", "current matcheds:")
		}
		listHub(cc.Screen, env, locals)
		return currCmdIdx, false
	}

	cnt := moveSavedFlowsToLocalDir(locals[0].Path, cc, env, cmd)

	if cnt > 0 {
		display.PrintTipTitle(cc.Screen, env,
			"all saved flow are moved to '"+locals[0].Path+"', it's the only matched local dir in hub")
	} else {
		display.PrintTipTitle(cc.Screen, env,
			"not saved flow to move")
	}
	return currCmdIdx, true
}

func listHub(screen core.Screen, env *core.Env, infos []meta.RepoInfo, filterStrs ...string) {
	for _, info := range infos {
		if len(filterStrs) != 0 {
			filtered := false
			for _, filterStr := range filterStrs {
				if !matchFindRepoInfo(info, filterStr) {
					filtered = true
					break
				}
			}
			if filtered {
				continue
			}
		}
		name := repoDisplayName(info, env)
		screen.Print(name)
		if info.OnOff != "on" {
			screen.Print(disabledStr(env))
		} else {
			screen.Print(enabledStr(env, false))
		}
		screen.Print("\n")
		if len(info.HelpStr) > 0 {
			screen.Print(fmt.Sprintf(display.ColorHelp("     '%s'\n", env), info.HelpStr))
		}
		if len(info.Addr) != 0 && name != info.Addr {
			screen.Print(fmt.Sprintf(display.ColorProp("    - addr: ", env)+"%s\n", info.Addr))
		}
		screen.Print(fmt.Sprintf(display.ColorProp("    - from: ", env)+"%s\n", getDisplayReason(info)))
		screen.Print(fmt.Sprintf(display.ColorProp("    - path: ", env)+"%s\n", info.Path))
	}
}

func purgeInactiveRepoFromHub(findStr string, cc *core.Cli, env *core.Env, cmd core.ParsedCmd) {
	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)

	var extracted []meta.RepoInfo
	var rest []meta.RepoInfo
	for _, info := range infos {
		if info.OnOff != "on" && (len(findStr) == 0 || matchFindRepoInfo(info, findStr)) {
			extracted = append(extracted, info)
		} else {
			rest = append(rest, info)
		}
	}
	checkFoundRepos(env, cmd, extracted, findStr, false)

	var unlinkeds int
	var removeds int

	for _, info := range extracted {
		if !info.IsLocal() {
			osRemoveDir(info.Path, cmd)
			removeds += 1
		} else {
			unlinkeds += 1
		}
		cc.Screen.Print(fmt.Sprintf("%s%s\n", repoDisplayName(info, env), purgedStr(env, info.IsLocal())))
		printInfoProps(cc.Screen, env, info)
	}

	if len(extracted) <= 0 {
		return
	}

	meta.WriteReposInfoFile(metaPath, rest, fieldSep)

	var helpStr []string
	if removeds > 0 {
		helpStr = append(helpStr, fmt.Sprintf("%v repos removed from local disk.", removeds))
	}
	if unlinkeds > 0 {
		helpStr = append(helpStr, fmt.Sprintf("%v local dir unlinked to %s, files are untouched.",
			unlinkeds, env.GetRaw("strs.self-name")))
	}
	display.PrintTipTitle(cc.Screen, env, helpStr)
}

func moveSavedFlowsToLocalDir(toDir string, cc *core.Cli, env *core.Env, cmd core.ParsedCmd) int {
	flowExt := env.GetRaw("strs.flow-ext")
	root := getFlowRoot(env, cmd)

	var count int
	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, flowExt) {
			return nil
		}

		// This dir is managed, so will be no sub-dir
		destPath := filepath.Join(toDir, filepath.Base(path))

		err = utils.MoveFile(path, destPath)
		if err != nil {
			panic(core.NewCmdError(cmd, fmt.Sprintf("rename file '%s' to '%s' failed: %v",
				path, destPath, err)))
		}
		cmdPath := getCmdPath(path, flowExt, cmd)
		cc.Screen.Print(fmt.Sprintf(display.ColorHub("[%s]\n", env), cmdPath))
		cc.Screen.Print(fmt.Sprintf(display.ColorProp("    - from:", env)+" %s\n", path))
		cc.Screen.Print(fmt.Sprintf(display.ColorProp("    - to: ", env)+"%s\n", destPath))
		count += 1
		return nil
	})
	return count
}

func addRepoToHub(
	gitAddr string,
	argv core.ArgVals,
	screen core.Screen,
	env *core.Env,
	cmd core.ParsedCmd) (addrs []string, helpStrs []string) {

	// A repo with this suffix should be a well controlled one, that we could assume some things
	repoExt := env.GetRaw("strs.mods-repo-ext")

	gitAddr = meta.NormalizeGitAddr(gitAddr)

	if !isOsCmdExists("git") {
		panic(core.NewCmdError(cmd, "cant't find 'git'"))
	}

	path := getHubPath(env, cmd)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		panic(core.NewCmdError(cmd, fmt.Sprintf("create hub path '%s' failed: %v", path, err)))
	}

	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	oldInfos, oldList := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	finisheds := map[string]bool{}
	for i, info := range oldInfos {
		if info.Addr == gitAddr {
			info.OnOff = "on"
			oldInfos[i] = info
		}
		if info.OnOff != "on" {
			finisheds[info.Addr] = true
		}
	}

	selfName := env.GetRaw("strs.self-name")
	listFileName := env.GetRaw("strs.repos-file-name")
	var topRepoHelpStr string
	topRepoHelpStr, addrs, helpStrs = meta.UpdateRepoAndSubRepos(
		screen, env, finisheds, path, gitAddr, repoExt, listFileName, selfName, cmd)

	addrs = append([]string{gitAddr}, addrs...)
	helpStrs = append([]string{topRepoHelpStr}, helpStrs...)

	var infos []meta.RepoInfo
	for i, addr := range addrs {
		if oldList[addr] {
			continue
		}
		repoPath := meta.GetRepoPath(path, addr)
		infos = append(infos, meta.RepoInfo{addr, gitAddr, repoPath, helpStrs[i], "on"})
	}

	infos = append(oldInfos, infos...)
	meta.WriteReposInfoFile(metaPath, infos, fieldSep)
	return
}

func printInfoProps(screen core.Screen, env *core.Env, info meta.RepoInfo) {
	if len(info.HelpStr) > 0 {
		screen.Print(fmt.Sprintf(display.ColorHelp("     '%s'\n", env), info.HelpStr))
	}
	screen.Print(fmt.Sprintf(display.ColorProp("    - from: ", env)+"%s\n", getDisplayReason(info)))
	screen.Print(fmt.Sprintf(display.ColorProp("    - path: ", env)+"%s\n", info.Path))
}

func getDisplayReason(info meta.RepoInfo) string {
	if info.AddReason == info.Addr {
		return "<manually-added>"
	}
	return info.AddReason
}

func checkFoundRepos(
	env *core.Env,
	cmd core.ParsedCmd,
	infos []meta.RepoInfo,
	findStr string,
	expectActive bool) {

	var status string
	if expectActive {
		status = "enabled"
	} else {
		status = "disabled"
	}
	if len(infos) == 0 {
		panic(core.NewCmdError(cmd, fmt.Sprintf("no %s repo matched find string '%s'", status, findStr)))
	}
}

func getHubPath(env *core.Env, cmd core.ParsedCmd) string {
	path := env.GetRaw("sys.paths.hub")
	if len(path) == 0 {
		panic(core.NewCmdError(cmd, "cant't get hub path from env, 'sys.paths.hub' is empty"))
	}
	return path
}

func matchFindRepoInfo(info meta.RepoInfo, findStr string) bool {
	if len(findStr) == 0 {
		return true
	}
	if strings.Index(info.Addr, findStr) >= 0 {
		return true
	}
	if strings.Index(info.Path, findStr) >= 0 {
		return true
	}
	if strings.Index(getDisplayReason(info), findStr) >= 0 {
		return true
	}
	if strings.Index(info.HelpStr, findStr) >= 0 {
		return true
	}
	if strings.Index(info.OnOff, findStr) >= 0 {
		return true
	}

	// TODO: better place for string "local"
	if len(info.Addr) == 0 && strings.Index("local", findStr) >= 0 {
		return true
	}
	return false
}

func showHubFindTip(screen core.Screen, env *core.Env) {
	display.PrintTipTitle(screen, env,
		"try to search/show commands in this repo:",
		"",
		display.SuggestFindCmdsInRepo(env))
}

func getReposInfoPath(env *core.Env, cmd core.ParsedCmd) string {
	path := getHubPath(env, cmd)
	reposInfoFileName := env.GetRaw("strs.hub-file-name")
	if len(reposInfoFileName) == 0 {
		panic(core.NewCmdError(cmd, "cant't hub meta file name"))
	}
	return filepath.Join(path, reposInfoFileName)
}

func repoDisplayName(info meta.RepoInfo, env *core.Env) string {
	var name string
	if len(info.Addr) == 0 {
		name = filepath.Base(info.Path)
	} else {
		name = meta.AddrDisplayName(info.Addr)
	}
	return display.ColorHub("["+name+"]", env)
}

func disabledStr(env *core.Env) string {
	return display.ColorDisabled(" (disabled)", env)
}

func enabledStr(env *core.Env, str bool) string {
	if str {
		return display.ColorEnabled(" (enabled)", env)
	} else {
		return ""
	}
}

func purgedStr(env *core.Env, isLocal bool) string {
	if isLocal {
		return display.ColorDisabled(" (unlinked)", env)
	} else {
		return display.ColorDisabled(" (purged)", env)
	}
}
