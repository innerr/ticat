package cli

import (
	"fmt"
	"strings"
	"time"
)

func printCmdResult(screen *Screen, cmd ParsedCmd, argv ArgVals, env *Env, succeeded bool,
	elapsed time.Duration, cmds []ParsedCmd, currCmdIdx int, sep string) {

	if !env.Get("runtime.display").GetBool() {
		return
	}
	if checkPrintFilter(cmd, env, sep) {
		return
	}
	cmds, currCmdIdx = filterBuiltinAndQuiet(env, cmds, currCmdIdx)
	if len(cmds) == 1 && !env.Get("runtime.display.one-cmd").GetBool() {
		return
	}
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

	timeStr := time.Now().Format("01-02 15:04:05")

	printRealname := env.Get("runtime.display.mod.realname").GetBool()
	line := " " + getCmdPath(cmd, sep, printRealname)
	durStr := formatDuration(elapsed)
	line = "│" + " " + resStr + padRight(line, " ", width-len(durStr)-2-2-2) + durStr + " " + "│"

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

func printCmdStack(screen *Screen, cmd ParsedCmd, argv ArgVals, env *Env,
	cmds []ParsedCmd, currCmdIdx int, sep string) {

	if !env.Get("runtime.display").GetBool() {
		return
	}
	if checkPrintFilter(cmd, env, sep) {
		return
	}
	cmds, currCmdIdx = filterBuiltinAndQuiet(env, cmds, currCmdIdx)
	if len(cmds) == 1 && !env.Get("runtime.display.one-cmd").GetBool() {
		return
	}
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
	if len(stackDepth) > 2 {
		stackDepth = ".."
	} else {
		stackDepth = stackDepth + strings.Repeat(" ", 2+1-len(stackDepth))
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

		argv := cmd.GenEnv(env).GetArgv(cmd.Path(), sep, cmd.Args())
		for k, v := range argv {
			line = "│" + padRight(strings.Repeat(" ", 8)+k+" = "+v.Raw,
				" ", width-3) + " " + "│"
			screen.Println(line)
		}
	}

	screen.Println("└" + strings.Repeat("─", width-2) + "┘")
}

func checkPrintFilter(cmd ParsedCmd, env *Env, sep string) bool {
	if len(cmd) == 0 {
		return true
	}
	lastSeg := cmd[len(cmd)-1]
	if lastSeg.Cmd.Cmd == nil || lastSeg.Cmd.Cmd.cmd == nil || lastSeg.Cmd.Cmd.IsQuiet() {
		return true
	}
	cmdFirstSegName := cmd[0].Cmd.Name
	if cmdFirstSegName == "builtin" && !env.Get("runtime.display.mod.builtin").GetBool() {
		return true
	}
	return false
}

func filterBuiltinAndQuiet(env *Env, cmds []ParsedCmd, currCmdIdx int) ([]ParsedCmd, int) {
	if env.Get("runtime.display.mod.builtin").GetBool() {
		return cmds, currCmdIdx
	}

	var newCmds []ParsedCmd
	newIdx := currCmdIdx
	for i, cmd := range cmds {
		if len(cmd) == 0 {
			continue
		}
		lastSeg := cmd[len(cmd)-1].Cmd
		if cmd[0].Cmd.Name == "builtin" || lastSeg.Cmd == nil || lastSeg.Cmd.cmd == nil || lastSeg.Cmd.IsQuiet() {
			if i < currCmdIdx {
				newIdx -= 1
			}
			continue
		}
		newCmds = append(newCmds, cmd)
	}
	return newCmds, newIdx
}

func debugDumpEnv(env *Env) {
	fmt.Println()
	fmt.Println("-- DBUG BEGIN --")
	lines := dumpEnv(env, true, true, true, nil)
	for _, line := range lines {
		fmt.Println(line)
	}
	fmt.Println("-- DBUG END --")
	fmt.Println()
}

func dumpEnv(env *Env, printEnvLayer bool, printDefEnv bool,
	printRuntimeEnv bool, filterPrefixs []string) (res []string) {

	if !printRuntimeEnv {
		filterPrefixs = append(filterPrefixs, EnvRuntimeSysPrefix)
	}
	if !printEnvLayer {
		compacted := env.Compact(printDefEnv, filterPrefixs)
		for k, v := range compacted {
			res = append(res, k+" = "+v)
		}
	} else {
		dumpEnvLayer(env, printEnvLayer, printDefEnv, filterPrefixs, &res, 0)
	}
	return
}

func dumpEnvLayer(env *Env, printEnvLayer bool, printDefEnv bool, filterPrefixs []string, res *[]string, depth int) {
	if env.tp == EnvLayerDefault && !printDefEnv {
		return
	}
	var output []string
	indent := strings.Repeat(" ", depth*4)
	for k, v := range env.pairs {
		filtered := false
		for _, filterPrefix := range filterPrefixs {
			if len(filterPrefix) != 0 && strings.HasPrefix(k, filterPrefix) {
				filtered = true
				break
			}
		}
		if !filtered {
			output = append(output, indent+"- "+k+" = "+v.Raw)
		}
	}
	if env.parent != nil {
		dumpEnvLayer(env.parent, printEnvLayer, printDefEnv, filterPrefixs, &output, depth+1)
	}
	if len(output) != 0 {
		*res = append(*res, indent+"["+EnvLayerName(env.tp)+"]")
		*res = append(*res, output...)
	}
}

func padRight(str string, pad string, width int) string {
	if len(str) >= width {
		return str
	}
	return str + strings.Repeat(pad, width-len(str))
}

func getCmdPath(cmd ParsedCmd, sep string, printRealname bool) string {
	var path []string
	for _, seg := range cmd {
		if seg.Cmd.Cmd != nil {
			name := seg.Cmd.Name
			realname := seg.Cmd.Cmd.Name()
			if printRealname && name != realname {
				name += "(=" + realname + ")"
			}
			path = append(path, name)
		}
	}
	return strings.Join(path, sep)
}

func formatDuration(dur time.Duration) string {
	return strings.ReplaceAll(fmt.Sprintf("%s", dur), "µ", "u")
}
