package builtin

import (
	"strings"

	"github.com/pingcap/ticat/pkg/cli/display"
	"github.com/pingcap/ticat/pkg/core/model"
)

func ListTags(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	screen := display.NewCacheScreen()
	display.ListTags(cc.Cmds, screen, env)
	if screen.OutputNum() > 0 {
		display.PrintTipTitle(cc.Screen, env, "all tags:")
	} else {
		display.PrintTipTitle(cc.Screen, env, "no command have tags")
	}
	screen.WriteTo(cc.Screen)
	return currCmdIdx, true
}

func FindByTags(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton()
	return findByTags(argv, cc, env, flow, currCmdIdx, dumpArgs)
}

func FindByTagsWithUsage(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs().SetSkeleton().SetShowUsage()
	return findByTags(argv, cc, env, flow, currCmdIdx, dumpArgs)
}

func FindByTagsWithDetails(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpCmdArgs()
	return findByTags(argv, cc, env, flow, currCmdIdx, dumpArgs)
}

func findByTags(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int,
	dumpArgs *display.DumpCmdArgs) (int, bool) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
	if len(findStrs) == 0 {
		return ListTags(argv, cc, env, flow, currCmdIdx)
	} else {
		dumpArgs.AddFindStrs(findStrs...).SetFindByTags()
	}

	findStr := strings.Join(dumpArgs.FindStrs, " ")
	if len(findStrs) > 1 {
		findStr = "tags '" + findStr
	} else {
		findStr = "tag '" + findStr
	}
	findStr = " commands matched " + findStr + "'"

	screen := display.NewCacheScreen()
	display.DumpCmds(cc.Cmds, screen, env, dumpArgs)
	if screen.OutputNum() > 0 {
		display.PrintTipTitle(cc.Screen, env, "search and found"+findStr)
	} else {
		display.PrintTipTitle(cc.Screen, env, "search but no"+findStr+", change keywords and retry")
	}

	screen.WriteTo(cc.Screen)

	if display.TooMuchOutput(env, screen) {
		if !dumpArgs.Skeleton || dumpArgs.ShowUsage {
			text := append([]string{"get a brief view by:", ""}, display.SuggestFindByTagLite(env)...)
			display.PrintTipTitle(cc.Screen, env, text)
		} else if len(findStrs) > 1 {
			display.PrintTipTitle(cc.Screen, env,
				"narrow down results by using less tags")
		}
	}
	return currCmdIdx, true
}
