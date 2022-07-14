package core

import (
	"fmt"
	"strings"
)

type Args struct {
	// map[arg-name]arg-index
	names map[string]int

	// Not support no-default-value arg yet, so names could be insteaded by defVals now
	defVals map[string]string

	orderedList []string
	abbrs       map[string][]string
	abbrsRevIdx map[string]string

	fromAutoMapAll map[string]bool
}

func newArgs() Args {
	return Args{
		map[string]int{},
		map[string]string{},
		[]string{},
		map[string][]string{},
		map[string]string{},
		map[string]bool{},
	}
}

func (self *Args) AddArg(owner *CmdTree, name string, defVal string, abbrs ...string) {
	if _, ok := self.names[name]; ok {
		panic(fmt.Errorf("[Args.AddArg] %s: arg name conflicted: %s",
			owner.DisplayPath(), name))
	}
	abbrs = append([]string{name}, abbrs...)
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
	self.names[name] = len(self.names)
	self.defVals[name] = defVal
	self.orderedList = append(self.orderedList, name)
}

func (self *Args) AddAutoMapAllArg(owner *CmdTree, name string, defVal string, abbrs ...string) {
	self.AddArg(owner, name, defVal, abbrs...)
	self.fromAutoMapAll[name] = true
}

func (self *Args) IsFromAutoMapAll(name string) bool {
	return self.fromAutoMapAll[name]
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

func (self *Args) Reorder(owner *Cmd, names []string) {
	changed := func() {
		panic(fmt.Errorf("[%s] args changed in reordering, origin: %s; new: %s",
			owner.Owner().DisplayPath(), strings.Join(self.orderedList, ","), strings.Join(names, ",")))
	}
	if len(names) != len(self.orderedList) {
		changed()
	}
	for i, name := range names {
		if _, ok := self.names[name]; !ok {
			changed()
		}
		self.names[name] = i
	}
	self.orderedList = names
}

func (self *Args) DefVal(name string, stackDepth int) string {
	if self.fromAutoMapAll[name] && stackDepth > 1 {
		return ""
	}
	return self.defVals[name]
}

func (self *Args) RawDefVal(name string) string {
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

func (self *Args) Has(name string) bool {
	_, ok := self.names[name]
	return ok
}

func (self *Args) HasArgOrAbbr(nameOrAbbr string) bool {
	_, ok := self.abbrsRevIdx[nameOrAbbr]
	return ok
}

func (self *Args) Index(name string) int {
	index, ok := self.names[name]
	if !ok {
		return -1
	}
	return index
}
