package cli

import (
	"fmt"
	"strings"
	"time"
)

func printCmdResult(isBootstrap bool, screen *Screen, cmd ParsedCmd, env *Env, succeeded bool,
	elapsed time.Duration, cmds []ParsedCmd, currCmdIdx int, sep string) {

	if isBootstrap && !env.GetBool("display.bootstrap") || !env.GetBool("display") {
		return
	}
	if checkPrintFilter(cmd, env, sep) {
		return
	}
	cmds, currCmdIdx = filterQuietCmds(env, cmds, currCmdIdx)
	if len(cmds) == 1 && !env.GetBool("display.one-cmd") {
		return
	}
	if len(cmds) == 0 {
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
	line := " " + getCmdPath(cmd, sep, printRealname)
	durStr := formatDuration(elapsed)
	if width - len(durStr) - 6 < 20 {
		width = len(durStr) + 6 + 20
	}
	line = "│" + " " + resStr + padRight(line, " ", width - len(durStr) - 6) + durStr + " " + "│"

	screen.Println("┌" + strings.Repeat("─", width-2) + "┐")
	screen.Println(line)
	/*
		for k, v := range argv {
			line = "│" + padRight(strings.Repeat(" ", len(resStr) + 4) + k + " = " + v.Raw,
				" ", width-3) + " " + "│"
			screen.Println(line)
		}
	*/
	screen.Println("└" + strings.Repeat("─", width-2) + "┘")

	if currCmdIdx >= len(cmds)-1 || !succeeded {
		screen.Println(strings.Repeat(" ", width-len(timeStr)-2) + timeStr)
	} else {
		screen.Println("\n")
	}
}

func printCmdStack(isBootstrap bool, screen *Screen, cmd ParsedCmd, env *Env,
	cmds []ParsedCmd, currCmdIdx int, sep string) {

	if isBootstrap && !env.GetBool("display.bootstrap") || !env.GetBool("display") {
		return
	}
	if checkPrintFilter(cmd, env, sep) {
		return
	}
	cmds, currCmdIdx = filterQuietCmds(env, cmds, currCmdIdx)
	if len(cmds) == 1 && !env.GetBool("display.one-cmd") {
		return
	}
	if len(cmds) == 0 {
		return
	}

	width := env.GetInt("display.width")

	const cmdCntKey = "display.max-cmd-cnt"
	cmdDisplayCnt := env.GetInt(cmdCntKey)
	if cmdDisplayCnt < 4 {
		panic(fmt.Errorf("[printCmdStack] %s should not less than 4", cmdCntKey))
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
	displayIdxEnd := len(cmds)
	if len(cmds) > cmdDisplayCnt {
		displayIdxStart = currCmdIdx - cmdDisplayCnt/2
		if displayIdxStart < 0 {
			displayIdxStart = 0
		}
		if displayIdxStart+cmdDisplayCnt > len(cmds) {
			displayIdxEnd = len(cmds)
			displayIdxStart = displayIdxEnd - cmdDisplayCnt
		} else {
			displayIdxEnd = displayIdxStart + cmdDisplayCnt
		}
	}

	screen.Println("┌" + strings.Repeat("─", titleWidth-2) + "┐")
	screen.Println(titleLine)
	screen.Println(topBorder)

	if printEnv {
		filterPrefixs := []string{strings.Join(cmd.Path(), sep) + sep}
		envLines := dumpEnv(env, printEnvLayer, printDefEnv, printRuntimeEnv, filterPrefixs)
		for _, line := range envLines {
			screen.Println("│" + padRight("    "+line, " ", width-2) + "│")
		}
		if len(envLines) != 0 {
			screen.Println("├" + strings.Repeat("─", width-2) + "┤")
		}
	}

	for i, cmd := range cmds {
		if i < displayIdxStart || i >= displayIdxEnd {
			continue
		}
		var line string
		if (i == displayIdxStart && i != 0) || (i+1 == displayIdxEnd && i+1 != len(cmds)) {
			line += "    ..."
		} else {
			if i == currCmdIdx {
				line += " >> "
			} else {
				line += "    "
			}
			line += getCmdPath(cmd, sep, printRealname)
		}
		screen.Println("│" + padRight(line, " ", width-2) + "│")

		args := cmd.Args()
		if args == nil {
			continue
		}
		argv := cmd.GenEnv(env.GetLayer(EnvLayerSession)).GetArgv(cmd.Path(), sep, cmd.Args())
		for _, line := range args.Dump(argv, false) {
			screen.Println("│" + padRight(strings.Repeat(" ", 8)+line, " ", width-2) + "│")
		}
	}

	screen.Println("└" + strings.Repeat("─", width-2) + "┘")
}

func checkPrintFilter(cmd ParsedCmd, env *Env, sep string) bool {
	if len(cmd) == 0 {
		return true
	}
	lastSeg := cmd[len(cmd)-1]
	if lastSeg.Cmd.Cmd == nil || lastSeg.Cmd.Cmd.cmd == nil ||
		(lastSeg.Cmd.Cmd.IsQuiet() && !env.GetBool("display.mod.quiet")) {
		return true
	}
	return false
}

func filterQuietCmds(env *Env, cmds []ParsedCmd, currCmdIdx int) ([]ParsedCmd, int) {
	if env.GetBool("display.mod.quiet") {
		return cmds, currCmdIdx
	}

	var newCmds []ParsedCmd
	newIdx := currCmdIdx
	for i, cmd := range cmds {
		if len(cmd) == 0 {
			continue
		}
		lastSeg := cmd[len(cmd)-1].Cmd
		if lastSeg.Cmd == nil || lastSeg.Cmd.cmd == nil || lastSeg.Cmd.IsQuiet() {
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
