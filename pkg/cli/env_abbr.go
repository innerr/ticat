package cli

import (
	"strings"
)

type EnvAbbr struct {
	name string
	subAbbrsRevIdx map[string]string
}

func NewEnvAbbr() *EnvAbbr {
	return &EnvAbbr {
		"",
		map[string]string{},
	}
}

func (self *EnvAbbr) TryMatch(path string, sep string, cmd *CmdTree) bool {
	for len(path) > 0 {
		i := strings.Index(path, sep)
		if i == 0 {
			path = path[1:]
			continue
		}
		var candidate string
		if i > 0 {
			candidate = path[0:i]
			path = path[i + 1:]
		} else {
			candidate = path
		}
		if candidate == self.name {
			return false
		}
		if cmd == nil {
			return false
		}
		realname := cmd.Realname(candidate);
		if len(realname) == 0 {
			return false
		}
	}
	return true
}
