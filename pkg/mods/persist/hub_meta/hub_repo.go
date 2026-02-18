package hub_meta

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/innerr/ticat/pkg/cli/display"
	"github.com/innerr/ticat/pkg/core/model"
)

func UpdateRepoAndSubRepos(
	screen model.Screen,
	env *model.Env,
	finisheds map[string]bool,
	hubPath string,
	gitAddr RepoAddr,
	repoExt string,
	listFileName string,
	selfName string,
	cmd model.ParsedCmd) (topRepoHelpStr string, addrs []RepoAddr, helpStrs []string, err error) {

	if finisheds[gitAddr.Str()] {
		return
	}
	topRepoHelpStr, addrs, helpStrs, err = updateRepoAndReadSubList(
		screen, env, hubPath, gitAddr, listFileName, selfName, cmd)
	if err != nil {
		return
	}
	finisheds[gitAddr.Str()] = true

	for i, addr := range addrs {
		var subTopHelpStr string
		var subAddrs []RepoAddr
		var subHelpStrs []string
		subTopHelpStr, subAddrs, subHelpStrs, err = UpdateRepoAndSubRepos(
			screen, env, finisheds, hubPath, addr, repoExt, listFileName, selfName, cmd)
		if err != nil {
			return
		}
		// If a repo has no help-str from hub-repo list, try to get the title from it's README
		if len(helpStrs[i]) == 0 && len(subTopHelpStr) != 0 {
			helpStrs[i] = subTopHelpStr
		}
		addrs = append(addrs, subAddrs...)
		helpStrs = append(helpStrs, subHelpStrs...)
	}

	return topRepoHelpStr, addrs, helpStrs, nil
}

func NormalizeGitAddr(addr string) string {
	if strings.HasPrefix(strings.ToLower(addr), "http") {
		return addr
	}
	if isSshAddr(addr) {
		return addr
	}
	return "https://github.com/" + addr
}

// TODO: need to improve this ssh-format address checking
func isSshAddr(addr string) bool {
	symAt := strings.Index(strings.ToLower(addr), "@")
	symCl := strings.Index(strings.ToLower(addr), ":")
	return symAt > 0 && symCl > symAt
}

func AddrDisplayName(addr RepoAddr) string {
	abbr := gitAddrAbbr(addr.Addr)
	if len(abbr) == 0 {
		return addr.Str()
	}
	return RepoAddr{abbr, addr.Branch}.Str()
}

func GetRepoPath(hubPath string, originGitAddr RepoAddr) (string, error) {
	addr := strings.ToLower(originGitAddr.Addr)
	for _, prefix := range []string{"http://", "https://"} {
		addr = strings.TrimPrefix(addr, prefix)
	}

	symAt := strings.Index(strings.ToLower(addr), "@")
	if symAt >= 0 {
		addr = addr[symAt+1:]
		symCl := strings.LastIndex(strings.ToLower(addr), ":")
		if symCl <= 0 {
			return "", fmt.Errorf("ill-format repo address '%v'", originGitAddr.Addr)
		}
		addr = addr[0:symCl] + "/" + addr[symCl+1:]
	}

	if len(originGitAddr.Branch) != 0 {
		addr = filepath.Join(addr, originGitAddr.Branch+".branch")
	} else {
		addr = filepath.Join(addr, "default.branch")
	}

	author := filepath.Dir(addr)
	return filepath.Join(hubPath,
		filepath.Dir(author),
		filepath.Base(author),
		filepath.Base(addr)), nil
}

func CheckRepoGitStatus(
	screen model.Screen,
	env *model.Env,
	hubPath string,
	gitAddr RepoAddr) error {

	name := AddrDisplayName(gitAddr)
	repoPath, err := GetRepoPath(hubPath, gitAddr)
	if err != nil {
		return err
	}
	var cmdStrs []string

	stat, err := os.Stat(repoPath)
	if os.IsNotExist(err) {
		screen.Print(fmt.Sprintf(display.ColorHub("[%s]\n", env)+display.ColorError("=> ", env)+
			"repo dir not exists: %s\n", name, repoPath))
		return nil
	}
	if !stat.IsDir() {
		screen.Print(fmt.Sprintf(display.ColorHub("[%s]\n", env)+display.ColorError("=> ", env)+
			"repo path exists but is not dir: %s\n", name, repoPath))
		return nil
	}
	screen.Print(fmt.Sprintf(display.ColorHub("[%s]\n", env)+display.ColorSymbol("=> ", env)+"git status\n"+
		display.ColorExplain("(%s)", env)+"\n", name, repoPath))
	cmdStrs = []string{"git", "status"}

	// Ignore errors
	c := exec.Command(cmdStrs[0], cmdStrs[1:]...)
	c.Dir = repoPath
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		// Ignore errors intentionally
		_ = err
	}
	return nil
}

func updateRepoAndReadSubList(
	screen model.Screen,
	env *model.Env,
	hubPath string,
	gitAddr RepoAddr,
	listFileName string,
	selfName string,
	cmd model.ParsedCmd) (helpStr string, addrs []RepoAddr, helpStrs []string, err error) {

	name := AddrDisplayName(gitAddr)
	repoPath, err := GetRepoPath(hubPath, gitAddr)
	if err != nil {
		return
	}
	var cmdStrs []string

	stat, statErr := os.Stat(repoPath)
	var pwd string
	if !os.IsNotExist(statErr) {
		if !stat.IsDir() {
			err = model.WrapCmdError(cmd, fmt.Errorf("repo path '%v' exists but is not dir",
				repoPath))
			return
		}
		screen.Print(fmt.Sprintf(display.ColorHub("[%s]", env)+display.ColorSymbol(" => ", env)+
			"git update\n", name))
		cmdStrs = []string{"git", "pull", "--recurse-submodules"}
		pwd = repoPath
	} else {
		screen.Print(fmt.Sprintf(display.ColorHub("[%s]", env)+display.ColorSymbol(" => ", env)+
			"git clone\n", name))
		cmdStrs = []string{"git", "clone", "--recursive", gitAddr.Addr}
		if len(gitAddr.Branch) != 0 {
			cmdStrs = append(cmdStrs, "-b", gitAddr.Branch)
		}
		cmdStrs = append(cmdStrs, repoPath)
	}

	c := exec.Command(cmdStrs[0], cmdStrs[1:]...)
	if len(pwd) != 0 {
		c.Dir = pwd
	}
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	runErr := c.Run()
	if runErr != nil {
		err = model.WrapCmdError(cmd, fmt.Errorf("run '%v' failed: %v", cmdStrs, runErr))
		return
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
