package core

import (
	"fmt"
	"strings"
)

// TODO: share some code with EnvAbbrs ?

type CmdTreeStrs struct {
	SelfName                 string
	RootDisplayName          string
	BuiltinName              string
	BuiltinDisplayName       string
	PathSep                  string
	PathAlterSeps            string
	AbbrsSep                 string
	EnvOpSep                 string
	EnvValDelAllMark         string
	EnvKeyValSep             string
	EnvPathSep               string
	ProtoSep                 string
	ListSep                  string
	FlowTemplateBracketLeft  string
	FlowTemplateBracketRight string
	FlowTemplateMultiplyMark string
	TagMark                  string
}

func CmdTreeStrsForTest() *CmdTreeStrs {
	return &CmdTreeStrs{"self", "<root>", "builtin", "<builtin>",
		".", ".", "|", ":", "--", "=", ".", "\t", ",", "[[", "]]", "*", "@"}
}

type CmdTree struct {
	Strs            *CmdTreeStrs
	name            string
	parent          *CmdTree
	subs            map[string]*CmdTree
	subOrderedNames []string
	cmd             *Cmd
	subAbbrs        map[string][]string
	subAbbrsRevIdx  map[string]string
	hidden          bool
	source          string
	tags            []string
	trivial         int
}

func NewCmdTree(strs *CmdTreeStrs) *CmdTree {
	return &CmdTree{
		strs,
		"",
		nil,
		map[string]*CmdTree{},
		[]string{},
		nil,
		map[string][]string{},
		map[string]string{},
		false,
		"",
		nil,
		0,
	}
}

func (self *CmdTree) Execute(
	argv ArgVals,
	sysArgv SysArgVals,
	cc *Cli,
	env *Env,
	mask *ExecuteMask,
	flow *ParsedCmds,
	currCmdIdx int) (int, bool) {

	if self.cmd == nil {
		return currCmdIdx, true
	} else {
		return self.cmd.Execute(argv, cc, env, mask, flow, currCmdIdx)
	}
}

func (self *CmdTree) cmdConflictCheck(help string, funName string) {
	if self.cmd == nil {
		return
	}
	// Allow to overwrite empty dir cmd
	if self.cmd.Type() == CmdTypeEmptyDir {
		return
	}
	err := &CmdTreeErrExecutableConflicted{
		fmt.Sprintf("reg-cmd conflicted. old-help '%s', new-help '%s'",
			strings.Split(self.cmd.Help(), "\n")[0],
			strings.Split(help, "\n")[0]),
		self.Path(),
		self.Source(),
	}
	panic(err)
}

func (self *CmdTree) SetHidden() *CmdTree {
	self.hidden = true
	return self
}

func (self *CmdTree) SetTrivial(val int) *CmdTree {
	self.trivial = val
	return self
}

func (self *CmdTree) IsHidden() bool {
	return self.hidden
}

func (self *CmdTree) AddTags(tags ...string) {
	self.tags = append(self.tags, tags...)
}

func (self *CmdTree) Tags() []string {
	return self.tags
}

func (self *CmdTree) Trivial() int {
	return self.trivial
}

func (self *CmdTree) MatchWriteKey(key string) bool {
	if self.cmd == nil {
		return false
	}
	return self.cmd.MatchWriteKey(key)
}

func (self *CmdTree) GatherNames() (names []string) {
	names = append(names, self.DisplayPath())
	for _, sub := range self.subs {
		// TODO: slow
		names = append(names, sub.GatherNames()...)
	}
	return names
}

func (self *CmdTree) RegCmd(cmd NormalCmd, help string) *Cmd {
	self.cmdConflictCheck(help, "RegCmd")
	self.cmd = NewCmd(self, help, cmd)
	return self.cmd
}

func (self *CmdTree) RegFileCmd(cmd string, help string) *Cmd {
	self.cmdConflictCheck(help, "RegFileCmd")
	self.cmd = NewFileCmd(self, help, cmd)
	return self.cmd
}

func (self *CmdTree) RegDirWithCmd(cmd string, help string) *Cmd {
	self.cmdConflictCheck(help, "RegDirWithCmd")
	self.cmd = NewDirWithCmd(self, help, cmd)
	return self.cmd
}

func (self *CmdTree) RegEmptyDirCmd(dir string, help string) *Cmd {
	//dirnore empty dir cmd register
	if self.cmd != nil {
		return self.cmd
	}
	self.cmd = NewEmptyDirCmd(self, help, dir)
	return self.cmd
}

func (self *CmdTree) RegEmptyCmd(help string) *Cmd {
	self.cmdConflictCheck(help, "RegEmptyCmd")
	self.cmd = NewEmptyCmd(self, help)
	return self.cmd
}

func (self *CmdTree) RegFlowCmd(flow []string, help string) *Cmd {
	self.cmdConflictCheck(help, "RegFlowCmd")
	self.cmd = NewFlowCmd(self, help, flow)
	return self.cmd
}

func (self *CmdTree) RegPowerCmd(cmd PowerCmd, help string) *Cmd {
	self.cmdConflictCheck(help, "RegPowerCmd")
	self.cmd = NewPowerCmd(self, help, cmd)
	return self.cmd
}

func (self *CmdTree) RegAdHotFlowCmd(cmd AdHotFlowCmd, help string) *Cmd {
	self.cmdConflictCheck(help, "RegAdHotFlowCmd")
	self.cmd = NewAdHotFlowCmd(self, help, cmd)
	return self.cmd
}

func (self *CmdTree) RegFileNFlowCmd(flow []string, cmd string, help string) *Cmd {
	self.cmdConflictCheck(help, "RegPowerNFlow")
	self.cmd = NewFileNFlowCmd(self, help, cmd, flow)
	return self.cmd
}

func (self *CmdTree) AddSub(name string, abbrs ...string) *CmdTree {
	if old, ok := self.subs[name]; ok && old.name != name {
		err := &CmdTreeErrSubCmdConflicted{
			fmt.Sprintf("sub-cmd name conflicted: %s", name),
			self.Path(),
			name,
			old.Source(),
		}
		panic(err)
	}
	sub := NewCmdTree(self.Strs)
	sub.name = name
	sub.parent = self
	self.subs[name] = sub
	self.subOrderedNames = append(self.subOrderedNames, name)
	self.addSubAbbrs(name, abbrs...)
	self.subAbbrsRevIdx[name] = name
	return sub
}

func (self *CmdTree) AddAbbrs(abbrs ...string) {
	if self.parent == nil {
		panic(fmt.Errorf("[CmdTree.AddAbbrs] can't add abbrs %v to root", abbrs))
	}
	self.parent.addSubAbbrs(self.name, abbrs...)
}

func (self *CmdTree) GetOrAddSub(path ...string) *CmdTree {
	return self.getOrAddSub("", true, path...)
}

func (self *CmdTree) GetOrAddSubEx(source string, path ...string) *CmdTree {
	return self.getOrAddSub(source, true, path...)
}

func (self *CmdTree) HasSub() bool {
	return len(self.subs) != 0 && !self.IsHidden()
}

func (self *CmdTree) GetSub(path ...string) *CmdTree {
	return self.getOrAddSub("", false, path...)
}

func (self *CmdTree) GetSubByPath(path string, panicOnNotFound bool) *CmdTree {
	cmds := self.GetSub(strings.Split(path, self.Strs.PathSep)...)
	if cmds == nil && panicOnNotFound {
		panic(fmt.Errorf("can't find sub cmd tree by path '%s'", path))
	}
	return cmds
}

func (self *CmdTree) IsQuiet() bool {
	return self.cmd != nil && self.cmd.IsQuiet()
}

func (self *CmdTree) IsNoExecutableCmd() bool {
	if self.cmd == nil {
		return true
	}
	return self.cmd.IsNoExecutableCmd()
}

func (self *CmdTree) IsPowerCmd() bool {
	return self.cmd != nil && self.cmd.IsPowerCmd()
}

func (self *CmdTree) AllowTailModeCall() bool {
	return self.cmd != nil && self.cmd.AllowTailModeCall()
}

func (self *CmdTree) Parent() *CmdTree {
	return self.parent
}

func (self *CmdTree) IsRoot() bool {
	return self.parent == nil
}

func (self *CmdTree) Name() string {
	return self.name
}

func (self *CmdTree) DisplayName() string {
	if len(self.name) == 0 {
		return self.Strs.RootDisplayName
	}
	return self.name
}

func (self *CmdTree) Cmd() *Cmd {
	return self.cmd
}

func (self *CmdTree) Args() (args Args) {
	if self.cmd == nil {
		return
	}
	return self.cmd.Args()
}

func (self *CmdTree) Path() []string {
	if self.parent == nil {
		return nil
	}
	return append(self.parent.Path(), self.name)
}

func (self *CmdTree) AbbrsPath() []string {
	if self.parent == nil {
		return nil
	}
	abbrs := self.parent.SubAbbrs(self.name)
	if len(abbrs) == 0 {
		return nil
	}
	return append(self.parent.AbbrsPath(), strings.Join(abbrs, self.Strs.AbbrsSep))
}

func (self *CmdTree) Depth() int {
	if self.parent == nil {
		return 0
	} else {
		return self.parent.Depth() + 1
	}
}

func (self *CmdTree) MatchSource(source string) bool {
	if len(source) == 0 {
		return true
	}
	if len(self.source) == 0 {
		return strings.Index(self.Strs.BuiltinName, source) >= 0
	}
	return strings.Index(self.source, source) >= 0
}

func (self *CmdTree) MatchTags(tags ...string) bool {
	for _, tag := range tags {
		for _, it := range self.tags {
			//if strings.Index(it, tag) >= 0 {
			if it == tag {
				return true
			}
		}
	}
	return false
}

// TODO: unused
func (self *CmdTree) MatchExactTags(tags ...string) bool {
	tagSet := map[string]bool{}
	for _, tag := range self.tags {
		tagSet[tag] = true
	}
	for _, tag := range tags {
		if !tagSet[tag] {
			return false
		}
	}
	return true
}

func (self *CmdTree) MatchFind(findStrs ...string) bool {
	for _, str := range findStrs {
		if !self.matchFind(str) {
			return false
		}
	}
	return true
}

func (self *CmdTree) matchFind(findStr string) bool {
	if len(findStr) == 0 {
		return true
	}
	for _, tag := range self.tags {
		if strings.Index(self.Strs.TagMark+tag, findStr) >= 0 {
			return true
		}
	}
	if strings.Index(self.name, findStr) >= 0 {
		return true
	}
	if self.cmd != nil && self.cmd.MatchFind(findStr) {
		return true
	}
	if self.parent != nil {
		for _, abbr := range self.parent.SubAbbrs(self.name) {
			if strings.Index(abbr, findStr) >= 0 {
				return true
			}
		}
	}
	if len(self.source) == 0 {
		if strings.Index("builtin", findStr) >= 0 {
			return true
		}
	} else {
		if strings.Index(self.source, findStr) >= 0 {
			return true
		}
	}
	return false
}

func (self *CmdTree) DisplayPath() string {
	path := self.Path()
	if len(path) == 0 {
		return self.Strs.RootDisplayName
	} else {
		return strings.Join(path, self.Strs.PathSep)
	}
}

func (self *CmdTree) DisplayAbbrsPath() string {
	path := self.AbbrsPath()
	if len(path) == 0 {
		return ""
	} else {
		return strings.Join(path, self.Strs.PathSep)
	}
}

func (self *CmdTree) Realname(nameOrAbbr string) (realname string) {
	if self.parent == nil {
		return
	}
	realname, _ = self.parent.subAbbrsRevIdx[nameOrAbbr]
	return
}

func (self *CmdTree) SubNames() []string {
	return self.subOrderedNames
}

func (self *CmdTree) SubAbbrs(name string) (abbrs []string) {
	abbrs, _ = self.subAbbrs[name]
	return
}

func (self *CmdTree) Abbrs() (abbrs []string) {
	if self.parent == nil {
		return
	}
	return self.parent.SubAbbrs(self.name)
}

func (self *CmdTree) Source() string {
	return self.source
}

func (self *CmdTree) IsBuiltin() bool {
	return len(self.source) == 0
}

func (self *CmdTree) SetSource(s string) {
	self.source = s
}

func (self *CmdTree) addSubAbbrs(name string, abbrs ...string) {
	for _, abbr := range append([]string{name}, abbrs...) {
		if len(abbr) == 0 {
			continue
		}
		old, ok := self.subAbbrsRevIdx[abbr]
		if old == name {
			continue
		}
		if ok {
			err := &CmdTreeErrSubAbbrConflicted{
				fmt.Sprintf("%s: sub command abbr name '%s' conflicted, "+
					"old for '%s', new for '%s'",
					self.DisplayPath(), abbr, old, name),
				self.Path(),
				abbr,
				old,
				name,
				self.GetSub(old).Source(),
			}
			panic(err)
		}
		self.subAbbrsRevIdx[abbr] = name
		olds, _ := self.subAbbrs[name]
		self.subAbbrs[name] = append(olds, abbr)
	}
}

func (self *CmdTree) getOrAddSub(source string, addIfNotExists bool, path ...string) *CmdTree {
	if len(path) == 0 {
		return self
	}
	name := path[0]
	if realName, ok := self.subAbbrsRevIdx[name]; ok {
		name = realName
	}
	sub, ok := self.subs[name]
	if !ok {
		if !addIfNotExists {
			return nil
		}
		sub = self.AddSub(name)
		sub.source = source
	}
	return sub.getOrAddSub(source, addIfNotExists, path[1:]...)
}
