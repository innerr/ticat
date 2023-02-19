package display

import (
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func HandleParseResult(
	cc *core.Cli,
	flow *core.ParsedCmds,
	env *core.Env,
	isSearch bool) bool {

	if flow.AttempTailModeCall && len(flow.Cmds) == 2 &&
		flow.Cmds[0].ParseResult.Error == nil && flow.Cmds[1].ParseResult.Error != nil {

		PrintErrTitle(cc.Screen, env,
			"["+flow.Cmds[0].DisplayPath(cc.Cmds.Strs.PathSep, true)+"] not support tail-mode call.",
			"",
			"it is using the wrong command?")
		return false
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
			return PrintCmdByParseError(cc, cmd, env, "doesn't have args")
		case core.ParseErrEnv:
			PrintErrTitle(cc.Screen, env,
				"["+cmd.DisplayPath(cc.Cmds.Strs.PathSep, true)+"] parse env failed.",
				"",
				"'"+inputStr+"' is not valid input.",
				"",
				"env setting examples:",
				"",
				SuggestEnvSetting(env),
				"")
			return false
		case core.ParseErrExpectArgs:
			return PrintCmdByParseError(cc, cmd, env, "parse args failed")
		case core.ParseErrExpectCmd:
			return PrintSubCmdByParseError(cc, flow, cmd, env, isSearch)
		default:
			return PrintFindResultByParseError(cc, cmd, env, "")
		}
	}
	return true
}

func PrintCmdByParseError(
	cc *core.Cli,
	cmd core.ParsedCmd,
	env *core.Env,
	title string) bool {

	sep := cc.Cmds.Strs.PathSep
	cmdName := cmd.DisplayPath(sep, true)
	printer := NewTipBoxPrinter(cc.Screen, env, true)
	input := cmd.ParseResult.Input

	printer.PrintWrap(
		"["+cmdName+"] "+title+".",
		"",
		"'"+strings.Join(input, " ")+"' is not valid input.")
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
	isSearch bool) bool {

	sep := cc.Cmds.Strs.PathSep
	cmdName := cmd.DisplayPath(sep, true)
	printer := NewTipBoxPrinter(cc.Screen, env, true)
	input := cmd.ParseResult.Input

	last := cmd.LastCmdNode()
	if last == nil {
		return PrintFreeSearchResultByParseError(cc, flow, env, isSearch, input...)
	}
	printer.PrintWrap(
		"parse sub command under ["+cmdName+"] failed.",
		"",
		"'"+strings.Join(input, " ")+"' is not valid input.")
	if last.HasSubs() {
		printer.Prints("", "commands on branch '"+last.DisplayPath()+"':")
		dumpArgs := NewDumpCmdArgs().SetSkeleton()
		printer.Finish()
		DumpCmds(last, cc.Screen, env, dumpArgs)
	} else {
		printer.Prints(
			"",
			"'"+last.DisplayPath()+"' branch doesn't have any sub commands.",
			"",
			"search commands by:",
			"")
		for _, line := range SuggestFindCmds(env) {
			printer.Prints(line)
		}
		printer.Finish()
	}
	return false
}

func PrintFreeSearchResultByParseError(
	cc *core.Cli,
	flow *core.ParsedCmds,
	env *core.Env,
	isSearch bool,
	findStr ...string) bool {

	selfName := env.GetRaw("strs.self-name")
	input := findStr
	inputStr := strings.Join(input, " ")
	notValidStr := "'" + inputStr + "' is not valid input."

	var lines int
	for len(input) > 0 {
		screen := NewCacheScreen()
		dumpArgs := NewDumpCmdArgs().SetSkeleton().AddFindStrs(input...)
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
		PrintErrTitle(cc.Screen, env, helpStr)
		screen.WriteTo(cc.Screen)
		return false
	}

	if isSearch {
		helpStr := []string{
			"search but no commands matched '" + inputStr + "'.",
			"",
			"try to change keywords on the leftside, ",
			selfName + " will filter results by kewords from left to right.",
		}
		helpStr = append([]string{notValidStr, ""}, helpStr...)
		PrintErrTitle(cc.Screen, env, notValidStr)
	} else {
		PrintErrTitle(cc.Screen, env, notValidStr)
	}
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
		PrintErrTitle(cc.Screen, env,
			title,
			"",
			"'"+inputStr+"' is not valid input, found related commands by search:")
		screen.WriteTo(cc.Screen)
	} else {
		PrintErrTitle(cc.Screen, env,
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
