package display

import (
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func PrintCmdByParseError(
	cc *core.Cli,
	cmd core.ParsedCmd,
	env *core.Env) bool {

	sep := cc.Cmds.Strs.PathSep
	cmdName := cmd.DisplayPath(sep, true)
	printer := NewTipBoxPrinter(cc.Screen, env, true)
	input := cmd.ParseError.Input

	printer.PrintWrap("[" + cmdName + "] parse args failed, '" +
		strings.Join(input, " ") + "' is not valid input.")
	printer.Prints("", "command detail:", "")
	DumpAllCmds(cmd.Last().Matched.Cmd, printer, false, 4, false, false)
	printer.Finish()
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
	input := cmd.ParseError.Input

	last := cmd.LastCmdNode()
	if last == nil {
		return PrintFreeSearchResultByParseError(cc, flow, env, isSearch, isMore, input...)
	}
	printer.PrintWrap("[" + cmdName + "] parse sub command failed, '" +
		strings.Join(input, " ") + "' is not valid input.")
	if last.HasSub() {
		printer.Prints("", "commands on branch '"+last.DisplayPath()+"':", "")
		DumpAllCmds(last, printer, true, 4, true, true)
	} else {
		printer.Prints("", "command branch '"+last.DisplayPath()+"' doesn't have any sub commands.")
		// TODO: search hint
	}
	printer.Finish()
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
		DumpAllCmds(cc.Cmds, screen, !isMore, 4, true, true, input...)
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
		PrintTipTitle(cc.Screen, env, helpStr...)
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
	PrintTipTitle(cc.Screen, env, helpStr...)
	return false
}

func PrintFindResultByParseError(
	cc *core.Cli,
	cmd core.ParsedCmd,
	env *core.Env,
	title string) bool {

	input := cmd.ParseError.Input
	inputStr := strings.Join(input, " ")
	screen := NewCacheScreen()
	DumpAllCmds(cc.Cmds, screen, true, 4, true, true, input...)

	if len(title) == 0 {
		title = cmd.ParseError.Error.Error()
	}

	if screen.OutputNum() > 0 {
		PrintTipTitle(cc.Screen, env, title, "",
			"'"+inputStr+"' is not valid input, found related commands by search:")
		screen.WriteTo(cc.Screen)
	} else {
		helpStr := []string{
			title, "",
			"'" + inputStr + "' is not valid input and no related commands found.",
			"", "try to change input,", "or search commands by:", "",
		}
		helpStr = append(helpStr, SuggestStrsFindCmds(env.GetRaw("strs.self-name"))...)
		helpStr = append(helpStr, "")
		PrintTipTitle(cc.Screen, env, helpStr...)
	}
	return false
}
