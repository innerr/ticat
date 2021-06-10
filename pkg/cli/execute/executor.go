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
	funcs           []ExecFunc
	sessionFileName string
}

func NewExecutor(sessionFileName string) *Executor {
	return &Executor{
		[]ExecFunc{
			// TODO: implement and add functions: flowFlatten, mockModInject, stepByStepInject
			handleParseError,
			filterEmptyCmdsAndReorderByPriority,
			verifyEnvOps,
		},
		sessionFileName,
	}
}

func (self *Executor) Run(cc *core.Cli, bootstrap string, input ...string) bool {
	overWriteBootstrap := cc.GlobalEnv.Get("sys.bootstrap").Raw
	if len(overWriteBootstrap) != 0 {
		bootstrap = overWriteBootstrap
	}
	if !self.execute(cc, true, bootstrap) {
		return false
	}
	return self.execute(cc, false, input...)
}

// Implement core.Executor
func (self *Executor) Execute(cc *core.Cli, input ...string) bool {
	return self.execute(cc, false, input...)
}

func (self *Executor) execute(cc *core.Cli, bootstrap bool, input ...string) bool {
	useCmdsAbbrs(cc.EnvAbbrs, cc.Cmds)
	env := cc.GlobalEnv.GetLayer(core.EnvLayerSession)

	if len(input) == 0 {
		display.PrintGlobalHelp(cc.Screen, env)
		return true
	}

	if !bootstrap {
		useEnvAbbrs(cc.EnvAbbrs, env, cc.Cmds.Strs.EnvPathSep)
	}
	flow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, input...)
	if flow.GlobalEnv != nil {
		flow.GlobalEnv.WriteNotArgTo(env, cc.Cmds.Strs.EnvValDelAllMark)
	}
	for _, function := range self.funcs {
		if !function(cc, flow, env) {
			return false
		}
	}

	if !bootstrap && !self.sessionInit(cc, flow, env) {
		return false
	}

	if !bootstrap {
		env.PlusInt("sys.stack-depth", 1)
	}
	if !self.executeFlow(cc, bootstrap, flow, env, input) {
		return false
	}
	if !bootstrap {
		env.PlusInt("sys.stack-depth", -1)
	}

	if !self.sessionFinish(cc, flow, env) {
		return false
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
	cmdEnv := cmd.GenEnv(env, cc.Cmds.Strs.EnvValDelAllMark)
	argv := cmdEnv.GetArgv(cmd.Path(), cc.Cmds.Strs.PathSep, cmd.Args())

	ln := cc.Screen.OutputNum()

	stackLines := display.PrintCmdStack(bootstrap, cc.Screen, cmd,
		cmdEnv, flow.Cmds, currCmdIdx, cc.Cmds.Strs)
	var width int
	if stackLines.Display {
		width = display.RenderCmdStack(stackLines, cmdEnv, cc.Screen)
		tryDelayAndStepByStep(cc, env)
	}

	last := cmd.LastCmdNode()
	start := time.Now()
	if last != nil {
		if last.IsEmptyDirCmd() {
			name := cmd.DisplayPath(cc.Cmds.Strs.PathSep, true)
			if !last.HasSub() {
				display.PrintTipTitle(cc.Screen, env,
					fmt.Sprintf("'%v' is not executable and has no commands on this branch.", name))
			} else {
				display.PrintTipTitle(cc.Screen, env,
					fmt.Sprintf("'%v' is not executable, but has commands on this branch:", name))
				display.DumpAllCmds(last, cc.Screen, true, 4, true, true)
			}
			newCurrCmdIdx, succeeded = currCmdIdx, true
		} else {
			newCurrCmdIdx, succeeded = last.Execute(argv, cc, cmdEnv, flow, currCmdIdx)
		}
	} else {
		newCurrCmdIdx, succeeded = currCmdIdx, false
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

func handleParseError(
	cc *core.Cli,
	flow *core.ParsedCmds,
	env *core.Env) bool {

	isSearch, isLess, isMore := isEndWithSearchCmd(flow)
	if isMore || isLess {
		return true
	}

	for _, cmd := range flow.Cmds {
		if cmd.ParseError.Error == nil {
			continue
		}
		// TODO: better handling: sub flow parse failed
		/*
			stackDepth := env.GetInt("sys.stack-depth")
			if stackDepth > 0 {
				panic(cmd.ParseError.Error)
			}
		*/

		input := cmd.ParseError.Input
		inputStr := strings.Join(input, " ")

		switch cmd.ParseError.Error.(type) {
		case core.ParseErrExpectNoArg:
			title := "[" + cmd.DisplayPath(cc.Cmds.Strs.PathSep, true) + "] doesn't have args."
			return display.PrintFindResultByParseError(cc, cmd, env, title)
		case core.ParseErrEnv:
			helpStr := []string{
				"[" + cmd.DisplayPath(cc.Cmds.Strs.PathSep, true) + "] parse env failed, " +
					"'" + inputStr + "' is not valid input.",
				"", "env setting examples:", "",
			}
			helpStr = append(helpStr, display.SuggestStrsEnvSetting(env.GetRaw("strs.self-name"))...)
			helpStr = append(helpStr, "")
			display.PrintTipTitle(cc.Screen, env, helpStr...)
		case core.ParseErrExpectArgs:
			return display.PrintCmdByParseError(cc, cmd, env)
		case core.ParseErrExpectCmd:
			return display.PrintSubCmdByParseError(cc, flow, cmd, env, isSearch, isMore)
		default:
			return display.PrintFindResultByParseError(cc, cmd, env, "")
		}
	}
	return true
}

// 1. remove the cmds only have cmd-level env definication but have no executable
// 2. move priority cmds to the head
// TODO: sort commands by priority-value, not just a bool flag, so '+' '-' can have the top priority
// TODO: move to core
func filterEmptyCmdsAndReorderByPriority(
	cc *core.Cli,
	flow *core.ParsedCmds,
	_ *core.Env) bool {

	// TODO: clean up this code
	//notFilterEmpty := doNotFilterEmptyCmds(flow)

	var unfiltered []core.ParsedCmd
	unfilteredGlobalCmdIdx := -1
	var priorities []core.ParsedCmd
	prioritiesGlobalCmdIdx := -1

	for i, cmd := range flow.Cmds {
		/*
			if cmd.IsAllEmptySegments() {
				if notFilterEmpty {
					unfiltered = append(unfiltered, cmd)
					if i == flow.GlobalCmdIdx {
						unfilteredGlobalCmdIdx = len(unfiltered) - 1
					}
				}
				continue
			}
		*/
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
		core.LoadEnvFromFile(env, sessionPath, cc.Cmds.Strs.ProtoSep)
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
		// TODO: warning
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

func (self *Executor) sessionFinish(cc *core.Cli, flow *core.ParsedCmds, env *core.Env) bool {
	sessionDir := env.GetRaw("session")
	if len(sessionDir) == 0 {
		return true
	}
	kvSep := env.GetRaw("strs.proto-sep")
	path := filepath.Join(sessionDir, self.sessionFileName)
	core.SaveEnvToFile(env, path, kvSep)
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
	core.CheckEnvOps(cc, flow, env, checker, true, &result)
	if len(result) == 0 {
		return true
	}
	display.DumpEnvOpsCheckResult(cc.Screen, env, result, cc.Cmds.Strs.PathSep)
	return false
}

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

func tryDelayAndStepByStep(cc *core.Cli, env *core.Env) {
	delaySec := env.GetInt("sys.execute-delay-sec")
	if delaySec > 0 {
		for i := 0; i < delaySec; i++ {
			time.Sleep(time.Second)
			cc.Screen.Print(".")
		}
		cc.Screen.Print("\n")
	}
	if env.GetBool("sys.step-by-step") {
		cc.Screen.Print("[confirm] press enter to run")
		utils.UserConfirm()
	}
}
