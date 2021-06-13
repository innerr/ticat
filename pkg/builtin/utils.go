package builtin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func getFindStrsFromArgv(argv core.ArgVals) (findStrs []string) {
	names := []string{
		"1st-str",
		"2nd-str",
		"3rd-str",
		"4th-str",
		"5th-str",
		"6th-str",
	}
	for _, name := range names {
		val := argv.GetRaw(name)
		if len(val) != 0 {
			findStrs = append(findStrs, val)
		}
	}
	return
}

func addFindStrArgs(cmd *core.Cmd) {
	cmd.AddArg("1st-str", "", "find-str").
		AddArg("2nd-str", "").
		AddArg("3rh-str", "").
		AddArg("4th-str", "").
		AddArg("5th-str", "").
		AddArg("6th-str", "")
}

func normalizeCmdPath(path string, sep string, alterSeps string) string {
	var segs []string
	for len(path) > 0 {
		i := strings.IndexAny(path, alterSeps)
		if i < 0 {
			segs = append(segs, path)
			break
		} else if i == 0 {
			path = path[1:]
		} else {
			segs = append(segs, path[0:i])
			path = path[i+1:]
		}
	}
	return strings.Join(segs, sep)
}

func getCmdPath(path string, flowExt string) string {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, flowExt) {
		panic(fmt.Errorf("[getCmdPath] flow file '%s' ext not match '%s'", path, flowExt))
	}
	return base[:len(base)-len(flowExt)]
}

func quoteIfHasSpace(str string) string {
	if strings.IndexAny(str, " \t\r\n") < 0 {
		return str
	}
	i := strings.Index(str, "\"")
	if i < 0 {
		return "\"" + str + "\""
	}
	i = strings.Index(str, "'")
	if i < 0 {
		return "'" + str + "'"
	}
	return str
}

func clearFlow(flow *core.ParsedCmds) (int, bool) {
	flow.Cmds = nil
	return 0, true
}

func getAndCheckArg(argv core.ArgVals, env *core.Env, cmd core.ParsedCmd, arg string) string {
	val := argv.GetRaw(arg)
	if len(val) == 0 {
		panic(core.NewCmdError(cmd, "arg '"+arg+"' is empty"))
	}
	return val
}

func isOsCmdExists(cmd string) bool {
	path, err := exec.LookPath(cmd)
	return err == nil && len(path) > 0
}

func osRemoveDir(path string, cmd core.ParsedCmd) {
	path = strings.TrimSpace(path)
	if len(path) <= 1 {
		panic(core.WrapCmdError(cmd, fmt.Errorf("removing path '%v', looks not right", path)))
	}
	err := os.RemoveAll(path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		panic(core.WrapCmdError(cmd, fmt.Errorf("remove repo '%s' failed: %v", path, err)))
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && !info.IsDir()
}
