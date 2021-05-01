package parser

import (
	"fmt"
	"strings"

	"github.com/pingcap/ticat/pkg/cli"
)

type envParser struct {
	brackets   *brackets
	spaces     string
	kvSep      string
	envPathSep string
}

func (self *envParser) TryParse(cmd *cli.CmdTree, envAbbrs *cli.EnvAbbrs,
	input []string) (env cli.ParsedEnv, rest []string, found bool, err error) {

	var again bool
	rest, found, again = self.findLeft(input)
	if again {
		return
	}
	if !found {
		return
	}

	var envStrs []string
	envStrs, rest, found = self.findRight(rest)
	if !found {
		return nil, tryTrimStrings(input), true,
			fmt.Errorf("[envParser.TryParse] unmatched env brackets '" + strings.Join(input, " ") + "'")
	}

	var envRest []string
	env, envRest = self.TryParseRaw(cmd, envAbbrs, envStrs)
	if len(envRest) != 0 {
		return nil, tryTrimStrings(input), true,
			fmt.Errorf("[envParser.TryParse] env difinition can't be recognized '" +
				strings.Join(envRest, " ") + "'")
	}

	return env, tryTrimStrings(rest), true, nil
}

func (self *envParser) TryParseRaw(cmd *cli.CmdTree, envAbbrs *cli.EnvAbbrs, input []string) (env cli.ParsedEnv, rest []string) {
	normalized, foundKvSep := normalizeEnvRawStr(input, self.kvSep, cli.Spaces)
	env = cli.ParsedEnv{}
	rest = normalized.data

	genResult := func(normalizedIdx int) []string {
		if normalizedIdx == len(normalized.data) {
			return nil
		}
		originIdx := normalized.originIdx[normalizedIdx]
		originStrIdx := normalized.originStrIdx[normalizedIdx]
		return tryTrimStrings(input[originIdx:][originStrIdx:])
	}

	// It's non-args env definition
	if cmd == nil || cmd.Cmd() == nil {
		if !foundKvSep {
			return tryTrimParsedEnv(env), tryTrimStrings(rest)
		}
		i := 0
		for ; i+2 < len(rest); i += 3 {
			if rest[i+1] != self.kvSep {
				return tryTrimParsedEnv(env), genResult(i)
			}
			key := rest[i]
			if envAbbrs != nil {
				matchedEnvPath, matched := envAbbrs.TryMatch(key, self.envPathSep)
				if matched {
					key = strings.Join(matchedEnvPath, self.envPathSep)
				}
			}
			value := rest[i+2]
			env[key] = cli.ParsedEnvVal{value, false}
		}
		return tryTrimParsedEnv(env), genResult(i)
	}

	// It's args env definition

	args := cmd.Args()
	i := 0
	if !foundKvSep {
		names := args.Names()
		curr := 0
		if len(rest) > len(names) || (len(rest) > 0 && len(args.Realname(rest[0])) != 0) {
			for ; i+1 < len(rest); i += 2 {
				key := args.Realname(rest[i])
				if len(key) == 0 {
					return tryTrimParsedEnv(env), genResult(i)
				}
				value := rest[i+1]
				env[key] = cli.ParsedEnvVal{value, true}
				curr += 1
			}
		} else {
			for ; i < len(rest); i += 1 {
				key := names[i]
				value := rest[i]
				env[key] = cli.ParsedEnvVal{value, true}
				curr += 1
			}
		}
	} else {
		for ; i+2 < len(rest); i += 3 {
			key := args.Realname(rest[i])
			if len(key) == 0 || rest[i+1] != self.kvSep {
				return tryTrimParsedEnv(env), genResult(i)
			}
			value := rest[i+2]
			env[key] = cli.ParsedEnvVal{value, true}
		}
	}
	return tryTrimParsedEnv(env), genResult(i)
}

func (self *envParser) findLeft(input []string) (rest []string, found bool, again bool) {
	rest = tryTrimStrings(input)
	found = false
	again = false

	if len(input) == 0 {
		return
	}
	i := strings.Index(input[0], self.brackets.Left)
	if i < 0 {
		return
	}
	found = true

	leftBrLen := len(self.brackets.Left)
	rest = tryTrimStrings(input[1:])
	if i == 0 {
		if len(input[0]) != leftBrLen {
			rest = append([]string{strings.Trim(input[0][leftBrLen:], self.spaces)}, rest...)
		}
	} else {
		lead := strings.Trim(input[0][0:i], self.spaces)
		tail := strings.TrimLeft(input[0][i+leftBrLen:], self.spaces)
		rest = append([]string{lead, self.brackets.Left, tail}, rest...)
		again = true
	}
	return
}

func (self *envParser) findRight(input []string) (env []string, rest []string, found bool) {
	rightLen := len(self.brackets.Right)

	for i, it := range input {
		k := strings.Index(it, self.brackets.Right)
		if k < 0 {
			if len(it) > 0 {
				env = append(env, it)
			}
			continue
		}
		found = true
		if k != 0 {
			env = append(env, strings.Trim(it[0:k], self.spaces))
		}
		rest = tryTrimStrings(input[i+1:])
		if rightLen != len(it)-k {
			tailOfIt := tryTrimStrings([]string{strings.Trim(it[k+rightLen:], self.spaces)})
			if len(rest) == 0 {
				rest = tailOfIt
			} else {
				rest = append(tailOfIt, rest...)
			}
		}
		return
	}
	return nil, nil, false
}

type normalizedMapping struct {
	data         []string
	originIdx    []int
	originStrIdx []int
}

func normalizeEnvRawStr(input []string, sep string, spaces string) (output normalizedMapping, foundSep bool) {
	for i, it := range input {
		k := strings.Index(it, sep)
		if k < 0 {
			if len(it) != 0 {
				output.data = append(output.data, strings.Trim(it, spaces))
				output.originIdx = append(output.originIdx, i)
				output.originStrIdx = append(output.originStrIdx, 0)
			}
			continue
		}

		foundSep = true

		head := strings.Trim(it[:k], spaces)
		if len(head) != 0 {
			output.data = append(output.data, head)
			output.originIdx = append(output.originIdx, i)
			output.originStrIdx = append(output.originStrIdx, 0)
		}

		output.data = append(output.data, sep)
		output.originIdx = append(output.originIdx, i)
		output.originStrIdx = append(output.originStrIdx, k)

		tail := strings.Trim(it[k+len(sep):], spaces)
		if len(tail) != 0 {
			output.data = append(output.data, tail)
			output.originIdx = append(output.originIdx, i)
			output.originStrIdx = append(output.originStrIdx, k+len(sep))
		}
	}
	return
}

type brackets struct {
	Left  string
	Right string
}

func tryTrimStrings(input []string) []string {
	if len(input) == 0 || len(input) == 1 && len(input[0]) == 0 {
		return nil
	}
	return input
}

func tryTrimParsedEnv(env cli.ParsedEnv) cli.ParsedEnv {
	if len(env) == 0 {
		return nil
	}
	return env
}
