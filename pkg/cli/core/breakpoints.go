package core

import (
	"fmt"
	"sort"
	"strings"
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
		parsed := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, cmd)
		if len(parsed.Cmds) != 1 || parsed.FirstErr() != nil {
			panic(fmt.Errorf("[BreakPoints.SetBefore] invalid break-point cmd name '%s'", cmd))
		}
		verifiedCmd := strings.Join(parsed.Cmds[0].Path(), cc.Cmds.Strs.PathSep)
		verifiedCmds = append(verifiedCmds, verifiedCmd)
		self.Before[verifiedCmd] = true
	}
	sort.Strings(verifiedCmds)
	return
}

func (self *BreakPoints) BreakBefore(cmd string) bool {
	return self.Before[cmd]
}
