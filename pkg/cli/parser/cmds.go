package parser

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

type Parser struct {
	seqParser *SequenceParser
	cmdParser *CmdParser
}

// A very simple implement of command line parsing, lack of char escaping
//   * The command line argv list have extra tokenizing info
//         - An example: a quoted string with space inside
//         - TODO: how to store this info(to flow file?) and still keep it human-editable ?
//   * The dynamite info(registered modules and env KVs) could use for disambiguation
//         - Inconvenient to use a LEX/YACC lib to parse
func (self *Parser) Parse(
	cmds *core.CmdTree,
	envAbbrs *core.EnvAbbrs,
	input ...string) *core.ParsedCmds {

	seqs, firstIsGlobal := self.seqParser.Parse(input)
	flow := core.ParsedCmds{core.ParsedEnv{}, nil, -1}
	for _, seq := range seqs {
		flow.Cmds = append(flow.Cmds, self.cmdParser.Parse(cmds, envAbbrs, seq))
	}
	if firstIsGlobal && len(flow.Cmds) != 0 {
		flow.GlobalSeqIdx = 0
		// TODO: remove GlobalEnv?
		if len(flow.Cmds[0]) != 0 {
			firstCmd := flow.Cmds[0]
			for _, seg := range firstCmd {
				flow.GlobalEnv.Merge(seg.Env)
				seg.Env = nil
			}
		}
	}
	return &flow
}

func NewParser(seqParser *SequenceParser, cmdParser *CmdParser) *Parser {
	return &Parser{seqParser, cmdParser}
}
