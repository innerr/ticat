package model

import (
	"testing"
)

func TestArgsAutoMapNestedFlowWithProvidedKeysAtDifferentLevels(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create level 3 command with val2env
	level3Cmd := tree.AddSub("deeplevel")
	level3 := level3Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "deep level command")
	level3.AddVal2Env("deep.provided.key", "deep-value")
	level3.AddArg("deeparg", "deep-default")
	level3.AddArg2Env("deep.arg.key", "deeparg")
	
	// Create level 2 command with arg2env
	level2Cmd := tree.AddSub("midlevel")
	level2 := level2Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "mid level command")
	level2.AddArg("midarg", "mid-default")
	level2.AddArg2Env("mid.arg.key", "midarg")
	
	// Create level 1 command
	level1Cmd := tree.AddSub("toplevel")
	level1 := level1Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "top level command")
	level1.AddArg("toparg", "top-default")
	level1.AddArg2Env("top.arg.key", "toparg")
	
	// Create target command with smart mapping
	targetCmd := tree.AddSub("nestedtarget")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "nested target command")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	status := target.GetArgsAutoMapStatus()
	
	// Mark all levels with different argv states
	argvDeep := ArgVals{
		"deeparg": ArgVal{Raw: "provided-deep", Provided: true},
	}
	argvMid := ArgVals{
		"midarg": ArgVal{Raw: "mid-default", Provided: false},
	}
	argvTop := ArgVals{
		"toparg": ArgVal{Raw: "provided-top", Provided: true},
	}
	
	status.MarkMetWithArgv(level3, argvDeep)
	status.MarkMetWithArgv(level2, argvMid)
	status.MarkMetWithArgv(level1, argvTop)
	
	// Verify all provided keys are recorded
	if !status.providedKeys["deep.provided.key"] {
		t.Error("deep level val2env key should be recorded")
	}
	if !status.providedKeys["deep.arg.key"] {
		t.Error("deep level arg2env key with provided arg should be recorded")
	}
	if !status.providedKeys["top.arg.key"] {
		t.Error("top level arg2env key with provided arg should be recorded")
	}
	if status.providedKeys["mid.arg.key"] {
		t.Error("mid level arg2env key with non-provided arg should NOT be recorded")
	}
}

func TestArgsAutoMapFiveLevelNestedFlow(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create 5 levels of nested commands
	for i := 5; i >= 1; i-- {
		cmdName := "nested5l" + string(rune('0'+i))
		cmdTree := tree.AddSub(cmdName)
		
		if i == 5 {
			// Deepest level - power cmd with arg
			cmd := cmdTree.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
				return currCmdIdx, nil
			}, cmdName+" command")
			cmd.AddArg("deeparg", "deepval")
			cmd.AddArg2Env("nested.deep.key", "deeparg")
		} else {
			// Upper levels - flow cmds
			prevPath := "nested5l" + string(rune('0'+i+1))
			cmdTree.RegFlowCmd([]string{prevPath}, cmdName+" command", "")
		}
	}
	
	// Create target command
	targetCmd := tree.AddSub("nested5target")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "5 level nested target")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	// Verify the structure is correct
	status := target.GetArgsAutoMapStatus()
	if status.mapNoProvider != true {
		t.Error("smart mapping should be enabled")
	}
}

func TestArgsAutoMapNestedFlowWithGlobalEnvAtDifferentLevels(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create inner command
	innerCmd := tree.AddSub("innercmd")
	inner := innerCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "inner command")
	inner.AddArg("innerarg", "inner-default")
	inner.AddArg2Env("inner.key", "innerarg")
	
	// Create outer command
	outerCmd := tree.AddSub("outercmd")
	outer := outerCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "outer command")
	outer.AddArg("outerarg", "outer-default")
	outer.AddArg2Env("outer.key", "outerarg")
	
	// Create target command
	targetCmd := tree.AddSub("nestedglobaltarget")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "nested global env target")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	status := target.GetArgsAutoMapStatus()
	
	// Record global env keys at different levels
	status.RecordGlobalEnvKey("global.level1.key")
	status.RecordGlobalEnvKey("global.level2.key")
	
	// Mark inner and outer with provided args
	innerArgv := ArgVals{
		"innerarg": ArgVal{Raw: "inner-provided", Provided: true},
	}
	outerArgv := ArgVals{
		"outerarg": ArgVal{Raw: "outer-provided", Provided: true},
	}
	
	status.MarkMetWithArgv(inner, innerArgv)
	status.MarkMetWithArgv(outer, outerArgv)
	
	// Verify all keys are recorded
	expectedKeys := []string{
		"global.level1.key",
		"global.level2.key",
		"inner.key",
		"outer.key",
	}
	
	for _, key := range expectedKeys {
		if !status.providedKeys[key] {
			t.Errorf("key %s should be recorded as provided", key)
		}
	}
}

func TestArgsAutoMapNestedFlowWithEnvOpsAtDifferentLevels(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create commands at different levels with env ops
	level1Cmd := tree.AddSub("envopsl1")
	level1 := level1Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "env ops level 1")
	level1.AddEnvOp("envop.l1.key1", EnvOpTypeWrite)
	level1.AddEnvOp("envop.l1.key2", EnvOpTypeWrite)
	
	level2Cmd := tree.AddSub("envopsl2")
	level2 := level2Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "env ops level 2")
	level2.AddEnvOp("envop.l2.key1", EnvOpTypeWrite)
	
	level3Cmd := tree.AddSub("envopsl3")
	level3 := level3Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "env ops level 3")
	level3.AddEnvOp("envop.l3.key1", EnvOpTypeWrite)
	
	// Create target command
	targetCmd := tree.AddSub("envopstarget")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "env ops target")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	status := target.GetArgsAutoMapStatus()
	
	// Mark all levels
	status.MarkMet(level1)
	status.MarkMet(level2)
	status.MarkMet(level3)
	
	// Verify all env op keys are recorded
	expectedKeys := []string{
		"envop.l1.key1",
		"envop.l1.key2",
		"envop.l2.key1",
		"envop.l3.key1",
	}
	
	for _, key := range expectedKeys {
		if !status.providedKeys[key] {
			t.Errorf("env op key %s should be recorded as provided", key)
		}
	}
}

func TestArgsAutoMapDeeplyNestedWithMixedProvidedKeys(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create 4 levels with mixed provided key types
	// Level 4: val2env + arg2env (provided)
	level4Cmd := tree.AddSub("mixl4")
	level4 := level4Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "mixed level 4")
	level4.AddVal2Env("mix.val.key", "val")
	level4.AddArg("arg4", "default4")
	level4.AddArg2Env("mix.arg4.key", "arg4")
	
	// Level 3: env op + arg2env (not provided)
	level3Cmd := tree.AddSub("mixl3")
	level3 := level3Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "mixed level 3")
	level3.AddEnvOp("mix.envop.key", EnvOpTypeWrite)
	level3.AddArg("arg3", "default3")
	level3.AddArg2Env("mix.arg3.key", "arg3")
	
	// Level 2: only arg2env (provided)
	level2Cmd := tree.AddSub("mixl2")
	level2 := level2Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "mixed level 2")
	level2.AddArg("arg2", "default2")
	level2.AddArg2Env("mix.arg2.key", "arg2")
	
	// Level 1: arg2env (not provided)
	level1Cmd := tree.AddSub("mixl1")
	level1 := level1Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "mixed level 1")
	level1.AddArg("arg1", "default1")
	level1.AddArg2Env("mix.arg1.key", "arg1")
	
	// Create target
	targetCmd := tree.AddSub("mixedtarget")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "mixed target")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	status := target.GetArgsAutoMapStatus()
	
	// Mark levels with different argv states
	argv4 := ArgVals{"arg4": ArgVal{Raw: "provided4", Provided: true}}
	argv3 := ArgVals{"arg3": ArgVal{Raw: "default3", Provided: false}}
	argv2 := ArgVals{"arg2": ArgVal{Raw: "provided2", Provided: true}}
	argv1 := ArgVals{"arg1": ArgVal{Raw: "default1", Provided: false}}
	
	status.MarkMetWithArgv(level4, argv4)
	status.MarkMetWithArgv(level3, argv3)
	status.MarkMetWithArgv(level2, argv2)
	status.MarkMetWithArgv(level1, argv1)
	status.RecordGlobalEnvKey("mix.global.key")
	
	// Verify expected provided keys
	shouldExist := []string{
		"mix.val.key",       // val2env always provided
		"mix.arg4.key",      // arg2env with provided arg
		"mix.envop.key",     // env op always provided
		"mix.arg2.key",      // arg2env with provided arg
		"mix.global.key",    // global env
	}
	
	shouldNotExist := []string{
		"mix.arg3.key",      // arg2env with non-provided arg
		"mix.arg1.key",      // arg2env with non-provided arg
	}
	
	for _, key := range shouldExist {
		if !status.providedKeys[key] {
			t.Errorf("key %s should be recorded as provided", key)
		}
	}
	
	for _, key := range shouldNotExist {
		if status.providedKeys[key] {
			t.Errorf("key %s should NOT be recorded as provided", key)
		}
	}
}
