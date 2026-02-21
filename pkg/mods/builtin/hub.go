package builtin

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/innerr/ticat/pkg/cli/display"
	"github.com/innerr/ticat/pkg/core/model"
	meta "github.com/innerr/ticat/pkg/mods/persist/hub_meta"
	"github.com/innerr/ticat/pkg/utils"
)

func LoadModsFromHub(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	metaPath := getReposInfoPath(env, flow.Cmds[currCmdIdx])
	fieldSep := env.GetRaw("strs.proto-sep")

	hubDir := env.GetRaw("sys.paths.hub")
	infos, _, err := meta.ReadReposInfoFile(hubDir, metaPath, true, fieldSep)
	if err != nil {
		return currCmdIdx, err
	}
	for _, info := range infos {
		if info.OnOff != "on" {
			continue
		}
		loadRepoMods(cc, env, info)
	}
	return currCmdIdx, nil
}

func AddGitRepoToHub(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	addr, err := tailModeCallArg(flow, currCmdIdx, argv, "git-address")
	if err != nil {
		return currCmdIdx, err
	}
	branch := argv.GetRaw("git-branch")
	if _, _, err := addRepoToHubAndLoadMods(cc, meta.RepoAddr{Addr: addr, Branch: branch},
		argv, cc.Screen, env, flow.Cmds[currCmdIdx]); err != nil {
		return currCmdIdx, err
	}

	initAddr := env.GetRaw("sys.hub.init-repo")
	if len(initAddr) != 0 {
		sep := env.GetRaw("strs.list-sep")
		addrs := strings.Split(initAddr, sep)
		for _, addr := range addrs {
			if _, _, err := addRepoToHubAndLoadMods(cc, meta.RepoAddr{Addr: addr, Branch: ""},
				argv, cc.Screen, env, flow.Cmds[currCmdIdx]); err != nil {
				return currCmdIdx, err
			}
		}
	}

	showHubFindTip(cc.Screen, env)
	return currCmdIdx, nil
}

func EnsureDefaultGitRepoInHub(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	hubDir := env.GetRaw("sys.paths.hub")
	if dirExists(hubDir) {
		return currCmdIdx, nil
	}
	display.PrintTipTitle(cc.Screen, env,
		"do 'hub.init' for the first time running")
	argv["show-tip"] = model.ArgVal{Raw: "false", Provided: true, Index: 0}
	newCurrCmdIdx, err := AddDefaultGitRepoToHub(argv, cc, env, flow, currCmdIdx)
	if err == nil {
		display.PrintTipTitle(cc.Screen, env,
			"init for the first time running: done")
	}
	return newCurrCmdIdx, err
}

func AddDefaultGitRepoToHub(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	addr := env.GetRaw("sys.hub.init-repo")
	if len(addr) == 0 {
		return currCmdIdx, nil
	}
	sep := env.GetRaw("strs.list-sep")
	addrs := strings.Split(addr, sep)
	for _, addr := range addrs {
		if _, _, err := addRepoToHubAndLoadMods(cc, meta.RepoAddr{Addr: addr, Branch: ""},
			argv, cc.Screen, env, flow.Cmds[currCmdIdx]); err != nil {
			return currCmdIdx, err
		}
	}
	if showTip := argv.GetBool("show-tip"); showTip {
		showHubFindTip(cc.Screen, env)
	}
	return currCmdIdx, nil
}

func ListHub(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	cmd := flow.Cmds[currCmdIdx]
	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)

	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	hubDir := env.GetRaw("sys.paths.hub")
	infos, _, err := meta.ReadReposInfoFile(hubDir, metaPath, true, fieldSep)
	if err != nil {
		return currCmdIdx, err
	}
	screen := display.NewCacheScreen()

	listHub(screen, env, infos, findStrs...)
	if screen.OutputtedLines() <= 0 {
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
	return currCmdIdx, nil
}

func RemoveAllFromHub(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	cmd := flow.Cmds[currCmdIdx]

	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	hubDir := env.GetRaw("sys.paths.hub")
	infos, _, err := meta.ReadReposInfoFile(hubDir, metaPath, true, fieldSep)
	if err != nil {
		return currCmdIdx, err
	}

	for _, info := range infos {
		if !info.IsLocal() {
			_ = osRemoveDir(info.Path, cmd)
		}
		_ = cc.Screen.Print(fmt.Sprintf("%s%s\n", repoDisplayName(info, env), purgedStr(env, info.IsLocal())))
		printInfoProps(cc.Screen, env, info)
	}

	err = os.Remove(metaPath)
	if err != nil {
		if os.IsNotExist(err) && len(infos) == 0 {
			display.PrintTipTitle(cc.Screen, env,
				"hub is empty.",
				"",
				"add more git repos to get more avaialable commands:",
				"",
				display.SuggestHubAddShort(env))
			return currCmdIdx, nil
		}
		return currCmdIdx, model.NewCmdError(cmd, fmt.Sprintf("remove '%s' failed: %v", metaPath, err))
	}

	display.PrintTipTitle(cc.Screen, env,
		"hub now is empty.",
		"",
		"add more git repos to get more avaialable commands:",
		"",
		display.SuggestHubAddShort(env))

	return currCmdIdx, nil
}

func PurgeAllInactiveReposFromHub(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	if err := purgeInactiveRepoFromHub("", cc, env, flow.Cmds[currCmdIdx]); err != nil {
		return currCmdIdx, err
	}
	return currCmdIdx, nil
}

func PurgeInactiveRepoFromHub(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	findStr, err := tailModeCallArg(flow, currCmdIdx, argv, "find-str")
	if err != nil {
		return currCmdIdx, err
	}
	if err := purgeInactiveRepoFromHub(findStr, cc, env, flow.Cmds[currCmdIdx]); err != nil {
		return currCmdIdx, err
	}
	return currCmdIdx, nil
}

func CheckGitRepoStatus(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	cmd := flow.Cmds[currCmdIdx]
	if !isOsCmdExists("git") {
		return currCmdIdx, model.NewCmdError(cmd, "cant't find 'git'")
	}

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)

	path := getHubPath(env, cmd)

	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	hubDir := env.GetRaw("sys.paths.hub")
	infos, _, err := meta.ReadReposInfoFile(hubDir, metaPath, true, fieldSep)
	if err != nil {
		return currCmdIdx, err
	}

	// Do not do exactl-match for readonly commands

	for _, info := range infos {
		if len(info.Addr.Str()) == 0 {
			continue
		}
		if len(findStrs) != 0 {
			filtered := false
			for _, filterStr := range findStrs {
				if !matchFindRepoInfo(info, filterStr) {
					filtered = true
					break
				}
			}
			if filtered {
				continue
			}
		}
		if err := meta.CheckRepoGitStatus(cc.Screen, env, path, info.Addr); err != nil {
			return currCmdIdx, err
		}
	}

	return currCmdIdx, nil
}

func UpdateHub(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	cmd := flow.Cmds[currCmdIdx]
	metaPath := getReposInfoPath(env, cmd)
	listFileName := env.GetRaw("strs.repos-file-name")
	repoExt := env.GetRaw("strs.mods-repo-ext")

	path := getHubPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	hubDir := env.GetRaw("sys.paths.hub")
	oldInfos, oldList, err := meta.ReadReposInfoFile(hubDir, metaPath, true, fieldSep)
	if err != nil {
		return currCmdIdx, err
	}
	finisheds := map[string]bool{}
	for _, info := range oldInfos {
		if info.OnOff != "on" {
			finisheds[info.Addr.Str()] = true
		}
	}

	selfName := env.GetRaw("strs.self-name")
	var infos []meta.RepoInfo

	for _, info := range oldInfos {
		if len(info.Addr.Str()) == 0 {
			continue
		}
		_, addrs, helpStrs, err := meta.UpdateRepoAndSubRepos(
			cc.Screen, env, finisheds, path, info.Addr, repoExt, listFileName, selfName, cmd)
		if err != nil {
			return currCmdIdx, err
		}
		for i, addr := range addrs {
			if oldList[addr.Str()] {
				continue
			}
			repoPath, err := meta.GetRepoPath(path, addr)
			if err != nil {
				return currCmdIdx, err
			}
			infos = append(infos, meta.RepoInfo{Addr: addr, AddReason: info.Addr.Str(), Path: repoPath, HelpStr: helpStrs[i], OnOff: "on"})
		}
	}

	infos = append(oldInfos, infos...)
	if len(infos) != len(oldInfos) {
		if err := meta.WriteReposInfoFile(hubDir, metaPath, infos, fieldSep); err != nil {
			return currCmdIdx, err
		}
	}

	if showTip := argv.GetBool("show-tip"); showTip {
		display.PrintTipTitle(cc.Screen, env, fmt.Sprintf(
			"local dir could also add to %s, use command 'hub.add.local'",
			env.GetRaw("strs.self-name")))
	}
	return currCmdIdx, nil
}

func EnableRepoInHub(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	cmd := flow.Cmds[currCmdIdx]
	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	hubDir := env.GetRaw("sys.paths.hub")
	infos, _, err := meta.ReadReposInfoFile(hubDir, metaPath, true, fieldSep)
	if err != nil {
		return currCmdIdx, err
	}
	findStr, err := tailModeCallArg(flow, currCmdIdx, argv, "find-str")
	if err != nil {
		return currCmdIdx, err
	}

	extracted, rest := meta.ExtractAddrFromList(infos, findStr)
	checkFoundRepos(env, cmd, extracted, findStr, true)

	var count int
	for i, info := range extracted {
		if info.OnOff == "on" {
			continue
		}
		count += 1
		_ = cc.Screen.Print(fmt.Sprintf("%s%s\n", repoDisplayName(info, env), enabledStr(env, true)))
		printInfoProps(cc.Screen, env, info)
		info.OnOff = "on"
		extracted[i] = info
	}

	if err := meta.WriteReposInfoFile(hubDir, metaPath, append(rest, extracted...), fieldSep); err != nil {
		return currCmdIdx, err
	}

	if count > 0 {
		display.PrintTipTitle(cc.Screen, env,
			"add a disabled repo manually will enable it")
	} else {
		display.PrintTipTitle(cc.Screen, env,
			"no disabled repo matched find string '"+findStr+"'")
	}
	return currCmdIdx, nil
}

func DisableRepoInHub(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	cmd := flow.Cmds[currCmdIdx]
	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	hubDir := env.GetRaw("sys.paths.hub")
	infos, _, err := meta.ReadReposInfoFile(hubDir, metaPath, true, fieldSep)
	if err != nil {
		return currCmdIdx, err
	}
	findStr, err := tailModeCallArg(flow, currCmdIdx, argv, "find-str")
	if err != nil {
		return currCmdIdx, err
	}

	extracted, rest := meta.ExtractAddrFromList(infos, findStr)
	checkFoundRepos(env, cmd, extracted, findStr, false)

	var count int
	for i, info := range extracted {
		if info.OnOff == "on" {
			_ = cc.Screen.Print(fmt.Sprintf("%s%s\n", repoDisplayName(info, env), disabledStr(env)))
			printInfoProps(cc.Screen, env, info)
			if info.OnOff != "disabled" {
				count += 1
			}
			info.OnOff = "disabled"
			extracted[i] = info
		}
	}

	_ = meta.WriteReposInfoFile(hubDir, metaPath, append(rest, extracted...), fieldSep)

	if count > 0 {
		display.PrintTipTitle(cc.Screen, env,
			"need two steps to remove a repo or unlink a dir: disable, purge")
	} else {
		display.PrintTipTitle(cc.Screen, env,
			"no enabled repo matched find string '"+findStr+"'")
	}
	return currCmdIdx, nil
}

func AddLocalDirToHub(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	path, err := tailModeCallArg(flow, currCmdIdx, argv, "path")
	if err != nil {
		return currCmdIdx, err
	}
	return addLocalDirToHub(argv, cc, env, flow, currCmdIdx, path)
}

func AddPwdToHub(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	return addLocalDirToHub(argv, cc, env, flow, currCmdIdx, ".")
}

func addLocalDirToHub(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int,
	path string) (int, error) {

	cmd := flow.Cmds[currCmdIdx]

	stat, err := os.Stat(path)
	if err != nil {
		return currCmdIdx, model.NewCmdError(cmd, fmt.Sprintf("access path '%v' failed: %v", path, err))
	}
	if !stat.IsDir() {
		return currCmdIdx, model.NewCmdError(cmd, fmt.Sprintf("path '%v' is not dir", path))
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return currCmdIdx, model.NewCmdError(cmd, fmt.Sprintf("get abs path of '%v' failed: %v", path, err))
	}

	screen := display.NewCacheScreen()

	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	hubDir := env.GetRaw("sys.paths.hub")
	infos, _, err := meta.ReadReposInfoFile(hubDir, metaPath, true, fieldSep)
	if err != nil {
		return currCmdIdx, err
	}
	found := false
	for i, info := range infos {
		if info.Path == path {
			if info.OnOff == "on" {
				_ = screen.Print(fmt.Sprintf("%s (exists)\n", repoDisplayName(info, env)))
				printInfoProps(screen, env, info)
				display.PrintTipTitle(cc.Screen, env,
					"local dir already in hub, nothing to do")
				return currCmdIdx, nil
			}
			info.OnOff = "on"
			infos[i] = info
			_ = screen.Print(fmt.Sprintf("%s%s\n", repoDisplayName(info, env), enabledStr(env, true)))
			printInfoProps(screen, env, info)
			found = true
			break
		}
	}

	if !found {
		listFileName := env.GetRaw("strs.repos-file-name")
		listFilePath := filepath.Join(path, listFileName)
		helpStr, _, _, _ := meta.ReadRepoListFromFile(env.GetRaw("strs.self-name"), listFilePath)
		info := meta.RepoInfo{Addr: meta.RepoAddr{Addr: "", Branch: ""}, AddReason: "<local>", Path: path, HelpStr: helpStr, OnOff: "on"}
		infos = append(infos, info)
		_ = screen.Print(fmt.Sprintf("%s\n", repoDisplayName(info, env)))
		printInfoProps(screen, env, info)
	}
	if err := meta.WriteReposInfoFile(hubDir, metaPath, infos, fieldSep); err != nil {
		return currCmdIdx, err
	}

	if !found {
		display.PrintTipTitle(cc.Screen, env,
			"local dir added to hub")
	} else {
		display.PrintTipTitle(cc.Screen, env,
			"local dir re-enabled")
	}
	screen.WriteTo(cc.Screen)

	// TODO: load mods now?
	return currCmdIdx, nil
}

func MoveSavedFlowsToPwd(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	return cmdMoveSavedFlowsToLocalDir(argv, cc, env, flow, currCmdIdx, ".")
}

func MoveSavedFlowsToLocalDir(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	return cmdMoveSavedFlowsToLocalDir(argv, cc, env, flow, currCmdIdx, "")
}

func cmdMoveSavedFlowsToLocalDir(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int,
	defaultDir string) (int, error) {

	// TODO: do accurate matching before search
	cmd := flow.Cmds[currCmdIdx]
	path := argv.GetRawEx("path", defaultDir)
	if len(path) == 0 {
		args := tailModeGetInput(flow, currCmdIdx, false)
		if len(args) > 1 {
			return currCmdIdx, model.NewCmdError(cmd, "too many input of arg 'path' in tail-mode")
		} else if len(args) == 1 {
			path = args[0]
		}
	}

	if len(path) != 0 {
		stat, err := os.Stat(path)
		if err != nil && !os.IsNotExist(err) {
			return currCmdIdx, model.NewCmdError(cmd, fmt.Sprintf("access path '%v' failed: %v", path, err))
		}

		if !os.IsNotExist(err) {
			if !stat.IsDir() {
				return currCmdIdx, model.NewCmdError(cmd, fmt.Sprintf("path '%v' exists but is not a dir", path))
			}
			moveSavedFlowsToLocalDir(path, cc, env, cmd)
			display.PrintTipTitle(cc.Screen, env,
				"all saved flow moved to '"+path+"'")
			return currCmdIdx, nil
		}
	}

	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	hubDir := env.GetRaw("sys.paths.hub")
	infos, _, err := meta.ReadReposInfoFile(hubDir, metaPath, true, fieldSep)
	if err != nil {
		return currCmdIdx, err
	}

	var locals []meta.RepoInfo
	for _, info := range infos {
		if len(info.Addr.Str()) != 0 {
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
		return currCmdIdx, fmt.Errorf("no local dirs found")
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
		return currCmdIdx, fmt.Errorf("multiple local dirs matched")
	}

	cnt := moveSavedFlowsToLocalDir(locals[0].Path, cc, env, cmd)

	if cnt > 0 {
		display.PrintTipTitle(cc.Screen, env,
			"all saved flow are moved to '"+locals[0].Path+"', it's the only matched local dir in hub")
	} else {
		display.PrintTipTitle(cc.Screen, env,
			"not saved flow to move")
	}
	return currCmdIdx, nil
}

func exactMatchHubRepo(infos []meta.RepoInfo, findStrs ...string) (info meta.RepoInfo,
	rest []meta.RepoInfo, matched bool) {

	if len(findStrs) != 1 {
		return
	}
	findStr := findStrs[0]
	if len(findStr) == 0 {
		return
	}
	for i, info := range infos {
		if info.Addr.Str() == findStr {
			return info, append(infos[:i], infos[i+1:]...), true
		}
	}
	return
}

func listHub(screen model.Screen, env *model.Env, infos []meta.RepoInfo, findStrs ...string) {
	// Do not do exactl-match for readonly commands
	for _, info := range infos {
		if len(findStrs) != 0 {
			filtered := false
			for _, filterStr := range findStrs {
				if !matchFindRepoInfo(info, filterStr) {
					filtered = true
					break
				}
			}
			if filtered {
				continue
			}
		}
		displayHubRepo(screen, env, info)
	}
}

func displayHubRepo(screen model.Screen, env *model.Env, info meta.RepoInfo) {
	name := repoDisplayName(info, env)
	_ = screen.Print(name)
	if info.OnOff != "on" {
		_ = screen.Print(disabledStr(env))
	} else {
		_ = screen.Print(enabledStr(env, false))
	}
	_ = screen.Print("\n")
	if len(info.HelpStr) > 0 {
		_ = screen.Print(fmt.Sprintf(display.ColorHelp("     '%s'\n", env), info.HelpStr))
	}
	if len(info.Addr.Addr) != 0 {
		_ = screen.Print(fmt.Sprintf(display.ColorProp("    - address: ", env)+"%s\n", info.Addr.Addr))
		if len(info.Addr.Branch) != 0 {
			_ = screen.Print(fmt.Sprintf(display.ColorProp("    - branch:  ", env)+"%s\n", info.Addr.Branch))
		}
	}
	_ = screen.Print(fmt.Sprintf(display.ColorProp("    - from:    ", env)+"%s\n", getDisplayReason(info)))
	_ = screen.Print(fmt.Sprintf(display.ColorProp("    - path:    ", env)+"%s\n", info.Path))
}

func purgeInactiveRepoFromHub(findStr string, cc *model.Cli, env *model.Env, cmd model.ParsedCmd) error {
	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	hubDir := env.GetRaw("sys.paths.hub")
	infos, _, err := meta.ReadReposInfoFile(hubDir, metaPath, true, fieldSep)
	if err != nil {
		return err
	}

	var extracted []meta.RepoInfo
	var rest []meta.RepoInfo

	if matchedInfo, matchedRest, matched := exactMatchHubRepo(infos, findStr); matched && matchedInfo.OnOff != "on" {
		extracted = append(extracted, matchedInfo)
		rest = matchedRest
	} else {
		for _, info := range infos {
			if info.OnOff != "on" && (len(findStr) == 0 || matchFindRepoInfo(info, findStr)) {
				extracted = append(extracted, info)
			} else {
				rest = append(rest, info)
			}
		}
		checkFoundRepos(env, cmd, extracted, findStr, false)
	}

	var unlinkeds int
	var removeds int

	for _, info := range extracted {
		if !info.IsLocal() {
			_ = osRemoveDir(info.Path, cmd)
			removeds += 1
		} else {
			unlinkeds += 1
		}
		_ = cc.Screen.Print(fmt.Sprintf("%s%s\n", repoDisplayName(info, env), purgedStr(env, info.IsLocal())))
		printInfoProps(cc.Screen, env, info)
	}

	if len(extracted) <= 0 {
		return nil
	}

	if err := meta.WriteReposInfoFile(hubDir, metaPath, rest, fieldSep); err != nil {
		return err
	}

	var helpStr []string
	if removeds > 0 {
		helpStr = append(helpStr, fmt.Sprintf("%v repos removed from local disk.", removeds))
	}
	if unlinkeds > 0 {
		helpStr = append(helpStr, fmt.Sprintf("%v local dir unlinked to %s, files are untouched.",
			unlinkeds, env.GetRaw("strs.self-name")))
	}
	display.PrintTipTitle(cc.Screen, env, helpStr)
	return nil
}

func moveSavedFlowsToLocalDir(toDir string, cc *model.Cli, env *model.Env, cmd model.ParsedCmd) int {
	flowExt := env.GetRaw("strs.flow-ext")
	root := getFlowRoot(env, cmd)

	var count int
	_ = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
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
			return model.NewCmdError(cmd, fmt.Sprintf("rename file '%s' to '%s' failed: %v",
				path, destPath, err))
		}
		cmdPath, _ := getCmdPath(path, flowExt, cmd)
		_ = cc.Screen.Print(fmt.Sprintf(display.ColorHub("[%s]\n", env), cmdPath))
		_ = cc.Screen.Print(fmt.Sprintf(display.ColorProp("    - from:", env)+" %s\n", path))
		_ = cc.Screen.Print(fmt.Sprintf(display.ColorProp("    - to: ", env)+"%s\n", destPath))
		count += 1
		return nil
	})
	return count
}

func addRepoToHubAndLoadMods(
	cc *model.Cli,
	gitAddr meta.RepoAddr,
	argv model.ArgVals,
	screen model.Screen,
	env *model.Env,
	cmd model.ParsedCmd) (addrs []meta.RepoAddr, helpStrs []string, err error) {

	// A repo with this suffix should be a well controlled one, that we could assume some things
	repoExt := env.GetRaw("strs.mods-repo-ext")

	gitAddr.Addr = meta.NormalizeGitAddr(gitAddr.Addr)

	if !isOsCmdExists("git") {
		return nil, nil, model.NewCmdError(cmd, "cant't find 'git'")
	}

	path := getHubPath(env, cmd)
	mkdirErr := os.MkdirAll(path, os.ModePerm)
	if mkdirErr != nil && !os.IsExist(mkdirErr) {
		return nil, nil, model.NewCmdError(cmd, fmt.Sprintf("create hub path '%s' failed: %v", path, mkdirErr))
	}

	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	hubDir := env.GetRaw("sys.paths.hub")
	oldInfos, oldList, err := meta.ReadReposInfoFile(hubDir, metaPath, true, fieldSep)
	if err != nil {
		return nil, nil, err
	}
	finisheds := map[string]bool{}
	for i, info := range oldInfos {
		if info.Addr == gitAddr {
			info.OnOff = "on"
			oldInfos[i] = info
		}
		if info.OnOff != "on" {
			finisheds[info.Addr.Str()] = true
		}
	}

	selfName := env.GetRaw("strs.self-name")
	listFileName := env.GetRaw("strs.repos-file-name")
	var topRepoHelpStr string
	topRepoHelpStr, addrs, helpStrs, err = meta.UpdateRepoAndSubRepos(
		screen, env, finisheds, path, gitAddr, repoExt, listFileName, selfName, cmd)
	if err != nil {
		return nil, nil, err
	}

	addrs = append([]meta.RepoAddr{gitAddr}, addrs...)
	helpStrs = append([]string{topRepoHelpStr}, helpStrs...)

	var infos []meta.RepoInfo
	for i, addr := range addrs {
		if oldList[addr.Str()] {
			continue
		}
		var repoPath string
		repoPath, err = meta.GetRepoPath(path, addr)
		if err != nil {
			return nil, nil, err
		}
		info := meta.RepoInfo{Addr: addr, AddReason: gitAddr.Str(), Path: repoPath, HelpStr: helpStrs[i], OnOff: "on"}
		loadRepoMods(cc, env, info)
		infos = append(infos, info)
	}

	infos = append(oldInfos, infos...)
	if err = meta.WriteReposInfoFile(hubDir, metaPath, infos, fieldSep); err != nil {
		return nil, nil, err
	}
	return addrs, helpStrs, nil
}

func loadRepoMods(cc *model.Cli, env *model.Env, info meta.RepoInfo) {
	metaExt := env.GetRaw("strs.meta-ext")
	flowExt := env.GetRaw("strs.flow-ext")
	helpExt := env.GetRaw("strs.help-ext")
	abbrsSep := env.GetRaw("strs.abbrs-sep")
	envPathSep := env.GetRaw("strs.env-path-sep")
	reposFileName := env.GetRaw("strs.repos-file-name")
	panicRecover := env.GetBool("sys.panic.recover")

	// TODO: move this login to RepoAddr
	source := info.Addr.Str()
	if len(source) == 0 {
		source = info.Path
	}

	loadLocalMods(cc, info.Path, reposFileName, metaExt, flowExt, helpExt,
		abbrsSep, envPathSep, source, panicRecover)
}

func printInfoProps(screen model.Screen, env *model.Env, info meta.RepoInfo) {
	if len(info.HelpStr) > 0 {
		_ = screen.Print(fmt.Sprintf(display.ColorHelp("     '%s'\n", env), info.HelpStr))
	}
	_ = screen.Print(fmt.Sprintf(display.ColorProp("    - from: ", env)+"%s\n", getDisplayReason(info)))
	_ = screen.Print(fmt.Sprintf(display.ColorProp("    - path: ", env)+"%s\n", info.Path))
}

func getDisplayReason(info meta.RepoInfo) string {
	if info.AddReason == info.Addr.Str() {
		return "<manually-added>"
	}
	return info.AddReason
}

func checkFoundRepos(
	env *model.Env,
	cmd model.ParsedCmd,
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
		// PANIC: Runtime error - no repo matched
		panic(model.NewCmdError(cmd, fmt.Sprintf("no %s repo matched find string '%s'", status, findStr)))
	}
}

func getHubPath(env *model.Env, cmd model.ParsedCmd) string {
	path := env.GetRaw("sys.paths.hub")
	if len(path) == 0 {
		// PANIC: Programming error - hub path not configured
		panic(model.NewCmdError(cmd, "cant't get hub path from env, 'sys.paths.hub' is empty"))
	}
	return path
}

func matchFindRepoInfo(info meta.RepoInfo, findStr string) bool {
	if len(findStr) == 0 {
		return true
	}
	if strings.Index(info.Addr.Str(), findStr) >= 0 {
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
	if len(info.Addr.Addr) == 0 && strings.Index("local", findStr) >= 0 {
		return true
	}
	return false
}

func showHubFindTip(screen model.Screen, env *model.Env) {
	display.PrintTipTitle(screen, env,
		"try to search/show commands in this repo:",
		"",
		display.SuggestFindCmdsInRepo(env))
}

func getReposInfoPath(env *model.Env, cmd model.ParsedCmd) string {
	path := getHubPath(env, cmd)
	reposInfoFileName := env.GetRaw("strs.hub-file-name")
	if len(reposInfoFileName) == 0 {
		// PANIC: Programming error - hub meta file name not configured
		panic(model.NewCmdError(cmd, "cant't hub meta file name"))
	}
	return filepath.Join(path, reposInfoFileName)
}

func repoDisplayName(info meta.RepoInfo, env *model.Env) string {
	var name string
	if len(info.Addr.Addr) == 0 {
		name = filepath.Base(info.Path)
	} else {
		name = meta.AddrDisplayName(info.Addr)
	}
	return display.ColorHub("["+name+"]", env)
}

func disabledStr(env *model.Env) string {
	return display.ColorDisabled(" (disabled)", env)
}

func enabledStr(env *model.Env, str bool) string {
	if str {
		return display.ColorEnabled(" (enabled)", env)
	} else {
		return ""
	}
}

func purgedStr(env *model.Env, isLocal bool) string {
	if isLocal {
		return display.ColorDisabled(" (unlinked)", env)
	} else {
		return display.ColorDisabled(" (purged)", env)
	}
}
