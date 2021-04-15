package cli

import (
	"fmt"
	"strings"
	"github.com/pingcap/ticat/pkg/parser"
)

const (
	cmdRootNodeName = "<root>"
	errStrPrefix    = "[ERR] "
)

type Word struct {
	Val   string
	Abbrs []string
}

func NewWord(val string, abbrs ...string) *Word {
	return &Word{val, abbrs}
}

type Cmd struct {
	IsPowerCmd bool
	Normal func(*Hub, *Env, []string) bool
	Power func(*Hub, *Env, []string) ([]string, bool)
}

func NewCmd(cmd func(*Hub, *Env, []string) bool) *Cmd {
	return &Cmd{false, cmd, nil}
}

func NewPowerCmd(cmd func(*Hub, *Env, []string) ([]string, bool)) *Cmd {
	return &Cmd{true, nil, cmd}
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

func (self *CmdTree) SetCmd(cmd func(*Hub, *Env, []string) bool) {
	self.cmd = NewCmd(cmd)
}

func (self *CmdTree) Name() string {
	return self.name
}

func (self *CmdTree) SetPowerCmd(cmd func(*Hub, *Env, []string) ([]string, bool)) {
	self.cmd = NewPowerCmd(cmd)
}

func (self *CmdTree) IsPowerCmd() bool {
	return self.cmd != nil && self.cmd.IsPowerCmd
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

type Brackets struct {
	Left  string
	Right string
}

func (self *CmdTree) Execute(flow *parser.ParsedCmds) bool {
	return false
}

func (self *CmdTree) ExecuteSequence(hub *Hub, env *Env, argvs [][]string) bool {
	// TODO: user-cmd log
	// TODO: executing log
	if len(argvs) == 1 {
		return self.ExecuteWithEnvStrs(hub, env, argvs[0], []string{})
	} else {
		for i, argv := range argvs {
			// If a mod modified the env, the modifications stay in session level
			res := self.ExecuteWithEnvStrs(hub, env, argv, []string{})
			if !res {
				return false
			}
			if i+1 != len(argvs) {
				hub.Screen.Print("")
			}
		}
	}
	return true
}

func (self *CmdTree) PrintErr(hub *Hub, env *Env, matchedCmdPath []string, msg string) {
	displayPath := cmdRootNodeName
	if len(matchedCmdPath) != 0 {
		displayPath = strings.Join(matchedCmdPath, ".")
	}
	hub.Screen.Print(errStrPrefix + displayPath + ": " + msg)
}

func (self *CmdTree) execute(hub *Hub, env *Env, argv []string, matchedCmdPath []string) bool {
	displayPath := cmdRootNodeName
	if len(matchedCmdPath) != 0 {
		displayPath = strings.Join(matchedCmdPath, ".")
	}
	hub.Screen.PrintSeperatingHeader(displayPath + "(" + strings.Join(argv, " ") + ")")
	if self.cmd == nil {
		self.PrintErr(hub, env, matchedCmdPath, "this cmd don't have an executable")
		return false
	}
	// TODO: power cmd
	return self.cmd.Normal(hub, env, argv)
}

type Hub struct {
	GlobalEnv   *Env
	Screen      *Screen
	Cmds        *CmdTree
}

func NewHub() *Hub {
	hub := &Hub{
	}
	RegisterBuiltins(hub.Cmds)
	return hub
}

func (self *Hub) Execute(argvs []string) bool {
	inputs := self.SeqParser.Parse(argvs)
	return self.Cmds.ExecuteSequence(self, self.GlobalEnv, inputs)
}
