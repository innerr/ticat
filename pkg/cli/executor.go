package cli

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
)

type ExecFunc func(cc *core.Cli, flow *core.ParsedCmds) bool

type Executor struct {
	funcs    []ExecFunc
	pipeName string
}

func NewExecutor(pipeName string) *Executor {
	return &Executor{
		[]ExecFunc{
			// TODO: implement and add functions: flowFlatten, mockModInject, stepByStepInject
			filterEmptyCmdsAndReorderByPriority,
			checkEnvOps,
		},
		pipeName,
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
	return self.Execute(cc, false, input...)
}

func (self *Executor) Execute(cc *core.Cli, bootstrap bool, input ...string) bool {
	cc.GlobalEnv.PlusInt("sys.stack-depth", 1)
	if !self.execute(cc, bootstrap, input...) {
		return false
	}
	cc.GlobalEnv.PlusInt("sys.stack-depth", -1)
	return true
}

func (self *Executor) execute(cc *core.Cli, bootstrap bool, input ...string) bool {
	if len(input) == 0 {
		return true
	}
	useCmdsAbbrs(cc.EnvAbbrs, cc.Cmds)
	flow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, input...)
	for _, function := range self.funcs {
		if !function(cc, flow) {
			return false
		}
	}
	if !bootstrap && !self.sessionInit(cc, flow) {
		return false
	}
	return self.executeFlow(cc, bootstrap, flow, input)
}

func (self *Executor) executeFlow(
	cc *core.Cli,
	bootstrap bool,
	flow *core.ParsedCmds,
	input []string) bool {

	env := cc.GlobalEnv.GetLayer(core.EnvLayerSession)
	if flow.GlobalEnv != nil {
		flow.GlobalEnv.WriteNotArgTo(env, cc.Cmds.Strs.EnvValDelMark, cc.Cmds.Strs.EnvValDelAllMark)
	}
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
	cmdEnv := cmd.GenEnv(env, cc.Cmds.Strs.EnvValDelMark, cc.Cmds.Strs.EnvValDelAllMark)
	argv := cmdEnv.GetArgv(cmd.Path(), cc.Cmds.Strs.PathSep, cmd.Args())

	ln := cc.Screen.OutputNum()

	stackLines := display.PrintCmdStack(bootstrap, cc.Screen, cmd,
		cmdEnv, flow.Cmds, currCmdIdx, cc.Cmds.Strs)
	if stackLines.Display {
		display.RenderCmdStack(stackLines, cmdEnv, cc.Screen)
	}

	last := cmd[len(cmd)-1].Cmd.Cmd
	start := time.Now()
	if last != nil {
		newCurrCmdIdx, succeeded = last.Execute(argv, cc, cmdEnv, flow, currCmdIdx)
	} else {
		newCurrCmdIdx, succeeded = currCmdIdx, false
	}
	elapsed := time.Now().Sub(start)

	if stackLines.Display {
		resultLines := display.PrintCmdResult(bootstrap, cc.Screen, cmd,
			cmdEnv, succeeded, elapsed, flow.Cmds, currCmdIdx, cc.Cmds.Strs)
		display.RenderCmdResult(resultLines, cmdEnv, cc.Screen)
	} else if currCmdIdx < len(flow.Cmds)-1 && ln != cc.Screen.OutputNum() {
		last := flow.Cmds[len(flow.Cmds)-1]
		if last.LastCmd() != nil && !last.LastCmd().IsQuiet() {
			cc.Screen.Print("\n")
		}
	}
	return
}

// 1. remove the cmds only have cmd-level env definication but have no executable
// 2. move priority cmds to the head
func filterEmptyCmdsAndReorderByPriority(
	cc *core.Cli,
	flow *core.ParsedCmds) bool {

	var unfiltered []core.ParsedCmd
	unfilteredGlobalSeqIdx := -1
	var priorities []core.ParsedCmd
	prioritiesGlobalSeqIdx := -1

	for i, cmd := range flow.Cmds {
		if cmd.TotallyEmpty() {
			continue
		}
		if cmd.IsPriority() {
			priorities = append(priorities, cmd)
			if i == flow.GlobalSeqIdx {
				prioritiesGlobalSeqIdx = len(priorities) - 1
			}
		} else {
			unfiltered = append(unfiltered, cmd)
			if i == flow.GlobalSeqIdx {
				unfilteredGlobalSeqIdx = len(unfiltered) - 1
			}
		}
	}
	flow.Cmds = append(priorities, unfiltered...)
	if prioritiesGlobalSeqIdx >= 0 {
		flow.GlobalSeqIdx = prioritiesGlobalSeqIdx
	}
	if unfilteredGlobalSeqIdx >= 0 {
		flow.GlobalSeqIdx = len(priorities) + unfilteredGlobalSeqIdx
	}
	return true
}

func (self *Executor) sessionInit(cc *core.Cli, flow *core.ParsedCmds) bool {
	env := cc.GlobalEnv

	if env.GetInt("sys.stack-depth") > 1 {
		return true
	}

	sessionsRoot := env.GetRaw("sys.paths.sessions")
	if len(sessionsRoot) == 0 {
		cc.Screen.Print("[ERR] can't get sessions' root path\n")
		return false
	}

	os.MkdirAll(sessionsRoot, os.ModePerm)
	dirs, err := os.ReadDir(sessionsRoot)
	if err != nil {
		cc.Screen.Print(fmt.Sprintf("[ERR] can't read sessions' root path '%s'\n",
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

	sessionPath := filepath.Join(sessionsRoot, pid)
	err = os.MkdirAll(sessionPath, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		cc.Screen.Print(fmt.Sprintf("[ERR] can't create session dir '%s'\n",
			sessionPath))
		return false
	}

	pipePath := filepath.Join(sessionPath, self.pipeName)
	core.SaveEnvToFile(env, pipePath, cc.Cmds.Strs.EnvKeyValSep)
	return true
}

func checkEnvOps(cc *core.Cli, flow *core.ParsedCmds) bool {
	checker := core.EnvOpsChecker{}
	sep := cc.Cmds.Strs.PathSep
	for _, cmd := range flow.Cmds {
		last := cmd.LastCmd()
		if last == nil {
			continue
		}
		result := checker.OnCallCmd(cmd, cc.Cmds.Strs.PathSep, last, true)
		// TODO: tell user more details, auto-find the provider
		for _, res := range result {
			realPath := strings.Join(cmd.Path(), sep)
			matchedPath := strings.Join(cmd.MatchedPath(), sep)
			var shortFor string
			if realPath != matchedPath {
				shortFor = " (short for '" + realPath + "')"
			}
			cc.Screen.Print(fmt.Sprintf("[ERR] cmd '%s'%s reads '%s' but no one provide it\n",
				matchedPath, shortFor, res.Key))
		}
		if len(result) != 0 {
			return false
		}
	}
	return true
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
