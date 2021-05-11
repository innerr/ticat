package builtin

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func RemoveFlow(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	cmdPath, filePath := getFlowCmdPath(argv, cc, env, true, "cmd-path", "RemoveFlow")
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		panic(fmt.Errorf("[RemoveFlow] path '%s' does not exist", filePath))
	}
	err = os.Remove(filePath)
	if err != nil {
		panic(fmt.Errorf("[RemoveFlow] remove flow file '%s' failed: %v",
			filePath, err))
	}
	cc.Screen.Print(fmt.Sprintf("=> %s\n", cmdPath))
	cc.Screen.Print(fmt.Sprintf("   %s\n", filePath))
	cc.Screen.Print(fmt.Sprintf("   (removed)\n"))
	return true
}

func SaveFlow(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath, filePath := getFlowCmdPath(argv, cc, env, false, "to-cmd-path", "SaveFlow")

	_, err := os.Stat(filePath)
	if !os.IsNotExist(err) {
		panic(fmt.Errorf("[SaveFlow] path '%s' exists", filePath))
	}

	w := bytes.NewBuffer(nil)
	flow.RmLeadingCmds(1)
	saveFlow(w, flow, cc.Cmds.Strs.PathSep, env)
	data := w.String()
	cc.Screen.Print(strings.Repeat("-", 80) + "\n")
	cc.Screen.Print(data)
	cc.Screen.Print(strings.Repeat("-", 80) + "\n")
	cc.Screen.Print(fmt.Sprintf("=> %s\n", cmdPath))
	cc.Screen.Print(fmt.Sprintf("   %s\n", filePath))

	dirPath := filepath.Dir(filePath)
	os.MkdirAll(dirPath, os.ModePerm)

	tmp := filePath + ".tmp"
	err = ioutil.WriteFile(tmp, []byte(data), 0644)
	if err != nil {
		panic(fmt.Errorf("[SaveFlow] write flow file '%s' failed: %v", tmp, err))
	}

	err = os.Rename(tmp, filePath)
	if err != nil {
		panic(fmt.Errorf("[SaveFlow] rename flow file from '%s' to '%s' failed: %v",
			tmp, filePath, err))
	}

	flow.Cmds = nil
	return 0, true
}

func LoadLocalFlows(_ core.ArgVals, cc *core.Cli, env *core.Env) bool {
	root := env.GetRaw("sys.paths.flows")
	if len(root) > 0 && root[len(root)-1] == filepath.Separator {
		root = root[:len(root)-1]
	}
	info, err := os.Stat(root)
	if os.IsNotExist(err) {
		return true
	} else {
		panic(fmt.Errorf("[LoadLocalFlows] access flows dir '%s' failed: %v", root, err))
	}
	if !info.IsDir() {
		panic(fmt.Errorf("[LoadLocalFlows] flows dir '%s' is not dir", root))
	}

	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		cmdPath := path[len(root)+1:]
		flow := cc.Cmds.GetOrAddSub(strings.Split(cmdPath, string(cc.Cmds.Strs.PathSep))...)
		data, err := ioutil.ReadFile(path)
		cmds := strings.TrimSpace(string(data))
		cmd := flow.RegFlowCmd(cmds, "(TODO: save/load help str of "+cmdPath+")")
		// TODO: abbrs, help
		_ = cmd
		return nil
	})
	return true
}

func saveFlow(w io.Writer, flow *core.ParsedCmds, cmdPathSep string, env *core.Env) {
	for i, cmd := range flow.Cmds {
		if len(flow.Cmds) > 1 {
			if i == 0 {
				if flow.GlobalSeqIdx < 0 {
					fmt.Fprint(w, ": ")
				}
			} else {
				fmt.Fprint(w, " : ")
			}
		}

		var path []string
		var lastSegHasNoCmd bool
		var cmdHasEnv bool

		for j, seg := range cmd {
			if len(cmd) > 1 && j != 0 && !lastSegHasNoCmd {
				fmt.Fprint(w, cmdPathSep)
			}
			fmt.Fprint(w, seg.Cmd.Name)

			if seg.Cmd.Cmd != nil {
				path = append(path, seg.Cmd.Cmd.Name())
			} else {
				path = append(path, seg.Cmd.Name)
			}
			lastSegHasNoCmd = (seg.Cmd.Cmd == nil)
			cmdHasEnv = saveEnv(w, seg.Env, path, cmdPathSep,
				!cmdHasEnv && j == len(cmd)-1) || cmdHasEnv
		}
	}
	fmt.Fprintf(w, "\n")
}

func saveEnv(
	w io.Writer,
	env core.ParsedEnv,
	prefixPath []string,
	cmdPathSep string,
	useArgsFmt bool) bool {

	if len(env) == 0 {
		return false
	}

	isAllArgs := true
	for _, v := range env {
		if !v.IsArg {
			isAllArgs = false
			break
		}
	}

	prefix := strings.Join(prefixPath, cmdPathSep) + cmdPathSep

	var kvs []string
	for k, v := range env {
		if strings.HasPrefix(k, prefix) && len(k) != len(prefix) {
			k = strings.Join(v.MatchedPath[len(prefixPath):], cmdPathSep)
		}
		kvs = append(kvs, fmt.Sprintf("%v=%v", k, v.Val))
	}

	format := "{%s}"
	if isAllArgs && useArgsFmt {
		format = " %s"
	}
	fmt.Fprintf(w, format, strings.Join(kvs, " "))
	return true
}

func getFlowCmdPath(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	expectExists bool,
	argName string,
	funcName string) (cmdPath string, filePath string) {

	cmdPath = normalizeCmdPath(argv.GetRaw(argName),
		cc.Cmds.Strs.PathSep, cc.Cmds.Strs.PathAlterSeps)
	if len(cmdPath) == 0 {
		panic(fmt.Errorf("[%s] arg '%s' is empty", funcName, argName))
	}

	root := env.GetRaw("sys.paths.flows")
	if len(root) == 0 {
		panic(fmt.Errorf("[%s] env 'sys.paths.flows' is empty", funcName))
	}

	filePath = filepath.Join(root, cmdPath)
	if !expectExists && fileExists(filePath) {
		panic(fmt.Errorf("[%s] cmd path '%s' is not empty", funcName, cmdPath))
	} else if expectExists && !fileExists(filePath) {
		panic(fmt.Errorf("[%s] cmd path '%s' is empty", funcName, cmdPath))
	}
	return
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
