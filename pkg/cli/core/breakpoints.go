package core

import (
	"sort"
)

type BreakPoints struct {
	Befores map[string]bool
	Afters  map[string]bool
}

func NewBreakPoints() *BreakPoints {
	return &BreakPoints{map[string]bool{}, map[string]bool{}}
}

func (self *BreakPoints) SetBefores(cc *Cli, env *Env, cmdList []string) (verifiedCmds []string) {
	for _, cmd := range cmdList {
		verifiedCmd := cc.ParseCmd(cmd, true)
		verifiedCmds = append(verifiedCmds, verifiedCmd)
		self.Befores[verifiedCmd] = true
	}
	sort.Strings(verifiedCmds)
	return
}

func (self *BreakPoints) SetAfters(cc *Cli, env *Env, cmdList []string) (verifiedCmds []string) {
	for _, cmd := range cmdList {
		verifiedCmd := cc.ParseCmd(cmd, true)
		verifiedCmds = append(verifiedCmds, verifiedCmd)
		self.Afters[verifiedCmd] = true
	}
	sort.Strings(verifiedCmds)
	return
}

func (self *BreakPoints) BreakBefore(cmd string) bool {
	return self.Befores[cmd]
}

func (self *BreakPoints) BreakAfter(cmd string) bool {
	return self.Afters[cmd]
}
