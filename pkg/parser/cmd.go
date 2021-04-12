package parser

import (
	"fmt"
	"strings"

	"github.com/pingcap/ticat/pkg/cli"
)

type cmdParser struct {
	envParser *envParser
	cmdSep string
	cmdAlterSeps string
	cmdRootNodeName string
}

func (self *cmdParser) Parse(tree *cli.CmdTree, input []string) ParsedCmd {
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
			if !curr.IsEmpty() {
				parsed = append(parsed, curr)
				curr = ParsedCmdSeg{nil, matchedCmd}
			} else {
				curr.Cmd = matchedCmd
			}
			path += matchedCmd.Cmd.Name() + self.cmdSep
		}
	}
	if !curr.IsEmpty() {
		parsed = append(parsed, curr)
	}
	return parsed
}

type ParsedCmd []ParsedCmdSeg

type ParsedCmdSeg struct {
	Env ParsedEnv
	Cmd MatchedCmd
}

func (self *ParsedCmdSeg) IsEmpty() bool {
	return self.Env == nil && len(self.Cmd.Name) == 0 && self.Cmd.Cmd == nil
}

type MatchedCmd struct {
	Name string
	Cmd *cli.CmdTree
}

func (self *cmdParser) parse(tree *cli.CmdTree, input []string) []parsedSeg {
	var parsed []parsedSeg
	var matchedCmdPath []string
	var curr = tree

	for len(input) != 0 {
		var env ParsedEnv
		var err error
		var succeeded bool
		env, input, succeeded, err = self.envParser.TryParse(curr, input)
		if err != nil {
			self.err(matchedCmdPath, "unmatched env brackets '" + strings.Join(input, " ") + "'")
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

		candidate := strings.TrimLeft(input[0], self.cmdAlterSeps)
		subInput := input[1:]
		i := strings.IndexAny(candidate, self.cmdAlterSeps)
		if i >= 0 {
			subInput = append([]string{strings.TrimLeft(candidate[i+1:], self.cmdAlterSeps)}, subInput...)
			candidate = candidate[0:i]
		}

		sub := curr.GetSub(candidate)
		if sub != nil {
			curr = sub
			parsed = append(parsed, parsedSeg{parsedSegTypeCmd, MatchedCmd{candidate, sub}})
			input = subInput
			matchedCmdPath = append(matchedCmdPath, candidate)
			continue
		}

		env, input = self.envParser.TryParseRaw(curr, input)
		if env != nil {
			parsed = append(parsed, parsedSeg{parsedSegTypeEnv, env})
		}
		if len(input) != 0 {
			self.err(matchedCmdPath, "unknow input '" + strings.Join(input, " ") + "'")
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
)

type parsedSeg struct {
	Type parsedSegType
	// Val should be 'ParsedEnv' or 'MatchedCmd'
	Val interface{}
}
