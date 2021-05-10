package builtin

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func AddToLocalHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	cc.Screen.Print("TODO: AddToLocalHub\n")
	return true
}

func AddDefaultToLocalHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	cc.Screen.Print("TODO: AddDefaultToLocalHub\n")
	return true
}

func ListLocalHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	cc.Screen.Print("TODO: ListLocalHub\n")
	return true
}

func UpdateLocalHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	cc.Screen.Print("TODO: UpdateLocalHub\n")
	return true
}

func EnableAddrInLocalHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	cc.Screen.Print("TODO: EnableAddrInLocalHub\n")
	return true
}

func DisableAddrInLocalHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	cc.Screen.Print("TODO: DisableAddrInLocalHub\n")
	return true
}

func MoveSavedFlowsToLocalDir(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	cc.Screen.Print("TODO: MoveFlowToDir\n")
	return true
}

func AddLocalDirToLocalHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	cc.Screen.Print("TODO: AddLocalDir\n")
	return true
}

// TODO: call on bootstrap
func LoadFromLocalHub(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	cc.Screen.Print("TODO: LoadFromLocalHub\n")
	return true
}

// TODO:
func Add(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	if !isOsCmdExists("git") {
		panic(fmt.Errorf("[LoadFromLocalHub] cant't find 'git'"))
	}

	path := env.Get("sys.paths.hub").Raw
	if len(path) == 0 {
		panic(fmt.Errorf("[LoadFromLocalHub] cant't get local hub path"))
	}
	err := os.MkdirAll(path, os.ModePerm)
	if os.IsExist(err) {
		panic(fmt.Errorf("[LoadFromLocalHub] create hub path '%s' failed: %v", path, err))
	}

	addr := env.Get("sys.hub.address").Raw
	if len(addr) == 0 {
		panic(fmt.Errorf("[LoadFromLocalHub] cant't get hub address"))
	}

	addrs, helpStrs := readRepoListFromHubRepo(path, addr)
	fmt.Printf("%v\n", addrs)
	fmt.Printf("%v\n", helpStrs)
	return true
}

func readRepoListFromHubRepo(hubPath string, gitAddr string) (addrs []string, helpStrs []string) {
	dir := filepath.Base(gitAddr)
	repoPath := filepath.Join(hubPath, dir)
	cmdStrs := []string{"git", "clone", gitAddr, repoPath}
	cmd := exec.Command(cmdStrs[0], cmdStrs[1:]...)
	err := cmd.Run()
	if err != nil {
		panic(fmt.Errorf("[readRepoListFromHubRepo] run '%v' failed: %v", cmdStrs, err))
	}
	return readRepoListFromLocalHubRepo(repoPath)
}

func readRepoListFromLocalHubRepo(repoPath string) (addrs []string, helpStrs []string) {
	listPath := filepath.Join(repoPath, "README.md")
	data, err := ioutil.ReadFile(listPath)
	if err != nil {
		panic(fmt.Errorf("[readRepoListFromLocalHubRepo] read list file '%v' failed: %v",
			listPath, err))
	}
	list := strings.Split(string(data), "\n")
	for _, line := range list {
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

func isOsCmdExists(cmd string) bool {
	path, err := exec.LookPath(cmd)
	return err == nil && len(path) > 0
}
