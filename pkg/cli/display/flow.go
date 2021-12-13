package display

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
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

	// The env will be modified during dumping (so it could show the real value)
	// so we need to clone the env to protect it
	env = env.Clone()

	writtenKeys := FlowWrittenKeys{}

	if executedFlow != nil {
	} else {
		PrintTipTitle(cc.Screen, env, "flow executing description:")
	}

	cc.Screen.Print(ColorFlowing("--->>>", env) + "\n")
	ok := dumpFlow(cc, env, envOpCmds, flow, fromCmdIdx, args, executedFlow, running,
		writtenKeys, args.MaxDepth, args.MaxTrivial, 0)
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
	writtenKeys FlowWrittenKeys,
	maxDepth int,
	maxTrivial int,
	indentAdjust int) bool {

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
				return false
			}
		}
		ok := dumpFlowCmd(cc, cc.Screen, env, envOpCmds, flow, fromCmdIdx+i, args, executedCmd, running,
			maxDepth, maxTrivial, indentAdjust, metFlows, writtenKeys)
		if !ok {
			return false
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
	maxDepth int,
	maxTrivial int,
	indentAdjust int,
	metFlows map[string]bool,
	writtenKeys FlowWrittenKeys) bool {

	// TODO: too complicated with executedCmd

	parsedCmd := flow.Cmds[currCmdIdx]
	cmd := parsedCmd.Last().Matched.Cmd
	if cmd == nil {
		return true
	}

	sep := cmd.Strs.PathSep

	prt := func(indentLvl int, msg string) {
		indentLvl += indentAdjust
		padding := rpt(" ", args.IndentSize*indentLvl)
		msg = autoPadNewLine(padding, msg)
		screen.Print(padding + msg + "\n")
	}

	cic := cmd.Cmd()
	if cic == nil {
		return true
	}

	trivialDelta := cmd.Trivial() + parsedCmd.TrivialLvl

	cmdEnv, argv := parsedCmd.ApplyMappingGenEnvAndArgv(env, cc.Cmds.Strs.EnvValDelAllMark, sep)

	if (executedCmd != nil && !executedCmd.Succeeded) || (maxTrivial > 0 && maxDepth > 0) {
		showTrivialMark := (executedCmd == nil || executedCmd.Succeeded) &&
			maxTrivial == 1 && trivialDelta > 0 && cic.HasSubFlow()
		name, ok := dumpCmdDisplayName(cmdEnv, parsedCmd, args, executedCmd, running, showTrivialMark)
		prt(0, name)
		if !ok {
			return false
		}

		dumpCmdHelp(cic.Help(), cmdEnv, args, prt)

		if executedCmd != nil {
			if executedCmd.Unexecuted {
				return true
			}
			dumpCmdExecutedErr(cmdEnv, args, executedCmd, prt)
		}
	}

	// TODO: this is slow
	originEnv := env.Clone()

	if (executedCmd != nil && !executedCmd.Succeeded) || (maxTrivial > 0 && maxDepth > 0) {
		dumpCmdArgv(cic, argv, cmdEnv, originEnv, prt, args, writtenKeys)
	}

	trivial := maxTrivial - trivialDelta

	// Folding if too trivial or too deep
	if (executedCmd == nil || executedCmd.Succeeded) &&
		(trivial <= 0 || maxDepth <= 0) {

		core.TryExeEnvOpCmds(argv, cc, cmdEnv, flow, currCmdIdx, envOpCmds, nil,
			"failed to execute env-op cmd in flow desc")

		// This is for render checking, even it's folded
		subFlow, rendered := cic.Flow(argv, cmdEnv, true)
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
			return dumpFlow(cc, env, envOpCmds, parsedFlow, 0, args, executedFlow, running,
				writtenKeys, maxDepth-1, trivial, indentAdjust+2)
		}
		return executedCmd == nil || executedCmd.Succeeded
	}

	dumpCmdEnvValues(cc, flow, parsedCmd, argv, cmdEnv, originEnv, prt, args, writtenKeys)
	dumpEnvOpsInFlow(cic, argv, cmdEnv, prt, args, executedCmd)
	dumpCmdTypeAndSource(cmd, cmdEnv, prt, args)

	if (len(cic.CmdLine()) != 0 || len(cic.FlowStrs()) != 0) &&
		cic.Type() != core.CmdTypeNormal && cic.Type() != core.CmdTypePower {
		metFlow := false
		if cic.Type() == core.CmdTypeFlow || cic.Type() == core.CmdTypeFileNFlow {
			flowStrs, _ := cic.RenderedFlowStrs(argv, cmdEnv, true)
			flowStr := strings.Join(flowStrs, " ")
			metFlow = metFlows[flowStr]
			if (executedCmd == nil || executedCmd.Succeeded) && metFlow {
				if !args.Skeleton {
					prt(1, ColorProp("- flow (duplicated):", env))
				} else {
					prt(1, ColorProp("(duplicated sub flow)", env))
				}
			} else {
				metFlows[flowStr] = true
				if (executedCmd == nil || executedCmd.Succeeded) && maxDepth <= 1 {
					if !args.Skeleton {
						prt(1, ColorProp("- flow (folded):", env))
					} else {
						prt(1, ColorProp("(folded flow)", env))
					}
				} else {
					if !args.Skeleton {
						prt(1, ColorProp("- flow:", env))
					}
				}
			}
			if !args.Skeleton {
				for _, flowStr := range flowStrs {
					prt(2, ColorFlow(flowStr, env))
				}
			}
		} else {
			dumpCmdExecutable(cic, env, prt, args)
		}

		if cic.Type() == core.CmdTypeFlow || cic.Type() == core.CmdTypeFileNFlow {
			subFlow, rendered := cic.Flow(argv, cmdEnv, true)
			if rendered && len(subFlow) != 0 {
				if !(metFlow && (executedCmd == nil || executedCmd.Succeeded)) {
					if (executedCmd != nil && !executedCmd.Succeeded) || maxDepth > 1 {
						prt(2, ColorFlowing("--->>>", env))
					}
					parsedFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
					err := parsedFlow.FirstErr()
					if err != nil {
						panic(err.Error)
					}

					// TODO: check this is ok or not
					subflowEnv := cmdEnv.NewLayer(core.EnvLayerSubFlow)
					parsedFlow.GlobalEnv.WriteNotArgTo(subflowEnv, cc.Cmds.Strs.EnvValDelAllMark)

					var executedFlow *core.ExecutedFlow
					if executedCmd != nil {
						executedFlow = executedCmd.SubFlow
					}
					newMaxDepth := maxDepth
					if executedCmd == nil || executedCmd.Succeeded {
						newMaxDepth -= 1
					}
					ok := dumpFlow(cc, subflowEnv, envOpCmds, parsedFlow, 0, args, executedFlow, running,
						writtenKeys, newMaxDepth, trivial, indentAdjust+2)
					if !ok {
						return false
					}
					exeMark := ""
					if cic.Type() == core.CmdTypeFileNFlow {
						exeMark = ColorCmd(" +", env)
					}
					if (executedCmd == nil || executedCmd.Succeeded) && maxDepth > 1 {
						prt(2, ColorFlowing("<<<---", env)+exeMark)
					}
				}
			}
		}
	}

	core.TryExeEnvOpCmds(argv, cc, cmdEnv, flow, currCmdIdx, envOpCmds, nil,
		"failed to execute env-op cmd in flow desc")

	startEnv := dumpExecutedEnvFull(cmdEnv, prt, args, executedCmd)
	dumpExecutedModifiedEnv(env, prt, args, startEnv, executedCmd)

	return executedCmd == nil || executedCmd.Succeeded
}

func dumpCmdHelp(help string, env *core.Env, args *DumpFlowArgs, prt func(indentLvl int, msg string)) {
	if len(help) == 0 {
		return
	}
	if !args.Skeleton {
		prt(1, " "+ColorHelp("'"+help+"'", env))
	} else {
		prt(1, ColorHelp("'"+help+"'", env))
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
		if len(executedCmd.Err) != 0 {
			prt(1, ColorError("- error:", env))
		}
		for _, line := range executedCmd.Err {
			prt(2, ColorError(strings.TrimSpace(line), env))
		}
	} else {
		if len(executedCmd.Err) != 0 {
			prt(0, "  "+ColorError(" - error:", env))
		}
		for _, line := range executedCmd.Err {
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
	showTrivialMark bool) (name string, ok bool) {

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
	if args.Skeleton {
		name = cmdId
	} else {
		name = parsedCmd.DisplayPath(sep, true)
	}
	if sysArgv.IsDelay() {
		name = ColorCmdDelay("["+name+"]", env)
	} else {
		name = ColorCmd("["+name+"]", env)
	}
	if showTrivialMark {
		name += ColorProp(trivialMark, env)
	}

	if sysArgv.IsDelay() {
		name += ColorCmdDelay(" (schedule in ", env) + sysArgv.GetDelayStr() + ColorCmdDelay(")", env)
	}

	if executedCmd != nil {
		if executedCmd.Cmd != cmdId {
			// TODO: better display
			name += ColorSymbol(" - ", env) + ColorError("flow not matched, origin cmd: ", env) +
				ColorCmd("["+executedCmd.Cmd+"]", env)
			return name, false
		}
		if executedCmd.Unexecuted {
			name += ColorSymbol(" - ", env) + ColorExplain("un-run", env)
		} else if executedCmd.Succeeded {
			name += ColorSymbol(" - ", env) + ColorCmdDone("OK", env)
		} else if running {
			name += ColorSymbol(" - ", env) + ColorError("not-done", env)
		} else {
			name += ColorSymbol(" - ", env) + ColorError("ERR", env)
		}
	}
	return name, true
}

func dumpCmdExecutable(
	cic *core.Cmd,
	env *core.Env,
	prt func(indentLvl int, msg string),
	args *DumpFlowArgs) {

	if args.Simple || args.Skeleton {
		return
	}

	if cic.Type() == core.CmdTypeEmptyDir {
		prt(1, ColorProp("- dir:", env))
		prt(2, cic.CmdLine())
	} else {
		prt(1, ColorProp("- executable:", env))
		prt(2, cic.CmdLine())
	}
	if len(cic.MetaFile()) != 0 {
		prt(1, ColorProp("- meta:", env))
		prt(2, cic.MetaFile())
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
	args *DumpFlowArgs,
	writtenKeys FlowWrittenKeys) {

	cmd := parsedCmd.Last().Matched.Cmd
	if cmd == nil {
		panic(fmt.Errorf("should never happen"))
	}
	cic := cmd.Cmd()
	if cic == nil {
		panic(fmt.Errorf("should never happen"))
	}

	if !args.Skeleton {
		keys, kvs := dumpFlowEnv(cc, originEnv, flow.GlobalEnv, parsedCmd, cmd, argv, writtenKeys)
		if len(keys) != 0 {
			prt(1, ColorProp("- env-pre-handle:", env))
		}
		for _, k := range keys {
			v := kvs[k]
			prt(2, ColorKey(k, env)+ColorSymbol(" = ", env)+mayQuoteMayTrimStr(v.Val, env)+" "+v.Source+"")
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
	writtenKeys FlowWrittenKeys) {

	if !args.Skeleton {
		args := cic.Args()
		arg2env := cic.GetArg2Env()
		argLines := DumpEffectedArgs(originEnv, arg2env, &args, argv, writtenKeys)
		if len(argLines) != 0 {
			prt(1, ColorProp("- args:", env))
		}
		for _, line := range argLines {
			prt(2, line)
		}
	} else {
		for name, val := range argv {
			if !val.Provided {
				continue
			}
			prt(1, " "+ColorArg(name, env)+ColorSymbol(" = ", env)+val.Raw)
		}
	}
}

func dumpEnvOpsInFlow(
	cic *core.Cmd,
	argv core.ArgVals,
	env *core.Env,
	prt func(indentLvl int, msg string),
	args *DumpFlowArgs,
	executedCmd *core.ExecutedCmd) {

	envOpSep := " " + env.GetRaw("strs.env-op-sep") + " "

	//if args.Skeleton || (executedCmd != nil && args.ShowExecutedEnvFull) {
	if args.Skeleton {
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
func dumpExecutedEnvFull(
	env *core.Env,
	prt func(indentLvl int, msg string),
	args *DumpFlowArgs,
	executedCmd *core.ExecutedCmd) (startEnv map[string]string) {

	if executedCmd != nil && executedCmd.StartEnv != nil {
		startEnv = executedCmd.StartEnv.FlattenAll()
	}

	keys := []string{}
	for k, _ := range startEnv {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// TODO
	if !args.ShowExecutedEnvFull || len(startEnv) == 0 {
		return
	}

	if !args.Skeleton {
		prt(1, ColorProp("- env-before-execute:", env))
		for _, k := range keys {
			prt(2, ColorKey(k, env)+ColorSymbol(" = ", env)+mayQuoteMayTrimStr(startEnv[k], env))
		}
	} else {
		prt(0, "  "+ColorProp(" - env-before-execute:", env))
		for _, k := range keys {
			prt(1, "   "+ColorKey(k, env)+ColorSymbol(" = ", env)+mayQuoteMayTrimStr(startEnv[k], env))
		}
	}
	return
}

// TODO: better display
func dumpExecutedModifiedEnv(
	env *core.Env,
	prt func(indentLvl int, msg string),
	args *DumpFlowArgs,
	startEnv map[string]string,
	executedCmd *core.ExecutedCmd) {

	if executedCmd == nil {
		return
	}
	if args.ShowExecutedEnvFull {
		//return
	}
	if !args.ShowExecutedModifiedEnv && executedCmd.Succeeded {
		return
	}
	if executedCmd.FinishEnv == nil && !executedCmd.Succeeded {
		return
	}

	var finishEnv map[string]string
	if executedCmd.FinishEnv != nil {
		finishEnv = executedCmd.FinishEnv.FlattenAll()
	}

	lines := []string{}
	for k, v := range startEnv {
		op := ""
		val := mayQuoteMayTrimStr(v, env)
		newV, ok := finishEnv[k]
		if !ok {
			op = ColorSymbol(" <- ", env) + ColorTip("(deleted)", env)
		} else if newV != v {
			op = ColorSymbol(" <- ", env) + ColorTip("(modified to) ", env) + mayQuoteMayTrimStr(newV, env)
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
		lines = append(lines, ColorKey(k, env)+ColorSymbol(" = ", env)+mayQuoteMayTrimStr(v, env)+op)
	}

	sort.Strings(lines)
	if !args.Skeleton {
		if len(lines) != 0 || !executedCmd.Succeeded {
			prt(1, ColorProp("- env-modified:", env))
		}
		if len(lines) == 0 && !executedCmd.Succeeded {
			prt(2, ColorExplain("(none)", env))
		} else {
			for _, line := range lines {
				prt(2, line)
			}
		}
	} else {
		if len(lines) != 0 || !executedCmd.Succeeded {
			prt(0, "  "+ColorProp(" - env-modified:", env))
		}
		if len(lines) == 0 && !executedCmd.Succeeded {
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

// TODO: better display
func mayQuoteMayTrimStr(s string, env *core.Env) string {
	limit := env.GetInt("display.width")*3/5*2/2 - 10
	if limit < 10 {
		limit = 10
	}
	if len(s) > limit {
		half := limit / 2
		s = s[0:half-1] + ColorExplain("...", env) + s[len(s)-(half-2):]
	}
	return mayQuoteStr(s)
}

type DumpFlowArgs struct {
	Simple                  bool
	Skeleton                bool
	IndentSize              int
	MaxDepth                int
	MaxTrivial              int
	ShowExecutedEnvFull     bool
	ShowExecutedModifiedEnv bool
}

func NewDumpFlowArgs() *DumpFlowArgs {
	return &DumpFlowArgs{false, false, 4, 32, 1, false, false}
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
