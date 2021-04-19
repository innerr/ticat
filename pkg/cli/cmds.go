package cli

import (
	"fmt"
	"strings"
)

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

func (self *CmdTree) Execute(cli *Cli, env *Env, cmds []ParsedCmd, currCmdIdx int) ([]ParsedCmd, bool) {
	if self.cmd == nil {
		return cmds, true
	} else {
		return self.cmd.Execute(cli, env, cmds, currCmdIdx)
	}
}

func (self *CmdTree) RegCmd(cmd NormalCmd) *Cmd {
	self.cmd = NewCmd(self, cmd, false)
	return self.cmd
}

func (self *CmdTree) RegQuietCmd(cmd NormalCmd) *Cmd {
	self.cmd = NewCmd(self, cmd, true)
	return self.cmd
}

func (self *CmdTree) RegPowerCmd(cmd PowerCmd) *Cmd {
	self.cmd = NewPowerCmd(self, cmd, false)
	return self.cmd
}

func (self *CmdTree) RegQuietPowerCmd(cmd PowerCmd) *Cmd {
	self.cmd = NewPowerCmd(self, cmd, true)
	return self.cmd
}

func (self *CmdTree) IsQuiet() bool {
	return self.cmd != nil && self.cmd.quiet
}

func (self *CmdTree) IsPowerCmd() bool {
	return self.cmd != nil && self.cmd.IsPowerCmd()
}

func (self *CmdTree) Name() string {
	return self.name
}

func (self *CmdTree) AddSub(name string, abbrs ...string) *CmdTree {
	if _, ok := self.sub[name]; ok {
		panic(fmt.Errorf("%s%s: sub-cmd name conflicted: %s", ErrStrPrefix, self.DisplayPath(), name))
	}
	for _, abbr := range abbrs {
		if old, ok := self.subAbbrsRevIdx[abbr]; ok {
			panic(fmt.Errorf("%s%s: command abbr name '%s' conflicted, old for '%s', new for '%s'",
				ErrStrPrefix, self.DisplayPath(), abbr, old, name))
		}
		self.subAbbrsRevIdx[abbr] = name
	}
	self.subAbbrsRevIdx[name] = name
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

func (self *CmdTree) Path() []string {
	if self.parent == nil {
		return []string{}
	}
	return append(self.parent.Path(), self.name)
}

func (self *CmdTree) DisplayPath() string {
	path := self.Path()
	if len(path) == 0 {
		return CmdRootNodeName
	} else {
		return strings.Join(path, ".")
	}
}
