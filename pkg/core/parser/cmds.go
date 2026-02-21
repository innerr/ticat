package parser

import (
	"strings"

	"github.com/innerr/ticat/pkg/core/model"
)

type Parser struct {
	seqParser *SequenceParser
	cmdParser *CmdParser
	helpFlags []string
	sep       string
	envPrefix string
	envSuffix string
}

func NewParser(seqParser *SequenceParser, cmdParser *CmdParser) *Parser {
	return &Parser{
		seqParser: seqParser,
		cmdParser: cmdParser,
		helpFlags: []string{"-h", "--help", "-?", "?", "-help"},
		sep:       ":",
		envPrefix: "{",
		envSuffix: "}",
	}
}

func (self *Parser) isEnvBlock(arg string) bool {
	return strings.HasPrefix(arg, self.envPrefix) && strings.HasSuffix(arg, self.envSuffix)
}

func (self *Parser) transformHelpFlag(input []string) []string {
	if len(input) == 0 {
		return input
	}

	lastArg := input[len(input)-1]
	isHelpFlag := false
	for _, hf := range self.helpFlags {
		if lastArg == hf {
			isHelpFlag = true
			break
		}
	}
	if !isHelpFlag {
		return input
	}

	args := self.seqParser.Normalize(input[:len(input)-1])

	if len(args) == 0 {
		return []string{"help"}
	}

	globalEnvs := []string{}
	remainingArgs := args
	for len(remainingArgs) > 0 && self.isEnvBlock(remainingArgs[0]) {
		globalEnvs = append(globalEnvs, remainingArgs[0])
		remainingArgs = remainingArgs[1:]
	}

	if len(remainingArgs) == 0 {
		return append(globalEnvs, "help")
	}

	firstNonSepIdx := -1
	for i, arg := range remainingArgs {
		if arg != self.sep {
			firstNonSepIdx = i
			break
		}
	}

	if firstNonSepIdx == -1 {
		return append(globalEnvs, "help")
	}

	hasCmdSequence := false
	for i := firstNonSepIdx; i < len(remainingArgs); i++ {
		if remainingArgs[i] == self.sep {
			hasCmdSequence = true
			break
		}
	}

	if hasCmdSequence {
		result := append(globalEnvs, "desc", ".", "more")
		for _, arg := range remainingArgs {
			if arg != self.sep {
				result = append(result, self.sep, arg)
			}
		}
		return result
	}

	return append(globalEnvs, append([]string{"cmd.full-with-flow"}, remainingArgs...)...)
}

func (self *Parser) Parse(
	cmds *model.CmdTree,
	envAbbrs *model.EnvAbbrs,
	input ...string) *model.ParsedCmds {

	input = self.transformHelpFlag(input)

	seqs, firstIsGlobal := self.seqParser.Parse(input)
	flow := model.ParsedCmds{GlobalEnv: model.ParsedEnv{}, Cmds: nil, GlobalCmdIdx: -1}
	for _, seq := range seqs {
		flow.Cmds = append(flow.Cmds, self.cmdParser.Parse(cmds, envAbbrs, seq))
	}
	if firstIsGlobal && len(flow.Cmds) != 0 {
		flow.GlobalCmdIdx = 0
		if !flow.Cmds[0].IsEmpty() {
			firstCmd := flow.Cmds[0]
			for _, seg := range firstCmd.Segments {
				flow.GlobalEnv.Merge(seg.Env)
				seg.Env = nil
			}
		}
	}
	return &flow
}
