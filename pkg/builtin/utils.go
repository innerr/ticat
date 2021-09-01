package builtin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func assertNotTailMode(flow *core.ParsedCmds, currCmdIdx int) {
	if flow.HasTailMode && !flow.TailModeCall && flow.Cmds[currCmdIdx].TailMode && len(flow.Cmds) > 1 {
		panic(core.NewCmdError(flow.Cmds[currCmdIdx], "tail-mode not support"))
	}
}

func assertNotTailModeFlow(flow *core.ParsedCmds, currCmdIdx int) {
	if flow.HasTailMode && !flow.TailModeCall && flow.Cmds[currCmdIdx].TailMode && len(flow.Cmds) > 1 {
		panic(core.NewCmdError(flow.Cmds[currCmdIdx], "tail-mode flow not support"))
	}
}

func assertNotTailModeCall(flow *core.ParsedCmds, currCmdIdx int) {
	if flow.TailModeCall {
		panic(core.NewCmdError(flow.Cmds[currCmdIdx], "tail-mode call not support"))
	}
}

func tailModeCallArg(
	flow *core.ParsedCmds,
	currCmdIdx int,
	argv core.ArgVals,
	arg string) string {

	args := tailModeCallArgs(flow, currCmdIdx, argv, arg, false)
	return args[0]
}

func tailModeCallArgs(
	flow *core.ParsedCmds,
	currCmdIdx int,
	argv core.ArgVals,
	arg string,
	allowMultiArgs bool) []string {

	val := argv.GetRaw(arg)
	if flow.TailModeCall && !flow.Cmds[currCmdIdx].TailMode {
		panic(core.NewCmdError(flow.Cmds[currCmdIdx],
			"should not happen: wrong command tail-mode flag"))
	}
	if !flow.TailModeCall {
		if len(val) == 0 {
			panic(core.NewCmdError(flow.Cmds[currCmdIdx], "arg '"+arg+"' is empty"))
		}
		return []string{val}
	}

	args := tailModeGetInput(flow, currCmdIdx, false)
	flowInputN := len(args)
	if len(val) != 0 {
		args = append(args, val)
	} else {
		if len(args) == 0 {
			panic(core.NewCmdError(flow.Cmds[currCmdIdx], "arg '"+arg+"' is empty"))
		}
	}
	if !allowMultiArgs && len(args) > 1 {
		if flowInputN > 0 && len(val) != 0 {
			panic(core.NewCmdError(flow.Cmds[currCmdIdx],
				"mixed usage of arg '"+arg+"' and tail-mode call"))
		} else {
			panic(core.NewCmdError(flow.Cmds[currCmdIdx],
				"too many input of arg '"+arg+"' in tail-mode call"))
		}
	}
	return args
}

func tailModeGetInput(flow *core.ParsedCmds, currCmdIdx int, allowMultiCmds bool) (input []string) {
	if !flow.Cmds[currCmdIdx].TailMode {
		return
	}
	if len(flow.Cmds) <= 1 {
		return
	}
	if !allowMultiCmds {
		cmd := flow.Cmds[len(flow.Cmds)-1]
		input = append(input, cmd.ParseResult.Input...)
	} else {
		for _, cmd := range flow.Cmds[currCmdIdx+1:] {
			input = append(input, cmd.ParseResult.Input...)
		}
	}
	return
}

func clearFlow(flow *core.ParsedCmds) (int, bool) {
	flow.Cmds = nil
	return 0, true
}

func getFindStrsFromArgvAndFlow(flow *core.ParsedCmds, currCmdIdx int, argv core.ArgVals) (findStrs []string) {
	findStrs = getFindStrsFromArgv(argv)
	if flow.TailModeCall && flow.Cmds[currCmdIdx].TailMode {
		findStrs = append(findStrs, tailModeGetInput(flow, currCmdIdx, false)...)
	}
	return
}

func getFindStrsFromArgv(argv core.ArgVals) (findStrs []string) {
	names := []string{
		"1st-str",
		"2nd-str",
		"3rd-str",
		"4th-str",
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
		AddArg("4th-str", "")
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

func getCmdPath(path string, flowExt string, cmd core.ParsedCmd) string {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, flowExt) {
		panic(core.NewCmdError(cmd, fmt.Sprintf("flow file '%s' ext not match '%s'", path, flowExt)))
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

func getAndCheckArg(argv core.ArgVals, cmd core.ParsedCmd, arg string) string {
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
		panic(core.NewCmdError(cmd, fmt.Sprintf("removing path '%v', looks not right", path)))
	}
	err := os.RemoveAll(path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		panic(core.NewCmdError(cmd, fmt.Sprintf("remove repo '%s' failed: %v", path, err)))
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && !info.IsDir()
}
