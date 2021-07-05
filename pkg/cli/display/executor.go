package display

import (
	"strings"
	"time"

	"github.com/pingcap/ticat/pkg/cli/core"
)

type CmdStackLines struct {
	Display       bool
	StackDepth    string
	StackDepthLen int
	Time          string
	TimeLen       int
	Env           []string
	EnvLen        []int
	Flow          []string
	FlowLen       []int
}

func PrintCmdStack(
	isBootstrap bool,
	screen core.Screen,
	cmd core.ParsedCmd,
	env *core.Env,
	flow []core.ParsedCmd,
	currCmdIdx int,
	strs *core.CmdTreeStrs) (lines CmdStackLines) {

	env = env.Clone()

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

	lines.Display = true

	const cmdCntKey = "display.max-cmd-cnt"
	cmdDisplayCnt := env.GetInt(cmdCntKey)
	if cmdDisplayCnt < 4 {
		// panic(fmt.Errorf("[PrintCmdStack] %s should not less than 4", cmdCntKey))
		cmdDisplayCnt = 4
	}

	stackDepth := env.Get("sys.stack-depth").Raw
	if len(stackDepth) > 2 {
		stackDepth = "[..]"
	} else {
		stackDepth = "[" + stackDepth + "]" + strings.Repeat(" ", 2+1-len(stackDepth))
	}
	lines.StackDepth = "stack-level: " + stackDepth
	lines.StackDepthLen = len(lines.StackDepth)

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
	printRealname := env.GetBool("display.mod.realname")

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
		envLines := dumpEnv(env, printEnvLayer, printDefEnv, printRuntimeEnv, false, filterPrefixs, 4)
		for _, line := range envLines {
			line := "   " + line
			lines.Env = append(lines.Env, line)
			lines.EnvLen = append(lines.EnvLen, len(line))
		}
	}

	for i, cmd := range flow {
		if i < displayIdxStart || i >= displayIdxEnd {
			continue
		}
		var line string
		endOmitting := (i+1 == displayIdxEnd && i+1 != len(flow))
		if (i == displayIdxStart && i != 0) || endOmitting {
			line += "   ..."
		} else {
			if i == currCmdIdx {
				line += ">> "
			} else {
				line += "   "
			}
			line += cmd.DisplayPath(strs.PathSep, printRealname)
		}
		lines.Flow = append(lines.Flow, line)
		lines.FlowLen = append(lines.FlowLen, len(line))
		if endOmitting {
			continue
		}
		_, argv := cmd.ApplyMappingGenEnvAndArgv(
			env.GetLayer(core.EnvLayerSession), strs.EnvValDelAllMark, strs.PathSep)
		args := cmd.Args()
		for _, line := range DumpArgs(&args, argv, false) {
			line := strings.Repeat(" ", 3+4) + line
			lines.Flow = append(lines.Flow, line)
			lines.FlowLen = append(lines.FlowLen, len(line))
		}

		cic := cmd.LastCmd()
		if cic != nil && cic.Type() == core.CmdTypeFlow {
			if i+1 == currCmdIdx || i == currCmdIdx {
				line := "       --->>>"
				lines.Flow = append(lines.Flow, line)
				lines.FlowLen = append(lines.FlowLen, len(line))
			}
			if i+1 == currCmdIdx {
				line := "       <<<---"
				lines.Flow = append(lines.Flow, line)
				lines.FlowLen = append(lines.FlowLen, len(line))
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
	isBootstrap bool,
	screen core.Screen,
	cmd core.ParsedCmd,
	env *core.Env,
	succeeded bool,
	elapsed time.Duration,
	flow []core.ParsedCmd,
	currCmdIdx int,
	strs *core.CmdTreeStrs) (lines CmdResultLines) {

	if isBootstrap && !env.GetBool("display.bootstrap") ||
		!env.GetBool("display.executor") || !env.GetBool("display.executor.end") {
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

	lines.Display = true
	lines.Dur = formatDuration(elapsed)
	lines.DurLen = len(lines.Dur)

	useUtf8 := env.GetBool("display.utf8.symbols")
	if useUtf8 {
		if succeeded {
			lines.Res = " ✓"
		} else {
			lines.Res = " ✘"
		}
	} else {
		if succeeded {
			lines.Res = "OK"
		} else {
			lines.Res = "EE"
		}
	}
	lines.ResLen = 2

	lines.Cmd = cmd.DisplayPath(strs.PathSep, env.GetBool("display.mod.realname"))
	lines.CmdLen = len(lines.Cmd)

	if currCmdIdx >= len(flow)-1 || !succeeded {
		lines.Footer = time.Now().Format("01-02 15:04:05")
	} else {
		lines.Footer = ""
	}
	lines.FooterLen = len(lines.Footer)
	return
}

func checkPrintFilter(cmd core.ParsedCmd, env *core.Env) bool {
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

func filterQuietCmds(env *core.Env, flow []core.ParsedCmd, currCmdIdx int) ([]core.ParsedCmd, int) {
	if env.GetBool("display.mod.quiet") {
		return flow, currCmdIdx
	}

	var newCmds []core.ParsedCmd
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
