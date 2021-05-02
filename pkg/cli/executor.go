package cli

import (
	"time"

	"github.com/pingcap/ticat/pkg/cli/core"
)

type Executor struct {
}

func (self *Executor) Execute(cc *core.Cli, bootstrap string, env *core.Env, script ...string) bool {
	// This could be very useful for customized mods-loader or env-loader
	// (those loaders will be loaded from 'bootstrap string' above)
	if env != nil {
		cc.GlobalEnv.GetLayer(core.EnvLayerSession).Merge(env)
	}

	if !self.execute(cc, true, bootstrap) {
		return false
	}

	extra := cc.GlobalEnv.Get("bootstrap").Raw
	if len(extra) != 0 && !self.execute(cc, true, extra) {
		return false
	}

	cc.GlobalEnv.PlusInt("sys.stack-depth", 1)
	if !self.execute(cc, false, script...) {
		return false
	}
	cc.GlobalEnv.PlusInt("sys.stack-depth", -1)
	return true
}

func (self *Executor) execute(cc *core.Cli, isBootstrap bool, script ...string) bool {
	if len(script) == 0 {
		return true
	}
	useCmdAbbrs(cc.EnvAbbrs, cc.Cmds)
	cmds := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, script...)
	return self.executeCmds(cc, isBootstrap, cmds)
}

// Remove the cmds only have cmd-level env definication but have no executable
func (self *Executor) filterEmptyCmds(cmds *core.ParsedCmds) {
	var filtered []core.ParsedCmd
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

func (self *Executor) executeCmds(cc *core.Cli, isBootstrap bool, cmds *core.ParsedCmds) bool {
	self.filterEmptyCmds(cmds)
	env := cc.GlobalEnv.GetLayer(core.EnvLayerSession)
	if cmds.GlobalEnv != nil {
		// TODO: XX
		cmds.GlobalEnv.WriteNotArgTo(env, "-", "--")
	}
	for i := 0; i < len(cmds.Cmds); i++ {
		var newCmds []core.ParsedCmd
		var succeeded bool
		cmd := cmds.Cmds[i]
		newCmds, i, succeeded = self.executeCmd(cc, isBootstrap, cmd, env, cmds.Cmds, i)
		if !succeeded {
			return false
		}
		cmds.Cmds = newCmds
	}
	return true
}

func (self *Executor) executeCmd(cc *core.Cli, isBootstrap bool, cmd core.ParsedCmd,
	env *core.Env, cmds []core.ParsedCmd, currCmdIdx int) (newCmds []core.ParsedCmd,
	newCurrCmdIdx int, succeeded bool) {

	sep := cc.Cmds.Strs.PathSep

	// The env modifications from script will be popped out after a command is executed
	// (TODO) But if a mod modified the env, the modifications stay in session level
	// TODO: XX
	cmdEnv := cmd.GenEnv(env, cc.Cmds.Strs.EnvValDelMark, cc.Cmds.Strs.EnvValDelAllMark)
	argv := cmdEnv.GetArgv(cmd.Path(), sep, cmd.Args())

	printCmdStack(isBootstrap, cc.Screen, cmd, cmdEnv, cmds, currCmdIdx, sep)
	last := cmd[len(cmd)-1].Cmd.Cmd
	start := time.Now()
	if last != nil {
		newCmds, newCurrCmdIdx, succeeded = last.Execute(argv, cc, cmdEnv, cmds, currCmdIdx)
	} else {
		newCmds, newCurrCmdIdx, succeeded = cmds, currCmdIdx, false
	}
	printCmdResult(isBootstrap, cc.Screen, cmd, cmdEnv, succeeded, time.Now().Sub(start), cmds, currCmdIdx, sep)
	return
}

func useCmdAbbrs(abbrs *core.EnvAbbrs, tree *core.CmdTree) {
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
