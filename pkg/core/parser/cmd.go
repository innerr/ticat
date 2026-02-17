package parser

import (
	"strings"

	"github.com/innerr/ticat/pkg/core/model"
)

type CmdParser struct {
	envParser       *EnvParser
	cmdSep          string
	cmdAlterSeps    string
	cmdSpaces       string
	cmdRootNodeName string
	TrivialMark     string
	fakeSepSuffixs  map[byte]bool
}

func NewCmdParser(
	envParser *EnvParser,
	cmdSep string,
	cmdAlterSeps string,
	cmdSpaces string,
	cmdRootNodeName string,
	TrivialMark string,
	fakeSepSuffixs string) *CmdParser {

	fakeSepSuffixMap := map[byte]bool{}
	for _, c := range []byte(fakeSepSuffixs) {
		fakeSepSuffixMap[c] = true
	}
	return &CmdParser{
		envParser,
		cmdSep,
		cmdAlterSeps,
		cmdSpaces,
		cmdRootNodeName,
		TrivialMark,
		fakeSepSuffixMap,
	}
}

func (self *CmdParser) Parse(
	cmds *model.CmdTree,
	envAbbrs *model.EnvAbbrs,
	input []string) (parsed model.ParsedCmd) {

	segs, trivialLvl, err, isMinorErr := self.parse(cmds, envAbbrs, input)

	curr := model.ParsedCmdSeg{Matched: model.MatchedCmd{}}
	var path []string
	for _, seg := range segs {
		if seg.Type == parsedSegTypeEnv {
			env := seg.Val.(model.ParsedEnv)
			if len(path) != 0 {
				env.AddPrefix(path, self.cmdSep)
			}
			if curr.Env != nil {
				curr.Env.Merge(env)
			} else {
				curr.Env = seg.Val.(model.ParsedEnv)
			}
		} else if seg.Type == parsedSegTypeCmd {
			matchedCmd := seg.Val.(model.MatchedCmd)
			if !curr.IsEmpty() {
				parsed.Segments = append(parsed.Segments, curr)
				curr = model.ParsedCmdSeg{Matched: matchedCmd}
			} else {
				curr.Matched = matchedCmd
			}
			if matchedCmd.Cmd != nil {
				path = append(path, matchedCmd.Cmd.Name())
			}
		}
	}
	if !curr.IsEmpty() {
		parsed.Segments = append(parsed.Segments, curr)
	}

	parsed.ParseResult = model.ParseResult{Input: input, Error: err, IsMinorErr: isMinorErr}
	parsed.TrivialLvl = trivialLvl
	return parsed
}

func (self *CmdParser) parse(
	cmds *model.CmdTree,
	envAbbrs *model.EnvAbbrs,
	input []string) (parsed []parsedSeg, trivialLvl int, err error, isMinorErr bool) {

	if len(input) == 0 {
		return nil, 0, nil, false
	}

	input, trivialLvl = self.stripTrivialMarks(input)

	tokens := self.tokenize(input)

	ctx := &yaccParseContext{
		cmdParser:    self,
		envParser:    self.envParser,
		cmds:         cmds,
		envAbbrs:     envAbbrs,
		currCmd:      cmds,
		currEnvAbbrs: envAbbrs,
		matchedPath:  nil,
		trivialLvl:   trivialLvl,
		allowSub:     true,
	}

	lexer := newYYLex(tokens, ctx)
	yyParse(lexer)

	if ctx.err != nil {
		return lexer.result, ctx.trivialLvl, model.ParseErrExpectCmd{Origin: ctx.err}, ctx.isMinorErr
	}

	return lexer.result, ctx.trivialLvl, nil, false
}

func (self *CmdParser) stripTrivialMarks(input []string) ([]string, int) {
	trivialLvl := 0
	for len(input) > 0 {
		stripped := strings.TrimLeft(input[0], self.TrivialMark)
		trivialLvl += len(input[0]) - len(stripped)
		if len(stripped) == 0 {
			input = input[1:]
		} else {
			input[0] = stripped
			break
		}
	}
	return input, trivialLvl
}

func (self *CmdParser) tokenize(input []string) []yyToken {
	var tokens []yyToken

	for _, s := range input {
		tokens = append(tokens, self.tokenizeString(s)...)
	}

	return tokens
}

func (self *CmdParser) tokenizeString(s string) []yyToken {
	var tokens []yyToken

	for len(s) > 0 {
		s = strings.TrimLeft(s, self.cmdSpaces)
		if len(s) == 0 {
			break
		}

		envIdx := strings.Index(s, self.envParser.brackets.Left)
		if envIdx >= 0 {
			if envIdx > 0 {
				tokens = append(tokens, self.tokenizePlain(s[:envIdx])...)
			}
			envStr, rest := self.extractEnvBlock(s[envIdx:])
			if envStr != "" {
				tokens = append(tokens, yyToken{typ: ENV, str: envStr})
				s = rest
				continue
			}
			break
		} else {
			tokens = append(tokens, self.tokenizePlain(s)...)
			break
		}
	}

	return tokens
}

func (self *CmdParser) tokenizePlain(s string) []yyToken {
	var tokens []yyToken

	for len(s) > 0 {
		s = strings.TrimLeft(s, self.cmdSpaces)
		if len(s) == 0 {
			break
		}

		sepIdx := strings.IndexAny(s, self.cmdAlterSeps)
		if sepIdx == 0 {
			if self.isFakeSep(s) {
				tokens = append(tokens, yyToken{typ: WORD, str: s})
				break
			}
			tokens = append(tokens, yyToken{typ: SEP, str: string(s[0])})
			s = s[1:]
		} else if sepIdx > 0 {
			head := strings.TrimRight(s[:sepIdx], self.cmdSpaces)
			if len(head) > 0 {
				tokens = append(tokens, yyToken{typ: WORD, str: head})
			}
			tokens = append(tokens, yyToken{typ: SEP, str: string(s[sepIdx])})
			s = s[sepIdx+1:]
		} else {
			if len(s) > 0 {
				tokens = append(tokens, yyToken{typ: WORD, str: s})
			}
			break
		}
	}

	return tokens
}

func (self *CmdParser) extractEnvBlock(s string) (string, string) {
	if !strings.HasPrefix(s, self.envParser.brackets.Left) {
		return "", s
	}

	left := self.envParser.brackets.Left
	right := self.envParser.brackets.Right
	depth := 0
	pos := 0

	for pos < len(s) {
		if strings.HasPrefix(s[pos:], left) {
			depth++
			pos += len(left)
		} else if strings.HasPrefix(s[pos:], right) {
			depth--
			pos += len(right)
			if depth == 0 {
				return s[:pos], s[pos:]
			}
		} else {
			pos++
		}
	}

	return s, ""
}

func (self *CmdParser) isFakeSep(s string) bool {
	if len(s) < 2 {
		return false
	}
	return self.fakeSepSuffixs[s[1]]
}

func (self *CmdParser) displayPath(matchedCmdPath []string) string {
	displayPath := self.cmdRootNodeName
	if len(matchedCmdPath) != 0 {
		displayPath = strings.Join(matchedCmdPath, self.cmdSep)
	}
	return displayPath
}

type parsedSegType uint

const (
	parsedSegTypeEnv parsedSegType = iota
	parsedSegTypeCmd
	parsedSegTypeSep
)

type parsedSeg struct {
	Type parsedSegType
	// Val should be 'ParsedEnv' or 'MatchedCmd'
	Val interface{}
}

func cmdHasArgs(cmd *model.CmdTree) bool {
	if cmd == nil || cmd.Cmd() == nil {
		return false
	}
	args := cmd.Args()
	return len(args.Names()) != 0
}
