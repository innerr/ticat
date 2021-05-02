package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func PrintCmdResult(
	isBootstrap bool,
	screen core.Screen,
	cmd core.ParsedCmd,
	env *core.Env,
	succeeded bool,
	elapsed time.Duration,
	flow []core.ParsedCmd,
	currCmdIdx int,
	strs *core.CmdTreeStrs) {

	if isBootstrap && !env.GetBool("display.bootstrap") || !env.GetBool("display.executor") {
		return
	}
	if checkPrintFilter(cmd, env) {
		return
	}
	flow, currCmdIdx = filterQuietCmds(env, flow, currCmdIdx)
	if len(flow) == 1 && !env.GetBool("display.one-cmd") {
		return
	}
	if len(flow) == 0 {
		return
	}

	width := env.GetInt("display.width")

	var resStr string
	if succeeded {
		resStr = " ✓" //"OK"
	} else {
		resStr = " ✘" //"EE"
	}

	timeStr := time.Now().Format("01-02 15:04:05")

	printRealname := env.GetBool("display.mod.realname")
	line := " " + getCmdPath(cmd, strs.PathSep, printRealname)
	durStr := formatDuration(elapsed)
	if width-len(durStr)-6 < 20 {
		width = len(durStr) + 6 + 20
	}
	line = "│" + " " + resStr + padRight(line, " ", width-len(durStr)-6) + durStr + " " + "│"

	screen.Print("┌" + strings.Repeat("─", width-2) + "┐" + "\n")
	screen.Print(line + "\n")
	/*
		for k, v := range argv {
			line = "│" + padRight(strings.Repeat(" ", len(resStr) + 4) + k + " = " + v.Raw,
				" ", width-3) + " " + "│"
			screen.Print(line + "\n")
		}
	*/
	screen.Print("└" + strings.Repeat("─", width-2) + "┘" + "\n")

	if currCmdIdx >= len(flow)-1 || !succeeded {
		screen.Print(strings.Repeat(" ", width-len(timeStr)-2) + timeStr + "\n")
	} else {
		screen.Print("\n\n")
	}
}

func PrintCmdStack(
	isBootstrap bool,
	screen core.Screen,
	cmd core.ParsedCmd,
	env *core.Env,
	flow []core.ParsedCmd,
	currCmdIdx int,
	strs *core.CmdTreeStrs) {

	if isBootstrap && !env.GetBool("display.bootstrap") || !env.GetBool("display.executor") {
		return
	}
	if checkPrintFilter(cmd, env) {
		return
	}
	flow, currCmdIdx = filterQuietCmds(env, flow, currCmdIdx)
	if len(flow) == 1 && !env.GetBool("display.one-cmd") {
		return
	}
	if len(flow) == 0 {
		return
	}

	width := env.GetInt("display.width")

	const cmdCntKey = "display.max-cmd-cnt"
	cmdDisplayCnt := env.GetInt(cmdCntKey)
	if cmdDisplayCnt < 4 {
		panic(fmt.Errorf("[PrintCmdStack] %s should not less than 4", cmdCntKey))
	}

	printEnv := env.GetBool("display.env")
	printEnvLayer := env.GetBool("display.env.layer")
	printDefEnv := env.GetBool("display.env.default")
	printRuntimeEnv := env.GetBool("display.env.sys")
	printRealname := env.GetBool("display.mod.realname")

	stackDepth := env.Get("sys.stack-depth").Raw
	if len(stackDepth) > 2 {
		stackDepth = "[..]"
	} else {
		stackDepth = "[" + stackDepth + "]" + strings.Repeat(" ", 2+1-len(stackDepth))
	}
	stackDepth = " stack-level: " + stackDepth
	titleLine := "│" + stackDepth + "│"

	// Notice that len(str) will return wrong size since we are using non-ascii chars
	titleWidth := len(stackDepth) + 2

	titleLine += "   (=`ω´=)   "
	titleLineLen := titleWidth + 3 + 7 + 3

	timeStr := time.Now().Format("01-02 15:04:05")
	// This "1" is for left border
	if width < titleLineLen+len(timeStr)+1 {
		width = titleLineLen + len(timeStr) + 1
	}
	if width < 1+titleLineLen+len(timeStr)+1 {
		width = 1 + titleLineLen + len(timeStr) + 1
	}
	titleLine += strings.Repeat(" ", width-1-titleLineLen-len(timeStr)-1) + timeStr

	topBorder := "├" + strings.Repeat("─", titleWidth-2) + "┴"
	topBorder = topBorder + strings.Repeat("─", width-1-titleWidth) + "┐"

	displayIdxStart := 0
	displayIdxEnd := len(flow)
	if len(flow) > cmdDisplayCnt {
		displayIdxStart = currCmdIdx - cmdDisplayCnt/2
		if displayIdxStart < 0 {
			displayIdxStart = 0
		}
		if displayIdxStart+cmdDisplayCnt > len(flow) {
			displayIdxEnd = len(flow)
			displayIdxStart = displayIdxEnd - cmdDisplayCnt
		} else {
			displayIdxEnd = displayIdxStart + cmdDisplayCnt
		}
	}

	screen.Print("┌" + strings.Repeat("─", titleWidth-2) + "┐" + "\n")
	screen.Print(titleLine + "\n")
	screen.Print(topBorder + "\n")

	if printEnv {
		filterPrefixs := []string{strings.Join(cmd.Path(), strs.PathSep) + strs.PathSep}
		envLines := dumpEnv(env, printEnvLayer, printDefEnv, printRuntimeEnv, false, filterPrefixs, 4)
		for _, line := range envLines {
			screen.Print("│" + padRight("    "+line, " ", width-2) + "│" + "\n")
		}
		if len(envLines) != 0 {
			screen.Print("├" + strings.Repeat("─", width-2) + "┤" + "\n")
		}
	}

	for i, cmd := range flow {
		if i < displayIdxStart || i >= displayIdxEnd {
			continue
		}
		var line string
		if (i == displayIdxStart && i != 0) || (i+1 == displayIdxEnd && i+1 != len(flow)) {
			line += "    ..."
		} else {
			if i == currCmdIdx {
				line += " >> "
			} else {
				line += "    "
			}
			line += getCmdPath(cmd, strs.PathSep, printRealname)
		}
		screen.Print("│" + padRight(line, " ", width-2) + "│" + "\n")

		cmdEnv := cmd.GenEnv(env.GetLayer(core.EnvLayerSession),
			strs.EnvValDelMark, strs.EnvValDelAllMark)
		args := cmd.Args()
		argv := cmdEnv.GetArgv(cmd.Path(), strs.PathSep, cmd.Args())
		for _, line := range DumpArgs(&args, argv, false) {
			screen.Print("│" + padRight(strings.Repeat(" ", 8)+line, " ", width-2) + "│" + "\n")
		}
	}

	screen.Print("└" + strings.Repeat("─", width-2) + "┘" + "\n")
}

func checkPrintFilter(cmd core.ParsedCmd, env *core.Env) bool {
	if len(cmd) == 0 {
		return true
	}
	lastSeg := cmd[len(cmd)-1]
	if lastSeg.Cmd.Cmd == nil || lastSeg.Cmd.Cmd.Cmd() == nil ||
		(lastSeg.Cmd.Cmd.IsQuiet() && !env.GetBool("display.mod.quiet")) {
		return true
	}
	return false
}

func filterQuietCmds(env *core.Env, flow []core.ParsedCmd, currCmdIdx int) ([]core.ParsedCmd, int) {
	if env.GetBool("display.mod.quiet") {
		return flow, currCmdIdx
	}

	var newCmds []core.ParsedCmd
	newIdx := currCmdIdx
	for i, cmd := range flow {
		if len(cmd) == 0 {
			continue
		}
		lastSeg := cmd[len(cmd)-1].Cmd
		if lastSeg.Cmd == nil || lastSeg.Cmd.Cmd() == nil || lastSeg.Cmd.IsQuiet() {
			if i < currCmdIdx {
				newIdx -= 1
			}
			continue
		}
		newCmds = append(newCmds, cmd)
	}
	return newCmds, newIdx
}

func padRight(str string, pad string, width int) string {
	if len(str) >= width {
		return str
	}
	return str + strings.Repeat(pad, width-len(str))
}

func formatDuration(dur time.Duration) string {
	return strings.ReplaceAll(fmt.Sprintf("%s", dur), "µ", "u")
}
