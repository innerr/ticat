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
			if model.IsJsonOutputMode(env) {
				return currCmdIdx, model.Output(cc, env, map[string]string{
					"command": cmdStr,
					"type":    string(node.Cmd().Type()),
				})
			}
			_, _ = fmt.Fprintf(os.Stdout, "%s\n", node.Cmd().Type())
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
			if model.IsJsonOutputMode(env) {
				return currCmdIdx, model.Output(cc, env, map[string]string{
					"command":   cmdStr,
					"meta_file": node.Cmd().MetaFile(),
				})
			}
			_, _ = fmt.Fprintf(os.Stdout, "%s\n", node.Cmd().MetaFile())
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
				if model.IsJsonOutputMode(env) {
					return currCmdIdx, model.Output(cc, env, map[string]string{
						"command": cmdStr,
						"path":    line,
					})
				}
				_, _ = fmt.Fprintf(os.Stdout, "%s\n", line)
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
			var dir string
			if cic.Type() == model.CmdTypeEmptyDir {
				dir = node.Cmd().CmdLine()
			} else {
				dir = filepath.Dir(node.Cmd().MetaFile())
			}
			if model.IsJsonOutputMode(env) {
				return currCmdIdx, model.Output(cc, env, map[string]string{
					"command": cmdStr,
					"dir":     dir,
				})
			}
			_, _ = fmt.Fprintf(os.Stdout, "%s\n", dir)
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
	if model.IsJsonOutputMode(env) {
		var names []string
		cmdCollectNames(cc.Cmds, &names)
		return currCmdIdx, model.Output(cc, env, map[string]any{
			"commands": names,
		})
	}
	cmdDumpName(cc.Cmds, cc.Screen)
	return currCmdIdx, nil
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
