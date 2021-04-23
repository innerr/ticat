package cli

import (
	"time"
)

type Cli struct {
	GlobalEnv *Env
	Screen    *Screen
	Cmds      *CmdTree
	Parser    *Parser
}

func NewCli(builtinLoader func(*CmdTree), envLoader func(*Env)) *Cli {
	env := NewEnv().NewLayers(
		EnvLayerDefault,
		EnvLayerPersisted,
		EnvLayerSession,
	)
	cli := &Cli{
		env,
		NewScreen(),
		NewCmdTree(),
		NewParser(),
	}
	builtinLoader(cli.Cmds)
	envLoader(cli.GlobalEnv)
	return cli
}

func (self *Cli) Execute(env *Env, script ...string) bool {
	// In termial:
	// $> ticat desc: B.E.L.L:B.E.L.R:B.M.L.L
	bootstrap := "B.E.L.L:B.E.L.R:B.M.L.L"
	if !self.execute(bootstrap) {
		return false
	}

	// This could be very useful for customized mods-loader or env-loader
	// (those loaders will be loaded from 'bootstrap string' above)
	if env != nil {
		self.GlobalEnv.GetLayer(EnvLayerSession).Merge(env)
	}
	extra := self.GlobalEnv.Get("bootstrap").Raw
	if len(extra) != 0 && !self.execute(extra) {
		return false
	}

	self.GlobalEnv.Get("runtime.sys.stack-depth").PlusInt(1)
	if !self.execute(script...) {
		return false
	}
	self.GlobalEnv.Get("runtime.sys.stack-depth").PlusInt(-1)
	return true
}

func (self *Cli) execute(script ...string) bool {
	if len(script) == 0 {
		return true
	}
	cmds := self.Parser.Parse(self.Cmds, script...)
	return self.executeCmds(cmds)
}

// Remove the cmds only have cmd-level env definication but have no executable
func (self *Cli) filterEmptyCmds(cmds *ParsedCmds) {
	var filtered []ParsedCmd
	for _, cmd := range cmds.Cmds {
		for _, seg := range cmd {
			if seg.Cmd.Cmd != nil {
				filtered = append(filtered, cmd)
				break
			}
		}
	}
	cmds.Cmds = filtered
}

func (self *Cli) executeCmds(cmds *ParsedCmds) bool {
	self.filterEmptyCmds(cmds)
	env := self.GlobalEnv.GetLayer(EnvLayerSession)
	if cmds.GlobalEnv != nil {
		cmds.GlobalEnv.WriteNotArgTo(env)
	}
	for i := 0; i < len(cmds.Cmds); i++ {
		var newCmds []ParsedCmd
		var succeeded bool
		cmd := cmds.Cmds[i]
		newCmds, i, succeeded = self.executeCmd(cmd, env, cmds.Cmds, i)
		if !succeeded {
			return false
		}
		cmds.Cmds = newCmds
	}
	return true
}

func (self *Cli) executeCmd(cmd ParsedCmd, env *Env, cmds []ParsedCmd,
	currCmdIdx int) (newCmds []ParsedCmd, newCurrCmdIdx int, succeeded bool) {

	sep := self.Parser.CmdPathSep()
	// The env modifications from script will be popped out after a command is executed
	// (TODO) But if a mod modified the env, the modifications stay in session level
	cmdEnv := cmd.GenEnv(env)
	argv := cmdEnv.GetArgv(cmd.Path(), sep, cmd.Args())

	printCmdStack(self.Screen, cmd, cmdEnv, cmds, currCmdIdx, sep)
	last := cmd[len(cmd)-1].Cmd.Cmd
	start := time.Now()
	if last != nil {
		newCmds, newCurrCmdIdx, succeeded = last.Execute(argv, self, cmdEnv, cmds, currCmdIdx)
	} else {
		newCmds, newCurrCmdIdx, succeeded = cmds, currCmdIdx, false
	}
	printCmdResult(self.Screen, cmd, cmdEnv, succeeded, time.Now().Sub(start), cmds, currCmdIdx, sep)
	return
}
