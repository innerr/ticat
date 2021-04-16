package cli

import (
	"fmt"
	"strings"
)

type cmdParser struct {
	envParser       *envParser
	cmdSep          string
	cmdAlterSeps    string
	cmdSpaces       string
	cmdRootNodeName string
}

func (self *cmdParser) Parse(tree *CmdTree, input []string) ParsedCmd {
	parsed := ParsedCmd{}
	segs := self.parse(tree, input)
	curr := ParsedCmdSeg{nil, MatchedCmd{}}
	path := ""
	for _, seg := range segs {
		if seg.Type == parsedSegTypeEnv {
			env := seg.Val.(ParsedEnv)
			if len(path) != 0 {
				env.AddPrefix(path)
			}
			if curr.Env != nil {
				curr.Env.Merge(env)
			} else {
				curr.Env = seg.Val.(ParsedEnv)
			}
		} else if seg.Type == parsedSegTypeCmd {
			matchedCmd := seg.Val.(MatchedCmd)
			if !curr.isEmpty() {
				parsed = append(parsed, curr)
				curr = ParsedCmdSeg{nil, matchedCmd}
			} else {
				curr.Cmd = matchedCmd
			}
			path += matchedCmd.Cmd.Name() + self.cmdSep
		} else {
			// ignore parsedSegTypeSep
		}
	}
	if !curr.isEmpty() {
		parsed = append(parsed, curr)
	}
	return parsed
}

type ParsedCmd []ParsedCmdSeg

func (self ParsedCmd) IsPowerCmd() bool {
	return len(self) != 0 && self[len(self)-1].IsPowerCmd()
}

type ParsedCmdSeg struct {
	Env ParsedEnv
	Cmd MatchedCmd
}

func (self ParsedCmdSeg) IsPowerCmd() bool {
	return self.Cmd.Cmd != nil && self.Cmd.Cmd.IsPowerCmd()
}

func (self *ParsedCmdSeg) isEmpty() bool {
	return self.Env == nil && len(self.Cmd.Name) == 0 && self.Cmd.Cmd == nil
}

type MatchedCmd struct {
	Name string
	Cmd  *CmdTree
}

func (self *cmdParser) parse(tree *CmdTree, input []string) []parsedSeg {
	var parsed []parsedSeg
	var matchedCmdPath []string
	var curr = tree

	for len(input) != 0 {
		var env ParsedEnv
		var err error
		var succeeded bool
		env, input, succeeded, err = self.envParser.TryParse(curr, input)
		if err != nil {
			self.err(matchedCmdPath, "unmatched env brackets '"+strings.Join(input, " ")+"'")
			break
		}
		if succeeded {
			if env != nil {
				parsed = append(parsed, parsedSeg{parsedSegTypeEnv, env})
			}
			continue
		}

		if len(input) == 0 {
			break
		}

		i := strings.IndexAny(input[0], self.cmdAlterSeps)
		if i == 0 {
			if len(parsed) != 0 && parsed[len(parsed)-1].Type != parsedSegTypeSep {
				parsed = append(parsed, parsedSeg{parsedSegTypeSep, nil})
			}
			if len(input[0]) == 1 {
				input = input[1:]
			} else {
				input = append([]string{input[0][1:]}, input[1:]...)
			}
			continue
		} else if i > 0 {
			head := input[0][0:i]
			rest := strings.TrimLeft(input[0][i+1:], self.cmdAlterSeps)
			input = input[1:]
			var lead []string
			if len(head) != 0 {
				lead = append(lead, head)
			}
			lead = append(lead, self.cmdSep)
			if len(rest) != 0 {
				lead = append(lead, rest)
			}
			input = append(lead, input...)
			continue
		}

		sub := curr.GetSub(input[0])
		if sub != nil {
			curr = sub
			parsed = append(parsed, parsedSeg{parsedSegTypeCmd, MatchedCmd{input[0], sub}})
			matchedCmdPath = append(matchedCmdPath, input[0])
			input = input[1:]
			continue
		}

		env, input = self.envParser.TryParseRaw(curr, input)
		if env != nil {
			parsed = append(parsed, parsedSeg{parsedSegTypeEnv, env})
		}
		if len(input) != 0 {
			self.err(matchedCmdPath, "unknow input '"+strings.Join(input, ",")+"'")
		}
		break
	}

	return parsed
}

func (self *cmdParser) err(matchedCmdPath []string, msg string) {
	displayPath := self.cmdRootNodeName
	if len(matchedCmdPath) != 0 {
		strings.Join(matchedCmdPath, self.cmdSep)
	}
	panic(fmt.Errorf("%s: %s", displayPath, msg))
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
