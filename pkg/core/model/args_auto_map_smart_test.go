package model

import (
	"testing"
)

func TestArgsAutoMapSmartMappingWithVal2Env(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create target command with smart mapping
	targetCmd := tree.AddSub("target")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "target command")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	// Create source command with val2env
	srcCmd := tree.AddSub("source")
	src := srcCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "source command")
	src.AddVal2Env("provided.key", "value")
	
	// Verify that the provided key is recorded
	status := target.GetArgsAutoMapStatus()
	status.MarkMet(src)
	
	if !status.providedKeys["provided.key"] {
		t.Error("val2env key should be recorded as provided")
	}
}

func TestArgsAutoMapSmartMappingWithArg2EnvProvided(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create target command with smart mapping
	targetCmd := tree.AddSub("target2")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "target command")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	// Create source command with arg2env
	srcCmd := tree.AddSub("source2")
	src := srcCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "source command")
	src.AddArg("myarg", "default-value")
	src.AddArg2Env("provided.arg.key", "myarg")
	
	// Create argv with provided argument
	argv := ArgVals{
		"myarg": ArgVal{Raw: "provided-value", Provided: true},
	}
	
	// Verify that the provided key is recorded when arg is provided
	status := target.GetArgsAutoMapStatus()
	status.MarkMetWithArgv(src, argv)
	
	if !status.providedKeys["provided.arg.key"] {
		t.Error("arg2env key should be recorded as provided when arg is provided")
	}
}

func TestArgsAutoMapSmartMappingWithArg2EnvNotProvided(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create target command with smart mapping
	targetCmd := tree.AddSub("target3")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "target command")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	// Create source command with arg2env
	srcCmd := tree.AddSub("source3")
	src := srcCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "source command")
	src.AddArg("myarg", "default-value")
	src.AddArg2Env("not.provided.key", "myarg")
	
	// Create argv without provided argument
	argv := ArgVals{
		"myarg": ArgVal{Raw: "default-value", Provided: false},
	}
	
	// Verify that the key is NOT recorded when arg is not provided
	status := target.GetArgsAutoMapStatus()
	status.MarkMetWithArgv(src, argv)
	
	if status.providedKeys["not.provided.key"] {
		t.Error("arg2env key should NOT be recorded as provided when arg is not provided")
	}
}

func TestArgsAutoMapSmartMappingWithGlobalEnv(t *testing.T) {
	strs := CmdTreeStrsForTest()
	tree := NewCmdTree(strs)
	
	// Create target command with smart mapping
	targetCmd := tree.AddSub("target4")
	target := targetCmd.RegPowerCmd(func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "target command")
	
	_, err := target.SetArg2EnvAutoMap([]string{"*"})
	if err != nil {
		t.Fatalf("failed to set auto map: %v", err)
	}
	
	// Verify that global env keys are recorded
	status := target.GetArgsAutoMapStatus()
	status.RecordGlobalEnvKey("global.env.key")
	
	if !status.providedKeys["global.env.key"] {
		t.Error("global env key should be recorded as provided")
	}
}
