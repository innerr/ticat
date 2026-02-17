package hub_meta

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type RepoAddr struct {
	Addr   string
	Branch string
}

func (self RepoAddr) Str() string {
	if len(self.Addr) == 0 {
		return self.Addr
	}
	if len(self.Branch) == 0 {
		return self.Addr
	}
	return self.Addr + RepoAddrBranchSep + self.Branch
}

func ParseRepoAddr(addr string) RepoAddr {
	branch := ""
	i := strings.LastIndex(addr, RepoAddrBranchSep)
	if i > 0 {
		branch = addr[i+1:]
		addr = addr[:i]
	}
	return RepoAddr{addr, branch}
}

type RepoInfo struct {
	Addr      RepoAddr
	AddReason string
	Path      string
	HelpStr   string
	OnOff     string
}

func (self RepoInfo) IsLocal() bool {
	return len(self.Addr.Addr) == 0
}

func WriteReposInfoFile(
	hubRootPath string,
	path string,
	infos []RepoInfo, sep string) {

	os.MkdirAll(filepath.Dir(path), os.ModePerm)
	tmp := path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(fmt.Errorf("[WriteReposInfoFile] open file '%s' failed: %v", tmp, err))
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(fmt.Errorf("[WriteReposInfoFile] close file '%s' failed: %v", tmp, err))
		}
	}()

	for _, info := range infos {
		_, err = fmt.Fprintf(file, "%s%s%s%s%s%s%s%s%s\n", info.Addr.Str(), sep, info.AddReason, sep,
			tryConvAbsPathToRelPath(hubRootPath, info.Path), sep, info.HelpStr, sep, info.OnOff)
		if err != nil {
			panic(fmt.Errorf("[WriteReposInfoFile] write file '%s' failed: %v", tmp, err))
		}
	}
	if err := file.Close(); err != nil {
		panic(fmt.Errorf("[WriteReposInfoFile] close file '%s' failed: %v", tmp, err))
	}

	err = os.Rename(tmp, path)
	if err != nil {
		panic(fmt.Errorf("[WriteReposInfoFile] rename file '%s' to '%s' failed: %v",
			tmp, path, err))
	}
}

func ReadReposInfoFile(
	hubRootPath string,
	path string,
	allowNotExist bool,
	sep string) (infos []RepoInfo, list map[string]bool) {

	list = map[string]bool{}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) && allowNotExist {
			return
		}
		panic(fmt.Errorf("[ReadReposInfoFile] open file '%s' failed: %v", path, err))
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(fmt.Errorf("[ReadReposInfoFile] close file '%s' failed: %v", path, err))
		}
	}()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), "\n\r")
		fields := strings.Split(line, sep)
		if len(fields) != 5 {
			panic(fmt.Errorf("[ReadReposInfoFile] file '%s' line '%s' can't be parsed",
				path, line))
		}
		info := RepoInfo{
			ParseRepoAddr(fields[0]),
			fields[1],
			tryConvRelPathToAbsPath(hubRootPath, fields[2]),
			fields[3],
			fields[4],
		}
		infos = append(infos, info)
		list[info.Addr.Str()] = true
	}
	return
}

func tryConvAbsPathToRelPath(hubRootPath string, path string) string {
	if strings.HasPrefix(path, hubRootPath) {
		return HubRootPathMark + path[len(hubRootPath):]
	}
	return path
}

func tryConvRelPathToAbsPath(hubRootPath string, path string) string {
	if strings.HasPrefix(path, HubRootPathMark) {
		return hubRootPath + path[len(HubRootPathMark):]
	}
	return path
}

// TODO: bad func name
func ExtractAddrFromList(
	infos []RepoInfo,
	findStr string) (extracted []RepoInfo, rest []RepoInfo) {

	if len(findStr) == 0 {
		return infos, nil
	}

	for i, info := range infos {
		findInStr := info.Addr.Str()
		if len(findInStr) == 0 {
			findInStr = info.Path
		}
		if findInStr == findStr {
			return []RepoInfo{info}, append(infos[:i], infos[i+1:]...)
		}
		if strings.Index(findInStr, findStr) >= 0 {
			extracted = append(extracted, info)
		} else {
			rest = append(rest, info)
		}
	}
	return
}

const (
	RepoAddrBranchSep = "#"
	HubRootPathMark   = "<hub>:"
)
