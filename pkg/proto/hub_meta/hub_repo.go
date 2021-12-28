package hub_meta

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func UpdateRepoAndSubRepos(
	screen core.Screen,
	env *core.Env,
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
		screen, env, hubPath, gitAddr, listFileName, selfName, cmd)
	finisheds[gitAddr] = true

	for i, addr := range addrs {
		subTopHelpStr, subAddrs, subHelpStrs := UpdateRepoAndSubRepos(
			screen, env, finisheds, hubPath, addr, repoExt, listFileName, selfName, cmd)
		// If a repo has no help-str from hub-repo list, try to get the title from it's README
		if len(helpStrs[i]) == 0 && len(subTopHelpStr) != 0 {
			helpStrs[i] = subTopHelpStr
		}
		addrs = append(addrs, subAddrs...)
		helpStrs = append(helpStrs, subHelpStrs...)
	}

	return topRepoHelpStr, addrs, helpStrs
}

func NormalizeGitAddr(addr string) string {
	if strings.HasPrefix(strings.ToLower(addr), "http") {
		return addr
	}
	if strings.HasPrefix(strings.ToLower(addr), "git") {
		return addr
	}
	//return "git@github.com:" + addr
	return "https://github.com/" + addr
}

func AddrDisplayName(addr string) string {
	abbr := gitAddrAbbr(addr)
	if len(abbr) == 0 {
		return addr
	}
	return abbr
}

func GetRepoPath(hubPath string, gitAddr string) string {
	addr := strings.ToLower(gitAddr)
	for _, prefix := range []string{"http://", "https://", "git@", "root@"} {
		addr = strings.TrimPrefix(addr, prefix)
	}
	author := filepath.Dir(addr)
	return filepath.Join(hubPath,
		filepath.Dir(author),
		filepath.Base(author),
		filepath.Base(addr))
}

func CheckRepoGitStatus(
	screen core.Screen,
	env *core.Env,
	hubPath string,
	gitAddr string) {

	name := AddrDisplayName(gitAddr)
	repoPath := GetRepoPath(hubPath, gitAddr)
	var cmdStrs []string

	stat, err := os.Stat(repoPath)
	if os.IsNotExist(err) {
		screen.Print(fmt.Sprintf(display.ColorHub("[%s]\n", env)+display.ColorError("=> ", env)+
			"repo dir not exists: %s\n", name, repoPath))
		return
	}
	if !stat.IsDir() {
		screen.Print(fmt.Sprintf(display.ColorHub("[%s]\n", env)+display.ColorError("=> ", env)+
			"repo path exists but is not dir: %s\n", name, repoPath))
		return
	}
	screen.Print(fmt.Sprintf(display.ColorHub("[%s]\n", env)+display.ColorSymbol("=> ", env)+"git status\n"+
		display.ColorExplain("(%s)", env)+"\n", name, repoPath))
	cmdStrs = []string{"git", "status"}

	// Ignore errors
	c := exec.Command(cmdStrs[0], cmdStrs[1:]...)
	c.Dir = repoPath
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Run()
}

func updateRepoAndReadSubList(
	screen core.Screen,
	env *core.Env,
	hubPath string,
	gitAddr string,
	listFileName string,
	selfName string,
	cmd core.ParsedCmd) (helpStr string, addrs []string, helpStrs []string) {

	name := AddrDisplayName(gitAddr)
	repoPath := GetRepoPath(hubPath, gitAddr)
	var cmdStrs []string

	stat, err := os.Stat(repoPath)
	var pwd string
	if !os.IsNotExist(err) {
		if !stat.IsDir() {
			panic(core.WrapCmdError(cmd, fmt.Errorf("repo path '%v' exists but is not dir",
				repoPath)))
		}
		screen.Print(fmt.Sprintf(display.ColorHub("[%s]", env)+display.ColorSymbol(" => ", env)+
			"git update\n", name))
		cmdStrs = []string{"git", "pull", "--recurse-submodules"}
		pwd = repoPath
	} else {
		screen.Print(fmt.Sprintf(display.ColorHub("[%s]", env)+display.ColorSymbol(" => ", env)+
			"git clone\n", name))
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
	return ReadRepoListFromFile(selfName, listFilePath)
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
