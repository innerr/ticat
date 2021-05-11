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
	infos := readReposInfoFile(metaPath, true, fieldSep)
	for _, info := range infos {
		loadLocalMods(cc, info.Path, metaExt, flowExt, abbrsSep, envPathSep)
	}
	return true
}

func AddGitAddrToHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	addr := argv.GetRaw("git-address")
	if len(addr) == 0 {
		panic(fmt.Errorf("[AddGitAddrToHub] cant't get hub address"))
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
	repoExt := env.GetRaw("strs.mods-repo-ext")
	fieldSep := env.GetRaw("strs.proto-sep")
	metaPath := getReposInfoPath(env, "ListHub")
	infos := readReposInfoFile(metaPath, true, fieldSep)

	for _, info := range infos {
		from := info.AddReason
		if from == info.Addr {
			from = "<manually added>"
		}
		abbr := gitAddrAbbr(info.Addr, repoExt)
		if len(abbr) != 0 {
			cc.Screen.Print(fmt.Sprintf("[%s]\n", abbr))
		} else {
			cc.Screen.Print(fmt.Sprintf("[%s]\n", info.Addr))
		}
		cc.Screen.Print(fmt.Sprintf("     '%s'\n", info.HelpStr))
		if len(abbr) != 0 {
			cc.Screen.Print(fmt.Sprintf("    - full: %s\n", info.Addr))
		}
		cc.Screen.Print(fmt.Sprintf("    - from: %s\n", from))
		cc.Screen.Print(fmt.Sprintf("    - path: %s\n", info.Path))
	}
	return true
}

func RemoveAllFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	fieldSep := env.GetRaw("strs.proto-sep")
	repoExt := env.GetRaw("strs.mods-repo-ext")

	metaPath := getReposInfoPath(env, "RemoveAllFromHub")
	infos := readReposInfoFile(metaPath, true, fieldSep)

	for _, info := range infos {
		osRemoveDir(info.Path)
		cc.Screen.Print(fmt.Sprintf("[%s] (removed)\n", gitAddrAbbr(info.Addr, repoExt)))
		cc.Screen.Print(fmt.Sprintf("    %s\n", info.Path))
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

func RemoveRepoFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	fieldSep := env.GetRaw("strs.proto-sep")
	repoExt := env.GetRaw("strs.mods-repo-ext")

	metaPath := getReposInfoPath(env, "RemoveGitAddrFromHub")
	infos := readReposInfoFile(metaPath, true, fieldSep)
	addr := argv.GetRaw("git-address")
	if len(addr) == 0 {
		panic(fmt.Errorf("[RemoveRepoFromHub] cant't get target repo addr from args"))
	}

	gitAddr := normalizeGitAddr(addr, repoExt)
	info, rest := extractAddrFromList(infos, gitAddr)
	if len(info.Path) == 0 {
		panic(fmt.Errorf("[RemoveRepoFromHub] cant't find repo '%s', normalized: %s",
			addr, gitAddr))
	}
	osRemoveDir(info.Path)

	cc.Screen.Print(fmt.Sprintf("[%s] (removed)\n", gitAddrAbbr(info.Addr, repoExt)))
	cc.Screen.Print(fmt.Sprintf("    %s\n", info.Path))

	writeReposInfoFile(metaPath, rest, fieldSep)
	return true
}

func UpdateHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	cc.Screen.Print("TODO: UpdateHub\n")
	return true
}

func EnableAddrInHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	cc.Screen.Print("TODO: EnableAddrInHub\n")
	return true
}

func DisableAddrInHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	cc.Screen.Print("TODO: DisableAddrInHub\n")
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

	repoExt := env.GetRaw("strs.mods-repo-ext")
	gitAddr = normalizeGitAddr(gitAddr, repoExt)

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

	listFileName := env.GetRaw("strs.repos-file-name")
	var topRepoHelpStr string
	topRepoHelpStr, addrs, helpStrs = readRepoListFromGitRepos(
		screen, path, gitAddr, repoExt, listFileName)

	addrs = append([]string{gitAddr}, addrs...)
	helpStrs = append([]string{topRepoHelpStr}, helpStrs...)

	var infos []repoInfo
	for i, addr := range addrs {
		repoPath := getRepoPath(path, addr)
		infos = append(infos, repoInfo{addr, gitAddr, repoPath, helpStrs[i]})
	}
	metaPath := getReposInfoPath(env, "addRepoToHub")
	fieldSep := env.GetRaw("strs.proto-sep")
	writeReposInfoFile(metaPath, infos, fieldSep)
	return
}

func readRepoListFromGitRepos(
	screen core.Screen,
	hubPath string,
	gitAddr string,
	repoExt string,
	listFileName string) (topRepoHelpStr string, addrs []string, helpStrs []string) {

	screen.Print(fmt.Sprintf("[%s] => git clone\n", gitAddrAbbr(gitAddr, repoExt)))
	topRepoHelpStr, addrs, helpStrs = readRepoListFromGitRepo(
		hubPath, gitAddr, listFileName)

	for _, addr := range addrs {
		_, subAddrs, subHelpStrs := readRepoListFromGitRepos(
			screen, hubPath, addr, repoExt, listFileName)
		addrs = append(addrs, subAddrs...)
		helpStrs = append(helpStrs, subHelpStrs...)
	}

	return topRepoHelpStr, addrs, helpStrs
}

func readRepoListFromGitRepo(
	hubPath string,
	gitAddr string,
	listFileName string) (helpStr string, addrs []string, helpStrs []string) {

	repoPath := getRepoPath(hubPath, gitAddr)
	cmdStrs := []string{"git", "clone", gitAddr, repoPath}
	cmd := exec.Command(cmdStrs[0], cmdStrs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(fmt.Errorf("[readRepoListFromGitRepo] run '%v' failed: %v", cmdStrs, err))
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

func extractAddrFromList(infos []repoInfo, target string) (extracted repoInfo, rest []repoInfo) {
	for _, info := range infos {
		if info.Addr == target {
			extracted = info
		} else {
			rest = append(rest, info)
		}
	}
	return
}

func normalizeGitAddr(addr string, repoExt string) string {
	if !strings.HasSuffix(strings.ToLower(addr), repoExt) {
		addr = addr + repoExt
	}
	if strings.HasPrefix(strings.ToLower(addr), "http") {
		return addr
	}
	if strings.HasPrefix(strings.ToLower(addr), "git") {
		return addr
	}
	return "git@github.com:" + addr
}

func gitAddrAbbr(addr string, repoExt string) (abbr string) {
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
	abbr = strings.TrimRight(abbr, repoExt)
	return
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
}

func writeReposInfoFile(path string, infos []repoInfo, sep string) {
	tmp := path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(fmt.Errorf("[writeReposInfoFile] open file '%s' failed: %v", tmp, err))
	}
	defer file.Close()

	for _, info := range infos {
		_, err = fmt.Fprintf(file, "%s%s%s%s%s%s%s\n",
			info.Addr, sep, info.AddReason, sep, info.Path, sep, info.HelpStr)
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

func readReposInfoFile(path string, allowNotExist bool, sep string) (infos []repoInfo) {
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
		if len(fields) != 4 {
			panic(fmt.Errorf("[readReposInfoFile] file '%s' line '%s' can't be parsed", path, line))
		}
		infos = append(infos, repoInfo{
			fields[0],
			fields[1],
			fields[2],
			fields[3],
		})
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
