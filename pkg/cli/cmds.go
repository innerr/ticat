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

func (self *CmdTree) Execute(argv ArgVals, cli *Cli, env *Env, cmds []ParsedCmd, currCmdIdx int) ([]ParsedCmd, int, bool) {
	if self.cmd == nil {
		return cmds, currCmdIdx, true
	} else {
		return self.cmd.Execute(argv, cli, env, cmds, currCmdIdx)
	}
}

func (self *CmdTree) RegCmd(cmd NormalCmd) *Cmd {
	self.cmd = NewCmd(self, cmd, false)
	return self.cmd
}

func (self *CmdTree) RegBashCmd(cmd string) *Cmd {
	self.cmd = NewBashCmd(self, cmd)
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

func (self *CmdTree) AddSub(name string, abbrs ...string) *CmdTree {
	if _, ok := self.sub[name]; ok {
		panic(fmt.Errorf("[CmdTree.AddSub] %s%s: sub-cmd name conflicted: %s", ErrStrPrefix, self.DisplayPath(), name))
	}
	sub := NewCmdTree()
	sub.name = name
	sub.parent = self
	self.sub[name] = sub
	self.addSubAbbr(name, abbrs...)
	return sub
}

func (self *CmdTree) AddAbbr(abbrs ...string) {
	if self.parent == nil {
		panic(fmt.Errorf("[CmdTree.AddAbbr] can't add abbrs %v to root", abbrs))
	}
	self.parent.addSubAbbr(self.name, abbrs...)
}

func (self *CmdTree) GetOrAddSub(path ...string) *CmdTree {
	return self.getSub(true, path...)
}

func (self *CmdTree) GetSub(path ...string) *CmdTree {
	return self.getSub(false, path...)
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

func (self *CmdTree) DisplayName() string {
	if len(self.name) == 0 {
		return CmdRootNodeName
	}
	return self.name
}

func (self *CmdTree) addSubAbbr(name string, abbrs ...string) {
	for _, abbr := range abbrs {
		if len(abbr) == 0 {
			continue
		}
		if old, ok := self.subAbbrsRevIdx[abbr]; ok {
			panic(fmt.Errorf(
				"[CmdTree.addSubAbbr] %s%s: command abbr name '%s' conflicted, old for '%s', new for '%s'",
				ErrStrPrefix, self.DisplayPath(), abbr, old, name))
		}
		self.subAbbrsRevIdx[abbr] = name
	}
	self.subAbbrsRevIdx[name] = name
}

func (self *CmdTree) getSub(addIfNotExists bool, path ...string) *CmdTree {
	if len(path) == 0 {
		return self
	}
	name := path[0]
	if realName, ok := self.subAbbrsRevIdx[name]; ok {
		name = realName
	}
	sub, ok := self.sub[name]
	if !ok {
		if !addIfNotExists {
			return nil
		}
		sub = self.AddSub(name)
	}
	return sub.getSub(addIfNotExists, path[1:]...)
}
