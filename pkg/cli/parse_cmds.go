package cli

type Parser struct {
	seqParser *sequenceParser
	cmdParser *cmdParser
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
			"\t\n\r./ ",
			"<root>",
		},
	}
}

// A simple implement of command line parsing, lack of char escaping
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

type ParsedCmds struct {
	GlobalEnv ParsedEnv
	Cmds []ParsedCmd
}

func (self *ParsedCmds) InsertPreparation(prep *ParsedCmds) {
	if prep.GlobalEnv != nil {
		self.GlobalEnv.Merge(prep.GlobalEnv)
	}

	hasPowerCmd := false
	for i, cmd := range self.Cmds {
		if cmd.IsPowerCmd() {
			hasPowerCmd = true
			continue
		}
		if hasPowerCmd {
			self.Cmds = append(append(self.Cmds[:i], prep.Cmds...), self.Cmds[i:]...)
			return
		}
	}

	if !hasPowerCmd {
		self.Cmds = append(prep.Cmds, self.Cmds...)
	} else {
		self.Cmds = append(self.Cmds, prep.Cmds...)
	}
}
