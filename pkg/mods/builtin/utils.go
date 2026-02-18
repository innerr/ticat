package builtin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/innerr/ticat/pkg/core/model"
)

func assertNotTailMode(flow *model.ParsedCmds, currCmdIdx int) error {
	if flow.HasTailMode && !flow.TailModeCall && flow.Cmds[currCmdIdx].TailMode && len(flow.Cmds) > 1 {
		return model.NewCmdError(flow.Cmds[currCmdIdx], "tail-mode not support")
	}
	return nil
}

/*
func assertNotTailModeFlow(flow *model.ParsedCmds, currCmdIdx int) error {
	if flow.HasTailMode && !flow.TailModeCall && flow.Cmds[currCmdIdx].TailMode && len(flow.Cmds) > 1 {
		return model.NewCmdError(flow.Cmds[currCmdIdx], "tail-mode flow not support")
	}
	return nil
}

func assertNotTailModeCall(flow *model.ParsedCmds, currCmdIdx int) error {
	if flow.TailModeCall {
		return model.NewCmdError(flow.Cmds[currCmdIdx], "tail-mode call not support")
	}
	return nil
}
*/

func tailModeCallArg(
	flow *model.ParsedCmds,
	currCmdIdx int,
	argv model.ArgVals,
	arg string) (string, error) {

	args, err := tailModeCallArgs(flow, currCmdIdx, argv, arg, false)
	if err != nil {
		return "", err
	}
	return args[0], nil
}

func tailModeCallArgs(
	flow *model.ParsedCmds,
	currCmdIdx int,
	argv model.ArgVals,
	arg string,
	allowMultiArgs bool) ([]string, error) {

	val := argv.GetRaw(arg)
	if flow.TailModeCall && !flow.Cmds[currCmdIdx].TailMode {
		return nil, model.NewCmdError(flow.Cmds[currCmdIdx],
			"should not happen: wrong command tail-mode flag")
	}
	if !flow.TailModeCall {
		if len(val) == 0 {
			return nil, model.NewCmdError(flow.Cmds[currCmdIdx], "arg '"+arg+"' is empty")
		}
		return []string{val}, nil
	}

	args := tailModeGetInput(flow, currCmdIdx, false)
	flowInputN := len(args)
	if len(val) != 0 {
		args = append(args, val)
	} else {
		if len(args) == 0 {
			return nil, model.NewCmdError(flow.Cmds[currCmdIdx], "arg '"+arg+"' is empty")
		}
	}
	if !allowMultiArgs && len(args) > 1 {
		if flowInputN > 0 && len(val) != 0 {
			return nil, model.NewCmdError(flow.Cmds[currCmdIdx],
				"mixed usage of arg '"+arg+"' and tail-mode call")
		} else {
			return nil, model.NewCmdError(flow.Cmds[currCmdIdx],
				"too many input of arg '"+arg+"' in tail-mode call")
		}
	}
	return args, nil
}

func tailModeGetInput(flow *model.ParsedCmds, currCmdIdx int, allowMultiCmds bool) (input []string) {
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

func clearFlow(flow *model.ParsedCmds) (int, error) {
	flow.Cmds = nil
	return 0, nil
}

func getFindStrsFromArgvAndFlow(flow *model.ParsedCmds, currCmdIdx int, argv model.ArgVals) (findStrs []string) {
	findStrs = getFindStrsFromArgv(argv)
	if flow.TailModeCall && flow.Cmds[currCmdIdx].TailMode {
		findStrs = append(findStrs, tailModeGetInput(flow, currCmdIdx, false)...)
	}
	return
}

func getFindStrsFromArgv(argv model.ArgVals) (findStrs []string) {
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

func addFindStrArgs(cmd *model.Cmd) {
	cmd.AddArg("1st-str", "", "find-str").
		AddArg("2nd-str", "").
		AddArg("3rd-str", "").
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

func getCmdPath(path string, flowExt string, cmd model.ParsedCmd) (string, error) {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, flowExt) {
		return "", model.NewCmdError(cmd, fmt.Sprintf("flow file '%s' ext not match '%s'", path, flowExt))
	}
	return base[:len(base)-len(flowExt)], nil
}

func getAndCheckArg(argv model.ArgVals, cmd model.ParsedCmd, arg string) (string, error) {
	val := argv.GetRaw(arg)
	if len(val) == 0 {
		return "", model.NewCmdError(cmd, "arg '"+arg+"' is empty")
	}
	return val, nil
}

func isOsCmdExists(cmd string) bool {
	path, err := exec.LookPath(cmd)
	return err == nil && len(path) > 0
}

func osRemoveDir(path string, cmd model.ParsedCmd) error {
	path = strings.TrimSpace(path)
	if len(path) <= 1 {
		return model.NewCmdError(cmd, fmt.Sprintf("removing path '%v', looks not right", path))
	}
	err := os.RemoveAll(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return model.NewCmdError(cmd, fmt.Sprintf("remove repo '%s' failed: %v", path, err))
	}
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && info.IsDir()
}
