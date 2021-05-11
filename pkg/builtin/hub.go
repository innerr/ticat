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
	path := env.GetRaw("sys.paths.hub")
	if len(path) == 0 {
		panic(fmt.Errorf("[addRepoToHub] cant't get hub path"))
	}
	// TODO: move to env
	const ReposInfoFileName = "repos.hub"
	metaExt := "." + env.GetRaw("strs.meta-ext")
	abbrsSep := env.GetRaw("strs.abbrs-sep")
	envPathSep := env.GetRaw("strs.env-path-sep")
	infos := readReposInfoFile(filepath.Join(path, ReposInfoFileName), true)
	for _, info := range infos {
		loadLocalMods(cc, info.Path, metaExt, abbrsSep, envPathSep)
	}
	return true
}

func AddGitAddrToHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	addr := argv.GetRaw("git-address")
	if len(addr) == 0 {
		panic(fmt.Errorf("[AddGitAddrToHub] cant't get hub address"))
	}
	addRepoToHub(addr, argv, env)
	return true
}

func AddGitDefaultToHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	addr := env.GetRaw("sys.hub.init-repo")
	if len(addr) == 0 {
		panic(fmt.Errorf("[AddGitDefaultToHub] cant't get init-repo address from env"))
	}
	addRepoToHub(addr, argv, env)
	return true
}

func ListHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	path := env.GetRaw("sys.paths.hub")
	if len(path) == 0 {
		panic(fmt.Errorf("[addRepoToHub] cant't get hub path"))
	}
	// TODO: move to env
	const ReposInfoFileName = "repos.hub"
	infos := readReposInfoFile(filepath.Join(path, ReposInfoFileName), true)
	for _, info := range infos {
		from := info.AddReason
		if from == info.Addr {
			from = "<manually added>"
		}
		abbr := gitAddrAbbr(info.Addr)
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

func RemoveGitAddrFromHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	path := env.GetRaw("sys.paths.hub")
	if len(path) == 0 {
		panic(fmt.Errorf("[RemoveGitAddrFromHub] cant't get hub path"))
	}
	// TODO: move to env
	const ReposInfoFileName = "repos.hub"
	infoPath := filepath.Join(path, ReposInfoFileName)
	infos := readReposInfoFile(infoPath, true)
	addr := argv.GetRaw("git-address")
	if len(addr) == 0 {
		panic(fmt.Errorf("[RemoveGitAddrFromHub] cant't get target repo address from args"))
	}
	related, rest := getRelatedAddrList(infos, addr)
	writeReposInfoFile(infoPath, rest)
	cc.Screen.Print(fmt.Sprintf("[RemoveGitAddrFromHub] TODO: remove dir '%v'\n", related))
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
	addr string,
	argv core.ArgVals,
	env *core.Env) (addrs []string, helpStrs []string) {

	addr = normalizeGitAddr(addr)

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

	addrs, helpStrs = readRepoListFromGitRepos(path, addr)
	var infos []RepoInfo
	for i, addr := range addrs {
		repoPath := getRepoPath(path, addr)
		infos = append(infos, RepoInfo{addr, addr, repoPath, helpStrs[i]})
	}
	// TODO: move to env
	// TODO: write to <data>/repos.hub?
	const ReposInfoFileName = "repos.hub"
	writeReposInfoFile(filepath.Join(path, ReposInfoFileName), infos)
	return
}

func readRepoListFromGitRepos(hubPath string, gitAddr string) (addrs []string, helpStrs []string) {
	addrs, helpStrs = readRepoListFromGitRepo(hubPath, gitAddr)
	for _, addr := range addrs {
		subAddrs, subHelpStrs := readRepoListFromGitRepo(hubPath, addr)
		addrs = append(addrs, subAddrs...)
		helpStrs = append(helpStrs, subHelpStrs...)
	}
	return
}

func readRepoListFromGitRepo(hubPath string, gitAddr string) (addrs []string, helpStrs []string) {
	repoPath := getRepoPath(hubPath, gitAddr)
	cmdStrs := []string{"git", "clone", gitAddr, repoPath}
	cmd := exec.Command(cmdStrs[0], cmdStrs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(fmt.Errorf("[readRepoListFromGitRepo] run '%v' failed: %v", cmdStrs, err))
	}
	// TODO: move to env
	const listFileName = "README.md"
	listFilePath := filepath.Join(repoPath, listFileName)
	return readRepoListFromFile(listFilePath)
}

func getRepoPath(hubPath string, gitAddr string) string {
	return filepath.Join(hubPath, filepath.Base(gitAddr))
}

func readRepoListFromFile(path string) (addrs []string, helpStrs []string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("[readRepoListFromFile] read list file '%v' failed: %v",
			path, err))
	}
	list := strings.Split(string(data), "\n")
	meetMark := false
	// TODO: move to env
	const StartMark = "[ticat.hub]"
	for _, line := range list {
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

func getRelatedAddrList(infos []RepoInfo, target string) (related []RepoInfo, rest []RepoInfo) {
	/* TODO
	for _, info := range infos {
		if info.AddedReason == target {
			related = append(related, info)
		} else {
			rest = append(rest, info)
		}
	}
	for _, rel := range related {
		subRelated, subRest := getRelatedAddrList(rest, rel)
	}
	*/
	for _, info := range infos {
		if info.Addr == target {
			related = append(related, info)
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
	// TODO: config from env
	abbr = strings.TrimRight(abbr, ".ticat")
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

type RepoInfo struct {
	Addr      string
	AddReason string
	Path      string
	HelpStr   string
}

func writeReposInfoFile(path string, infos []RepoInfo) {
	tmp := path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(fmt.Errorf("[writeReposInfoFile] open file '%s' failed: %v", tmp, err))
	}
	defer file.Close()

	// TODO: move to env
	const sep = "\t"
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

func readReposInfoFile(path string, allowNotExist bool) (infos []RepoInfo) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) && allowNotExist {
			return
		}
		panic(fmt.Errorf("[readReposInfoFile] open file '%s' failed: %v", path, err))
	}
	defer file.Close()

	// TODO: move to env
	const sep = "\t"

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), "\n\r")
		fields := strings.Split(line, sep)
		if len(fields) != 4 {
			panic(fmt.Errorf("[readReposInfoFile] file '%s' line '%s' can't be parsed", path, line))
		}
		infos = append(infos, RepoInfo{
			fields[0],
			fields[1],
			fields[2],
			fields[3],
		})
	}
	return
}
