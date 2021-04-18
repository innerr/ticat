package cli

import (
	"fmt"
	"strings"
)

type NormalCmd func(cli *Cli, env *Env) (succeeded bool)
type PowerCmd func(cli *Cli, env *Env, cmds []ParsedCmd, currCmdIdx int) (modified []ParsedCmd, succeeded bool)

type Cmd struct {
	IsPowerCmd bool
	Quiet      bool
	Normal     NormalCmd
	Power      PowerCmd
}

func (self *Cmd) AddArg(name string, abbrs ...string) *Cmd {
	// TODO:
	return self
}

func NewCmd(cmd NormalCmd, quiet bool) *Cmd {
	return &Cmd{false, quiet, cmd, nil}
}

func NewPowerCmd(cmd PowerCmd, quiet bool) *Cmd {
	return &Cmd{true, quiet, nil, cmd}
}

type CmdTree struct {
	name           string
	parent         *CmdTree
	sub            map[string]*CmdTree
	cmd            *Cmd
	subAbbrsRevIdx map[string]string
}

func NewCmdTree() *CmdTree {
	return &CmdTree{"", nil, map[string]*CmdTree{}, nil, map[string]string{}}
}

func (self *CmdTree) RegCmd(cmd NormalCmd) *Cmd {
	self.cmd = NewCmd(cmd, false)
	return self.cmd
}

func (self *CmdTree) RegQuietCmd(cmd NormalCmd) *Cmd {
	self.cmd = NewCmd(cmd, true)
	return self.cmd
}

func (self *CmdTree) RegPowerCmd(cmd PowerCmd) *Cmd {
	self.cmd = NewPowerCmd(cmd, false)
	return self.cmd
}

func (self *CmdTree) RegQuietPowerCmd(cmd PowerCmd) *Cmd {
	self.cmd = NewPowerCmd(cmd, true)
	return self.cmd
}

func (self *CmdTree) Name() string {
	return self.name
}

func (self *CmdTree) IsQuiet() bool {
	return self.cmd != nil && self.cmd.Quiet
}

func (self *CmdTree) IsPowerCmd() bool {
	return self.cmd != nil && self.cmd.IsPowerCmd
}

func (self *CmdTree) Execute(cli *Cli, env *Env, cmds []ParsedCmd, currCmdIdx int) ([]ParsedCmd, bool) {
	if self.cmd == nil {
		return cmds, true
	}
	if self.cmd.IsPowerCmd {
		return self.cmd.Power(cli, env, cmds, currCmdIdx)
	} else {
		return cmds, self.cmd.Normal(cli, env)
	}
}

func (self *CmdTree) path() []string {
	if self.parent == nil {
		return []string{}
	}
	return append(self.parent.path(), self.name)
}

func (self *CmdTree) displayPath() string {
	path := self.path()
	if len(path) == 0 {
		return cmdRootNodeName
	} else {
		return strings.Join(self.path(), ".")
	}
}

func (self *CmdTree) AddSub(name string, abbrs ...string) *CmdTree {
	if _, ok := self.sub[name]; ok {
		panic(fmt.Errorf("%s%s: sub-cmd name conflicted: %s", errStrPrefix, self.displayPath(), name))
	}
	for _, abbr := range abbrs {
		if _, ok := self.subAbbrsRevIdx[abbr]; ok {
			// TODO: full info
			panic(fmt.Errorf("%s%s: cmd abbr name conflicted: %s", errStrPrefix, self.displayPath(), abbr))
		}
		self.subAbbrsRevIdx[abbr] = name
	}
	sub := NewCmdTree()
	sub.name = name
	self.sub[name] = sub
	return sub
}

func (self *CmdTree) GetSub(name string) *CmdTree {
	if realName, ok := self.subAbbrsRevIdx[name]; ok {
		name = realName
	}
	sub, _ := self.sub[name]
	return sub
}

const (
	cmdRootNodeName string = "<root>"
	errStrPrefix    string = "[ERR] "
)
