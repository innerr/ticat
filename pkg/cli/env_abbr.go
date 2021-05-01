package cli

import (
	"fmt"
	"strings"
)

type EnvAbbrs struct {
	rootDisplayName string

	name           string
	parent         *EnvAbbrs
	subs           map[string]*EnvAbbrs
	subAbbrs       map[string][]string
	subAbbrsRevIdx map[string]string
}

func NewEnvAbbrs(rootDisplayName string) *EnvAbbrs {
	return &EnvAbbrs{
		rootDisplayName,
		"",
		nil,
		map[string]*EnvAbbrs{},
		map[string][]string{},
		map[string]string{},
	}
}

func (self *EnvAbbrs) GetSub(name string) *EnvAbbrs {
	sub, _ := self.subs[name]
	return sub
}

func (self *EnvAbbrs) AddSub(name string, abbrs ...string) *EnvAbbrs {
	if old, ok := self.subs[name]; ok && old.name != name {
		panic(fmt.Errorf("[EnvAbbrs.AddSub] %s%s: sub-node name conflicted: %s", ErrStrPrefix, self.DisplayPath(), name))
	}
	sub := NewEnvAbbrs(self.rootDisplayName)
	sub.name = name
	sub.parent = self
	self.subs[name] = sub
	self.AddSubAbbrs(name, abbrs...)
	self.subAbbrsRevIdx[name] = name
	return sub
}

func (self *EnvAbbrs) TryMatch(path string, sep string) (matchedPath []string, matched bool) {
	for len(path) > 0 {
		i := strings.Index(path, sep)
		if i == 0 {
			path = path[1:]
			continue
		}
		var candidate string
		if i > 0 {
			candidate = path[0:i]
			path = path[i+1:]
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

func (self *EnvAbbrs) Path() []string {
	if self.parent == nil {
		return []string{}
	}
	return append(self.parent.Path(), self.name)
}

func (self *EnvAbbrs) DisplayPath() string {
	path := self.Path()
	if len(path) == 0 {
		return self.rootDisplayName
	} else {
		return strings.Join(path, ".")
	}
}

func (self *EnvAbbrs) Name() string {
	return self.name
}

func (self *EnvAbbrs) DisplayName() string {
	if len(self.name) == 0 {
		return self.rootDisplayName
	}
	return self.name
}

func (self *EnvAbbrs) AddSubAbbrs(name string, abbrs ...string) {
	for _, abbr := range abbrs {
		if len(abbr) == 0 {
			continue
		}
		if old, ok := self.subAbbrsRevIdx[abbr]; ok && old != name {
			panic(fmt.Errorf(
				"[EnvAbbrs.AddSubAbbrs] %s%s: command abbr name '%s' conflicted, old for '%s', new for '%s'",
				ErrStrPrefix, self.DisplayPath(), abbr, old, name))
		}
		self.subAbbrsRevIdx[abbr] = name
	}
	olds, _ := self.subAbbrs[name]
	self.subAbbrs[name] = append(olds, abbrs...)
}
