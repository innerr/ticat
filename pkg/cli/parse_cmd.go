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

func (self ParsedCmd) Args() *Args {
	if len(self) == 0 {
		return nil
	}
	last := self[len(self)-1].Cmd.Cmd
	if last == nil || last.cmd == nil {
		return nil
	}
	return &last.cmd.args
}

func (self ParsedCmd) IsPowerCmd() bool {
	return len(self) != 0 && self[len(self)-1].IsPowerCmd()
}

func (self ParsedCmd) Path() (path []string) {
	for _, it := range self {
		if it.Cmd.Cmd != nil {
			path = append(path, it.Cmd.Cmd.Name())
		}
	}
	return
}

func (self ParsedCmd) GenEnv(env *Env) *Env {
	env = env.NewLayer(EnvLayerCmd)
	for _, seg := range self {
		if seg.Env != nil {
			seg.Env.WriteTo(env)
		}
	}
	return env
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

	var lastNotExpectArg bool
	var notExpectArg bool

	for len(input) != 0 {
		var env ParsedEnv
		var err error
		var succeeded bool

		// Try to parse input to env
		env, input, succeeded, err = self.envParser.TryParse(curr, input)
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
			parsed = append(parsed, parsedSeg{parsedSegTypeCmd, MatchedCmd{input[0], sub}})
			matchedCmdPath = append(matchedCmdPath, input[0])
			input = input[1:]
			continue
		}

		// Try to parse cmd args
		if lastNotExpectArg {
			self.err("parse", matchedCmdPath, "unknow input '"+strings.Join(input, ",")+
				"', should be a cmd or env definition")
		}
		env, input = self.envParser.TryParseRaw(curr, input)
		if env != nil {
			parsed = append(parsed, parsedSeg{parsedSegTypeEnv, env})
		}
		if len(input) != 0 {
			self.err("parse", matchedCmdPath, "unknow input '"+strings.Join(input, ",")+"', should be args")
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
