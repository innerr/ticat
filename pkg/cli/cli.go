package cli

import (
	"time"
)

type Cli struct {
	GlobalEnv *Env
	Screen    *Screen
	Cmds      *CmdTree
	Parser    CliParser
	EnvAbbrs  *EnvAbbrs
}

func NewCli(builtinModsLoader func(*CmdTree), envLoader func(*Env), parser CliParser) *Cli {
	env := NewEnv().NewLayers(
		EnvLayerDefault,
		EnvLayerPersisted,
		EnvLayerSession,
	)
	cc := &Cli{
		env,
		NewScreen(),
		NewCmdTree(CmdRootDisplayName, CmdPathSep),
		parser,
		NewEnvAbbrs(CmdRootDisplayName),
	}
	builtinModsLoader(cc.Cmds)
	envLoader(cc.GlobalEnv)
	return cc
}

func (self *Cli) Execute(bootstrap string, env *Env, script ...string) bool {
	// This could be very useful for customized mods-loader or env-loader
	// (those loaders will be loaded from 'bootstrap string' above)
	if env != nil {
		self.GlobalEnv.GetLayer(EnvLayerSession).Merge(env)
	}

	if !self.execute(true, bootstrap) {
		return false
	}

	extra := self.GlobalEnv.Get("bootstrap").Raw
	if len(extra) != 0 && !self.execute(true, extra) {
		return false
	}

	self.GlobalEnv.PlusInt("sys.stack-depth", 1)
	if !self.execute(false, script...) {
		return false
	}
	self.GlobalEnv.PlusInt("sys.stack-depth", -1)
	return true
}

func (self *Cli) execute(isBootstrap bool, script ...string) bool {
	if len(script) == 0 {
		return true
	}
	useCmdAbbrs(self.EnvAbbrs, self.Cmds)
	cmds := self.Parser.Parse(self.Cmds, self.EnvAbbrs, script...)
	return self.executeCmds(isBootstrap, cmds)
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

func (self *Cli) executeCmds(isBootstrap bool, cmds *ParsedCmds) bool {
	self.filterEmptyCmds(cmds)
	env := self.GlobalEnv.GetLayer(EnvLayerSession)
	if cmds.GlobalEnv != nil {
		cmds.GlobalEnv.WriteNotArgTo(env)
	}
	for i := 0; i < len(cmds.Cmds); i++ {
		var newCmds []ParsedCmd
		var succeeded bool
		cmd := cmds.Cmds[i]
		newCmds, i, succeeded = self.executeCmd(isBootstrap, cmd, env, cmds.Cmds, i)
		if !succeeded {
			return false
		}
		cmds.Cmds = newCmds
	}
	return true
}

func (self *Cli) executeCmd(isBootstrap bool, cmd ParsedCmd, env *Env, cmds []ParsedCmd,
	currCmdIdx int) (newCmds []ParsedCmd, newCurrCmdIdx int, succeeded bool) {

	sep := self.Parser.CmdPathSep()
	// The env modifications from script will be popped out after a command is executed
	// (TODO) But if a mod modified the env, the modifications stay in session level
	cmdEnv := cmd.GenEnv(env)
	argv := cmdEnv.GetArgv(cmd.Path(), sep, cmd.Args())

	printCmdStack(isBootstrap, self.Screen, cmd, cmdEnv, cmds, currCmdIdx, sep)
	last := cmd[len(cmd)-1].Cmd.Cmd
	start := time.Now()
	if last != nil {
		newCmds, newCurrCmdIdx, succeeded = last.Execute(argv, self, cmdEnv, cmds, currCmdIdx)
	} else {
		newCmds, newCurrCmdIdx, succeeded = cmds, currCmdIdx, false
	}
	printCmdResult(isBootstrap, self.Screen, cmd, cmdEnv, succeeded, time.Now().Sub(start), cmds, currCmdIdx, sep)
	return
}

func useCmdAbbrs(abbrs *EnvAbbrs, tree *CmdTree) {
	if tree == nil {
		return
	}
	for _, subName := range tree.SubNames() {
		subAbbrs := tree.SubAbbrs(subName)
		subEnv := abbrs.GetSub(subName)
		if subEnv == nil {
			subEnv = abbrs.AddSub(subName, subAbbrs...)
		} else {
			abbrs.AddSubAbbrs(subName, subAbbrs...)
		}
		subTree := tree.GetSub(subName)
		useCmdAbbrs(subEnv, subTree)
	}
}
