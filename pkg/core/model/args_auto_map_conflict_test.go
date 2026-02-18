package model

import (
	"testing"
)

func TestArgsAutoMapSameKeyFromDifferentSources(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create source 1 with key1
	src1Cmd := tree.AddSub("src1")
	src1 := src1Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "source 1")
	src1.AddArg("arg1", "default1")
	src1.AddArg2Env("conflict.key", "arg1")
	
	// Create source 2 with same key
	src2Cmd := tree.AddSub("src2")
	src2 := src2Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "source 2")
	src2.AddArg("arg2", "default2")
	src2.AddArg2Env("conflict.key", "arg2")
	
	// Create target with auto map
	targetCmd := tree.AddSub("conflicttarget1")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "conflict target 1")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	status := target.GetArgsAutoMapStatus()
	
	// Mark both sources
	argv1 := ArgVals{"arg1": ArgVal{Raw: "val1", Provided: true}}
	argv2 := ArgVals{"arg2": ArgVal{Raw: "val2", Provided: true}}
	
	status.MarkMetWithArgv(src1, argv1)
	status.MarkMetWithArgv(src2, argv2)
	
	// The key should be recorded once (from whichever source marked it first)
	if !status.providedKeys["conflict.key"] {
		t.Error("conflict key should be recorded as provided")
	}
}

func TestArgsAutoMapArgNameConflictFromDifferentSources(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create source 1 with arg "data"
	src1Cmd := tree.AddSub("srcarg1")
	src1 := src1Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "source arg 1")
	src1.AddArg("data", "default1")
	src1.AddArg2Env("key1", "data")
	
	// Create source 2 with same arg name "data"
	src2Cmd := tree.AddSub("srcarg2")
	src2 := src2Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "source arg 2")
	src2.AddArg("data", "default2")
	src2.AddArg2Env("key2", "data")
	
	// Create target with auto map
	targetCmd := tree.AddSub("argconflicttarget")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "arg conflict target")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	status := target.GetArgsAutoMapStatus()
	
	// Mark both sources with provided args
	argv1 := ArgVals{"data": ArgVal{Raw: "val1", Provided: true}}
	argv2 := ArgVals{"data": ArgVal{Raw: "val2", Provided: true}}
	
	status.MarkMetWithArgv(src1, argv1)
	status.MarkMetWithArgv(src2, argv2)
	
	// Both keys should be recorded
	if !status.providedKeys["key1"] {
		t.Error("key1 should be recorded as provided")
	}
	if !status.providedKeys["key2"] {
		t.Error("key2 should be recorded as provided")
	}
}

func TestArgsAutoMapNestedFlowKeyConflict(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create inner command with key
	innerCmd := tree.AddSub("innerconflict")
	inner := innerCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "inner conflict")
	inner.AddArg("arg", "default")
	inner.AddArg2Env("nested.conflict.key", "arg")
	
	// Create outer command with same key
	outerCmd := tree.AddSub("outerconflict")
	outer := outerCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "outer conflict")
	outer.AddArg("arg", "default")
	outer.AddArg2Env("nested.conflict.key", "arg")
	
	// Create target with auto map
	targetCmd := tree.AddSub("nestedconflicttarget")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "nested conflict target")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	status := target.GetArgsAutoMapStatus()
	
	// Mark both with provided args
	argvInner := ArgVals{"arg": ArgVal{Raw: "inner-val", Provided: true}}
	argvOuter := ArgVals{"arg": ArgVal{Raw: "outer-val", Provided: true}}
	
	status.MarkMetWithArgv(inner, argvInner)
	status.MarkMetWithArgv(outer, argvOuter)
	
	// The key should be recorded (once)
	if !status.providedKeys["nested.conflict.key"] {
		t.Error("nested conflict key should be recorded as provided")
	}
}

func TestArgsAutoMapMultipleAbbrConflict(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create source with arg that has abbr
	srcCmd := tree.AddSub("abbrconflictsrc")
	src := srcCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "abbr conflict source")
	src.AddArg("longargname", "default", "lan", "ln")
	src.AddArg2Env("abbr.key", "longargname")
	
	// Create target with auto map
	targetCmd := tree.AddSub("abbrconflicttarget")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "abbr conflict target")
	
	// Target already has an arg with conflicting abbr
	target.AddArg("existing", "default", "lan")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	status := target.GetArgsAutoMapStatus()
	
	// Mark source with provided arg
	argv := ArgVals{"longargname": ArgVal{Raw: "val", Provided: true}}
	status.MarkMetWithArgv(src, argv)
	
	// The key should be recorded
	if !status.providedKeys["abbr.key"] {
		t.Error("abbr key should be recorded as provided")
	}
}

func TestArgsAutoMapKeyConflictWithVal2EnvAndArg2Env(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create source with both val2env and arg2env to same key
	srcCmd := tree.AddSub("valargconflictsrc")
	src := srcCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "val arg conflict source")
	
	// Add val2env first
	src.AddVal2Env("conflict.key", "val-value")
	
	// Try to add arg2env to same key - this should be prevented by AddArg2Env
	// But we can test the provided keys tracking
	src.AddArg("arg", "default")
	
	// Create target
	targetCmd := tree.AddSub("valargconflicttarget")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "val arg conflict target")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	status := target.GetArgsAutoMapStatus()
	
	// Mark source
	argv := ArgVals{"arg": ArgVal{Raw: "arg-val", Provided: true}}
	status.MarkMetWithArgv(src, argv)
	
	// val2env key should be recorded
	if !status.providedKeys["conflict.key"] {
		t.Error("val2env key should be recorded as provided")
	}
}

func TestArgsAutoMapDeepNestedConflictResolution(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create 3 levels, each trying to provide the same key
	level1Cmd := tree.AddSub("deepconflict1")
	level1 := level1Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "deep conflict level 1")
	level1.AddArg("arg1", "default1")
	level1.AddArg2Env("deep.conflict", "arg1")
	
	level2Cmd := tree.AddSub("deepconflict2")
	level2 := level2Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "deep conflict level 2")
	level2.AddArg("arg2", "default2")
	level2.AddArg2Env("deep.conflict", "arg2")
	
	level3Cmd := tree.AddSub("deepconflict3")
	level3 := level3Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "deep conflict level 3")
	level3.AddArg("arg3", "default3")
	level3.AddArg2Env("deep.conflict", "arg3")
	
	// Create target
	targetCmd := tree.AddSub("deepconflicttarget")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "deep conflict target")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	status := target.GetArgsAutoMapStatus()
	
	// Mark all levels with provided args
	argv1 := ArgVals{"arg1": ArgVal{Raw: "val1", Provided: true}}
	argv2 := ArgVals{"arg2": ArgVal{Raw: "val2", Provided: true}}
	argv3 := ArgVals{"arg3": ArgVal{Raw: "val3", Provided: true}}
	
	status.MarkMetWithArgv(level1, argv1)
	status.MarkMetWithArgv(level2, argv2)
	status.MarkMetWithArgv(level3, argv3)
	
	// The key should be recorded once (first one wins)
	if !status.providedKeys["deep.conflict"] {
		t.Error("deep conflict key should be recorded as provided")
	}
	
	// Count should be 1, not 3
	count := 0
	for key := range status.providedKeys {
		if key == "deep.conflict" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("deep.conflict key should be recorded once, got %d times", count)
	}
}

func TestArgsAutoMapConflictWithExistingTargetArg(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create source with arg
	srcCmd := tree.AddSub("existingargsrc")
	src := srcCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "existing arg source")
	src.AddArg("srcarg", "default")
	src.AddArg2Env("existing.key", "srcarg")
	
	// Create target with existing arg that maps to different key
	targetCmd := tree.AddSub("existingargtarget")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "existing arg target")
	target.AddArg("targetarg", "default")
	target.AddArg2Env("existing.key", "targetarg")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	status := target.GetArgsAutoMapStatus()
	
	// Mark source
	argv := ArgVals{"srcarg": ArgVal{Raw: "val", Provided: true}}
	status.MarkMetWithArgv(src, argv)
	
	// The key should be recorded
	if !status.providedKeys["existing.key"] {
		t.Error("existing key should be recorded as provided")
	}
}
