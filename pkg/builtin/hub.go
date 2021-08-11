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
)

func LoadModsFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env, flow []core.ParsedCmd) bool {
	metaExt := env.GetRaw("strs.meta-ext")
	flowExt := env.GetRaw("strs.flow-ext")
	helpExt := env.GetRaw("strs.help-ext")
	abbrsSep := env.GetRaw("strs.abbrs-sep")
	envPathSep := env.GetRaw("strs.env-path-sep")
	reposFileName := env.GetRaw("strs.repos-file-name")

	metaPath := getReposInfoPath(env, flow[0])
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
	return true
}

func AddGitRepoToHub(argv core.ArgVals, cc *core.Cli, env *core.Env, flow []core.ParsedCmd) bool {
	cmd := flow[0]
	addr := getAndCheckArg(argv, env, cmd, "git-address")
	addRepoToHub(addr, argv, cc.Screen, env, cmd)
	showHubFindTip(cc.Screen, env)
	return true
}

func AddGitDefaultToHub(argv core.ArgVals, cc *core.Cli, env *core.Env, flow []core.ParsedCmd) bool {
	cmd := flow[0]
	addr := env.GetRaw("sys.hub.init-repo")
	if len(addr) == 0 {
		panic(core.NewCmdError(cmd, "cant't get init-repo address from env, 'sys.hub.init-repo' is empty"))
	}
	addRepoToHub(addr, argv, cc.Screen, env, cmd)
	showHubFindTip(cc.Screen, env)
	return true
}

func ListHub(argv core.ArgVals, cc *core.Cli, env *core.Env, flow []core.ParsedCmd) bool {
	cmd := flow[0]
	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	findStrs := getFindStrsFromArgv(argv)
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
	return true
}

func RemoveAllFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env, flow []core.ParsedCmd) bool {
	cmd := flow[0]
	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)

	for _, info := range infos {
		if !info.IsLocal() {
			osRemoveDir(info.Path, cmd)
		}
		cc.Screen.Print(fmt.Sprintf("[%s]%s\n", repoDisplayName(info), purgedStr(env, info.IsLocal())))
		printInfoProps(cc.Screen, info)
	}

	err := os.Remove(metaPath)
	if err != nil {
		if os.IsNotExist(err) && len(infos) == 0 {
			return true
		}
		panic(core.WrapCmdError(cmd, fmt.Errorf("remove '%s' failed: %v", metaPath, err)))
	}

	display.PrintTipTitle(cc.Screen, env,
		"hub now is empty.",
		"",
		"add more git repos to get more avaialable commands:",
		"",
		display.SuggestHubAddShort(env))

	return true
}

func PurgeAllInactiveReposFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env, flow []core.ParsedCmd) bool {
	purgeInactiveRepoFromHub("", cc, env, flow[0])
	return true
}

func PurgeInactiveRepoFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env, flow []core.ParsedCmd) bool {
	cmd := flow[0]
	findStr := getAndCheckArg(argv, env, cmd, "find-str")
	purgeInactiveRepoFromHub(findStr, cc, env, cmd)
	return true
}

func UpdateHub(argv core.ArgVals, cc *core.Cli, env *core.Env, flow []core.ParsedCmd) bool {
	cmd := flow[0]
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
			cc.Screen, finisheds, path, info.Addr, repoExt, listFileName, selfName, cmd)
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
	return true
}

func EnableRepoInHub(argv core.ArgVals, cc *core.Cli, env *core.Env, flow []core.ParsedCmd) bool {
	cmd := flow[0]
	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	findStr := getAndCheckArg(argv, env, cmd, "find-str")

	extracted, rest := meta.ExtractAddrFromList(infos, findStr)
	checkFoundRepos(env, cmd, extracted, findStr)

	var count int
	for i, info := range extracted {
		if info.OnOff == "on" {
			continue
		}
		count += 1
		cc.Screen.Print(fmt.Sprintf("[%s] (enabled)\n", repoDisplayName(info)))
		printInfoProps(cc.Screen, info)
		info.OnOff = "on"
		extracted[i] = info
	}

	meta.WriteReposInfoFile(metaPath, append(rest, extracted...), fieldSep)

	if count > 0 {
		display.PrintTipTitle(cc.Screen, env,
			"add a disabled repo manually will enable it")
	} else {
		display.PrintTipTitle(cc.Screen, env,
			"no disabled repos matched find string '"+findStr+"'")
	}
	return true
}

func DisableRepoInHub(argv core.ArgVals, cc *core.Cli, env *core.Env, flow []core.ParsedCmd) bool {
	cmd := flow[0]
	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	findStr := getAndCheckArg(argv, env, cmd, "find-str")

	extracted, rest := meta.ExtractAddrFromList(infos, findStr)
	checkFoundRepos(env, cmd, extracted, findStr)

	var count int
	for i, info := range extracted {
		if info.OnOff == "on" {
			cc.Screen.Print(fmt.Sprintf("[%s] (disabled)\n", repoDisplayName(info)))
			printInfoProps(cc.Screen, info)
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
			"no enabled repos matched find string '"+findStr+"'")
	}
	return true
}

func AddLocalDirToHub(argv core.ArgVals, cc *core.Cli, env *core.Env, flow []core.ParsedCmd) bool {
	cmd := flow[0]
	path := getAndCheckArg(argv, env, cmd, "path")

	stat, err := os.Stat(path)
	if err != nil {
		panic(core.WrapCmdError(cmd, fmt.Errorf("access path '%v' failed: %v", path, err)))
	}
	if !stat.IsDir() {
		panic(core.WrapCmdError(cmd, fmt.Errorf("path '%v' is not dir", path)))
	}

	path, err = filepath.Abs(path)
	if err != nil {
		panic(core.WrapCmdError(cmd, fmt.Errorf("get abs path of '%v' failed: %v", path, err)))
	}

	screen := display.NewCacheScreen()

	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	found := false
	for i, info := range infos {
		if info.Path == path {
			if info.OnOff == "on" {
				screen.Print(fmt.Sprintf("[%s] (exists)\n", repoDisplayName(info)))
				printInfoProps(screen, info)
				display.PrintTipTitle(cc.Screen, env,
					"local dir already in hub, nothing to do")
				return true
			}
			info.OnOff = "on"
			infos[i] = info
			screen.Print(fmt.Sprintf("[%s] (enabled)\n", repoDisplayName(info)))
			printInfoProps(screen, info)
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
		screen.Print(fmt.Sprintf("[%s]\n", repoDisplayName(info)))
		printInfoProps(screen, info)
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
	return true
}

func MoveSavedFlowsToLocalDir(argv core.ArgVals, cc *core.Cli, env *core.Env, flow []core.ParsedCmd) bool {
	// TODO: do accurate matching before search
	cmd := flow[0]
	path := argv.GetRaw("path")
	if len(path) != 0 {
		stat, err := os.Stat(path)
		if err != nil && !os.IsNotExist(err) {
			panic(core.WrapCmdError(cmd, fmt.Errorf("access path '%v' failed: %v", path, err)))
		}

		if !os.IsNotExist(err) {
			if !stat.IsDir() {
				panic(core.WrapCmdError(cmd, fmt.Errorf("path '%v' exists but is not a dir", path)))
			}
			moveSavedFlowsToLocalDir(path, cc, env, cmd)
			display.PrintTipTitle(cc.Screen, env,
				"all saved flow moved to '"+path+"'.")
			return true
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
		display.PrintTipTitle(cc.Screen, env,
			fmt.Sprintf("cant't find matched dir by string '%s'.", path),
			"notice this command only search added local dirs, not repos")
		return false
	}
	if len(locals) > 1 {
		display.PrintErrTitle(cc.Screen, env,
			"cant't determine which dir by string '"+path+"'.",
			"only could move to the one and only matched dir.",
			"", "current matcheds:")
		listHub(cc.Screen, env, locals)
		return false
	}

	moveSavedFlowsToLocalDir(locals[0].Path, cc, env, cmd)

	display.PrintTipTitle(cc.Screen, env,
		"all saved flow moved to '"+locals[0].Path+"', it's the only local dir in hub")
	return true
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
		name := repoDisplayName(info)
		screen.Print(fmt.Sprintf("[%s]", name))
		if info.OnOff != "on" {
			screen.Print(disabledStr(env))
		} else {
			screen.Print(enabledStr(env, false))
		}
		screen.Print("\n")
		if len(info.HelpStr) > 0 {
			screen.Print(fmt.Sprintf("     '%s'\n", info.HelpStr))
		}
		if len(info.Addr) != 0 && name != info.Addr {
			screen.Print(fmt.Sprintf("    - addr: %s\n", info.Addr))
		}
		screen.Print(fmt.Sprintf("    - from: %s\n", getDisplayReason(info)))
		screen.Print(fmt.Sprintf("    - path: %s\n", info.Path))
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
	checkFoundRepos(env, cmd, extracted, findStr)

	var unlinkeds int
	var removeds int

	for _, info := range extracted {
		if !info.IsLocal() {
			osRemoveDir(info.Path, cmd)
			removeds += 1
		} else {
			unlinkeds += 1
		}
		cc.Screen.Print(fmt.Sprintf("[%s]%s\n", repoDisplayName(info), purgedStr(env, info.IsLocal())))
		printInfoProps(cc.Screen, info)
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
	root := env.GetRaw("sys.paths.flows")
	if len(root) == 0 {
		panic(core.NewCmdError(cmd, "env 'sys.paths.flows' is empty"))
	}

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

		err = os.Rename(path, destPath)
		if err != nil {
			panic(core.WrapCmdError(cmd, fmt.Errorf("rename file '%s' to '%s' failed: %v",
				path, destPath, err)))
		}
		cmdPath := getCmdPath(path, flowExt, cmd)
		cc.Screen.Print(fmt.Sprintf("[%s]\n", cmdPath))
		cc.Screen.Print(fmt.Sprintf("    - from: %s\n", path))
		cc.Screen.Print(fmt.Sprintf("    - to: %s\n", destPath))
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
		panic(core.WrapCmdError(cmd, fmt.Errorf("create hub path '%s' failed: %v", path, err)))
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
		screen, finisheds, path, gitAddr, repoExt, listFileName, selfName, cmd)

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

func printInfoProps(screen core.Screen, info meta.RepoInfo) {
	if len(info.HelpStr) > 0 {
		screen.Print(fmt.Sprintf("     '%s'\n", info.HelpStr))
	}
	screen.Print(fmt.Sprintf("    - from: %s\n", getDisplayReason(info)))
	screen.Print(fmt.Sprintf("    - path: %s\n", info.Path))
}

func getDisplayReason(info meta.RepoInfo) string {
	if info.AddReason == info.Addr {
		return "<manually-added>"
	}
	return info.AddReason
}

func checkFoundRepos(env *core.Env, cmd core.ParsedCmd, infos []meta.RepoInfo, findStr string) {
	if len(infos) == 0 {
		panic(core.WrapCmdError(cmd, fmt.Errorf("cant't find repo by string '%s'", findStr)))
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

func repoDisplayName(info meta.RepoInfo) string {
	if len(info.Addr) == 0 {
		return filepath.Base(info.Path)
	}
	return meta.AddrDisplayName(info.Addr)
}

func disabledStr(env *core.Env) string {
	if env.GetBool("display.utf8.symbols") {
		errStr := env.GetRaw("display.utf8.symbols.err")
		return errStr + "(disabled)"
	} else {
		return " (disabled)"
	}
}

func enabledStr(env *core.Env, str bool) string {
	if str {
		return " (enabled)"
	} else {
		return ""
	}
}

func purgedStr(env *core.Env, isLocal bool) string {
	if isLocal {
		return " (unlinked)"
	} else {
		return " (purged)"
	}
}
