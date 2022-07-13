package execute

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pingcap/ticat/pkg/builtin"
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
	"github.com/pingcap/ticat/pkg/utils"
)

type ExecFunc func(cc *core.Cli, flow *core.ParsedCmds, env *core.Env) bool

type Executor struct {
	sessionFileName       string
	sessionStatusFileName string
	callerNameBootstrap   string
	callerNameEntry       string
}

func NewExecutor(
	sessionFileName string,
	sessionStatusFileName string,
	callerNameBootstrap string,
	callerNameEntry string) *Executor {

	return &Executor{
		sessionFileName,
		sessionStatusFileName,
		callerNameBootstrap,
		callerNameEntry,
	}
}

func (self *Executor) Clone() core.Executor {
	return NewExecutor(
		self.sessionFileName,
		self.sessionStatusFileName,
		self.callerNameBootstrap,
		self.callerNameEntry)
}

func (self *Executor) Run(cc *core.Cli, env *core.Env, bootstrap string, input ...string) bool {
	overWriteBootstrap := env.Get("sys.bootstrap").Raw
	if len(overWriteBootstrap) != 0 {
		bootstrap = overWriteBootstrap
	}
	if !self.execute(self.callerNameBootstrap, cc, env, nil, true, false, bootstrap) {
		return false
	}

	// Do arg2env auto mapping between bootstrap (commands are loaded after this point) and executing
	cc.Arg2EnvAutoMapCmds.AutoMapArg2Env(cc, env, builtin.EnvOpCmds(), env.GetInt("sys.stack-depth"))

	ok := self.execute(self.callerNameEntry, cc, env, nil, false, false, input...)

	tryBreakAtEnd(cc, env)

	var errs []error
	if env.GetBool("sys.bg.wait") {
		errs = builtin.WaitBgTasks(cc, env, "")
		for _, err := range errs {
			display.PrintError(cc, env, err)
		}
	}

	ok = ok && len(errs) == 0

	if cc.FlowStatus != nil {
		cc.FlowStatus.OnFlowFinish(env, ok)
	}
	return ok
}

// Implement core.Executor
func (self *Executor) Execute(caller string, innerCall bool, cc *core.Cli, env *core.Env,
	masks []*core.ExecuteMask, input ...string) bool {
	return self.execute(caller, cc, env, masks, false, innerCall, input...)
}

func (self *Executor) execute(caller string, cc *core.Cli, env *core.Env, masks []*core.ExecuteMask,
	bootstrap bool, innerCall bool, input ...string) bool {

	if !innerCall && env.GetBool("sys.env.use-cmd-abbrs") {
		useCmdsAbbrs(cc.EnvAbbrs, cc.Cmds)
	}

	if !innerCall && len(input) == 0 {
		if !env.GetBool("sys.interact.inside") {
			display.PrintGlobalHelp(cc, env)
		}
		return true
	}

	if !innerCall && !bootstrap {
		useEnvAbbrs(cc.EnvAbbrs, env, cc.Cmds.Strs.EnvPathSep)
	}
	flow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, input...)
	if flow.GlobalEnv != nil {
		env = env.GetOneOfLayers(core.EnvLayerSubFlow, core.EnvLayerSession)
		flow.GlobalEnv.WriteNotArgTo(env, cc.Cmds.Strs.EnvValDelAllMark)
	}

	if !innerCall && !bootstrap {
		reordered, moved, tailModeCall, attempTailModeCall := moveLastPriorityCmdToFront(flow.Cmds)
		flow.Cmds = reordered
		flow.HasTailMode = moved
		flow.TailModeCall = tailModeCall
		flow.AttempTailModeCall = attempTailModeCall
		// TODO: this may not right if flow was changed in recursive process, but not big deal
		if moved && flow.GlobalCmdIdx == 0 {
			flow.GlobalCmdIdx = 1
		}
	}

	removeEmptyCmds(flow)

	if !innerCall && !bootstrap {
		checkTailModeCalls(flow)
	}

	if !innerCall {
		cc.Blender.Invoke(cc, env, flow)
	}

	if !allowParseError(flow) {
		isSearch := isStartWithSearchCmd(flow)
		if !display.HandleParseResult(cc, flow, env, isSearch) {
			return false
		}
	}

	display.PrintTolerableErrs(cc.Screen, env, cc.TolerableErrs)

	crossProcessInnerCall := false

	if !innerCall && !bootstrap && !env.GetBool("sys.interact.inside") {
		noSession := env.GetBool("sys.session.disable") || noSessionCmds(flow)
		if !noSession {
			statusWriter, sessionExisted, ok := core.SessionInit(cc, flow, env, self.sessionFileName, self.sessionStatusFileName)
			if !ok {
				return false
			}
			crossProcessInnerCall = sessionExisted
			cc.SetFlowStatusWriter(statusWriter)
		} else {
			core.SessionSetId(env)
		}
	}

	if !innerCall && !bootstrap {
		if !flow.TailModeCall && !verifyEnvOps(cc, flow, env) {
			return false
		}
		if !verifyOsDepCmds(cc, flow, env) {
			return false
		}
	}

	if !bootstrap && !crossProcessInnerCall {
		stackStepIn(caller, env)
	}
	if !self.executeFlow(cc, bootstrap, flow, env, masks, input) {
		return false
	}
	if !bootstrap && !crossProcessInnerCall {
		stackStepOut(caller, self.callerNameEntry, env)
	}
	return true
}

func (self *Executor) executeFlow(
	cc *core.Cli,
	bootstrap bool,
	flow *core.ParsedCmds,
	env *core.Env,
	masks []*core.ExecuteMask,
	input []string) bool {

	breakAtNext := false
	for i := 0; i < len(flow.Cmds); i++ {
		cmd := flow.Cmds[i]
		var succeeded bool
		var mask *core.ExecuteMask
		if i < len(masks) {
			mask = masks[i]
		}
		cmdEnv := env
		if cc.ForestMode.AtForestTopLvl(env) {
			cmdEnv = env.Clone()
		}
		i, succeeded, breakAtNext = self.executeCmd(cc, bootstrap, cmd, cmdEnv, mask, flow, i,
			breakAtNext, i+1 == len(flow.Cmds))
		if !succeeded {
			return false
		}
	}
	if cc.ForestMode.AtForestTopLvl(env) {
		cc.ForestMode.Pop(env)
	}
	return true
}

func (self *Executor) executeCmd(
	cc *core.Cli,
	bootstrap bool,
	cmd core.ParsedCmd,
	env *core.Env,
	mask *core.ExecuteMask,
	flow *core.ParsedCmds,
	currCmdIdx int,
	breakByPrev bool,
	lastCmdInFlow bool) (newCurrCmdIdx int, succeeded bool, breakAtNext bool) {

	// This is a fake apply just for calculate sys args, the env is a clone
	cmdEnv, argv := cmd.ApplyMappingGenEnvAndArgv(
		env.Clone(), cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep, env.GetInt("sys.stack-depth"))
	sysArgv := cmdEnv.GetSysArgv(cmd.Path(), cc.Cmds.Strs.PathSep)

	ln := cc.Screen.OutputNum()

	stackLines := display.PrintCmdStack(bootstrap, cc.Screen, cmd, mask,
		cmdEnv, flow.Cmds, currCmdIdx, cc.Cmds.Strs, cc.BgTasks, flow.TailModeCall)

	showStack := func() {
		display.RenderCmdStack(stackLines, cmdEnv, cc.Screen)
	}

	var width int
	if stackLines.Display {
		width = display.RenderCmdStack(stackLines, cmdEnv, cc.Screen)
	}

	last := cmd.LastCmdNode()

	if !sysArgv.IsDelay() {
		bpa := tryWaitSecAndStepByStepAndBreakBefore(cc, env, cmd, mask, breakByPrev, lastCmdInFlow, bootstrap, showStack)
		if bpa == BPASkip {
			mask = copyMask(last.DisplayPath(), mask)
			mask.ExecPolicy = core.ExecPolicySkip
			breakAtNext = true
		} else if bpa == BPAStepOver {
			breakAtNext = true
		} else if bpa != BPAContinue {
			return
		}
	} else {
		// TODO: maybe use env to pass the breaking status for all cases will be better?
		breakAtNext = breakByPrev
	}

	start := time.Now()
	if last != nil {
		if last.IsNoExecutableCmd() {
			display.PrintEmptyDirCmdHint(cc.Screen, env, cmd)
			newCurrCmdIdx, succeeded = currCmdIdx, true
		} else {
			if !sysArgv.IsDelay() {
				// This cmdEnv is different from env, it included values from 'val2env' and 'arg2env'
				cmdEnv, argv = cmd.ApplyMappingGenEnvAndArgv(
					env, cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep, cmdEnv.GetInt("sys.stack-depth"))
				if width > 0 {
					cmdEnv.SetInt("display.executor.displayed", cmdEnv.GetInt("sys.stack-depth"))
				}
				tryBreakInsideFileNFlowWrap := func(cc *core.Cli, env *core.Env, cmd *core.Cmd) bool {
					return tryBreakInsideFileNFlow(cc, env, cmd, breakByPrev, showStack)
				}
				newCurrCmdIdx, succeeded = last.Execute(argv, sysArgv, cc, cmdEnv, mask, flow, currCmdIdx, tryBreakInsideFileNFlowWrap)
				cmdEnv.SetInt("display.executor.displayed", 0)
			} else {
				dur := sysArgv.GetDelayDuration()
				asyncCC := cc.CloneForAsyncExecuting(cmdEnv)
				var tid string
				tid, succeeded = asyncExecute(cc.Screen, sysArgv.GetDelayStr(),
					dur, last.Cmd(), argv, asyncCC, cmdEnv, mask, flow.CloneOne(currCmdIdx), 0)
				if cc.FlowStatus != nil {
					cc.FlowStatus.OnAsyncTaskSchedule(flow, currCmdIdx, env, tid)
				}
				newCurrCmdIdx = currCmdIdx
			}
		}
	} else {
		// Maybe it's an empty global-env definition
		newCurrCmdIdx, succeeded = currCmdIdx, true
	}
	elapsed := time.Now().Sub(start)

	if stackLines.Display {
		resultLines := display.PrintCmdResult(cc, bootstrap, cc.Screen, cmd,
			cmdEnv, succeeded, elapsed, flow.Cmds, currCmdIdx, cc.Cmds.Strs)
		display.RenderCmdResult(resultLines, cmdEnv, cc.Screen, width)
	} else if currCmdIdx < len(flow.Cmds)-1 && ln != cc.Screen.OutputNum() {
		last := flow.Cmds[len(flow.Cmds)-1]
		if last.LastCmd() != nil && !last.LastCmd().IsQuiet() {
			// Not pretty, disable for now
			//cc.Screen.Print("\n")
		}
	}

	if !sysArgv.IsDelay() {
		bpa := tryWaitSecAndBreakAfter(cc, env, cmd, bootstrap, lastCmdInFlow, showStack)
		if bpa == BPAStepOver {
			breakAtNext = true
		} else if bpa != BPAContinue {
			return
		}
	}
	return
}

func removeEmptyCmds(flow *core.ParsedCmds) {
	var cmds []core.ParsedCmd
	for _, cmd := range flow.Cmds {
		if !cmd.IsAllEmptySegments() {
			cmds = append(cmds, cmd)
		}
	}
	flow.Cmds = cmds
}

func checkTailModeCalls(flow *core.ParsedCmds) {
	if !flow.TailModeCall && flow.AttempTailModeCall && len(flow.Cmds) > 0 {
		panic(core.NewCmdError(flow.Cmds[0], "tail-mode call not support"))
	}
}

// TODO: trim the empty cmds at tail, like: '<cmd> : <cmd> ::'
func moveLastPriorityCmdToFront(
	flow []core.ParsedCmd) (reordered []core.ParsedCmd, doMove bool, tailModeCall bool, attempTailModeCall bool) {

	cnt := len(flow)
	if cnt <= 1 {
		return flow, false, false, false
	}

	last := flow[cnt-1]

	trim := func(flow []core.ParsedCmd) []core.ParsedCmd {
		for {
			if len(flow) == 0 {
				break
			}
			if !flow[len(flow)-1].IsAllEmptySegments() {
				break
			}
			flow = flow[:len(flow)-1]
		}
		return flow
	}

	if cnt >= 2 && len(flow[cnt-2].ParseResult.Input) == 0 {
		flow = trim(flow[:cnt-2])
	} else if last.IsPriority() {
		flow = trim(flow[:cnt-1])
	} else {
		return flow, false, false, false
	}

	flow, _, _, _ = moveLastPriorityCmdToFront(flow)

	last.TailMode = true
	reordered = append([]core.ParsedCmd{last}, flow...)
	attempTailModeCall = !last.IsPriority() && !last.AllowTailModeCall() && len(reordered) == 2
	tailModeCall = last.AllowTailModeCall() && len(reordered) <= 2
	doMove = true
	return
}

func verifyEnvOps(cc *core.Cli, flow *core.ParsedCmds, env *core.Env) bool {
	if len(flow.Cmds) == 0 {
		return true
	}
	if allowCheckEnvOpsFail(flow) {
		return true
	}
	checker := &core.EnvOpsChecker{}
	result := []core.EnvOpsCheckResult{}
	env = env.Clone()
	core.CheckEnvOps(cc, flow, env, checker, true, builtin.EnvOpCmds(), &result)
	if len(result) == 0 {
		return true
	}
	display.DumpEnvOpsCheckResult(cc.Screen, flow.Cmds, env, result, cc.Cmds.Strs.PathSep)
	return false
}

func verifyOsDepCmds(cc *core.Cli, flow *core.ParsedCmds, env *core.Env) bool {
	deps := core.Depends{}
	env = env.Clone()
	core.CollectDepends(cc, env, flow, 0, deps, true, builtin.EnvOpCmds())
	screen := display.NewCacheScreen()
	hasMissedOsCmds := display.DumpDepends(screen, env, deps)
	if hasMissedOsCmds {
		screen.WriteTo(cc.Screen)
		return false
	}
	return true
}

// Borrow abbrs from cmds to env
func useCmdsAbbrs(abbrs *core.EnvAbbrs, cmds *core.CmdTree) {
	if cmds == nil {
		return
	}
	for _, subName := range cmds.SubNames() {
		subAbbrs := cmds.SubAbbrs(subName)
		subEnv := abbrs.GetSub(subName)
		if subEnv == nil {
			subEnv = abbrs.AddSub(subName, subAbbrs...)
		} else {
			abbrs.AddSubAbbrs(subName, subAbbrs...)
		}
		subTree := cmds.GetSub(subName)
		useCmdsAbbrs(subEnv, subTree)
	}
}

func useEnvAbbrs(abbrs *core.EnvAbbrs, env *core.Env, sep string) {
	for k, _ := range env.Flatten(true, nil, true) {
		curr := abbrs
		for _, seg := range strings.Split(k, sep) {
			curr = curr.GetOrAddSub(seg)
		}
	}
}

func stackStepIn(caller string, env *core.Env) {
	env = env.GetLayer(core.EnvLayerSession)
	env.PlusInt("sys.stack-depth", 1)
	sep := env.GetRaw("strs.list-sep")
	stack := env.GetRaw("sys.stack")
	if len(stack) == 0 {
		env.Set("sys.stack", caller)
	} else {
		env.Set("sys.stack", stack+sep+caller)
	}
}

func stackStepOut(caller string, callerNameEntry string, env *core.Env) {
	env = env.GetLayer(core.EnvLayerSession)
	env.PlusInt("sys.stack-depth", -1)
	sep := env.GetRaw("strs.list-sep")
	stack := env.GetRaw("sys.stack")
	if !strings.HasSuffix(stack, sep+caller) {
		if stack == callerNameEntry {
			stack = ""
		} else {
			fields := strings.Split(stack, sep)
			if len(fields) != 1 || fields[0] != caller {
				panic(fmt.Errorf("stack string not match when stepping out from '%s', stack: '%s'",
					caller, stack))
			}
		}
	} else {
		stack = stack[0 : len(stack)-len(sep)-len(caller)]
	}
	env.Set("sys.stack", stack)
}

func asyncExecute(
	screen core.Screen,
	durStr string,
	dur time.Duration,
	cic *core.Cmd,
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	mask *core.ExecuteMask,
	flow *core.ParsedCmds,
	currCmdIdx int) (tid string, scheduled bool) {

	if env.GetBool("sys.in-bg-task") {
		tid := utils.GoRoutineIdStr()
		panic(fmt.Errorf("can't delay a command when not in main thread, current thread: %s", tid))
	}

	tidChan := make(chan string, 1)

	// TODO: fixme, do real clone for ready-only instances
	go func(
		dur time.Duration,
		argv core.ArgVals,
		cc *core.Cli,
		env *core.Env,
		flow *core.ParsedCmds,
		currCmdIdx int) {

		cmd := flow.Cmds[currCmdIdx]
		name := cmd.DisplayPath(cc.Cmds.Strs.PathSep, true)
		tid := utils.GoRoutineIdStr()

		sessionDir := env.GetRaw("session")
		sessionDir = filepath.Join(sessionDir, tid)
		os.MkdirAll(sessionDir, os.ModePerm)

		statusFileName := env.GetRaw("strs.session-status-file")
		statusPath := filepath.Join(sessionDir, statusFileName)
		cc.SetFlowStatusWriter(core.NewExecutingFlow(statusPath, flow, env))

		bgSessionEnv := env.GetLayer(core.EnvLayerSession)
		bgSessionEnv.Set("session", sessionDir)
		bgSessionEnv.SetBool("display.one-cmd", true)
		bgSessionEnv.SetBool("sys.in-bg-task", true)

		clearBreakPointStatusInEnv(bgSessionEnv)

		task := cc.BgTasks.GetOrAddTask(tid, name, cc.Screen.(*core.BgTaskScreen).GetBgStdout())
		tidChan <- tid

		time.Sleep(dur)

		task.OnStart()

		defer func() {
			var err error
			if !env.GetBool("sys.panic.recover") {
				// TODO: not sure how to handle if config is not-recover
			} else {
				if r := recover(); r != nil {
					err = r.(error)
				}
			}
			task.OnFinish(err)
		}()

		stackLines := display.PrintCmdStack(false, cc.Screen, cmd, mask,
			env, flow.Cmds, currCmdIdx, cc.Cmds.Strs, nil, false)
		var width int
		if stackLines.Display {
			width = display.RenderCmdStack(stackLines, env, cc.Screen)
		}

		if width > 0 {
			env.SetInt("display.executor.displayed", env.GetInt("sys.stack-depth"))
		}
		start := time.Now()
		_, ok := cic.Execute(argv, cc, env, mask, flow, currCmdIdx, nil)
		elapsed := time.Now().Sub(start)
		env.SetInt("display.executor.displayed", 0)
		if !ok {
			// Should already panic inside cmd.Execute
			panic(fmt.Errorf("delay-command fail, thread: %s", tid))
		}

		if stackLines.Display {
			resultLines := display.PrintCmdResult(cc, false, cc.Screen, cmd,
				env, ok, elapsed, flow.Cmds, currCmdIdx, cc.Cmds.Strs)
			display.RenderCmdResult(resultLines, env, cc.Screen, width)
		}
		cc.FlowStatus.OnFlowFinish(env, ok)

	}(dur, argv, cc, env, flow, currCmdIdx)

	tid = <-tidChan
	screen.Print(display.ColorExplain("(current command scheduled to thread "+tid+")\n", env))
	return tid, true
}

func copyMask(cmd string, mask *core.ExecuteMask) *core.ExecuteMask {
	if mask != nil {
		mask = mask.Copy()
	} else {
		mask = core.NewExecuteMask(cmd)
	}
	return mask
}
