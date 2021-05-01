package parser

import (
	"fmt"
	"strings"

	"github.com/pingcap/ticat/pkg/cli"
)

type cmdParser struct {
	envParser       *envParser
	cmdSep          string
	cmdAlterSeps    string
	cmdSpaces       string
	cmdRootNodeName string
}

func (self *cmdParser) Parse(tree *cli.CmdTree, envAbbrs *cli.EnvAbbrs, input []string) cli.ParsedCmd {
	parsed := cli.ParsedCmd{}
	segs := self.parse(tree, envAbbrs, input)
	curr := cli.ParsedCmdSeg{nil, cli.MatchedCmd{}}
	var path string
	for _, seg := range segs {
		if seg.Type == parsedSegTypeEnv {
			env := seg.Val.(cli.ParsedEnv)
			if len(path) != 0 {
				env.AddPrefix(path)
			}
			if curr.Env != nil {
				curr.Env.Merge(env)
			} else {
				curr.Env = seg.Val.(cli.ParsedEnv)
			}
		} else if seg.Type == parsedSegTypeCmd {
			matchedCmd := seg.Val.(cli.MatchedCmd)
			if !curr.IsEmpty() {
				parsed = append(parsed, curr)
				curr = cli.ParsedCmdSeg{nil, matchedCmd}
			} else {
				curr.Cmd = matchedCmd
			}
			path += matchedCmd.Cmd.Name() + self.cmdSep
		} else {
			// ignore parsedSegTypeSep
		}
	}
	if !curr.IsEmpty() {
		parsed = append(parsed, curr)
	}
	return parsed
}

func (self *cmdParser) parse(tree *cli.CmdTree, envAbbrs *cli.EnvAbbrs, input []string) []parsedSeg {
	var parsed []parsedSeg
	var matchedCmdPath []string
	var curr = tree
	var currEnvAbbrs = envAbbrs

	var lastNotExpectArg bool
	var notExpectArg bool

	for len(input) != 0 {
		var env cli.ParsedEnv
		var err error
		var succeeded bool

		// Try to parse input to env
		env, input, succeeded, err = self.envParser.TryParse(curr, currEnvAbbrs, input)
		if err != nil {
			self.err("parse", matchedCmdPath, "unmatched env brackets '"+strings.Join(input, " ")+"'")
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

		// Try to split input by cmd-sep
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
		}
		lastNotExpectArg = notExpectArg
		notExpectArg = (i >= 0)
		if i >= 0 {
			continue
		}

		// Try to parse input as cmd-seg to sub
		sub := curr.GetSub(input[0])
		if sub != nil {
			curr = sub
			if currEnvAbbrs != nil {
				currEnvAbbrs = currEnvAbbrs.GetSub(input[0])
			}
			parsed = append(parsed, parsedSeg{parsedSegTypeCmd, cli.MatchedCmd{input[0], sub}})
			matchedCmdPath = append(matchedCmdPath, input[0])
			input = input[1:]
			continue
		}

		// Try to parse cmd args
		if lastNotExpectArg {
			self.err("parse", matchedCmdPath, "unknow input '"+strings.Join(input, ",")+
				"', should be a cmd or env definition")
		}
		env, input = self.envParser.TryParseRaw(curr, currEnvAbbrs, input)
		if env != nil {
			parsed = append(parsed, parsedSeg{parsedSegTypeEnv, env})
		}
		if len(input) != 0 {
			self.err("parse", matchedCmdPath, "unknow input '"+strings.Join(input, ","))
		}
		break
	}

	return parsed
}

func (self *cmdParser) err(function string, matchedCmdPath []string, msg string) {
	displayPath := self.cmdRootNodeName
	if len(matchedCmdPath) != 0 {
		displayPath = strings.Join(matchedCmdPath, self.cmdSep)
	}
	panic(fmt.Errorf("[cmdParser.%s] %s: %s", function, displayPath, msg))
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
