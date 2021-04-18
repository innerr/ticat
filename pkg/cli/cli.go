package cli

func Execute(preparation string, script ...string) bool {
	return NewCli().Execute(preparation, script...)
}

type Cli struct {
	GlobalEnv *Env
	Screen    *Screen
	Cmds      *CmdTree
	Parser    *Parser
}

func NewCli() *Cli {
	cli := &Cli{
		NewEnv(),
		NewScreen(),
		NewCmdTree(),
		NewParser(),
	}
	RegisterBuiltins(cli.Cmds)
	LoadBuiltinEnv(cli.GlobalEnv)
	return cli
}

func (self *Cli) Execute(preparation string, script ...string) bool {
	prep := self.Parser.Parse(self.Cmds, preparation)
	flow := self.Parser.Parse(self.Cmds, script...)
	self.insertPreparation(flow, prep)
	return self.executeCmds(flow)
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
	env := self.GlobalEnv.NewLayer(EnvLayerSession)
	if flow.GlobalEnv != nil {
		flow.GlobalEnv.WriteTo(env)
	}
	for i := 0; i < len(flow.Cmds); i++ {
		cmd := flow.Cmds[i]
		modified, succeeded := self.executeCmd(cmd, env, flow.Cmds, i)
		if !succeeded {
			return false
		}
		flow.Cmds = modified
	}
	return true
}

func (self *Cli) executeCmd(cmd ParsedCmd, env *Env, cmds []ParsedCmd, currCmdIdx int) (modified []ParsedCmd, succeeded bool) {
	sep := self.Parser.CmdPathSep()
	// The env modifications from script will be popped out after a cmd is executed
	// (TODO) But if a mod modified the env, the modifications stay in session level
	cmdEnv := env.NewLayer(EnvLayerMod)
	for _, seg := range cmd {
		if seg.Env != nil {
			seg.Env.WriteTo(cmdEnv)
		}
	}

	printCmdStack(self.Screen, cmd, cmdEnv, cmds, currCmdIdx, sep)
	last := cmd[len(cmd)-1]
	if last.Cmd.Cmd != nil {
		modified, succeeded = last.Cmd.Cmd.Execute(self, env, cmds, currCmdIdx)
	} else {
		modified, succeeded = cmds, false
	}
	printCmdResult(self.Screen, cmd, cmdEnv, succeeded, cmds, currCmdIdx, sep)
	return
}
