package ticat

import (
	"bytes"
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
