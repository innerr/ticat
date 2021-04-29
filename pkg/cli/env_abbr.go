package cli

import (
	"fmt"
	"strings"
)

type EnvAbbr struct {
	name string
	parent *EnvAbbr
	subs            map[string]*EnvAbbr
	subAbbrsRevIdx map[string]string
}

func NewEnvAbbr() *EnvAbbr {
	return &EnvAbbr {
		"",
		nil,
		map[string]*EnvAbbr{},
		map[string]string{},
	}
}

func (self *EnvAbbr) AddSub(name string, abbrs ...string) *EnvAbbr {
	if _, ok := self.subs[name]; ok {
		panic(fmt.Errorf("[EnvAbbr.AddSub] %s%s: sub-node name conflicted: %s", ErrStrPrefix, self.DisplayPath(), name))
	}
	sub := NewEnvAbbr()
	sub.name = name
	sub.parent = self
	self.subs[name] = sub
	self.addSubAbbr(name, abbrs...)
	return sub
}

func (self *EnvAbbr) addSubAbbr(name string, abbrs ...string) {
	for _, abbr := range abbrs {
		if len(abbr) == 0 {
			continue
		}
		if old, ok := self.subAbbrsRevIdx[abbr]; ok {
			panic(fmt.Errorf(
				"[EnvAbbr.addSubAbbr] %s%s: command abbr name '%s' conflicted, old for '%s', new for '%s'",
				ErrStrPrefix, self.DisplayPath(), abbr, old, name))
		}
		self.subAbbrsRevIdx[abbr] = name
	}
	self.subAbbrsRevIdx[name] = name
}
func (self *EnvAbbr) TryMatch(path string, sep string) (matchedPath []string, matched bool) {
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
			path = ""
		}
		subName, ok := self.subAbbrsRevIdx[candidate]
		if !ok {
			matched = false
			return
		}
		sub, _ := self.subs[subName]
		var subMatchedPath []string
		subMatchedPath, matched = sub.TryMatch(path, sep)
		if len(subMatchedPath) != 0 {
			matchedPath = append(matchedPath, subMatchedPath...)
		}
		return
	}
	matched = true
	return
}

func (self *EnvAbbr) Path() []string {
	if self.parent == nil {
		return []string{}
	}
	return append(self.parent.Path(), self.name)
}

func (self *EnvAbbr) DisplayPath() string {
	path := self.Path()
	if len(path) == 0 {
		return CmdRootNodeName
	} else {
		return strings.Join(path, ".")
	}
}

func (self *EnvAbbr) Name() string {
	return self.name
}

func (self *EnvAbbr) DisplayName() string {
	if len(self.name) == 0 {
		return CmdRootNodeName
	}
	return self.name
}
