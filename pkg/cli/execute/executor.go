package execute

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
	"github.com/pingcap/ticat/pkg/utils"
)

type ExecFunc func(cc *core.Cli, flow *core.ParsedCmds, env *core.Env) bool

type Executor struct {
	sessionFileName     string
	callerNameBootstrap string
	callerNameEntry     string
}

func NewExecutor(
	sessionFileName string,
	callerNameBootstrap string,
	callerNameEntry string) *Executor {

	return &Executor{
		sessionFileName,
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
	return self.execute(self.callerNameEntry, cc, false, false, input...)
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
		reordered, moved := moveLastPriorityCmdToFront(flow.Cmds)
		flow.Cmds = reordered
		flow.TailMode = moved
		// TODO: this may not right if flow was changed in recursive process, but not big deal
		if moved && flow.GlobalCmdIdx == 0 {
			flow.GlobalCmdIdx = 1
		}
	}
	removeEmptyCmds(flow)

	if !flow.TailMode && !allowParseError(flow) {
		isSearch, isLess, isMore := isEndWithSearchCmd(flow)
		if !display.HandleParseResult(cc, flow, env, isSearch, isLess, isMore) {
			return false
		}
	}

	display.PrintTolerableErrs(cc.Screen, env, cc.TolerableErrs)

	if !innerCall && !bootstrap {
		if !flow.TailMode && !verifyEnvOps(cc, flow, env) {
			return false
		}
		if !flow.TailMode && !verifyOsDepCmds(cc, flow, env) {
			return false
		}
	}

	if !innerCall && !bootstrap && !self.sessionInit(cc, flow, env) {
		return false
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
		cmdEnv, flow.Cmds, currCmdIdx, cc.Cmds.Strs, flow.TailMode)
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
			// This cmdEnv is different, it included values from 'val2env' and 'arg2env'
			cmdEnv, argv := cmd.ApplyMappingGenEnvAndArgv(env, cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep)
			newCurrCmdIdx, succeeded = last.Execute(argv, cc, cmdEnv, flow, currCmdIdx)
		}
	} else {
		// Maybe a empty global-env definition
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

func moveLastPriorityCmdToFront(flow []core.ParsedCmd) (reordered []core.ParsedCmd, doMove bool) {
	cnt := len(flow)
	if cnt <= 1 {
		return flow, false
	}

	last := flow[cnt-1]

	trim := func(flow []core.ParsedCmd) []core.ParsedCmd {
		for {
			if len(flow) == 0 || !flow[len(flow)-1].IsAllEmptySegments() {
				break
			}
			flow = flow[:len(flow)-1]
		}
		return flow
	}

	if cnt >= 2 && len(flow[cnt-2].ParseResult.Input) == 0 {
		flow, _ = moveLastPriorityCmdToFront(trim(flow[:cnt-2]))
	} else if last.IsPriority() {
		flow, _ = moveLastPriorityCmdToFront(trim(flow[:cnt-1]))
	} else {
		return flow, false
	}

	flow = append([]core.ParsedCmd{last}, flow...)
	return flow, true
}

// TODO: remove this, not use anymore
func _moveLastPriorityCmdToFront(
	cc *core.Cli,
	flow *core.ParsedCmds,
	_ *core.Env) bool {

	for i := len(flow.Cmds) - 1; i >= 0; i-- {
		cmd := flow.Cmds[i]
		if cmd.IsPriority() {
			if i == 0 {
				return true
			}
			if i == len(flow.Cmds)-1 {
				flow.Cmds = append([]core.ParsedCmd{cmd}, flow.Cmds[:i]...)
			} else {
				tail := flow.Cmds[i+1:]
				flow.Cmds = append([]core.ParsedCmd{cmd}, flow.Cmds[:i]...)
				flow.Cmds = append(flow.Cmds, tail...)
			}
			if flow.GlobalCmdIdx == 0 {
				flow.GlobalCmdIdx = 1
			}
			return true
		}
	}
	return true
}

// TODO: remove this, not use anymore
// Move priority cmds to the front
func reorderByPriority(
	cc *core.Cli,
	flow *core.ParsedCmds,
	_ *core.Env) bool {

	var unfiltered []core.ParsedCmd
	unfilteredGlobalCmdIdx := -1
	var priorities []core.ParsedCmd
	prioritiesGlobalCmdIdx := -1

	for i, cmd := range flow.Cmds {
		if cmd.IsPriority() {
			priorities = append(priorities, cmd)
			if i == flow.GlobalCmdIdx {
				prioritiesGlobalCmdIdx = len(priorities) - 1
			}
		} else {
			unfiltered = append(unfiltered, cmd)
			if i == flow.GlobalCmdIdx {
				unfilteredGlobalCmdIdx = len(unfiltered) - 1
			}
		}
	}
	flow.Cmds = append(priorities, unfiltered...)
	if prioritiesGlobalCmdIdx >= 0 {
		flow.GlobalCmdIdx = prioritiesGlobalCmdIdx
	}
	if unfilteredGlobalCmdIdx >= 0 {
		flow.GlobalCmdIdx = len(priorities) + unfilteredGlobalCmdIdx
	}
	return true
}

func (self *Executor) sessionInit(cc *core.Cli, flow *core.ParsedCmds, env *core.Env) bool {
	sessionDir := env.GetRaw("session")
	sessionPath := filepath.Join(sessionDir, self.sessionFileName)
	if len(sessionDir) != 0 {
		core.LoadEnvFromFile(env, sessionPath, cc.Cmds.Strs.EnvKeyValSep)
		return true
	}

	sessionsRoot := env.GetRaw("sys.paths.sessions")
	if len(sessionsRoot) == 0 {
		cc.Screen.Print("[sessionInit] can't get sessions' root path\n")
		return false
	}

	os.MkdirAll(sessionsRoot, os.ModePerm)
	dirs, err := os.ReadDir(sessionsRoot)
	if err != nil {
		cc.Screen.Print(fmt.Sprintf("[sessionInit] can't read sessions' root path '%s'\n",
			sessionsRoot))
		return false
	}

	pid := fmt.Sprintf("%d", os.Getpid())

	for _, dir := range dirs {
		pid, err := strconv.Atoi(dir.Name())
		if err != nil {
			continue
		}
		err = syscall.Kill(pid, syscall.Signal(0))
		if err != nil && err == syscall.ESRCH {
			os.RemoveAll(filepath.Join(sessionsRoot, dir.Name()))
		}
	}

	sessionDir = filepath.Join(sessionsRoot, pid)
	err = os.MkdirAll(sessionDir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		cc.Screen.Print(fmt.Sprintf("[sessionInit] can't create session dir '%s'\n",
			sessionDir))
		return false
	}

	env.GetLayer(core.EnvLayerSession).Set("session", sessionDir)
	return true
}

// TODO: clean ti
// Seems not very useful, no user now.
func (self *Executor) sessionFinish(cc *core.Cli, flow *core.ParsedCmds, env *core.Env) bool {
	sessionDir := env.GetRaw("session")
	if len(sessionDir) == 0 {
		return true
	}
	path := filepath.Join(sessionDir, self.sessionFileName)
	core.SaveEnvToFile(env, path, cc.Cmds.Strs.EnvKeyValSep)
	return true
}

func verifyEnvOps(cc *core.Cli, flow *core.ParsedCmds, env *core.Env) bool {
	if len(flow.Cmds) == 0 {
		return true
	}
	allowFail := allowCheckEnvOpsFail(flow)
	if allowFail {
		return true
	}
	checker := &core.EnvOpsChecker{}
	result := []core.EnvOpsCheckResult{}
	env = env.Clone()
	core.CheckEnvOps(cc, flow, env, checker, true, &result)
	if len(result) == 0 {
		return true
	}
	display.DumpEnvOpsCheckResult(cc.Screen, flow.Cmds, env, result, cc.Cmds.Strs.PathSep)
	return false
}

func verifyOsDepCmds(cc *core.Cli, flow *core.ParsedCmds, env *core.Env) bool {
	deps := display.Depends{}
	env = env.Clone()
	display.CollectDepends(cc, env, flow.Cmds, deps, true)
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
