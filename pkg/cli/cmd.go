package cli

import (
	"fmt"
)

type CmdType int

const (
	CmdTypeNormal CmdType = iota
	CmdTypePower
	CmdTypeBash
)

type NormalCmd func(argv ArgVals, cli *Cli, env *Env) (succeeded bool)
type PowerCmd func(argv ArgVals, cli *Cli, env *Env, cmds []ParsedCmd,
	currCmdIdx int) (newCmds []ParsedCmd, newCurrCmdIdx int, succeeded bool)

type Cmd struct {
	owner  *CmdTree
	ty     CmdType
	quiet  bool
	args   Args
	normal NormalCmd
	power  PowerCmd
	bash   string
}

func NewCmd(owner *CmdTree, cmd NormalCmd, quiet bool) *Cmd {
	return &Cmd{owner, CmdTypeNormal, quiet, newArgs(), cmd, nil, ""}
}

func NewPowerCmd(owner *CmdTree, cmd PowerCmd, quiet bool) *Cmd {
	return &Cmd{owner, CmdTypePower, quiet, newArgs(), nil, cmd, ""}
}

func NewBashCmd(owner *CmdTree, cmd string) *Cmd {
	return &Cmd{owner, CmdTypeBash, false, newArgs(), nil, nil, cmd}
}

func (self *Cmd) Execute(argv ArgVals, cli *Cli, env *Env, cmds []ParsedCmd, currCmdIdx int) ([]ParsedCmd, int, bool) {
	switch self.ty {
	case CmdTypePower:
		return self.power(argv, cli, env, cmds, currCmdIdx)
	case CmdTypeNormal:
		return cmds, currCmdIdx, self.normal(argv, cli, env)
	case CmdTypeBash:
		fmt.Println("TODO: execute bash command:", self.bash)
		return cmds, currCmdIdx, true
	default:
		panic(fmt.Errorf("[Cmd.Execute] unknown command executable type: %d", self.ty))
	}
}

func (self *Cmd) IsPowerCmd() bool {
	return self.ty == CmdTypePower
}

func (self *Cmd) AddArg(name string, defVal string, abbrs ...string) *Cmd {
	self.args.AddArg(self.owner, name, defVal, abbrs...)
	return self
}
