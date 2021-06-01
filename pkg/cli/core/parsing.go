package core

import (
	"strings"
)

//  ParsedCmds                - A list of cmd
//      ParsedEnv             - Global env, map[string]string
//      []ParsedCmd           - Full path of cmd
//          []ParsedCmdSeg    - A path = a segment list
//              MatchedCmd    - A segment
//                  Name      - string
//                  *CmdTree  - The executable function
//              ParsedEnv     - The function's env, include argv
//      GlobalSeqIdx          - Point to the global seequence in []ParsedCmd

type CliParser interface {
	Parse(cmds *CmdTree, envAbbrs *EnvAbbrs, input ...string) *ParsedCmds
}

type ParsedCmds struct {
	GlobalEnv    ParsedEnv
	Cmds         []ParsedCmd
	GlobalSeqIdx int
}

func (self *ParsedCmds) RmLeadingCmds(count int) {
	self.GlobalSeqIdx -= count
	if self.GlobalSeqIdx < 0 {
		self.GlobalSeqIdx = -1
	}
	self.Cmds = self.Cmds[count:]
}

type ParsedCmd []ParsedCmdSeg

func (self ParsedCmd) Args() (args Args) {
	if len(self) == 0 {
		return
	}
	last := self[len(self)-1].Cmd.Cmd
	if last == nil || last.Cmd() == nil {
		return
	}
	args = last.cmd.Args()
	return
}

func (self ParsedCmd) IsPowerCmd() bool {
	return len(self) != 0 && self[len(self)-1].IsPowerCmd()
}

func (self ParsedCmd) Help() (help string) {
	if len(self) == 0 {
		return
	}
	return self[len(self)-1].Help()
}

func (self ParsedCmd) IsPriority() bool {
	return len(self) != 0 && self[len(self)-1].IsPriority()
}

func (self ParsedCmd) LastCmd() (cmd *Cmd) {
	if len(self) == 0 {
		return
	}
	last := self[len(self)-1].Cmd.Cmd
	if last == nil {
		return
	}
	return last.Cmd()
}

func (self ParsedCmd) DisplayPath(sep string, displayRealname bool) string {
	var path []string
	for _, seg := range self {
		if seg.Cmd.Cmd != nil {
			name := seg.Cmd.Name
			realname := seg.Cmd.Cmd.Name()
			if displayRealname && name != realname {
				name += "(=" + realname + ")"
			}
			path = append(path, name)
		}
	}
	return strings.Join(path, sep)
}

func (self ParsedCmd) TotallyEmpty() bool {
	for _, seg := range self {
		cmd := seg.Cmd.Cmd
		if cmd != nil && cmd.Cmd() != nil {
			return false
		}
	}
	return true
}

func (self ParsedCmd) Path() (path []string) {
	for _, it := range self {
		if it.Cmd.Cmd != nil {
			path = append(path, it.Cmd.Cmd.Name())
		}
	}
	return
}

func (self ParsedCmd) MatchedPath() (path []string) {
	for _, it := range self {
		if it.Cmd.Cmd != nil {
			path = append(path, it.Cmd.Name)
		}
	}
	return
}

func (self ParsedCmd) GenEnv(env *Env, valDelAllMark string) *Env {
	env = env.NewLayer(EnvLayerCmd)
	for _, seg := range self {
		if seg.Env != nil {
			seg.Env.WriteTo(env, valDelAllMark)
		}
	}
	return env
}

type ParsedCmdSeg struct {
	Env ParsedEnv
	Cmd MatchedCmd
}

func (self ParsedCmdSeg) IsPowerCmd() bool {
	return self.Cmd.Cmd != nil && self.Cmd.Cmd.Cmd() != nil && self.Cmd.Cmd.IsPowerCmd()
}

func (self ParsedCmdSeg) IsPriority() bool {
	return self.Cmd.Cmd != nil && self.Cmd.Cmd.Cmd() != nil && self.Cmd.Cmd.Cmd().IsPriority()
}

func (self *ParsedCmdSeg) IsEmpty() bool {
	return self.Env == nil && len(self.Cmd.Name) == 0 && self.Cmd.Cmd == nil
}

func (self ParsedCmdSeg) Help() (help string) {
	if self.Cmd.Cmd == nil || self.Cmd.Cmd.Cmd() == nil {
		return
	}
	return self.Cmd.Cmd.Cmd().Help()
}

type MatchedCmd struct {
	Name string
	Cmd  *CmdTree
}

type ParsedEnv map[string]ParsedEnvVal

func (self ParsedEnv) AddPrefix(prefix []string, sep string) {
	var keys []string
	var vals []ParsedEnvVal
	for k, v := range self {
		keys = append(keys, k)
		vals = append(vals, v)
	}

	prefixPath := strings.Join(prefix, sep) + sep
	for i, k := range keys {
		v := vals[i]
		self[prefixPath+k] = ParsedEnvVal{v.Val, v.IsArg, append(prefix, v.MatchedPath...)}
		delete(self, k)
	}
}

func (self ParsedEnv) Merge(x ParsedEnv) {
	for k, v := range x {
		self[k] = v
	}
}

func (self ParsedEnv) Equal(x ParsedEnv) bool {
	if len(self) != len(x) {
		return false
	}
	for k, v := range x {
		if self[k].Val != v.Val {
			return false
		}
	}
	return true
}

func (self ParsedEnv) WriteTo(env *Env, valDelAllMark string) {
	for k, v := range self {
		if v.Val == valDelAllMark {
			env.Delete(k)
		} else {
			env.SetEx(k, v.Val, v.IsArg)
		}
	}
}

func (self ParsedEnv) WriteNotArgTo(env *Env, valDelAllMark string) {
	for k, v := range self {
		if v.Val == valDelAllMark {
			env.Delete(k)
		} else if !v.IsArg {
			env.SetEx(k, v.Val, v.IsArg)
		}
	}
}

type ParsedEnvVal struct {
	Val         string
	IsArg       bool
	MatchedPath []string
}

func NewParsedEnvVal(key string, val string) ParsedEnvVal {
	return ParsedEnvVal{val, false, []string{key}}
}

func NewParsedEnvArgv(key string, val string) ParsedEnvVal {
	return ParsedEnvVal{val, true, []string{key}}
}
