package model

import (
	"testing"
)

func TestArgsAutoMapSmartMappingIntegration(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create target command with smart mapping
	targetCmd := tree.AddSub("inttarget")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "integration target command")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	// Create source command with multiple env operations
	srcCmd := tree.AddSub("intsource")
	src := srcCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "integration source command")
	
	// Add val2env
	src.AddVal2Env("val.key", "val-value")
	
	// Add arg2env
	src.AddArg("arg1", "default1")
	src.AddArg2Env("arg.key", "arg1")
	
	// Add env op
	src.AddEnvOp("env.op.key", EnvOpTypeWrite)
	
	// Create argv with provided argument
	argv := ArgVals{
		"arg1": ArgVal{Raw: "provided-value", Provided: true},
	}
	
	// Mark the source command with argv
	status := target.GetArgsAutoMapStatus()
	status.MarkMetWithArgv(src, argv)
	
	// Verify all provided keys are recorded
	if !status.providedKeys["val.key"] {
		t.Error("val2env key should be recorded as provided")
	}
	if !status.providedKeys["arg.key"] {
		t.Error("arg2env key with provided arg should be recorded as provided")
	}
	if !status.providedKeys["env.op.key"] {
		t.Error("env op key should be recorded as provided")
	}
	
	// Test that non-provided arg2env is not recorded
	src2Cmd := tree.AddSub("intsource2")
	src2 := src2Cmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "integration source command 2")
	
	src2.AddArg("arg2", "default2")
	src2.AddArg2Env("not.provided.arg.key", "arg2")
	
	argv2 := ArgVals{
		"arg2": ArgVal{Raw: "default2", Provided: false},
	}
	
	status.MarkMetWithArgv(src2, argv2)
	
	if status.providedKeys["not.provided.arg.key"] {
		t.Error("arg2env key with non-provided arg should NOT be recorded as provided")
	}
}

func TestArgsAutoMapFlushCacheWithSmartMapping(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create target command with smart mapping
	targetCmd := tree.AddSub("flushtarget")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "flush target command")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	// Create source command
	srcCmd := tree.AddSub("flushsource")
	src := srcCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "flush source command")
	
	src.AddArg("arg1", "default1")
	
	// Mark source as met and cache a mapping
	status := target.GetArgsAutoMapStatus()
	status.MarkAndCacheMapping(src, "test.key", "test.arg", "default-val", nil, true)
	
	// Record the key as provided
	status.RecordGlobalEnvKey("test.key")
	
	// Flush cache
	err = status.FlushCache(target)
	if err != nil {
		t.Fatalf("failed to flush cache: %v", err)
	}
	
	// Verify the provided key was NOT mapped (smart mapping)
	if target.GetArg2Env().Has("test.key") {
		t.Error("provided key should not be mapped in smart mode")
	}
}
