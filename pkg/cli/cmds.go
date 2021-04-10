package cli

import (
	"fmt"
	"strings"
)

type Word struct {
	Val string
	Abbrs []string
}

func NewWord(val string, abbrs ...string) *Word {
	return &Word{val, abbrs}
}

type CmdTree struct {
	name string
	parent *CmdTree
	sub map[string]*CmdTree
	cmd func(*Hub, *Env, []string)bool
	subAbbrsRevIdx map[string]string
}

func NewCmdTree() *CmdTree {
	return &CmdTree{ "", nil, map[string]*CmdTree{}, nil, map[string]string{} }
}

func (self *CmdTree) SetCmd(cmd func(*Hub, *Env, []string)bool) {
	self.cmd = cmd
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
		return "(root)"
	} else {
		return strings.Join(self.path(), " ")
	}
}

func (self *CmdTree) AddSub(name string, abbrs ...string) *CmdTree {
	if _, ok := self.sub[name]; ok {
		panic(fmt.Errorf("[ERR] %s: sub-cmd name conflicted: %s", self.displayPath(), name))
	}
	for _, abbr := range abbrs {
		if _, ok := self.subAbbrsRevIdx[abbr]; ok {
			// TODO: full info
			panic(fmt.Errorf("[ERR] %s: cmd abbr name conflicted: %s", self.displayPath(), abbr))
		}
		self.subAbbrsRevIdx[abbr] = name
	}
	sub := NewCmdTree()
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
	Left string
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
			if i + 1 != len(argvs) {
				hub.Screen.Print("")
			}
		}
	}
	return true
}

func (self *CmdTree) PrintError(hub *Hub, env *Env, matchedCmdPath []string, msg string) {
	hub.Screen.Print("[ERR] " + strings.Join(append([]string{"ticat"}, matchedCmdPath...), " ") + msg)
}

func (self *CmdTree) ExecuteWithEnvStrs(hub *Hub, env *Env, argv []string, matchedCmdPath []string) bool {
	if len(argv) == 0 {
		return self.execute(hub, env, argv, matchedCmdPath)
	}

	sub := self.GetSub(argv[0])
	if sub != nil {
		if strings.HasPrefix(argv[0], hub.envBrackets.Left) {
			self.PrintError(hub, env, matchedCmdPath,
				"confused '" + argv[0] + "', could be env-begining or sub-cmd")
			return false
		}
		return sub.ExecuteWithEnvStrs(hub, env, argv[1:], append(matchedCmdPath, argv[0]))
	}

	i := strings.Index(argv[0], hub.envBrackets.Left)
	if i < 0 {
		return self.execute(hub, env, argv, matchedCmdPath)
	}
	if i == 0 {
		if len(argv[0]) != len(hub.envBrackets.Left) {
			argv = append([]string{argv[0][len(hub.envBrackets.Left):]}, argv[1:]...)
		}
		envStrs, newArgv, ok := self.extractEnvStrs(argv, hub.envBrackets.Right)
		if !ok {
			self.PrintError(hub, env, matchedCmdPath,
				"env definition not close properly '" + strings.Join(argv, " ") + "'")
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
			self.PrintError(hub, env, matchedCmdPath,
				"sub-cmd '" + subCmd + "' not found")
			return false
		}
		rest := argv[0][i+len(hub.envBrackets.Left):]
		subArgv := argv[1:]
		if len(rest) != 0 {
			subArgv = append([]string{rest}, subArgv...)
		}
		envStrs, subArgv, ok := self.extractEnvStrs(subArgv, hub.envBrackets.Right)
		if !ok {
			self.PrintError(hub, env, matchedCmdPath,
				"env definition not close properly '" + strings.Join(subArgv, " ") + "'")
		}
		subEnv := env.NewLayerIfTypeNotMatch(EnvLayerMod)
		subEnv.ParseAndSet(envStrs)
		return sub.ExecuteWithEnvStrs(hub, subEnv, subArgv, append(matchedCmdPath, subCmd))
	}
}

func (self *CmdTree) execute(hub *Hub, env *Env, argv []string, matchedCmdPath []string) bool {
	hub.Screen.PrintSeperatingHeader(strings.Join(matchedCmdPath, " ") + "(" + strings.Join(argv, " ") + ")")
	return self.cmd(hub, env, argv)
}

func (self *CmdTree) extractEnvStrs(argv []string, endingMark string) (env []string, rest[]string, ok bool) {
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
			if len(endingMark) == len(arg) - k {
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

type Hub struct {
	breaker *SequenceBreaker
	global *Env
	envBrackets *Brackets
	Screen *Screen
	Cmds * CmdTree
}

func NewHub() *Hub {
	hub := &Hub {
		&SequenceBreaker{":", []string{"http", "HTTP"}, []string{"/"}},
		NewEnv(),
		&Brackets{"{", "}"},
		&Screen{},
		NewCmdTree(),
	}
	RegisterBuiltins(hub.Cmds)
	return hub
}

func (self *Hub) Execute(argvs []string) bool {
	return self.Cmds.ExecuteSequence(self, self.global, self.breaker.Break(argvs))
}
