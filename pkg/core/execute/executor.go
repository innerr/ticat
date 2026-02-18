package execute

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/innerr/ticat/pkg/cli/display"
	"github.com/innerr/ticat/pkg/core/model"
	"github.com/innerr/ticat/pkg/mods/builtin"
	"github.com/innerr/ticat/pkg/utils"
)

type ExecFunc func(cc *model.Cli, flow *model.ParsedCmds, env *model.Env) bool

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

func (self *Executor) Clone() model.Executor {
	return NewExecutor(
		self.sessionFileName,
		self.sessionStatusFileName,
		self.callerNameBootstrap,
		self.callerNameEntry)
}

func (self *Executor) Run(cc *model.Cli, env *model.Env, bootstrap string, input ...string) bool {
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
		errs = builtin.WaitBgTasks(cc, env, false)
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

// Implement model.Executor
func (self *Executor) Execute(caller string, innerCall bool, cc *model.Cli, env *model.Env,
	masks []*model.ExecuteMask, input ...string) bool {
	return self.execute(caller, cc, env, masks, false, innerCall, input...)
}

func (self *Executor) execute(caller string, cc *model.Cli, env *model.Env, masks []*model.ExecuteMask,
	bootstrap bool, innerCall bool, input ...string) bool {

	if !innerCall && env.GetBool("sys.env.use-cmd-abbrs") {
		useCmdsAbbrs(cc.EnvAbbrs, cc.Cmds)
	}

	if !innerCall && emptyInput(input...) {
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
		env = env.GetOneOfLayers(model.EnvLayerSubFlow, model.EnvLayerSession)
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
	if len(flow.Cmds) == 0 {
		return true
	}

	if !innerCall && !bootstrap {
		checkTailModeCalls(flow)
	}

	if !innerCall {
		_ = cc.Blender.Invoke(cc, env, flow)
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
			statusWriter, sessionExisted, ok := model.SessionInit(cc, flow, env, self.sessionFileName, self.sessionStatusFileName)
			if !ok {
				return false
			}
			crossProcessInnerCall = sessionExisted
			cc.SetFlowStatusWriter(statusWriter)
		} else {
			model.SessionSetId(env)
		}
	}

	if !innerCall && !bootstrap {
		if !flow.TailModeCall && !verifyEnvOps(cc, flow, env) {
			return false
		}
		var firstCmd *model.Cmd
		if len(flow.Cmds) != 0 {
			firstCmd = flow.Cmds[0].LastCmd()
		}
		if firstCmd == nil || !firstCmd.ShouldIgnoreFollowingDeps() {
			if !verifyOsDepCmds(cc, flow, env) {
				return false
			}
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
	cc *model.Cli,
	bootstrap bool,
	flow *model.ParsedCmds,
	env *model.Env,
	masks []*model.ExecuteMask,
	input []string) bool {

	breakAtNext := false
	for i := 0; i < len(flow.Cmds); i++ {
		cmd := flow.Cmds[i]
		var err error
		var mask *model.ExecuteMask
		if i < len(masks) {
			mask = masks[i]
		}
		cmdEnv := env
		if cc.ForestMode.AtForestTopLvl(env) {
			cmdEnv = env.Clone()
		}
		i, err, breakAtNext = self.executeCmd(cc, bootstrap, cmd, cmdEnv, mask, flow, i,
			breakAtNext, i+1 == len(flow.Cmds))
		if err != nil {
			return false
		}
	}
	if cc.ForestMode.AtForestTopLvl(env) {
		cc.ForestMode.Pop(env)
	}
	return true
}

func (self *Executor) executeCmd(
	cc *model.Cli,
	bootstrap bool,
	cmd model.ParsedCmd,
	env *model.Env,
	mask *model.ExecuteMask,
	flow *model.ParsedCmds,
	currCmdIdx int,
	breakByPrev bool,
	lastCmdInFlow bool) (newCurrCmdIdx int, err error, breakAtNext bool) {

	// This is a fake apply just for calculate sys args, the env is a clone
	cmdEnv, argv := cmd.ApplyMappingGenEnvAndArgv(
		env.Clone(), cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep, env.GetInt("sys.stack-depth"))
	sysArgv := cmdEnv.GetSysArgv(cmd.Path(), cc.Cmds.Strs.PathSep)

	ln := cc.Screen.OutputtedLines()

	stackLines := display.PrintCmdStack(bootstrap, cc.Screen, cmd, mask,
		cmdEnv, cc.EnvKeysInfo, flow.Cmds, currCmdIdx, cc.Cmds.Strs, cc.BgTasks, flow.TailModeCall)

	showStack := func() {
		display.RenderCmdStack(stackLines, cmdEnv, cc.Screen)
	}

	var width int
	if stackLines.Display {
		width = display.RenderCmdStack(stackLines, cmdEnv, cc.Screen)
	}

	last := cmd.LastCmdNode()

	// TODO: change name `IsDelay`, maybe `IsBackgroundCmd`
	if !sysArgv.IsDelay() {
		bpa := tryWaitSecAndBreakBefore(cc, env, cmd, mask, breakByPrev, lastCmdInFlow, bootstrap, showStack)
		if bpa == BPASkip {
			mask = copyMask(last.DisplayPath(), mask)
			mask.ExecPolicy = model.ExecPolicySkip
			// Use env `sys.breakpoint.at-next` to do skip to affect subflow
			env.GetLayer(model.EnvLayerSession).SetBool("sys.breakpoint.at-next", true)
		} else if bpa == BPAStepOver {
			// Use `breakAtNext = true` to do step-over to not affect subflow
			breakAtNext = true
		} else if bpa != BPAContinue {
			return
		}
	} else {
		// TODO: need to do anything for bg cmd?
		_ = true
	}

	start := time.Now()
	if last != nil {
		if last.IsNoExecutableCmd() {
			display.PrintEmptyDirCmdHint(cc.Screen, env, cmd)
			newCurrCmdIdx, err = currCmdIdx, nil
		} else {
			if !sysArgv.IsDelay() {
				// This cmdEnv is different from env, it included values from 'val2env' and 'arg2env'
				cmdEnv, argv = cmd.ApplyMappingGenEnvAndArgv(
					env, cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep, cmdEnv.GetInt("sys.stack-depth"))
				if width > 0 {
					cmdEnv.SetInt("display.executor.displayed", cmdEnv.GetInt("sys.stack-depth"))
				}
				tryBreakInsideFileNFlowWrap := func(cc *model.Cli, env *model.Env, cmd *model.Cmd) bool {
					return tryBreakInsideFileNFlow(cc, env, cmd, breakByPrev, showStack)
				}
				newCurrCmdIdx, err = last.Execute(argv, sysArgv, cc, cmdEnv, mask, flow, currCmdIdx, tryBreakInsideFileNFlowWrap)
				cmdEnv.SetInt("display.executor.displayed", 0)
			} else {
				if sysArgv.IsDelayEnvEarlyApply() {
					cmd.ApplyMappingGenEnvAndArgv(
						env, cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep, cmdEnv.GetInt("sys.stack-depth"))
				}
				dur, durErr := sysArgv.GetDelayDuration()
				if durErr != nil {
					err = durErr
					newCurrCmdIdx = currCmdIdx
				} else {
					asyncCC := cc.CloneForAsyncExecuting(cmdEnv)
					var tid string
					var asyncSucceeded bool
					tid, asyncSucceeded = asyncExecute(cc.Screen, sysArgv.GetDelayStr(), sysArgv.AllowError(),
						dur, last.Cmd(), argv, asyncCC, cmdEnv.Clone(), mask, flow.CloneOne(currCmdIdx), 0)
					if !asyncSucceeded {
						err = fmt.Errorf("async execute failed")
					}
					if cc.FlowStatus != nil {
						cc.FlowStatus.OnAsyncTaskSchedule(flow, currCmdIdx, env, tid)
					}
					newCurrCmdIdx = currCmdIdx
				}
			}
		}
	} else {
		// Maybe it's an empty global-env definition
		newCurrCmdIdx, err = currCmdIdx, nil
	}
	elapsed := time.Since(start)

	if stackLines.Display {
		resultLines := display.PrintCmdResult(cc, bootstrap, cc.Screen, cmd,
			cmdEnv, err == nil, elapsed, flow.Cmds, currCmdIdx, cc.Cmds.Strs)
		display.RenderCmdResult(resultLines, cmdEnv, cc.Screen, width)
	} else if currCmdIdx < len(flow.Cmds)-1 && ln != cc.Screen.OutputtedLines() {
		last := flow.Cmds[len(flow.Cmds)-1]
		if last.LastCmd() != nil && !last.LastCmd().IsQuiet() {
			// Not pretty, disable for now
			//cc.Screen.Print("\n")
			_ = true
		}
	}

	if !sysArgv.IsDelay() {
		bpa := tryWaitSecAndBreakAfter(cc, env, cmd, bootstrap, lastCmdInFlow, showStack)
		if bpa == BPAStepOver {
			// Use `breakAtNext = true` to do step-over to not affect subflow
			breakAtNext = true
		} else if bpa != BPAContinue {
			return
		}
	}
	return
}

func removeEmptyCmds(flow *model.ParsedCmds) {
	var cmds []model.ParsedCmd
	for _, cmd := range flow.Cmds {
		if !cmd.IsAllEmptySegments() {
			cmds = append(cmds, cmd)
		}
	}
	flow.Cmds = cmds
}

func checkTailModeCalls(flow *model.ParsedCmds) {
	if !flow.TailModeCall && flow.AttempTailModeCall && len(flow.Cmds) > 0 {
		// PANIC: Runtime error - tail-mode call not supported
		panic(model.NewCmdError(flow.Cmds[0], "tail-mode call not support"))
	}
}

// TODO: trim the empty cmds at tail, like: '<cmd> : <cmd> ::'
func moveLastPriorityCmdToFront(
	flow []model.ParsedCmd) (reordered []model.ParsedCmd, doMove bool, tailModeCall bool, attempTailModeCall bool) {

	cnt := len(flow)
	if cnt <= 1 {
		return flow, false, false, false
	}

	last := flow[cnt-1]

	trim := func(flow []model.ParsedCmd) []model.ParsedCmd {
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
	reordered = append([]model.ParsedCmd{last}, flow...)
	attempTailModeCall = !last.IsPriority() && !last.AllowTailModeCall() && len(reordered) == 2
	tailModeCall = last.AllowTailModeCall() && len(reordered) <= 2
	doMove = true
	return
}

func verifyEnvOps(cc *model.Cli, flow *model.ParsedCmds, env *model.Env) bool {
	if len(flow.Cmds) == 0 {
		return true
	}
	if allowCheckEnvOpsFail(flow) {
		return true
	}
	checker := &model.EnvOpsChecker{}
	result := []model.EnvOpsCheckResult{}
	env = env.Clone()
	model.CheckEnvOps(cc, flow, env, checker, true, builtin.EnvOpCmds(), &result)
	if len(result) == 0 {
		return true
	}
	display.DumpEnvOpsCheckResult(cc.Screen, flow.Cmds, env, result, cc.Cmds.Strs.PathSep)
	return false
}

func verifyOsDepCmds(cc *model.Cli, flow *model.ParsedCmds, env *model.Env) bool {
	deps := model.Depends{}
	env = env.Clone()
	model.CollectDepends(cc, env, flow, 0, deps, true, builtin.EnvOpCmds())
	screen := display.NewCacheScreen()
	hasMissedOsCmds := display.DumpDepends(screen, env, deps)
	if hasMissedOsCmds {
		screen.WriteTo(cc.Screen)
		return false
	}
	return true
}

// Borrow abbrs from cmds to env
func useCmdsAbbrs(abbrs *model.EnvAbbrs, cmds *model.CmdTree) {
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

func useEnvAbbrs(abbrs *model.EnvAbbrs, env *model.Env, sep string) {
	for k := range env.Flatten(true, nil, true) {
		curr := abbrs
		for _, seg := range strings.Split(k, sep) {
			curr = curr.GetOrAddSub(seg)
		}
	}
}

func stackStepIn(caller string, env *model.Env) {
	env = env.GetLayer(model.EnvLayerSession)
	env.PlusInt("sys.stack-depth", 1)
	sep := env.GetRaw("strs.list-sep")
	stack := env.GetRaw("sys.stack")
	if len(stack) == 0 {
		env.Set("sys.stack", caller)
	} else {
		env.Set("sys.stack", stack+sep+caller)
	}
}

func stackStepOut(caller string, callerNameEntry string, env *model.Env) {
	env = env.GetLayer(model.EnvLayerSession)
	env.PlusInt("sys.stack-depth", -1)
	sep := env.GetRaw("strs.list-sep")
	stack := env.GetRaw("sys.stack")
	if !strings.HasSuffix(stack, sep+caller) {
		if stack == callerNameEntry {
			stack = ""
		} else {
			fields := strings.Split(stack, sep)
			if len(fields) != 1 || fields[0] != caller {
				// PANIC: Programming error - stack string not match when stepping out
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
	screen model.Screen,
	durStr string,
	allowError bool,
	dur time.Duration,
	cic *model.Cmd,
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	mask *model.ExecuteMask,
	flow *model.ParsedCmds,
	currCmdIdx int) (tid string, scheduled bool) {

	if env.GetBool("sys.in-bg-task") {
		tid := utils.GoRoutineIdStr()
		// PANIC: Runtime error - can't delay a command when not in main thread
		panic(fmt.Errorf("can't delay a command when not in main thread, current thread: %s", tid))
	}

	tidChan := make(chan string, 1)

	// TODO: fixme, do real clone for ready-only instances
	go func(
		dur time.Duration,
		argv model.ArgVals,
		cc *model.Cli,
		env *model.Env,
		flow *model.ParsedCmds,
		currCmdIdx int) {

		cmd := flow.Cmds[currCmdIdx]
		displayName := cmd.DisplayPath(cc.Cmds.Strs.PathSep, true)
		realName := cmd.LastCmdNode().DisplayPath()
		tid := utils.GoRoutineIdStr()

		sessionDir := env.GetRaw("session")
		sessionDir = filepath.Join(sessionDir, tid)
		err := os.MkdirAll(sessionDir, os.ModePerm)
		if err != nil {
			// PANIC: Runtime error - could not create session dir for bg task
			panic(fmt.Errorf("could not create session dir '%s' for bg task: %w", sessionDir, err))
		}

		statusFileName := env.GetRaw("strs.session-status-file")
		statusPath := filepath.Join(sessionDir, statusFileName)
		cc.SetFlowStatusWriter(model.NewExecutingFlow(statusPath, flow, env))

		bgSessionEnv := env.GetLayer(model.EnvLayerSession)
		bgSessionEnv.Set("session", sessionDir)
		bgSessionEnv.SetBool("display.one-cmd", true)
		bgSessionEnv.SetBool("sys.in-bg-task", true)

		clearBreakPointStatusInEnv(bgSessionEnv)

		task := cc.BgTasks.GetOrAddTask(tid, displayName, realName, cc.Screen.(*model.BgTaskScreen).GetBgStdout())
		tidChan <- tid

		time.Sleep(dur)

		task.OnStart()

		defer func() {
			var err error
			if !env.GetBool("sys.panic.recover") {
				// TODO: not sure how to handle if config is not-recover
			} else {
				if r := recover(); r != nil {
					var ok bool
					err, ok = r.(error)
					if !ok {
						err = fmt.Errorf("panic with non-error: %v", r)
					}
				}
			}
			task.OnFinish(err)
		}()

		stackLines := display.PrintCmdStack(false, cc.Screen, cmd, mask,
			env, cc.EnvKeysInfo, flow.Cmds, currCmdIdx, cc.Cmds.Strs, nil, false)
		var width int
		if stackLines.Display {
			width = display.RenderCmdStack(stackLines, env, cc.Screen)
		}

		if width > 0 {
			env.SetInt("display.executor.displayed", env.GetInt("sys.stack-depth"))
		}
		start := time.Now()
		_, asyncErr := cic.Execute(argv, cc, env, mask, flow, allowError, currCmdIdx, nil)
		elapsed := time.Since(start)
		env.SetInt("display.executor.displayed", 0)
		if asyncErr != nil {
			// PANIC: Runtime error - delay-command failed
			panic(fmt.Errorf("delay-command fail, thread: %s: %w", tid, asyncErr))
		}

		if stackLines.Display {
			resultLines := display.PrintCmdResult(cc, false, cc.Screen, cmd,
				env, asyncErr == nil, elapsed, flow.Cmds, currCmdIdx, cc.Cmds.Strs)
			display.RenderCmdResult(resultLines, env, cc.Screen, width)
		}
		cc.FlowStatus.OnFlowFinish(env, asyncErr == nil)

	}(dur, argv, cc, env, flow, currCmdIdx)

	tid = <-tidChan
	_ = screen.Print(display.ColorExplain("(current command scheduled to thread "+tid+")\n", env))
	return tid, true
}

func copyMask(cmd string, mask *model.ExecuteMask) *model.ExecuteMask {
	if mask != nil {
		mask = mask.Copy()
	} else {
		mask = model.NewExecuteMask(cmd)
	}
	return mask
}

func emptyInput(input ...string) bool {
	for _, it := range input {
		if len(it) != 0 {
			return false
		}
	}
	return true
}
