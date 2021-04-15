package cli

import (
	"fmt"
	"strings"
	"github.com/pingcap/ticat/pkg/hub/cli"
	"github.com/pingcap/ticat/pkg/hub/parser"
)

func Execute(preparation string, script ...string) {
	NewHub().Execute(preparation, script...)
}

type Hub struct {
	GlobalEnv   *cli.Env
	Screen      *cli.Screen
	Cmds        *cli.CmdTree
	Parser      *parser.Parser
}

func NewHub() *Hub {
	hub := &Hub{
		cli.NewEnv(),
		cli.NewScreen(),
		cli.NewCmdTree(),
		parser.NewParser(),
	}
	cli.RegisterBuiltins(hub.Cmds)
	return hub
}

func (self *Hub) Execute(preparation string, script ...string) bool {
	prep := self.Parser.Parse(self.Cmds, preparation)
	flow := self.Parser.Parse(self.Cmds, script...)
	flow.InsertPreparation(prep)
	return self.execute(flow)
}

func (self *hub) executeCmds(flow *parser.ParsedCmds) bool {
	// If a mod modified the env, the modifications stay in session level
	env := self.env.NewLayer(EnvLayerSession)
	if flow.GlobalEnv != nil {
		flow.GlobalEnv.WriteTo(env)
	}
	for i := 0; i < len(flow.Cmds); i++ {
		cmd := flow.Cmds[i]
		// The env modifications from script will be popped out after a cmd is executed
		cmdEnv := env.NewLayer(EnvLayerCmd)
		for _, seg := range cmd {
			if seg.Env != nil {
				seg.Env.WriteTo(cmdEnv)
			}
		}
		seg := cmd[len(cmd)-1]
		if seg.Cmd.Cmd != nil {
			seg.Cmd.Cmd.Execute()
		}
	}
	return true
}
