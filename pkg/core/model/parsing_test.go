package model

import (
	"strconv"
	"testing"
)

func TestApplyArg2EnvBasic(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "arg1", "default1")
	cmd.AddArg2Env("env.key1", "arg1")

	env := NewEnv()
	argv := ArgVals{
		"arg1": ArgVal{Raw: "value1", Provided: true},
	}

	ApplyArg2Env(env, cmd, argv)

	val := env.Get("env.key1")
	if val.Raw != "value1" {
		t.Errorf("Expected env.key1 to be 'value1', got '%s'", val.Raw)
	}
}

func TestApplyArg2EnvNotProvided(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "arg1", "default1")
	cmd.AddArg2Env("env.key1", "arg1")

	env := NewEnv()
	argv := ArgVals{
		"arg1": ArgVal{Raw: "default1", Provided: false},
	}

	ApplyArg2Env(env, cmd, argv)

	val := env.Get("env.key1")
	if val.Raw != "default1" {
		t.Errorf("Expected env.key1 to be 'default1', got '%s'", val.Raw)
	}
}

func TestApplyArg2EnvExistingEnvNotProvided(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "arg1", "default1")
	cmd.AddArg2Env("env.key1", "arg1")

	env := NewEnv()
	env.Set("env.key1", "existing")
	argv := ArgVals{
		"arg1": ArgVal{Raw: "default1", Provided: false},
	}

	ApplyArg2Env(env, cmd, argv)

	val := env.Get("env.key1")
	if val.Raw != "existing" {
		t.Errorf("Expected env.key1 to remain 'existing', got '%s'", val.Raw)
	}
}

func TestApplyArg2EnvRandom(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "randarg", "[[RANDOM]]")
	cmd.AddArg2Env("env.random", "randarg")

	env := NewEnv()
	argv := ArgVals{
		"randarg": ArgVal{Raw: "[[RANDOM]]", Provided: true},
	}

	ApplyArg2Env(env, cmd, argv)

	val := env.Get("env.random")
	if val.Raw == "[[RANDOM]]" {
		t.Error("Expected [[RANDOM]] to be rendered, but it was not")
	}
	if val.Raw == "" {
		t.Error("Expected [[RANDOM]] to be rendered to a non-empty value")
	}
	_, err := strconv.Atoi(val.Raw)
	if err != nil {
		t.Errorf("Expected [[RANDOM]] to be rendered to an integer, got '%s'", val.Raw)
	}
}

func TestApplyArg2EnvRandomMultiple(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "randarg", "[[RANDOM]]")
	cmd.AddArg2Env("env.random1", "randarg")

	env1 := NewEnv()
	env2 := NewEnv()
	argv := ArgVals{
		"randarg": ArgVal{Raw: "[[RANDOM]]", Provided: true},
	}

	ApplyArg2Env(env1, cmd, argv)
	ApplyArg2Env(env2, cmd, argv)

	val1 := env1.Get("env.random1").Raw
	val2 := env2.Get("env.random1").Raw

	if val1 == "[[RANDOM]]" || val2 == "[[RANDOM]]" {
		t.Error("Expected [[RANDOM]] to be rendered")
	}
}

func TestApplyArg2EnvNoMapping(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "arg1", "default1")

	env := NewEnv()
	argv := ArgVals{
		"arg1": ArgVal{Raw: "value1", Provided: true},
	}

	ApplyArg2Env(env, cmd, argv)

	if env.Has("arg1") {
		t.Error("Expected no env mapping when arg2env is not set")
	}
}

func TestApplyArg2EnvEmptyArg(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "arg1", "")
	cmd.AddArg2Env("env.key1", "arg1")

	env := NewEnv()
	argv := ArgVals{
		"arg1": ArgVal{Raw: "", Provided: false},
	}

	ApplyArg2Env(env, cmd, argv)

	if env.Has("env.key1") {
		t.Error("Expected no env mapping for empty and not provided arg")
	}
}

func TestApplyArg2EnvMultipleArgs(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "arg1", "default1")
	cmd.args.AddArg(tree, "arg2", "default2")
	cmd.AddArg2Env("env.key1", "arg1")
	cmd.AddArg2Env("env.key2", "arg2")

	env := NewEnv()
	argv := ArgVals{
		"arg1": ArgVal{Raw: "value1", Provided: true},
		"arg2": ArgVal{Raw: "value2", Provided: true},
	}

	ApplyArg2Env(env, cmd, argv)

	if env.Get("env.key1").Raw != "value1" {
		t.Errorf("Expected env.key1 to be 'value1', got '%s'", env.Get("env.key1").Raw)
	}
	if env.Get("env.key2").Raw != "value2" {
		t.Errorf("Expected env.key2 to be 'value2', got '%s'", env.Get("env.key2").Raw)
	}
}

func TestApplyArg2EnvWithEnvVar(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "arg1", "")
	cmd.AddArg2Env("env.key1", "arg1")

	env := NewEnv()
	env.Set("prefix", "hello")
	argv := ArgVals{
		"arg1": ArgVal{Raw: "[[prefix]]-world", Provided: true},
	}

	ApplyArg2Env(env, cmd, argv)

	val := env.Get("env.key1")
	if val.Raw != "hello-world" {
		t.Errorf("Expected env.key1 to be 'hello-world', got '%s'", val.Raw)
	}
}

func TestApplyVal2EnvBasic(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.AddVal2Env("env.key1", "value1")

	env := NewEnv()
	argv := ArgVals{}

	ApplyVal2Env(env, cmd, argv)

	val := env.Get("env.key1")
	if val.Raw != "value1" {
		t.Errorf("Expected env.key1 to be 'value1', got '%s'", val.Raw)
	}
}

func TestApplyVal2EnvRandom(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.AddVal2Env("env.random", "[[RANDOM]]")

	env := NewEnv()
	argv := ArgVals{}

	ApplyVal2Env(env, cmd, argv)

	val := env.Get("env.random")
	if val.Raw == "[[RANDOM]]" {
		t.Error("Expected [[RANDOM]] to be rendered, but it was not")
	}
	if val.Raw == "" {
		t.Error("Expected [[RANDOM]] to be rendered to a non-empty value")
	}
	_, err := strconv.Atoi(val.Raw)
	if err != nil {
		t.Errorf("Expected [[RANDOM]] to be rendered to an integer, got '%s'", val.Raw)
	}
}

func TestApplyVal2EnvWithArg(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "name", "default")
	cmd.AddVal2Env("env.greeting", "[[name]]-world")

	env := NewEnv()
	argv := ArgVals{
		"name": ArgVal{Raw: "hello", Provided: true},
	}

	ApplyVal2Env(env, cmd, argv)

	val := env.Get("env.greeting")
	if val.Raw != "hello-world" {
		t.Errorf("Expected env.greeting to be 'hello-world', got '%s'", val.Raw)
	}
}
