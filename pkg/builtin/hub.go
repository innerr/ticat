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

func LoadModsFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env, cmd core.ParsedCmd) bool {
	metaExt := env.GetRaw("strs.meta-ext")
	flowExt := env.GetRaw("strs.flow-ext")
	abbrsSep := env.GetRaw("strs.abbrs-sep")
	envPathSep := env.GetRaw("strs.env-path-sep")

	metaPath := getReposInfoPath(env, cmd)
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

func AddGitRepoToHub(argv core.ArgVals, cc *core.Cli, env *core.Env, cmd core.ParsedCmd) bool {
	addr := getAndCheckArg(argv, env, cmd, "git-address")
	addRepoToHub(addr, argv, cc.Screen, env, cmd)
	showFindTip(cc.Screen, env)
	return true
}

func AddGitDefaultToHub(argv core.ArgVals, cc *core.Cli, env *core.Env, cmd core.ParsedCmd) bool {
	addr := env.GetRaw("sys.hub.init-repo")
	if len(addr) == 0 {
		panic(core.NewCmdError(cmd, "cant't get init-repo address from env, 'sys.hub.init-repo' is empty"))
	}
	addRepoToHub(addr, argv, cc.Screen, env, cmd)
	showFindTip(cc.Screen, env)
	return true
}

func ListHub(argv core.ArgVals, cc *core.Cli, env *core.Env, cmd core.ParsedCmd) bool {
	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	findStrs := getFindStrsFromArgv(argv)
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)

	screen := display.NewCacheScreen()

	listHub(screen, env, infos, findStrs...)
	if screen.OutputNum() <= 0 {
		helpStr := []string{
			"'hub' manages all added git repos, now it's empty.",
			"",
			"add more git repos to get more avaialable commands:",
			"",
		}
		helpStr = append(helpStr, display.SuggestHubAddShort(env)...)
		display.PrintTipTitle(cc.Screen, env, helpStr...)
	} else {
		display.PrintTipTitle(cc.Screen, env, "repo list in hub:")
		screen.WriteTo(cc.Screen)
		cmdName := cmd.DisplayPath(cc.Cmds.Strs.PathSep, true)
		helpStr := []string{
			"command branch '" + cmdName + "' manages the repos in local disk.",
			"", "to see more usage:", "",
		}
		helpStr = append(helpStr, display.SuggestHubBranch(env)...)
		display.PrintTipTitle(cc.Screen, env, helpStr...)
	}
	return true
}

func RemoveAllFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env, cmd core.ParsedCmd) bool {
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

	helpStr := []string{
		"hub now is empty.",
		"",
		"add more git repos to get more avaialable commands:",
		"",
	}
	helpStr = append(helpStr, display.SuggestHubAddShort(env)...)
	display.PrintTipTitle(cc.Screen, env, helpStr...)

	return true
}

func PurgeAllInactiveReposFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env, cmd core.ParsedCmd) bool {
	purgeInactiveRepoFromHub("", cc, env, cmd)
	return true
}

func PurgeInactiveRepoFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env, cmd core.ParsedCmd) bool {
	findStr := getAndCheckArg(argv, env, cmd, "find-str")
	purgeInactiveRepoFromHub(findStr, cc, env, cmd)
	return true
}

func UpdateHub(argv core.ArgVals, cc *core.Cli, env *core.Env, cmd core.ParsedCmd) bool {
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
		_, addrs, helpStrs := updateRepoAndSubRepos(
			cc.Screen, finisheds, path, info.Addr, repoExt, listFileName, selfName, cmd)
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

	display.PrintTipTitle(cc.Screen, env, fmt.Sprintf(
		"local dir could also add to %s, use command 'h.add.local'",
		env.GetRaw("strs.self-name")))
	return true
}

func EnableRepoInHub(argv core.ArgVals, cc *core.Cli, env *core.Env, cmd core.ParsedCmd) bool {
	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	findStr := getAndCheckArg(argv, env, cmd, "find-str")

	extracted, rest := extractAddrFromList(infos, findStr)
	checkFoundRepos(env, cmd, extracted, findStr)

	var count int
	for i, info := range extracted {
		if info.OnOff == "on" {
			continue
		}
		count += 1
		cc.Screen.Print(fmt.Sprintf("[%s]%s\n", repoDisplayName(info), enabledStr(env, true)))
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

func DisableRepoInHub(argv core.ArgVals, cc *core.Cli, env *core.Env, cmd core.ParsedCmd) bool {
	metaPath := getReposInfoPath(env, cmd)
	fieldSep := env.GetRaw("strs.proto-sep")
	infos, _ := meta.ReadReposInfoFile(metaPath, true, fieldSep)
	findStr := getAndCheckArg(argv, env, cmd, "find-str")

	extracted, rest := extractAddrFromList(infos, findStr)
	checkFoundRepos(env, cmd, extracted, findStr)

	var count int
	for i, info := range extracted {
		if info.OnOff == "on" {
			cc.Screen.Print(fmt.Sprintf("[%s]%s\n", repoDisplayName(info), disabledStr(env)))
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
			"a disabled repo will not update, commands in it are not available")
	} else {
		display.PrintTipTitle(cc.Screen, env,
			"no enabled repos matched find string '"+findStr+"'")
	}
	return true
}

func AddLocalDirToHub(argv core.ArgVals, cc *core.Cli, env *core.Env, cmd core.ParsedCmd) bool {
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

	metaPath := getReposInfoPath(env, cmd)
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
		helpStr, _, _ := meta.ReadRepoListFromFile(env.GetRaw("strs.self-name"), listFilePath)
		info := meta.RepoInfo{"", "<local>", path, helpStr, "on"}
		infos = append(infos, info)
		cc.Screen.Print(fmt.Sprintf("[%s]\n", repoDisplayName(info)))
		printInfoProps(cc.Screen, info)
	}
	meta.WriteReposInfoFile(metaPath, infos, fieldSep)

	display.PrintTipTitle(cc.Screen, env,
		"need two steps to remove a repo or unlink a dir: disable, purge")

	// TODO: load mods now?
	return true
}

func MoveSavedFlowsToLocalDir(argv core.ArgVals, cc *core.Cli, env *core.Env, cmd core.ParsedCmd) bool {
	path := argv.GetRaw("path")
	if len(path) == 0 {
		panic(core.NewCmdError(cmd, "arg 'path' is empty"))
	}

	stat, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		panic(core.WrapCmdError(cmd, fmt.Errorf("access path '%v' failed: %v", path, err)))
	}

	if !os.IsNotExist(err) {
		if !stat.IsDir() {
			panic(core.WrapCmdError(cmd, fmt.Errorf("path '%v' exists but is not a dir", path)))
		}
		moveSavedFlowsToLocalDir(path, cc, env, cmd)
		return true
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
			"only could move to the one and only one matched dir.",
			"", "current matcheds:")
		listHub(cc.Screen, env, locals)
		return false
	}

	moveSavedFlowsToLocalDir(locals[0].Path, cc, env, cmd)
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
	display.PrintTipTitle(cc.Screen, env, helpStr...)
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
		cmdPath := getCmdPath(path, flowExt)
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

	gitAddr = normalizeGitAddr(gitAddr)

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
	topRepoHelpStr, addrs, helpStrs = updateRepoAndSubRepos(
		screen, finisheds, path, gitAddr, repoExt, listFileName, selfName, cmd)

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
	listFileName string,
	selfName string,
	cmd core.ParsedCmd) (topRepoHelpStr string, addrs []string, helpStrs []string) {

	if finisheds[gitAddr] {
		return
	}
	topRepoHelpStr, addrs, helpStrs = updateRepoAndReadSubList(
		screen, hubPath, gitAddr, listFileName, selfName, cmd)
	finisheds[gitAddr] = true

	for i, addr := range addrs {
		subTopHelpStr, subAddrs, subHelpStrs := updateRepoAndSubRepos(
			screen, finisheds, hubPath, addr, repoExt, listFileName, selfName, cmd)
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
	listFileName string,
	selfName string,
	cmd core.ParsedCmd) (helpStr string, addrs []string, helpStrs []string) {

	name := addrDisplayName(gitAddr)
	repoPath := getRepoPath(hubPath, gitAddr)
	var cmdStrs []string

	stat, err := os.Stat(repoPath)
	var pwd string
	if !os.IsNotExist(err) {
		if !stat.IsDir() {
			panic(core.WrapCmdError(cmd, fmt.Errorf("repo path '%v' exists but is not dir",
				repoPath)))
		}
		screen.Print(fmt.Sprintf("[%s] => git update\n", name))
		cmdStrs = []string{"git", "pull", "--recurse-submodules"}
		pwd = repoPath
	} else {
		screen.Print(fmt.Sprintf("[%s] => git clone\n", name))
		cmdStrs = []string{"git", "clone", "--recursive", gitAddr, repoPath}
	}

	c := exec.Command(cmdStrs[0], cmdStrs[1:]...)
	if len(pwd) != 0 {
		c.Dir = pwd
	}
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err = c.Run()
	if err != nil {
		panic(core.WrapCmdError(cmd, fmt.Errorf("run '%v' failed: %v", cmdStrs, err)))
	}
	listFilePath := filepath.Join(repoPath, listFileName)
	return meta.ReadRepoListFromFile(selfName, listFilePath)
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

func getReposInfoPath(env *core.Env, cmd core.ParsedCmd) string {
	path := getHubPath(env, cmd)
	reposInfoFileName := env.GetRaw("strs.hub-file-name")
	if len(reposInfoFileName) == 0 {
		panic(core.NewCmdError(cmd, "cant't hub meta file name"))
	}
	return filepath.Join(path, reposInfoFileName)
}

func getRepoPath(hubPath string, gitAddr string) string {
	return filepath.Join(hubPath, filepath.Base(gitAddr))
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

func disabledStr(env *core.Env) string {
	if env.GetBool("display.utf8.symbols") {
		return "❎(disabled)"
	} else {
		return " (disabled)"
	}
}

func enabledStr(env *core.Env, str bool) string {
	if env.GetBool("display.utf8.symbols") {
		if str {
			return "✅(enabled)"
		} else {
			return "✅"
		}
	} else {
		if str {
			return " (enabled)"
		} else {
			return ""
		}
	}
}

func purgedStr(env *core.Env, isLocal bool) string {
	if env.GetBool("display.utf8.symbols") {
		if isLocal {
			return "❎(unlinked)"
		} else {
			return "🚮(purged)"
		}
	} else {
		if isLocal {
			return " (unlinked)"
		} else {
			return " (purged)"
		}
	}
}

func showFindTip(screen core.Screen, env *core.Env) {
	helpStr := []string{
		"try to search commands by tag @ready, it means 'out-of-the-box':",
		"",
	}
	helpStr = append(helpStr, display.SuggestFindRepoTag(env)...)
	display.PrintTipTitle(screen, env, helpStr...)
}
