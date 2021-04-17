package cli

import (
	"fmt"
	"time"
	"strings"
)

func printCmdResult(screen *Screen, cmd ParsedCmd, env *Env, succeeded bool, cmds []ParsedCmd, currCmdIdx int, sep string) {
	const widthKey = "runtime.display.executor.width"
	widthVal := env.Get(widthKey)
	if len(widthVal.Raw) == 0 {
		panic(fmt.Errorf("%s not found in env", widthKey))
	}
	width := widthVal.GetInt()

	var resStr string
	if succeeded {
		resStr = "OK"
	} else {
		resStr = "EE"
	}

	timeStrOrigin := time.Now().Format("01-02 15:04:05")

	var timeStr string
	if currCmdIdx >= len(cmds) {
		timeStr = timeStrOrigin + " "
	}

	line := " " + resStr + " " + cmdDisplayPath(cmd, sep)
	line = "│" + padRight(line, " ", width - len(timeStr) - 2) + timeStr + "│"

	screen.Println("┌" + strings.Repeat("─", width - 2) + "┐")
	screen.Println(line)
	screen.Println("└" + strings.Repeat("─", width - 2) + "┘")

	if currCmdIdx >= len(cmds) - 1 || !succeeded {
		screen.Println(strings.Repeat(" ", width - len(timeStrOrigin) - 2) + timeStrOrigin)
	} else {
		screen.Println("")
	}
}

func printCmdStack(screen *Screen, cmd ParsedCmd, env *Env, cmds []ParsedCmd, currCmdIdx int, sep string) {
	const widthKey = "runtime.display.executor.width"
	widthVal := env.Get(widthKey)
	if len(widthVal.Raw) == 0 {
		return
	}
	width := widthVal.GetInt()

	const cmdCntKey = "runtime.display.executor.max-cmd-cnt"
	cmdDisplayCnt := env.Get(cmdCntKey).GetInt()
	if cmdDisplayCnt < 4 {
		panic(fmt.Errorf("%s should not less than 4", cmdCntKey))
	}

	stackDepth := env.Get("runtime.stack-depth").Raw
	if len(stackDepth) > 3 {
		stackDepth = "[...]"
	} else {
		stackDepth = "[" + stackDepth + "]" + strings.Repeat(" ", 3 + 1 - len(stackDepth))
	}
	stackDepth = " depth " + stackDepth + " "
	titleLine := "│" + stackDepth + "│"

	// Notice that len(str) will return wrong size since we are using non-ascii chars
	titleWidth := len(stackDepth) + 2

	titleLine += "   (=`ω´=)   "
	titleLineLen := titleWidth + 3 + 7 + 3

	timeStr := time.Now().Format("01-02 15:04:05")
	// This "1" is for left border
	if width < titleLineLen + len(timeStr) + 1 {
		width = titleLineLen + len(timeStr) + 1
	}
	titleLine += strings.Repeat(" ", width - 1 - titleLineLen - len(timeStr) - 1) + timeStr

	topBorder := "├" + strings.Repeat("─", titleWidth - 2) + "┴"
	topBorder = topBorder + strings.Repeat("─", width - 1 - titleWidth) + "┐"

	displayIdxStart := 0
	displayIdxEnd := len(cmds)
	if len(cmds) > cmdDisplayCnt {
		displayIdxStart = currCmdIdx - cmdDisplayCnt / 2
		if displayIdxStart < 0 {
			displayIdxStart = 0
		}
		if displayIdxStart + cmdDisplayCnt > len(cmds) {
			displayIdxEnd = len(cmds)
			displayIdxStart = displayIdxEnd - cmdDisplayCnt
		} else {
			displayIdxEnd = displayIdxStart + cmdDisplayCnt
		}
	}

	screen.Println("┌" + strings.Repeat("─", titleWidth - 2) + "┐")
	screen.Println(titleLine)
	screen.Println(topBorder)
	for i, cmd := range cmds {
		if i < displayIdxStart || i >= displayIdxEnd {
			continue
		}
		var line string
		if (i == displayIdxStart && i != 0) || (i + 1 == displayIdxEnd && i + 1 != len(cmds)) {
			line += "    ..."
		} else {
			if i == currCmdIdx {
				line += " >> "
			} else {
				line += "    "
			}
			line += cmdDisplayPath(cmd, sep)
		}
		screen.Println("│" + padRight(line, " ", width - 2) + "│")
	}
	screen.Println("└" + strings.Repeat("─", width - 2) + "┘")
}

func padRight(str string, pad string, width int) string {
	if len(str) >= width {
		return str
	}
	return str + strings.Repeat(pad, width - len(str))
}

func cmdDisplayPath(cmd ParsedCmd, sep string) string {
	var path []string
	for _, seg := range cmd {
		if seg.Cmd.Cmd != nil {
			path = append(path, seg.Cmd.Name)
		}
	}
	return strings.Join(path, sep)
}
