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

type Context struct {
	Executor *Executor
}

type CmdTree struct {
	sub map[string]*CmdTree
	cmd func(*Context, []string)
	subAbbrsRevIdx map[string]string
}

func NewCmdTree() *CmdTree {
	return &CmdTree{ map[string]*CmdTree{}, nil, map[string]string{} }
}

func (self *CmdTree) SetCmd(cmd func(*Context, []string)) {
	self.cmd = cmd
}

func (self *CmdTree) AddSub(name string, abbrs ...string) *CmdTree {
	if _, ok := self.sub[name]; ok {
		// TODO: full path
		panic(fmt.Errorf("cmd name conflicted: {}", name))
	}
	for _, abbr := range abbrs {
		if _, ok := self.subAbbrsRevIdx[abbr]; ok {
			// TODO: full path
			panic(fmt.Errorf("cmd abbr name conflicted: {}", abbr))
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

func (self *CmdTree) ExecuteSequence(argvs [][]string, ctx *Context, env *Env, screen *Screen, brackets *Brackets) bool {
	if len(argvs) == 1 {
		return self.ExecuteWithEnvStrs(argvs[0], ctx, env, screen, brackets)
	} else {
		for i, argv := range argvs {
			res := self.ExecuteWithEnvStrs(argv, ctx, env, screen, brackets)
			if !res {
				return false
			}
			if i + 1 != len(argvs) {
				screen.Print("")
			}
		}
	}
	return true
}

func (self *CmdTree) execute(argv []string, ctx *Context, screen *Screen) bool {
	screen.PrintSeperatingHeader(strings.Join(argv, " "))
	self.cmd(ctx, argv)
	return true
}

func (self *CmdTree) ExecuteWithEnvStrs(argv []string, ctx *Context, env *Env, screen *Screen, brackets *Brackets) bool {
	if len(argv) == 0 {
		return self.execute(argv, ctx, screen)
	}

	sub := self.GetSub(argv[0])
	if sub != nil {
		if argv[0] == brackets.Left {
			self.errInvalidCmdName(argv, screen)
			return false
		}
		return sub.ExecuteWithEnvStr(argv[1:], env, screen, brackets)
	}

	i := arg.Index(brackets.Left)
	if i < 0 {
		self.errCmdNotFound(argv, screen)
		return false
	}
	if i == 0 {
		if len(argv[0]) != len(brackets.Left) {
			argv = append(argv[0][len(brackets.Left):], arvg[1:]...)
		}
		envStrs, newArgv := self.extractEnvStrs(argv, brackets.Right)
		newEnv := env.ParseAdd(envStrs)
		return sub.ExecuteWithEnvStr(newArgv, newEnv, screen, brackets)
	} else {
		sub, ok := self.GetSub(argv[0][0:i])
		if !ok {
			self.errCmdNotFound(argv, screen)
			return false
		}
		subArgv := append(argv[0][i+len(brackets.Left):], arvg[1:]...)
		envStrs, subArgv = self.extractEnvStrs(subArgv, brackets.Right)
		subEnv := env.ParseAdd(envStrs)
		return sub.Execute(subArgv, subEnv, screen, brackets)
	}
}

func (self *CmdTree) extractEnvStrs(argv []string, endingStr string) (env []string, rest[]string) {
	rest = []string{}
	inEnvStr := false
	for _, arg := range argv {
		i := arg.Index(brackets.Left)
		if i < 0 {
			
			rest = append(rest, arg)
			
		}
		if arg == brackets.Left {
			inEnvStr = true
		} else {
			rest = append(rest, arg)
		}
		rest = argv
		
	}
	return
}

type Executor struct {
	brackets *Brackets
	breaker *SequenceBreaker
	sreen *Screen
	env *Env
	cmds * CmdTree
}

func NewExecutor() *Executor {
	executor := &Executor {
		&Brackets{"{", "}"},
		&SequenceBreaker{":", []string{"http", "HTTP"}, []string{"/"}},
		&Screen{},
		newEnv(),
		NewCmdTree(),
	}
	RegisterBuiltins(executor.cmds)
	return executor
}

func (self *Executor) Execute(argvs [][]string) bool {
	return self.Cmds.ExecuteSequence(
		self.breaker.Break(argvs), self.screen, self.brackets)
}
