package cli

type Parser struct {
	seqParser *sequenceParser
	cmdParser *cmdParser
}

// A very simple implement of command line parsing, lack of char escaping
//
//  ParsedCmds                - A list of cmd
//      ParsedEnv             - Global env, map[string]string
//      []ParsedCmd           - Full path of cmd
//          []ParsedCmdSeg    - A path = a segment list
//              MatchedCmd    - A segment
//                  Name      - string
//                  *CmdTree  - The executable function
//              ParsedEnv     - The function's env, include argv
//
func (self *Parser) Parse(tree *CmdTree, input ...string) *ParsedCmds {
	seqs, firstIsGlobal := self.seqParser.Parse(input)
	cmds := ParsedCmds{nil, nil}
	for _, seq := range seqs {
		cmds.Cmds = append(cmds.Cmds, self.cmdParser.Parse(tree, seq))
	}
	if firstIsGlobal && len(cmds.Cmds) != 0 && len(cmds.Cmds[0]) != 0 {
		cmds.GlobalEnv = cmds.Cmds[0][0].Env
		cmds.Cmds[0][0].Env = nil
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
			&envParser{&brackets{"{", "}"}},
			".",
			"\t\n\r ./",
			"\t\n\r ",
			"<root>",
		},
	}
}

func (self *Parser) CmdPathSep() string {
	return self.cmdParser.cmdSep
}

type ParsedCmds struct {
	GlobalEnv ParsedEnv
	Cmds      []ParsedCmd
}
