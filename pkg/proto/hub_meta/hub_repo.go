package hub_meta

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func UpdateRepoAndSubRepos(
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
		subTopHelpStr, subAddrs, subHelpStrs := UpdateRepoAndSubRepos(
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
	return filepath.Join(hubPath, filepath.Base(gitAddr))
}

func updateRepoAndReadSubList(
	screen core.Screen,
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
