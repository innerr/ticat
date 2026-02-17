package model

import (
	"testing"
)

func TestNewArgs(t *testing.T) {
	args := newArgs()
	if args.names == nil {
		t.Error("newArgs should return initialized Args with names map")
	}

	if !args.IsEmpty() {
		t.Error("New Args should be empty")
	}
}

func TestArgsIsEmpty(t *testing.T) {
	args := newArgs()
	if !args.IsEmpty() {
		t.Error("New Args should be empty")
	}

	tree := NewCmdTree(CmdTreeStrsForTest())
	args.AddArg(tree, "arg1", "default")
	if args.IsEmpty() {
		t.Error("Args with added arg should not be empty")
	}
}

func TestArgsAddArg(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "arg1", "default1")
	args.AddArg(tree, "arg2", "default2", "a2", "alias2")

	if !args.Has("arg1") {
		t.Error("arg1 should exist")
	}

	if !args.Has("arg2") {
		t.Error("arg2 should exist")
	}

	if args.DefVal("arg1", 0) != "default1" {
		t.Errorf("Expected default1, got %s", args.DefVal("arg1", 0))
	}

	if args.DefVal("arg2", 0) != "default2" {
		t.Errorf("Expected default2, got %s", args.DefVal("arg2", 0))
	}
}

func TestArgsAddArgWithAbbrs(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "verbose", "false", "v", "V")

	if !args.HasArgOrAbbr("verbose") {
		t.Error("verbose should exist")
	}

	if !args.HasArgOrAbbr("v") {
		t.Error("v abbreviation should exist")
	}

	if !args.HasArgOrAbbr("V") {
		t.Error("V abbreviation should exist")
	}

	if args.Realname("v") != "verbose" {
		t.Errorf("Expected verbose, got %s", args.Realname("v"))
	}

	if args.Realname("V") != "verbose" {
		t.Errorf("Expected verbose, got %s", args.Realname("V"))
	}
}

func TestArgsAbbrs(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "output", "stdout", "o", "O", "out")

	abbrs := args.Abbrs("output")
	if len(abbrs) != 4 {
		t.Errorf("Expected 4 abbreviations, got %d", len(abbrs))
	}

	expected := map[string]bool{"output": true, "o": true, "O": true, "out": true}
	for _, abbr := range abbrs {
		if !expected[abbr] {
			t.Errorf("Unexpected abbreviation: %s", abbr)
		}
	}
}

func TestArgsRealname(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "count", "1", "c", "n")

	if args.Realname("count") != "count" {
		t.Errorf("Expected count, got %s", args.Realname("count"))
	}

	if args.Realname("c") != "count" {
		t.Errorf("Expected count for abbreviation c, got %s", args.Realname("c"))
	}

	if args.Realname("n") != "count" {
		t.Errorf("Expected count for abbreviation n, got %s", args.Realname("n"))
	}

	if args.Realname("nonexistent") != "" {
		t.Errorf("Expected empty string for nonexistent, got %s", args.Realname("nonexistent"))
	}
}

func TestArgsNames(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "arg1", "default1")
	args.AddArg(tree, "arg2", "default2")
	args.AddArg(tree, "arg3", "default3")

	names := args.Names()
	if len(names) != 3 {
		t.Errorf("Expected 3 names, got %d", len(names))
	}

	if names[0] != "arg1" || names[1] != "arg2" || names[2] != "arg3" {
		t.Errorf("Names order is incorrect: %v", names)
	}
}

func TestArgsIndex(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "first", "1")
	args.AddArg(tree, "second", "2")
	args.AddArg(tree, "third", "3")

	if args.Index("first") != 0 {
		t.Errorf("Expected index 0 for first, got %d", args.Index("first"))
	}

	if args.Index("second") != 1 {
		t.Errorf("Expected index 1 for second, got %d", args.Index("second"))
	}

	if args.Index("third") != 2 {
		t.Errorf("Expected index 2 for third, got %d", args.Index("third"))
	}

	if args.Index("nonexistent") != -1 {
		t.Errorf("Expected -1 for nonexistent, got %d", args.Index("nonexistent"))
	}
}

func TestArgsHas(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "existing", "value")

	if !args.Has("existing") {
		t.Error("Has should return true for existing arg")
	}

	if args.Has("nonexistent") {
		t.Error("Has should return false for nonexistent arg")
	}
}

func TestArgsHasArgOrAbbr(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "verbose", "false", "v")

	if !args.HasArgOrAbbr("verbose") {
		t.Error("HasArgOrAbbr should return true for name")
	}

	if !args.HasArgOrAbbr("v") {
		t.Error("HasArgOrAbbr should return true for abbreviation")
	}

	if args.HasArgOrAbbr("nonexistent") {
		t.Error("HasArgOrAbbr should return false for nonexistent")
	}
}

func TestArgsMatchFind(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "output", "stdout", "o")
	args.AddArg(tree, "verbose", "false", "v")

	if !args.MatchFind("output") {
		t.Error("MatchFind should find 'output'")
	}

	if !args.MatchFind("verb") {
		t.Error("MatchFind should find partial match 'verb'")
	}

	if !args.MatchFind("v") {
		t.Error("MatchFind should find abbreviation 'v'")
	}

	if args.MatchFind("nonexistent") {
		t.Error("MatchFind should not find 'nonexistent'")
	}
}

func TestArgsDefVal(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "count", "10")

	if args.DefVal("count", 0) != "10" {
		t.Errorf("Expected default value 10, got %s", args.DefVal("count", 0))
	}

	if args.DefVal("nonexistent", 0) != "" {
		t.Errorf("Expected empty string for nonexistent, got %s", args.DefVal("nonexistent", 0))
	}
}

func TestArgsRawDefVal(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "name", "default")

	if args.RawDefVal("name") != "default" {
		t.Errorf("Expected default, got %s", args.RawDefVal("name"))
	}
}

func TestArgsAddAutoMapAllArg(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddAutoMapAllArg(tree, "autoArg", "value")

	if !args.Has("autoArg") {
		t.Error("autoArg should exist")
	}

	if !args.IsFromAutoMapAll("autoArg") {
		t.Error("autoArg should be marked as from auto map all")
	}

	if args.IsFromAutoMapAll("nonexistent") {
		t.Error("nonexistent should not be marked as from auto map all")
	}
}

func TestArgsIsFromAutoMapAll(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "normalArg", "value")
	args.AddAutoMapAllArg(tree, "autoArg", "value")

	if args.IsFromAutoMapAll("normalArg") {
		t.Error("normalArg should not be from auto map all")
	}

	if !args.IsFromAutoMapAll("autoArg") {
		t.Error("autoArg should be from auto map all")
	}
}

func TestArgsDefValStackDepth(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddAutoMapAllArg(tree, "autoArg", "defaultValue")

	if args.DefVal("autoArg", 1) != "defaultValue" {
		t.Error("Stack depth 1 should return default value")
	}

	if args.DefVal("autoArg", 2) != "" {
		t.Error("Stack depth > 1 should return empty for auto map all arg")
	}
}

func TestArgsClone(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "arg1", "value1", "a1")
	args.AddArg(tree, "arg2", "value2", "a2")

	cloned := args.Clone()

	if !cloned.Has("arg1") {
		t.Error("Cloned args should have arg1")
	}

	if !cloned.HasArgOrAbbr("a1") {
		t.Error("Cloned args should have abbreviation a1")
	}

	if cloned.DefVal("arg1", 0) != "value1" {
		t.Error("Cloned args should have same default value")
	}
}

func TestArgsSetArgEnums(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "color", "red")
	args.SetArgEnums(tree, "color", "red", "green", "blue")

	enums := args.EnumVals("color")
	if len(enums) != 3 {
		t.Errorf("Expected 3 enum values, got %d", len(enums))
	}

	expected := map[string]bool{"red": true, "green": true, "blue": true}
	for _, v := range enums {
		if !expected[v] {
			t.Errorf("Unexpected enum value: %s", v)
		}
	}
}

func TestArgsEnumVals(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "color", "red")
	args.SetArgEnums(tree, "color", "red", "green", "blue")

	enums := args.EnumVals("color")
	if enums == nil {
		t.Error("EnumVals should return values")
	}

	if len(enums) != 3 {
		t.Errorf("Expected 3 values, got %d", len(enums))
	}

	enums = args.EnumVals("nonexistent")
	if enums != nil {
		t.Error("EnumVals for nonexistent should return nil")
	}
}

func TestArgsReorder(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "first", "1")
	args.AddArg(tree, "second", "2")
	args.AddArg(tree, "third", "3")

	cmd := NewEmptyCmd(tree, "")
	args.Reorder(cmd, []string{"third", "first", "second"})

	names := args.Names()
	if names[0] != "third" || names[1] != "first" || names[2] != "second" {
		t.Errorf("Reorder failed: %v", names)
	}

	if args.Index("third") != 0 {
		t.Errorf("Expected third at index 0, got %d", args.Index("third"))
	}
}

func TestArgsAddArgConflict(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "arg1", "default1")

	// Adding arg with same name should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("AddArg with conflicting name should panic")
		}
	}()

	args.AddArg(tree, "arg1", "default2")
}

func TestArgsAddArgEmptyAbbr(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	// Empty abbr should be skipped
	args.AddArg(tree, "verbose", "false", "", "v", "")

	if !args.Has("verbose") {
		t.Error("verbose should exist")
	}

	if !args.HasArgOrAbbr("v") {
		t.Error("v abbreviation should exist")
	}
}

func TestArgsAddArgAbbrConflict(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "verbose", "false", "v")

	// Adding arg with conflicting abbr should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("AddArg with conflicting abbr should panic")
		}
	}()

	args.AddArg(tree, "version", "1.0", "v")
}

func TestArgsSetArgEnumsNotFound(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	// Setting enums for non-existent arg should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("SetArgEnums for non-existent arg should panic")
		}
	}()

	args.SetArgEnums(tree, "nonexistent", "val1", "val2")
}

func TestArgsSetArgEnumsDuplicate(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "color", "red")
	args.SetArgEnums(tree, "color", "red", "green", "blue")

	// Setting enums twice should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("SetArgEnums twice should panic")
		}
	}()

	args.SetArgEnums(tree, "color", "yellow")
}

func TestArgsEnumValsEmpty(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "arg", "default")

	// No enums set, should return nil
	enums := args.EnumVals("arg")
	if enums != nil {
		t.Errorf("Expected nil for arg without enums, got %v", enums)
	}
}

func TestArgsDefValNonExistent(t *testing.T) {
	args := newArgs()

	val := args.DefVal("nonexistent", 0)
	if val != "" {
		t.Errorf("Expected empty string for nonexistent arg, got %q", val)
	}
}

func TestArgsRawDefValNonExistent(t *testing.T) {
	args := newArgs()

	val := args.RawDefVal("nonexistent")
	if val != "" {
		t.Errorf("Expected empty string for nonexistent arg, got %q", val)
	}
}

func TestArgsRealnameSelf(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "verbose", "false", "v")

	// Realname of full name should return itself
	if args.Realname("verbose") != "verbose" {
		t.Errorf("Expected 'verbose', got %s", args.Realname("verbose"))
	}
}

func TestArgsRealnameNonExistent(t *testing.T) {
	args := newArgs()

	// Realname of non-existent should return empty
	if args.Realname("nonexistent") != "" {
		t.Errorf("Expected empty string for nonexistent, got %s", args.Realname("nonexistent"))
	}
}

func TestArgsAbbrsNonExistent(t *testing.T) {
	args := newArgs()

	// Abbrs of non-existent should return nil
	abbrs := args.Abbrs("nonexistent")
	if abbrs != nil {
		t.Errorf("Expected nil for nonexistent arg, got %v", abbrs)
	}
}

func TestArgsIndexNonExistent(t *testing.T) {
	args := newArgs()

	args.AddArg(NewCmdTree(CmdTreeStrsForTest()), "arg1", "default")

	// Index of non-existent should return -1
	idx := args.Index("nonexistent")
	if idx != -1 {
		t.Errorf("Expected -1 for nonexistent arg, got %d", idx)
	}
}

func TestArgsCloneEmpty(t *testing.T) {
	args := newArgs()

	cloned := args.Clone()

	if !cloned.IsEmpty() {
		t.Error("Cloned empty args should be empty")
	}

	if cloned.names == nil {
		t.Error("Cloned args should have initialized names map")
	}
}

func TestArgsCloneIndependence(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "arg1", "value1", "a1")

	cloned := args.Clone()

	// Modify cloned args
	cloned.AddArg(tree, "arg2", "value2")

	// Original should not be affected
	if args.Has("arg2") {
		t.Error("Original args should not have arg2")
	}

	if cloned.Index("arg2") != 1 {
		t.Error("Cloned args should have arg2 at index 1")
	}
}

func TestArgsMatchFindEmpty(t *testing.T) {
	args := newArgs()

	// Empty args should not match anything
	if args.MatchFind("anything") {
		t.Error("Empty args should not match any string")
	}
}

func TestArgsReorderLengthMismatch(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "arg1", "default1")
	args.AddArg(tree, "arg2", "default2")

	cmd := NewEmptyCmd(tree, "")

	// Reorder with different length should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Reorder with length mismatch should panic")
		}
	}()

	args.Reorder(cmd, []string{"arg1"})
}

func TestArgsReorderInvalidName(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	args := newArgs()

	args.AddArg(tree, "arg1", "default1")
	args.AddArg(tree, "arg2", "default2")

	cmd := NewEmptyCmd(tree, "")

	// Reorder with invalid name should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Reorder with invalid name should panic")
		}
	}()

	args.Reorder(cmd, []string{"arg1", "nonexistent"})
}

func TestArgsIsFromAutoMapAllNonExistent(t *testing.T) {
	args := newArgs()

	// Non-existent arg should return false
	if args.IsFromAutoMapAll("nonexistent") {
		t.Error("IsFromAutoMapAll should return false for non-existent arg")
	}
}
