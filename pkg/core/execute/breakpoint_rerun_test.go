package execute

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/innerr/ticat/pkg/core/model"
)

type multiLayerFlowTest struct {
	name          string
	cmds          []string
	subFlows      map[int][]string
	breakPoints   []string
	actions       []string
	expectedFlow  []string
	expectedSkips []string
}

func createMultiLayerFlow(cmdCount int, subFlowDepth int) *model.ExecutedFlow {
	sessionName := fmt.Sprintf("multilayer-%d-%d", cmdCount, subFlowDepth)
	flow := model.NewExecutedFlow(sessionName)
	flow.Flow = fmt.Sprintf("multilayer-flow-%d", cmdCount)

	for i := 0; i < cmdCount; i++ {
		cmd := model.NewExecutedCmd(fmt.Sprintf("cmd%d", i))
		cmd.Result = model.ExecutedResultSucceeded
		cmd.StartTs = time.Now().Add(-time.Duration(cmdCount-i+1) * time.Second)
		cmd.FinishTs = time.Now().Add(-time.Duration(cmdCount-i) * time.Second)

		if subFlowDepth > 0 && i%3 == 0 {
			subFlow := model.NewExecutedFlow(fmt.Sprintf("subflow-%d", i))
			for j := 0; j < subFlowDepth; j++ {
				subCmd := model.NewExecutedCmd(fmt.Sprintf("sub%d-cmd%d", i, j))
				subCmd.Result = model.ExecutedResultSucceeded
				subFlow.Cmds = append(subFlow.Cmds, subCmd)
			}
			cmd.SubFlow = subFlow
		}

		flow.Cmds = append(flow.Cmds, cmd)
	}
	return flow
}

func createLargeMultiLayerFlow(totalCmds int, layers int) *model.ExecutedFlow {
	sessionName := fmt.Sprintf("large-flow-L%d-C%d", layers, totalCmds)
	flow := model.NewExecutedFlow(sessionName)
	flow.Flow = "large-multilayer-flow"

	var buildRecursive func(depth int, prefix string, count int) []*model.ExecutedCmd
	buildRecursive = func(depth int, prefix string, count int) []*model.ExecutedCmd {
		var cmds []*model.ExecutedCmd
		for i := 0; i < count; i++ {
			cmdName := fmt.Sprintf("%s-L%d-C%d", prefix, depth, i)
			cmd := model.NewExecutedCmd(cmdName)
			cmd.Result = model.ExecutedResultSucceeded
			cmd.StartTs = time.Now().Add(-time.Duration(depth*count+i+1) * time.Second)
			cmd.FinishTs = time.Now().Add(-time.Duration(depth*count+i) * time.Second)

			if depth < layers-1 && i%2 == 0 {
				subFlow := model.NewExecutedFlow(fmt.Sprintf("sub-%s", cmdName))
				subFlow.Cmds = buildRecursive(depth+1, cmdName, count/2+1)
				cmd.SubFlow = subFlow
			}
			cmds = append(cmds, cmd)
		}
		return cmds
	}

	flow.Cmds = buildRecursive(0, "root", totalCmds/layers+1)
	return flow
}

func TestRerunWithBreakPointStepOver(t *testing.T) {
	tests := []struct {
		name        string
		cmdCount    int
		breakAt     int
		actions     []string
		minExpected int
		maxExpected int
	}{
		{
			name:        "step_over_5_cmds",
			cmdCount:    10,
			breakAt:     3,
			actions:     []string{"d", "c", "c", "c", "c", "c", "c", "c"},
			minExpected: 5,
			maxExpected: 8,
		},
		{
			name:        "step_over_all",
			cmdCount:    5,
			breakAt:     0,
			actions:     []string{"d", "d", "d", "d", "d"},
			minExpected: 1,
			maxExpected: 5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hook := &mockTestingHook{actions: tc.actions}
			cc := newTestCliWithHook(hook)
			env := newTestEnv()

			tree := model.NewCmdTree(model.CmdTreeStrsForTest())

			var cmdNames []string
			for i := 0; i < tc.cmdCount; i++ {
				name := fmt.Sprintf("cmd%d", i)
				cmdNames = append(cmdNames, name)
				sub := tree.AddSub(name)
				sub.RegEmptyCmd(fmt.Sprintf("command %d", i))
			}

			if tc.breakAt >= 0 && tc.breakAt < tc.cmdCount {
				cc.BreakPoints.Befores[cmdNames[tc.breakAt]] = true
			}

			var triggeredBreaks []string
			for i := 0; i < tc.cmdCount; i++ {
				name := cmdNames[i]
				cmd := newTestParsedCmd(tree, name)

				var mask *model.ExecuteMask
				if i >= tc.breakAt {
					mask = model.NewExecuteMask(name)
				}

				bpa := tryWaitSecAndBreakBefore(cc, env, cmd, mask, i > tc.breakAt, i == tc.cmdCount-1, false, func() {})
				if bpa != BPAContinue || len(hook.recordReason) > len(triggeredBreaks) {
					triggeredBreaks = append(triggeredBreaks, name)
				}
			}

			if len(triggeredBreaks) < tc.minExpected || len(triggeredBreaks) > tc.maxExpected {
				t.Errorf("expected %d-%d breakpoint triggers, got %d", tc.minExpected, tc.maxExpected, len(triggeredBreaks))
			}
		})
	}
}

func TestRerunWithBreakPointSkip(t *testing.T) {
	tests := []struct {
		name          string
		cmdCount      int
		breakAt       int
		skipAt        int
		actions       []string
		expectedSkips int
	}{
		{
			name:          "skip_one_command",
			cmdCount:      5,
			breakAt:       0,
			skipAt:        1,
			actions:       []string{"c", "s", "c", "c"},
			expectedSkips: 1,
		},
		{
			name:          "skip_multiple",
			cmdCount:      8,
			breakAt:       0,
			skipAt:        2,
			actions:       []string{"c", "c", "s", "s", "c", "c"},
			expectedSkips: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hook := &mockTestingHook{actions: tc.actions}
			cc := newTestCliWithHook(hook)
			env := newTestEnv()

			tree := model.NewCmdTree(model.CmdTreeStrsForTest())

			var cmdNames []string
			for i := 0; i < tc.cmdCount; i++ {
				name := fmt.Sprintf("skipcmd%d", i)
				cmdNames = append(cmdNames, name)
				sub := tree.AddSub(name)
				sub.RegEmptyCmd(fmt.Sprintf("skip command %d", i))
			}

			cc.BreakPoints.Befores[cmdNames[tc.breakAt]] = true

			skipCount := 0
			actionIdx := 0
			for i := 0; i < tc.cmdCount; i++ {
				name := cmdNames[i]
				cmd := newTestParsedCmd(tree, name)

				mask := model.NewExecuteMask(name)
				breakByPrev := i > tc.breakAt && actionIdx > 0

				bpa := tryWaitSecAndBreakBefore(cc, env, cmd, mask, breakByPrev, i == tc.cmdCount-1, false, func() {})

				if bpa == BPASkip {
					skipCount++
				}
				if len(hook.recordReason) > actionIdx {
					actionIdx = len(hook.recordReason)
				}
			}

			if skipCount != tc.expectedSkips {
				t.Errorf("expected %d skips, got %d", tc.expectedSkips, skipCount)
			}
		})
	}
}

func TestRerunWithBreakPointStepIn(t *testing.T) {
	tests := []struct {
		name            string
		hasSubFlow      bool
		actions         []string
		expectStepInEnv bool
	}{
		{
			name:            "step_in_with_subflow",
			hasSubFlow:      true,
			actions:         []string{"t", "c"},
			expectStepInEnv: true,
		},
		{
			name:            "continue_without_step_in",
			hasSubFlow:      true,
			actions:         []string{"c"},
			expectStepInEnv: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hook := &mockTestingHook{actions: tc.actions}
			cc := newTestCliWithHook(hook)
			env := newTestEnv()

			tree := model.NewCmdTree(model.CmdTreeStrsForTest())
			sub := tree.AddSub("flowcmd")
			sub.RegFlowCmd([]string{"sub1", ":", "sub2"}, "flow command with subflow", "")

			cmd := model.ParsedCmd{
				Segments: []model.ParsedCmdSeg{
					{
						Matched: model.MatchedCmd{
							Name: "flowcmd",
							Cmd:  sub,
						},
					},
				},
			}

			cc.BreakPoints.Befores["flowcmd"] = true

			mask := model.NewExecuteMask("flowcmd")
			if tc.hasSubFlow {
				mask.SubFlow = []*model.ExecuteMask{}
			}

			bpa := tryWaitSecAndBreakBefore(cc, env, cmd, mask, false, false, false, func() {})

			if bpa != BPAContinue {
				t.Errorf("expected BPAContinue, got %s", bpa)
			}

			hasStepIn := env.GetBool("sys.breakpoint.status.step-in")
			if hasStepIn != tc.expectStepInEnv {
				t.Errorf("expected step-in env %v, got %v", tc.expectStepInEnv, hasStepIn)
			}
		})
	}
}

func TestRerunMultiLayerFlowWithBreakPoints(t *testing.T) {
	tests := []struct {
		name           string
		layers         int
		cmdsPerLayer   int
		breakAtLayers  []int
		actions        []string
		expectedBreaks int
	}{
		{
			name:           "two_layers_break_at_each",
			layers:         2,
			cmdsPerLayer:   3,
			breakAtLayers:  []int{0, 1},
			actions:        []string{"c", "c", "c", "c"},
			expectedBreaks: 2,
		},
		{
			name:           "three_layers_break_at_first",
			layers:         3,
			cmdsPerLayer:   2,
			breakAtLayers:  []int{0},
			actions:        []string{"c", "c", "c", "c", "c", "c"},
			expectedBreaks: 1,
		},
		{
			name:           "four_layers_step_over",
			layers:         4,
			cmdsPerLayer:   2,
			breakAtLayers:  []int{0},
			actions:        []string{"d", "c", "c", "c"},
			expectedBreaks: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hook := &mockTestingHook{actions: tc.actions}
			cc := newTestCliWithHook(hook)
			env := newTestEnv()

			tree := model.NewCmdTree(model.CmdTreeStrsForTest())

			var allCmdNames []string
			for layer := 0; layer < tc.layers; layer++ {
				for cmd := 0; cmd < tc.cmdsPerLayer; cmd++ {
					name := fmt.Sprintf("L%dC%d", layer, cmd)
					allCmdNames = append(allCmdNames, name)
					sub := tree.AddSub(name)
					sub.RegEmptyCmd(fmt.Sprintf("layer %d cmd %d", layer, cmd))
				}
			}

			for _, layer := range tc.breakAtLayers {
				if layer*tc.cmdsPerLayer < len(allCmdNames) {
					cc.BreakPoints.Befores[allCmdNames[layer*tc.cmdsPerLayer]] = true
				}
			}

			breakCount := 0
			for i, name := range allCmdNames {
				cmd := newTestParsedCmd(tree, name)
				mask := model.NewExecuteMask(name)

				bpa := tryWaitSecAndBreakBefore(cc, env, cmd, mask, false, i == len(allCmdNames)-1, false, func() {})

				if bpa == BPAStepOver {
					mask.SetExecPolicyForAll(model.ExecPolicyExec)
				}

				if len(hook.recordReason) > breakCount {
					breakCount = len(hook.recordReason)
				}
			}

			if breakCount != tc.expectedBreaks {
				t.Errorf("expected %d breakpoint triggers, got %d", tc.expectedBreaks, breakCount)
			}
		})
	}
}

func TestRerunLargeFlowWithBreakPoints(t *testing.T) {
	tests := []struct {
		name           string
		totalCmds      int
		breakEvery     int
		actions        []string
		expectedBreaks int
	}{
		{
			name:           "twenty_cmds_break_every_5",
			totalCmds:      20,
			breakEvery:     5,
			actions:        []string{"c", "c", "c", "c", "c", "c", "c", "c"},
			expectedBreaks: 4,
		},
		{
			name:           "fifty_cmds_break_at_start_step_over",
			totalCmds:      50,
			breakEvery:     0,
			actions:        []string{"d"},
			expectedBreaks: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hook := &mockTestingHook{actions: tc.actions}
			cc := newTestCliWithHook(hook)
			env := newTestEnv()

			tree := model.NewCmdTree(model.CmdTreeStrsForTest())

			var cmdNames []string
			for i := 0; i < tc.totalCmds; i++ {
				name := fmt.Sprintf("largecmd%d", i)
				cmdNames = append(cmdNames, name)
				sub := tree.AddSub(name)
				sub.RegEmptyCmd(fmt.Sprintf("large command %d", i))
			}

			if tc.breakEvery > 0 {
				for i := 0; i < tc.totalCmds; i += tc.breakEvery {
					cc.BreakPoints.Befores[cmdNames[i]] = true
				}
			} else {
				cc.BreakPoints.Befores[cmdNames[0]] = true
			}

			breakCount := 0
			for i, name := range cmdNames {
				cmd := newTestParsedCmd(tree, name)

				var mask *model.ExecuteMask
				if i == 0 || (tc.breakEvery > 0 && i%tc.breakEvery == 0) {
					mask = model.NewExecuteMask(name)
				}

				bpa := tryWaitSecAndBreakBefore(cc, env, cmd, mask, false, i == len(cmdNames)-1, false, func() {})

				if bpa == BPAStepOver && mask != nil {
					mask.SetExecPolicyForAll(model.ExecPolicyExec)
				}

				if len(hook.recordReason) > breakCount {
					breakCount = len(hook.recordReason)
				}
			}

			if breakCount != tc.expectedBreaks {
				t.Errorf("expected %d breakpoint triggers, got %d", tc.expectedBreaks, breakCount)
			}
		})
	}
}

func TestRerunWithBreakPointQuit(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"q"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("quitcmd%d", i)
		sub := tree.AddSub(name)
		sub.RegEmptyCmd(fmt.Sprintf("quit test command %d", i))
	}

	cc.BreakPoints.Befores["quitcmd0"] = true

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for BPAQuit")
			return
		}
		err, ok := r.(error)
		if !ok {
			t.Errorf("expected error type, got %T", r)
			return
		}
		if !strings.Contains(err.Error(), "abort") {
			t.Errorf("expected abort error, got %v", err)
		}
	}()

	name := "quitcmd0"
	cmd := newTestParsedCmd(tree, name)
	tryBreakBefore(cc, env, cmd, nil, false, func() {})
}

func TestRerunWithBreakPointInteractChoice(t *testing.T) {
	t.Run("interact_choice_available", func(t *testing.T) {
		hook := &mockTestingHook{actions: []string{"c"}}
		cc := newTestCliWithHook(hook)
		env := newTestEnv()

		tree := model.NewCmdTree(model.CmdTreeStrsForTest())

		sub := tree.AddSub("interactcmd")
		sub.RegFlowCmd([]string{"sub1", ":", "sub2"}, "interact test command", "")

		cmd := model.ParsedCmd{
			Segments: []model.ParsedCmdSeg{
				{Matched: model.MatchedCmd{Name: "interactcmd", Cmd: sub}},
			},
		}

		cc.BreakPoints.Befores["interactcmd"] = true

		mask := model.NewExecuteMask("interactcmd")
		mask.SubFlow = []*model.ExecuteMask{}

		bpa := tryBreakBefore(cc, env, cmd, mask, false, func() {})

		if bpa != BPAContinue {
			t.Errorf("expected BPAContinue, got %s", bpa)
		}

		if len(hook.recordReason) == 0 {
			t.Error("expected breakpoint to trigger")
		}
	})

	t.Run("continue_after_interact_choice_displayed", func(t *testing.T) {
		hook := &mockTestingHook{actions: []string{"c", "c", "c"}}
		cc := newTestCliWithHook(hook)
		env := newTestEnv()

		tree := model.NewCmdTree(model.CmdTreeStrsForTest())

		for i := 0; i < 3; i++ {
			name := fmt.Sprintf("interactcmd%d", i)
			sub := tree.AddSub(name)
			sub.RegEmptyCmd(fmt.Sprintf("interact test command %d", i))
		}

		cc.BreakPoints.Befores["interactcmd0"] = true
		cc.BreakPoints.Befores["interactcmd1"] = true

		executedCmds := 0
		for i := 0; i < 3; i++ {
			name := fmt.Sprintf("interactcmd%d", i)
			cmd := newTestParsedCmd(tree, name)

			bpa := tryBreakBefore(cc, env, cmd, nil, false, func() {})

			if bpa == BPAContinue {
				executedCmds++
			}
		}

		if executedCmds != 3 {
			t.Errorf("expected 3 commands executed, got %d", executedCmds)
		}
	})
}

func TestRerunWithBreakPointAfter(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"d", "c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	for i := 0; i < 4; i++ {
		name := fmt.Sprintf("aftercmd%d", i)
		sub := tree.AddSub(name)
		sub.RegEmptyCmd(fmt.Sprintf("after test command %d", i))
	}

	cc.BreakPoints.Afters["aftercmd0"] = true
	cc.BreakPoints.Afters["aftercmd2"] = true

	afterBreaks := 0
	for i := 0; i < 4; i++ {
		name := fmt.Sprintf("aftercmd%d", i)
		cmd := newTestParsedCmd(tree, name)

		bpa := tryBreakAfter(cc, env, cmd, func() {})

		if bpa == BPAStepOver {
			afterBreaks++
		}
	}

	if afterBreaks != 1 {
		t.Errorf("expected 1 after-break with step-over, got %d", afterBreaks)
	}
}

func TestRerunComplexScenarioWithBreakPoints(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"c", "s", "t", "c", "d", "c", "c", "c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	setupCmd := tree.AddSub("setup")
	setupCmd.RegEmptyCmd("setup command")

	deployCmd := tree.AddSub("deploy")
	deployCmd.RegFlowCmd([]string{"sub1", ":", "sub2"}, "deploy with subflow", "")

	testCmd := tree.AddSub("test")
	testCmd.RegEmptyCmd("test command")

	cleanupCmd := tree.AddSub("cleanup")
	cleanupCmd.RegEmptyCmd("cleanup command")

	cmds := []model.ParsedCmd{
		newTestParsedCmd(tree, "setup"),
		{
			Segments: []model.ParsedCmdSeg{
				{Matched: model.MatchedCmd{Name: "deploy", Cmd: deployCmd}},
			},
		},
		newTestParsedCmd(tree, "test"),
		newTestParsedCmd(tree, "cleanup"),
	}

	cc.BreakPoints.Befores["setup"] = true
	cc.BreakPoints.Befores["deploy"] = true
	cc.BreakPoints.Befores["test"] = true

	var results []string
	for i, cmd := range cmds {
		name := cmd.DisplayPath(".", false)
		mask := model.NewExecuteMask(name)

		if name == "deploy" {
			mask.SubFlow = []*model.ExecuteMask{}
		}

		bpa := tryWaitSecAndBreakBefore(cc, env, cmd, mask, false, i == len(cmds)-1, false, func() {})

		results = append(results, fmt.Sprintf("%s:%s", name, string(bpa)))

		if bpa == BPASkip {
			env.GetLayer(model.EnvLayerSession).SetBool("sys.breakpoint.status.step-out", true)
		}
	}

	expectedResults := []string{
		"setup:continue",
		"deploy:continue",
		"test:continue",
		"cleanup:continue",
	}

	for i, expected := range expectedResults {
		if i < len(results) && !strings.HasPrefix(results[i], strings.Split(expected, ":")[0]) {
			t.Errorf("result[%d]: expected prefix %s, got %s", i, strings.Split(expected, ":")[0], results[i])
		}
	}
}

func TestRerunHundredCommandsWithBreakPoints(t *testing.T) {
	actions := make([]string, 10)
	for i := range actions {
		actions[i] = "c"
	}

	hook := &mockTestingHook{actions: actions}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	var cmdNames []string
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("hundred%d", i)
		cmdNames = append(cmdNames, name)
		sub := tree.AddSub(name)
		sub.RegEmptyCmd(fmt.Sprintf("hundred command %d", i))
	}

	for i := 0; i < 100; i += 10 {
		cc.BreakPoints.Befores[cmdNames[i]] = true
	}

	breakCount := 0
	for i, name := range cmdNames {
		cmd := newTestParsedCmd(tree, name)

		var mask *model.ExecuteMask
		if i%10 == 0 {
			mask = model.NewExecuteMask(name)
		}

		tryWaitSecAndBreakBefore(cc, env, cmd, mask, false, i == len(cmdNames)-1, false, func() {})

		if len(hook.recordReason) > breakCount {
			breakCount = len(hook.recordReason)
		}
	}

	if breakCount != 10 {
		t.Errorf("expected 10 breakpoint triggers, got %d", breakCount)
	}
}

func TestRerunNestedFlowWithBreakPoints(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"c", "t", "c", "c", "c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	level0 := tree.AddSub("level0")
	level0.RegFlowCmd([]string{"sub1", ":", "sub2"}, "level 0 with subflow", "")

	level1 := tree.AddSub("level1")
	level1.RegFlowCmd([]string{"sub3", ":", "sub4"}, "level 1 with subflow", "")

	level2 := tree.AddSub("level2")
	level2.RegEmptyCmd("level 2 command")

	cmds := []model.ParsedCmd{
		{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "level0", Cmd: level0}}}},
		{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "level1", Cmd: level1}}}},
		{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "level2", Cmd: level2}}}},
	}

	cc.BreakPoints.Befores["level0"] = true
	cc.BreakPoints.Befores["level1"] = true

	for i, cmd := range cmds {
		name := cmd.DisplayPath(".", false)
		mask := model.NewExecuteMask(name)

		if cmd.LastCmd() != nil && cmd.LastCmd().HasSubFlow(false) {
			mask.SubFlow = []*model.ExecuteMask{}
		}

		bpa := tryWaitSecAndBreakBefore(cc, env, cmd, mask, false, i == len(cmds)-1, false, func() {})

		if bpa == BPAStepOver && mask != nil {
			mask.SetExecPolicyForAll(model.ExecPolicyExec)
		}
	}

	if hook.actionIndex < 3 {
		t.Errorf("expected at least 3 actions consumed, got %d", hook.actionIndex)
	}
}

func TestRerunWithStepInOutSequence(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"t", "c", "c", "c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	flowCmd := tree.AddSub("flowcmd")
	flowCmd.RegFlowCmd([]string{"sub1", ":", "sub2"}, "flow command", "")

	normalCmd := tree.AddSub("normalcmd")
	normalCmd.RegEmptyCmd("normal command")

	cmds := []model.ParsedCmd{
		{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "flowcmd", Cmd: flowCmd}}}},
		{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "normalcmd", Cmd: normalCmd}}}},
	}

	cc.BreakPoints.Befores["flowcmd"] = true

	stepInTriggered := false
	for i, cmd := range cmds {
		name := cmd.DisplayPath(".", false)
		mask := model.NewExecuteMask(name)

		if name == "flowcmd" {
			mask.SubFlow = []*model.ExecuteMask{}
		}

		bpa := tryWaitSecAndBreakBefore(cc, env, cmd, mask, false, i == len(cmds)-1, false, func() {})

		if bpa == BPAStepOver {
			mask.SetExecPolicyForAll(model.ExecPolicyExec)
		}

		if env.GetBool("sys.breakpoint.status.step-in") {
			stepInTriggered = true
		}
	}

	if !stepInTriggered {
		t.Error("expected step-in to be triggered")
	}
}

func TestRerunWithAllBreakPointActions(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"s", "d", "t", "c", "c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	skipCmd := tree.AddSub("skipcmd")
	skipCmd.RegEmptyCmd("skip command")

	stepOverCmd := tree.AddSub("stepovercmd")
	stepOverCmd.RegEmptyCmd("step over command")

	stepInCmd := tree.AddSub("stepincmd")
	stepInCmd.RegFlowCmd([]string{"sub1", ":", "sub2"}, "step in command", "")

	continueCmd := tree.AddSub("continuecmd")
	continueCmd.RegEmptyCmd("continue command")

	cmds := []model.ParsedCmd{
		newTestParsedCmd(tree, "skipcmd"),
		newTestParsedCmd(tree, "stepovercmd"),
		{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "stepincmd", Cmd: stepInCmd}}}},
		newTestParsedCmd(tree, "continuecmd"),
	}

	for _, name := range []string{"skipcmd", "stepovercmd", "stepincmd"} {
		cc.BreakPoints.Befores[name] = true
	}

	var bpas []BreakPointAction
	for i, cmd := range cmds {
		name := cmd.DisplayPath(".", false)
		mask := model.NewExecuteMask(name)

		if name == "stepincmd" {
			mask.SubFlow = []*model.ExecuteMask{}
		}

		bpa := tryWaitSecAndBreakBefore(cc, env, cmd, mask, false, i == len(cmds)-1, false, func() {})
		bpas = append(bpas, bpa)

		if bpa == BPASkip && i == len(cmds)-1 {
			env.GetLayer(model.EnvLayerSession).SetBool("sys.breakpoint.status.step-out", true)
		}
	}

	expectedBpas := []BreakPointAction{BPASkip, BPAStepOver, BPAContinue, BPAContinue}

	for i, expected := range expectedBpas {
		if i < len(bpas) && bpas[i] != expected {
			t.Errorf("cmd %d: expected %s, got %s", i, expected, bpas[i])
		}
	}
}
