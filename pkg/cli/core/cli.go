package core

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
	Cmd               string
	OverWriteStartEnv *Env
	ExecPolicy        ExecPolicy
	SubFlow           []*ExecuteMask
}

func NewExecuteMask(cmd string) *ExecuteMask {
	return &ExecuteMask{cmd, nil, ExecPolicyExec, nil}
}

func (self *ExecuteMask) Copy() *ExecuteMask {
	return &ExecuteMask{self.Cmd, self.OverWriteStartEnv, self.ExecPolicy, self.SubFlow}
}

type Executor interface {
	Execute(caller string, inerrCall bool, cc *Cli, env *Env, masks []*ExecuteMask, input ...string) bool
	Clone() Executor
}

type HandledErrors map[interface{}]bool

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
	HandledErrors HandledErrors
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
		HandledErrors{},
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
	}
}

func (self *Cli) ParseCmd(cmd string, panicOnError bool) (verifiedCmdPath string) {
	parsed := self.Parser.Parse(self.Cmds, self.EnvAbbrs, cmd)
	if len(parsed.Cmds) != 1 || parsed.FirstErr() != nil {
		if panicOnError {
			panic(fmt.Errorf("[Cli.ParseCmd] invalid cmd name '%s'", cmd))
		} else {
			return
		}
	}
	return strings.Join(parsed.Cmds[0].Path(), self.Cmds.Strs.PathSep)
}
