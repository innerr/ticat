package display

import (
	"strings"
	"testing"
	"time"

	"github.com/innerr/ticat/pkg/core/model"
)

func TestFilterQuietCmds(t *testing.T) {
	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	quietSub := tree.GetOrAddSub("quiet")
	quietSub.RegEmptyCmd("quiet command").SetQuiet()

	newParsedCmd := func(name string, isQuiet bool) model.ParsedCmd {
		sub := tree.GetOrAddSub(name)
		if sub.Cmd() == nil {
			cic := sub.RegEmptyCmd("test command")
			if isQuiet {
				cic.SetQuiet()
			}
		}
		return model.ParsedCmd{
			Segments: []model.ParsedCmdSeg{
				{
					Matched: model.MatchedCmd{
						Name: name,
						Cmd:  sub,
					},
				},
			},
		}
	}

	t.Run("filter_first_quiet_cmd", func(t *testing.T) {
		env := model.NewEnv()
		flow := []model.ParsedCmd{newParsedCmd("quiet", true), newParsedCmd("normal", false)}
		newFlow, newIdx := filterQuietCmds(env, flow, 1)

		if len(newFlow) != 1 {
			t.Errorf("expected 1 command, got %d", len(newFlow))
		}
		if newIdx != 0 {
			t.Errorf("expected newIdx 0, got %d", newIdx)
		}
	})

	t.Run("filter_multiple_quiet_cmds", func(t *testing.T) {
		env := model.NewEnv()
		flow := []model.ParsedCmd{newParsedCmd("q1", true), newParsedCmd("q2", true), newParsedCmd("normal", false)}
		newFlow, newIdx := filterQuietCmds(env, flow, 2)

		if len(newFlow) != 1 {
			t.Errorf("expected 1 command, got %d", len(newFlow))
		}
		if newIdx != 0 {
			t.Errorf("expected newIdx 0, got %d", newIdx)
		}
	})

	t.Run("no_filter_when_display_quiet_enabled", func(t *testing.T) {
		env := model.NewEnv()
		env.SetBool("display.mod.quiet", true)
		flow := []model.ParsedCmd{newParsedCmd("quiet", true), newParsedCmd("normal", false)}
		newFlow, newIdx := filterQuietCmds(env, flow, 1)

		if len(newFlow) != 2 {
			t.Errorf("expected 2 commands, got %d", len(newFlow))
		}
		if newIdx != 1 {
			t.Errorf("expected newIdx 1, got %d", newIdx)
		}
	})

	t.Run("handle_empty_cmd", func(t *testing.T) {
		env := model.NewEnv()
		emptyCmd := model.ParsedCmd{}
		flow := []model.ParsedCmd{emptyCmd, newParsedCmd("n1", false), newParsedCmd("n2", false)}
		newFlow, newIdx := filterQuietCmds(env, flow, 2)

		if len(newFlow) != 2 {
			t.Errorf("expected 2 commands, got %d", len(newFlow))
		}
		if newIdx != 1 {
			t.Errorf("expected newIdx 1, got %d", newIdx)
		}
	})

	t.Run("all_quiet_cmds", func(t *testing.T) {
		env := model.NewEnv()
		flow := []model.ParsedCmd{newParsedCmd("q1", true), newParsedCmd("q2", true)}
		newFlow, newIdx := filterQuietCmds(env, flow, 0)

		if len(newFlow) != 0 {
			t.Errorf("expected 0 commands, got %d", len(newFlow))
		}
		if newIdx != 0 {
			t.Errorf("expected newIdx 0, got %d", newIdx)
		}
	})

	t.Run("middle_cmd_index", func(t *testing.T) {
		env := model.NewEnv()
		flow := []model.ParsedCmd{newParsedCmd("n1", false), newParsedCmd("n2", false), newParsedCmd("n3", false)}
		newFlow, newIdx := filterQuietCmds(env, flow, 1)

		if len(newFlow) != 3 {
			t.Errorf("expected 3 commands, got %d", len(newFlow))
		}
		if newIdx != 1 {
			t.Errorf("expected newIdx 1, got %d", newIdx)
		}
	})

	t.Run("quiet_cmd_in_middle", func(t *testing.T) {
		env := model.NewEnv()
		flow := []model.ParsedCmd{newParsedCmd("n1", false), newParsedCmd("quiet", true), newParsedCmd("n2", false)}
		newFlow, newIdx := filterQuietCmds(env, flow, 2)

		if len(newFlow) != 2 {
			t.Errorf("expected 2 commands, got %d", len(newFlow))
		}
		if newIdx != 1 {
			t.Errorf("expected newIdx 1, got %d", newIdx)
		}
	})
}

func TestCheckPrintFilter(t *testing.T) {
	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	quietSub := tree.GetOrAddSub("quiet")
	quietSub.RegEmptyCmd("quiet command").SetQuiet()

	normalSub := tree.GetOrAddSub("normal")
	normalSub.RegEmptyCmd("normal command")

	t.Run("empty_cmd_should_skip", func(t *testing.T) {
		env := model.NewEnv()
		cmd := model.ParsedCmd{}
		if !checkPrintFilter(cmd, env) {
			t.Error("empty cmd should return true (skip)")
		}
	})

	t.Run("quiet_cmd_should_skip", func(t *testing.T) {
		env := model.NewEnv()
		cmd := model.ParsedCmd{
			Segments: []model.ParsedCmdSeg{
				{Matched: model.MatchedCmd{Name: "quiet", Cmd: quietSub}},
			},
		}
		if !checkPrintFilter(cmd, env) {
			t.Error("quiet cmd should return true (skip)")
		}
	})

	t.Run("quiet_cmd_with_display_quiet_should_not_skip", func(t *testing.T) {
		env := model.NewEnv()
		env.SetBool("display.mod.quiet", true)
		cmd := model.ParsedCmd{
			Segments: []model.ParsedCmdSeg{
				{Matched: model.MatchedCmd{Name: "quiet", Cmd: quietSub}},
			},
		}
		if checkPrintFilter(cmd, env) {
			t.Error("quiet cmd with display.mod.quiet should return false (not skip)")
		}
	})

	t.Run("normal_cmd_should_not_skip", func(t *testing.T) {
		env := model.NewEnv()
		cmd := model.ParsedCmd{
			Segments: []model.ParsedCmdSeg{
				{Matched: model.MatchedCmd{Name: "normal", Cmd: normalSub}},
			},
		}
		if checkPrintFilter(cmd, env) {
			t.Error("normal cmd should return false (not skip)")
		}
	})
}

func TestPrintCmdStackLines(t *testing.T) {
	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("first command")
	cmd2 := tree.AddSub("cmd2")
	cmd2.RegEmptyCmd("second command")
	cmd3 := tree.AddSub("cmd3")
	cmd3.RegEmptyCmd("third command")

	createEnv := func() *model.Env {
		env := model.NewEnvEx(model.EnvLayerDefault).NewLayer(model.EnvLayerSession)
		env.SetBool("display.utf8", false)
		env.SetBool("display.color", false)
		env.SetBool("display.executor", true)
		env.Set("strs.env-path-sep", ".")
		env.Set("strs.cmd-path-sep", ".")
		return env
	}

	t.Run("simple_stack_with_two_cmds", func(t *testing.T) {
		screen := &memoryScreen{}
		env := createEnv()

		flow := []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd2", Cmd: cmd2}}}},
		}

		lines := PrintCmdStack(false, screen, flow[0], nil, env, nil, flow, 0, tree.Strs, nil, false)
		if !lines.Display {
			t.Error("lines should be displayed for multi-cmd flow")
		}
		if len(lines.Flow) < 1 {
			t.Errorf("expected at least 1 flow line, got %d", len(lines.Flow))
		}
	})

	t.Run("stack_with_display_stack", func(t *testing.T) {
		screen := &memoryScreen{}
		env := createEnv()
		env.SetBool("display.stack", true)
		env.Set("sys.stack", "caller1,caller2")
		env.Set("strs.list-sep", ",")

		flow := []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd2", Cmd: cmd2}}}},
		}

		lines := PrintCmdStack(false, screen, flow[0], nil, env, nil, flow, 0, tree.Strs, nil, false)
		if !lines.Display {
			t.Error("lines should be displayed")
		}
		if len(lines.Stack) == 0 {
			t.Error("expected stack lines when display.stack is true")
		}
	})

	t.Run("stack_with_env_display", func(t *testing.T) {
		screen := &memoryScreen{}
		env := createEnv()
		env.SetBool("display.env", true)
		env.Set("test.key", "test-value")

		flow := []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd2", Cmd: cmd2}}}},
		}

		lines := PrintCmdStack(false, screen, flow[0], nil, env, nil, flow, 0, tree.Strs, nil, false)
		if !lines.Display {
			t.Error("lines should be displayed")
		}
	})

	t.Run("stack_with_mask_executed", func(t *testing.T) {
		screen := &memoryScreen{}
		env := createEnv()

		flow := []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd2", Cmd: cmd2}}}},
		}

		mask := model.NewExecuteMask("cmd1")
		mask.ResultIfExecuted = model.ExecutedResultSucceeded
		executedCmd := model.NewExecutedCmd("cmd1")
		executedCmd.Result = model.ExecutedResultSucceeded
		executedCmd.StartTs = time.Now().Add(-time.Second)
		executedCmd.FinishTs = time.Now()
		mask.ExecutedCmd = executedCmd

		lines := PrintCmdStack(false, screen, flow[0], mask, env, nil, flow, 0, tree.Strs, nil, false)
		if !lines.Display {
			t.Error("lines should be displayed")
		}
	})

	t.Run("disabled_executor", func(t *testing.T) {
		screen := &memoryScreen{}
		env := createEnv()
		env.SetBool("display.executor", false)

		flow := []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
		}

		lines := PrintCmdStack(false, screen, flow[0], nil, env, nil, flow, 0, tree.Strs, nil, false)
		if lines.Display {
			t.Error("lines should not be displayed when display.executor is false")
		}
	})

	t.Run("single_cmd_without_display_one_cmd", func(t *testing.T) {
		screen := &memoryScreen{}
		env := createEnv()

		flow := []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
		}

		lines := PrintCmdStack(false, screen, flow[0], nil, env, nil, flow, 0, tree.Strs, nil, false)
		if lines.Display {
			t.Error("single cmd without display.one-cmd should not display")
		}
	})

	t.Run("single_cmd_with_display_one_cmd", func(t *testing.T) {
		screen := &memoryScreen{}
		env := createEnv()
		env.SetBool("display.one-cmd", true)

		flow := []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
		}

		lines := PrintCmdStack(false, screen, flow[0], nil, env, nil, flow, 0, tree.Strs, nil, false)
		if !lines.Display {
			t.Error("single cmd with display.one-cmd should display")
		}
	})
}

func TestCmdResultLines(t *testing.T) {
	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	createEnv := func() *model.Env {
		env := model.NewEnvEx(model.EnvLayerDefault).NewLayer(model.EnvLayerSession)
		env.SetBool("display.utf8", false)
		env.SetBool("display.color", false)
		env.SetBool("display.executor", true)
		env.SetBool("display.executor.end", true)
		env.Set("strs.cmd-path-sep", ".")
		return env
	}

	t.Run("result_succeeded_with_multi_cmd_flow", func(t *testing.T) {
		screen := &memoryScreen{}
		env := createEnv()

		flow := []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
		}

		lines := PrintCmdResult(nil, false, screen, flow[0], env, true, time.Second, flow, 0, tree.Strs)
		if !lines.Display {
			t.Error("result lines should be displayed")
		}
		if !strings.Contains(lines.Res, "OK") && !strings.Contains(lines.Res, "✓") {
			t.Errorf("expected success indicator, got: %s", lines.Res)
		}
	})

	t.Run("result_failed_with_multi_cmd_flow", func(t *testing.T) {
		screen := &memoryScreen{}
		env := createEnv()

		flow := []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
		}

		lines := PrintCmdResult(nil, false, screen, flow[0], env, false, time.Second, flow, 0, tree.Strs)
		if !lines.Display {
			t.Error("result lines should be displayed")
		}
		if !strings.Contains(lines.Res, "E") && !strings.Contains(lines.Res, "✘") {
			t.Errorf("expected error indicator, got: %s", lines.Res)
		}
	})

	t.Run("result_with_duration", func(t *testing.T) {
		screen := &memoryScreen{}
		env := createEnv()

		flow := []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
		}

		lines := PrintCmdResult(nil, false, screen, flow[0], env, true, 1500*time.Millisecond, flow, 0, tree.Strs)
		if !lines.Display {
			t.Error("result lines should be displayed")
		}
		if lines.DurLen == 0 {
			t.Error("expected duration to be displayed")
		}
	})

	t.Run("disabled_executor_end", func(t *testing.T) {
		screen := &memoryScreen{}
		env := createEnv()
		env.SetBool("display.executor.end", false)

		flow := []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
		}

		lines := PrintCmdResult(nil, false, screen, flow[0], env, true, time.Second, flow, 0, tree.Strs)
		if lines.Display {
			t.Error("result should not display when display.executor.end is false")
		}
	})

	t.Run("single_cmd_with_display_one_cmd", func(t *testing.T) {
		screen := &memoryScreen{}
		env := createEnv()
		env.SetBool("display.one-cmd", true)

		flow := []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
		}

		lines := PrintCmdResult(nil, false, screen, flow[0], env, true, time.Second, flow, 0, tree.Strs)
		if !lines.Display {
			t.Error("result should display with display.one-cmd for single cmd")
		}
	})

	t.Run("single_cmd_without_display_one_cmd", func(t *testing.T) {
		screen := &memoryScreen{}
		env := createEnv()

		flow := []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
		}

		lines := PrintCmdResult(nil, false, screen, flow[0], env, true, time.Second, flow, 0, tree.Strs)
		if lines.Display {
			t.Error("result should not display for single cmd without display.one-cmd")
		}
	})
}

func TestPrintCmdStackWithEnvFilter(t *testing.T) {
	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")
	cmd2 := tree.AddSub("cmd2")
	cmd2.RegEmptyCmd("test command")

	createEnv := func() *model.Env {
		env := model.NewEnvEx(model.EnvLayerDefault).NewLayer(model.EnvLayerSession)
		env.SetBool("display.utf8", false)
		env.SetBool("display.color", false)
		env.SetBool("display.executor", true)
		env.SetBool("display.env", true)
		env.Set("strs.env-path-sep", ".")
		env.Set("strs.cmd-path-sep", ".")
		env.Set("strs.list-sep", ",")
		return env
	}

	t.Run("env_filter_hides_matching_keys", func(t *testing.T) {
		screen := &memoryScreen{}
		env := createEnv()
		env.Set("test.key1", "value1")
		env.Set("test.key2", "value2")
		env.Set("foo.bar", "value3")
		env.Set("display.env.filter.prefix", "test.")

		flow := []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd2", Cmd: cmd2}}}},
		}

		lines := PrintCmdStack(false, screen, flow[0], nil, env, nil, flow, 0, tree.Strs, nil, false)
		if !lines.Display {
			t.Error("lines should be displayed")
		}

		envOutput := strings.Join(lines.Env, " ")
		if strings.Contains(envOutput, "test.key1") {
			t.Error("test.key1 should be filtered out")
		}
		if strings.Contains(envOutput, "test.key2") {
			t.Error("test.key2 should be filtered out")
		}
		if !strings.Contains(envOutput, "foo.bar") {
			t.Error("foo.bar should NOT be filtered out")
		}
	})

	t.Run("env_filter_multiple_prefixes", func(t *testing.T) {
		screen := &memoryScreen{}
		env := createEnv()
		env.Set("test.key1", "value1")
		env.Set("foo.bar", "value2")
		env.Set("baz.qux", "value3")
		env.Set("keep.me", "value4")
		env.Set("display.env.filter.prefix", "test.,foo.")

		flow := []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd2", Cmd: cmd2}}}},
		}

		lines := PrintCmdStack(false, screen, flow[0], nil, env, nil, flow, 0, tree.Strs, nil, false)
		if !lines.Display {
			t.Error("lines should be displayed")
		}

		envOutput := strings.Join(lines.Env, " ")
		if strings.Contains(envOutput, "test.key1") {
			t.Error("test.key1 should be filtered out")
		}
		if strings.Contains(envOutput, "foo.bar") {
			t.Error("foo.bar should be filtered out")
		}
		if !strings.Contains(envOutput, "baz.qux") {
			t.Error("baz.qux should NOT be filtered out")
		}
		if !strings.Contains(envOutput, "keep.me") {
			t.Error("keep.me should NOT be filtered out")
		}
	})

	t.Run("env_filter_empty_shows_all", func(t *testing.T) {
		screen := &memoryScreen{}
		env := createEnv()
		env.Set("test.key1", "value1")
		env.Set("foo.bar", "value2")

		flow := []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd1", Cmd: cmd1}}}},
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "cmd2", Cmd: cmd2}}}},
		}

		lines := PrintCmdStack(false, screen, flow[0], nil, env, nil, flow, 0, tree.Strs, nil, false)
		if !lines.Display {
			t.Error("lines should be displayed")
		}

		envOutput := strings.Join(lines.Env, " ")
		if !strings.Contains(envOutput, "test.key1") {
			t.Error("test.key1 should be visible when no filter")
		}
		if !strings.Contains(envOutput, "foo.bar") {
			t.Error("foo.bar should be visible when no filter")
		}
	})
}
