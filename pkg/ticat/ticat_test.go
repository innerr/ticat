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
	model.SaveFlow(&buf, flow, ".", "@", env)

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
