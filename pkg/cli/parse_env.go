package cli

import (
	"fmt"
	"strings"
)

type envParser struct {
	brackets *brackets
}

func (self *envParser) TryParse(cmd *CmdTree, input []string) (env ParsedEnv, rest []string, found bool, err error) {
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
		return nil, tryTrim(input), true, fmt.Errorf("unmatched env brackets '" + strings.Join(input, " ") + "'")
	}

	var envRest []string
	env, envRest = self.TryParseRaw(cmd, envStrs)
	if len(envRest) != 0 {
		return nil, tryTrim(input), true, fmt.Errorf("env difinition can't be recognized '" + strings.Join(envRest, " ") + "'")
	}

	return env, tryTrim(rest), true, nil
}

func (self *envParser) TryParseRaw(cmd *CmdTree, input []string) (env ParsedEnv, rest []string) {
	// TODO: more forms
	// TODO: use cmd info
	rest = input
	env = ParsedEnv{}
	for _, it := range input {
		kv := strings.Split(it, "=")
		if len(kv) != 2 {
			return
		}
		env[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		rest = rest[1:]
	}
	if len(env) == 0 {
		env = nil
	}
	rest = tryTrim(rest)
	return
}

type ParsedEnv map[string]string

func (self ParsedEnv) AddPrefix(prefix string) {
	// TODO: slow
	var keys []string
	for k, _ := range self {
		keys = append(keys, k)
	}
	for _, k := range keys {
		self[prefix + k] = self[k]
		delete(self, k)
	}
}

func (self ParsedEnv) Merge(x ParsedEnv) {
	for k, v := range x {
		self[k] = v
	}
}

func (self ParsedEnv) Equal(x ParsedEnv) bool {
	if len(self) != len(x) {
		return false
	}
	for k, v := range x {
		if self[k] != v {
			return false
		}
	}
	return true
}

func (self ParsedEnv) WriteTo(env *Env) {
	for k, v := range self {
		env.Set(k, EnvVal{v, nil})
	}
}

func (self *envParser) findLeft(input []string) (rest []string, found bool, again bool) {
	rest = tryTrim(input)
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
	if i == 0 {
		if len(input[0]) != leftBrLen {
			rest = tryTrim(append([]string{strings.TrimSpace(input[0][leftBrLen:])}, input[1:]...))
		} else {
			rest = tryTrim(input[1:])
		}
	} else {
		lead := strings.TrimSpace(input[0][0:i])
		tail := strings.TrimSpace(input[0][i+leftBrLen:])
		rest = tryTrim(append([]string{lead, self.brackets.Left, tail}, input[1:]...))
		again = true
	}
	return
}

func (self *envParser) findRight(input []string) (env []string, rest []string, found bool) {
	rightLen := len(self.brackets.Right)

	for i, it := range input {
		k := strings.Index(it, self.brackets.Right)
		if k < 0 {
			env = append(env, it)
			continue
		}
		found = true
		if k == 0 {
			if rightLen == len(it) {
				rest = tryTrim(input[i+1:])
				env = tryTrim(input[:i])
				return
			} else {
				rest = tryTrim(append(rest, strings.TrimSpace(it[rightLen:])))
				env = tryTrim(append([]string{strings.TrimSpace(it[rightLen:])}, input[i+1:]...))
				return
			}
		} else {
			if rightLen == len(it)-k {
				env = tryTrim(append(env, strings.TrimSpace(it[0:k])))
				rest = tryTrim(input[i+1:])
				return
			} else {
				env = tryTrim(append(env, strings.TrimSpace(it[0:k])))
				rest = tryTrim(append([]string{strings.TrimSpace(it[k+rightLen:])}, input[i+1:]...))
				return
			}
		}
	}
	return nil, nil, false
}

type brackets struct {
	Left  string
	Right string
}

func tryTrim(input []string) []string {
	if len(input) == 0 || len(input) == 1 && len(input[0]) == 0 {
		return nil
	}
	return input
}
