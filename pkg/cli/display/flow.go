package display

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/utils"
)

func DumpFlow(
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	fromCmdIdx int,
	args *DumpFlowArgs,
	envOpCmds []core.EnvOpCmd) {

	DumpFlowEx(cc, env, flow, fromCmdIdx, args, nil, false, envOpCmds)
}

// 'parsedGlobalEnv' + env in 'flow' = all env
func DumpFlowEx(
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	fromCmdIdx int,
	args *DumpFlowArgs,
	executedFlow *core.ExecutedFlow,
	running bool,
	envOpCmds []core.EnvOpCmd) {

	if len(flow.Cmds) == 0 {
		return
	}
	if args.MaxDepth <= 0 {
		args.MaxDepth = math.MaxInt64
	}
	if args.MaxTrivial <= 0 {
		args.MaxTrivial = math.MaxInt64
	}

	// Do not fold the only one cmd
	if len(flow.Cmds[fromCmdIdx:]) == 1 && getCmdTrivial(flow.Cmds[fromCmdIdx]) > 0 {
		if args.MaxDepth == 1 {
			args.MaxDepth = 2
		}
		if args.MaxTrivial == 1 {
			args.MaxTrivial = 2
		}
	}

	// The env will be modified during dumping (so it could show the real value)
	// so we need to clone the env to protect it
	env = env.Clone()

	writtenKeys := FlowWrittenKeys{}

	if executedFlow != nil {
		if args.MonitorMode {
			title := ColorTip("["+executedFlow.DirName+"]", env)
			ip := env.GetRaw("sys.session.id.ip")
			if len(ip) != 0 {
				title += ColorExplain(" - "+ip, env)
			}
			cc.Screen.Print(title + "\n")
		} else {
			PrintTipTitle(cc.Screen, env, "session-id ["+executedFlow.DirName+"], flow executed status:")
		}
	} else {
		PrintTipTitle(cc.Screen, env, "flow executing description:")
	}

	// TODO: show executed status and duration here

	cc.Screen.Print(ColorFlowing("--->>>", env) + "\n")
	ok := dumpFlow(cc, env, envOpCmds, flow, fromCmdIdx, args, executedFlow, running,
		false, writtenKeys, args.MaxDepth, args.MaxTrivial, 0, false)
	if ok {
		cc.Screen.Print(ColorFlowing("<<<---", env) + "\n")
	}
}

func dumpFlow(
	cc *core.Cli,
	env *core.Env,
	envOpCmds []core.EnvOpCmd,
	flow *core.ParsedCmds,
	fromCmdIdx int,
	args *DumpFlowArgs,
	executedFlow *core.ExecutedFlow,
	running bool,
	parentInBg bool,
	writtenKeys FlowWrittenKeys,
	maxDepth int,
	maxTrivial int,
	depth int,
	parentIncompleted bool) bool {

	sep := cc.Cmds.Strs.PathSep

	metFlows := map[string]bool{}

	for i, cmd := range flow.Cmds[fromCmdIdx:] {
		if cmd.IsEmpty() {
			continue
		}
		var executedCmd *core.ExecutedCmd
		if executedFlow != nil {
			if i < len(executedFlow.Cmds) {
				executedCmd = executedFlow.GetCmd(i)
			} else {
				// TODO: better display
				name := strings.Join(cmd.Path(), sep)
				executedCmd = core.NewExecutedCmd(name)
				executedCmd.Result = core.ExecutedResultUnRun
			}
		}

		cmdInBg, ok := dumpFlowCmd(cc, cc.Screen, env, envOpCmds, flow, fromCmdIdx+i, args, executedCmd, running,
			parentInBg, maxDepth, maxTrivial, depth, metFlows, writtenKeys, parentIncompleted)
		if !ok {
			if parentInBg {
				return false
			}
			if !cmdInBg {
				return false
			}
		}
	}
	return true
}

func dumpFlowCmd(
	cc *core.Cli,
	screen core.Screen,
	env *core.Env,
	envOpCmds []core.EnvOpCmd,
	flow *core.ParsedCmds,
	currCmdIdx int,
	args *DumpFlowArgs,
	executedCmd *core.ExecutedCmd,
	running bool,
	parentInBg bool,
	maxDepth int,
	maxTrivial int,
	depth int,
	metFlows map[string]bool,
	writtenKeys FlowWrittenKeys,
	parentIncompleted bool) (cmdInBg bool, ok bool) {

	parsedCmd := flow.Cmds[currCmdIdx]
	cmd := parsedCmd.Last().Matched.Cmd
	if cmd == nil {
		return false, true
	}
	cic := cmd.Cmd()
	if cic == nil {
		return false, true
	}
	if cmd.IsQuiet() && !cmd.HasSub() && !env.GetBool("display.mod.quiet") {
		return false, true
	}

	// TODO: this is slow
	originEnv := env.Clone()
	cmdEnv, argv := parsedCmd.ApplyMappingGenEnvAndArgv(env, cc.Cmds.Strs.EnvValDelAllMark, cmd.Strs.PathSep, depth+1)

	padLenCal := func(indentLvl int) int {
		indentLvl += depth * 2
		return args.IndentSize * indentLvl
	}

	prt := func(indentLvl int, msg string) {
		padding := ""
		if !args.MonitorMode {
			padding = rpt(" ", padLenCal(indentLvl))
			msg = autoPadNewLine(padding, msg)
		}
		screen.Print(padding + msg + "\n")
	}

	trivialDelta := getCmdTrivial(parsedCmd)

	cmdFailed := func() bool {
		return executedCmd != nil && executedCmd.Result != core.ExecutedResultSucceeded &&
			executedCmd.Result != core.ExecutedResultSkipped
	}
	cmdSkipped := func() bool {
		return executedCmd != nil && executedCmd.Result == core.ExecutedResultSkipped
	}

	notFold := func() bool {
		return parentIncompleted || cmdFailed() || maxTrivial > 0 && maxDepth > 0
	}

	foldSubFlowByTrivial := func() bool {
		return !cmdFailed() && maxTrivial <= trivialDelta && cic.HasSubFlow(false)
	}
	foldSubFlowByDepth := func() bool {
		return !cmdFailed() && maxDepth <= 1
	}
	foldSubFlow := func() bool {
		return foldSubFlowByDepth() || foldSubFlowByTrivial()
	}

	sysArgv := cmdEnv.GetSysArgv(cmd.Path(), cc.Cmds.Strs.PathSep)
	cmdInBg = parentInBg || sysArgv.IsDelay()

	// Folding if too trivial or too deep
	if !notFold() {
		core.TryExeEnvOpCmds(argv, cc, cmdEnv, flow, currCmdIdx, envOpCmds, nil,
			"failed to execute env-op cmd in flow desc")

		// This is for render checking, even it's folded
		subFlow, _, rendered := cic.Flow(argv, cc, cmdEnv, true, true)
		if rendered {
			parsedFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
			err := parsedFlow.FirstErr()
			if err != nil {
				panic(err.Error)
			}
			parsedFlow.GlobalEnv.WriteNotArgTo(env, cc.Cmds.Strs.EnvValDelAllMark)
			var executedFlow *core.ExecutedFlow
			if executedCmd != nil {
				executedFlow = executedCmd.SubFlow
			}
			return cmdInBg, dumpFlow(cc, env, envOpCmds, parsedFlow, 0, args, executedFlow, running,
				cmdInBg, writtenKeys, maxDepth-1, maxTrivial-trivialDelta, depth+1, cmdFailed() || parentIncompleted)
		}
		return cmdInBg, !cmdFailed()
	}

	showTrivialMark := foldSubFlowByTrivial() && trivialDelta > 0
	name, ok := dumpCmdDisplayName(cmdEnv, parsedCmd, args, executedCmd, running, cmdInBg, showTrivialMark)
	if len(name) != 0 {
		prt(0, name)
	}
	if !ok {
		return cmdInBg, false
	}

	dumpCmdHelp(cic.Help(), cmdEnv, args, prt)

	//lineLimit := env.GetInt("display.width")
	_, lineLimit := utils.GetTerminalWidth(50, 120)
	if args.MonitorMode {
		lineLimit = 9999
	}

	var startEnv map[string]string
	if !cmdSkipped() {
		startEnv = dumpExecutedStartEnv(cmdEnv, prt, padLenCal, args, executedCmd, lineLimit)
	}

	if !cmdSkipped() || executedCmd == nil {
		dumpCmdArgv(cic, argv, cmdEnv, originEnv, prt, args, executedCmd, writtenKeys)
	}

	dumpCmdExecutedLog(cmdEnv, args, executedCmd, prt, padLenCal, lineLimit)
	dumpCmdExecutedErr(cmdEnv, args, executedCmd, prt)

	if !cmdSkipped() && cmdFailed() || executedCmd == nil {
		dumpCmdEnvValues(cc, flow, parsedCmd, argv, cmdEnv, originEnv, prt, padLenCal, args, writtenKeys, lineLimit)
	}
	if !cmdSkipped() && cmdFailed() || executedCmd == nil || (!args.Skeleton && !args.Simple) {
		dumpEnvOpsDefinition(cic, argv, cmdEnv, prt, args, executedCmd)
	}
	if !cmdSkipped() && cmdFailed() || executedCmd == nil {
		dumpCmdTypeAndSource(cmd, cmdEnv, prt, args)
	}

	if !foldSubFlow() && !cic.IsBuiltinCmd() && (cic.HasCmdLine() || cic.HasSubFlow(false)) {
		metFlow := false
		if cic.HasSubFlow(false) {
			flowStrs, _, _ := cic.RenderedFlowStrs(argv, cc, cmdEnv, true, true)
			flowStr := strings.Join(flowStrs, " ")
			metFlow = metFlows[flowStr]
			if !cmdFailed() && metFlow && executedCmd == nil {
				if !args.Skeleton {
					prt(1, ColorProp("- flow (duplicated):", env))
				} else {
					prt(1, ColorProp("(duplicated sub flow)", env))
				}
			} else {
				metFlows[flowStr] = true
				if foldSubFlowByDepth() {
					if !args.Skeleton {
						prt(1, ColorProp("- flow (folded):", env))
					} else {
						prt(1, ColorProp("(folded flow)", env))
					}
				} else {
					if !args.Skeleton && !foldSubFlow() {
						prt(1, ColorProp("- flow:", env))
					}
				}
			}
			if !args.Skeleton && !foldSubFlow() && !cmdSkipped() {
				for _, flowStr := range flowStrs {
					prt(2, ColorFlow(flowStr, env))
				}
			}
		} else if !cmdSkipped() {
			dumpCmdExecutable(cic, env, args, prt, padLenCal, lineLimit)
		}

		if cic.HasSubFlow(false) && !cmdSkipped() && (executedCmd == nil || executedCmd.Result != core.ExecutedResultUnRun) {
			subFlow, _, rendered := cic.Flow(argv, cc, cmdEnv, true, true)
			if rendered && len(subFlow) != 0 {
				if !metFlow || executedCmd != nil || cmdFailed() {
					depthMark := ""
					if executedCmd != nil && !args.Skeleton {
						depthMark = fmt.Sprintf(" L%v", depth+1)
					}
					if !foldSubFlow() && !args.MonitorMode {
						prt(2, ColorFlowing("--->>>"+depthMark, env))
					}
					parsedFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
					err := parsedFlow.FirstErr()
					if err != nil {
						panic(err.Error)
					}

					subflowEnv := cmdEnv.NewLayer(core.EnvLayerSubFlow)
					parsedFlow.GlobalEnv.WriteNotArgTo(subflowEnv, cc.Cmds.Strs.EnvValDelAllMark)

					var executedFlow *core.ExecutedFlow
					if executedCmd != nil {
						executedFlow = executedCmd.SubFlow
					}
					newMaxDepth := maxDepth
					if !cmdFailed() {
						newMaxDepth -= 1
					}
					ok := dumpFlow(cc, subflowEnv, envOpCmds, parsedFlow, 0, args, executedFlow, running,
						cmdInBg, writtenKeys, newMaxDepth, maxTrivial-trivialDelta, depth+1, cmdFailed() || parentIncompleted)
					if !ok {
						return cmdInBg, false
					}
					exeMark := ""
					if cic.Type() == core.CmdTypeFileNFlow {
						exeMark += ColorCmd(" +", env)
					}
					if !foldSubFlow() && !args.MonitorMode {
						prt(2, ColorFlowing("<<<---"+depthMark, env)+exeMark)
					}
				}
			}
		}
	}

	core.TryExeEnvOpCmds(argv, cc, cmdEnv, flow, currCmdIdx, envOpCmds, nil,
		"failed to execute env-op cmd in flow desc")

	if !cmdSkipped() {
		dumpExecutedModifiedEnv(env, prt, padLenCal, args, startEnv, executedCmd, lineLimit)
	}

	return cmdInBg, !cmdFailed()
}

func dumpCmdHelp(help string, env *core.Env, args *DumpFlowArgs, prt func(indentLvl int, msg string)) {
	if args.MonitorMode {
		return
	}
	if len(help) == 0 {
		return
	}
	if !args.Skeleton {
		prt(1, " "+ColorHelp("'"+help+"'", env))
	} else {
		prt(1, ColorHelp("'"+help+"'", env))
	}
}

func dumpCmdExecutedLog(
	env *core.Env,
	args *DumpFlowArgs,
	executedCmd *core.ExecutedCmd,
	prt func(indentLvl int, msg string),
	padLenCal func(indentLvl int) int,
	lineLimit int) {

	if executedCmd == nil || len(executedCmd.LogFilePath) == 0 {
		return
	}
	if (args.Skeleton || args.Simple) && executedCmd.Result == core.ExecutedResultSucceeded {
		return
	}

	padLen := padLenCal(2)
	limit := lineLimit - padLen

	if !args.MonitorMode {
		if executedCmd.Result == core.ExecutedResultSucceeded {
			prt(1, ColorProp("- execute-log:", env))
		} else {
			//prt(1, ColorHighLight("- execute-log:", env))
			prt(1, ColorProp("- execute-log:", env))
		}
	}
	prt(2, mayTrimStr(executedCmd.LogFilePath, env, limit))

	displayLogLines := 8
	if args.MonitorMode {
		displayLogLines = 3
	}
	lines := utils.ReadLogFileLastLines(executedCmd.LogFilePath, 1024*16, displayLogLines)
	//if len(lines) != 0 {
	//	prt(2, ColorExplain("...", env))
	//}
	for _, line := range lines {
		i := 0
		line = strings.TrimSpace(line)
		line = strings.Map(func(r rune) rune {
			if unicode.IsGraphic(r) {
				return r
			}
			return -1
		}, line)
		line = stripTermStyle(line)
		if len(line) > 2*lineLimit {
			line = line[0:limit]
		}
		for {
			j := i + limit
			if j < len(line) {
				prt(2, ColorExplain(line[i:j], env))
				i = j
			} else {
				prt(2, ColorExplain(line[i:], env))
				break
			}
		}
	}
}

func dumpCmdExecutedErr(
	env *core.Env,
	args *DumpFlowArgs,
	executedCmd *core.ExecutedCmd,
	prt func(indentLvl int, msg string)) {

	if executedCmd == nil {
		return
	}

	if !args.Skeleton {
		if len(executedCmd.ErrStrs) != 0 && !args.MonitorMode {
			prt(1, ColorError("- err-msg:", env))
		}
		for _, line := range executedCmd.ErrStrs {
			prt(2, ColorError(strings.TrimSpace(line), env))
		}
	} else {
		if len(executedCmd.ErrStrs) != 0 && !args.MonitorMode {
			prt(0, "  "+ColorError(" - err-msg:", env))
		}
		for _, line := range executedCmd.ErrStrs {
			prt(2, ColorError(strings.TrimSpace(line), env))
		}
	}
}

func dumpCmdDisplayName(
	env *core.Env,
	parsedCmd core.ParsedCmd,
	args *DumpFlowArgs,
	executedCmd *core.ExecutedCmd,
	running bool,
	inBg bool,
	showTrivialMark bool) (name string, ok bool) {

	if args.MonitorMode &&
		!(executedCmd != nil && (executedCmd.Result == core.ExecutedResultError || executedCmd.Result == core.ExecutedResultIncompleted)) {
		return "", true
	}

	cmd := parsedCmd.Last().Matched.Cmd
	if cmd == nil {
		panic(fmt.Errorf("should never happen"))
	}
	cic := cmd.Cmd()
	if cic == nil {
		panic(fmt.Errorf("should never happen"))
	}

	sep := cmd.Strs.PathSep
	trivialMark := env.GetRaw("strs.trivial-mark")

	sysArgv := env.GetSysArgv(cmd.Path(), sep)

	cmdId := strings.Join(parsedCmd.Path(), sep)
	if args.Skeleton || args.Simple {
		name = cmdId
	} else {
		name = parsedCmd.DisplayPath(sep, true)
	}
	if inBg {
		name = ColorCmdDelay("["+name+"]", env)
	} else {
		name = ColorCmd("["+name+"]", env)
	}
	if showTrivialMark {
		name += ColorProp(trivialMark, env)
	}

	if sysArgv.IsDelay() {
		//if sysArgv.GetDelayDuration().Nanoseconds() == 0 {
		//	name += ColorCmdDelay(" (background)", env)
		//} else {
		name += ColorCmdDelay(" (schedule to bg in ", env) + sysArgv.GetDelayStr() + ColorCmdDelay(")", env)
		//}
	}

	if executedCmd != nil {
		if executedCmd.Cmd != cmdId {
			// TODO: better display
			name += ColorSymbol(" - ", env) + ColorError("flow not matched, origin cmd: ", env) +
				ColorCmd("["+executedCmd.Cmd+"]", env)
			return name, false
		}

		name += " " + ColorExplain(executedCmd.StartTs.Format(core.SessionTimeShortFormat), env) + " "

		resultStr := string(executedCmd.Result)
		if executedCmd.Result == core.ExecutedResultError {
			name += ColorError(resultStr, env)
		} else if executedCmd.Result == core.ExecutedResultSucceeded {
			name += ColorCmdDone(resultStr, env)
		} else if executedCmd.Result == core.ExecutedResultSkipped {
			name += ColorExplain(resultStr, env)
		} else if executedCmd.Result == core.ExecutedResultIncompleted {
			if running {
				name += ColorHighLight("running", env)
			} else {
				name += ColorWarn("failed", env)
			}
		} else if executedCmd.Result == core.ExecutedResultUnRun {
			name += ColorHighLight(resultStr, env)
		} else {
			name += ColorExplain(resultStr, env)
		}

		if !executedCmd.StartTs.IsZero() {
			finishTs := executedCmd.FinishTs
			if running {
				finishTs = time.Now().Round(time.Second)
			}
			var durStr string
			if executedCmd.StartTs != finishTs || executedCmd.Result != core.ExecutedResultIncompleted || running {
				dur := finishTs.Sub(executedCmd.StartTs)
				durStr += formatDuration(dur)
			}
			if (executedCmd.Result == core.ExecutedResultIncompleted) && !running && len(durStr) != 0 {
				durStr += ColorExplain("+?", env)
			}
			if len(durStr) != 0 {
				durStr += " "
			}
			name += " " + ColorExplain(durStr, env)
		}
	}
	return name, true
}

func dumpCmdExecutable(
	cic *core.Cmd,
	env *core.Env,
	args *DumpFlowArgs,
	prt func(indentLvl int, msg string),
	padLenCal func(indentLvl int) int,
	lineLimit int) {

	if args.Simple || args.Skeleton {
		return
	}

	padLen := padLenCal(2)
	limit := lineLimit - padLen

	if cic.Type() == core.CmdTypeEmptyDir {
		prt(1, ColorProp("- dir:", env))
		prt(2, mayTrimStr(cic.CmdLine(), env, limit))
	} else {
		prt(1, ColorProp("- executable:", env))
		prt(2, mayTrimStr(cic.CmdLine(), env, limit))
	}
	if len(cic.MetaFile()) != 0 {
		prt(1, ColorProp("- meta:", env))
		prt(2, mayTrimStr(cic.MetaFile(), env, limit))
	}
}

func dumpCmdTypeAndSource(
	cmd *core.CmdTree,
	env *core.Env,
	prt func(indentLvl int, msg string),
	args *DumpFlowArgs) {

	cic := cmd.Cmd()
	if cic == nil {
		panic(fmt.Errorf("should never happen"))
	}

	if args.Simple || args.Skeleton {
		return
	}

	/*
		line := string(cic.Type())
		if cic.IsQuiet() {
			line += " quiet"
		}
		if cic.IsPriority() {
			line += " priority"
		}
		prt(1, ColorProp("- cmd-type:", env))
		prt(2, line)
	*/

	if len(cmd.Source()) != 0 && !strings.HasPrefix(cic.CmdLine(), cmd.Source()) {
		prt(1, ColorProp("- from:", env))
		prt(2, cmd.Source())
	}
}

func dumpCmdEnvValues(
	cc *core.Cli,
	flow *core.ParsedCmds,
	parsedCmd core.ParsedCmd,
	argv core.ArgVals,
	env *core.Env,
	originEnv *core.Env,
	prt func(indentLvl int, msg string),
	padLenCal func(indentLvl int) int,
	args *DumpFlowArgs,
	writtenKeys FlowWrittenKeys,
	lineLimit int) {

	cmd := parsedCmd.Last().Matched.Cmd
	if cmd == nil {
		panic(fmt.Errorf("should never happen"))
	}
	cic := cmd.Cmd()
	if cic == nil {
		panic(fmt.Errorf("should never happen"))
	}

	padLen := padLenCal(2)

	if !args.Skeleton {
		keys, kvs := dumpFlowEnv(cc, originEnv, flow.GlobalEnv, parsedCmd, cmd, argv, writtenKeys)
		if len(keys) != 0 {
			prt(1, ColorProp("- env-filling:", env))
		}
		for _, k := range keys {
			v := kvs[k]
			limit := lineLimit - (padLen + len(k) + 3)
			prt(2, ColorKey(k, env)+ColorSymbol(" = ", env)+normalizeValForDisplay(k, v.Val, env, limit)+" "+v.Source+"")
		}
	}
	writtenKeys.AddCmd(argv, env, cic)
}

func dumpCmdArgv(
	cic *core.Cmd,
	argv core.ArgVals,
	env *core.Env,
	originEnv *core.Env,
	prt func(indentLvl int, msg string),
	args *DumpFlowArgs,
	executedCmd *core.ExecutedCmd,
	writtenKeys FlowWrittenKeys) {

	if args.MonitorMode &&
		!(executedCmd != nil && (executedCmd.Result == core.ExecutedResultError || executedCmd.Result == core.ExecutedResultIncompleted)) {
		return
	}

	stackDepth := env.GetInt("sys.stack-depth")

	if args.Skeleton || args.MonitorMode {
		cicArgs := cic.Args()
		for _, name := range cicArgs.Names() {
			if _, ok := argv[name]; !ok {
				continue
			}
			val := argv[name]
			if !val.Provided {
				continue
			}
			indent := " "
			if args.MonitorMode {
				indent = "   "
			}
			prt(1, indent+ColorArg(name, env)+ColorSymbol(" = ", env)+val.Raw)
		}
	} else {
		args := cic.Args()
		arg2env := cic.GetArg2Env()
		argLines := DumpEffectedArgs(originEnv, arg2env, &args, argv, writtenKeys, stackDepth)
		if len(argLines) != 0 {
			prt(1, ColorProp("- args:", env))
		}
		for _, line := range argLines {
			prt(2, line)
		}
	}
}

func dumpEnvOpsDefinition(
	cic *core.Cmd,
	argv core.ArgVals,
	env *core.Env,
	prt func(indentLvl int, msg string),
	args *DumpFlowArgs,
	executedCmd *core.ExecutedCmd) {

	envOpSep := " " + env.GetRaw("strs.env-op-sep") + " "

	if args.Skeleton {
		return
	}
	if args.Simple && executedCmd != nil && !args.ShowExecutedEnvFull {
		return
	}

	envOps := cic.EnvOps()
	envOpKeys, origins, _ := envOps.RenderedEnvKeys(argv, env, cic, false)
	if len(envOpKeys) != 0 {
		prt(1, ColorProp("- env-ops:", env))
	}
	for i, k := range envOpKeys {
		prt(2, ColorKey(k, env)+ColorSymbol(" = ", env)+
			dumpEnvOps(envOps.Ops(origins[i]), envOpSep)+dumpIsAutoTimerKey(env, cic, k))
	}
}

// TODO: env-before-execute => env-full-trace
func dumpExecutedStartEnv(
	env *core.Env,
	prt func(indentLvl int, msg string),
	padLenCal func(indentLvl int) int,
	args *DumpFlowArgs,
	executedCmd *core.ExecutedCmd,
	lineLimit int) (startEnv map[string]string) {

	if args.MonitorMode {
		return
	}

	if executedCmd == nil {
		return
	}

	if executedCmd.StartEnv != nil {
		startEnv = executedCmd.StartEnv.FlattenAll()
	}

	keys := []string{}
	for k, _ := range startEnv {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if len(startEnv) == 0 {
		return
	}

	if !args.ShowExecutedEnvFull && (executedCmd.Result != core.ExecutedResultError) {
		return
	}

	padLen := padLenCal(2)

	if !args.Skeleton {
		prt(1, ColorProp("- env-before-execute:", env))
		for _, k := range keys {
			limit := lineLimit - (padLen + len(k) + 3)
			prt(2, ColorKey(k, env)+ColorSymbol(" = ", env)+normalizeValForDisplay(k, startEnv[k], env, limit))
		}
	} else {
		prt(0, "  "+ColorProp(" - env-before-execute:", env))
		for _, k := range keys {
			limit := lineLimit - (padLen + len(k) + 3)
			prt(1, "   "+ColorKey(k, env)+ColorSymbol(" = ", env)+normalizeValForDisplay(k, startEnv[k], env, limit))
		}
	}
	return
}

// TODO: better display
func dumpExecutedModifiedEnv(
	env *core.Env,
	prt func(indentLvl int, msg string),
	padLenCal func(indentLvl int) int,
	args *DumpFlowArgs,
	startEnv map[string]string,
	executedCmd *core.ExecutedCmd,
	lineLimit int) {

	if args.MonitorMode {
		return
	}

	if executedCmd == nil {
		return
	}
	if args.ShowExecutedEnvFull {
		// TODO:
		// return
	}
	if !args.ShowExecutedModifiedEnv && executedCmd.Result != core.ExecutedResultError {
		return
	}
	if executedCmd.FinishEnv == nil && executedCmd.Result != core.ExecutedResultError {
		return
	}

	var finishEnv map[string]string
	if executedCmd.FinishEnv != nil {
		finishEnv = executedCmd.FinishEnv.FlattenAll()
	}

	padLen := padLenCal(2)

	lines := []string{}
	for k, v := range startEnv {
		op := ""
		limit := lineLimit
		tipLen := 4
		afterV, hasAfter := finishEnv[k]
		if !hasAfter {
			tipLen += 9
		} else {
			tipLen += 14
		}
		prefixLen := padLen + len(k) + 3 + tipLen
		val := normalizeValForDisplay(k, v, env, limit-prefixLen)
		if len(val) != len(v) {
			limit = 0
		}
		if !hasAfter {
			op = ColorSymbol(" <- ", env) + ColorTip("(deleted)", env)
		} else if afterV != v {
			op = ColorSymbol(" <- ", env) + ColorTip("(modified to) ", env)
			prefixLen += len(val) + len(op) - ColorExtraLen(env, "symbol", "tip")
			op = op + normalizeValForDisplay(k, afterV, env, limit-prefixLen)
		} else {
			continue
		}
		lines = append(lines, ColorKey(k, env)+ColorSymbol(" = ", env)+val+op)
	}

	for k, v := range finishEnv {
		_, ok := startEnv[k]
		if ok {
			continue
		}
		op := ColorSymbol(" <- ", env) + ColorTip("(added)", env)
		prefixLen := padLen + len(k) + 3 + len(op) - ColorExtraLen(env, "symbol", "tip")
		lines = append(lines, ColorKey(k, env)+ColorSymbol(" = ", env)+normalizeValForDisplay(k, v, env, lineLimit-prefixLen)+op)
	}

	if len(lines) == 0 {
		return
	}

	sort.Strings(lines)
	if !args.Skeleton {
		if len(lines) != 0 || executedCmd.Result != core.ExecutedResultSucceeded {
			prt(1, ColorProp("- env-modified:", env))
		}
		if len(lines) == 0 && executedCmd.Result != core.ExecutedResultSucceeded {
			prt(2, ColorExplain("(none)", env))
		} else {
			for _, line := range lines {
				prt(2, line)
			}
		}
	} else {
		if len(lines) != 0 || executedCmd.Result != core.ExecutedResultSucceeded {
			prt(0, "  "+ColorProp(" - env-modified:", env))
		}
		if len(lines) == 0 && executedCmd.Result != core.ExecutedResultSucceeded {
			prt(2, ColorExplain("(none)", env))
		} else {
			for _, line := range lines {
				prt(2, line)
			}
		}
	}
}

type flowEnvVal struct {
	Val    string
	Source string
}

func dumpFlowEnv(
	cc *core.Cli,
	env *core.Env,
	parsedGlobalEnv core.ParsedEnv,
	parsedCmd core.ParsedCmd,
	cmd *core.CmdTree,
	argv core.ArgVals,
	writtenKeys FlowWrittenKeys) (keys []string, kvs map[string]flowEnvVal) {

	kvs = map[string]flowEnvVal{}
	cic := cmd.Cmd()

	tempEnv := core.NewEnv()
	parsedGlobalEnv.WriteNotArgTo(tempEnv, cc.Cmds.Strs.EnvValDelAllMark)
	cmdEssEnv := parsedCmd.GenCmdEnv(tempEnv, cc.Cmds.Strs.EnvValDelAllMark)
	val2env := cic.GetVal2Env()
	for _, k := range val2env.EnvKeys() {
		kvs[k] = flowEnvVal{val2env.Val(k), ColorSymbol("<- mod", env)}
	}

	flatten := cmdEssEnv.Flatten(true, nil, true)
	for k, v := range flatten {
		kvs[k] = flowEnvVal{v, ColorSymbol("<- flow", env)}
	}

	arg2env := cic.GetArg2Env()
	for name, val := range argv {
		if !val.Provided && len(val.Raw) == 0 {
			continue
		}
		key, hasMapping := arg2env.GetEnvKey(name)
		if !hasMapping {
			continue
		}
		_, inEnv := env.GetEx(key)
		if !val.Provided && inEnv {
			continue
		}
		if writtenKeys[key] {
			continue
		}
		kvs[key] = flowEnvVal{val.Raw, ColorSymbol("<- arg", env) +
			ColorArg(" '"+name+"'", env)}
	}

	for k, _ := range kvs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return
}

func getCmdTrivial(parsedCmd core.ParsedCmd) (trivial int) {
	trivial += parsedCmd.TrivialLvl
	cmd := parsedCmd.Last().Matched.Cmd
	if cmd == nil {
		return
	}
	trivial += cmd.Trivial()
	return
}

func normalizeValForDisplay(key string, val string, env *core.Env, limit int) string {
	val = mayMaskSensitiveVal(env, key, val)
	return mayQuoteMayTrimStr(val, env, limit)
}

func mayQuoteMayTrimStr(s string, env *core.Env, limit int) string {
	return mayTrimStr(mayQuoteStr(s), env, limit)
}

func mayTrimStr(s string, env *core.Env, limit int) string {
	if len(s) > limit {
		if limit <= 1 {
			return ColorExplain(".", env)
		}
		if limit <= 2 {
			return ColorExplain("..", env)
		}
		if limit <= 3 {
			return ColorExplain("...", env)
		}
		//half := limit / 2
		//s = s[0:half-2] + ColorExplain("...", env) + s[len(s)-half+1:]
		s = ColorExplain("...", env) + s[len(s)-limit+3:]
	}
	return s
}

type DumpFlowArgs struct {
	Simple                  bool
	Skeleton                bool
	IndentSize              int
	MaxDepth                int
	MaxTrivial              int
	ShowExecutedEnvFull     bool
	ShowExecutedModifiedEnv bool
	MonitorMode             bool
}

func NewDumpFlowArgs() *DumpFlowArgs {
	return &DumpFlowArgs{false, false, 4, 32, 1, false, false, false}
}

func (self *DumpFlowArgs) SetSimple() *DumpFlowArgs {
	self.Simple = true
	return self
}

func (self *DumpFlowArgs) SetMaxDepth(val int) *DumpFlowArgs {
	self.MaxDepth = val
	return self
}

func (self *DumpFlowArgs) SetMaxTrivial(val int) *DumpFlowArgs {
	self.MaxTrivial = val
	return self
}

func (self *DumpFlowArgs) SetSkeleton() *DumpFlowArgs {
	self.Simple = true
	self.Skeleton = true
	return self
}

func (self *DumpFlowArgs) SetShowExecutedEnvFull() *DumpFlowArgs {
	self.ShowExecutedEnvFull = true
	return self
}

func (self *DumpFlowArgs) SetShowExecutedModifiedEnv() *DumpFlowArgs {
	self.ShowExecutedModifiedEnv = true
	return self
}

func (self *DumpFlowArgs) SetMonitorMode() *DumpFlowArgs {
	self.MonitorMode = true
	return self
}

type FlowWrittenKeys map[string]bool

func (self FlowWrittenKeys) AddCmd(argv core.ArgVals, env *core.Env, cic *core.Cmd) {
	if cic == nil {
		return
	}
	ops := cic.EnvOps()
	keys, _, _ := ops.RenderedEnvKeys(argv, env, cic, false)
	for _, k := range keys {
		// If is read-op, then the key must exists, so no need to check the op flags
		self[k] = true
	}
}

func stripTermStyle(str string) string {
	return re.ReplaceAllString(str, "")
}

func init() {
	re = regexp.MustCompile(AnsiChars)
}

var re *regexp.Regexp

const AnsiChars = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
