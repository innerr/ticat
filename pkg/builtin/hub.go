package builtin

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
	meta "github.com/pingcap/ticat/pkg/proto/hub_meta"
)

func LoadModsFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	metaExt := env.GetRaw("strs.meta-ext")
	flowExt := env.GetRaw("strs.flow-ext")
	abbrsSep := env.GetRaw("strs.abbrs-sep")
	envPathSep := env.GetRaw("strs.env-path-sep")

	metaPath := getReposInfoPath(env, "LoadModsFromHub")
	fieldSep := env.GetRaw("strs.proto-sep")

	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	for _, info := range infos {
		if info.OnOff != "on" {
			continue
		}
		source := info.Addr
		if len(source) == 0 {
			source = info.Path
		}
		loadLocalMods(cc, info.Path, metaExt, flowExt, abbrsSep, envPathSep, source)
	}
	return true
}

func AddGitRepoToHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	addr := argv.GetRaw("git-address")
	if len(addr) == 0 {
		panic(fmt.Errorf("[AddGitRepoToHub] cant't get hub address"))
	}
	addRepoToHub(addr, argv, cc.Screen, env)
	return true
}

func AddGitDefaultToHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	addr := env.GetRaw("sys.hub.init-repo")
	if len(addr) == 0 {
		panic(fmt.Errorf("[AddGitDefaultToHub] cant't get init-repo address from env"))
	}
	addRepoToHub(addr, argv, cc.Screen, env)
	return true
}

func ListHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	metaPath := getReposInfoPath(env, "ListHub")
	fieldSep := env.GetRaw("strs.proto-sep")
	findStrs := getFindStrsFromArgv(argv)
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	screen := display.NewCacheScreen()
	listHub(screen, env, infos, findStrs...)
	if screen.OutputNum() <= 0 {
		display.PrintTipTitle(cc.Screen, env,
			"add more commands by adding more git repos to local with 'h.+'")
	} else {
		display.PrintTipTitle(cc.Screen, env,
			"show more hub usage by:")
	}
	screen.WriteTo(cc.Screen)
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
			screen.Print(enabledStr(env))
		}
		screen.Print("\n")
		screen.Print(fmt.Sprintf("     '%s'\n", info.HelpStr))
		if len(info.Addr) != 0 && name != info.Addr {
			screen.Print(fmt.Sprintf("    - addr: %s\n", info.Addr))
		}
		screen.Print(fmt.Sprintf("    - from: %s\n", getDisplayReason(info)))
		screen.Print(fmt.Sprintf("    - path: %s\n", info.Path))
	}
}

func RemoveAllFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	metaPath := getReposInfoPath(env, "RemoveAllFromHub")
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)

	for _, info := range infos {
		if len(info.Addr) != 0 {
			osRemoveDir(info.Path)
		}
		cc.Screen.Print(fmt.Sprintf("[%s]\n", repoDisplayName(info)))
		printInfoProps(cc.Screen, info)
		if len(info.Addr) != 0 {
			cc.Screen.Print("      (removed)\n")
		} else {
			cc.Screen.Print("      (unlinked)\n")
		}
	}

	err := os.Remove(metaPath)
	if err != nil {
		if os.IsNotExist(err) && len(infos) == 0 {
			return true
		}
		panic(fmt.Errorf("[RemoveAllFromHub] remove '%s' failed: %v", metaPath, err))
	}
	return true
}

func PurgeAllInactiveReposFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	purgeInactiveRepoFromHub("", cc, env)
	return true
}

func PurgeInactiveRepoFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	findStr := argv.GetRaw("find-str")
	if len(findStr) == 0 {
		panic(fmt.Errorf("[PurgeInactiveRepoFromHub] cant't get target repo addr from args"))
	}
	purgeInactiveRepoFromHub(findStr, cc, env)
	return true
}

func purgeInactiveRepoFromHub(findStr string, cc *core.Cli, env *core.Env) {
	metaPath := getReposInfoPath(env, "PurgeInactiveRepoFromHub")
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)

	var extracted []meta.RepoInfo
	var rest []meta.RepoInfo
	for _, info := range infos {
		if info.OnOff != "on" && (len(findStr) == 0 || strings.Index(info.Addr, findStr) >= 0) {
			extracted = append(extracted, info)
		} else {
			rest = append(rest, info)
		}
	}
	if len(extracted) == 0 {
		panic(fmt.Errorf("[PurgeInactiveRepoFromHub] cant't find repo by string '%s'", findStr))
	}

	for _, info := range extracted {
		if len(info.Addr) != 0 {
			osRemoveDir(info.Path)
		}
		cc.Screen.Print(fmt.Sprintf("[%s]\n", repoDisplayName(info)))
		printInfoProps(cc.Screen, info)
		cc.Screen.Print("      (purged)\n")
	}

	meta.WriteReposInfoFile(metaPath, rest, fieldSep)
}

// TODO: support partly update, not all
func UpdateHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	metaPath := getReposInfoPath(env, "UpdateHub")
	listFileName := env.GetRaw("strs.repos-file-name")
	repoExt := env.GetRaw("strs.mods-repo-ext")

	path := env.GetRaw("sys.paths.hub")
	if len(path) == 0 {
		panic(fmt.Errorf("[UpdateHub] cant't get hub path"))
	}

	fieldSep := env.GetRaw("strs.proto-sep")
	oldInfos, oldList := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	finisheds := map[string]bool{}
	for _, info := range oldInfos {
		if info.OnOff != "on" {
			finisheds[info.Addr] = true
		}
	}

	var infos []meta.RepoInfo

	for _, info := range oldInfos {
		if len(info.Addr) == 0 {
			continue
		}
		_, addrs, helpStrs := updateRepoAndSubRepos(
			cc.Screen, finisheds, path, info.Addr, repoExt, listFileName)
		for i, addr := range addrs {
			if oldList[addr] {
				continue
			}
			repoPath := getRepoPath(path, addr)
			infos = append(infos, meta.RepoInfo{addr, info.Addr, repoPath, helpStrs[i], "on"})
		}
	}

	infos = append(oldInfos, infos...)
	if len(infos) != len(oldInfos) {
		meta.WriteReposInfoFile(metaPath, infos, fieldSep)
	}
	return true
}

func EnableRepoInHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	metaPath := getReposInfoPath(env, "EnableRepoInHub")
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	findStr := argv.GetRaw("find-str")
	if len(findStr) == 0 {
		panic(fmt.Errorf("[EnableRepoInHub] cant't get target repo addr from args"))
	}

	extracted, rest := extractAddrFromList(infos, findStr)
	if len(extracted) == 0 {
		panic(fmt.Errorf("[EnableRepoInHub] cant't find repo by string '%s'", findStr))
	}

	for i, info := range extracted {
		if info.OnOff == "on" {
			continue
		}
		cc.Screen.Print(fmt.Sprintf("[%s] (enabled)\n", repoDisplayName(info)))
		printInfoProps(cc.Screen, info)
		info.OnOff = "on"
		extracted[i] = info
	}

	meta.WriteReposInfoFile(metaPath, append(rest, extracted...), fieldSep)
	return true
}

func DisableRepoInHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	metaPath := getReposInfoPath(env, "DisableRepoInHub")
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	findStr := argv.GetRaw("find-str")
	if len(findStr) == 0 {
		panic(fmt.Errorf("[DisableRepoInHub] cant't get target repo addr from args"))
	}

	extracted, rest := extractAddrFromList(infos, findStr)
	if len(extracted) == 0 {
		panic(fmt.Errorf("[DisableRepoInHub] cant't find repo by string '%s'", findStr))
	}

	for i, info := range extracted {
		if info.OnOff == "on" {
			cc.Screen.Print(fmt.Sprintf("[%s]%s\n", disabledStr(env), repoDisplayName(info)))
			cc.Screen.Print(fmt.Sprintf("    %s\n", info.Path))
			info.OnOff = "disabled"
			extracted[i] = info
		}
	}

	meta.WriteReposInfoFile(metaPath, append(rest, extracted...), fieldSep)
	return true
}

func MoveSavedFlowsToLocalDir(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	path := argv.GetRaw("path")
	if len(path) == 0 {
		panic("[MoveSavedFlowsToLocalDir] arg 'path' is empty")
	}

	stat, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		panic(fmt.Errorf("[MoveSavedFlowsToLocalDir] access path '%v' failed: %v",
			path, err))
	}

	if !os.IsNotExist(err) {
		if !stat.IsDir() {
			panic(fmt.Errorf("[MoveSavedFlowsToLocalDir] path '%v' exists but is not dir",
				path))
		}
		moveSavedFlowsToLocalDir(path, cc, env)
		return true
	}

	metaPath := getReposInfoPath(env, "LoadModsFromHub")
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
		panic(fmt.Errorf("[MoveSavedFlowsToLocalDir] cant't find dir by string '%s'", path))
	}
	if len(locals) > 1 {
		listHub(cc.Screen, env, locals)
		cc.Screen.Print(fmt.Sprintf(
			"\n[MoveSavedFlowsToLocalDir] cant't determine which dir by string '%s'\n",
			path))
		return false
	}

	moveSavedFlowsToLocalDir(locals[0].Path, cc, env)
	return true
}

func moveSavedFlowsToLocalDir(toDir string, cc *core.Cli, env *core.Env) {
	flowExt := env.GetRaw("strs.flow-ext")
	root := env.GetRaw("sys.paths.flows")
	if len(root) == 0 {
		panic(fmt.Errorf("[moveSavedFlowsToLocalDir] env 'sys.paths.flows' is empty"))
	}

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
			panic(fmt.Errorf("[moveSavedFlowsToLocalDir] rename file '%s' to '%s' failed: %v",
				path, destPath, err))
		}
		cmdPath := getCmdPath(path, flowExt)
		cc.Screen.Print(fmt.Sprintf("[%s]\n", cmdPath))
		cc.Screen.Print(fmt.Sprintf("    - from: %s\n", path))
		cc.Screen.Print(fmt.Sprintf("    - to: %s\n", destPath))
		return nil
	})
}

func AddLocalDirToHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	path := argv.GetRaw("path")
	if len(path) == 0 {
		panic("[AddLocalDirToHub] arg 'path' is empty")
	}

	stat, err := os.Stat(path)
	if err != nil {
		panic(fmt.Errorf("[AddLocalDirToHub] access path '%v' failed: %v",
			path, err))
	}
	if !stat.IsDir() {
		panic(fmt.Errorf("[AddLocalDirToHub] path '%v' is not dir", path))
	}

	path, err = filepath.Abs(path)
	if err != nil {
		panic(fmt.Errorf("[AddLocalDirToHub] get abs path of '%v' failed: %v",
			path, err))
	}

	metaPath := getReposInfoPath(env, "addRepoToHub")
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	found := false
	for i, info := range infos {
		if info.Path == path {
			if info.OnOff == "on" {
				cc.Screen.Print(fmt.Sprintf("[%s] (exists)\n", repoDisplayName(info)))
				printInfoProps(cc.Screen, info)
				return true
			}
			info.OnOff = "on"
			infos[i] = info
			cc.Screen.Print(fmt.Sprintf("[%s] (%s)\n", repoDisplayName(info), info.OnOff))
			printInfoProps(cc.Screen, info)
			found = true
			break
		}
	}

	if !found {
		listFileName := env.GetRaw("strs.repos-file-name")
		listFilePath := filepath.Join(path, listFileName)
		helpStr, _, _ := meta.ReadRepoListFromFile(listFilePath)
		info := meta.RepoInfo{"", "<local>", path, helpStr, "on"}
		infos = append(infos, info)
		cc.Screen.Print(fmt.Sprintf("[%s]\n", repoDisplayName(info)))
		printInfoProps(cc.Screen, info)
	}
	meta.WriteReposInfoFile(metaPath, infos, fieldSep)

	// TODO: load mods now?
	return true
}

func addRepoToHub(
	gitAddr string,
	argv core.ArgVals,
	screen core.Screen,
	env *core.Env) (addrs []string, helpStrs []string) {

	// A repo with this suffix should be a well controlled one, that we could assume some things
	repoExt := env.GetRaw("strs.mods-repo-ext")

	gitAddr = normalizeGitAddr(gitAddr)

	if !isOsCmdExists("git") {
		panic(fmt.Errorf("[addRepoToHub] cant't find 'git'"))
	}

	path := env.GetRaw("sys.paths.hub")
	if len(path) == 0 {
		panic(fmt.Errorf("[addRepoToHub] cant't get hub path"))
	}
	err := os.MkdirAll(path, os.ModePerm)
	if os.IsExist(err) {
		panic(fmt.Errorf("[addRepoToHub] create hub path '%s' failed: %v", path, err))
	}

	metaPath := getReposInfoPath(env, "addRepoToHub")
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

	listFileName := env.GetRaw("strs.repos-file-name")
	var topRepoHelpStr string
	topRepoHelpStr, addrs, helpStrs = updateRepoAndSubRepos(
		screen, finisheds, path, gitAddr, repoExt, listFileName)

	addrs = append([]string{gitAddr}, addrs...)
	helpStrs = append([]string{topRepoHelpStr}, helpStrs...)

	var infos []meta.RepoInfo
	for i, addr := range addrs {
		if oldList[addr] {
			continue
		}
		repoPath := getRepoPath(path, addr)
		infos = append(infos, meta.RepoInfo{addr, gitAddr, repoPath, helpStrs[i], "on"})
	}

	infos = append(oldInfos, infos...)
	meta.WriteReposInfoFile(metaPath, infos, fieldSep)
	return
}

func updateRepoAndSubRepos(
	screen core.Screen,
	finisheds map[string]bool,
	hubPath string,
	gitAddr string,
	repoExt string,
	listFileName string) (topRepoHelpStr string, addrs []string, helpStrs []string) {

	if finisheds[gitAddr] {
		return
	}
	topRepoHelpStr, addrs, helpStrs = updateRepoAndReadSubList(
		screen, hubPath, gitAddr, listFileName)
	finisheds[gitAddr] = true

	for i, addr := range addrs {
		subTopHelpStr, subAddrs, subHelpStrs := updateRepoAndSubRepos(
			screen, finisheds, hubPath, addr, repoExt, listFileName)
		// If a repo has no help-str from hub-repo list, try to get the title from it's README
		if len(helpStrs[i]) == 0 && len(subTopHelpStr) != 0 {
			helpStrs[i] = subTopHelpStr
		}
		addrs = append(addrs, subAddrs...)
		helpStrs = append(helpStrs, subHelpStrs...)
	}

	return topRepoHelpStr, addrs, helpStrs
}

func updateRepoAndReadSubList(
	screen core.Screen,
	hubPath string,
	gitAddr string,
	listFileName string) (helpStr string, addrs []string, helpStrs []string) {

	name := addrDisplayName(gitAddr)
	repoPath := getRepoPath(hubPath, gitAddr)
	var cmdStrs []string

	stat, err := os.Stat(repoPath)
	var pwd string
	if !os.IsNotExist(err) {
		if !stat.IsDir() {
			panic(fmt.Errorf("[updateRepoAndReadSubList] repo path '%v' exists but is not dir",
				repoPath))
		}
		screen.Print(fmt.Sprintf("[%s] => git update\n", name))
		cmdStrs = []string{"git", "pull", "--recurse-submodules"}
		pwd = repoPath
	} else {
		screen.Print(fmt.Sprintf("[%s] => git clone\n", name))
		cmdStrs = []string{"git", "clone", "--recursive", gitAddr, repoPath}
	}

	cmd := exec.Command(cmdStrs[0], cmdStrs[1:]...)
	if len(pwd) != 0 {
		cmd.Dir = pwd
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(fmt.Errorf("[updateRepoAndReadSubList] run '%v' failed: %v", cmdStrs, err))
	}
	listFilePath := filepath.Join(repoPath, listFileName)
	return meta.ReadRepoListFromFile(listFilePath)
}

func extractAddrFromList(
	infos []meta.RepoInfo,
	findStr string) (extracted []meta.RepoInfo, rest []meta.RepoInfo) {

	for _, info := range infos {
		findInStr := info.Addr
		if len(info.Addr) == 0 {
			findInStr = info.Path
		}
		if strings.Index(findInStr, findStr) >= 0 {
			extracted = append(extracted, info)
		} else {
			rest = append(rest, info)
		}
	}
	return
}

func normalizeGitAddr(addr string) string {
	if strings.HasPrefix(strings.ToLower(addr), "http") {
		return addr
	}
	if strings.HasPrefix(strings.ToLower(addr), "git") {
		return addr
	}
	return "git@github.com:" + addr
}

func gitAddrAbbr(addr string) (abbr string) {
	// TODO: support other git platform
	abbrExtractors := []func(string) string{
		githubAddrAbbr,
	}
	for _, extractor := range abbrExtractors {
		abbr = extractor(addr)
		if len(abbr) != 0 {
			break
		}
	}
	return
}

func repoDisplayName(info meta.RepoInfo) string {
	if len(info.Addr) == 0 {
		return filepath.Base(info.Path)
	}
	return addrDisplayName(info.Addr)
}

func addrDisplayName(addr string) string {
	abbr := gitAddrAbbr(addr)
	if len(abbr) == 0 {
		return addr
	}
	return abbr
}

func githubAddrAbbr(addr string) (abbr string) {
	httpPrefix := "http://github.com/"
	if strings.HasPrefix(strings.ToLower(addr), httpPrefix) {
		return addr[len(httpPrefix):]
	}
	sshPrefix := "git@github.com:"
	if strings.HasPrefix(strings.ToLower(addr), sshPrefix) {
		return addr[len(sshPrefix):]
	}
	return
}

func isOsCmdExists(cmd string) bool {
	path, err := exec.LookPath(cmd)
	return err == nil && len(path) > 0
}

func getReposInfoPath(env *core.Env, funcName string) string {
	path := env.GetRaw("sys.paths.hub")
	if len(path) == 0 {
		panic(fmt.Errorf("[addRepoToHub] cant't get hub path"))
	}
	reposInfoFileName := env.GetRaw("strs.hub-file-name")
	if len(reposInfoFileName) == 0 {
		panic(fmt.Errorf("[%s] cant't hub meta path", funcName))
	}
	return filepath.Join(path, reposInfoFileName)
}

func getRepoPath(hubPath string, gitAddr string) string {
	return filepath.Join(hubPath, filepath.Base(gitAddr))
}

func printInfoProps(screen core.Screen, info meta.RepoInfo) {
	screen.Print(fmt.Sprintf("     '%s'\n", info.HelpStr))
	screen.Print(fmt.Sprintf("    - from: %s\n", getDisplayReason(info)))
	screen.Print(fmt.Sprintf("    - path: %s\n", info.Path))
}

func getDisplayReason(info meta.RepoInfo) string {
	if info.AddReason == info.Addr {
		return "<manually-added>"
	}
	return info.AddReason
}

func osRemoveDir(path string) {
	path = strings.TrimSpace(path)
	if len(path) <= 1 {
		panic(fmt.Errorf("[osRemoveDir] removing path '%v', looks not right", path))
	}
	err := os.RemoveAll(path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		panic(fmt.Errorf("[osRemoveDir] remove repo '%s' failed: %v", path, err))
	}
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

func disabledStr(env *core.Env) string {
	if env.GetBool("display.utf8.symbols") {
		return "⛔(disabled)"
	}
	return " (disabled)"
}

func enabledStr(env *core.Env) string {
	if env.GetBool("display.utf8.symbols") {
		return "✅"
	}
	return ""
}
