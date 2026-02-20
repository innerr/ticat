package display

import (
	"strings"
	"testing"

	"github.com/innerr/ticat/pkg/core/model"
)

func TestDumpEnvOpsCheckResultWithSuggestions(t *testing.T) {
	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	writerCmd := tree.AddSub("writer")
	cic := writerCmd.RegEmptyCmd("writes test.key")
	cic.AddEnvOp("test.key", model.EnvOpTypeWrite)

	readerCmd := tree.AddSub("reader")
	readerCmd.RegEmptyCmd("reads test.key").AddEnvOp("test.key", model.EnvOpTypeRead)

	env := model.NewEnvEx(model.EnvLayerDefault).NewLayer(model.EnvLayerSession)
	env.SetBool("display.utf8", false)
	env.SetBool("display.color", false)
	env.SetBool("display.tip", true)
	env.Set("strs.self-name", "ticat")
	env.Set("strs.cmd-path-sep", ".")
	env.Set("display.env.suggest.max-cmds", "3")

	t.Run("verify_writer_matches_write_key", func(t *testing.T) {
		if !writerCmd.MatchWriteKey("test.key") {
			t.Error("writerCmd should match test.key")
		}
		if !cic.MatchWriteKey("test.key") {
			t.Error("cic (writer cmd) should match test.key")
		}
	})

	t.Run("verify_dumpcmds_finds_writer", func(t *testing.T) {
		dumpArgs := NewDumpCmdArgs().SetSkeleton().SetMatchWriteKey("test.key")
		cacheScreen := NewCacheScreen()
		DumpCmds(tree, cacheScreen, env, dumpArgs)
		lines := cacheScreen.Lines()
		if len(lines) == 0 {
			t.Error("DumpCmds should find writer command")
		}
		for i, line := range lines {
			t.Logf("line[%d]: %q", i, line)
		}
		// First line should be [writer]
		if len(lines) > 0 {
			cmdPath := strings.Trim(lines[0], " []")
			if cmdPath != "writer" {
				t.Errorf("expected cmdPath='writer', got %q", cmdPath)
			}
		}
	})

	t.Run("verify_dumpenvopscheckresult_calls_findcmds", func(t *testing.T) {
		result := []model.EnvOpsCheckResult{
			{
				Key:            "test.key",
				CmdDisplayPath: "reader",
				ReadNotExist:   true,
			},
		}
		screen := &memoryScreen{}
		DumpEnvOpsCheckResult(screen, nil, env, result, ".", tree)

		output := screen.GetOutput()
		if !strings.Contains(output, "writer") {
			t.Errorf("expected output to contain 'writer' suggestion, got:\n%s", output)
		}
	})

	tests := []struct {
		name           string
		result         []model.EnvOpsCheckResult
		expectKey      string
		expectCmd      string
		notExpectEmpty bool
	}{
		{
			name: "fatal_with_writer_suggestion",
			result: []model.EnvOpsCheckResult{
				{
					Key:            "test.key",
					CmdDisplayPath: "reader",
					ReadNotExist:   true,
				},
			},
			expectKey:      "test.key",
			expectCmd:      "writer",
			notExpectEmpty: true,
		},
		{
			name: "fatal_without_writer",
			result: []model.EnvOpsCheckResult{
				{
					Key:            "unknown.key",
					CmdDisplayPath: "reader",
					ReadNotExist:   true,
				},
			},
			expectKey:      "unknown.key",
			notExpectEmpty: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			screen := &memoryScreen{}
			DumpEnvOpsCheckResult(screen, nil, env, tc.result, ".", tree)

			output := screen.GetOutput()

			if !strings.Contains(output, tc.expectKey) {
				t.Errorf("expected output to contain key [%s], got:\n%s", tc.expectKey, output)
			}

			if tc.expectCmd != "" && !strings.Contains(output, tc.expectCmd) {
				t.Errorf("expected output to contain command [%s], got:\n%s", tc.expectCmd, output)
			}

			if tc.notExpectEmpty && len(strings.TrimSpace(output)) == 0 {
				t.Error("expected non-empty output")
			}
		})
	}
}

func TestDumpEnvOpsCheckResultEmpty(t *testing.T) {
	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	env := model.NewEnvEx(model.EnvLayerDefault).NewLayer(model.EnvLayerSession)

	screen := &memoryScreen{}
	DumpEnvOpsCheckResult(screen, nil, env, nil, ".", tree)

	if screen.GetOutput() != "" {
		t.Error("expected empty output for empty result")
	}
}

func TestDumpEnvOpsCheckResultRisk(t *testing.T) {
	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	env := model.NewEnvEx(model.EnvLayerDefault).NewLayer(model.EnvLayerSession)
	env.SetBool("display.utf8", false)
	env.SetBool("display.color", false)

	result := []model.EnvOpsCheckResult{
		{
			Key:             "risk.key",
			CmdDisplayPath:  "cmd1",
			MayReadNotExist: true,
		},
	}

	screen := &memoryScreen{}
	DumpEnvOpsCheckResult(screen, nil, env, result, ".", tree)

	output := screen.GetOutput()
	if !strings.Contains(output, "risk") {
		t.Errorf("expected output to contain 'risk', got:\n%s", output)
	}
}

func TestDumpEnvOpsCheckResultWithArg2Env(t *testing.T) {
	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	cmd := tree.AddSub("cmd-with-arg")
	cic := cmd.RegEmptyCmd("command with arg mapped to env")
	cic.AddArg("myarg", "default", "a")
	cic.GetArg2Env().Add("myarg", "test.arg.key")
	cic.AddEnvOp("test.arg.key", model.EnvOpTypeRead)

	env := model.NewEnvEx(model.EnvLayerDefault).NewLayer(model.EnvLayerSession)
	env.SetBool("display.utf8", false)
	env.SetBool("display.color", false)
	env.Set("strs.self-name", "ticat")

	result := []model.EnvOpsCheckResult{
		{
			Key:            "test.arg.key",
			CmdDisplayPath: "cmd-with-arg",
			ReadNotExist:   true,
			FirstArg2Env: &model.ParsedCmd{
				Segments: []model.ParsedCmdSeg{
					{
						Matched: model.MatchedCmd{
							Name: "cmd-with-arg",
							Cmd:  cmd,
						},
					},
				},
			},
		},
	}

	screen := &memoryScreen{}
	DumpEnvOpsCheckResult(screen, nil, env, result, ".", tree)

	output := screen.GetOutput()
	if !strings.Contains(output, "test.arg.key") {
		t.Errorf("expected output to contain 'test.arg.key', got:\n%s", output)
	}
}

func TestDumpEnvOpsCheckResultMultipleFatals(t *testing.T) {
	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	writer1 := tree.AddSub("writer1")
	writer1.RegEmptyCmd("writes key1").AddEnvOp("fatal.key1", model.EnvOpTypeWrite)

	writer2 := tree.AddSub("writer2")
	writer2.RegEmptyCmd("writes key2").AddEnvOp("fatal.key2", model.EnvOpTypeWrite)

	env := model.NewEnvEx(model.EnvLayerDefault).NewLayer(model.EnvLayerSession)
	env.SetBool("display.utf8", false)
	env.SetBool("display.color", false)
	env.SetBool("display.tip", true)
	env.Set("strs.self-name", "ticat")
	env.Set("strs.cmd-path-sep", ".")
	env.Set("display.env.suggest.max-cmds", "3")

	result := []model.EnvOpsCheckResult{
		{
			Key:            "fatal.key1",
			CmdDisplayPath: "reader1",
			ReadNotExist:   true,
		},
		{
			Key:            "fatal.key2",
			CmdDisplayPath: "reader2",
			ReadNotExist:   true,
		},
	}

	screen := &memoryScreen{}
	DumpEnvOpsCheckResult(screen, nil, env, result, ".", tree)

	output := screen.GetOutput()
	if !strings.Contains(output, "fatal.key1") {
		t.Errorf("expected output to contain 'fatal.key1', got:\n%s", output)
	}
	if !strings.Contains(output, "fatal.key2") {
		t.Errorf("expected output to contain 'fatal.key2', got:\n%s", output)
	}
	if !strings.Contains(output, "writer1") {
		t.Errorf("expected output to contain 'writer1' suggestion, got:\n%s", output)
	}
	if !strings.Contains(output, "writer2") {
		t.Errorf("expected output to contain 'writer2' suggestion, got:\n%s", output)
	}
}

func TestDumpEnvOpsCheckResultWithColorEnabled(t *testing.T) {
	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	writerCmd := tree.AddSub("writer")
	cic := writerCmd.RegEmptyCmd("writes test.key")
	cic.AddEnvOp("test.key", model.EnvOpTypeWrite)

	env := model.NewEnvEx(model.EnvLayerDefault).NewLayer(model.EnvLayerSession)
	env.SetBool("display.utf8", false)
	env.SetBool("display.color", true)
	env.SetBool("display.tip", true)
	env.Set("strs.self-name", "ticat")
	env.Set("strs.cmd-path-sep", ".")
	env.Set("display.env.suggest.max-cmds", "3")

	result := []model.EnvOpsCheckResult{
		{
			Key:            "test.key",
			CmdDisplayPath: "reader",
			ReadNotExist:   true,
		},
	}

	screen := &memoryScreen{}
	DumpEnvOpsCheckResult(screen, nil, env, result, ".", tree)

	output := screen.GetOutput()
	if !strings.Contains(output, "commands which can provide these keys") {
		t.Errorf("expected output to contain 'commands which can provide these keys', got:\n%s", output)
	}
	if !strings.Contains(output, "writer") {
		t.Errorf("expected output to contain 'writer' suggestion, got:\n%s", output)
	}
}

func TestAggEnvOpsCheckResult(t *testing.T) {
	tests := []struct {
		name                string
		result              []model.EnvOpsCheckResult
		expectFatalCount    int
		expectRiskCount     int
		expectArg2EnvCanFix bool
	}{
		{
			name:                "empty",
			result:              nil,
			expectFatalCount:    0,
			expectRiskCount:     0,
			expectArg2EnvCanFix: true,
		},
		{
			name: "all_fatals_with_arg2env",
			result: []model.EnvOpsCheckResult{
				{Key: "k1", ReadNotExist: true, FirstArg2Env: &model.ParsedCmd{}},
				{Key: "k2", ReadNotExist: true, FirstArg2Env: &model.ParsedCmd{}},
			},
			expectFatalCount:    2,
			expectRiskCount:     0,
			expectArg2EnvCanFix: true,
		},
		{
			name: "all_fatals_without_arg2env",
			result: []model.EnvOpsCheckResult{
				{Key: "k1", ReadNotExist: true, FirstArg2Env: nil},
				{Key: "k2", ReadNotExist: true, FirstArg2Env: nil},
			},
			expectFatalCount:    2,
			expectRiskCount:     0,
			expectArg2EnvCanFix: false,
		},
		{
			name: "mixed_fatals_arg2env",
			result: []model.EnvOpsCheckResult{
				{Key: "k1", ReadNotExist: true, FirstArg2Env: &model.ParsedCmd{}},
				{Key: "k2", ReadNotExist: true, FirstArg2Env: nil},
			},
			expectFatalCount:    2,
			expectRiskCount:     0,
			expectArg2EnvCanFix: false,
		},
		{
			name: "all_risks",
			result: []model.EnvOpsCheckResult{
				{Key: "k1", MayReadNotExist: true},
				{Key: "k2", ReadMayWrite: true},
			},
			expectFatalCount:    0,
			expectRiskCount:     2,
			expectArg2EnvCanFix: true,
		},
		{
			name: "mixed_fatals_and_risks",
			result: []model.EnvOpsCheckResult{
				{Key: "k1", ReadNotExist: true, FirstArg2Env: &model.ParsedCmd{}},
				{Key: "k2", MayReadNotExist: true},
			},
			expectFatalCount:    1,
			expectRiskCount:     1,
			expectArg2EnvCanFix: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fatals, risks, isArg2EnvCanFix := AggEnvOpsCheckResult(tc.result)

			if len(fatals.Result) != tc.expectFatalCount {
				t.Errorf("expected %d fatals, got %d", tc.expectFatalCount, len(fatals.Result))
			}

			if len(risks.Result) != tc.expectRiskCount {
				t.Errorf("expected %d risks, got %d", tc.expectRiskCount, len(risks.Result))
			}

			if isArg2EnvCanFix != tc.expectArg2EnvCanFix {
				t.Errorf("expected isArg2EnvCanFix=%v, got %v", tc.expectArg2EnvCanFix, isArg2EnvCanFix)
			}
		})
	}
}
