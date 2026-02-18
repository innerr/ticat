package builtin

import (
	"fmt"

	"github.com/innerr/ticat/pkg/cli/display"
	"github.com/innerr/ticat/pkg/core/model"
)

func GlobalHelp(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	target := argv.GetRaw("target")
	if len(target) != 0 {
		cmdPath := cc.NormalizeCmd(false, target)
		if len(cmdPath) == 0 {
			display.PrintErrTitle(cc.Screen, env, fmt.Sprintf("'%s' is not a valid command", target))
		} else {
			dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage().NoRecursive()
			dumpCmdByPath(cc, env, dumpArgs, cmdPath, "")
		}
	} else {
		display.PrintGlobalHelp(cc, env)
	}
	return currCmdIdx, nil
}

func SelfHelp(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	display.PrintSelfHelp(cc.Screen, env)
	return currCmdIdx, nil
}

func GlobalFindCmds(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton()
	return globalFindCmds(argv, cc, env, flow, currCmdIdx, dumpArgs, "", "find.more")
}

func GlobalFindCmdsWithUsage(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage()
	return globalFindCmds(argv, cc, env, flow, currCmdIdx, dumpArgs, "find", "find.full")
}

func GlobalFindCmdsWithDetails(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	dumpArgs := display.NewDumpCmdArgs()
	return globalFindCmds(argv, cc, env, flow, currCmdIdx, dumpArgs, "find.more", "")
}

func globalFindCmds(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int,
	dumpArgs *display.DumpCmdArgs,
	lessDetailCmd string,
	moreDetailCmd string) (int, error) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
	if len(findStrs) != 0 {
		dumpArgs.AddFindStrs(findStrs...)
	}
	display.DumpCmdsWithTips(cc.Cmds, cc.Screen, env, dumpArgs, "", lessDetailCmd, moreDetailCmd)
	return clearFlow(flow)
}

func DumpCmdsWhoWriteKey(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	key, err := tailModeCallArg(flow, currCmdIdx, argv, "key")
	if err != nil {
		return currCmdIdx, err
	}
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetMatchWriteKey(key)

	screen := display.NewCacheScreen()
	display.DumpCmds(cc.Cmds, screen, env, dumpArgs)

	if screen.OutputtedLines() > 0 {
		display.PrintTipTitle(cc.Screen, env, "all commands which write key '"+key+"':")
	} else {
		display.PrintTipTitle(cc.Screen, env, "no command writes key '"+key+"':")
	}
	screen.WriteTo(cc.Screen)
	return currCmdIdx, nil
}

func DumpCmdsTree(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	firstErr := flow.FirstErr()
	if firstErr != nil {
		return currCmdIdx, firstErr.Error
	}

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().NoFlatten()

	cmdPath := ""
	cmds := cc.Cmds
	if len(argv.GetRaw("cmd-path")) != 0 {
		var err error
		cmdPath, err = tailModeCallArg(flow, currCmdIdx, argv, "cmd-path")
		if err != nil {
			return currCmdIdx, err
		}
		cmds = cmds.GetSubByPath(cmdPath, true)
	}

	depth := 0
	if len(argv.GetRaw("depth")) != 0 {
		depth = argv.GetInt("depth")
		dumpArgs.SetMaxDepth(depth)
	}

	screen := display.NewCacheScreen()
	allShown := display.DumpCmds(cmds, screen, env, dumpArgs)

	text := ""
	if len(cmdPath) == 0 {
		text = "the tree of all commands:"
	} else {
		text = "the tree branch of '" + cmdPath + "'"
	}
	if !allShown {
		text += fmt.Sprintf(", some may not shown by arg depth='%d'", depth)
	}

	display.PrintTipTitle(cc.Screen, env, text)
	screen.WriteTo(cc.Screen)
	return clearFlow(flow)
}
