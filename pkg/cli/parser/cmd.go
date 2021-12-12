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
	TrivialMark     string
}

func NewCmdParser(
	envParser *EnvParser,
	cmdSep string,
	cmdAlterSeps string,
	cmdSpaces string,
	cmdRootNodeName string,
	TrivialMark string) *CmdParser {

	return &CmdParser{
		envParser,
		cmdSep,
		cmdAlterSeps,
		cmdSpaces,
		cmdRootNodeName,
		TrivialMark,
	}
}

func (self *CmdParser) Parse(
	cmds *core.CmdTree,
	envAbbrs *core.EnvAbbrs,
	input []string) (parsed core.ParsedCmd) {

	// Delay err check
	segs, trivialLvl, err := self.parse(cmds, envAbbrs, input)

	curr := core.ParsedCmdSeg{nil, core.MatchedCmd{}}
	var path []string
	for _, seg := range segs {
		if seg.Type == parsedSegTypeEnv {
			env := seg.Val.(core.ParsedEnv)
			if len(path) != 0 {
				env.AddPrefix(path, self.cmdSep)
			}
			if curr.Env != nil {
				curr.Env.Merge(env)
			} else {
				curr.Env = seg.Val.(core.ParsedEnv)
			}
		} else if seg.Type == parsedSegTypeCmd {
			matchedCmd := seg.Val.(core.MatchedCmd)
			if !curr.IsEmpty() {
				parsed.Segments = append(parsed.Segments, curr)
				curr = core.ParsedCmdSeg{nil, matchedCmd}
			} else {
				curr.Matched = matchedCmd
			}
			path = append(path, matchedCmd.Cmd.Name())
		} else {
			// ignore parsedSegTypeSep
		}
	}
	if !curr.IsEmpty() {
		parsed.Segments = append(parsed.Segments, curr)
	}

	parsed.ParseResult = core.ParseResult{input, err}
	parsed.TrivialLvl = trivialLvl
	return parsed
}

func (self *CmdParser) parse(
	cmds *core.CmdTree,
	envAbbrs *core.EnvAbbrs,
	input []string) (parsed []parsedSeg, trivialLvl int, err error) {

	var matchedCmdPath []string
	var curr = cmds
	var currEnvAbbrs = envAbbrs

	allowSub := true

	for len(input) != 0 {
		var env core.ParsedEnv
		var err error
		var succeeded bool

		// Try to parse trivial level
		for len(input) != 0 {
			stripped := strings.TrimLeft(input[0], "^")
			trivialLvl += len(input[0]) - len(stripped)
			if len(stripped) == 0 {
				input = input[1:]
			} else {
				input[0] = stripped
				break
			}
		}

		// Try to parse input to env
		env, input, succeeded, err = self.envParser.TryParse(curr, currEnvAbbrs, input)
		if err != nil {
			err = fmt.Errorf("[CmdParser.parse] %s: %s", self.displayPath(matchedCmdPath), err.Error())
			return parsed, trivialLvl, core.ParseErrEnv{err}
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
			rest := strings.TrimLeft(input[0][i+1:], self.cmdAlterSeps+self.cmdSpaces)
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
				errStr := "unknow input '" + input[0] + "' ..., should be sub cmd"
				err = fmt.Errorf("[CmdParser.parse] %s: %s", self.displayPath(matchedCmdPath), errStr)
				return parsed, trivialLvl, core.ParseErrExpectCmd{err}
			}
		} else {
			// Try to parse cmd args
			env, input = self.envParser.TryParseRaw(curr, currEnvAbbrs, input)
			if env != nil {
				parsed = append(parsed, parsedSeg{parsedSegTypeEnv, env})
			}
			if len(input) != 0 {
				var errStr string
				if cmdHasArgs(curr) {
					errStr = "args parse failed"
					errStr = "unknow input '" + strings.Join(input, " ") + "', " + errStr
					err = fmt.Errorf("[CmdParser.parse] %s: %s", self.displayPath(matchedCmdPath), errStr)
					return parsed, trivialLvl, core.ParseErrExpectArgs{err}
				} else {
					errStr = "looks like args, but curr cmd has no args"
					errStr = "unknow input '" + strings.Join(input, " ") + "', " + errStr
					err = fmt.Errorf("[CmdParser.parse] %s: %s", self.displayPath(matchedCmdPath), errStr)
					return parsed, trivialLvl, core.ParseErrExpectNoArg{err}
				}
			}
			break
		}
	}

	return parsed, trivialLvl, nil
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

func cmdHasArgs(cmd *core.CmdTree) bool {
	if cmd == nil || cmd.Cmd() == nil {
		return false
	}
	args := cmd.Args()
	return len(args.Names()) != 0
}
