package parser

import (
	"fmt"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

type CmdParser struct {
	envParser       *EnvParser
	cmdSep          string
	cmdAlterSeps    string
	cmdSpaces       string
	cmdRootNodeName string
}

func NewCmdParser(
	envParser *EnvParser,
	cmdSep string,
	cmdAlterSeps string,
	cmdSpaces string,
	cmdRootNodeName string) *CmdParser {

	return &CmdParser{
		envParser,
		cmdSep,
		cmdAlterSeps,
		cmdSpaces,
		cmdRootNodeName,
	}
}

func (self *CmdParser) Parse(
	cmds *core.CmdTree,
	envAbbrs *core.EnvAbbrs,
	input []string) core.ParsedCmd {

	parsed := core.ParsedCmd{}
	segs := self.parse(cmds, envAbbrs, input)
	curr := core.ParsedCmdSeg{nil, core.MatchedCmd{}}
	var path string
	for _, seg := range segs {
		if seg.Type == parsedSegTypeEnv {
			env := seg.Val.(core.ParsedEnv)
			if len(path) != 0 {
				env.AddPrefix(path)
			}
			if curr.Env != nil {
				curr.Env.Merge(env)
			} else {
				curr.Env = seg.Val.(core.ParsedEnv)
			}
		} else if seg.Type == parsedSegTypeCmd {
			matchedCmd := seg.Val.(core.MatchedCmd)
			if !curr.IsEmpty() {
				parsed = append(parsed, curr)
				curr = core.ParsedCmdSeg{nil, matchedCmd}
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

func (self *CmdParser) parse(
	cmds *core.CmdTree,
	envAbbrs *core.EnvAbbrs,
	input []string) []parsedSeg {

	var parsed []parsedSeg
	var matchedCmdPath []string
	var curr = cmds
	var currEnvAbbrs = envAbbrs

	allowSub := true

	for len(input) != 0 {
		var env core.ParsedEnv
		var err error
		var succeeded bool

		// Try to parse input to env
		env, input, succeeded, err = self.envParser.TryParse(curr, currEnvAbbrs, input)
		if err != nil {
			self.err("parse", matchedCmdPath, err.Error())
			break
		}
		if succeeded {
			if env != nil {
				parsed = append(parsed, parsedSeg{parsedSegTypeEnv, env})
			}
			// Allow use an env segment as cmd-path-sep
			allowSub = true
			continue
		}

		if len(input) == 0 {
			break
		}

		// Try to split input by cmd-sep
		i := strings.IndexAny(input[0], self.cmdAlterSeps)
		if i == 0 {
			// Tolerat redundant path-sep
			if len(parsed) != 0 && parsed[len(parsed)-1].Type != parsedSegTypeSep {
				parsed = append(parsed, parsedSeg{parsedSegTypeSep, nil})
			}
			if len(input[0]) == 1 {
				input = input[1:]
			} else {
				rest := strings.TrimLeft(input[0][1:], self.cmdSpaces)
				input = append([]string{rest}, input[1:]...)
			}
			allowSub = true
			continue
		} else if i > 0 && allowSub {
			head := strings.TrimRight(input[0][0:i], self.cmdSpaces)
			rest := strings.TrimLeft(input[0][i+1:], self.cmdAlterSeps + self.cmdSpaces)
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

		if allowSub {
			// Try to parse input as cmd-seg to sub
			sub := curr.GetSub(input[0])
			if sub != nil {
				curr = sub
				if currEnvAbbrs != nil {
					currEnvAbbrs = currEnvAbbrs.GetSub(input[0])
				}
				parsed = append(parsed, parsedSeg{parsedSegTypeCmd, core.MatchedCmd{input[0], sub}})
				matchedCmdPath = append(matchedCmdPath, input[0])
				input = input[1:]
				allowSub = false
				continue
			} else {
				self.err("parse", matchedCmdPath, "unknow input '"+
					strings.Join(input, " ")+"', shoud be sub cmd")
			}
		} else {
			// Try to parse cmd args
			env, input = self.envParser.TryParseRaw(curr, currEnvAbbrs, input)
			if env != nil {
				parsed = append(parsed, parsedSeg{parsedSegTypeEnv, env})
			}
			if len(input) != 0 {
				brackets := strings.Join(self.envParser.Brackets(), "")
				self.err("parse", matchedCmdPath, "unknow input '"+strings.Join(input, " ")+
					"', should be args, tips: try to enclose env definition or args with '"+brackets+"' to disambiguation")
			}
			break
		}
	}

	return parsed
}

func (self *CmdParser) err(function string, matchedCmdPath []string, msg string) {
	displayPath := self.cmdRootNodeName
	if len(matchedCmdPath) != 0 {
		displayPath = strings.Join(matchedCmdPath, self.cmdSep)
	}
	panic(fmt.Errorf("[CmdParser.%s] %s: %s", function, displayPath, msg))
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
