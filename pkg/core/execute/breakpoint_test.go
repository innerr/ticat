package execute

import (
	"strings"
	"testing"

	"github.com/innerr/ticat/pkg/core/model"
)

type mockTestingHook struct {
	actions      []string
	actionIndex  int
	promptLines  []string
	promptIndex  int
	skipBash     bool
	recordReason []string
}

func (h *mockTestingHook) OnBreakPoint(reason string, choices []string, actions map[string]string) string {
	h.recordReason = append(h.recordReason, reason)
	if h.actionIndex < len(h.actions) {
		action := h.actions[h.actionIndex]
		h.actionIndex++
		return action
	}
	return "c"
}

func (h *mockTestingHook) OnInteractPrompt(prompt string) (string, bool) {
	if h.promptIndex < len(h.promptLines) {
		line := h.promptLines[h.promptIndex]
		h.promptIndex++
		return line, true
	}
	return "", false
}

func (h *mockTestingHook) ShouldSkipBash() bool {
	return h.skipBash
}

func newTestCliWithHook(hook model.TestingHook) *model.Cli {
	strs := model.CmdTreeStrsForTest()
	tree := model.NewCmdTree(strs)
	return &model.Cli{
		Screen:        &model.QuietScreen{},
		Cmds:          tree,
		BreakPoints:   model.NewBreakPoints(),
		TestingHook:   hook,
		EnvAbbrs:      model.NewEnvAbbrs("test"),
		TolerableErrs: model.NewTolerableErrs(),
	}
}

func newTestEnv() *model.Env {
	env := model.NewEnv()
	return env.NewLayer(model.EnvLayerSession)
}

func newTestParsedCmd(tree *model.CmdTree, name string) model.ParsedCmd {
	sub := tree.GetOrAddSub(name)
	sub.RegEmptyCmd("test command")
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

func TestBreakPointActionContinue(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd := newTestParsedCmd(tree, "testcmd")
	cc.BreakPoints.Befores["testcmd"] = true

	bpa := tryBreakBefore(cc, env, cmd, nil, false, func() {})

	if bpa != BPAContinue {
		t.Errorf("expected BPAContinue, got %s", bpa)
	}
	if len(hook.recordReason) != 1 {
		t.Errorf("expected 1 breakpoint trigger, got %d", len(hook.recordReason))
	}
	if !strings.Contains(hook.recordReason[0], "testcmd") {
		t.Errorf("reason should contain 'testcmd', got %s", hook.recordReason[0])
	}
}

func TestBreakPointActionSkip(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"s"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd := newTestParsedCmd(tree, "testcmd")
	cc.BreakPoints.Befores["testcmd"] = true

	bpa := tryBreakBefore(cc, env, cmd, nil, false, func() {})

	if bpa != BPASkip {
		t.Errorf("expected BPASkip, got %s", bpa)
	}
}

func TestBreakPointActionStepOver(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"d"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd := newTestParsedCmd(tree, "testcmd")
	cc.BreakPoints.Befores["testcmd"] = true

	mask := model.NewExecuteMask("testcmd")
	bpa := tryWaitSecAndBreakBefore(cc, env, cmd, mask, false, false, false, func() {})

	if bpa != BPAStepOver {
		t.Errorf("expected BPAStepOver, got %s", bpa)
	}
	if mask.ExecPolicy != model.ExecPolicyExec {
		t.Errorf("mask ExecPolicy should be ExecPolicyExec for StepOver")
	}
}

func TestBreakPointActionStepIn(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"t"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	sub := tree.AddSub("testcmd")
	sub.RegFlowCmd([]string{"subcmd1", ":", "subcmd2"}, "test command with subflow", "")

	cmd := model.ParsedCmd{
		Segments: []model.ParsedCmdSeg{
			{
				Matched: model.MatchedCmd{
					Name: "testcmd",
					Cmd:  sub,
				},
			},
		},
	}
	cc.BreakPoints.Befores["testcmd"] = true

	mask := model.NewExecuteMask("testcmd")
	mask.SubFlow = []*model.ExecuteMask{}

	bpa := tryWaitSecAndBreakBefore(cc, env, cmd, mask, false, false, false, func() {})

	if bpa != BPAContinue {
		t.Errorf("expected BPAContinue after step-in conversion, got %s", bpa)
	}
	if !env.GetBool("sys.breakpoint.status.step-in") {
		t.Error("step-in status should be set after StepIn action")
	}
}

func TestBreakPointActionQuit(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"q"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd := newTestParsedCmd(tree, "testcmd")
	cc.BreakPoints.Befores["testcmd"] = true

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

	tryBreakBefore(cc, env, cmd, nil, false, func() {})
}

func TestBreakPointActionInteract(t *testing.T) {
	t.Run("interact action requires executor", func(t *testing.T) {
		hook := &mockTestingHook{
			actions:     []string{"i", "c"},
			promptLines: []string{"dbg.interact.leave"},
		}
		cc := newTestCliWithHook(hook)
		env := newTestEnv()

		tree := model.NewCmdTree(model.CmdTreeStrsForTest())
		cmd := newTestParsedCmd(tree, "testcmd")
		cc.BreakPoints.Befores["testcmd"] = true

		defer func() {
			r := recover()
			if r == nil {
				t.Error("expected panic when executor is nil")
			}
		}()

		tryBreakBefore(cc, env, cmd, nil, false, func() {})
	})
}

func TestBreakPointAfter(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd := newTestParsedCmd(tree, "testcmd")
	cc.BreakPoints.Afters["testcmd"] = true

	bpa := tryBreakAfter(cc, env, cmd, func() {})

	if bpa != BPAContinue {
		t.Errorf("expected BPAContinue, got %s", bpa)
	}
}

func TestBreakPointAtEnd(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	cc.BreakPoints.SetAtEnd(true)

	tryBreakAtEnd(cc, env)

	if len(hook.recordReason) != 1 {
		t.Errorf("expected 1 breakpoint trigger at end, got %d", len(hook.recordReason))
	}
}

func TestBreakPointByEnv(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	env.SetBool("sys.breakpoint.here.now", true)

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd := newTestParsedCmd(tree, "testcmd")

	bpa := tryBreakBefore(cc, env, cmd, nil, false, func() {})

	if bpa != BPAContinue {
		t.Errorf("expected BPAContinue, got %s", bpa)
	}
	if env.GetBool("sys.breakpoint.here.now") {
		t.Error("sys.breakpoint.here.now should be deleted after trigger")
	}
}

func TestBreakPointAtNext(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	env.SetBool("sys.breakpoint.at-next", true)

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd := newTestParsedCmd(tree, "testcmd")

	bpa := tryBreakBefore(cc, env, cmd, nil, false, func() {})

	if bpa != BPAContinue {
		t.Errorf("expected BPAContinue, got %s", bpa)
	}
}

func TestStepInStatus(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	env.SetBool("sys.breakpoint.status.step-in", true)

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd := newTestParsedCmd(tree, "testcmd")

	bpa := tryBreakBefore(cc, env, cmd, nil, false, func() {})

	if bpa != BPAContinue {
		t.Errorf("expected BPAContinue, got %s", bpa)
	}
	if env.GetBool("sys.breakpoint.status.step-in") {
		t.Error("step-in status should be deleted after trigger")
	}
}

func TestStepOutStatus(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	env.SetBool("sys.breakpoint.status.step-out", true)

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd := newTestParsedCmd(tree, "testcmd")

	bpa := tryBreakBefore(cc, env, cmd, nil, false, func() {})

	if bpa != BPAContinue {
		t.Errorf("expected BPAContinue, got %s", bpa)
	}
}

func TestBreakPointInsideInteract(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	env.SetBool("sys.interact.inside", true)

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd := newTestParsedCmd(tree, "testcmd")
	cc.BreakPoints.Befores["testcmd"] = true

	bpa := tryWaitSecAndBreakBefore(cc, env, cmd, nil, false, false, false, func() {})

	if bpa != BPAContinue {
		t.Errorf("expected BPAContinue when inside interact, got %s", bpa)
	}
	if len(hook.recordReason) != 0 {
		t.Errorf("should not trigger breakpoint when inside interact, got %d triggers", len(hook.recordReason))
	}
}

func TestBreakPointQuietCommand(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	sub := tree.GetOrAddSub("quietcmd")
	sub.RegEmptyCmd("quiet command").SetQuiet()

	cmd := model.ParsedCmd{
		Segments: []model.ParsedCmdSeg{
			{
				Matched: model.MatchedCmd{
					Name: "quietcmd",
					Cmd:  sub,
				},
			},
		},
	}
	cc.BreakPoints.Befores["quietcmd"] = true

	bpa := tryWaitSecAndBreakBefore(cc, env, cmd, nil, false, false, false, func() {})

	if bpa != BPAContinue {
		t.Errorf("expected BPAContinue for quiet command, got %s", bpa)
	}
	if len(hook.recordReason) != 0 {
		t.Errorf("should not trigger breakpoint for quiet command, got %d triggers", len(hook.recordReason))
	}
}

func TestMultipleBreakPoints(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"c", "c", "c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd1 := newTestParsedCmd(tree, "cmd1")
	cmd2 := newTestParsedCmd(tree, "cmd2")
	cmd3 := newTestParsedCmd(tree, "cmd3")

	cc.BreakPoints.Befores["cmd1"] = true
	cc.BreakPoints.Befores["cmd2"] = true
	cc.BreakPoints.Befores["cmd3"] = true

	bpa1 := tryBreakBefore(cc, env, cmd1, nil, false, func() {})
	bpa2 := tryBreakBefore(cc, env, cmd2, nil, false, func() {})
	bpa3 := tryBreakBefore(cc, env, cmd3, nil, false, func() {})

	if bpa1 != BPAContinue || bpa2 != BPAContinue || bpa3 != BPAContinue {
		t.Errorf("expected all BPAContinue, got %s, %s, %s", bpa1, bpa2, bpa3)
	}
	if len(hook.recordReason) != 3 {
		t.Errorf("expected 3 breakpoint triggers, got %d", len(hook.recordReason))
	}
	if hook.actionIndex != 3 {
		t.Errorf("expected 3 actions consumed, got %d", hook.actionIndex)
	}
}

func TestStepOverSequence(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"d", "c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd1 := newTestParsedCmd(tree, "cmd1")
	cmd2 := newTestParsedCmd(tree, "cmd2")

	cc.BreakPoints.Befores["cmd1"] = true

	mask1 := model.NewExecuteMask("cmd1")
	bpa1 := tryWaitSecAndBreakBefore(cc, env, cmd1, mask1, false, false, false, func() {})

	if bpa1 != BPAStepOver {
		t.Errorf("expected BPAStepOver for first command, got %s", bpa1)
	}

	env.SetBool("sys.breakpoint.at-next", true)

	mask2 := model.NewExecuteMask("cmd2")
	bpa2 := tryWaitSecAndBreakBefore(cc, env, cmd2, mask2, true, false, false, func() {})

	if bpa2 != BPAContinue {
		t.Errorf("expected BPAContinue for second command after step over, got %s", bpa2)
	}
}

func TestSkipLastCmdInFlow(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"s"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd := newTestParsedCmd(tree, "testcmd")
	cc.BreakPoints.Befores["testcmd"] = true

	bpa := tryWaitSecAndBreakBefore(cc, env, cmd, nil, false, true, false, func() {})

	if bpa != BPASkip {
		t.Errorf("expected BPASkip, got %s", bpa)
	}
	if !env.GetBool("sys.breakpoint.status.step-out") {
		t.Error("step-out status should be set when skipping last command in flow")
	}
}

func TestBreakPointChoices(t *testing.T) {
	tests := []struct {
		name           string
		hasSubFlow     bool
		expectedChoice string
	}{
		{"normal command", false, "s"},
		{"command with subflow", true, "t"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hook := &mockTestingHook{actions: []string{"c"}}
			cc := newTestCliWithHook(hook)
			env := newTestEnv()

			tree := model.NewCmdTree(model.CmdTreeStrsForTest())
			sub := tree.AddSub("testcmd")

			if tc.hasSubFlow {
				sub.RegFlowCmd([]string{"subcmd1", ":", "subcmd2"}, "test command with subflow", "")
			} else {
				sub.RegEmptyCmd("test command")
			}

			cmd := model.ParsedCmd{
				Segments: []model.ParsedCmdSeg{
					{
						Matched: model.MatchedCmd{
							Name: "testcmd",
							Cmd:  sub,
						},
					},
				},
			}
			cc.BreakPoints.Befores["testcmd"] = true

			mask := model.NewExecuteMask("testcmd")
			if tc.hasSubFlow {
				mask.SubFlow = []*model.ExecuteMask{}
			}

			tryBreakBefore(cc, env, cmd, mask, false, func() {})

			if tc.hasSubFlow && hook.actionIndex != 1 {
				t.Error("should have consumed one action for command with subflow")
			}
		})
	}
}

func TestBreakPointReasonMessages(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*model.Cli, *model.Env, *model.CmdTree) model.ParsedCmd
		containsStr string
	}{
		{
			name: "break before",
			setup: func(cc *model.Cli, env *model.Env, tree *model.CmdTree) model.ParsedCmd {
				cmd := newTestParsedCmd(tree, "testcmd")
				cc.BreakPoints.Befores["testcmd"] = true
				return cmd
			},
			containsStr: "break-point: before command",
		},
		{
			name: "step in",
			setup: func(cc *model.Cli, env *model.Env, tree *model.CmdTree) model.ParsedCmd {
				env.SetBool("sys.breakpoint.status.step-in", true)
				return newTestParsedCmd(tree, "testcmd")
			},
			containsStr: "just stepped in",
		},
		{
			name: "step out",
			setup: func(cc *model.Cli, env *model.Env, tree *model.CmdTree) model.ParsedCmd {
				env.SetBool("sys.breakpoint.status.step-out", true)
				return newTestParsedCmd(tree, "testcmd")
			},
			containsStr: "just stepped out",
		},
		{
			name: "break by prev",
			setup: func(cc *model.Cli, env *model.Env, tree *model.CmdTree) model.ParsedCmd {
				env.SetBool("sys.breakpoint.at-next", true)
				return newTestParsedCmd(tree, "testcmd")
			},
			containsStr: "previous choice",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hook := &mockTestingHook{actions: []string{"c"}}
			cc := newTestCliWithHook(hook)
			env := newTestEnv()
			tree := model.NewCmdTree(model.CmdTreeStrsForTest())

			cmd := tc.setup(cc, env, tree)

			tryBreakBefore(cc, env, cmd, nil, false, func() {})

			if len(hook.recordReason) == 0 {
				t.Fatal("no breakpoint reason recorded")
			}
			if !strings.Contains(hook.recordReason[0], tc.containsStr) {
				t.Errorf("reason should contain '%s', got '%s'", tc.containsStr, hook.recordReason[0])
			}
		})
	}
}

func TestContinueClearsAtNext(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	env.SetBool("sys.breakpoint.at-next", true)

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd := newTestParsedCmd(tree, "testcmd")
	cc.BreakPoints.Befores["testcmd"] = true

	tryBreakBefore(cc, env, cmd, nil, false, func() {})

	if env.GetBool("sys.breakpoint.at-next") {
		t.Error("sys.breakpoint.at-next should be cleared after continue")
	}
}

func TestInvalidAction(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"x"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd := newTestParsedCmd(tree, "testcmd")
	cc.BreakPoints.Befores["testcmd"] = true

	bpa := tryBreakBefore(cc, env, cmd, nil, false, func() {})

	if bpa != BPAContinue {
		t.Errorf("expected BPAContinue for invalid input, got %s", bpa)
	}
}

func TestBreakPointAfterStepOver(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"d"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd := newTestParsedCmd(tree, "testcmd")
	cc.BreakPoints.Afters["testcmd"] = true

	bpa := tryBreakAfter(cc, env, cmd, func() {})

	if bpa != BPAStepOver {
		t.Errorf("expected BPAStepOver, got %s", bpa)
	}
}

func TestClearBreakPointStatusInEnv(t *testing.T) {
	env := newTestEnv()
	env.SetBool("sys.interact.leaving", true)
	env.SetBool("sys.breakpoint.status.step-in", true)
	env.SetBool("sys.breakpoint.status.step-out", true)

	clearBreakPointStatusInEnv(env)

	if env.GetBool("sys.interact.leaving") {
		t.Error("sys.interact.leaving should be cleared")
	}
	if env.GetBool("sys.breakpoint.status.step-in") {
		t.Error("sys.breakpoint.status.step-in should be cleared")
	}
	if env.GetBool("sys.breakpoint.status.step-out") {
		t.Error("sys.breakpoint.status.step-out should be cleared")
	}
}

func TestGetAllBPAs(t *testing.T) {
	bpas := getAllBPAs()

	expectedActions := map[string]BreakPointAction{
		"c": BPAContinue,
		"s": BPASkip,
		"q": BPAQuit,
		"t": BPAStepIn,
		"d": BPAStepOver,
		"i": BPAInteract,
	}

	for key, expected := range expectedActions {
		if bpas[key] != expected {
			t.Errorf("expected bpas[%s] = %s, got %s", key, expected, bpas[key])
		}
	}

	if len(bpas) != len(expectedActions) {
		t.Errorf("expected %d actions, got %d", len(expectedActions), len(bpas))
	}
}

func TestBreakPointNoTrigger(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd := newTestParsedCmd(tree, "testcmd")

	bpa := tryBreakBefore(cc, env, cmd, nil, false, func() {})

	if bpa != BPAContinue {
		t.Errorf("expected BPAContinue when no breakpoint set, got %s", bpa)
	}
	if len(hook.recordReason) != 0 {
		t.Errorf("should not trigger when no breakpoint, got %d triggers", len(hook.recordReason))
	}
}

func TestBreakPointEmptyChoices(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	bpa := readUserBPAChoice("test reason", []string{"c"}, BPAs{"c": BPAContinue}, true, cc, env, func() {})

	if bpa != BPAContinue {
		t.Errorf("expected BPAContinue, got %s", bpa)
	}
}

func TestBreakPointActionCombinations(t *testing.T) {
	t.Run("continue then skip", func(t *testing.T) {
		hook := &mockTestingHook{actions: []string{"c", "s"}}
		cc := newTestCliWithHook(hook)
		env := newTestEnv()

		tree := model.NewCmdTree(model.CmdTreeStrsForTest())
		cmd1 := newTestParsedCmd(tree, "cmd1")
		cmd2 := newTestParsedCmd(tree, "cmd2")

		cc.BreakPoints.Befores["cmd1"] = true
		cc.BreakPoints.Befores["cmd2"] = true

		bpa1 := tryBreakBefore(cc, env, cmd1, nil, false, func() {})
		bpa2 := tryBreakBefore(cc, env, cmd2, nil, false, func() {})

		if bpa1 != BPAContinue {
			t.Errorf("expected BPAContinue for cmd1, got %s", bpa1)
		}
		if bpa2 != BPASkip {
			t.Errorf("expected BPASkip for cmd2, got %s", bpa2)
		}
	})

	t.Run("step over then continue", func(t *testing.T) {
		hook := &mockTestingHook{actions: []string{"d", "c"}}
		cc := newTestCliWithHook(hook)
		env := newTestEnv()

		tree := model.NewCmdTree(model.CmdTreeStrsForTest())
		cmd1 := newTestParsedCmd(tree, "cmd1")
		cmd2 := newTestParsedCmd(tree, "cmd2")

		cc.BreakPoints.Befores["cmd1"] = true

		mask1 := model.NewExecuteMask("cmd1")
		bpa1 := tryWaitSecAndBreakBefore(cc, env, cmd1, mask1, false, false, false, func() {})

		env.SetBool("sys.breakpoint.at-next", true)

		mask2 := model.NewExecuteMask("cmd2")
		bpa2 := tryWaitSecAndBreakBefore(cc, env, cmd2, mask2, true, false, false, func() {})

		if bpa1 != BPAStepOver {
			t.Errorf("expected BPAStepOver for cmd1, got %s", bpa1)
		}
		if bpa2 != BPAContinue {
			t.Errorf("expected BPAContinue for cmd2, got %s", bpa2)
		}
	})

	t.Run("step in sets env status", func(t *testing.T) {
		hook := &mockTestingHook{actions: []string{"t"}}
		cc := newTestCliWithHook(hook)
		env := newTestEnv()

		tree := model.NewCmdTree(model.CmdTreeStrsForTest())
		sub := tree.AddSub("flowcmd")
		sub.RegFlowCmd([]string{"sub1", ":", "sub2"}, "flow command", "")

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
		mask.SubFlow = []*model.ExecuteMask{}

		bpa := tryWaitSecAndBreakBefore(cc, env, cmd, mask, false, false, false, func() {})

		if bpa != BPAContinue {
			t.Errorf("expected BPAContinue after step-in conversion, got %s", bpa)
		}
		if !env.GetBool("sys.breakpoint.status.step-in") {
			t.Error("sys.breakpoint.status.step-in should be set")
		}
	})

	t.Run("quit action panics with abort error", func(t *testing.T) {
		hook := &mockTestingHook{actions: []string{"q"}}
		cc := newTestCliWithHook(hook)
		env := newTestEnv()

		tree := model.NewCmdTree(model.CmdTreeStrsForTest())
		cmd := newTestParsedCmd(tree, "testcmd")
		cc.BreakPoints.Befores["testcmd"] = true

		defer func() {
			r := recover()
			if r == nil {
				t.Error("expected panic for BPAQuit")
				return
			}
		}()

		tryBreakBefore(cc, env, cmd, nil, false, func() {})
	})
}

func TestBreakPointEnvironmentTriggers(t *testing.T) {
	t.Run("sys.breakpoint.here.now triggers break", func(t *testing.T) {
		hook := &mockTestingHook{actions: []string{"c"}}
		cc := newTestCliWithHook(hook)
		env := newTestEnv()

		env.SetBool("sys.breakpoint.here.now", true)

		tree := model.NewCmdTree(model.CmdTreeStrsForTest())
		cmd := newTestParsedCmd(tree, "testcmd")

		bpa := tryBreakBefore(cc, env, cmd, nil, false, func() {})

		if bpa != BPAContinue {
			t.Errorf("expected BPAContinue, got %s", bpa)
		}
		if env.GetBool("sys.breakpoint.here.now") {
			t.Error("sys.breakpoint.here.now should be cleared after trigger")
		}
		if len(hook.recordReason) == 0 {
			t.Error("should have triggered breakpoint")
		}
	})

	t.Run("sys.breakpoint.at-next triggers break", func(t *testing.T) {
		hook := &mockTestingHook{actions: []string{"c"}}
		cc := newTestCliWithHook(hook)
		env := newTestEnv()

		env.SetBool("sys.breakpoint.at-next", true)

		tree := model.NewCmdTree(model.CmdTreeStrsForTest())
		cmd := newTestParsedCmd(tree, "testcmd")

		bpa := tryBreakBefore(cc, env, cmd, nil, false, func() {})

		if bpa != BPAContinue {
			t.Errorf("expected BPAContinue, got %s", bpa)
		}
		if len(hook.recordReason) == 0 {
			t.Error("should have triggered breakpoint")
		}
	})

	t.Run("sys.breakpoint.status.step-in triggers break", func(t *testing.T) {
		hook := &mockTestingHook{actions: []string{"c"}}
		cc := newTestCliWithHook(hook)
		env := newTestEnv()

		env.SetBool("sys.breakpoint.status.step-in", true)

		tree := model.NewCmdTree(model.CmdTreeStrsForTest())
		cmd := newTestParsedCmd(tree, "testcmd")

		bpa := tryBreakBefore(cc, env, cmd, nil, false, func() {})

		if bpa != BPAContinue {
			t.Errorf("expected BPAContinue, got %s", bpa)
		}
		if len(hook.recordReason) == 0 {
			t.Error("should have triggered breakpoint")
		}
	})

	t.Run("sys.breakpoint.status.step-out triggers break", func(t *testing.T) {
		hook := &mockTestingHook{actions: []string{"c"}}
		cc := newTestCliWithHook(hook)
		env := newTestEnv()

		env.SetBool("sys.breakpoint.status.step-out", true)

		tree := model.NewCmdTree(model.CmdTreeStrsForTest())
		cmd := newTestParsedCmd(tree, "testcmd")

		bpa := tryBreakBefore(cc, env, cmd, nil, false, func() {})

		if bpa != BPAContinue {
			t.Errorf("expected BPAContinue, got %s", bpa)
		}
		if len(hook.recordReason) == 0 {
			t.Error("should have triggered breakpoint")
		}
	})
}

func TestBreakPointSkipBehavior(t *testing.T) {
	t.Run("skip sets step-out for last cmd in flow", func(t *testing.T) {
		hook := &mockTestingHook{actions: []string{"s"}}
		cc := newTestCliWithHook(hook)
		env := newTestEnv()

		tree := model.NewCmdTree(model.CmdTreeStrsForTest())
		cmd := newTestParsedCmd(tree, "testcmd")
		cc.BreakPoints.Befores["testcmd"] = true

		bpa := tryWaitSecAndBreakBefore(cc, env, cmd, nil, false, true, false, func() {})

		if bpa != BPASkip {
			t.Errorf("expected BPASkip, got %s", bpa)
		}
		if !env.GetBool("sys.breakpoint.status.step-out") {
			t.Error("sys.breakpoint.status.step-out should be set for last cmd in flow")
		}
	})

	t.Run("skip does not set step-out for non-last cmd", func(t *testing.T) {
		hook := &mockTestingHook{actions: []string{"s"}}
		cc := newTestCliWithHook(hook)
		env := newTestEnv()

		tree := model.NewCmdTree(model.CmdTreeStrsForTest())
		cmd := newTestParsedCmd(tree, "testcmd")
		cc.BreakPoints.Befores["testcmd"] = true

		bpa := tryWaitSecAndBreakBefore(cc, env, cmd, nil, false, false, false, func() {})

		if bpa != BPASkip {
			t.Errorf("expected BPASkip, got %s", bpa)
		}
		if env.GetBool("sys.breakpoint.status.step-out") {
			t.Error("sys.breakpoint.status.step-out should not be set for non-last cmd")
		}
	})
}

func TestBreakPointContinueClearsAtNext(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"c"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	env.SetBool("sys.breakpoint.at-next", true)

	tree := model.NewCmdTree(model.CmdTreeStrsForTest())
	cmd := newTestParsedCmd(tree, "testcmd")
	cc.BreakPoints.Befores["testcmd"] = true

	tryBreakBefore(cc, env, cmd, nil, false, func() {})

	if env.GetBool("sys.breakpoint.at-next") {
		t.Error("sys.breakpoint.at-next should be cleared after continue")
	}
}

func TestBreakPointLowerInput(t *testing.T) {
	hook := &mockTestingHook{actions: []string{"C"}}
	cc := newTestCliWithHook(hook)
	env := newTestEnv()

	bpa := readUserBPAChoice("test reason", []string{"c"}, BPAs{"c": BPAContinue}, true, cc, env, func() {})

	if bpa != BPAContinue {
		t.Errorf("expected BPAContinue for uppercase input with lowerInput=true, got %s", bpa)
	}
}
