package core

import (
	"sort"
)

type BreakPoints struct {
	Before map[string]bool
	After  map[string]bool
}

func NewBreakPoints() *BreakPoints {
	return &BreakPoints{map[string]bool{}, map[string]bool{}}
}

func (self *BreakPoints) SetBefore(cc *Cli, env *Env, cmdList []string) (verifiedCmds []string) {
	for _, cmd := range cmdList {
		verifiedCmd := cc.ParseCmd(cmd, true)
		verifiedCmds = append(verifiedCmds, verifiedCmd)
		self.Before[verifiedCmd] = true
	}
	sort.Strings(verifiedCmds)
	return
}

func (self *BreakPoints) BreakBefore(cmd string) bool {
	return self.Before[cmd]
}
