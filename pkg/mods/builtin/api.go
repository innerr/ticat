package builtin

import (
	"fmt"
	"os"
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
			fmt.Fprintf(os.Stdout, "%s\n", node.Cmd().Type())
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
			fmt.Fprintf(os.Stdout, "%s\n", node.Cmd().MetaFile())
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
				fmt.Fprintf(os.Stdout, "%s\n", line)
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
				fmt.Fprintf(os.Stdout, "%s\n", node.Cmd().CmdLine())
			} else {
				dir := filepath.Dir(node.Cmd().MetaFile())
				fmt.Fprintf(os.Stdout, "%s\n", dir)
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

func cmdDumpName(cmd *model.CmdTree, screen model.Screen) {
	if !cmd.IsEmpty() {
		screen.Print(cmd.DisplayPath() + "\n")
	}
	for _, name := range cmd.SubNames() {
		cmdDumpName(cmd.GetSub(name), screen)
	}
}
