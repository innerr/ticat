package hub_meta

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type RepoInfo struct {
	Addr      string
	AddReason string
	Path      string
	HelpStr   string
	OnOff     string
}

func WriteReposInfoFile(path string, infos []RepoInfo, sep string) {
	tmp := path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(fmt.Errorf("[WriteReposInfoFile] open file '%s' failed: %v", tmp, err))
	}
	defer file.Close()

	for _, info := range infos {
		_, err = fmt.Fprintf(file, "%s%s%s%s%s%s%s%s%s\n", info.Addr, sep,
			info.AddReason, sep, info.Path, sep, info.HelpStr, sep, info.OnOff)
		if err != nil {
			panic(fmt.Errorf("[WriteReposInfoFile] write file '%s' failed: %v", tmp, err))
		}
	}
	file.Close()

	err = os.Rename(tmp, path)
	if err != nil {
		panic(fmt.Errorf("[WriteReposInfoFile] rename file '%s' to '%s' failed: %v",
			tmp, path, err))
	}
}

func ReadReposInfoFile(
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
	defer file.Close()

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
