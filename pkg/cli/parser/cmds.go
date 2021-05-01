package parser

import (
	"github.com/pingcap/ticat/pkg/cli"
)

type Parser struct {
	seqParser *sequenceParser
	cmdParser *cmdParser
}

// A very simple implement of command line parsing, lack of char escaping
//   * The command line argv list have extra tokenizing info
//         - An example: a quoted string with space inside
//         - TODO: how to store this info(to a file?) and still keep it human-editable ?
//   * The dynamite info(registered modules and env KVs) could use for disambiguation
//         - Inconvenient to use a LEX/YACC lib to parse
func (self *Parser) Parse(tree *cli.CmdTree, envAbbrs *cli.EnvAbbrs, input ...string) *cli.ParsedCmds {
	seqs, firstIsGlobal := self.seqParser.Parse(input)
	cmds := cli.ParsedCmds{cli.ParsedEnv{}, nil}
	for _, seq := range seqs {
		cmds.Cmds = append(cmds.Cmds, self.cmdParser.Parse(tree, envAbbrs, seq))
	}
	if firstIsGlobal && len(cmds.Cmds) != 0 && len(cmds.Cmds[0]) != 0 {
		firstCmd := cmds.Cmds[0]
		for _, seg := range firstCmd {
			cmds.GlobalEnv.Merge(seg.Env)
			seg.Env = nil
		}
	}
	return &cmds
}

func NewParser() *Parser {
	return &Parser{
		&sequenceParser{
			":",
			[]string{"http", "HTTP"},
			[]string{"/"},
		},
		&cmdParser{
			&envParser{
				&brackets{cli.EnvBracketLeft, cli.EnvBracketRight},
				cli.Spaces,
				cli.EnvKeyValSep,
				cli.EnvPathSep,
			},
			cli.CmdPathSep,
			cli.Spaces + cli.CmdPathAlterSeps,
			cli.Spaces,
			cli.CmdRootDisplayName,
		},
	}
}

func (self *Parser) CmdPathSep() string {
	return self.cmdParser.cmdSep
}
