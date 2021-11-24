package core

// TODO: use io.Write ?
type Screen interface {
	Print(text string)
	Error(text string)
	// Same as line-number, but it's the count of 'Print'
	OutputNum() int
}

type QuietScreen struct {
	outN int
}

func (self *QuietScreen) Print(text string) {
	self.outN += 1
}

func (self *QuietScreen) Error(text string) {
}

func (self *QuietScreen) OutputNum() int {
	return self.outN
}

type Executor interface {
	Execute(caller string, cc *Cli, input ...string) bool
}

type Cli struct {
	GlobalEnv     *Env
	Screen        Screen
	Cmds          *CmdTree
	Parser        CliParser
	EnvAbbrs      *EnvAbbrs
	TolerableErrs *TolerableErrs
	Executor      Executor
	Helps         *Helps
}

func NewCli(env *Env, screen Screen, cmds *CmdTree, parser CliParser, abbrs *EnvAbbrs) *Cli {
	return &Cli{
		env,
		screen,
		cmds,
		parser,
		abbrs,
		NewTolerableErrs(),
		nil,
		NewHelps(),
	}
}

func (self *Cli) Copy() *Cli {
	return &Cli{
		self.GlobalEnv,
		self.Screen,
		self.Cmds,
		self.Parser,
		self.EnvAbbrs,
		self.TolerableErrs,
		nil,
		self.Helps,
	}
}

// TODO: fixme, do real clone for ready-only instances
func (self *Cli) CloneForAsyncExecuting() *Cli {
	return &Cli{
		self.GlobalEnv.Clone(),
		&QuietScreen{},
		self.Cmds,
		self.Parser,
		self.EnvAbbrs,
		NewTolerableErrs(),
		nil,
		NewHelps(),
	}
}
