package builtin

import (
	"path/filepath"

	"github.com/innerr/ticat/pkg/core/model"
)

func ApiCmdType(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	cmdStr, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "cmd")
	if err != nil {
		return currCmdIdx, err
	}
	cmd, _ := cc.ParseCmd(true, cmdStr)
	node := cmd.LastCmdNode()
	if node != nil {
		if node.Cmd() != nil {
			_ = cc.Screen.Print(string(node.Cmd().Type()) + "\n")
		}
	}
	return currCmdIdx, nil
}

func ApiCmdMeta(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	cmdStr, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "cmd")
	if err != nil {
		return currCmdIdx, err
	}
	cmd, _ := cc.ParseCmd(true, cmdStr)
	node := cmd.LastCmdNode()
	if node != nil {
		if node.Cmd() != nil {
			_ = cc.Screen.Print(node.Cmd().MetaFile() + "\n")
		}
	}
	return currCmdIdx, nil
}

func ApiCmdPath(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	cmdStr, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "cmd")
	if err != nil {
		return currCmdIdx, err
	}
	cmd, _ := cc.ParseCmd(true, cmdStr)
	node := cmd.LastCmdNode()
	if node != nil {
		cic := node.Cmd()
		if cic != nil {
			line := cic.CmdLine()
			if len(line) != 0 && cic.Type() != model.CmdTypeEmptyDir {
				_ = cc.Screen.Print(line + "\n")
			}
		}
	}
	return currCmdIdx, nil
}

func ApiCmdDir(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	cmdStr, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "cmd")
	if err != nil {
		return currCmdIdx, err
	}
	cmd, _ := cc.ParseCmd(true, cmdStr)
	node := cmd.LastCmdNode()
	if node != nil {
		cic := node.Cmd()
		if cic != nil {
			if cic.Type() == model.CmdTypeEmptyDir {
				_ = cc.Screen.Print(node.Cmd().CmdLine() + "\n")
			} else {
				dir := filepath.Dir(node.Cmd().MetaFile())
				_ = cc.Screen.Print(dir + "\n")
			}
		}
	}
	return currCmdIdx, nil
}

func ApiCmdListAll(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	cmdDumpName(cc.Cmds, cc.Screen)
	return currCmdIdx, nil
}

// ApiCmdJson returns combined command info (type, meta, path, dir) as a single JSON object.
func ApiCmdJson(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	cmdStr, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "cmd")
	if err != nil {
		return currCmdIdx, err
	}
	cmd, _ := cc.ParseCmd(true, cmdStr)
	node := cmd.LastCmdNode()
	result := map[string]string{
		"command": cmdStr,
	}
	if node != nil {
		cic := node.Cmd()
		if cic != nil {
			result["type"] = string(cic.Type())
			result["meta_file"] = cic.MetaFile()
			line := cic.CmdLine()
			if len(line) != 0 && cic.Type() != model.CmdTypeEmptyDir {
				result["path"] = line
			}
			if cic.Type() == model.CmdTypeEmptyDir {
				result["dir"] = cic.CmdLine()
			} else {
				result["dir"] = filepath.Dir(cic.MetaFile())
			}
		}
	}
	return currCmdIdx, model.OutputJson(cc, result)
}

// ApiCmdJsonListAll returns all command names as a JSON array.
func ApiCmdJsonListAll(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	var names []string
	cmdCollectNames(cc.Cmds, &names)
	return currCmdIdx, model.OutputJson(cc, map[string]any{
		"commands": names,
	})
}

func cmdCollectNames(cmd *model.CmdTree, names *[]string) {
	if !cmd.IsEmpty() {
		*names = append(*names, cmd.DisplayPath())
	}
	for _, name := range cmd.SubNames() {
		cmdCollectNames(cmd.GetSub(name), names)
	}
}

func cmdDumpName(cmd *model.CmdTree, screen model.Screen) {
	if !cmd.IsEmpty() {
		_ = screen.Print(cmd.DisplayPath() + "\n")
	}
	for _, name := range cmd.SubNames() {
		cmdDumpName(cmd.GetSub(name), screen)
	}
}
