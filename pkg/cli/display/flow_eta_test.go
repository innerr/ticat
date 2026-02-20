package display

import (
	"strings"
	"testing"
	"time"

	"github.com/innerr/ticat/pkg/core/model"
)

func TestCalculateETAStats(t *testing.T) {
	tests := []struct {
		name            string
		executedCmds    []*model.ExecutedCmd
		totalFlowCmds   int
		fromCmdIdx      int
		running         bool
		expectCompleted int
		expectRemaining int
		expectHasETA    bool
	}{
		{
			name: "no_commands",
			executedCmds: []*model.ExecutedCmd{
				{Cmd: "cmd1", Result: model.ExecutedResultUnRun},
				{Cmd: "cmd2", Result: model.ExecutedResultUnRun},
			},
			totalFlowCmds:   2,
			fromCmdIdx:      0,
			running:         true,
			expectCompleted: 0,
			expectRemaining: 2,
			expectHasETA:    false,
		},
		{
			name: "one_completed_one_remaining",
			executedCmds: []*model.ExecutedCmd{
				{Cmd: "cmd1", Result: model.ExecutedResultSucceeded, StartTs: time.Now().Add(-2 * time.Second), FinishTs: time.Now().Add(-1 * time.Second)},
				{Cmd: "cmd2", Result: model.ExecutedResultUnRun},
			},
			totalFlowCmds:   2,
			fromCmdIdx:      0,
			running:         true,
			expectCompleted: 1,
			expectRemaining: 1,
			expectHasETA:    true,
		},
		{
			name: "all_completed",
			executedCmds: []*model.ExecutedCmd{
				{Cmd: "cmd1", Result: model.ExecutedResultSucceeded, StartTs: time.Now().Add(-3 * time.Second), FinishTs: time.Now().Add(-2 * time.Second)},
				{Cmd: "cmd2", Result: model.ExecutedResultSucceeded, StartTs: time.Now().Add(-2 * time.Second), FinishTs: time.Now().Add(-1 * time.Second)},
			},
			totalFlowCmds:   2,
			fromCmdIdx:      0,
			running:         true,
			expectCompleted: 2,
			expectRemaining: 0,
			expectHasETA:    false,
		},
		{
			name: "running_command",
			executedCmds: []*model.ExecutedCmd{
				{Cmd: "cmd1", Result: model.ExecutedResultSucceeded, StartTs: time.Now().Add(-3 * time.Second), FinishTs: time.Now().Add(-2 * time.Second)},
				{Cmd: "cmd2", Result: model.ExecutedResultIncompleted, StartTs: time.Now().Add(-1 * time.Second)},
				{Cmd: "cmd3", Result: model.ExecutedResultUnRun},
			},
			totalFlowCmds:   3,
			fromCmdIdx:      0,
			running:         true,
			expectCompleted: 2,
			expectRemaining: 1,
			expectHasETA:    true,
		},
		{
			name: "skipped_command",
			executedCmds: []*model.ExecutedCmd{
				{Cmd: "cmd1", Result: model.ExecutedResultSucceeded, StartTs: time.Now().Add(-3 * time.Second), FinishTs: time.Now().Add(-2 * time.Second)},
				{Cmd: "cmd2", Result: model.ExecutedResultSkipped},
				{Cmd: "cmd3", Result: model.ExecutedResultUnRun},
			},
			totalFlowCmds:   3,
			fromCmdIdx:      0,
			running:         true,
			expectCompleted: 2,
			expectRemaining: 1,
			expectHasETA:    true,
		},
		{
			name: "error_command",
			executedCmds: []*model.ExecutedCmd{
				{Cmd: "cmd1", Result: model.ExecutedResultSucceeded, StartTs: time.Now().Add(-3 * time.Second), FinishTs: time.Now().Add(-2 * time.Second)},
				{Cmd: "cmd2", Result: model.ExecutedResultError, StartTs: time.Now().Add(-2 * time.Second), FinishTs: time.Now().Add(-1 * time.Second)},
				{Cmd: "cmd3", Result: model.ExecutedResultUnRun},
			},
			totalFlowCmds:   3,
			fromCmdIdx:      0,
			running:         true,
			expectCompleted: 2,
			expectRemaining: 1,
			expectHasETA:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flow := &model.ParsedCmds{}
			for i := 0; i < tc.totalFlowCmds; i++ {
				flow.Cmds = append(flow.Cmds, model.ParsedCmd{})
			}

			executedFlow := model.NewExecutedFlow("test-session")
			executedFlow.Cmds = tc.executedCmds

			stats := calculateETAStats(executedFlow, flow, tc.fromCmdIdx, tc.running)

			if stats.completedCmds != tc.expectCompleted {
				t.Errorf("completedCmds: expected %d, got %d", tc.expectCompleted, stats.completedCmds)
			}
			if stats.remainingCmds != tc.expectRemaining {
				t.Errorf("remainingCmds: expected %d, got %d", tc.expectRemaining, stats.remainingCmds)
			}

			hasETA := stats.eta > 0
			if hasETA != tc.expectHasETA {
				t.Errorf("hasETA: expected %v, got %v (eta=%v)", tc.expectHasETA, hasETA, stats.eta)
			}
		})
	}
}

func TestCalculateMonitorETA(t *testing.T) {
	tests := []struct {
		name          string
		executedCmds  []*model.ExecutedCmd
		totalFlowCmds int
		fromCmdIdx    int
		running       bool
		expectEmpty   bool
		shouldContain []string
	}{
		{
			name: "no_completed_commands",
			executedCmds: []*model.ExecutedCmd{
				{Cmd: "cmd1", Result: model.ExecutedResultUnRun},
				{Cmd: "cmd2", Result: model.ExecutedResultUnRun},
			},
			totalFlowCmds: 2,
			fromCmdIdx:    0,
			running:       true,
			expectEmpty:   true,
		},
		{
			name: "all_completed",
			executedCmds: []*model.ExecutedCmd{
				{Cmd: "cmd1", Result: model.ExecutedResultSucceeded, StartTs: time.Now().Add(-2 * time.Second), FinishTs: time.Now().Add(-1 * time.Second)},
				{Cmd: "cmd2", Result: model.ExecutedResultSucceeded, StartTs: time.Now().Add(-1 * time.Second), FinishTs: time.Now()},
			},
			totalFlowCmds: 2,
			fromCmdIdx:    0,
			running:       true,
			expectEmpty:   true,
		},
		{
			name: "one_completed_one_remaining",
			executedCmds: []*model.ExecutedCmd{
				{Cmd: "cmd1", Result: model.ExecutedResultSucceeded, StartTs: time.Now().Add(-2 * time.Second), FinishTs: time.Now().Add(-1 * time.Second)},
				{Cmd: "cmd2", Result: model.ExecutedResultUnRun},
			},
			totalFlowCmds: 2,
			fromCmdIdx:    0,
			running:       true,
			expectEmpty:   false,
			shouldContain: []string{"avg:", "progress:", "1/2"},
		},
		{
			name: "multiple_remaining",
			executedCmds: []*model.ExecutedCmd{
				{Cmd: "cmd1", Result: model.ExecutedResultSucceeded, StartTs: time.Now().Add(-3 * time.Second), FinishTs: time.Now().Add(-2 * time.Second)},
				{Cmd: "cmd2", Result: model.ExecutedResultSucceeded, StartTs: time.Now().Add(-2 * time.Second), FinishTs: time.Now().Add(-1 * time.Second)},
				{Cmd: "cmd3", Result: model.ExecutedResultUnRun},
				{Cmd: "cmd4", Result: model.ExecutedResultUnRun},
			},
			totalFlowCmds: 4,
			fromCmdIdx:    0,
			running:       true,
			expectEmpty:   false,
			shouldContain: []string{"avg:", "2/4"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flow := &model.ParsedCmds{}
			for i := 0; i < tc.totalFlowCmds; i++ {
				flow.Cmds = append(flow.Cmds, model.ParsedCmd{})
			}

			executedFlow := model.NewExecutedFlow("test-session")
			executedFlow.Cmds = tc.executedCmds

			etaStr := calculateMonitorETA(executedFlow, flow, tc.fromCmdIdx)

			if tc.expectEmpty {
				if len(etaStr) != 0 {
					t.Errorf("expected empty ETA, got: %s", etaStr)
				}
				return
			}

			if len(etaStr) == 0 {
				t.Error("expected non-empty ETA, got empty string")
				return
			}

			for _, substr := range tc.shouldContain {
				if !strings.Contains(etaStr, substr) {
					t.Errorf("ETA string '%s' should contain '%s'", etaStr, substr)
				}
			}
		})
	}
}

func TestMonitorModeETA(t *testing.T) {
	tree := createTestCmdTree()
	cc := createTestCli(&memoryScreen{}, tree)
	env := createTestEnv()

	cmd1 := tree.AddSub("prepare")
	cmd1.RegEmptyCmd("prepare command")

	cmd2 := tree.AddSub("process")
	cmd2.RegEmptyCmd("process command")

	cmd3 := tree.AddSub("cleanup")
	cmd3.RegEmptyCmd("cleanup command")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			newParsedCmd("prepare", cmd1),
			newParsedCmd("process", cmd2),
			newParsedCmd("cleanup", cmd3),
		},
	}

	executedFlow := model.NewExecutedFlow("monitor-eta-session")
	executedFlow.Cmds = []*model.ExecutedCmd{
		{
			Cmd:      "prepare",
			Result:   model.ExecutedResultSucceeded,
			StartTs:  time.Now().Add(-2 * time.Second),
			FinishTs: time.Now().Add(-1 * time.Second),
		},
		{
			Cmd:     "process",
			Result:  model.ExecutedResultIncompleted,
			StartTs: time.Now().Add(-500 * time.Millisecond),
		},
	}

	screen := &memoryScreen{}
	cc.Screen = screen

	args := NewDumpFlowArgs().SetMonitorMode()
	DumpFlowEx(cc, env, flow, 0, args, executedFlow, true, nil)

	output := screen.GetOutput()

	if !strings.Contains(output, "ETA:") {
		t.Errorf("monitor mode output should contain ETA, got:\n%s", output)
	}

	if !strings.Contains(output, "progress:") {
		t.Errorf("monitor mode output should contain progress, got:\n%s", output)
	}

	t.Logf("Monitor output:\n%s", output)
}

func TestMonitorModeNoETAWhenAllCompleted(t *testing.T) {
	tree := createTestCmdTree()
	cc := createTestCli(&memoryScreen{}, tree)
	env := createTestEnv()

	cmd1 := tree.AddSub("task1")
	cmd1.RegEmptyCmd("task 1")

	cmd2 := tree.AddSub("task2")
	cmd2.RegEmptyCmd("task 2")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			newParsedCmd("task1", cmd1),
			newParsedCmd("task2", cmd2),
		},
	}

	executedFlow := model.NewExecutedFlow("completed-session")
	executedFlow.Cmds = []*model.ExecutedCmd{
		{
			Cmd:      "task1",
			Result:   model.ExecutedResultSucceeded,
			StartTs:  time.Now().Add(-2 * time.Second),
			FinishTs: time.Now().Add(-1 * time.Second),
		},
		{
			Cmd:      "task2",
			Result:   model.ExecutedResultSucceeded,
			StartTs:  time.Now().Add(-1 * time.Second),
			FinishTs: time.Now(),
		},
	}

	screen := &memoryScreen{}
	cc.Screen = screen

	args := NewDumpFlowArgs().SetMonitorMode()
	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()

	if strings.Contains(output, "ETA:") {
		t.Errorf("monitor mode should NOT contain ETA when all completed, got:\n%s", output)
	}
}

func TestMonitorModeNoETAWhenNoCompleted(t *testing.T) {
	tree := createTestCmdTree()
	cc := createTestCli(&memoryScreen{}, tree)
	env := createTestEnv()

	cmd1 := tree.AddSub("task1")
	cmd1.RegEmptyCmd("task 1")

	cmd2 := tree.AddSub("task2")
	cmd2.RegEmptyCmd("task 2")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			newParsedCmd("task1", cmd1),
			newParsedCmd("task2", cmd2),
		},
	}

	executedFlow := model.NewExecutedFlow("not-started-session")
	executedFlow.Cmds = []*model.ExecutedCmd{
		{Cmd: "task1", Result: model.ExecutedResultUnRun},
		{Cmd: "task2", Result: model.ExecutedResultUnRun},
	}

	screen := &memoryScreen{}
	cc.Screen = screen

	args := NewDumpFlowArgs().SetMonitorMode()
	DumpFlowEx(cc, env, flow, 0, args, executedFlow, true, nil)

	output := screen.GetOutput()

	if strings.Contains(output, "ETA:") {
		t.Errorf("monitor mode should NOT contain ETA when no commands completed, got:\n%s", output)
	}
}

func TestMonitorModeETAFormat(t *testing.T) {
	tree := createTestCmdTree()
	cc := createTestCli(&memoryScreen{}, tree)
	env := createTestEnv()

	for i := 0; i < 5; i++ {
		cmd := tree.AddSub("cmd" + string(rune('0'+i)))
		cmd.RegEmptyCmd("command")
	}

	flow := &model.ParsedCmds{}
	for i := 0; i < 5; i++ {
		flow.Cmds = append(flow.Cmds, newParsedCmd("cmd"+string(rune('0'+i)), tree.GetSub("cmd"+string(rune('0'+i)))))
	}

	executedFlow := model.NewExecutedFlow("format-test-session")
	executedFlow.Cmds = []*model.ExecutedCmd{
		{Cmd: "cmd0", Result: model.ExecutedResultSucceeded, StartTs: time.Now().Add(-10 * time.Second), FinishTs: time.Now().Add(-5 * time.Second)},
		{Cmd: "cmd1", Result: model.ExecutedResultSucceeded, StartTs: time.Now().Add(-5 * time.Second), FinishTs: time.Now().Add(-2 * time.Second)},
		{Cmd: "cmd2", Result: model.ExecutedResultIncompleted, StartTs: time.Now().Add(-1 * time.Second)},
		{Cmd: "cmd3", Result: model.ExecutedResultUnRun},
		{Cmd: "cmd4", Result: model.ExecutedResultUnRun},
	}

	screen := &memoryScreen{}
	cc.Screen = screen

	args := NewDumpFlowArgs().SetMonitorMode()
	DumpFlowEx(cc, env, flow, 0, args, executedFlow, true, nil)

	output := screen.GetOutput()

	if !strings.Contains(output, "ETA:") {
		t.Errorf("output should contain ETA, got:\n%s", output)
	}

	if !strings.Contains(output, "avg:") {
		t.Errorf("output should contain avg, got:\n%s", output)
	}

	if !strings.Contains(output, "3/5") {
		t.Errorf("output should contain progress 3/5, got:\n%s", output)
	}

	t.Logf("Output:\n%s", output)
}
