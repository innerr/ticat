package builtin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func ApiCmdType(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	cmdStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "cmd")
	cmd, _ := cc.ParseCmd(true, cmdStr)
	node := cmd.LastCmdNode()
	if node != nil {
		if node.Cmd() != nil {
			fmt.Fprintf(os.Stdout, "%s\n", node.Cmd().Type())
		}
	}
	return currCmdIdx, true
}

func ApiCmdMeta(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	cmdStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "cmd")
	cmd, _ := cc.ParseCmd(true, cmdStr)
	node := cmd.LastCmdNode()
	if node != nil {
		if node.Cmd() != nil {
			fmt.Fprintf(os.Stdout, "%s\n", node.Cmd().MetaFile())
		}
	}
	return currCmdIdx, true
}

func ApiCmdPath(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	cmdStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "cmd")
	cmd, _ := cc.ParseCmd(true, cmdStr)
	node := cmd.LastCmdNode()
	if node != nil {
		cic := node.Cmd()
		if cic != nil {
			line := cic.CmdLine()
			if len(line) != 0 && cic.Type() != core.CmdTypeEmptyDir {
				fmt.Fprintf(os.Stdout, "%s\n", line)
			}
		}
	}
	return currCmdIdx, true
}

func ApiCmdDir(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	cmdStr := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "cmd")
	cmd, _ := cc.ParseCmd(true, cmdStr)
	node := cmd.LastCmdNode()
	if node != nil {
		cic := node.Cmd()
		if cic != nil {
			if cic.Type() == core.CmdTypeEmptyDir {
				fmt.Fprintf(os.Stdout, "%s\n", node.Cmd().CmdLine())
			} else {
				dir := filepath.Dir(node.Cmd().MetaFile())
				fmt.Fprintf(os.Stdout, "%s\n", dir)
			}
		}
	}
	return currCmdIdx, true
}

func ApiCmdListAll(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	cmdDumpName(cc.Cmds, cc.Screen)
	return currCmdIdx, true
}

func cmdDumpName(cmd *core.CmdTree, screen core.Screen) {
	if !cmd.IsEmpty() {
		screen.Print(cmd.DisplayPath() + "\n")
	}
	for _, name := range cmd.SubNames() {
		cmdDumpName(cmd.GetSub(name), screen)
	}
}
