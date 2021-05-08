package core

import (
	"fmt"
	"strings"
)

type Args struct {
	// Use a map as a set
	names map[string]bool
	// Not support no-default-value arg yet, so names could be insteaded by defVals now
	defVals map[string]string

	orderedList []string
	abbrs       map[string][]string
	abbrsRevIdx map[string]string
}

func newArgs() Args {
	return Args{
		map[string]bool{},
		map[string]string{},
		[]string{},
		map[string][]string{},
		map[string]string{},
	}
}

func (self *Args) AddArg(owner *CmdTree, name string, defVal string, abbrs ...string) {
	if _, ok := self.names[name]; ok {
		panic(fmt.Errorf("[Args.AddArg] %s: arg name conflicted: %s",
			owner.DisplayPath(), name))
	}
	for _, abbr := range abbrs {
		if len(abbr) == 0 {
			continue
		}
		if old, ok := self.abbrsRevIdx[abbr]; ok {
			panic(fmt.Errorf("[Args.AddArg] %s: arg abbr name '%s' conflicted,"+
				" old for '%s', new for '%s'",
				owner.DisplayPath(), abbr, old, name))
		}
		self.abbrsRevIdx[abbr] = name
	}
	self.abbrs[name] = abbrs
	self.abbrsRevIdx[name] = name
	self.names[name] = true
	self.defVals[name] = defVal
	self.orderedList = append(self.orderedList, name)
}

func (self Args) MatchFind(findStr string) bool {
	for k, _ := range self.abbrsRevIdx {
		if strings.Index(k, findStr) >= 0 {
			return true
		}
	}
	return false
}

func (self *Args) Names() []string {
	return self.orderedList
}

func (self *Args) DefVal(name string) string {
	return self.defVals[name]
}

func (self *Args) Realname(nameOrAbbr string) string {
	name, _ := self.abbrsRevIdx[nameOrAbbr]
	return name
}

func (self *Args) Abbrs(name string) (abbrs []string) {
	abbrs, _ = self.abbrs[name]
	return
}
