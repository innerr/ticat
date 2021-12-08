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

func (self *Executor) Run(cc *core.Cli, bootstrap string, input ...string) bool {
	overWriteBootstrap := cc.GlobalEnv.Get("sys.bootstrap").Raw
	if len(overWriteBootstrap) != 0 {
		bootstrap = overWriteBootstrap
	}
	if !self.execute(self.callerNameBootstrap, cc, true, false, bootstrap) {
		return false
	}
	ok := self.execute(self.callerNameEntry, cc, false, false, input...)
	builtin.WaitAllBgTasks(cc)
	if cc.FlowStatus != nil {
		cc.FlowStatus.OnFlowFinish()
	}
	return ok
}

// Implement core.Executor
func (self *Executor) Execute(caller string, cc *core.Cli, input ...string) bool {
	return self.execute(caller, cc, false, true, input...)
}

func (self *Executor) execute(caller string, cc *core.Cli, bootstrap bool, innerCall bool, input ...string) bool {
	if !innerCall && cc.GlobalEnv.GetBool("sys.env.use-cmd-abbrs") {
		useCmdsAbbrs(cc.EnvAbbrs, cc.Cmds)
	}
	env := cc.GlobalEnv.GetLayer(core.EnvLayerSession)

	if !innerCall && len(input) == 0 {
		display.PrintGlobalHelp(cc, env)
		return true
	}

	if !innerCall && !bootstrap {
		useEnvAbbrs(cc.EnvAbbrs, env, cc.Cmds.Strs.EnvPathSep)
	}
	flow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, input...)
	if flow.GlobalEnv != nil {
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

	if !allowParseError(flow) {
		isSearch := isStartWithSearchCmd(flow)
		if !display.HandleParseResult(cc, flow, env, isSearch) {
			return false
		}
	}

	display.PrintTolerableErrs(cc.Screen, env, cc.TolerableErrs)

	if !innerCall && !bootstrap && !noSessionCmds(flow) {
		statusWriter, ok := core.SessionInit(cc, flow, env, self.sessionFileName, self.sessionStatusFileName)
		if !ok {
			return false
		}
		cc.SetFlowStatusWriter(statusWriter)
	}

	if !innerCall && !bootstrap {
		if !flow.HasTailMode && !verifyEnvOps(cc, flow, env) {
			return false
		}
		if !flow.HasTailMode && !verifyOsDepCmds(cc, flow, env) {
			return false
		}
	}

	if !bootstrap {
		stackStepIn(caller, env)
	}
	if !self.executeFlow(cc, bootstrap, flow, env, input) {
		return false
	}
	if !bootstrap {
		stackStepOut(caller, self.callerNameEntry, env)
	}
	return true
}

func (self *Executor) executeFlow(
	cc *core.Cli,
	bootstrap bool,
	flow *core.ParsedCmds,
	env *core.Env,
	input []string) bool {

	for i := 0; i < len(flow.Cmds); i++ {
		cmd := flow.Cmds[i]
		var succeeded bool
		i, succeeded = self.executeCmd(cc, bootstrap, cmd, env, flow, i)
		if !succeeded {
			return false
		}
	}
	return true
}

func (self *Executor) executeCmd(
	cc *core.Cli,
	bootstrap bool,
	cmd core.ParsedCmd,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (newCurrCmdIdx int, succeeded bool) {

	// The env modifications from input will be popped out after a command is executed
	// But if a mod modified the env, the modifications stay in session level
	cmdEnv := cmd.GenCmdEnv(env, cc.Cmds.Strs.EnvValDelAllMark)

	ln := cc.Screen.OutputNum()

	stackLines := display.PrintCmdStack(bootstrap, cc.Screen, cmd,
		cmdEnv, flow.Cmds, currCmdIdx, cc.Cmds.Strs, cc.BgTasks, flow.TailModeCall)
	var width int
	if stackLines.Display {
		width = display.RenderCmdStack(stackLines, cmdEnv, cc.Screen)
		succeeded = tryDelayAndStepByStep(cc, env)
		if !succeeded {
			return
		}
	}

	last := cmd.LastCmdNode()
	start := time.Now()
	if last != nil {
		if last.IsNoExecutableCmd() {
			display.PrintEmptyDirCmdHint(cc.Screen, env, cmd)
			newCurrCmdIdx, succeeded = currCmdIdx, true
		} else {
			// This is a fake apply just for calculate sys args, the env is a clone
			cmdEnv, argv := cmd.ApplyMappingGenEnvAndArgv(
				env.Clone(), cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep)
			sysArgv := cmdEnv.GetSysArgv(cmd.Path(), cc.Cmds.Strs.PathSep)
			if !sysArgv.IsDelay() {
				// This cmdEnv is different from env, it included values from 'val2env' and 'arg2env'
				cmdEnv, argv = cmd.ApplyMappingGenEnvAndArgv(
					env, cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep)
				newCurrCmdIdx, succeeded = last.Execute(argv, sysArgv, cc, cmdEnv, flow, currCmdIdx)
			} else {
				dur := sysArgv.GetDelayDuration()
				asyncCC := cc.CloneForAsyncExecuting(cmdEnv)
				var tid string
				tid, succeeded = asyncExecute(cc.Screen, sysArgv.GetDelayStr(),
					dur, last.Cmd(), argv, asyncCC, cmdEnv, flow.CloneOne(currCmdIdx), 0)
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
		resultLines := display.PrintCmdResult(bootstrap, cc.Screen, cmd,
			cmdEnv, succeeded, elapsed, flow.Cmds, currCmdIdx, cc.Cmds.Strs)
		display.RenderCmdResult(resultLines, cmdEnv, cc.Screen, width)
	} else if currCmdIdx < len(flow.Cmds)-1 && ln != cc.Screen.OutputNum() {
		last := flow.Cmds[len(flow.Cmds)-1]
		if last.LastCmd() != nil && !last.LastCmd().IsQuiet() {
			cc.Screen.Print("\n")
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
	flow = append([]core.ParsedCmd{last}, flow...)
	attempTailModeCall = !last.IsPriority() && !last.AllowTailModeCall() && len(flow) == 2
	return flow, true, last.AllowTailModeCall() && len(flow) <= 2, attempTailModeCall
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

func tryDelayAndStepByStep(cc *core.Cli, env *core.Env) bool {
	delaySec := env.GetInt("sys.execute-delay-sec")
	if delaySec > 0 {
		for i := 0; i < delaySec; i++ {
			time.Sleep(time.Second)
			cc.Screen.Print(".")
		}
		cc.Screen.Print("\n")
	}
	if env.GetBool("sys.step-by-step") {
		cc.Screen.Print(display.ColorTip("[confirm]", env) + " type " +
			display.ColorWarn("'y'", env) + " and press enter:\n")
		return utils.UserConfirm()
	}
	return true
}

func stackStepIn(caller string, env *core.Env) {
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
	env.PlusInt("sys.stack-depth", -1)
	sep := env.GetRaw("strs.list-sep")
	stack := env.GetRaw("sys.stack")
	if !strings.HasSuffix(stack, sep+caller) {
		if stack == callerNameEntry {
			stack = ""
		} else {
			panic(fmt.Errorf("stack string not match when stepping out from '%s', stack: '%s'",
				caller, stack))
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

		envBgSession := env.GetLayer(core.EnvLayerSession)
		envBgSession.Set("session", sessionDir)
		envBgSession.SetBool("display.one-cmd", true)
		envBgSession.SetBool("sys.in-bg-task", true)

		task := cc.BgTasks.GetOrAddTask(tid, name, cc.Screen.(*core.BgTaskScreen).GetBgStdout())
		tidChan <- tid

		time.Sleep(dur)

		defer func() {
			if !cc.GlobalEnv.GetBool("sys.panic.recover") {
				return
			}
			if r := recover(); r != nil {
				display.PrintError(cc, cc.GlobalEnv, r.(error))
				os.Exit(-1)
			}
		}()

		task.OnStart()

		stackLines := display.PrintCmdStack(false, cc.Screen, cmd,
			env, flow.Cmds, currCmdIdx, cc.Cmds.Strs, nil, false)
		var width int
		if stackLines.Display {
			width = display.RenderCmdStack(stackLines, env, cc.Screen)
		}

		start := time.Now()
		_, ok := cic.Execute(argv, cc, env, flow, currCmdIdx)
		elapsed := time.Now().Sub(start)
		if !ok {
			// Should already panic inside cmd.Execute
			panic(fmt.Errorf("delay-command fail, thread: %s", tid))
		}

		if stackLines.Display {
			resultLines := display.PrintCmdResult(false, cc.Screen, cmd,
				env, ok, elapsed, flow.Cmds, currCmdIdx, cc.Cmds.Strs)
			display.RenderCmdResult(resultLines, env, cc.Screen, width)
		}
		cc.FlowStatus.OnFlowFinish()
		task.OnFinish()

	}(dur, argv, cc, env, flow, currCmdIdx)

	tid = <-tidChan
	screen.Print(display.ColorExplain("(current command scheduled to thread "+tid+")\n", env))
	return tid, true
}
