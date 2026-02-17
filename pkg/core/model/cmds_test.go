package model

import (
	"strings"
	"testing"
)

func TestNewCmdTree(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)

	if tree == nil {
		t.Fatal("NewCmdTree returned nil")
	}

	if tree.Strs != strs {
		t.Error("Strs should be set correctly")
	}

	if !tree.IsRoot() {
		t.Error("New tree should be root")
	}
}

func TestCmdTreeAddSub(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.AddSub("sub1")
	if sub == nil {
		t.Fatal("AddSub returned nil")
	}

	if sub.Name() != "sub1" {
		t.Errorf("Expected name 'sub1', got '%s'", sub.Name())
	}

	if sub.Parent() != tree {
		t.Error("Parent should be set correctly")
	}

	if !tree.HasSubs() {
		t.Error("Tree should have subs")
	}
}

func TestCmdTreeAddSubWithAbbrs(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.AddSub("verbose", "v", "V")

	if sub.Name() != "verbose" {
		t.Errorf("Expected name 'verbose', got '%s'", sub.Name())
	}

	abbrs := tree.SubAbbrs("verbose")
	if len(abbrs) != 3 {
		t.Errorf("Expected 3 abbreviations, got %d", len(abbrs))
	}

	subByAbbr := tree.GetSub("v")
	if subByAbbr == nil || subByAbbr.Name() != "verbose" {
		t.Errorf("Expected to find 'verbose' by abbreviation 'v'")
	}
}

func TestCmdTreeGetSub(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	tree.AddSub("level1").AddSub("level2").AddSub("level3")

	sub := tree.GetSub("level1", "level2", "level3")
	if sub == nil {
		t.Fatal("GetSub returned nil")
	}

	if sub.Name() != "level3" {
		t.Errorf("Expected name 'level3', got '%s'", sub.Name())
	}

	sub = tree.GetSub("nonexistent")
	if sub != nil {
		t.Error("GetSub for nonexistent path should return nil")
	}
}

func TestCmdTreeGetOrAddSub(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.GetOrAddSub("level1", "level2", "level3")
	if sub == nil {
		t.Fatal("GetOrAddSub returned nil")
	}

	if sub.Name() != "level3" {
		t.Errorf("Expected name 'level3', got '%s'", sub.Name())
	}

	sub2 := tree.GetOrAddSub("level1", "level2", "level3")
	if sub != sub2 {
		t.Error("GetOrAddSub should return same sub for same path")
	}
}

func TestCmdTreePath(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	level1 := tree.AddSub("level1")
	level2 := level1.AddSub("level2")
	level3 := level2.AddSub("level3")

	path := level3.Path()
	if len(path) != 3 {
		t.Errorf("Expected path length 3, got %d", len(path))
	}

	if path[0] != "level1" || path[1] != "level2" || path[2] != "level3" {
		t.Errorf("Path is incorrect: %v", path)
	}
}

func TestCmdTreeDisplayPath(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	level1 := tree.AddSub("level1")
	level2 := level1.AddSub("level2")

	displayPath := level2.DisplayPath()
	if displayPath != "level1.level2" {
		t.Errorf("Expected 'level1.level2', got '%s'", displayPath)
	}
}

func TestCmdTreeDepth(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	if tree.Depth() != 0 {
		t.Errorf("Root depth should be 0, got %d", tree.Depth())
	}

	level1 := tree.AddSub("level1")
	if level1.Depth() != 1 {
		t.Errorf("Level1 depth should be 1, got %d", level1.Depth())
	}

	level2 := level1.AddSub("level2")
	if level2.Depth() != 2 {
		t.Errorf("Level2 depth should be 2, got %d", level2.Depth())
	}
}

func TestCmdTreeRegCmd(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.GetOrAddSubEx("test-source", "testcmd")

	dummyFunc := func(argv ArgVals, cc *Cli, env *Env, flow []ParsedCmd) bool {
		return true
	}

	cmd := sub.RegCmd(dummyFunc, "test command", "test-source")
	if cmd == nil {
		t.Fatal("RegCmd returned nil")
	}

	if sub.Cmd() != cmd {
		t.Error("Cmd should be set correctly")
	}

	if cmd.Help() != "test command" {
		t.Errorf("Expected help 'test command', got '%s'", cmd.Help())
	}
}

func TestCmdTreeRegEmptyCmd(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.AddSub("emptycmd")
	cmd := sub.RegEmptyCmd("empty command")
	if cmd == nil {
		t.Fatal("RegEmptyCmd returned nil")
	}

	if cmd.Type() != CmdTypeEmpty {
		t.Errorf("Expected CmdTypeEmpty, got %s", cmd.Type())
	}
}

func TestCmdTreeRegPowerCmd(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.AddSub("powercmd")

	powerFunc := func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, bool) {
		return currCmdIdx, true
	}

	cmd := sub.RegPowerCmd(powerFunc, "power command")
	if cmd == nil {
		t.Fatal("RegPowerCmd returned nil")
	}

	if cmd.Type() != CmdTypePower {
		t.Errorf("Expected CmdTypePower, got %s", cmd.Type())
	}
}

func TestCmdTreeRegFlowCmd(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.GetOrAddSubEx("test-source", "flowcmd")

	flow := []string{"cmd1", ":", "cmd2"}
	cmd := sub.RegFlowCmd(flow, "flow command", "test-source")
	if cmd == nil {
		t.Fatal("RegFlowCmd returned nil")
	}

	if cmd.Type() != CmdTypeFlow {
		t.Errorf("Expected CmdTypeFlow, got %s", cmd.Type())
	}

	flowStrs := cmd.FlowStrs()
	if len(flowStrs) != 3 {
		t.Errorf("Expected flow length 3, got %d", len(flowStrs))
	}
}

func TestCmdTreeIsQuiet(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.AddSub("quietcmd")

	if sub.IsQuiet() {
		t.Error("Tree without cmd should not be quiet")
	}

	dummyFunc := func(argv ArgVals, cc *Cli, env *Env, flow []ParsedCmd) bool {
		return true
	}

	cmd := sub.RegCmd(dummyFunc, "test", "")
	cmd.SetQuiet()

	if !sub.IsQuiet() {
		t.Error("Tree with quiet cmd should be quiet")
	}
}

func TestCmdTreeIsPowerCmd(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.AddSub("powercmd")

	if sub.IsPowerCmd() {
		t.Error("Tree without cmd should not be power cmd")
	}

	powerFunc := func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, bool) {
		return currCmdIdx, true
	}

	sub.RegPowerCmd(powerFunc, "power")

	if !sub.IsPowerCmd() {
		t.Error("Tree with power cmd should be power cmd")
	}
}

func TestCmdTreeIsEmpty(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	if !tree.IsEmpty() {
		t.Error("Tree without cmd should be empty")
	}

	tree.RegEmptyCmd("empty")
	if !tree.IsEmpty() {
		t.Error("Tree with empty cmd should be empty")
	}
}

func TestCmdTreeIsNoExecutableCmd(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	if !tree.IsNoExecutableCmd() {
		t.Error("Tree without cmd should be no executable")
	}

	tree.RegEmptyCmd("empty")
	if !tree.IsNoExecutableCmd() {
		t.Error("Tree with empty cmd should be no executable")
	}
}

func TestCmdTreeSetHidden(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.AddSub("sub")
	sub.SetHidden()

	if !sub.IsHidden() {
		t.Error("Sub should be hidden")
	}
}

func TestCmdTreeSetTrivial(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.AddSub("sub")
	sub.SetTrivial(2)

	if sub.Trivial() != 2 {
		t.Errorf("Expected trivial level 2, got %d", sub.Trivial())
	}
}

func TestCmdTreeAddTags(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	tree.AddTags("@ready", "@test")

	tags := tree.Tags()
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}
}

func TestCmdTreeMatchTags(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	tree.AddTags("@ready", "@test")

	if !tree.MatchTags("@ready") {
		t.Error("Should match @ready tag")
	}

	if !tree.MatchTags("@test") {
		t.Error("Should match @test tag")
	}

	if tree.MatchTags("@nonexistent") {
		t.Error("Should not match nonexistent tag")
	}
}

func TestCmdTreeMatchFind(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	tree.AddSub("benchmark")
	tree.AddSub("test").AddTags("@ready")

	if !tree.GetSub("benchmark").MatchFind("bench") {
		t.Error("Should find 'bench' in 'benchmark'")
	}

	if !tree.GetSub("test").MatchFind("@ready") {
		t.Error("Should find '@ready' tag")
	}
}

func TestCmdTreeGatherSubNames(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	tree.AddSub("cmd1")
	tree.AddSub("cmd2", "c2")

	names := tree.GatherSubNames(false, true)
	if len(names) != 2 {
		t.Errorf("Expected 2 names (no abbrs), got %d", len(names))
	}

	names = tree.GatherSubNames(true, true)
	if len(names) != 3 {
		t.Errorf("Expected 3 names (with abbrs), got %d", len(names))
	}
}

func TestCmdTreeSubNames(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	tree.AddSub("cmd1")
	tree.AddSub("cmd2")
	tree.AddSub("cmd3")

	names := tree.SubNames()
	if len(names) != 3 {
		t.Errorf("Expected 3 names, got %d", len(names))
	}

	if names[0] != "cmd1" || names[1] != "cmd2" || names[2] != "cmd3" {
		t.Errorf("Names order is incorrect: %v", names)
	}
}

func TestCmdTreeClone(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	tree.AddSub("sub1").AddSub("sub2")

	cloned := tree.Clone()
	if cloned == nil {
		t.Fatal("Clone returned nil")
	}

	if cloned.GetSub("sub1") == nil {
		t.Error("Cloned tree should have sub1")
	}

	if cloned.GetSub("sub1", "sub2") == nil {
		t.Error("Cloned tree should have sub1.sub2")
	}
}

func TestCmdTreeSetSource(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	tree.SetSource("test-source")

	if tree.Source() != "test-source" {
		t.Errorf("Expected source 'test-source', got '%s'", tree.Source())
	}
}

func TestCmdTreeIsBuiltin(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	if !tree.IsBuiltin() {
		t.Error("Tree without source should be builtin")
	}

	tree.SetSource("external")
	if tree.IsBuiltin() {
		t.Error("Tree with source should not be builtin")
	}
}

func TestCmdTreeDisplayName(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	if tree.DisplayName() != "<root>" {
		t.Errorf("Expected '<root>', got '%s'", tree.DisplayName())
	}

	sub := tree.AddSub("mysub")
	if sub.DisplayName() != "mysub" {
		t.Errorf("Expected 'mysub', got '%s'", sub.DisplayName())
	}
}

func TestCmdTreeParent(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.AddSub("sub")
	if sub.Parent() != tree {
		t.Error("Parent should be set correctly")
	}

	if tree.Parent() != nil {
		t.Error("Root parent should be nil")
	}
}

func TestCmdTreeIsRoot(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	if !tree.IsRoot() {
		t.Error("Tree should be root")
	}

	sub := tree.AddSub("sub")
	if sub.IsRoot() {
		t.Error("Sub should not be root")
	}
}

func TestCmdTreeSetIsApi(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	if tree.IsApi() {
		t.Error("Tree should not be API by default")
	}

	tree.SetIsApi()
	if !tree.IsApi() {
		t.Error("Tree should be API after SetIsApi")
	}
}

func TestCmdTreeGetSubByPath(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	tree.AddSub("level1").AddSub("level2")

	sub := tree.GetSubByPath("level1.level2", false)
	if sub == nil {
		t.Error("Should find sub by path")
	}

	if sub.Name() != "level2" {
		t.Errorf("Expected 'level2', got '%s'", sub.Name())
	}
}

func TestCmdTreeMatchSource(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	if !tree.MatchSource("") {
		t.Error("Should match empty source")
	}

	tree.SetSource("github.com/user/repo")
	if !tree.MatchSource("user/repo") {
		t.Error("Should match partial source")
	}

	if tree.MatchSource("other/repo") {
		t.Error("Should not match different source")
	}
}

func TestCmdTreeHasSubs(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	if tree.HasSubs() {
		t.Error("Tree without subs should not have subs")
	}

	sub := tree.AddSub("sub")
	if !tree.HasSubs() {
		t.Error("Tree with subs should have subs")
	}

	if sub.HasSubs() {
		t.Error("Leaf node should not have subs")
	}
}

func TestCmdTreeArgs(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.AddSub("cmdwithargs")

	args := sub.Args()
	if !args.IsEmpty() {
		t.Error("Tree without cmd should have empty args")
	}

	dummyFunc := func(argv ArgVals, cc *Cli, env *Env, flow []ParsedCmd) bool {
		return true
	}

	cmd := sub.RegCmd(dummyFunc, "test", "")
	cmd.AddArg("arg1", "default")

	args = sub.Args()
	if args.IsEmpty() {
		t.Error("Args should not be empty after adding arg")
	}

	if !args.Has("arg1") {
		t.Error("Args should have arg1")
	}
}

func TestCmdTreeRealname(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	verbose := tree.AddSub("verbose", "v")

	realname := verbose.Realname("v")
	if realname != "verbose" {
		t.Errorf("Expected 'verbose' for abbreviation 'v', got '%s'", realname)
	}

	realname = verbose.Realname("verbose")
	if realname != "verbose" {
		t.Errorf("Expected 'verbose' for full name, got '%s'", realname)
	}

	realname = verbose.Realname("nonexistent")
	if realname != "" {
		t.Errorf("Expected empty string for nonexistent, got '%s'", realname)
	}
}

func TestCmdTreeAbbrs(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.AddSub("verbose", "v", "V")
	abbrs := sub.Abbrs()

	if len(abbrs) != 3 {
		t.Errorf("Expected 3 abbreviations, got %d", len(abbrs))
	}

	hasV := false
	hasVUpper := false
	hasVerbose := false
	for _, abbr := range abbrs {
		if abbr == "v" {
			hasV = true
		}
		if abbr == "V" {
			hasVUpper = true
		}
		if abbr == "verbose" {
			hasVerbose = true
		}
	}

	if !hasV || !hasVUpper || !hasVerbose {
		t.Error("Missing expected abbreviations")
	}
}

func TestIsShortcutCmdName(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"cmd", false},
		{"Cmd", false},
		{"CMD", false},
		{"+cmd", true},
		{"-cmd", true},
		{"@cmd", true},
		{":cmd", true},
		{"", false},
	}

	for _, test := range tests {
		result := isShortcutCmdName(test.name)
		if result != test.expected {
			t.Errorf("isShortcutCmdName(%s) = %v, expected %v", test.name, result, test.expected)
		}
	}
}

func TestCmdTreeString(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	path := tree.Path()
	pathStr := strings.Join(path, ".")
	if pathStr != "" {
		t.Errorf("Root path should be empty, got '%s'", pathStr)
	}

	sub := tree.AddSub("sub1")
	path = sub.Path()
	pathStr = strings.Join(path, ".")
	if pathStr != "sub1" {
		t.Errorf("Expected 'sub1', got '%s'", pathStr)
	}
}

func TestCmdTreeAddSubConflict(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	tree.AddSub("sub1")

	// Adding sub with same name should NOT panic (it's allowed for in-source overwrite)
	// The conflict check is: if old, ok := self.subs[name]; ok && old.name != name
	// Since we're adding the same name, old.name == name, so no panic
	sub := tree.AddSub("sub1")
	if sub == nil {
		t.Error("AddSub with same name should return the existing sub")
	}
}

func TestCmdTreeAddSubAbbrConflict(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	tree.AddSub("verbose", "v")

	// Adding sub with conflicting abbr should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("AddSub with conflicting abbr should panic")
		}
	}()

	tree.AddSub("version", "v")
}

func TestCmdTreeAddAbbrsToRoot(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	// Adding abbrs to root should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("AddAbbrs to root should panic")
		}
	}()

	tree.AddAbbrs("r", "R")
}

func TestCmdTreeGetSubByPathPanic(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	tree.AddSub("level1")

	// GetSubByPath with panicOnNotFound should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("GetSubByPath with non-existent path should panic")
		}
	}()

	tree.GetSubByPath("nonexistent", true)
}

func TestCmdTreeGetSubByPathNoPanic(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	tree.AddSub("level1")

	// GetSubByPath with panicOnNotFound=false should return nil
	result := tree.GetSubByPath("nonexistent", false)
	if result != nil {
		t.Error("GetSubByPath should return nil for non-existent path")
	}
}

func TestCmdTreeMatchExactTags(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	tree.AddTags("@ready", "@test", "@v1")

	// Should match all exact tags
	if !tree.MatchExactTags("@ready", "@test") {
		t.Error("Should match exact tags")
	}

	// Should not match partial
	if tree.MatchExactTags("@ready", "@nonexistent") {
		t.Error("Should not match non-existent tag")
	}
}

func TestCmdTreeMatchFindEmpty(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	// Empty find string should match
	if !tree.matchFind("") {
		t.Error("Empty find string should match")
	}
}

func TestCmdTreeMatchFindBuiltin(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.AddSub("mysub")

	// Should find "builtin" in source
	if !sub.matchFind("builtin") {
		t.Error("Should find 'builtin' in source")
	}
}

func TestCmdTreeDisplayPathRoot(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	displayPath := tree.DisplayPath()
	if displayPath != tree.Strs.RootDisplayName {
		t.Errorf("Expected root display name '%s', got '%s'", tree.Strs.RootDisplayName, displayPath)
	}
}

func TestCmdTreeRealnameRoot(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	// Root has no parent, Realname should return empty
	realname := tree.Realname("anything")
	if realname != "" {
		t.Errorf("Expected empty string for root Realname, got %s", realname)
	}
}

func TestCmdTreeAbbrsRoot(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	// Root has no parent, Abbrs should return nil
	abbrs := tree.Abbrs()
	if abbrs != nil {
		t.Errorf("Expected nil for root Abbrs, got %v", abbrs)
	}
}

func TestCmdTreeCloneIndependence(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	tree.AddSub("sub1")
	tree.AddTags("@tag1")

	cloned := tree.Clone()

	// Modify cloned tree
	cloned.AddSub("sub2")
	cloned.AddTags("@tag2")

	// Original should not be affected
	if cloned.GetSub("sub2") == nil {
		t.Error("Cloned tree should have sub2")
	}

	if tree.GetSub("sub2") != nil {
		t.Error("Original tree should not have sub2")
	}

	if len(cloned.Tags()) != 2 {
		t.Errorf("Cloned tree should have 2 tags, got %d", len(cloned.Tags()))
	}

	if len(tree.Tags()) != 1 {
		t.Errorf("Original tree should have 1 tag, got %d", len(tree.Tags()))
	}
}

func TestCmdTreeGetOrAddSubEx(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.GetOrAddSubEx("external-source", "level1", "level2")

	if sub.Name() != "level2" {
		t.Errorf("Expected name 'level2', got '%s'", sub.Name())
	}

	if sub.Source() != "external-source" {
		t.Errorf("Expected source 'external-source', got '%s'", sub.Source())
	}
}

func TestCmdTreeMatchSourceBuiltin(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	// Builtin tree should match "builtin"
	if !tree.MatchSource("builtin") {
		t.Error("Builtin tree should match 'builtin'")
	}
}

func TestCmdTreeMatchFindWithSource(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	tree.SetSource("github.com/user/repo")

	if !tree.matchFind("github") {
		t.Error("Should find 'github' in source")
	}

	if !tree.matchFind("user/repo") {
		t.Error("Should find 'user/repo' in source")
	}
}

func TestCmdTreeMatchFindWithTags(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	tree.AddTags("@important")

	if !tree.matchFind("@important") {
		t.Error("Should find tag '@important'")
	}

	if !tree.matchFind("important") {
		t.Error("Should find tag content 'important'")
	}
}

func TestCmdTreeMatchFindWithAbbr(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	tree.AddSub("verbose", "v", "V")

	sub := tree.GetSub("verbose")

	if !sub.matchFind("v") {
		t.Error("Should find abbreviation 'v'")
	}
}

func TestCmdTreeGetSubEmptyPath(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	// Empty path should return self
	result := tree.GetSub()
	if result != tree {
		t.Error("GetSub with empty path should return self")
	}
}

func TestCmdTreeGetOrAddSubEmptyPath(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	// Empty path should return self
	result := tree.GetOrAddSub()
	if result != tree {
		t.Error("GetOrAddSub with empty path should return self")
	}
}

func TestCmdTreeDepthDeep(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	level1 := tree.AddSub("level1")
	level2 := level1.AddSub("level2")
	level3 := level2.AddSub("level3")
	level4 := level3.AddSub("level4")

	if level4.Depth() != 4 {
		t.Errorf("Expected depth 4, got %d", level4.Depth())
	}
}

func TestCmdTreeSubAbbrsNonExistent(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	// Non-existent sub should return nil abbrs
	abbrs := tree.SubAbbrs("nonexistent")
	if abbrs != nil {
		t.Errorf("Expected nil for non-existent sub abbrs, got %v", abbrs)
	}
}

func TestCmdTreeAddSubWithMultipleAbbrs(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	sub := tree.AddSub("output", "o", "O", "out", "OUT")

	abbrs := tree.SubAbbrs("output")
	if len(abbrs) != 5 {
		t.Errorf("Expected 5 abbreviations, got %d", len(abbrs))
	}

	// All abbrs should resolve to the sub
	expectedAbbrs := []string{"output", "o", "O", "out", "OUT"}
	for _, abbr := range expectedAbbrs {
		result := tree.GetSub(abbr)
		if result == nil || result != sub {
			t.Errorf("Abbreviation '%s' should resolve to the same sub", abbr)
		}
	}
}

func TestCmdTreeAddSubEmptyAbbr(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())

	// Empty abbrs should be skipped
	tree.AddSub("verbose", "", "v", "", "V")

	abbrs := tree.SubAbbrs("verbose")
	// Should have: verbose, v, V (empty strings skipped)
	if len(abbrs) != 3 {
		t.Errorf("Expected 3 abbreviations (empty skipped), got %d: %v", len(abbrs), abbrs)
	}
}
