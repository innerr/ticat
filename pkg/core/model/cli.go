package model

import (
	"fmt"
	"strings"
)

type Screen interface {
	Print(text string)
	Error(text string)
	// Same as line-number, but it's the count of 'Print'
	OutputNum() int
}

type ExecPolicy string

const ExecPolicyExec ExecPolicy = "exec"
const ExecPolicySkip ExecPolicy = "skip"

type ExecuteMask struct {
	Cmd                 string
	OverWriteStartEnv   *Env
	OverWriteFinishEnv  *Env
	ExecPolicy          ExecPolicy
	FileNFlowExecPolicy ExecPolicy
	SubFlow             []*ExecuteMask
	ResultIfExecuted    ExecutedResult
	ExecutedCmd         *ExecutedCmd
}

func NewExecuteMask(cmd string) *ExecuteMask {
	return &ExecuteMask{cmd, nil, nil, ExecPolicyExec, ExecPolicyExec, nil, ExecutedResultUnRun, nil}
}

func (self *ExecuteMask) Copy() *ExecuteMask {
	return &ExecuteMask{
		self.Cmd,
		self.OverWriteStartEnv,
		self.OverWriteFinishEnv,
		self.ExecPolicy,
		self.FileNFlowExecPolicy,
		self.SubFlow,
		ExecutedResultUnRun,
		self.ExecutedCmd,
	}
}

func (self *ExecuteMask) SetExecPolicyForAll(policy ExecPolicy) {
	self.ExecPolicy = policy
	self.FileNFlowExecPolicy = policy
	for _, sub := range self.SubFlow {
		sub.SetExecPolicyForAll(policy)
	}
}

type Executor interface {
	Execute(caller string, inerrCall bool, cc *Cli, env *Env, masks []*ExecuteMask, input ...string) bool
	Clone() Executor
}

type HandledErrors map[interface{}]bool

type Cli struct {
	Screen             Screen
	Cmds               *CmdTree
	Parser             CliParser
	EnvAbbrs           *EnvAbbrs
	TolerableErrs      *TolerableErrs
	Executor           Executor
	Helps              *Helps
	BgTasks            *BgTasks
	CmdIO              *CmdIO
	FlowStatus         *ExecutingFlow
	BreakPoints        *BreakPoints
	HandledErrors      HandledErrors
	ForestMode         *ForestMode
	Blender            *Blender
	Arg2EnvAutoMapCmds Arg2EnvAutoMapCmds
	EnvKeysInfo        *EnvKeysInfo
}

func NewCli(screen Screen, cmds *CmdTree, parser CliParser, abbrs *EnvAbbrs, cmdIO *CmdIO,
	envKeysInfo *EnvKeysInfo) *Cli {

	return &Cli{
		screen,
		cmds,
		parser,
		abbrs,
		NewTolerableErrs(),
		nil,
		NewHelps(),
		NewBgTasks(),
		cmdIO,
		nil,
		NewBreakPoints(),
		HandledErrors{},
		NewForestMode(),
		NewBlender(),
		Arg2EnvAutoMapCmds{},
		envKeysInfo,
	}
}

func (self *Cli) SetFlowStatusWriter(status *ExecutingFlow) {
	if self.FlowStatus != nil {
		panic(fmt.Errorf("[SetFlowStatusWriter] should never happen"))
	}
	self.FlowStatus = status
}

// Shadow copy
func (self *Cli) CopyForInteract() *Cli {
	return &Cli{
		self.Screen,
		self.Cmds,
		self.Parser,
		self.EnvAbbrs,
		self.TolerableErrs,
		self.Executor,
		self.Helps,
		self.BgTasks,
		self.CmdIO,
		nil,
		self.BreakPoints,
		HandledErrors{},
		self.ForestMode,
		self.Blender,
		Arg2EnvAutoMapCmds{},
		self.EnvKeysInfo,
	}
}

// TODO: fixme, do real clone for ready-only instances
func (self *Cli) CloneForAsyncExecuting(env *Env) *Cli {
	screen := NewBgTaskScreen()
	bgStdout := screen.GetBgStdout()
	return &Cli{
		screen,
		self.Cmds,
		self.Parser,
		self.EnvAbbrs,
		NewTolerableErrs(),
		self.Executor.Clone(),
		self.Helps,
		self.BgTasks,
		NewCmdIO(nil, bgStdout, bgStdout),
		nil,
		NewBreakPoints(),
		HandledErrors{},
		self.ForestMode.Clone(),
		self.Blender.Clone(),
		Arg2EnvAutoMapCmds{},
		self.EnvKeysInfo,
	}
}

func (self *Cli) CloneForChecking() *Cli {
	return &Cli{
		self.Screen,
		self.Cmds,
		self.Parser,
		self.EnvAbbrs,
		self.TolerableErrs,
		self.Executor,
		self.Helps,
		self.BgTasks,
		self.CmdIO,
		self.FlowStatus,
		self.BreakPoints,
		self.HandledErrors,
		self.ForestMode,
		self.Blender.Clone(),
		Arg2EnvAutoMapCmds{},
		self.EnvKeysInfo,
	}
}

func (self *Cli) ParseCmd(panicOnError bool, cmdAndArgs ...string) (parsedCmd ParsedCmd, ok bool) {
	parsed := self.Parser.Parse(self.Cmds, self.EnvAbbrs, cmdAndArgs...)
	err := parsed.FirstErr()
	if len(parsed.Cmds) != 1 || err != nil {
		if panicOnError {
			cmdStr := strings.Join(cmdAndArgs, " ")
			if err != nil {
				panic(err.Error)
			}
			if len(parsed.Cmds) > 1 {
				panic(fmt.Errorf("[Cli.ParseCmd] too many result commands by parsing '%s': %d",
					cmdStr, len(parsed.Cmds)))
			} else {
				panic(fmt.Errorf("[Cli.ParseCmd] no result command by parsing '%s'", cmdStr))
			}
		} else {
			return
		}
	}
	return parsed.Cmds[0], true
}

func (self *Cli) NormalizeCmd(panicOnError bool, cmd string) (verifiedCmdPath string) {
	parsed, ok := self.ParseCmd(panicOnError, cmd)
	if ok {
		verifiedCmdPath = strings.Join(parsed.Path(), self.Cmds.Strs.PathSep)
	}
	return
}
