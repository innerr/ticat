package model

import (
	"sort"
)

type BreakPoints struct {
	AtEnd   bool
	Befores map[string]bool
	Afters  map[string]bool
}

func NewBreakPoints() *BreakPoints {
	return &BreakPoints{false, map[string]bool{}, map[string]bool{}}
}

func (self *BreakPoints) SetAtEnd(enabled bool) {
	self.AtEnd = enabled
}

func (self *BreakPoints) BreakAtEnd() bool {
	return self.AtEnd
}

func (self *BreakPoints) SetBefores(cc *Cli, env *Env, cmdList []string) (verifiedCmds []string) {
	for _, cmd := range cmdList {
		verifiedCmd := cc.NormalizeCmd(true, cmd)
		verifiedCmds = append(verifiedCmds, verifiedCmd)
		self.Befores[verifiedCmd] = true
	}
	sort.Strings(verifiedCmds)
	return
}

func (self *BreakPoints) SetAfters(cc *Cli, env *Env, cmdList []string) (verifiedCmds []string) {
	for _, cmd := range cmdList {
		verifiedCmd := cc.NormalizeCmd(true, cmd)
		verifiedCmds = append(verifiedCmds, verifiedCmd)
		self.Afters[verifiedCmd] = true
	}
	sort.Strings(verifiedCmds)
	return
}

func (self *BreakPoints) IsEmpty() bool {
	return self.AtEnd == false && len(self.Befores) == 0 && len(self.Afters) == 0
}

func (self *BreakPoints) Clean(cc *Cli, env *Env) {
	self.AtEnd = false
	self.Befores = map[string]bool{}
	self.Afters = map[string]bool{}
}

func (self *BreakPoints) BreakBefore(cmd string) bool {
	return self.Befores[cmd]
}

func (self *BreakPoints) BreakAfter(cmd string) bool {
	return self.Afters[cmd]
}
