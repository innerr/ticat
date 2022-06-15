package hub_meta

import (
	"fmt"
	"os"

	"github.com/pingcap/ticat/pkg/proto/meta_file"
)

func ReadRepoListFromFile(selfName string, path string) (helpStr string, addrs []string, helpStrs []string) {
	metas, err := meta_file.NewMetaFileEx(path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		panic(fmt.Errorf("[ReadRepoListFromFile] read mod meta file '%s' failed: %v", path, err))
	}
	if len(metas) != 1 {
		panic(fmt.Errorf("[ReadRepoListFromFile] repo meta file '%s' should not be a combined file", path))
	}
	meta := metas[0].Meta

	helpStr = meta.Get("help")
	repos := meta.GetSection("repos")
	if repos == nil {
		repos = meta.GetSection("repo")
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
