package core

import (
	"fmt"
	"strings"
)

type CmdTreeStrs struct {
	RootDisplayName string
	PathSep         string
	EnvValDelMark string
	EnvValDelAllMark string
	ProtoEnvMark    string
	ProtoSep        string
}

type CmdTree struct {
	Strs            *CmdTreeStrs
	name            string
	parent          *CmdTree
	subs            map[string]*CmdTree
	subOrderedNames []string
	cmd             *Cmd
	subAbbrs        map[string][]string
	subAbbrsRevIdx  map[string]string
}

func NewCmdTree(strs *CmdTreeStrs) *CmdTree {
	return &CmdTree{
		strs,
		"",
		nil,
		map[string]*CmdTree{},
		[]string{},
		nil,
		map[string][]string{},
		map[string]string{},
	}
}

func (self *CmdTree) Execute(argv ArgVals, cc *Cli, env *Env, cmds []ParsedCmd, currCmdIdx int) ([]ParsedCmd, int, bool) {
	if self.cmd == nil {
		return cmds, currCmdIdx, true
	} else {
		return self.cmd.Execute(argv, cc, env, cmds, currCmdIdx)
	}
}

func (self *CmdTree) RegCmd(cmd NormalCmd, help string) *Cmd {
	self.cmd = NewCmd(self, help, cmd)
	return self.cmd
}

func (self *CmdTree) RegBashCmd(cmd string, help string) *Cmd {
	self.cmd = NewBashCmd(self, help, cmd)
	return self.cmd
}

func (self *CmdTree) RegPowerCmd(cmd PowerCmd, help string) *Cmd {
	self.cmd = NewPowerCmd(self, help, cmd)
	return self.cmd
}

func (self *CmdTree) AddSub(name string, abbrs ...string) *CmdTree {
	if old, ok := self.subs[name]; ok && old.name != name {
		panic(fmt.Errorf("[CmdTree.AddSub] %s: sub-cmd name conflicted: %s", self.DisplayPath(), name))
	}
	sub := NewCmdTree(self.Strs)
	sub.name = name
	sub.parent = self
	self.subs[name] = sub
	self.subOrderedNames = append(self.subOrderedNames, name)
	self.addSubAbbr(name, abbrs...)
	self.subAbbrsRevIdx[name] = name
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
	return self.cmd != nil && self.cmd.IsQuiet()
}

func (self *CmdTree) IsPowerCmd() bool {
	return self.cmd != nil && self.cmd.IsPowerCmd()
}

func (self *CmdTree) Parent() *CmdTree {
	return self.parent
}

func (self *CmdTree) Name() string {
	return self.name
}

func (self *CmdTree) DisplayName() string {
	if len(self.name) == 0 {
		return self.Strs.RootDisplayName
	}
	return self.name
}

func (self *CmdTree) Cmd() *Cmd {
	return self.cmd
}

func (self *CmdTree) Args() (args Args) {
	if self.cmd == nil {
		return
	}
	return self.cmd.Args()
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
		return self.Strs.RootDisplayName
	} else {
		return strings.Join(path, self.Strs.PathSep)
	}
}

func (self *CmdTree) Realname(abbr string) (realname string) {
	if abbr == self.name {
		return abbr
	}
	if self.parent == nil {
		return
	}
	realname, _ = self.parent.subAbbrsRevIdx[abbr]
	return
}

func (self *CmdTree) SubNames() []string {
	return self.subOrderedNames
}

func (self *CmdTree) SubAbbrs(name string) (abbrs []string) {
	abbrs, _ = self.subAbbrs[name]
	return
}

func (self *CmdTree) Abbrs() (abbrs []string) {
	if self.parent == nil {
		return
	}
	return self.parent.SubAbbrs(self.name)
}

func (self *CmdTree) addSubAbbr(name string, abbrs ...string) {
	for _, abbr := range abbrs {
		if len(abbr) == 0 {
			continue
		}
		if old, ok := self.subAbbrsRevIdx[abbr]; ok && old != name {
			panic(fmt.Errorf(
				"[CmdTree.addSubAbbr] %s: command abbr name '%s' conflicted, old for '%s', new for '%s'",
				self.DisplayPath(), abbr, old, name))
		}
		self.subAbbrsRevIdx[abbr] = name
	}
	olds, _ := self.subAbbrs[name]
	self.subAbbrs[name] = append(olds, abbrs...)
}

func (self *CmdTree) getSub(addIfNotExists bool, path ...string) *CmdTree {
	if len(path) == 0 {
		return self
	}
	name := path[0]
	if realName, ok := self.subAbbrsRevIdx[name]; ok {
		name = realName
	}
	sub, ok := self.subs[name]
	if !ok {
		if !addIfNotExists {
			return nil
		}
		sub = self.AddSub(name)
	}
	return sub.getSub(addIfNotExists, path[1:]...)
}
