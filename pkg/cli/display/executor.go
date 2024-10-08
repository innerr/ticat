package display

import (
	"strings"
	"time"

	"github.com/innerr/ticat/pkg/core/model"
	"github.com/innerr/ticat/pkg/utils"
)

type CmdStackLines struct {
	Display  bool
	Title    string
	TitleLen int
	Time     string
	TimeLen  int
	Stack    []string
	StackLen []int
	Env      []string
	EnvLen   []int
	Flow     []string
	FlowLen  []int
	Bg       []string
	BgLen    []int
}

func PrintCmdStack(
	isBootstrap bool,
	screen model.Screen,
	cmd model.ParsedCmd,
	mask *model.ExecuteMask,
	env *model.Env,
	envKeysInfo *model.EnvKeysInfo,
	flow []model.ParsedCmd,
	currCmdIdx int,
	strs *model.CmdTreeStrs,
	bgTasks *model.BgTasks,
	tailModeCall bool) (lines CmdStackLines) {

	if flow[currCmdIdx].LastCmdNode() != nil && flow[currCmdIdx].LastCmdNode().IsApi() {
		return
	}
	if tailModeCall {
		return
	}
	if isBootstrap && !env.GetBool("display.bootstrap") || !env.GetBool("display.executor") {
		return
	}
	if checkPrintFilter(cmd, env) {
		return
	}
	flow, currCmdIdx = filterQuietCmds(env, flow, currCmdIdx)
	stackDepth := env.GetInt("sys.stack-depth")
	if len(flow) == 1 && !env.GetBool("display.one-cmd") && stackDepth <= 1 {
		if !env.GetBool("sys.breakpoint.here.now") {
			return
		}
	}
	if len(flow) == 0 {
		return
	}

	env = env.Clone()
	lines.Display = true

	useUtf8 := env.GetBool("display.utf8.symbols")
	inBg := env.GetBool("sys.in-bg-task")
	if bgTasks != nil && !inBg {
		for _, bg := range bgTasks.GetStat() {
			line := ""
			lineLen := 0
			bgCmd := ColorCmdDelay(bg.Cmd, env) + ColorSymbol(" -- ", env) + bg.Tid
			bgCmdLen := len(bg.Cmd) + len(bg.Tid) + 4
			if bg.Finished {
				if bg.Err != nil {
					line += ColorError(" E ", env) + bgCmd
				} else {
					doneStr := "OK"
					if useUtf8 {
						doneStr = " ✓"
					}
					line += ColorCmd(doneStr+" ", env) + bgCmd
				}
				lineLen = 3 + bgCmdLen
			} else if bg.Started {
				line += ColorCmdCurr(">> ", env) + bgCmd
				lineLen = 3 + bgCmdLen
			} else {
				line += ColorExplain("zZ ", env) + ColorExplain(bg.Cmd, env) + ColorExplain(" - "+bg.Tid, env)
				lineLen = len(line) - ColorExtraLen(env, "explain")*3
			}
			lines.Bg = append(lines.Bg, line)
			lines.BgLen = append(lines.BgLen, lineLen)
		}
	}

	if env.GetBool("display.stack") {
		listSep := env.GetRaw("strs.list-sep")
		stack := strings.Split(env.GetRaw("sys.stack"), listSep)
		if len(stack) > 1 {
			for i, frame := range stack {
				line := rpt(" ", i*4+3) + ColorProp("+ ", env)
				extraLen := ColorExtraLen(env, "prop")
				if i == 0 {
					line += ColorExplain(frame, env)
					extraLen += ColorExtraLen(env, "explain")
				} else {
					line += ColorCmd(frame, env)
					extraLen += ColorExtraLen(env, "cmd")
				}
				lines.Stack = append(lines.Stack, line)
				lines.StackLen = append(lines.StackLen, len(line)-extraLen)
			}
		}
	}

	const cmdCntKey = "display.max-cmd-cnt"
	cmdDisplayCnt := env.GetInt(cmdCntKey)
	if cmdDisplayCnt < 4 {
		// panic(fmt.Errorf("[PrintCmdStack] %s should not less than 4", cmdCntKey))
		cmdDisplayCnt = 4
	}

	// TODO: show stack depth when no background tasks
	/*
		stackDepth := env.GetRaw("sys.stack-depth")
		if len(stackDepth) > 2 {
			stackDepth = "[..]"
		} else {
			stackDepth = "[" + stackDepth + "]"// + strings.Repeat(" ", 2+1-len(stackDepth))
		}
		lines.Title = "stack-level: " + stackDepth
		lines.TitleLen = len(lines.Title)
	*/
	lines.Title = ColorThread("thread: ", env) + utils.GoRoutineIdStr()
	lines.TitleLen = len(lines.Title) - ColorExtraLen(env, "thread")

	lines.Time = time.Now().Format("01-02 15:04:05")
	lines.TimeLen = len(lines.Time)

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

	printEnv := env.GetBool("display.env")
	printEnvLayer := env.GetBool("display.env.layer")
	printDefEnv := env.GetBool("display.env.default")
	printRuntimeEnv := env.GetBool("display.env.sys")
	printInputName := env.GetBool("display.mod.input-name")
	printRealname := env.GetBool("display.mod.input-name.with-realname")

	if printEnv {
		// TODO: 'session' 'strs' => config
		filterPrefixs := []string{
			"session",
			"strs" + strs.EnvPathSep,
			"display.utf8" + strs.EnvPathSep,
			strings.Join(cmd.Path(), strs.PathSep) + strs.PathSep,
		}
		if !env.GetBool("display.env.sys.paths") {
			filterPrefixs = append(filterPrefixs, "sys.paths")
		}
		if !env.GetBool("display.env.display") {
			filterPrefixs = append(filterPrefixs, "display.")
		}
		// TODO: the extra color char len is ugly, handle it better
		envLines, envLinesExtraLens := dumpEnv(env, envKeysInfo, printEnvLayer, printDefEnv,
			printRuntimeEnv, false, filterPrefixs, 4)
		for i, line := range envLines {
			line := "   " + line
			extraLen := envLinesExtraLens[i]
			lines.Env = append(lines.Env, line)
			lines.EnvLen = append(lines.EnvLen, len(line)-extraLen)
		}
	}

	for i, cmd := range flow {
		if i < displayIdxStart || i >= displayIdxEnd {
			continue
		}
		cmdEnv, argv := cmd.ApplyMappingGenEnvAndArgv(env.GetLayer(model.EnvLayerSession),
			strs.EnvValDelAllMark, strs.PathSep, stackDepth)
		sysArgv := cmdEnv.GetSysArgv(cmd.Path(), strs.PathSep)
		var name string
		if !printInputName && cmd.LastCmdNode() != nil {
			name = cmd.LastCmdNode().DisplayPath()
		} else {
			name = cmd.DisplayMatchedPath(strs.PathSep, printRealname)
		}
		var line string
		lineExtraLen := 0
		endOmitting := (i+1 == displayIdxEnd && i+1 != len(flow))
		if (i == displayIdxStart && i != 0) || endOmitting {
			line += "   ..."
		} else {
			if i == currCmdIdx {
				if sysArgv.IsDelay() && !inBg {
					line += ColorCmdCurr(">> "+name+" (schedule to bg in ", env) + sysArgv.GetDelayStr() + ColorCmdCurr(")", env)
					lineExtraLen += ColorExtraLen(env, "cmd-curr", "cmd-curr")
				} else {
					line += ColorCmdCurr(">> "+name, env)
					lineExtraLen += ColorExtraLen(env, "cmd-curr")
				}
				if mask != nil {
					resultStr := string(mask.ResultIfExecuted)
					if mask.ResultIfExecuted == model.ExecutedResultError {
						line += ColorExplain(" - executed: ", env) + ColorError(resultStr, env)
						lineExtraLen += ColorExtraLen(env, "explain", "error")
					} else if mask.ResultIfExecuted == model.ExecutedResultSucceeded {
						line += ColorExplain(" - executed: ", env) + ColorCmdDone(resultStr, env)
						lineExtraLen += ColorExtraLen(env, "explain", "cmd-done")
					} else if mask.ResultIfExecuted == model.ExecutedResultSkipped {
						line += ColorExplain(" - executed: ", env) + ColorExplain(resultStr, env)
						lineExtraLen += ColorExtraLen(env, "explain", "explain")
					} else if mask.ResultIfExecuted != model.ExecutedResultUnRun || mask.ResultIfExecuted == model.ExecutedResultIncompleted {
						line += ColorExplain(" - executed: ", env) + ColorHighLight(resultStr, env)
						lineExtraLen += ColorExtraLen(env, "explain", "highlight")
					}
					if mask.ResultIfExecuted != model.ExecutedResultSkipped && mask.ResultIfExecuted != model.ExecutedResultUnRun {
						durStr, durExtraLen := executedCmdDurStr(mask.ExecutedCmd, false, env)
						line += " " + durStr
						lineExtraLen += durExtraLen
					}
				}
			} else if i < currCmdIdx {
				if sysArgv.IsDelay() && !inBg {
					line += "   " + ColorCmdDelay(name+" (scheduled to bg in ", env) + sysArgv.GetDelayStr() + ColorCmdDelay(")", env)
					lineExtraLen += ColorExtraLen(env, "cmd-delay", "cmd-delay")
				} else {
					line += "   " + ColorCmdDone(name, env)
					lineExtraLen += ColorExtraLen(env, "cmd-done")
				}
			} else {
				if sysArgv.IsDelay() && !inBg {
					line += "   " + name + " (schedule to bg in " + sysArgv.GetDelayStr() + ")"
				} else {
					line += "   " + name
				}
			}
		}

		lines.Flow = append(lines.Flow, line)
		lines.FlowLen = append(lines.FlowLen, len(line)-lineExtraLen)
		if endOmitting {
			continue
		}
		args := cmd.Args()
		// TODO: use DumpEffectedArgs instead of DumpProvidedArgs
		colorizeArg := i <= currCmdIdx
		for _, line := range DumpProvidedArgs(env, &args, argv, colorizeArg) {
			line := strings.Repeat(" ", 3+4) + line
			extraLen := 0
			if colorizeArg {
				extraLen += ColorExtraLen(env, "arg", "symbol")
			}
			lines.Flow = append(lines.Flow, line)
			lines.FlowLen = append(lines.FlowLen, len(line)-extraLen)
		}
		//for _, line := range DumpSysArgs(env, sysArgv, colorizeArg) {
		//	line = strings.Repeat(" ", 3+4) + line
		//	extraLen := 0
		//	if colorizeArg {
		//		extraLen += ColorExtraLen(env, "explain", "arg", "symbol")
		//	}
		//	lines.Flow = append(lines.Flow, line)
		//	lines.FlowLen = append(lines.FlowLen, len(line)-extraLen)
		//}

		cic := cmd.LastCmd()
		if cic != nil && !sysArgv.IsDelay() && cic.HasSubFlow(false) && (mask == nil || mask.SubFlow != nil) {
			if i < currCmdIdx || i == currCmdIdx {
				line := ColorFlowing("       --->>>", env)
				lines.Flow = append(lines.Flow, line)
				lines.FlowLen = append(lines.FlowLen, len(line)-ColorExtraLen(env, "flowing"))
			}
			if i < currCmdIdx {
				line := ColorFlowing("       <<<---", env)
				lines.Flow = append(lines.Flow, line)
				lines.FlowLen = append(lines.FlowLen, len(line)-ColorExtraLen(env, "flowing"))
			}
		}
	}

	return
}

type CmdResultLines struct {
	Display   bool
	Cmd       string
	CmdLen    int
	Res       string
	ResLen    int
	Dur       string
	DurLen    int
	Footer    string
	FooterLen int
}

func PrintCmdResult(
	cc *model.Cli,
	isBootstrap bool,
	screen model.Screen,
	cmd model.ParsedCmd,
	env *model.Env,
	succeeded bool,
	elapsed time.Duration,
	flow []model.ParsedCmd,
	currCmdIdx int,
	strs *model.CmdTreeStrs) (lines CmdResultLines) {

	if isBootstrap && !env.GetBool("display.bootstrap") || !env.GetBool("display.executor") {
		return
	}
	if checkPrintFilter(cmd, env) {
		return
	}

	betweenFileNFlow := (currCmdIdx == len(flow)-1) && callerIsFileNFlow(cc, env)

	if !env.GetBool("display.executor.end") && !betweenFileNFlow {
		return
	}
	flow, currCmdIdx = filterQuietCmds(env, flow, currCmdIdx)
	if len(flow) == 1 && !env.GetBool("display.one-cmd") && !betweenFileNFlow {
		return
	}
	if len(flow) == 0 {
		return
	}

	lines.Display = true
	lines.Dur = formatDuration(elapsed)
	lines.DurLen = len(lines.Dur)

	if env.GetBool("display.utf8.symbols") {
		if succeeded {
			lines.Res = " ✓"
		} else {
			lines.Res = " ✘"
		}
	} else {
		if succeeded {
			lines.Res = "OK"
		} else {
			lines.Res = " E"
		}
	}
	lines.ResLen = 2

	if succeeded {
		lines.Res = ColorCmd(lines.Res, env)
	} else {
		lines.Res = ColorError(lines.Res, env)
	}

	lines.Cmd = ColorCmdDone(cmd.DisplayPath(strs.PathSep, env.GetBool("display.mod.realname")), env)
	lines.CmdLen = len(lines.Cmd) - ColorExtraLen(env, "cmd-done")

	if currCmdIdx >= len(flow)-1 || !succeeded {
		lines.Footer = time.Now().Format("01-02 15:04:05")
	} else {
		lines.Footer = ""
	}
	lines.FooterLen = len(lines.Footer)
	return
}

func checkPrintFilter(cmd model.ParsedCmd, env *model.Env) bool {
	if cmd.IsEmpty() {
		return true
	}
	last := cmd.LastCmd()
	if last == nil {
		return true
	}
	if last.IsQuiet() && !env.GetBool("display.mod.quiet") {
		return true
	}
	return false
}

func filterQuietCmds(env *model.Env, flow []model.ParsedCmd, currCmdIdx int) ([]model.ParsedCmd, int) {
	if env.GetBool("display.mod.quiet") {
		return flow, currCmdIdx
	}

	var newCmds []model.ParsedCmd
	newIdx := currCmdIdx
	for i, cmd := range flow {
		if cmd.IsEmpty() {
			continue
		}
		last := cmd.LastCmd()
		if last == nil || last.IsQuiet() {
			if i < currCmdIdx {
				newIdx -= 1
			}
			continue
		}
		newCmds = append(newCmds, cmd)
	}
	return newCmds, newIdx
}

func callerIsFileNFlow(cc *model.Cli, env *model.Env) bool {
	listSep := env.GetRaw("strs.list-sep")
	stack := strings.Split(env.GetRaw("sys.stack"), listSep)
	if len(stack) <= 1 {
		return false
	}
	stack = stack[1:]
	callerName := stack[len(stack)-1]
	callerCmd, ok := cc.ParseCmd(false, callerName)
	if !ok {
		return false
	}
	last := callerCmd.LastCmd()
	return last.Type() == model.CmdTypeFileNFlow
}
