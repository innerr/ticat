package cli

import (
	"fmt"
)

type CmdType int

const (
	CmdTypeNormal CmdType = iota
	CmdTypePower
	CmdTypeBash
	// TODO: Py2, Py3
)

type NormalCmd func(cli *Cli, env *Env) (succeeded bool)
type PowerCmd func(cli *Cli, env *Env, cmds []ParsedCmd, currCmdIdx int) (modified []ParsedCmd, succeeded bool)

type Cmd struct {
	owner      *CmdTree
	ty         CmdType
	quiet      bool
	args       Args
	normal     NormalCmd
	power      PowerCmd
}

func NewCmd(owner *CmdTree, cmd NormalCmd, quiet bool) *Cmd {
	return &Cmd{owner, CmdTypeNormal, quiet, newArgs(), cmd, nil}
}

func NewPowerCmd(owner *CmdTree, cmd PowerCmd, quiet bool) *Cmd {
	return &Cmd{owner, CmdTypePower, quiet, newArgs(), nil, cmd}
}

func (self *Cmd) Execute(cli *Cli, env *Env, cmds []ParsedCmd, currCmdIdx int) ([]ParsedCmd, bool) {
	switch self.ty {
	case CmdTypePower:
		return self.power(cli, env, cmds, currCmdIdx)
	case CmdTypeNormal:
		return cmds, self.normal(cli, env)
	default:
		panic(fmt.Errorf("unknown command executable type: %d", self.ty))
	}
}

func (self *Cmd) IsPowerCmd() bool {
	return self.ty == CmdTypePower
}

func (self *Cmd) AddArg(name string, abbrs ...string) *Cmd {
	self.args.AddArg(self.owner, name, abbrs...)
	return self
}

type Args struct {
	// Use a map as a set
	pairs map[string]bool
	abbrsRevIdx map[string]string
}

func newArgs() Args {
	return Args{
		map[string]bool{},
		map[string]string{},
	}
}

func (self *Args) AddArg(owner *CmdTree, name string, abbrs ...string) {
	if _, ok := self.pairs[name]; ok {
		panic(fmt.Errorf("%s%s: arg name conflicted: %s", ErrStrPrefix, owner.DisplayPath(), name))
	}
	for _, abbr := range abbrs {
		if old, ok := self.abbrsRevIdx[abbr]; ok {
			panic(fmt.Errorf("%s%s: arg abbr name '%s' conflicted, old for '%s', new for '%s'",
				ErrStrPrefix, owner.DisplayPath(), abbr, old, name))
		}
		self.abbrsRevIdx[abbr] = name
	}
	self.abbrsRevIdx[name] = name
	self.pairs[name] = true
}


