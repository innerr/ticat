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

func (self *Cli) Execute(preparation string, script ...string) bool {
	prep := self.Parser.Parse(self.Cmds, preparation)
	flow := self.Parser.Parse(self.Cmds, script...)
	self.insertPreparation(flow, prep)
	self.filterEmptyCmds(flow)
	return self.executeCmds(flow)
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

func (self *Cli) insertPreparation(cmds *ParsedCmds, prep *ParsedCmds) {
	if prep.GlobalEnv != nil {
		cmds.GlobalEnv.Merge(prep.GlobalEnv)
	}

	hasPowerCmd := false
	for i, cmd := range cmds.Cmds {
		if cmd.IsPowerCmd() {
			hasPowerCmd = true
			continue
		}
		if hasPowerCmd {
			cmds.Cmds = append(append(cmds.Cmds[:i], prep.Cmds...), cmds.Cmds[i:]...)
			return
		}
	}

	if !hasPowerCmd {
		cmds.Cmds = append(prep.Cmds, cmds.Cmds...)
	} else {
		cmds.Cmds = append(cmds.Cmds, prep.Cmds...)
	}
}

func (self *Cli) executeCmds(flow *ParsedCmds) bool {
	env := self.GlobalEnv.GetLayer(EnvLayerSession)
	if flow.GlobalEnv != nil {
		flow.GlobalEnv.WriteTo(env)
	}
	for i := 0; i < len(flow.Cmds); i++ {
		var newCmds []ParsedCmd
		var succeeded bool
		cmd := flow.Cmds[i]
		newCmds, i, succeeded = self.executeCmd(cmd, env, flow.Cmds, i)
		if !succeeded {
			return false
		}
		flow.Cmds = newCmds
	}
	return true
}

func (self *Cli) executeCmd(cmd ParsedCmd, env *Env, cmds []ParsedCmd,
	currCmdIdx int) (newCmds []ParsedCmd, newCurrCmdIdx int, succeeded bool) {

	sep := self.Parser.CmdPathSep()
	// The env modifications from script will be popped out after a command is executed
	// (TODO) But if a mod modified the env, the modifications stay in session level
	cmdEnv := env.NewLayer(EnvLayerCmd)
	for _, seg := range cmd {
		if seg.Env != nil {
			seg.Env.WriteTo(cmdEnv)
		}
	}

	printCmdStack(self.Screen, cmd, cmdEnv, cmds, currCmdIdx, sep)
	last := cmd[len(cmd)-1]
	start := time.Now()
	if last.Cmd.Cmd != nil {
		newCmds, newCurrCmdIdx, succeeded = last.Cmd.Cmd.Execute(self, cmdEnv, cmds, currCmdIdx)
	} else {
		newCmds, newCurrCmdIdx, succeeded = cmds, currCmdIdx, false
	}
	printCmdResult(self.Screen, cmd, cmdEnv, succeeded, time.Now().Sub(start), cmds, currCmdIdx, sep)
	return
}
