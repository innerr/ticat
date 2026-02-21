package builtin

import (
	"strings"

	"github.com/innerr/ticat/pkg/cli/display"
	"github.com/innerr/ticat/pkg/core/model"
)

func DumpCmdUsage(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	cmdPath, err := tailModeCallArg(flow, currCmdIdx, argv, "cmd-path")
	if err != nil {
		return currCmdIdx, err
	}
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage().NoRecursive()
	dumpCmdByPath(cc, env, dumpArgs, cmdPath, "cmd.full")
	return currCmdIdx, nil
}

func DumpCmdWithDetails(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	cmdPath, err := tailModeCallArg(flow, currCmdIdx, argv, "cmd-path")
	if err != nil {
		return currCmdIdx, err
	}
	dumpArgs := display.NewDumpCmdArgs().NoRecursive()
	dumpCmdByPath(cc, env, dumpArgs, cmdPath, "")
	return currCmdIdx, nil
}

func DumpCmdWithDetailsAndFlow(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	cmdPath, err := tailModeCallArg(flow, currCmdIdx, argv, "cmd-path")
	if err != nil {
		return currCmdIdx, err
	}

	cmdTree := cc.Cmds.GetSubByPath(cmdPath, true)
	printUsageExample(cc, env, cmdPath, cmdTree)

	dumpArgs := display.NewDumpCmdArgs().NoRecursive()
	dumpCmdByPath(cc, env, dumpArgs, cmdPath, "")

	if cmdTree != nil && cmdTree.Cmd() != nil && cmdTree.Cmd().HasSubFlow(true) {
		subFlow, _, _ := cmdTree.Cmd().Flow(argv, cc, env, true, true)
		if len(subFlow) > 0 {
			parsedFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
			dumpFlowArgs := display.NewDumpFlowArgs().SetSkeleton().SetMaxDepth(32).SetSkipTipTitle()
			display.DumpFlow(cc, env, parsedFlow, 0, dumpFlowArgs, EnvOpCmds())
		}
	}

	return currCmdIdx, nil
}

func printUsageExample(cc *model.Cli, env *model.Env, cmdPath string, cmdTree *model.CmdTree) {
	selfName := env.GetRaw("strs.self-name")
	argsExample := ""
	if cmdTree != nil && cmdTree.Cmd() != nil {
		args := cmdTree.Cmd().Args()
		names := args.Names()
		if len(names) > 0 {
			argParts := []string{}
			for _, name := range names {
				argParts = append(argParts, name+"=<"+name+">")
			}
			argsExample = " " + strings.Join(argParts, " ")
		}
	}
	_ = cc.Screen.Print(display.ColorTip("usage:\n    $> "+selfName+" "+cmdPath+argsExample, env) + "\n")
}

func DumpTailCmdWithUsage(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	err := flow.FirstErr()
	if err != nil {
		return currCmdIdx, err.Error
	}

	cmdPath := argv.GetRaw("cmd-path")
	if len(cmdPath) == 0 {
		cmdPath = flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
	} else {
		cmdPath = cc.NormalizeCmd(true, cmdPath)
	}
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage().NoRecursive()
	dumpCmdByPath(cc, env, dumpArgs, cmdPath, "==")
	return clearFlow(flow)
}

func DumpTailCmdWithDetails(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	err := flow.FirstErr()
	if err != nil {
		return currCmdIdx, err.Error
	}

	cmdPath := argv.GetRaw("cmd-path")
	if len(cmdPath) == 0 {
		cmdPath = flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
	} else {
		cmdPath = cc.NormalizeCmd(true, cmdPath)
	}
	dumpArgs := display.NewDumpCmdArgs().NoRecursive()
	dumpCmdByPath(cc, env, dumpArgs, cmdPath, "")
	return clearFlow(flow)
}

func dumpCmdByPath(cc *model.Cli, env *model.Env, args *display.DumpCmdArgs, path string, fullDetailCmd string) {
	cmds := cc.Cmds
	if len(path) != 0 {
		cmds = cmds.GetSubByPath(path, true)
	}
	if args.Skeleton {
		if len(fullDetailCmd) != 0 {
			display.PrintTipTitle(cc.Screen, env,
				"command usage: (use '"+fullDetailCmd+"' for full details)")
		} else {
			display.PrintTipTitle(cc.Screen, env, "command usage:")
		}
	} else {
		display.PrintTipTitle(cc.Screen, env, "full command details:")
	}
	display.DumpCmds(cmds, cc.Screen, env, args)
}
