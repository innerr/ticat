package cli

import (
	"fmt"
	"strings"
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
	Normal func(*Hub, *Env, []string) bool
	Power func(*Hub, *Env, []string) ([]string, bool)
}

func NewCmd(cmd func(*Hub, *Env, []string) bool) *Cmd {
	return &Cmd{cmd, nil}
}

func NewPowerCmd(cmd func(*Hub, *Env, []string) ([]string, bool)) *Cmd {
	return &Cmd{nil, cmd}
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

// TODO: too slow
func (self *CmdTree) ExecuteWithEnvStrs(hub *Hub, env *Env, argv []string, matchedCmdPath []string) bool {
	if len(argv) == 0 {
		return self.execute(hub, env, argv, matchedCmdPath)
	}
	sub := self.GetSub(argv[0])
	if sub != nil {
		if strings.HasPrefix(argv[0], hub.CmdParser.EnvBrackets.Left) {
			self.PrintErr(hub, env, matchedCmdPath,
				"confused '"+argv[0]+"', could be env-begining or sub-cmd")
			return false
		}
		return sub.ExecuteWithEnvStrs(hub, env, argv[1:], append(matchedCmdPath, argv[0]))
	}

	i := strings.Index(argv[0], hub.CmdParser.EnvBrackets.Left)
	if i < 0 {
		return self.execute(hub, env, argv, matchedCmdPath)
	}
	if i == 0 {
		if len(argv[0]) != len(hub.CmdParser.EnvBrackets.Left) {
			argv = append([]string{argv[0][len(hub.CmdParser.EnvBrackets.Left):]}, argv[1:]...)
		}
		envStrs, newArgv, ok := self.extractEnvStrs(argv, hub.CmdParser.EnvBrackets.Right)
		if !ok {
			self.PrintErr(hub, env, matchedCmdPath,
				"env definition not close properly '"+strings.Join(argv, " ")+"'")
		}
		layerType := EnvLayerSession
		if len(matchedCmdPath) != 0 {
			layerType = EnvLayerMod
		}
		newEnv := env.NewLayerIfTypeNotMatch(layerType)
		newEnv.ParseAndSet(envStrs)
		return self.ExecuteWithEnvStrs(hub, newEnv, newArgv, matchedCmdPath)
	} else {
		subCmd := argv[0][0:i]
		sub := self.GetSub(subCmd)
		if sub == nil {
			self.PrintErr(hub, env, matchedCmdPath,
				"sub-cmd '"+subCmd+"' not found")
			return false
		}
		rest := argv[0][i+len(hub.CmdParser.EnvBrackets.Left):]
		subArgv := argv[1:]
		if len(rest) != 0 {
			subArgv = append([]string{rest}, subArgv...)
		}
		envStrs, subArgv, ok := self.extractEnvStrs(subArgv, hub.CmdParser.EnvBrackets.Right)
		if !ok {
			self.PrintErr(hub, env, matchedCmdPath,
				"env definition not close properly '"+strings.Join(subArgv, " ")+"'")
		}
		subEnv := env.NewLayerIfTypeNotMatch(EnvLayerMod)
		subEnv.ParseAndSet(envStrs)
		return sub.ExecuteWithEnvStrs(hub, subEnv, subArgv, append(matchedCmdPath, subCmd))
	}
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

func (self *CmdTree) extractEnvStrs(argv []string, endingMark string) (env []string, rest []string, ok bool) {
	rest = []string{}
	ok = true
	for i, arg := range argv {
		k := strings.Index(arg, endingMark)
		if k < 0 {
			rest = append(rest, arg)
			continue
		}
		if k == 0 {
			if len(endingMark) == len(arg) {
				env = argv[i+1:]
				return
			} else {
				rest = append(rest, arg[len(endingMark):])
				env = append([]string{arg[len(endingMark):]}, argv[i+1:]...)
				return
			}
		} else {
			if len(endingMark) == len(arg)-k {
				rest = append(rest, arg[0:k])
				env = argv[i+1:]
				return
			} else {
				rest = append(rest, arg[0:k])
				rest = append(rest, arg[k+len(endingMark):])
				env = argv[i+1:]
				return
			}
		}
	}
	return nil, nil, false
}

type CmdParser struct {
	EnvBrackets *Brackets
}

type Hub struct {
	SeqParser *SequenceParser
	CmdParser *CmdParser
	GlobalEnv   *Env
	Screen      *Screen
	Cmds        *CmdTree
}

func NewHub() *Hub {
	hub := &Hub{
		&SequenceParser{":", []string{"http", "HTTP"}, []string{"/"}},
		&CmdParser{
			&Brackets{"{", "}"},
		},
		NewEnv(),
		&Screen{},
		NewCmdTree(),
	}
	RegisterBuiltins(hub.Cmds)
	return hub
}

func (self *Hub) Execute(argvs []string) bool {
	inputs := self.SeqParser.Parse(argvs)
	return self.Cmds.ExecuteSequence(self, self.GlobalEnv, inputs)
}
