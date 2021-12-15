package core

import (
	"fmt"
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
	Cmd               string
	OverWriteStartEnv *Env
	ExecPolicy        ExecPolicy
	SubFlow           []*ExecuteMask
}

type Executor interface {
	Execute(caller string, cc *Cli, env *Env, masks []*ExecuteMask, input ...string) bool
}

type Cli struct {
	Screen        Screen
	Cmds          *CmdTree
	Parser        CliParser
	EnvAbbrs      *EnvAbbrs
	TolerableErrs *TolerableErrs
	Executor      Executor
	Helps         *Helps
	BgTasks       *BgTasks
	CmdIO         *CmdIO
	FlowStatus    *ExecutingFlow
	BreakPoints   *BreakPoints
}

func NewCli(screen Screen, cmds *CmdTree, parser CliParser, abbrs *EnvAbbrs, cmdIO *CmdIO) *Cli {
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
	}
}

func (self *Cli) SetFlowStatusWriter(status *ExecutingFlow) {
	if self.FlowStatus != nil {
		panic(fmt.Errorf("[SetExecutingFlowStatusLogger] should never happen"))
	}
	self.FlowStatus = status
}

// Shadow copy
func (self *Cli) Copy() *Cli {
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
		self.Executor,
		self.Helps,
		self.BgTasks,
		NewCmdIO(nil, bgStdout, bgStdout),
		nil,
		NewBreakPoints(),
	}
}
