package builtin

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func LoadModsFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	metaExt := env.GetRaw("strs.meta-ext")
	flowExt := env.GetRaw("strs.flow-ext")
	abbrsSep := env.GetRaw("strs.abbrs-sep")
	envPathSep := env.GetRaw("strs.env-path-sep")
	fieldSep := env.GetRaw("strs.proto-sep")

	metaPath := getReposInfoPath(env, "LoadModsFromHub")
	infos, _ := readReposInfoFile(metaPath, true, fieldSep)
	for _, info := range infos {
		if info.OnOff != "on" {
			continue
		}
		loadLocalMods(cc, info.Path, metaExt, flowExt, abbrsSep, envPathSep)
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
	fieldSep := env.GetRaw("strs.proto-sep")
	metaPath := getReposInfoPath(env, "ListHub")
	infos, _ := readReposInfoFile(metaPath, true, fieldSep)
	listHub(cc.Screen, infos)
	return true
}

func listHub(screen core.Screen, infos []repoInfo) {
	for _, info := range infos {
		name := addrDisplayName(info.Addr)
		screen.Print(fmt.Sprintf("[%s]", name))
		if info.OnOff != "on" {
			screen.Print(" (" + info.OnOff + ")")
		}
		screen.Print("\n")
		screen.Print(fmt.Sprintf("     '%s'\n", info.HelpStr))
		if name != info.Addr {
			screen.Print(fmt.Sprintf("    - full: %s\n", info.Addr))
		}
		screen.Print(fmt.Sprintf("    - from: %s\n", getDisplayReason(info)))
		screen.Print(fmt.Sprintf("    - path: %s\n", info.Path))
	}
}

func RemoveAllFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	fieldSep := env.GetRaw("strs.proto-sep")

	metaPath := getReposInfoPath(env, "RemoveAllFromHub")
	infos, _ := readReposInfoFile(metaPath, true, fieldSep)

	for _, info := range infos {
		osRemoveDir(info.Path)
		cc.Screen.Print(fmt.Sprintf("[%s]\n", addrDisplayName(info.Addr)))
		printInfoProps(cc.Screen, info)
		cc.Screen.Print("      (removed)\n")
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
	fieldSep := env.GetRaw("strs.proto-sep")

	metaPath := getReposInfoPath(env, "PurgeInactiveRepoFromHub")
	infos, _ := readReposInfoFile(metaPath, true, fieldSep)

	var extracted []repoInfo
	var rest []repoInfo
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
		osRemoveDir(info.Path)
		cc.Screen.Print(fmt.Sprintf("[%s]\n", addrDisplayName(info.Addr)))
		printInfoProps(cc.Screen, info)
		cc.Screen.Print("      (purged)\n")
	}

	writeReposInfoFile(metaPath, rest, fieldSep)
}

func UpdateHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	metaPath := getReposInfoPath(env, "UpdateHub")
	fieldSep := env.GetRaw("strs.proto-sep")
	listFileName := env.GetRaw("strs.repos-file-name")
	repoExt := env.GetRaw("strs.mods-repo-ext")

	path := env.GetRaw("sys.paths.hub")
	if len(path) == 0 {
		panic(fmt.Errorf("[UpdateHub] cant't get hub path"))
	}

	oldInfos, oldList := readReposInfoFile(metaPath, true, fieldSep)
	finisheds := map[string]bool{}
	for _, info := range oldInfos {
		if info.OnOff != "on" {
			finisheds[info.Addr] = true
		}
	}

	var infos []repoInfo

	for _, info := range oldInfos {
		_, addrs, helpStrs := updateRepoAndSubRepos(
			cc.Screen, finisheds, path, info.Addr, repoExt, listFileName)
		for i, addr := range addrs {
			if oldList[addr] {
				continue
			}
			repoPath := getRepoPath(path, addr)
			infos = append(infos, repoInfo{addr, info.Addr, repoPath, helpStrs[i], "on"})
		}
	}

	infos = append(oldInfos, infos...)
	if len(infos) != len(oldInfos) {
		writeReposInfoFile(metaPath, infos, fieldSep)
	}
	return true
}

func EnableRepoInHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	fieldSep := env.GetRaw("strs.proto-sep")

	metaPath := getReposInfoPath(env, "EnableRepoInHub")
	infos, _ := readReposInfoFile(metaPath, true, fieldSep)
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
		cc.Screen.Print(fmt.Sprintf("[%s] (enabled)\n", addrDisplayName(info.Addr)))
		printInfoProps(cc.Screen, info)
		info.OnOff = "on"
		extracted[i] = info
	}

	writeReposInfoFile(metaPath, append(rest, extracted...), fieldSep)
	return true
}

func DisableRepoInHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	fieldSep := env.GetRaw("strs.proto-sep")

	metaPath := getReposInfoPath(env, "DisableRepoInHub")
	infos, _ := readReposInfoFile(metaPath, true, fieldSep)
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
			cc.Screen.Print(fmt.Sprintf("[%s] (disabled)\n", addrDisplayName(info.Addr)))
			cc.Screen.Print(fmt.Sprintf("    %s\n", info.Path))
			info.OnOff = "disabled"
			extracted[i] = info
		}
	}

	writeReposInfoFile(metaPath, append(rest, extracted...), fieldSep)
	return true
}

func MoveSavedFlowsToLocalDir(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	cc.Screen.Print("TODO: MoveFlowToDir\n")
	return true
}

func AddLocalDirToHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	cc.Screen.Print("TODO: AddLocalDir\n")
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
	oldInfos, oldList := readReposInfoFile(metaPath, true, fieldSep)
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

	var infos []repoInfo
	for i, addr := range addrs {
		if oldList[addr] {
			continue
		}
		repoPath := getRepoPath(path, addr)
		infos = append(infos, repoInfo{addr, gitAddr, repoPath, helpStrs[i], "on"})
	}

	infos = append(oldInfos, infos...)
	writeReposInfoFile(metaPath, infos, fieldSep)
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
	if !os.IsNotExist(err) {
		if !stat.IsDir() {
			panic(fmt.Errorf("[updateRepoAndReadSubList] repo path '%v' exists but is not dir",
				repoPath))
		}
		screen.Print(fmt.Sprintf("[%s] => git update\n", name))
		cmdStrs = []string{"git", "-C", repoPath, "pull"}
	} else {
		screen.Print(fmt.Sprintf("[%s] => git clone\n", name))
		cmdStrs = []string{"git", "clone", gitAddr, repoPath}
	}

	cmd := exec.Command(cmdStrs[0], cmdStrs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(fmt.Errorf("[updateRepoAndReadSubList] run '%v' failed: %v", cmdStrs, err))
	}
	listFilePath := filepath.Join(repoPath, listFileName)
	return readRepoListFromFile(listFilePath)
}

func readRepoListFromFile(path string) (helpStr string, addrs []string, helpStrs []string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("[readRepoListFromFile] read list file '%v' failed: %v",
			path, err))
	}
	list := strings.Split(string(data), "\n")
	meetMark := false

	// TODO: move to specific package
	const StartMark = "[ticat.hub]"
	for i, line := range list {
		line = strings.TrimSpace(line)
		if i != 0 && len(line) > 0 && len(helpStr) == 0 {
			j := strings.LastIndex(line, ":")
			if j < 0 {
				helpStr = line
			} else {
				text := strings.TrimSpace(line[j+1:])
				if len(text) > 0 {
					helpStr = text
				}
			}
		}
		if strings.HasPrefix(line, StartMark) {
			meetMark = true
		}
		if !meetMark {
			continue
		}
		if len(line) > 0 && line[0:1] == "*" {
			line = strings.TrimSpace(line[1:])
			i := strings.Index(line, "[")
			if i < 0 {
				continue
			}
			line = line[i+1:]
			j := strings.Index(line, "]")
			if j < 0 {
				continue
			}
			addrs = append(addrs, strings.TrimSpace(line[:j]))
			line := line[j+1:]
			k := strings.LastIndex(line, ":")
			if k < 0 {
				continue
			}
			helpStrs = append(helpStrs, strings.TrimSpace(line[k+1:]))
		}
	}
	return
}

func extractAddrFromList(
	infos []repoInfo,
	findStr string) (extracted []repoInfo, rest []repoInfo) {

	for _, info := range infos {
		if strings.Index(info.Addr, findStr) >= 0 {
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

type repoInfo struct {
	Addr      string
	AddReason string
	Path      string
	HelpStr   string
	OnOff     string
}

func writeReposInfoFile(path string, infos []repoInfo, sep string) {
	tmp := path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(fmt.Errorf("[writeReposInfoFile] open file '%s' failed: %v", tmp, err))
	}
	defer file.Close()

	for _, info := range infos {
		_, err = fmt.Fprintf(file, "%s%s%s%s%s%s%s%s%s\n", info.Addr, sep,
			info.AddReason, sep, info.Path, sep, info.HelpStr, sep, info.OnOff)
		if err != nil {
			panic(fmt.Errorf("[writeReposInfoFile] write file '%s' failed: %v", tmp, err))
		}
	}
	file.Close()

	err = os.Rename(tmp, path)
	if err != nil {
		panic(fmt.Errorf("[writeReposInfoFile] rename file '%s' to '%s' failed: %v",
			tmp, path, err))
	}
}

func readReposInfoFile(
	path string,
	allowNotExist bool,
	sep string) (infos []repoInfo, list map[string]bool) {

	list = map[string]bool{}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) && allowNotExist {
			return
		}
		panic(fmt.Errorf("[readReposInfoFile] open file '%s' failed: %v", path, err))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), "\n\r")
		fields := strings.Split(line, sep)
		if len(fields) != 5 {
			panic(fmt.Errorf("[readReposInfoFile] file '%s' line '%s' can't be parsed",
				path, line))
		}
		info := repoInfo{
			fields[0],
			fields[1],
			fields[2],
			fields[3],
			fields[4],
		}
		infos = append(infos, info)
		list[info.Addr] = true
	}
	return
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

func printInfoProps(screen core.Screen, info repoInfo) {
	screen.Print(fmt.Sprintf("     '%s'\n", info.HelpStr))
	screen.Print(fmt.Sprintf("    - from: %s\n", getDisplayReason(info)))
	screen.Print(fmt.Sprintf("    - path: %s\n", info.Path))
}

func getDisplayReason(info repoInfo) string {
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
