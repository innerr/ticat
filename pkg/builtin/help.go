package builtin

import (
	"fmt"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func GlobalHelp(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	target := argv.GetRaw("target")
	if len(target) != 0 {
		cmdPath := cc.ParseCmd(target, false)
		if len(cmdPath) == 0 {
			display.PrintErrTitle(cc.Screen, env, fmt.Sprintf("'%s' is not a valid command", target))
		} else {
			dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage().NoRecursive()
			dumpCmdByPath(cc, env, dumpArgs, cmdPath)
		}
	} else {
		display.PrintGlobalHelp(cc, env)
	}
	return currCmdIdx, true
}

func SelfHelp(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	display.PrintSelfHelp(cc.Screen, env)
	return currCmdIdx, true
}

func GlobalFindCmds(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton()
	return globalFindCmds(argv, cc, env, flow, currCmdIdx, dumpArgs)
}

func GlobalFindCmdsWithUsage(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage()
	return globalFindCmds(argv, cc, env, flow, currCmdIdx, dumpArgs)
}

func GlobalFindCmdsWithDetails(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs()
	return globalFindCmds(argv, cc, env, flow, currCmdIdx, dumpArgs)
}

func globalFindCmds(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int,
	dumpArgs *display.DumpCmdArgs) (int, bool) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
	if len(findStrs) != 0 {
		dumpArgs.AddFindStrs(findStrs...)
	}

	findStr := strings.Join(dumpArgs.FindStrs, " ")
	matchStr := " commands matched '" + findStr + "'"
	selfName := env.GetRaw("strs.self-name")

	screen := display.NewCacheScreen()
	display.DumpCmds(cc.Cmds, screen, env, dumpArgs)

	if len(dumpArgs.FindStrs) != 0 {
		if screen.OutputNum() > 0 {
			if dumpArgs.Skeleton && screen.OutputNum() <= env.GetInt("display.height")/2 {
				display.PrintTipTitle(cc.Screen, env, "search and found"+matchStr,
					"",
					// TODO: XXX
					"get more details by using [find.with-usage|with-full-info] for search.")
			} else {
				display.PrintTipTitle(cc.Screen, env, "search and found"+matchStr)
			}
		} else {
			display.PrintTipTitle(cc.Screen, env, "search but no"+matchStr+", change find-strs and try again")
		}
	} else {
		if screen.OutputNum() > 0 {
			display.PrintTipTitle(cc.Screen, env, "all commands loaded to "+selfName+":")
		} else {
			display.PrintTipTitle(cc.Screen, env, selfName+" has no loaded commands. (this should never happen)")
		}
	}

	screen.WriteTo(cc.Screen)

	if display.TooMuchOutput(env, screen) {
		if !dumpArgs.Skeleton {
			display.PrintTipTitle(screen, env,
				// TODO: XXX
				"get a brief view by using [find] for search.",
				"",
				"or/and locate exact commands by adding more keywords:",
				"",
				display.SuggestFindCmds(env))
		} else {
			display.PrintTipTitle(screen, env,
				"locate exact commands by adding more keywords:",
				"",
				display.SuggestFindCmdsLess(env))
		}
	}

	return clearFlow(flow)
}

func DumpCmdsWhoWriteKey(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	key := tailModeCallArg(flow, currCmdIdx, argv, "key")
	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetMatchWriteKey(key)

	screen := display.NewCacheScreen()
	display.DumpCmds(cc.Cmds, screen, env, dumpArgs)

	if screen.OutputNum() > 0 {
		display.PrintTipTitle(cc.Screen, env, "all commands which write key '"+key+"':")
	} else {
		display.PrintTipTitle(cc.Screen, env, "no command writes key '"+key+"':")
	}
	screen.WriteTo(cc.Screen)
	return currCmdIdx, true
}

func DumpCmdsTree(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().NoFlatten()

	cmdPath := ""
	cmds := cc.Cmds
	if len(argv.GetRaw("cmd-path")) != 0 {
		cmdPath = tailModeCallArg(flow, currCmdIdx, argv, "cmd-path")
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
		text += fmt.Sprintf(", some are not showed by arg depth='%d'", depth)
	}

	display.PrintTipTitle(cc.Screen, env, text)
	screen.WriteTo(cc.Screen)
	return clearFlow(flow)
}
