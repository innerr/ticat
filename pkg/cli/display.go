package cli

import (
	"fmt"
	"time"
	"strings"
)

func printCmdResult(screen *Screen, cmd ParsedCmd, env *Env, succeeded bool, cmds []ParsedCmd, currCmdIdx int, sep string) {
	if !env.Get("runtime.display").GetBool() {
		return
	}
	if checkPrintFilter(cmd, env, sep) {
		return
	}
	cmds, currCmdIdx = filterBuiltins(env, cmds, currCmdIdx)
	if len(cmds) == 0 {
		return
	}

	const widthKey = "runtime.display.width"
	width := env.Get(widthKey).GetInt()

	var resStr string
	if succeeded {
		resStr = " ✓" //"OK"
	} else {
		resStr = " ✘" //"EE"
	}

	timeStrOrigin := time.Now().Format("01-02 15:04:05")

	var timeStr string
	if currCmdIdx >= len(cmds) {
		timeStr = timeStrOrigin + " "
	}

	printRealname := env.Get("runtime.display.mod.realname").GetBool()
	line := " " + getCmdPath(cmd, sep, printRealname)
	line = "│" + " " + resStr + padRight(line, " ", width - len(timeStr) - 2 - 3) + timeStr + "│"

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
	if !env.Get("runtime.display").GetBool() {
		return
	}
	if checkPrintFilter(cmd, env, sep) {
		return
	}
	cmds, currCmdIdx = filterBuiltins(env, cmds, currCmdIdx)
	if len(cmds) == 0 {
		return
	}

	width := env.Get("runtime.display.width").GetInt()

	const cmdCntKey = "runtime.display.max-cmd-cnt"
	cmdDisplayCnt := env.Get(cmdCntKey).GetInt()
	if cmdDisplayCnt < 4 {
		panic(fmt.Errorf("%s should not less than 4", cmdCntKey))
	}

	printEnv := env.Get("runtime.display.env").GetBool()
	printEnvLayer := env.Get("runtime.display.env.layer").GetBool()
	printDefEnv := env.Get("runtime.display.env.default").GetBool()
	printRuntimeEnv := env.Get("runtime.display.env.runtime.sys").GetBool()

	printRealname := env.Get("runtime.display.mod.realname").GetBool()

	stackDepth := env.Get("runtime.stack-depth").Raw
	if len(stackDepth) > 3 {
		stackDepth = "..."
	} else {
		stackDepth = stackDepth + strings.Repeat(" ", 3 + 1 - len(stackDepth))
	}
	stackDepth = " stack-level: " + stackDepth + " "
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
	if width < 1 + titleLineLen + len(timeStr) + 1  {
		width = 1 + titleLineLen + len(timeStr) + 1
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

	if printEnv {
		envLines := dumpEnv(env, printEnvLayer, printDefEnv, printRuntimeEnv)
		for _, line := range envLines {
			screen.Println("│" + padRight("    " + line, " ", width - 2) + "│")
		}
		if len(envLines) != 0 {
			screen.Println("├" + strings.Repeat("─", width - 2) + "┤")
		}
	}

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
			line += getCmdPath(cmd, sep, printRealname)
		}
		screen.Println("│" + padRight(line, " ", width - 2) + "│")
	}

	screen.Println("└" + strings.Repeat("─", width - 2) + "┘")
}

func checkPrintFilter(cmd ParsedCmd, env *Env, sep string) bool {
	var cmdFirstSegName string
	if len(cmd) != 0 {
		if cmd[len(cmd)-1].Cmd.Cmd != nil && cmd[len(cmd)-1].Cmd.Cmd.IsQuiet() {
			return true
		}
		cmdFirstSegName = cmd[0].Cmd.Name
	}
	if cmdFirstSegName == "builtin" && !env.Get("runtime.display.mod.builtin").GetBool() {
		return true
	}
	return false
}

func filterBuiltins(env *Env, cmds []ParsedCmd, currCmdIdx int) ([]ParsedCmd, int) {
	if env.Get("runtime.display.mod.builtin").GetBool() {
		return cmds, currCmdIdx
	}

	var newCmds []ParsedCmd
	newIdx := currCmdIdx
	for i, cmd := range cmds {
		if len(cmd) != 0 && cmd[0].Cmd.Name == "builtin" {
			if i < currCmdIdx {
				newIdx -= 1
			}
		} else {
			newCmds = append(newCmds, cmd)
		}
	}
	return newCmds, newIdx
}

func dumpEnv(env *Env, printEnvLayer bool, printDefEnv bool, printRuntimeEnv bool) (res []string) {
	var filterPrefix string
	if !printRuntimeEnv {
		filterPrefix = "runtime.sys."
	}
	if !printEnvLayer {
		compacted := env.Compact(printDefEnv, filterPrefix)
		for k, v := range compacted {
			res = append(res, k + " = " + v)
		}
	} else {
		dumpEnvLayer(env, printEnvLayer, printDefEnv, filterPrefix, &res, 0)
	}
	return
}

func dumpEnvLayer(env *Env, printEnvLayer bool, printDefEnv bool, filterPrefix string, res *[]string, depth int) {
	if env.Type == EnvLayerDefault && !printDefEnv {
		return
	}
	var output []string
	indent := strings.Repeat(" ", depth * 4)
	for k, v := range env.Pairs {
		if len(filterPrefix) != 0 && strings.HasPrefix(k, filterPrefix) {
			continue
		}
		output = append(output, indent + "- " + k + " = " + v.Raw)
	}
	if env.Parent != nil {
		dumpEnvLayer(env.Parent, printEnvLayer, printDefEnv, filterPrefix, &output, depth + 1)
	}
	if len(output) != 0 {
		*res = append(*res, indent + "[" + EnvLayerName(env.Type) + "]")
		*res = append(*res, output...)
	}
}

func padRight(str string, pad string, width int) string {
	if len(str) >= width {
		return str
	}
	return str + strings.Repeat(pad, width - len(str))
}

func getCmdPath(cmd ParsedCmd, sep string, printRealname bool) string {
	var path []string
	for _, seg := range cmd {
		if seg.Cmd.Cmd != nil {
			name := seg.Cmd.Name
			if printRealname {
				name += "(=" + seg.Cmd.Cmd.Name() + ")"
			}
			path = append(path, name)
		}
	}
	return strings.Join(path, sep)
}
