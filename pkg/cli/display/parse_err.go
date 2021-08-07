package display

import (
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func HandleParseResult(
	cc *core.Cli,
	flow *core.ParsedCmds,
	env *core.Env,
	isSearch bool,
	isLess bool,
	isMore bool) bool {

	if isMore || isLess {
		return true
	}

	for _, cmd := range flow.Cmds {
		if cmd.ParseResult.Error == nil {
			continue
		}
		// TODO: better handling: sub flow parse failed
		/*
			stackDepth := env.GetInt("sys.stack-depth")
			if stackDepth > 0 {
				panic(cmd.ParseResult.Error)
			}
		*/

		input := cmd.ParseResult.Input
		inputStr := strings.Join(input, " ")

		switch cmd.ParseResult.Error.(type) {
		case core.ParseErrExpectNoArg:
			title := "[" + cmd.DisplayPath(cc.Cmds.Strs.PathSep, true) + "] doesn't have args."
			return PrintFindResultByParseError(cc, cmd, env, title)
		case core.ParseErrEnv:
			PrintTipTitle(cc.Screen, env,
				"["+cmd.DisplayPath(cc.Cmds.Strs.PathSep, true)+"] parse env failed, "+
					"'"+inputStr+"' is not valid input.",
				"",
				"env setting examples:",
				"",
				SuggestEnvSetting(env),
				"")
		case core.ParseErrExpectArgs:
			return PrintCmdByParseError(cc, cmd, env)
		case core.ParseErrExpectCmd:
			return PrintSubCmdByParseError(cc, flow, cmd, env, isSearch, isMore)
		default:
			return PrintFindResultByParseError(cc, cmd, env, "")
		}
	}
	return true
}

func PrintCmdByParseError(
	cc *core.Cli,
	cmd core.ParsedCmd,
	env *core.Env) bool {

	sep := cc.Cmds.Strs.PathSep
	cmdName := cmd.DisplayPath(sep, true)
	printer := NewTipBoxPrinter(cc.Screen, env, true)
	input := cmd.ParseResult.Input

	printer.PrintWrap("[" + cmdName + "] parse args failed, '" +
		strings.Join(input, " ") + "' is not valid input.")
	printer.Prints("", "command detail:")
	printer.Finish()
	dumpArgs := NewDumpCmdArgs().NoFlatten().NoRecursive()
	DumpCmds(cmd.Last().Matched.Cmd, cc.Screen, env, dumpArgs)
	return false
}

func PrintSubCmdByParseError(
	cc *core.Cli,
	flow *core.ParsedCmds,
	cmd core.ParsedCmd,
	env *core.Env,
	isSearch bool,
	isMore bool) bool {

	sep := cc.Cmds.Strs.PathSep
	cmdName := cmd.DisplayPath(sep, true)
	printer := NewTipBoxPrinter(cc.Screen, env, true)
	input := cmd.ParseResult.Input

	last := cmd.LastCmdNode()
	if last == nil {
		return PrintFreeSearchResultByParseError(cc, flow, env, isSearch, isMore, input...)
	}
	printer.PrintWrap("[" + cmdName + "] parse sub command failed, '" +
		strings.Join(input, " ") + "' is not valid input.")
	if last.HasSub() {
		printer.Prints("", "commands on branch '"+last.DisplayPath()+"':")
		dumpArgs := NewDumpCmdArgs().SetSkeleton()
		printer.Finish()
		DumpCmds(last, cc.Screen, env, dumpArgs)
	} else {
		printer.Prints("", "command branch '"+last.DisplayPath()+"' doesn't have any sub commands.")
		printer.Finish()
		// TODO: search hint
	}
	return false
}

func PrintFreeSearchResultByParseError(
	cc *core.Cli,
	flow *core.ParsedCmds,
	env *core.Env,
	isSearch bool,
	isMore bool,
	findStr ...string) bool {

	selfName := env.GetRaw("strs.self-name")
	input := findStr
	inputStr := strings.Join(input, " ")
	notValidStr := "'" + inputStr + "' is not valid input."

	var lines int
	for len(input) > 0 {
		screen := NewCacheScreen()
		dumpArgs := NewDumpCmdArgs().AddFindStrs(input...)
		dumpArgs.Skeleton = !isMore
		DumpCmds(cc.Cmds, screen, env, dumpArgs)
		lines = screen.OutputNum()
		if lines <= 0 {
			input = input[:len(input)-1]
			continue
		}
		helpStr := []string{
			"search and found commands matched '" + strings.Join(input, " ") + "':",
		}
		if !isSearch {
			helpStr = append([]string{notValidStr, ""}, helpStr...)
		}
		PrintTipTitle(cc.Screen, env, helpStr)
		screen.WriteTo(cc.Screen)
		return false
	}

	helpStr := []string{
		"search but no commands matched '" + inputStr + "'.",
		"",
		"try to change keywords on the leftside, ",
		selfName + " will filter results by kewords from left to right.",
	}
	if !isSearch {
		helpStr = append([]string{notValidStr, ""}, helpStr...)
	}
	PrintTipTitle(cc.Screen, env, helpStr)
	return false
}

func PrintFindResultByParseError(
	cc *core.Cli,
	cmd core.ParsedCmd,
	env *core.Env,
	title string) bool {

	input := cmd.ParseResult.Input
	inputStr := strings.Join(input, " ")
	screen := NewCacheScreen()
	dumpArgs := NewDumpCmdArgs().SetSkeleton().AddFindStrs(input...)
	DumpCmds(cc.Cmds, screen, env, dumpArgs)

	if len(title) == 0 {
		title = cmd.ParseResult.Error.Error()
	}

	if screen.OutputNum() > 0 {
		PrintTipTitle(cc.Screen, env,
			title,
			"",
			"'"+inputStr+"' is not valid input, found related commands by search:")
		screen.WriteTo(cc.Screen)
	} else {
		PrintTipTitle(cc.Screen, env,
			title,
			"",
			"'"+inputStr+"' is not valid input and no related commands found.",
			"",
			"try to change input,", "or search commands by:", "",
			SuggestFindCmds(env),
			"")
	}
	return false
}
