package cli

import (
	"fmt"
)

type Args struct {
	// Use a map as a set
	pairs       map[string]bool
	list        []string
	defVals     map[string]string
	abbrsRevIdx map[string]string
}

func newArgs() Args {
	return Args{
		map[string]bool{},
		[]string{},
		map[string]string{},
		map[string]string{},
	}
}

func (self *Args) AddArg(owner *CmdTree, name string, defVal string, abbrs ...string) {
	if _, ok := self.pairs[name]; ok {
		panic(fmt.Errorf("[Args.AddArg] %s%s: arg name conflicted: %s", ErrStrPrefix, owner.DisplayPath(), name))
	}
	for _, abbr := range abbrs {
		if old, ok := self.abbrsRevIdx[abbr]; ok {
			panic(fmt.Errorf("[Args.AddArg] %s%s: arg abbr name '%s' conflicted, old for '%s', new for '%s'",
				ErrStrPrefix, owner.DisplayPath(), abbr, old, name))
		}
		self.abbrsRevIdx[abbr] = name
	}
	self.abbrsRevIdx[name] = name
	self.pairs[name] = true
	self.defVals[name] = defVal
	self.list = append(self.list, name)
}

func (self *Args) List() []string {
	return self.list
}

func (self *Args) DefVal(name string) string {
	return self.defVals[name]
}

func (self *Args) Realname(nameOrAbbr string) string {
	name, _ := self.abbrsRevIdx[nameOrAbbr]
	return name
}
