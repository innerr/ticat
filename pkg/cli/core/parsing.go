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
//      GlobalCmdIdx          - Point to the global area command in sequence in []ParsedCmd

type CliParser interface {
	Parse(cmds *CmdTree, envAbbrs *EnvAbbrs, input ...string) *ParsedCmds
}

type ParsedCmds struct {
	GlobalEnv    ParsedEnv
	Cmds         ParsedCmdSeq
	GlobalCmdIdx int
}

type ParsedCmdSeq []ParsedCmd

func (self *ParsedCmds) Last() (last ParsedCmd) {
	return self.Cmds[len(self.Cmds)-1]
}

func (self ParsedCmdSeq) LastCmd() (last ParsedCmd) {
	if len(self) > 0 {
		last = self[len(self)-1]
	}
	return
}

func (self *ParsedCmds) RemoveLeadingCmds(count int) {
	self.GlobalCmdIdx -= count
	if self.GlobalCmdIdx < 0 {
		self.GlobalCmdIdx = -1
	}
	self.Cmds = self.Cmds[count:]
}

type ParseError struct {
	Input []string
	Error error
}

type ParsedCmd struct {
	Segments   []ParsedCmdSeg
	ParseError ParseError
}

func (self ParsedCmd) IsEmpty() bool {
	return len(self.Segments) == 0
}

func (self ParsedCmd) Last() (seg ParsedCmdSeg) {
	if self.IsEmpty() {
		return
	}
	return self.Segments[len(self.Segments)-1]
}

func (self ParsedCmd) LastCmdNode() (cmd *CmdTree) {
	return self.Last().Matched.Cmd
}

func (self ParsedCmd) LastCmd() (cmd *Cmd) {
	last := self.LastCmdNode()
	if last == nil {
		return
	}
	return last.Cmd()
}

func (self ParsedCmd) Args() (args Args) {
	return self.LastCmd().Args()
}

func (self ParsedCmd) IsPowerCmd() bool {
	return self.Last().IsPowerCmd()
}

func (self ParsedCmd) Help() (help string) {
	return self.Last().Help()
}

func (self ParsedCmd) IsPriority() bool {
	return self.Last().IsPriority()
}

func (self ParsedCmd) DisplayPath(sep string, displayRealname bool) string {
	var path []string
	for _, seg := range self.Segments {
		if seg.Matched.Cmd != nil {
			name := seg.Matched.Name
			realname := seg.Matched.Cmd.Name()
			if displayRealname && name != realname {
				name += "(=" + realname + ")"
			}
			path = append(path, name)
		}
	}
	return strings.Join(path, sep)
}

func (self ParsedCmd) IsAllEmptySegments() bool {
	for _, seg := range self.Segments {
		cmd := seg.Matched.Cmd
		if cmd != nil && cmd.Cmd() != nil {
			return false
		}
	}
	return true
}

func (self ParsedCmd) Path() (path []string) {
	for _, seg := range self.Segments {
		if seg.Matched.Cmd != nil {
			path = append(path, seg.Matched.Cmd.Name())
		}
	}
	return
}

func (self ParsedCmd) MatchedPath() (path []string) {
	for _, seg := range self.Segments {
		if seg.Matched.Cmd != nil {
			path = append(path, seg.Matched.Name)
		}
	}
	return
}

func (self ParsedCmd) GenEnv(env *Env, valDelAllMark string) *Env {
	env = env.NewLayer(EnvLayerCmd)
	for _, seg := range self.Segments {
		if seg.Env != nil {
			seg.Env.WriteTo(env, valDelAllMark)
		}
	}
	return env
}

type ParsedCmdSeg struct {
	Env     ParsedEnv
	Matched MatchedCmd
}

func (self ParsedCmdSeg) IsPowerCmd() bool {
	return !self.Matched.IsEmptyCmd() && self.Matched.GetCmd().IsPowerCmd()
}

func (self ParsedCmdSeg) IsPriority() bool {
	return !self.Matched.IsEmptyCmd() && self.Matched.GetCmd().IsPriority()
}

func (self ParsedCmdSeg) Help() (help string) {
	if self.Matched.IsEmptyCmd() {
		return
	}
	return self.Matched.GetCmd().Help()
}

func (self *ParsedCmdSeg) IsEmpty() bool {
	return self.Env == nil && self.Matched.IsEmpty()
}

type MatchedCmd struct {
	Name string
	Cmd  *CmdTree
}

func (self MatchedCmd) GetCmd() *Cmd {
	if self.Cmd != nil {
		return self.Cmd.Cmd()
	}
	return nil
}

func (self MatchedCmd) IsEmptyCmd() bool {
	return self.Cmd == nil || self.Cmd.Cmd() == nil
}

func (self MatchedCmd) IsEmpty() bool {
	return self.IsEmptyCmd() && len(self.Name) == 0
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

type ParseErrExpectCmd struct {
	Origin error
}

func (self ParseErrExpectCmd) Error() string {
	return self.Origin.Error()
}

type ParseErrExpectArgs struct {
	Origin error
}

func (self ParseErrExpectArgs) Error() string {
	return self.Origin.Error()
}

type ParseErrExpectNoArg struct {
	Origin error
}

func (self ParseErrExpectNoArg) Error() string {
	return self.Origin.Error()
}

type ParseErrEnv struct {
	Origin error
}

func (self ParseErrEnv) Error() string {
	return self.Origin.Error()
}
