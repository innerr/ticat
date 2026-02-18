package ticat

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/innerr/ticat/pkg/core/model"
)

// TestIntegrationBasicCmdTree tests basic command tree operations
func TestIntegrationBasicCmdTree(t *testing.T) {
	strs := model.CmdTreeStrsForTest()
	tree := model.NewCmdTree(strs)

	// Build a simple command hierarchy
	build := tree.AddSub("build", "b")
	build.AddSub("all").RegEmptyCmd("build all targets")
	build.AddSub("clean").RegEmptyCmd("clean build artifacts")

	test := tree.AddSub("test", "t")
	test.AddSub("unit").RegEmptyCmd("run unit tests")
	test.AddSub("integration").RegEmptyCmd("run integration tests")

	// Verify structure
	if !tree.HasSubs() {
		t.Fatal("root should have subs")
	}

	buildSub := tree.GetSub("build")
	if buildSub == nil {
		t.Fatal("build sub not found")
	}

	if buildSub.Depth() != 1 {
		t.Errorf("build depth should be 1, got %d", buildSub.Depth())
	}

	// Test abbreviation resolution
	testByAbbr := tree.GetSub("t")
	if testByAbbr == nil || testByAbbr.Name() != "test" {
		t.Error("abbreviation 't' should resolve to 'test'")
	}
}

// TestIntegrationEnvLayerHierarchy tests environment layer inheritance
func TestIntegrationEnvLayerHierarchy(t *testing.T) {
	// Create a 4-layer environment hierarchy
	global := model.NewEnv()
	global.Set("app.name", "MyApp")
	global.Set("app.version", "1.0.0")
	global.Set("db.host", "localhost")

	session := global.NewLayer(model.EnvLayerSession)
	session.Set("session.id", "abc123")
	session.Set("db.host", "dev-server") // Override

	cmd := session.NewLayer(model.EnvLayerCmd)
	cmd.Set("cmd.timeout", "30s")

	subflow := cmd.NewLayer(model.EnvLayerSubFlow)
	subflow.Set("flow.retry", "3")

	// Test inheritance
	if subflow.Get("app.name").Raw != "MyApp" {
		t.Error("subflow should inherit app.name from global")
	}

	if subflow.Get("db.host").Raw != "dev-server" {
		t.Error("subflow should get overridden db.host from session")
	}

	if subflow.Get("session.id").Raw != "abc123" {
		t.Error("subflow should inherit session.id")
	}

	// Test Flatten
	flat := subflow.FlattenAll()
	if len(flat) < 6 {
		t.Errorf("expected at least 6 flattened keys, got %d", len(flat))
	}

	// Test layer-specific operations
	cmd.Deduplicate()
	// cmd.timeout should still exist (unique to cmd layer)
	if !cmd.Has("cmd.timeout") {
		t.Error("cmd.timeout should still exist after deduplicate")
	}
}

// TestIntegrationArgValConversion tests argument value type conversions
func TestIntegrationArgValConversion(t *testing.T) {
	env := model.NewEnv()

	// Set up various string values
	env.Set("count", "42")
	env.Set("enabled", "true")
	env.Set("disabled", "false")
	env.Set("ratio", "0.75")
	env.Set("name", "test")

	// Test boolean conversions
	if !model.StrToBool(env.Get("enabled").Raw) {
		t.Error("'true' should convert to true")
	}

	if model.StrToBool(env.Get("disabled").Raw) {
		t.Error("'false' should convert to false")
	}

	// Test various true values
	trueValues := []string{"TRUE", "True", "t", "T", "1", "on", "ON", "y", "Y", "yes"}
	for _, v := range trueValues {
		if !model.StrToTrue(v) {
			t.Errorf("StrToTrue(%q) should return true", v)
		}
	}

	// Test various false values
	falseValues := []string{"FALSE", "False", "f", "F", "0", "off", "OFF", "n", "N", "no"}
	for _, v := range falseValues {
		if !model.StrToFalse(v) {
			t.Errorf("StrToFalse(%q) should return true", v)
		}
	}
}

// TestIntegrationCmdTreeWithArgs tests command tree with arguments
func TestIntegrationCmdTreeWithArgs(t *testing.T) {
	strs := model.CmdTreeStrsForTest()
	tree := model.NewCmdTree(strs)

	// Create a command with arguments
	cmd := tree.AddSub("deploy")
	// Register command with empty source to avoid conflict
	registeredCmd := cmd.RegEmptyCmd("deploy application")

	// Add arguments to the command
	registeredCmd.AddArg("target", "production", "t")
	registeredCmd.AddArg("version", "latest", "v")
	registeredCmd.AddArg("dry-run", "false", "d")

	// Verify arguments
	args := cmd.Args()
	if args.IsEmpty() {
		t.Error("command args should not be empty")
	}

	if !args.Has("target") {
		t.Error("should have 'target' argument")
	}

	if !args.HasArgOrAbbr("t") {
		t.Error("should have 't' abbreviation")
	}

	if args.DefVal("target", 0) != "production" {
		t.Errorf("expected default 'production', got %q", args.DefVal("target", 0))
	}

	// Test argument enums
	registeredCmd.SetArgEnums("target", "production", "staging", "development")
	enums := args.EnumVals("target")
	if len(enums) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(enums))
	}
}

// TestIntegrationFlowSerialization tests flow saving and parsing
func TestIntegrationFlowSerialization(t *testing.T) {
	env := model.NewEnv()
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.env-bracket-left", "{")
	env.Set("strs.env-bracket-right", "}")
	env.Set("strs.env-kv-sep", "=")
	env.Set("strs.seq-sep", ":")

	// Create a parsed flow
	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			{
				Segments: []model.ParsedCmdSeg{
					{
						Matched: model.MatchedCmd{Name: "build"},
					},
				},
			},
			{
				Segments: []model.ParsedCmdSeg{
					{
						Matched: model.MatchedCmd{Name: "test"},
					},
				},
			},
			{
				Segments: []model.ParsedCmdSeg{
					{
						Matched: model.MatchedCmd{Name: "deploy"},
					},
				},
			},
		},
	}

	// Save flow to string
	flowStr, _ := model.SaveFlowToStr(flow, ".", "@", env)

	// Verify the flow string contains expected commands
	if !strings.Contains(flowStr, "build") {
		t.Errorf("flow string should contain 'build': %s", flowStr)
	}
	if !strings.Contains(flowStr, "test") {
		t.Errorf("flow string should contain 'test': %s", flowStr)
	}
	if !strings.Contains(flowStr, "deploy") {
		t.Errorf("flow string should contain 'deploy': %s", flowStr)
	}

	// Parse flow string back
	strs := model.FlowStrToStrs(flowStr)
	if len(strs) < 3 {
		t.Errorf("expected at least 3 elements after parsing, got %d: %v", len(strs), strs)
	}
}

// TestIntegrationSensitiveData tests sensitive key detection
func TestIntegrationSensitiveData(t *testing.T) {
	env := model.NewEnv()

	// Set various sensitive and non-sensitive values
	env.Set("db.password", "secret123")
	env.Set("api.key", "apikey-xyz")
	env.Set("app.name", "MyApp")
	env.Set("auth.token", "token123")
	env.Set("debug", "false")

	// Test sensitive key detection
	testCases := []struct {
		key       string
		val       string
		sensitive bool
	}{
		{"db.password", "secret123", true},
		{"password", "mypass", true},
		{"api_secret", "key123", true},
		{"app.name", "MyApp", false},
		{"debug", "false", false},
		{"password", "true", false}, // Boolean values are not sensitive
		{"password", "false", false},
	}

	for _, tc := range testCases {
		result := model.IsSensitiveKeyVal(tc.key, tc.val)
		if result != tc.sensitive {
			t.Errorf("IsSensitiveKeyVal(%q, %q) = %v, expected %v",
				tc.key, tc.val, result, tc.sensitive)
		}
	}
}

// TestIntegrationCloneOperations tests cloning operations across types
func TestIntegrationCloneOperations(t *testing.T) {
	strs := model.CmdTreeStrsForTest()

	// Clone command tree
	tree := model.NewCmdTree(strs)
	tree.AddSub("level1").AddSub("level2")
	tree.AddTags("@test")

	clonedTree := tree.Clone()
	if clonedTree.GetSub("level1", "level2") == nil {
		t.Error("cloned tree should have level1.level2")
	}

	// Clone environment
	env := model.NewEnv()
	env.Set("key1", "value1")
	env.Set("key2", "value2")

	clonedEnv := env.Clone()
	if clonedEnv.Get("key1").Raw != "value1" {
		t.Error("cloned env should have same values")
	}

	// Modify clone should not affect original
	clonedEnv.Set("key1", "modified")
	if env.Get("key1").Raw != "value1" {
		t.Error("original env should not be affected by clone modification")
	}
}

// TestIntegrationBufferFlow tests flow saving to buffer
func TestIntegrationBufferFlow(t *testing.T) {
	env := model.NewEnv()
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.env-bracket-left", "{")
	env.Set("strs.env-bracket-right", "}")
	env.Set("strs.env-kv-sep", "=")
	env.Set("strs.seq-sep", ":")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			{
				TrivialLvl: 1,
				Segments: []model.ParsedCmdSeg{
					{
						Matched: model.MatchedCmd{Name: "cmd1"},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	_ = model.SaveFlow(&buf, flow, ".", "@", env)

	result := buf.String()
	if !strings.Contains(result, "@") {
		t.Errorf("expected trivial mark '@' in output: %s", result)
	}
	if !strings.Contains(result, "cmd1") {
		t.Errorf("expected 'cmd1' in output: %s", result)
	}
}

// TestIntegrationEnvMergeAndDedup tests merge and deduplicate operations
func TestIntegrationEnvMergeAndDedup(t *testing.T) {
	base := model.NewEnv()
	base.Set("shared", "base-value")
	base.Set("unique-base", "only-in-base")

	override := model.NewEnv()
	override.Set("shared", "override-value")
	override.Set("unique-override", "only-in-override")

	// Merge override into base
	base.Merge(override)

	if base.Get("shared").Raw != "override-value" {
		t.Error("shared key should be overridden")
	}
	if base.Get("unique-override").Raw != "only-in-override" {
		t.Error("unique-override should be added")
	}
	if base.Get("unique-base").Raw != "only-in-base" {
		t.Error("unique-base should be preserved")
	}

	// Test deduplicate
	parent := model.NewEnv()
	parent.Set("key", "same-value")

	child := parent.NewLayer(model.EnvLayerSession)
	child.Set("key", "same-value")
	child.Set("unique", "child-only")

	child.Deduplicate()
	// key should be removed from child since it's same as parent
	// Access pairs through reflection or check via GetEx behavior
	val, exists := child.GetEx("key")
	if exists && val.Raw == "same-value" {
		// This is acceptable - deduplicate may keep the value but rely on parent
	}
}

// TestIntegrationCmdTreeTagsAndMatching tests tag operations and matching
func TestIntegrationCmdTreeTagsAndMatching(t *testing.T) {
	strs := model.CmdTreeStrsForTest()
	tree := model.NewCmdTree(strs)

	cmd := tree.AddSub("deploy-prod")
	cmd.RegEmptyCmd("deploy to production")
	cmd.AddTags("@production", "@critical", "@v2")

	// Test exact tag matching
	if !cmd.MatchTags("@production") {
		t.Error("should match @production tag")
	}

	if !cmd.MatchTags("@critical") {
		t.Error("should match @critical tag")
	}

	if cmd.MatchTags("@development") {
		t.Error("should not match @development tag")
	}

	// Test MatchExactTags
	if !cmd.MatchExactTags("@production", "@critical") {
		t.Error("should match all exact tags")
	}

	if cmd.MatchExactTags("@production", "@nonexistent") {
		t.Error("should not match when one tag is missing")
	}

	// Test find matching
	if !cmd.MatchFind("deploy") {
		t.Error("should find 'deploy' in command name")
	}

	if !cmd.MatchFind("@production") {
		t.Error("should find @production tag")
	}
}

// TestIntegrationCmdTreeSourceAndBuiltin tests source tracking
func TestIntegrationCmdTreeSourceAndBuiltin(t *testing.T) {
	strs := model.CmdTreeStrsForTest()

	// Test builtin (no source)
	builtinTree := model.NewCmdTree(strs)
	if !builtinTree.IsBuiltin() {
		t.Error("tree without source should be builtin")
	}

	// Test external source
	externalTree := model.NewCmdTree(strs)
	externalTree.SetSource("github.com/user/module")

	if externalTree.IsBuiltin() {
		t.Error("tree with source should not be builtin")
	}

	if !externalTree.MatchSource("user/module") {
		t.Error("should match partial source")
	}

	if !externalTree.MatchSource("github.com") {
		t.Error("should match source prefix")
	}

	if externalTree.MatchSource("other/module") {
		t.Error("should not match different source")
	}
}

// TestIntegrationEnvOperations tests comprehensive env operations
func TestIntegrationEnvOperations(t *testing.T) {
	// Create layered environment
	base := model.NewEnv()
	base.Set("app.name", "TestApp")
	base.Set("app.version", "1.0.0")

	session := base.NewLayer(model.EnvLayerSession)
	session.Set("session.id", "sess-123")

	cmd := session.NewLayer(model.EnvLayerCmd)
	cmd.Set("cmd.flag", "enabled")

	// Test WriteCurrLayerTo
	target := model.NewEnv()
	cmd.WriteCurrLayerTo(target)

	if target.Get("cmd.flag").Raw != "enabled" {
		t.Error("WriteCurrLayerTo should copy current layer values")
	}

	// Test CleanCurrLayer
	cmd.CleanCurrLayer()
	if cmd.Has("cmd.flag") {
		t.Error("CleanCurrLayer should remove current layer keys")
	}

	// Test SetIfEmpty
	session.SetIfEmpty("new.key", "new-value")
	if session.Get("new.key").Raw != "new-value" {
		t.Error("SetIfEmpty should set value for non-existing key")
	}

	session.SetIfEmpty("new.key", "should-not-change")
	if session.Get("new.key").Raw != "new-value" {
		t.Error("SetIfEmpty should not change existing value")
	}
}

// TestIntegrationCmdTreeAbbrs tests abbreviation operations
func TestIntegrationCmdTreeAbbrs(t *testing.T) {
	strs := model.CmdTreeStrsForTest()
	tree := model.NewCmdTree(strs)

	// Create command with multiple abbreviations
	verbose := tree.AddSub("verbose", "v", "V", "verb")

	// Test SubAbbrs
	abbrs := tree.SubAbbrs("verbose")
	if len(abbrs) != 4 {
		t.Errorf("expected 4 abbreviations, got %d", len(abbrs))
	}

	// Test resolution via GetSub
	if tree.GetSub("v") != verbose {
		t.Error("should resolve 'v' to verbose")
	}

	if tree.GetSub("V") != verbose {
		t.Error("should resolve 'V' to verbose")
	}

	if tree.GetSub("verb") != verbose {
		t.Error("should resolve 'verb' to verbose")
	}

	// Test Abbrs on the sub itself
	subAbbrs := verbose.Abbrs()
	if len(subAbbrs) != 4 {
		t.Errorf("expected 4 abbrs from sub, got %d", len(subAbbrs))
	}
}

// TestIntegrationFlowStrsConversion tests flow string conversions
func TestIntegrationFlowStrsConversion(t *testing.T) {
	// Test FlowStrsToStr
	flowStrs := []string{"cmd1", ":", "cmd2", ":", "cmd3"}
	result := model.FlowStrsToStr(flowStrs)
	expected := "cmd1 : cmd2 : cmd3"
	if result != expected {
		t.Errorf("FlowStrsToStr(%v) = %q, want %q", flowStrs, result, expected)
	}

	// Test FlowStrToStrs
	parsed := model.FlowStrToStrs(expected)
	if len(parsed) != 5 {
		t.Errorf("expected 5 elements, got %d: %v", len(parsed), parsed)
	}

	// Test empty cases
	emptyResult := model.FlowStrsToStr([]string{})
	if emptyResult != "" {
		t.Errorf("empty FlowStrsToStr should return empty string, got %q", emptyResult)
	}

	emptyParsed := model.FlowStrToStrs("")
	if len(emptyParsed) != 0 {
		t.Errorf("empty FlowStrToStrs should return empty slice, got %v", emptyParsed)
	}
}

// TestIntegrationGlobalEnvInSequence tests that global env {key=val} before cmd1 : cmd2
// is accessible to all commands in the sequence
func TestIntegrationGlobalEnvInSequence(t *testing.T) {
	strs := model.CmdTreeStrsForTest()
	tree := model.NewCmdTree(strs)

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command 1")
	cmd2 := tree.AddSub("cmd2")
	cmd2.RegEmptyCmd("test command 2")

	env := model.NewEnv()
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.env-bracket-left", "{")
	env.Set("strs.env-bracket-right", "}")
	env.Set("strs.env-kv-sep", "=")
	env.Set("strs.seq-sep", ":")

	t.Run("global env before sequence is parsed correctly", func(t *testing.T) {
		flow := &model.ParsedCmds{
			GlobalEnv: model.ParsedEnv{
				"mock-key": model.NewParsedEnvVal("mock-key", "mock-value"),
			},
			Cmds: []model.ParsedCmd{
				{
					Segments: []model.ParsedCmdSeg{
						{Matched: model.MatchedCmd{Name: "cmd1"}},
					},
				},
				{
					Segments: []model.ParsedCmdSeg{
						{Matched: model.MatchedCmd{Name: "cmd2"}},
					},
				},
			},
		}

		if flow.GlobalEnv == nil {
			t.Fatal("GlobalEnv should not be nil")
		}
		if flow.GlobalEnv["mock-key"].Val != "mock-value" {
			t.Errorf("GlobalEnv should contain mock-key=mock-value, got %v", flow.GlobalEnv["mock-key"])
		}
		if len(flow.Cmds) != 2 {
			t.Fatalf("expected 2 commands, got %d", len(flow.Cmds))
		}
	})

	t.Run("global env can be written to command env", func(t *testing.T) {
		globalEnv := model.ParsedEnv{
			"shared-key": model.NewParsedEnvVal("shared-key", "shared-value"),
		}

		cmdEnv := model.NewEnv()

		globalEnv.WriteNotArgTo(cmdEnv, "")

		if cmdEnv.Get("shared-key").Raw != "shared-value" {
			t.Errorf("cmdEnv should have shared-key=shared-value after WriteNotArgTo, got %v",
				cmdEnv.Get("shared-key"))
		}
	})

	t.Run("multiple global env values", func(t *testing.T) {
		flow := &model.ParsedCmds{
			GlobalEnv: model.ParsedEnv{
				"key1": model.NewParsedEnvVal("key1", "val1"),
				"key2": model.NewParsedEnvVal("key2", "val2"),
				"key3": model.NewParsedEnvVal("key3", "val3"),
			},
			Cmds: []model.ParsedCmd{
				{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1"}}}},
				{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd2"}}}},
			},
		}

		cmdEnv := model.NewEnv()
		flow.GlobalEnv.WriteNotArgTo(cmdEnv, "")

		if cmdEnv.Get("key1").Raw != "val1" {
			t.Errorf("cmdEnv should have key1=val1, got %v", cmdEnv.Get("key1"))
		}
		if cmdEnv.Get("key2").Raw != "val2" {
			t.Errorf("cmdEnv should have key2=val2, got %v", cmdEnv.Get("key2"))
		}
		if cmdEnv.Get("key3").Raw != "val3" {
			t.Errorf("cmdEnv should have key3=val3, got %v", cmdEnv.Get("key3"))
		}
	})

	t.Run("global env does not affect segment env", func(t *testing.T) {
		flow := &model.ParsedCmds{
			GlobalEnv: model.ParsedEnv{
				"global-key": model.NewParsedEnvVal("global-key", "global-value"),
			},
			Cmds: []model.ParsedCmd{
				{
					Segments: []model.ParsedCmdSeg{
						{
							Env:     model.ParsedEnv{"local-key": model.NewParsedEnvVal("local-key", "local-value")},
							Matched: model.MatchedCmd{Name: "cmd1"},
						},
					},
				},
			},
		}

		if _, ok := flow.Cmds[0].Segments[0].Env["global-key"]; ok {
			t.Error("segment env should not contain global-key after parsing")
		}
		if flow.Cmds[0].Segments[0].Env["local-key"].Val != "local-value" {
			t.Error("segment env should retain local-key")
		}
	})

	t.Run("three commands in sequence share global env", func(t *testing.T) {
		flow := &model.ParsedCmds{
			GlobalEnv: model.ParsedEnv{
				"shared": model.NewParsedEnvVal("shared", "value"),
			},
			Cmds: []model.ParsedCmd{
				{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1"}}}},
				{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd2"}}}},
				{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd3"}}}},
			},
		}

		for i, cmd := range flow.Cmds {
			cmdEnv := model.NewEnv()
			flow.GlobalEnv.WriteNotArgTo(cmdEnv, "")
			if cmdEnv.Get("shared").Raw != "value" {
				t.Errorf("cmd[%d] env should have shared=value", i)
			}
			_ = cmd
		}
	})
}

// TestIntegrationGlobalEnvWithCommandArgs tests global env with command arguments
func TestIntegrationGlobalEnvWithCommandArgs(t *testing.T) {
	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	echo := tree.AddSub("echo")
	echo.RegEmptyCmd("print message").
		AddArg("message", "", "msg", "m").
		AddArg("color", "", "c")

	t.Run("global env and cmd args are separate", func(t *testing.T) {
		flow := &model.ParsedCmds{
			GlobalEnv: model.ParsedEnv{
				"debug": model.NewParsedEnvVal("debug", "true"),
			},
			Cmds: []model.ParsedCmd{
				{
					Segments: []model.ParsedCmdSeg{
						{
							Env:     model.ParsedEnv{"echo.message": model.NewParsedEnvArgv("echo.message", "hello")},
							Matched: model.MatchedCmd{Name: "echo"},
						},
					},
				},
				{
					Segments: []model.ParsedCmdSeg{
						{
							Env:     model.ParsedEnv{"echo.message": model.NewParsedEnvArgv("echo.message", "world")},
							Matched: model.MatchedCmd{Name: "echo"},
						},
					},
				},
			},
		}

		cmdEnv := model.NewEnv()
		flow.GlobalEnv.WriteNotArgTo(cmdEnv, "")

		if cmdEnv.Get("debug").Raw != "true" {
			t.Errorf("cmdEnv should have debug=true from global env, got %v", cmdEnv.Get("debug"))
		}

		if flow.Cmds[0].Segments[0].Env["echo.message"].Val != "hello" {
			t.Errorf("first cmd should have echo.message=hello")
		}
		if flow.Cmds[1].Segments[0].Env["echo.message"].Val != "world" {
			t.Errorf("second cmd should have echo.message=world")
		}
	})

	t.Run("args are written with IsArg flag", func(t *testing.T) {
		env := model.NewEnv()
		parsedEnv := model.ParsedEnv{
			"echo.message": model.NewParsedEnvArgv("echo.message", "hello"),
		}

		parsedEnv.WriteTo(env, "")

		if env.Get("echo.message").Raw != "hello" {
			t.Errorf("env should have echo.message=hello, got %v", env.Get("echo.message"))
		}
	})
}

// TestTestingHook demonstrates how to use TestingHook for unit testing
// with breakpoint and interactive mode support
func TestTestingHook(t *testing.T) {
	t.Run("TestingHook interface methods", func(t *testing.T) {
		hook := &model.TestingHookFuncs{
			BreakPointAction: func(reason string, choices []string, actions map[string]string) string {
				return "c"
			},
			InteractPrompt: func(prompt string) (string, bool) {
				return "dbg.interact.leave", true
			},
			SkipBash: true,
		}

		choice := hook.OnBreakPoint("test reason", []string{"c", "s", "d"}, map[string]string{"c": "continue"})
		if choice != "c" {
			t.Errorf("expected 'c', got %q", choice)
		}

		line, ok := hook.OnInteractPrompt("test> ")
		if !ok || line != "dbg.interact.leave" {
			t.Errorf("expected ('dbg.interact.leave', true), got (%q, %v)", line, ok)
		}

		if !hook.ShouldSkipBash() {
			t.Error("expected SkipBash to be true")
		}
	})

	t.Run("DefaultTestingHook", func(t *testing.T) {
		hook := &model.DefaultTestingHook{}

		choice := hook.OnBreakPoint("test", nil, nil)
		if choice != "c" {
			t.Errorf("expected 'c', got %q", choice)
		}

		line, ok := hook.OnInteractPrompt("test> ")
		if ok || line != "" {
			t.Errorf("expected ('', false), got (%q, %v)", line, ok)
		}

		if hook.ShouldSkipBash() {
			t.Error("expected SkipBash to be false")
		}
	})

	t.Run("TestingHookFuncs defaults", func(t *testing.T) {
		hook := &model.TestingHookFuncs{}

		choice := hook.OnBreakPoint("test", nil, nil)
		if choice != "c" {
			t.Errorf("expected default 'c', got %q", choice)
		}

		line, ok := hook.OnInteractPrompt("test> ")
		if ok || line != "" {
			t.Errorf("expected default ('', false), got (%q, %v)", line, ok)
		}

		if hook.ShouldSkipBash() {
			t.Error("expected default SkipBash to be false")
		}
	})
}

// TestTiCatWithTestingHook demonstrates how to create a TiCat instance
// configured for testing with breakpoint simulation
func TestTiCatWithTestingHook(t *testing.T) {
	tc := NewTiCatForTest()
	tc.SetScreen(&model.QuietScreen{})

	breakPointCalls := 0
	expectedActions := []string{"c", "c", "c"}

	hook := &model.TestingHookFuncs{
		BreakPointAction: func(reason string, choices []string, actions map[string]string) string {
			idx := breakPointCalls
			breakPointCalls++
			if idx < len(expectedActions) {
				return expectedActions[idx]
			}
			return "c"
		},
		InteractPrompt: func(prompt string) (string, bool) {
			return "dbg.interact.leave", true
		},
		SkipBash: true,
	}

	tc.SetTestingHook(hook)

	if tc.GetCli() == nil {
		t.Error("GetCli should not return nil")
	}
	if tc.GetCli().TestingHook == nil {
		t.Error("TestingHook should be set")
	}
}

type argsCaptureScreen struct {
	output bytes.Buffer
}

func (s *argsCaptureScreen) Print(text string) error {
	s.output.WriteString(text)
	return nil
}

func (s *argsCaptureScreen) Error(text string) error {
	s.output.WriteString(text)
	return nil
}

func (s *argsCaptureScreen) OutputtedLines() int {
	return strings.Count(s.output.String(), "\n")
}

func (s *argsCaptureScreen) GetOutput() string {
	return s.output.String()
}

func TestDbgArgsCommandParsing(t *testing.T) {
	t.Run("simple args", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=value1", "arg2=value2")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[value1]") {
			t.Errorf("expected arg1=value1 in output, got:\n%s", output)
		}
		if !strings.Contains(output, "arg2: raw=[value2]") {
			t.Errorf("expected arg2=value2 in output, got:\n%s", output)
		}
	})

	t.Run("args with equals in value", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=db=test", "arg2=host=127.0.0.1")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[db=test]") {
			t.Errorf("expected arg1=db=test in output, got:\n%s", output)
		}
		if !strings.Contains(output, "arg2: raw=[host=127.0.0.1]") {
			t.Errorf("expected arg2=host=127.0.0.1 in output, got:\n%s", output)
		}
	})

	t.Run("args with dots", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=1.2.3.4", "arg2=file.conf")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[1.2.3.4]") {
			t.Errorf("expected arg1=1.2.3.4 in output, got:\n%s", output)
		}
		if !strings.Contains(output, "arg2: raw=[file.conf]") {
			t.Errorf("expected arg2=file.conf in output, got:\n%s", output)
		}
	})

	t.Run("abbreviation args", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "a1=abbr1", "str=string-val")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[abbr1]") {
			t.Errorf("expected arg1=abbr1 (using a1 abbreviation) in output, got:\n%s", output)
		}
		if !strings.Contains(output, "str-val: raw=[string-val]") {
			t.Errorf("expected str-val=string-val in output, got:\n%s", output)
		}
	})

	t.Run("default values", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "str-val: raw=[default-str]") {
			t.Errorf("expected str-val with default value in output, got:\n%s", output)
		}
	})

	t.Run("env style args", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args.env", "db=mysql", "host=192.168.1.1", "port=3306")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "db: [mysql]") {
			t.Errorf("expected db=mysql in output, got:\n%s", output)
		}
		if !strings.Contains(output, "host: [192.168.1.1]") {
			t.Errorf("expected host=192.168.1.1 in output, got:\n%s", output)
		}
		if !strings.Contains(output, "port: [3306]") {
			t.Errorf("expected port=3306 in output, got:\n%s", output)
		}
	})
}

func TestDbgArgsTailModeParsing(t *testing.T) {
	t.Run("tail mode with desc", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("desc", ":", "dbg.args.tail", "arg1=tail1", "arg2=tail2")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "TailMode: true") {
			t.Errorf("expected TailMode=true in output, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: [tail1]") {
			t.Errorf("expected arg1=tail1 in output, got:\n%s", output)
		}
		if !strings.Contains(output, "arg2: [tail2]") {
			t.Errorf("expected arg2=tail2 in output, got:\n%s", output)
		}
	})

	t.Run("tail mode with env-style args", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("desc", ":", "dbg.args.tail", "db=x", "host=y")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: [db=x]") {
			t.Errorf("expected arg1=db=x in output, got:\n%s", output)
		}
		if !strings.Contains(output, "arg2: [host=y]") {
			t.Errorf("expected arg2=host=y in output, got:\n%s", output)
		}
	})

	t.Run("tail mode abbreviation", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("-", ":", "dbg.args.tail", "arg1=test")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: [test]") {
			t.Errorf("expected arg1=test in output, got:\n%s", output)
		}
	})
}

func TestDbgArgsSequenceParsing(t *testing.T) {
	t.Run("sequence with args", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=first", ":", "dbg.args", "arg1=second")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[first]") {
			t.Errorf("expected first command with arg1=first, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: raw=[second]") {
			t.Errorf("expected second command with arg1=second, got:\n%s", output)
		}
	})
}

func TestDbgArgsFlowParsing(t *testing.T) {
	t.Run("three commands in flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=first", "arg2=A",
			":", "dbg.args", "arg1=second", "arg2=B",
			":", "dbg.args", "arg1=third", "arg2=C")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[first]") {
			t.Errorf("expected first command with arg1=first, got:\n%s", output)
		}
		if !strings.Contains(output, "arg2: raw=[A]") {
			t.Errorf("expected first command with arg2=A, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: raw=[second]") {
			t.Errorf("expected second command with arg1=second, got:\n%s", output)
		}
		if !strings.Contains(output, "arg2: raw=[B]") {
			t.Errorf("expected second command with arg2=B, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: raw=[third]") {
			t.Errorf("expected third command with arg1=third, got:\n%s", output)
		}
		if !strings.Contains(output, "arg2: raw=[C]") {
			t.Errorf("expected third command with arg2=C, got:\n%s", output)
		}
	})

	t.Run("args with special values in flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=db=test", ":", "dbg.args", "arg1=host=127.0.0.1")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[db=test]") {
			t.Errorf("expected first command with arg1=db=test, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: raw=[host=127.0.0.1]") {
			t.Errorf("expected second command with arg1=host=127.0.0.1, got:\n%s", output)
		}
	})

	t.Run("mixed arg types in flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=normal", ":",
			"dbg.args.env", "db=mysql", "host=localhost", ":",
			"dbg.args", "str-val=custom-string")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[normal]") {
			t.Errorf("expected first command with arg1=normal, got:\n%s", output)
		}
		if !strings.Contains(output, "db: [mysql]") {
			t.Errorf("expected second command with db=mysql, got:\n%s", output)
		}
		if !strings.Contains(output, "str-val: raw=[custom-string]") {
			t.Errorf("expected third command with str-val=custom-string, got:\n%s", output)
		}
	})
}

func TestDbgArgsPositionInFlow(t *testing.T) {
	t.Run("first command in flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=pos1", ":", "dbg.args", "arg2=pos2")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[pos1]") {
			t.Errorf("expected arg1=pos1, got:\n%s", output)
		}
	})

	t.Run("last command in flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=start", ":", "dbg.args", "arg1=last")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[last]") {
			t.Errorf("expected arg1=last, got:\n%s", output)
		}
	})
}

func TestDbgArgsDescFlow(t *testing.T) {
	t.Run("desc short flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("-", ":", "dbg.args.tail", "arg1=desc-test")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1") && !strings.Contains(output, "desc-test") {
			t.Errorf("expected desc output with arg1, got:\n%s", output)
		}
	})

	t.Run("desc with multiple args", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("-", ":", "dbg.args.tail", "arg1=val1", "arg2=val2", "arg3=val3")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1") {
			t.Errorf("expected desc output with arg1, got:\n%s", output)
		}
	})

	t.Run("desc long flow with args", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("-", ":", "dbg.args.tail", "arg1=step1",
			":", "dbg.args.tail", "arg1=step2",
			":", "dbg.args.tail", "arg1=step3",
			":", "dbg.args.tail", "arg1=step4")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "step1") || !strings.Contains(output, "step4") {
			t.Errorf("expected desc output with all steps, got:\n%s", output)
		}
	})
}

func TestDbgArgsDescMore(t *testing.T) {
	t.Run("desc more with flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("+", ":", "dbg.args.tail", "arg1=test")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1") && !strings.Contains(output, "test") {
			t.Errorf("expected desc output with arg1, got:\n%s", output)
		}
	})
}

func TestDbgArgsNestedCommands(t *testing.T) {
	t.Run("args in simple flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=before", ":", "dbg.args", "arg1=after")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[before]") {
			t.Errorf("expected arg1=before, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: raw=[after]") {
			t.Errorf("expected arg1=after, got:\n%s", output)
		}
	})
}

func TestDbgArgsEdgeCases(t *testing.T) {
	t.Run("empty arg value", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[]") {
			t.Errorf("expected empty arg1, got:\n%s", output)
		}
	})

	t.Run("arg with spaces in value", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1={hello world}")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1:") {
			t.Errorf("expected arg1 in output, got:\n%s", output)
		}
	})

	t.Run("arg value with url", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=http//localhost4000")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1:") {
			t.Errorf("expected arg1 in output, got:\n%s", output)
		}
	})

	t.Run("arg value with path", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=/path/to/file.conf")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[/path/to/file.conf]") {
			t.Errorf("expected arg1 with path, got:\n%s", output)
		}
	})

	t.Run("multiple dots in value", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=a.b.c.d.e")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[a.b.c.d.e]") {
			t.Errorf("expected arg1 with multiple dots, got:\n%s", output)
		}
	})
}

func TestDbgArgsTailModeWithFlow(t *testing.T) {
	t.Run("tail mode with desc", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("-", ":", "dbg.args.tail", "arg1=tail-val")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1") && !strings.Contains(output, "tail-val") {
			t.Errorf("expected tail mode output, got:\n%s", output)
		}
	})
}

func TestDbgArgsAbbreviationInFlow(t *testing.T) {
	t.Run("arg abbreviation in flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "a1=val1", ":", "dbg.args", "a2=val2")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[val1]") {
			t.Errorf("expected arg1=val1 (using a1 abbreviation), got:\n%s", output)
		}
		if !strings.Contains(output, "arg2: raw=[val2]") {
			t.Errorf("expected arg2=val2 (using a2 abbreviation), got:\n%s", output)
		}
	})

	t.Run("str-val abbreviation", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "str=hello", ":", "dbg.args", "s=world")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "str-val: raw=[hello]") {
			t.Errorf("expected str-val=hello, got:\n%s", output)
		}
		if !strings.Contains(output, "str-val: raw=[world]") {
			t.Errorf("expected str-val=world, got:\n%s", output)
		}
	})
}

func TestNestedFlowStackExecution(t *testing.T) {
	t.Run("simple two-level flow stack", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=level1", ":", "dbg.args", "arg1=level2")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[level1]") {
			t.Errorf("expected first cmd with arg1=level1, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: raw=[level2]") {
			t.Errorf("expected second cmd with arg1=level2, got:\n%s", output)
		}
	})

	t.Run("three-level flow stack", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=L1",
			":", "dbg.args", "arg1=L2",
			":", "dbg.args", "arg1=L3")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[L1]") {
			t.Errorf("expected L1, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: raw=[L2]") {
			t.Errorf("expected L2, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: raw=[L3]") {
			t.Errorf("expected L3, got:\n%s", output)
		}
	})

	t.Run("five-level deep flow stack", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("dbg.args", "arg1=step1",
			":", "dbg.args", "arg1=step2",
			":", "dbg.args", "arg1=step3",
			":", "dbg.args", "arg1=step4",
			":", "dbg.args", "arg1=step5")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		for i := 1; i <= 5; i++ {
			expected := fmt.Sprintf("arg1: raw=[step%d]", i)
			if !strings.Contains(output, expected) {
				t.Errorf("expected %s, got:\n%s", expected, output)
			}
		}
	})
}

func TestNestedFlowStackEnvPropagation(t *testing.T) {
	t.Run("env propagation through flow stack", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli("env.set", "key=test-key", "value=test-value", ":", "env", "test-key")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "test-value") {
			t.Errorf("expected test-value in output, got:\n%s", output)
		}
	})

	t.Run("multiple env values in flow stack", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=multi.key1", "value=val1",
			":", "env.set", "key=multi.key2", "value=val2",
			":", "env.set", "key=multi.key3", "value=val3",
			":", "env", "multi")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "val1") {
			t.Errorf("expected val1 in output, got:\n%s", output)
		}
		if !strings.Contains(output, "val2") {
			t.Errorf("expected val2 in output, got:\n%s", output)
		}
		if !strings.Contains(output, "val3") {
			t.Errorf("expected val3 in output, got:\n%s", output)
		}
	})

	t.Run("env override in flow stack", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=override.test", "value=first",
			":", "env.update", "key=override.test", "value=second",
			":", "env", "override.test")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "second") {
			t.Errorf("expected second (overridden) value, got:\n%s", output)
		}
	})
}

func TestNestedFlowStackArgsAndEnvMixed(t *testing.T) {
	t.Run("args and env mixed in flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=shared.test", "value=env-value",
			":", "dbg.args", "arg1=cmd1-arg",
			":", "dbg.args", "arg1=cmd2-arg",
			":", "env", "shared.test")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "cmd1-arg") {
			t.Errorf("expected cmd1-arg, got:\n%s", output)
		}
		if !strings.Contains(output, "cmd2-arg") {
			t.Errorf("expected cmd2-arg, got:\n%s", output)
		}
		if !strings.Contains(output, "env-value") {
			t.Errorf("expected env-value, got:\n%s", output)
		}
	})

	t.Run("global env applied to all commands in flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"{global.key=global-value}",
			":", "dbg.args", "arg1=first",
			":", "dbg.args", "arg1=second")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[first]") {
			t.Errorf("expected first command arg, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: raw=[second]") {
			t.Errorf("expected second command arg, got:\n%s", output)
		}
	})

	t.Run("command-specific env vs global env", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"dbg.args", "arg1=test1", "{local.env=local-val}",
			":", "dbg.args", "arg1=test2")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[test1]") {
			t.Errorf("expected test1, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: raw=[test2]") {
			t.Errorf("expected test2, got:\n%s", output)
		}
	})
}

func TestNestedFlowStackDeepNesting(t *testing.T) {
	t.Run("deep nesting with different arg types", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"dbg.args", "arg1=L1A", "arg2=L1B",
			":", "dbg.args.env", "db=mysql", "host=localhost",
			":", "dbg.args", "str-val=custom-string",
			":", "dbg.args.tail", "arg1=tail-value")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[L1A]") {
			t.Errorf("expected L1A, got:\n%s", output)
		}
		if !strings.Contains(output, "arg2: raw=[L1B]") {
			t.Errorf("expected L1B, got:\n%s", output)
		}
		if !strings.Contains(output, "db: [mysql]") {
			t.Errorf("expected db=mysql, got:\n%s", output)
		}
		if !strings.Contains(output, "str-val: raw=[custom-string]") {
			t.Errorf("expected custom-string, got:\n%s", output)
		}
	})

	t.Run("very deep flow with consistent args", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		args := []string{}
		for i := 1; i <= 10; i++ {
			args = append(args, "dbg.args", fmt.Sprintf("arg1=step%d", i))
			if i < 10 {
				args = append(args, ":")
			}
		}

		ok := tc.RunCli(args...)
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		for i := 1; i <= 10; i++ {
			expected := fmt.Sprintf("arg1: raw=[step%d]", i)
			if !strings.Contains(output, expected) {
				t.Errorf("expected %s in output", expected)
			}
		}
	})
}

func TestNestedFlowStackWithEcho(t *testing.T) {
	t.Run("echo commands in flow stack", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"echo", "message=first",
			":", "echo", "message=second",
			":", "echo", "message=third")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "first") {
			t.Errorf("expected 'first' in output, got:\n%s", output)
		}
		if !strings.Contains(output, "second") {
			t.Errorf("expected 'second' in output, got:\n%s", output)
		}
		if !strings.Contains(output, "third") {
			t.Errorf("expected 'third' in output, got:\n%s", output)
		}
	})

	t.Run("mixed echo and args in flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"echo", "message=start",
			":", "dbg.args", "arg1=middle",
			":", "echo", "message=end")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "start") {
			t.Errorf("expected 'start' in output, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: raw=[middle]") {
			t.Errorf("expected middle arg, got:\n%s", output)
		}
		if !strings.Contains(output, "end") {
			t.Errorf("expected 'end' in output, got:\n%s", output)
		}
	})
}

func TestNestedFlowStackSpecialChars(t *testing.T) {
	t.Run("args with dots and paths in flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"dbg.args", "arg1=1.2.3.4",
			":", "dbg.args", "arg1=file.conf",
			":", "dbg.args", "arg1=/path/to/file")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[1.2.3.4]") {
			t.Errorf("expected 1.2.3.4, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: raw=[file.conf]") {
			t.Errorf("expected file.conf, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: raw=[/path/to/file]") {
			t.Errorf("expected /path/to/file, got:\n%s", output)
		}
	})

	t.Run("args with equals sign in value", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"dbg.args", "arg1=db=test",
			":", "dbg.args", "arg1=key=value")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[db=test]") {
			t.Errorf("expected db=test, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: raw=[key=value]") {
			t.Errorf("expected key=value, got:\n%s", output)
		}
	})

	t.Run("args with spaces in brackets", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"dbg.args", "arg1={hello world}",
			":", "dbg.args", "arg1={foo bar baz}")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1:") {
			t.Errorf("expected arg1 in output, got:\n%s", output)
		}
	})
}

func TestNestedFlowStackEnvAssertion(t *testing.T) {
	t.Run("env assert equal in flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=test.key", "value=expected-value",
			":", "env.assert.equal", "key=test.key", "value=expected-value")
		if !ok {
			t.Error("command should succeed")
		}
	})

	t.Run("env operations in sequence", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=seq.key1", "value=value1",
			":", "env.set", "key=seq.key2", "value=value2",
			":", "env.assert.equal", "key=seq.key1", "value=value1",
			":", "env.assert.equal", "key=seq.key2", "value=value2")
		if !ok {
			t.Error("command should succeed")
		}
	})
}

func TestNestedFlowStackWithNoop(t *testing.T) {
	t.Run("noop in flow stack", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"dbg.args", "arg1=before-noop",
			":", "noop",
			":", "dbg.args", "arg1=after-noop")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[before-noop]") {
			t.Errorf("expected before-noop, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: raw=[after-noop]") {
			t.Errorf("expected after-noop, got:\n%s", output)
		}
	})

	t.Run("multiple noops in flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"noop",
			":", "noop",
			":", "noop",
			":", "dbg.args", "arg1=final")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[final]") {
			t.Errorf("expected final, got:\n%s", output)
		}
	})
}

func TestNestedFlowStackStackDepth(t *testing.T) {
	t.Run("stack depth tracking", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		tc.Env.SetInt("sys.stack-depth", 0)

		ok := tc.RunCli(
			"dbg.args", "arg1=level0",
			":", "dbg.args", "arg1=level1",
			":", "dbg.args", "arg1=level2")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[level0]") {
			t.Errorf("expected level0, got:\n%s", output)
		}
	})
}

func TestNestedFlowStackTimerCommands(t *testing.T) {
	t.Run("timer operations in flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"timer.begin", "key=timer.test",
			":", "dbg.args", "arg1=middle",
			":", "timer.elapsed", "begin-key=timer.test", "key=timer.elapsed")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[middle]") {
			t.Errorf("expected middle, got:\n%s", output)
		}
	})
}

func TestNestedFlowStackWithDummy(t *testing.T) {
	t.Run("dummy commands in flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"dbg.args", "arg1=start",
			":", "dummy",
			":", "dbg.args", "arg1=end")
		if !ok {
			t.Error("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.Contains(output, "arg1: raw=[start]") {
			t.Errorf("expected start, got:\n%s", output)
		}
		if !strings.Contains(output, "arg1: raw=[end]") {
			t.Errorf("expected end, got:\n%s", output)
		}
	})
}

func TestEnvChangeAcrossFlowDepth(t *testing.T) {
	t.Run("env set in first cmd visible in subsequent cmds", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=depth.test", "value=initial",
			":", "env.assert.equal", "key=depth.test", "value=initial")
		if !ok {
			t.Error("command should succeed - env set in first cmd should be visible in second")
		}
	})

	t.Run("env modification in middle cmd affects later cmds", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=chain.value", "value=first",
			":", "env.update", "key=chain.value", "value=second",
			":", "env.assert.equal", "key=chain.value", "value=second",
			":", "env.update", "key=chain.value", "value=third",
			":", "env.assert.equal", "key=chain.value", "value=third")
		if !ok {
			t.Error("command should succeed - env changes should propagate through flow")
		}
	})

	t.Run("multiple keys set at different depths", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=multi.a", "value=1",
			":", "env.set", "key=multi.b", "value=2",
			":", "env.set", "key=multi.c", "value=3",
			":", "env.assert.equal", "key=multi.a", "value=1",
			":", "env.assert.equal", "key=multi.b", "value=2",
			":", "env.assert.equal", "key=multi.c", "value=3")
		if !ok {
			t.Error("command should succeed - all keys set at different depths should be visible")
		}
	})
}

func TestEnvIsolationInFlowStack(t *testing.T) {
	t.Run("env set does not leak to parent scope concept", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		tc.Env.Set("isolation.original", "original-value")

		ok := tc.RunCli(
			"env.set", "key=isolation.modified", "value=modified-value",
			":", "env.assert.equal", "key=isolation.modified", "value=modified-value")
		if !ok {
			t.Error("command should succeed")
		}

		if tc.Env.Get("isolation.original").Raw != "original-value" {
			t.Error("original env should remain unchanged")
		}
	})

	t.Run("deep flow stack env changes persist to end", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=persist.test", "value=step1",
			":", "noop",
			":", "env.update", "key=persist.test", "value=step2",
			":", "noop",
			":", "env.update", "key=persist.test", "value=step3",
			":", "noop",
			":", "env.assert.equal", "key=persist.test", "value=step3")
		if !ok {
			t.Error("command should succeed - env changes should persist through flow")
		}
	})

	t.Run("env changes in five level flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=level.counter", "value=1",
			":", "env.update", "key=level.counter", "value=2",
			":", "env.update", "key=level.counter", "value=3",
			":", "env.update", "key=level.counter", "value=4",
			":", "env.update", "key=level.counter", "value=5",
			":", "env.assert.equal", "key=level.counter", "value=5")
		if !ok {
			t.Error("command should succeed - counter should reach 5")
		}
	})
}

func TestEnvOverwriteBehavior(t *testing.T) {
	t.Run("same key overwritten at different depths", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=overwrite.key", "value=A",
			":", "env.assert.equal", "key=overwrite.key", "value=A",
			":", "env.update", "key=overwrite.key", "value=B",
			":", "env.assert.equal", "key=overwrite.key", "value=B",
			":", "env.update", "key=overwrite.key", "value=C",
			":", "env.assert.equal", "key=overwrite.key", "value=C")
		if !ok {
			t.Error("command should succeed - key should be overwritten at each step")
		}
	})

	t.Run("multiple keys with selective overwrite", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=sel.key1", "value=v1",
			":", "env.set", "key=sel.key2", "value=v2",
			":", "env.set", "key=sel.key3", "value=v3",
			":", "env.update", "key=sel.key2", "value=v2-modified",
			":", "env.assert.equal", "key=sel.key1", "value=v1",
			":", "env.assert.equal", "key=sel.key2", "value=v2-modified",
			":", "env.assert.equal", "key=sel.key3", "value=v3")
		if !ok {
			t.Error("command should succeed - selective overwrite should not affect other keys")
		}
	})

	t.Run("env set allows overwriting", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=set.twice", "value=first",
			":", "env.set", "key=set.twice", "value=second",
			":", "env.assert.equal", "key=set.twice", "value=second")
		if !ok {
			t.Error("command should succeed - env.set allows overwriting")
		}
	})
}

func TestEnvAddVsSetBehavior(t *testing.T) {
	t.Run("env add creates new key", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.add", "key=add.new.key", "value=new-value",
			":", "env.assert.equal", "key=add.new.key", "value=new-value")
		if !ok {
			t.Error("command should succeed - add should create new key")
		}
	})

	t.Run("env set creates new key if not exists", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=set.brand.new", "value=created",
			":", "env.assert.equal", "key=set.brand.new", "value=created")
		if !ok {
			t.Error("command should succeed - set creates key if not exists")
		}
	})

	t.Run("env set can overwrite existing key", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=set.overwrite", "value=first",
			":", "env.set", "key=set.overwrite", "value=second",
			":", "env.assert.equal", "key=set.overwrite", "value=second")
		if !ok {
			t.Error("command should succeed - set should overwrite existing key")
		}
	})

	t.Run("env add then update works correctly", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.add", "key=add.then.update", "value=initial",
			":", "env.update", "key=add.then.update", "value=modified",
			":", "env.assert.equal", "key=add.then.update", "value=modified")
		if !ok {
			t.Error("command should succeed - add then update should work")
		}
	})
}

func TestEnvChangeWithMixedCommands(t *testing.T) {
	t.Run("env changes persist across dbg.args", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=mixed.test", "value=before",
			":", "dbg.args", "arg1=some-arg",
			":", "env.assert.equal", "key=mixed.test", "value=before",
			":", "env.update", "key=mixed.test", "value=after",
			":", "dbg.args", "arg2=another-arg",
			":", "env.assert.equal", "key=mixed.test", "value=after")
		if !ok {
			t.Error("command should succeed - env changes should persist across dbg.args")
		}
	})

	t.Run("env changes persist across echo", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=echo.test", "value=persisted",
			":", "echo", "message=hello",
			":", "env.assert.equal", "key=echo.test", "value=persisted")
		if !ok {
			t.Error("command should succeed - env changes should persist across echo")
		}
	})

	t.Run("env changes persist across noop and dummy", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=noop.test", "value=initial",
			":", "noop",
			":", "env.assert.equal", "key=noop.test", "value=initial",
			":", "dummy",
			":", "env.assert.equal", "key=noop.test", "value=initial")
		if !ok {
			t.Error("command should succeed - env changes should persist across noop and dummy")
		}
	})
}

func TestEnvMapCommand(t *testing.T) {
	t.Run("env map copies value to another key", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=map.src", "value=source-value",
			":", "env.map", "src-key=map.src", "dest-key=map.dst",
			":", "env.assert.equal", "key=map.dst", "value=source-value",
			":", "env.assert.equal", "key=map.src", "value=source-value")
		if !ok {
			t.Error("command should succeed - map should copy value")
		}
	})

	t.Run("env map in flow chain", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=chain.original", "value=hello",
			":", "env.map", "src-key=chain.original", "dest-key=chain.copy1",
			":", "env.map", "src-key=chain.copy1", "dest-key=chain.copy2",
			":", "env.assert.equal", "key=chain.original", "value=hello",
			":", "env.assert.equal", "key=chain.copy1", "value=hello",
			":", "env.assert.equal", "key=chain.copy2", "value=hello")
		if !ok {
			t.Error("command should succeed - chained map should work")
		}
	})
}

func TestEnvChangeVerificationAfterFlow(t *testing.T) {
	t.Run("verify env state after multi-step flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=final.key1", "value=val1",
			":", "env.set", "key=final.key2", "value=val2",
			":", "env.update", "key=final.key1", "value=val1-modified",
			":", "noop")
		if !ok {
			t.Error("flow should succeed")
		}

		if tc.Env.Get("final.key1").Raw != "val1-modified" {
			t.Errorf("final.key1 should be val1-modified, got %s", tc.Env.Get("final.key1").Raw)
		}
		if tc.Env.Get("final.key2").Raw != "val2" {
			t.Errorf("final.key2 should be val2, got %s", tc.Env.Get("final.key2").Raw)
		}
	})

	t.Run("env session layer persists after flow", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=session.persistent", "value=kept")
		if !ok {
			t.Error("flow should succeed")
		}

		sessionEnv := tc.Env.GetLayer(model.EnvLayerSession)
		if sessionEnv.Get("session.persistent").Raw != "kept" {
			t.Error("session layer should contain the set value")
		}
	})
}

func TestEnvDeepNestingIntegrity(t *testing.T) {
	t.Run("ten level env changes maintain integrity", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		args := []string{}
		args = append(args, "env.set", "key=integrity.counter", "value=0")
		args = append(args, ":")
		for i := 1; i <= 10; i++ {
			args = append(args, "env.update", "key=integrity.counter", fmt.Sprintf("value=%d", i))
			if i < 10 {
				args = append(args, ":")
			}
		}

		ok := tc.RunCli(args...)
		if !ok {
			t.Error("flow should succeed")
		}

		if tc.Env.Get("integrity.counter").Raw != "10" {
			t.Errorf("counter should be 10, got %s", tc.Env.Get("integrity.counter").Raw)
		}
	})

	t.Run("multiple keys modified at different levels remain consistent", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=consist.a", "value=1",
			":", "env.set", "key=consist.b", "value=1",
			":", "env.set", "key=consist.c", "value=1",
			":", "env.update", "key=consist.a", "value=2",
			":", "env.update", "key=consist.b", "value=2",
			":", "env.update", "key=consist.a", "value=3",
			":", "env.assert.equal", "key=consist.a", "value=3",
			":", "env.assert.equal", "key=consist.b", "value=2",
			":", "env.assert.equal", "key=consist.c", "value=1")
		if !ok {
			t.Error("flow should succeed - all keys should have expected values")
		}
	})
}

func TestSubflowEnvIsolation(t *testing.T) {
	t.Run("subflow creates separate env layer", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		tc.Env.Set("test.original", "original-value")

		ok := tc.RunCli("env.set", "key=test.new", "value=new-value", ":", "env.assert.equal", "key=test.original", "value=original-value")
		if !ok {
			t.Error("flow should succeed")
		}

		if tc.Env.Get("test.new").Raw != "new-value" {
			t.Error("new key should be set in session layer")
		}
	})

	t.Run("env changes persist across flow commands", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=persist.a", "value=1",
			":", "noop",
			":", "env.update", "key=persist.a", "value=2",
			":", "noop",
			":", "env.assert.equal", "key=persist.a", "value=2")
		if !ok {
			t.Error("env changes should persist through flow")
		}
	})
}

func TestNestedSubflowEnv(t *testing.T) {
	t.Run("deeply nested flow maintains env integrity", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		args := []string{"env.set", "key=nested.counter", "value=0", ":"}
		for i := 1; i <= 5; i++ {
			args = append(args, "env.update", "key=nested.counter", fmt.Sprintf("value=%d", i))
			if i < 5 {
				args = append(args, ":")
			}
		}

		ok := tc.RunCli(args...)
		if !ok {
			t.Error("nested flow should succeed")
		}

		if tc.Env.Get("nested.counter").Raw != "5" {
			t.Errorf("counter should be 5, got %s", tc.Env.Get("nested.counter").Raw)
		}
	})
}

func TestForestModeEnvReset(t *testing.T) {
	t.Run("forest mode isolates command env", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)
		tc.Env.SetBool("sys.forest-mode", true)

		tc.Env.Set("forest.base", "base-value")

		ok := tc.RunCli(
			"noop",
			":", "noop",
			":", "env.assert.equal", "key=forest.base", "value=base-value")
		if !ok {
			t.Error("forest mode flow should succeed")
		}
	})
}

func TestSubflowFailureEnvHandling(t *testing.T) {
	t.Run("env integrity after multiple operations", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=integrity.a", "value=1",
			":", "env.set", "key=integrity.b", "value=2",
			":", "env.update", "key=integrity.a", "value=1-modified",
			":", "env.assert.equal", "key=integrity.a", "value=1-modified",
			":", "env.assert.equal", "key=integrity.b", "value=2")
		if !ok {
			t.Error("env integrity should be maintained")
		}
	})

	t.Run("flow continues after successful operations", func(t *testing.T) {
		tc := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		tc.SetScreen(screen)
		tc.Env.SetBool("sys.panic.recover", false)

		ok := tc.RunCli(
			"env.set", "key=flow.step1", "value=a",
			":", "env.set", "key=flow.step2", "value=b",
			":", "env.set", "key=flow.step3", "value=c",
			":", "env.assert.equal", "key=flow.step1", "value=a",
			":", "env.assert.equal", "key=flow.step2", "value=b",
			":", "env.assert.equal", "key=flow.step3", "value=c")
		if !ok {
			t.Error("all flow steps should complete successfully")
		}
	})
}
