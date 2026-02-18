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

func TestApplyArg2EnvWithIPValue(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "hosts", "")
	cmd.AddArg2Env("env.hosts", "hosts")

	env := NewEnv()
	argv := ArgVals{
		"hosts": ArgVal{Raw: "172.16.5.34,172.16.5.37", Provided: true},
	}

	ApplyArg2Env(env, cmd, argv)

	val := env.Get("env.hosts")
	if val.Raw != "172.16.5.34,172.16.5.37" {
		t.Errorf("Expected env.hosts to be '172.16.5.34,172.16.5.37', got '%s'", val.Raw)
	}
}

func TestApplyArg2EnvWithEqualsInValue(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "config", "")
	cmd.AddArg2Env("env.config", "config")

	env := NewEnv()
	argv := ArgVals{
		"config": ArgVal{Raw: "key=value", Provided: true},
	}

	ApplyArg2Env(env, cmd, argv)

	val := env.Get("env.config")
	if val.Raw != "key=value" {
		t.Errorf("Expected env.config to be 'key=value', got '%s'", val.Raw)
	}
}

func TestApplyArg2EnvWithURLEnvRef(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "url", "")
	cmd.AddArg2Env("env.url", "url")

	env := NewEnv()
	argv := ArgVals{
		"url": ArgVal{Raw: "http://example.com?a=1&b=2", Provided: true},
	}

	ApplyArg2Env(env, cmd, argv)

	val := env.Get("env.url")
	if val.Raw != "http://example.com?a=1&b=2" {
		t.Errorf("Expected env.url to be 'http://example.com?a=1&b=2', got '%s'", val.Raw)
	}
}

func TestApplyArg2EnvWithComplexTemplate(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "host", "")
	cmd.args.AddArg(tree, "port", "")
	cmd.AddArg2Env("env.address", "host")
	cmd.AddArg2Env("env.port", "port")

	env := NewEnv()
	env.Set("prefix", "my")
	argv := ArgVals{
		"host": ArgVal{Raw: "[[prefix]]-server.local", Provided: true},
		"port": ArgVal{Raw: "8080", Provided: true},
	}

	ApplyArg2Env(env, cmd, argv)

	if env.Get("env.address").Raw != "my-server.local" {
		t.Errorf("Expected env.address to be 'my-server.local', got '%s'", env.Get("env.address").Raw)
	}
	if env.Get("env.port").Raw != "8080" {
		t.Errorf("Expected env.port to be '8080', got '%s'", env.Get("env.port").Raw)
	}
}

func TestApplyArg2EnvProvidedOverwritesExisting(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "arg1", "default")
	cmd.AddArg2Env("env.key1", "arg1")

	env := NewEnv()
	env.Set("env.key1", "existing")
	argv := ArgVals{
		"arg1": ArgVal{Raw: "newvalue", Provided: true},
	}

	ApplyArg2Env(env, cmd, argv)

	val := env.Get("env.key1")
	if val.Raw != "newvalue" {
		t.Errorf("Expected env.key1 to be 'newvalue', got '%s'", val.Raw)
	}
}

func TestApplyArg2EnvDefaultWhenNoExisting(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "arg1", "defaultvalue")
	cmd.AddArg2Env("env.key1", "arg1")

	env := NewEnv()
	argv := ArgVals{
		"arg1": ArgVal{Raw: "defaultvalue", Provided: false},
	}

	ApplyArg2Env(env, cmd, argv)

	val := env.Get("env.key1")
	if val.Raw != "defaultvalue" {
		t.Errorf("Expected env.key1 to be 'defaultvalue', got '%s'", val.Raw)
	}
}

func TestApplyArg2EnvEmptyStringProvided(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "arg1", "default")
	cmd.AddArg2Env("env.key1", "arg1")

	env := NewEnv()
	argv := ArgVals{
		"arg1": ArgVal{Raw: "", Provided: true},
	}

	ApplyArg2Env(env, cmd, argv)

	val := env.Get("env.key1")
	if val.Raw != "" {
		t.Errorf("Expected env.key1 to be empty, got '%s'", val.Raw)
	}
}

func TestApplyVal2EnvMultipleValues(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.AddVal2Env("env.key1", "value1")
	cmd.AddVal2Env("env.key2", "value2")
	cmd.AddVal2Env("env.key3", "value3")

	env := NewEnv()
	argv := ArgVals{}

	ApplyVal2Env(env, cmd, argv)

	if env.Get("env.key1").Raw != "value1" {
		t.Errorf("Expected env.key1 to be 'value1', got '%s'", env.Get("env.key1").Raw)
	}
	if env.Get("env.key2").Raw != "value2" {
		t.Errorf("Expected env.key2 to be 'value2', got '%s'", env.Get("env.key2").Raw)
	}
	if env.Get("env.key3").Raw != "value3" {
		t.Errorf("Expected env.key3 to be 'value3', got '%s'", env.Get("env.key3").Raw)
	}
}

func TestApplyVal2EnvWithMultipleEnvRefs(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "host", "")
	cmd.args.AddArg(tree, "port", "")
	cmd.AddVal2Env("env.url", "[[host]]:[[port]]")

	env := NewEnv()
	argv := ArgVals{
		"host": ArgVal{Raw: "localhost", Provided: true},
		"port": ArgVal{Raw: "8080", Provided: true},
	}

	ApplyVal2Env(env, cmd, argv)

	val := env.Get("env.url")
	if val.Raw != "localhost:8080" {
		t.Errorf("Expected env.url to be 'localhost:8080', got '%s'", val.Raw)
	}
}

func TestApplyVal2EnvWithRandomAndStatic(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.AddVal2Env("env.id", "prefix-[[RANDOM]]-suffix")

	env := NewEnv()
	argv := ArgVals{}

	ApplyVal2Env(env, cmd, argv)

	val := env.Get("env.id")
	if val.Raw == "[[RANDOM]]" {
		t.Error("Expected [[RANDOM]] to be rendered")
	}
	if !containsPrefixAndSuffix(val.Raw, "prefix-", "-suffix") {
		t.Errorf("Expected env.id to have prefix 'prefix-' and suffix '-suffix', got '%s'", val.Raw)
	}
}

func containsPrefixAndSuffix(s, prefix, suffix string) bool {
	return len(s) >= len(prefix)+len(suffix) &&
		s[:len(prefix)] == prefix &&
		s[len(s)-len(suffix):] == suffix
}

func TestApplyArg2EnvWithSpecialChars(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"colon separated", "a:b:c", "a:b:c"},
		{"semicolon separated", "a;b;c", "a;b;c"},
		{"slashes", "/path/to/file", "/path/to/file"},
		{"backslashes", `C:\Users\test`, `C:\Users\test`},
		{"at symbol", "user@example.com", "user@example.com"},
		{"hash symbol", "value#comment", "value#comment"},
		{"dollar sign", "$100", "$100"},
		{"parentheses", "(value)", "(value)"},
		{"brackets", "[value]", "[value]"},
		{"braces", "{value}", "{value}"},
		{"spaces", "hello world", "hello world"},
		{"tabs", "hello\tworld", "hello\tworld"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := NewCmdTree(CmdTreeStrsForTest())
			cmd := NewEmptyCmd(tree, "test cmd")
			cmd.args.AddArg(tree, "arg", "")
			cmd.AddArg2Env("env.arg", "arg")

			env := NewEnv()
			argv := ArgVals{
				"arg": ArgVal{Raw: tc.input, Provided: true},
			}

			ApplyArg2Env(env, cmd, argv)

			val := env.Get("env.arg")
			if val.Raw != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, val.Raw)
			}
		})
	}
}

func TestApplyArg2EnvConcurrency(t *testing.T) {
	tree := NewCmdTree(CmdTreeStrsForTest())
	cmd := NewEmptyCmd(tree, "test cmd")
	cmd.args.AddArg(tree, "arg", "[[RANDOM]]")
	cmd.AddArg2Env("env.random", "arg")

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			env := NewEnv()
			argv := ArgVals{
				"arg": ArgVal{Raw: "[[RANDOM]]", Provided: true},
			}
			ApplyArg2Env(env, cmd, argv)
			val := env.Get("env.random")
			if val.Raw == "[[RANDOM]]" || val.Raw == "" {
				t.Error("Random not rendered properly in concurrent access")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
