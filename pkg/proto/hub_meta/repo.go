package hub_meta

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pingcap/ticat/pkg/proto/meta_file"
)

func ReadRepoListFromFile(selfName string, path string) (helpStr string, addrs []string, helpStrs []string) {
	meta, err := meta_file.NewModMetaEx(path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		panic(fmt.Errorf("[ReadRepoListFromFile] read mod meta file '%s' failed: %v", path, err))
	}
	helpStr = meta.Get("help")
	repos := meta.GetSession("repos")
	if repos == nil {
		repos = meta.GetSession("repo")
	}
	if repos == nil {
		return
	}
	for _, addr := range repos.Keys() {
		addrs = append(addrs, addr)
		helpStrs = append(helpStrs, repos.Get(addr))
	}
	return
}

func ReadRepoListFromFileOld(selfName string, path string) (helpStr string, addrs []string, helpStrs []string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		panic(fmt.Errorf("[readRepoListFromFile] read list file '%v' failed: %v",
			path, err))
	}
	list := strings.Split(string(data), "\n")
	meetMark := false

	StartMark := "[" + selfName + ".hub]"
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
			addr := strings.TrimSpace(line[:j])
			if len(addr) == 0 {
				continue
			}
			addrs = append(addrs, addr)
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
