package execute

import (
	"testing"

	"github.com/innerr/ticat/pkg/core/model"
)

func newTestCmdTree() *model.CmdTree {
	return model.NewCmdTree(model.CmdTreeStrsForTest())
}

func createParsedCmd(tree *model.CmdTree, name string, allowTailMode bool, isPriority bool, input []string) model.ParsedCmd {
	sub := tree.GetOrAddSub(name)
	cmd := sub.RegEmptyCmd("test command")
	if allowTailMode {
		cmd.SetAllowTailModeCall()
	}
	if isPriority {
		cmd.SetPriority()
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
		ParseResult: model.ParseResult{
			Input: input,
		},
	}
}

func createEmptyParsedCmd() model.ParsedCmd {
	return model.ParsedCmd{
		Segments:    []model.ParsedCmdSeg{},
		ParseResult: model.ParseResult{Input: []string{}},
	}
}

func TestMoveLastPriorityCmdToFront_GenericCases(t *testing.T) {
	t.Run("single command should not trigger tail mode", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "cmd1", false, false, []string{"cmd1"}),
		}
		reordered, doMove, tailModeCall, attempTailModeCall := moveLastPriorityCmdToFront(flow)

		if doMove {
			t.Error("single command should not trigger move")
		}
		if tailModeCall {
			t.Error("single command should not trigger tailModeCall")
		}
		if attempTailModeCall {
			t.Error("single command should not trigger attempTailModeCall")
		}
		if len(reordered) != 1 {
			t.Errorf("expected 1 command, got %d", len(reordered))
		}
	})

	t.Run("two commands without sep should not trigger tail mode", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "cmd1", false, false, []string{"cmd1"}),
			createParsedCmd(tree, "cmd2", false, false, []string{"cmd2"}),
		}
		reordered, doMove, tailModeCall, attempTailModeCall := moveLastPriorityCmdToFront(flow)

		if doMove {
			t.Error("commands without sep should not trigger move")
		}
		if tailModeCall {
			t.Error("commands without sep should not trigger tailModeCall")
		}
		if attempTailModeCall {
			t.Error("commands without sep should not trigger attempTailModeCall")
		}
		if len(reordered) != 2 {
			t.Errorf("expected 2 commands, got %d", len(reordered))
		}
	})

	t.Run("empty sep + tail-mode-supported command should trigger tail mode", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "arg1", false, false, []string{"arg1"}),
			createEmptyParsedCmd(),
			createParsedCmd(tree, "tailcmd", true, true, []string{"tailcmd"}),
		}
		reordered, doMove, tailModeCall, attempTailModeCall := moveLastPriorityCmdToFront(flow)

		if !doMove {
			t.Error("should trigger move")
		}
		if !tailModeCall {
			t.Error("should trigger tailModeCall")
		}
		if attempTailModeCall {
			t.Error("should not trigger attempTailModeCall for supported command")
		}
		if len(reordered) != 2 {
			t.Errorf("expected 2 commands after trim, got %d", len(reordered))
		}
		if reordered[0].MatchedPath()[0] != "tailcmd" {
			t.Errorf("expected first command to be 'tailcmd', got '%s'", reordered[0].MatchedPath()[0])
		}
		if !reordered[0].TailMode {
			t.Error("first command should have TailMode=true")
		}
	})

	t.Run("empty sep + non-tail-mode command should trigger attempt", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "arg1", false, false, []string{"arg1"}),
			createEmptyParsedCmd(),
			createParsedCmd(tree, "normalcmd", false, false, []string{"normalcmd"}),
		}
		reordered, doMove, tailModeCall, attempTailModeCall := moveLastPriorityCmdToFront(flow)

		if !doMove {
			t.Error("should trigger move")
		}
		if tailModeCall {
			t.Error("should not trigger tailModeCall for unsupported command")
		}
		if !attempTailModeCall {
			t.Error("should trigger attempTailModeCall for unsupported command")
		}
		if len(reordered) != 2 {
			t.Errorf("expected 2 commands after trim, got %d", len(reordered))
		}
	})

	t.Run("priority command at end without tail-mode support should still move", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "arg1", false, false, []string{"arg1"}),
			createParsedCmd(tree, "prioritycmd", false, true, []string{"prioritycmd"}),
		}
		reordered, doMove, _, _ := moveLastPriorityCmdToFront(flow)

		if !doMove {
			t.Error("priority command should trigger move")
		}
		if reordered[0].MatchedPath()[0] != "prioritycmd" {
			t.Errorf("expected first command to be 'prioritycmd', got '%s'", reordered[0].MatchedPath()[0])
		}
	})

	t.Run("three commands with empty sep before last should work", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "arg1", false, false, []string{"arg1"}),
			createParsedCmd(tree, "arg2", false, false, []string{"arg2"}),
			createEmptyParsedCmd(),
			createParsedCmd(tree, "tailcmd", true, true, []string{"tailcmd"}),
		}
		reordered, doMove, tailModeCall, _ := moveLastPriorityCmdToFront(flow)

		if !doMove {
			t.Error("should trigger move")
		}
		if tailModeCall {
			t.Error("should not trigger tailModeCall when len(reordered) > 2")
		}
		if len(reordered) != 3 {
			t.Errorf("expected 3 commands, got %d", len(reordered))
		}
		if reordered[0].MatchedPath()[0] != "tailcmd" {
			t.Errorf("expected first command to be 'tailcmd', got '%s'", reordered[0].MatchedPath()[0])
		}
	})

	t.Run("tail mode with exactly 2 reordered commands should trigger tailModeCall", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "arg1", false, false, []string{"arg1"}),
			createEmptyParsedCmd(),
			createParsedCmd(tree, "tailcmd", true, true, []string{"tailcmd"}),
		}
		_, _, tailModeCall, _ := moveLastPriorityCmdToFront(flow)

		if !tailModeCall {
			t.Error("tailModeCall should be true when len(reordered) == 2")
		}
	})

	t.Run("tail mode with single reordered command should trigger tailModeCall", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createEmptyParsedCmd(),
			createParsedCmd(tree, "tailcmd", true, true, []string{"tailcmd"}),
		}
		reordered, doMove, tailModeCall, _ := moveLastPriorityCmdToFront(flow)

		if !doMove {
			t.Error("should trigger move")
		}
		if !tailModeCall {
			t.Error("tailModeCall should be true when len(reordered) == 1")
		}
		if len(reordered) != 1 {
			t.Errorf("expected 1 command, got %d", len(reordered))
		}
	})

	t.Run("empty commands at tail before move should be trimmed", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "cmd1", false, false, []string{"cmd1"}),
			createEmptyParsedCmd(),
			createEmptyParsedCmd(),
			createParsedCmd(tree, "tailcmd", true, true, []string{"tailcmd"}),
		}
		reordered, doMove, _, _ := moveLastPriorityCmdToFront(flow)

		if !doMove {
			t.Error("should trigger move")
		}
		for _, cmd := range reordered {
			if cmd.IsAllEmptySegments() {
				t.Error("empty commands should be trimmed")
			}
		}
	})

	t.Run("priority command at end without sep should move", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "cmd1", false, false, []string{"cmd1"}),
			createParsedCmd(tree, "tailcmd", true, true, []string{"tailcmd"}),
		}
		reordered, doMove, tailModeCall, attempTailModeCall := moveLastPriorityCmdToFront(flow)

		if !doMove {
			t.Error("should trigger move for priority command at end")
		}
		if !tailModeCall {
			t.Error("should trigger tailModeCall when len(reordered) == 2 and command supports tail mode")
		}
		if attempTailModeCall {
			t.Error("should not trigger attempTailModeCall")
		}
		if reordered[0].MatchedPath()[0] != "tailcmd" {
			t.Errorf("expected first command to be 'tailcmd', got '%s'", reordered[0].MatchedPath()[0])
		}
	})

	t.Run("non-priority command at end without sep should not move", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "cmd1", false, false, []string{"cmd1"}),
			createParsedCmd(tree, "normalcmd", false, false, []string{"normalcmd"}),
		}
		_, doMove, tailModeCall, attempTailModeCall := moveLastPriorityCmdToFront(flow)

		if doMove {
			t.Error("should not trigger move for non-priority command without sep")
		}
		if tailModeCall {
			t.Error("should not trigger tailModeCall")
		}
		if attempTailModeCall {
			t.Error("should not trigger attempTailModeCall")
		}
	})
}

func TestMoveLastPriorityCmdToFront_SupportedCommands(t *testing.T) {
	supportedCommands := []struct {
		name        string
		allowTail   bool
		isPriority  bool
		description string
	}{
		{"find", true, true, "find command"},
		{"/", true, true, "find shortcut"},
		{"//", true, true, "find with usage shortcut"},
		{"///", true, true, "find with full info shortcut"},
		{"tail-sub", true, true, "tail-sub command"},
		{"~", true, true, "tail-sub shortcut"},
		{"~~", true, true, "tail-sub with usage shortcut"},
		{"~~~", true, true, "tail-sub with full info shortcut"},
		{"cmds", true, true, "cmds command"},
		{"cmds.tree", true, false, "cmds.tree command"},
		{"tag", true, false, "tag command"},
		{"cmd", true, true, "cmd command"},
		{"=", true, true, "cmd shortcut"},
		{"==", true, true, "cmd full shortcut"},
		{"desc", true, true, "desc command"},
		{"-", true, true, "desc shortcut"},
		{"--", true, true, "desc shortcut 2"},
		{"+", true, true, "desc more shortcut"},
	}

	for _, tc := range supportedCommands {
		t.Run(tc.description+" should support tail mode", func(t *testing.T) {
			tree := newTestCmdTree()
			flow := []model.ParsedCmd{
				createParsedCmd(tree, "arg1", false, false, []string{"arg1"}),
				createEmptyParsedCmd(),
				createParsedCmd(tree, tc.name, tc.allowTail, tc.isPriority, []string{tc.name}),
			}
			reordered, doMove, tailModeCall, attempTailModeCall := moveLastPriorityCmdToFront(flow)

			if !doMove {
				t.Errorf("%s should trigger move", tc.name)
			}
			if !tailModeCall {
				t.Errorf("%s should trigger tailModeCall", tc.name)
			}
			if attempTailModeCall {
				t.Errorf("%s should not trigger attempTailModeCall", tc.name)
			}
			if reordered[0].MatchedPath()[0] != tc.name {
				t.Errorf("expected first command to be '%s', got '%s'", tc.name, reordered[0].MatchedPath()[0])
			}
		})
	}
}

func TestMoveLastPriorityCmdToFront_UnsupportedCommands(t *testing.T) {
	unsupportedCommands := []string{
		"version",
		"help",
		"tags",
		"randomcmd",
	}

	for _, cmdName := range unsupportedCommands {
		t.Run(cmdName+" should not support tail mode", func(t *testing.T) {
			tree := newTestCmdTree()
			flow := []model.ParsedCmd{
				createParsedCmd(tree, "arg1", false, false, []string{"arg1"}),
				createEmptyParsedCmd(),
				createParsedCmd(tree, cmdName, false, false, []string{cmdName}),
			}
			_, doMove, tailModeCall, attempTailModeCall := moveLastPriorityCmdToFront(flow)

			if !doMove {
				t.Errorf("%s should trigger move", cmdName)
			}
			if tailModeCall {
				t.Errorf("%s should not trigger tailModeCall", cmdName)
			}
			if !attempTailModeCall {
				t.Errorf("%s should trigger attempTailModeCall", cmdName)
			}
		})
	}
}

func TestMoveLastPriorityCmdToFront_MultipleArgs(t *testing.T) {
	t.Run("multiple args with tail mode command", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "arg1", false, false, []string{"arg1"}),
			createParsedCmd(tree, "arg2", false, false, []string{"arg2"}),
			createParsedCmd(tree, "arg3", false, false, []string{"arg3"}),
			createEmptyParsedCmd(),
			createParsedCmd(tree, "find", true, true, []string{"find"}),
		}
		reordered, doMove, _, _ := moveLastPriorityCmdToFront(flow)

		if !doMove {
			t.Error("should trigger move")
		}
		if reordered[0].MatchedPath()[0] != "find" {
			t.Errorf("expected first command to be 'find', got '%s'", reordered[0].MatchedPath()[0])
		}
		if len(reordered) != 4 {
			t.Errorf("expected 4 commands (1 find + 3 args), got %d", len(reordered))
		}
	})

	t.Run("args with env should work with tail mode", func(t *testing.T) {
		tree := newTestCmdTree()
		argSub := tree.GetOrAddSub("arg1")
		argSub.RegEmptyCmd("test arg")

		tailSub := tree.GetOrAddSub("find")
		tailSub.RegEmptyCmd("find command").SetAllowTailModeCall().SetPriority()

		flow := []model.ParsedCmd{
			{
				Segments: []model.ParsedCmdSeg{
					{
						Env: model.ParsedEnv{
							"key": model.NewParsedEnvVal("key", "value"),
						},
						Matched: model.MatchedCmd{Name: "arg1", Cmd: argSub},
					},
				},
				ParseResult: model.ParseResult{Input: []string{"{key=value}"}},
			},
			createEmptyParsedCmd(),
			{
				Segments: []model.ParsedCmdSeg{
					{
						Matched: model.MatchedCmd{Name: "find", Cmd: tailSub},
					},
				},
				ParseResult: model.ParseResult{Input: []string{"find"}},
			},
		}

		reordered, doMove, tailModeCall, _ := moveLastPriorityCmdToFront(flow)

		if !doMove {
			t.Error("should trigger move")
		}
		if !tailModeCall {
			t.Error("should trigger tailModeCall")
		}
		if reordered[0].MatchedPath()[0] != "find" {
			t.Errorf("expected first command to be 'find', got '%s'", reordered[0].MatchedPath()[0])
		}
	})
}

func TestMoveLastPriorityCmdToFront_EdgeCases(t *testing.T) {
	t.Run("empty flow should not crash", func(t *testing.T) {
		flow := []model.ParsedCmd{}
		reordered, doMove, tailModeCall, attempTailModeCall := moveLastPriorityCmdToFront(flow)

		if doMove {
			t.Error("empty flow should not trigger move")
		}
		if tailModeCall {
			t.Error("empty flow should not trigger tailModeCall")
		}
		if attempTailModeCall {
			t.Error("empty flow should not trigger attempTailModeCall")
		}
		if len(reordered) != 0 {
			t.Errorf("expected 0 commands, got %d", len(reordered))
		}
	})

	t.Run("all empty commands should be processed correctly", func(t *testing.T) {
		flow := []model.ParsedCmd{
			createEmptyParsedCmd(),
			createEmptyParsedCmd(),
			createEmptyParsedCmd(),
		}
		reordered, _, _, _ := moveLastPriorityCmdToFront(flow)

		for _, cmd := range reordered {
			if !cmd.IsAllEmptySegments() {
				t.Error("all commands should be empty")
			}
		}
	})

	t.Run("recursive tail mode calls", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "cmd1", false, false, []string{"cmd1"}),
			createEmptyParsedCmd(),
			createParsedCmd(tree, "cmd2", false, false, []string{"cmd2"}),
			createEmptyParsedCmd(),
			createParsedCmd(tree, "find", true, true, []string{"find"}),
		}
		reordered, doMove, _, _ := moveLastPriorityCmdToFront(flow)

		if !doMove {
			t.Error("should trigger move")
		}
		if reordered[0].MatchedPath()[0] != "find" {
			t.Errorf("expected first command to be 'find', got '%s'", reordered[0].MatchedPath()[0])
		}
	})
}

func TestCheckTailModeCalls(t *testing.T) {
	t.Run("tailModeCall true should not panic", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := &model.ParsedCmds{
			Cmds: []model.ParsedCmd{
				createParsedCmd(tree, "find", true, true, []string{"find"}),
			},
			TailModeCall:       true,
			AttempTailModeCall: false,
		}
		defer func() {
			if r := recover(); r != nil {
				t.Error("should not panic for valid tail mode call")
			}
		}()
		checkTailModeCalls(flow)
	})

	t.Run("attempTailModeCall true should panic", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := &model.ParsedCmds{
			Cmds: []model.ParsedCmd{
				createParsedCmd(tree, "normalcmd", false, false, []string{"normalcmd"}),
			},
			TailModeCall:       false,
			AttempTailModeCall: true,
		}
		defer func() {
			if r := recover(); r == nil {
				t.Error("should panic for unsupported tail mode call")
			}
		}()
		checkTailModeCalls(flow)
	})

	t.Run("neither flag set should not panic", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := &model.ParsedCmds{
			Cmds: []model.ParsedCmd{
				createParsedCmd(tree, "cmd1", false, false, []string{"cmd1"}),
			},
			TailModeCall:       false,
			AttempTailModeCall: false,
		}
		defer func() {
			if r := recover(); r != nil {
				t.Error("should not panic for normal call")
			}
		}()
		checkTailModeCalls(flow)
	})
}

func TestMoveLastPriorityCmdToFront_ConditionCombinations(t *testing.T) {
	t.Run("no sep + no priority + no tail mode support = no tail mode", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "arg1", false, false, []string{"arg1"}),
			createParsedCmd(tree, "cmd2", false, false, []string{"cmd2"}),
		}
		_, doMove, tailModeCall, attempTailModeCall := moveLastPriorityCmdToFront(flow)

		if doMove {
			t.Error("should not trigger move")
		}
		if tailModeCall {
			t.Error("should not trigger tailModeCall")
		}
		if attempTailModeCall {
			t.Error("should not trigger attempTailModeCall")
		}
	})

	t.Run("has sep + has priority + no tail mode support = move but not tail mode call", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "arg1", false, false, []string{"arg1"}),
			createEmptyParsedCmd(),
			createParsedCmd(tree, "cmd2", false, true, []string{"cmd2"}),
		}
		_, doMove, tailModeCall, attempTailModeCall := moveLastPriorityCmdToFront(flow)

		if !doMove {
			t.Error("should trigger move for priority command")
		}
		if tailModeCall {
			t.Error("should not trigger tailModeCall without tail mode support")
		}
		if attempTailModeCall {
			t.Error("should not trigger attempTailModeCall for priority command")
		}
	})

	t.Run("has sep + no priority + has tail mode support = tail mode call", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "arg1", false, false, []string{"arg1"}),
			createEmptyParsedCmd(),
			createParsedCmd(tree, "cmd2", true, false, []string{"cmd2"}),
		}
		_, doMove, tailModeCall, attempTailModeCall := moveLastPriorityCmdToFront(flow)

		if !doMove {
			t.Error("should trigger move")
		}
		if !tailModeCall {
			t.Error("should trigger tailModeCall")
		}
		if attempTailModeCall {
			t.Error("should not trigger attempTailModeCall")
		}
	})

	t.Run("has sep + no priority + no tail mode support = attempt tail mode call", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "arg1", false, false, []string{"arg1"}),
			createEmptyParsedCmd(),
			createParsedCmd(tree, "cmd2", false, false, []string{"cmd2"}),
		}
		_, doMove, tailModeCall, attempTailModeCall := moveLastPriorityCmdToFront(flow)

		if !doMove {
			t.Error("should trigger move")
		}
		if tailModeCall {
			t.Error("should not trigger tailModeCall")
		}
		if !attempTailModeCall {
			t.Error("should trigger attempTailModeCall")
		}
	})

	t.Run("has sep + has priority + has tail mode support = tail mode call", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "arg1", false, false, []string{"arg1"}),
			createEmptyParsedCmd(),
			createParsedCmd(tree, "cmd2", true, true, []string{"cmd2"}),
		}
		_, doMove, tailModeCall, attempTailModeCall := moveLastPriorityCmdToFront(flow)

		if !doMove {
			t.Error("should trigger move")
		}
		if !tailModeCall {
			t.Error("should trigger tailModeCall")
		}
		if attempTailModeCall {
			t.Error("should not trigger attempTailModeCall")
		}
	})

	t.Run("priority command at end without sep = move", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "arg1", false, false, []string{"arg1"}),
			createParsedCmd(tree, "prioritycmd", false, true, []string{"prioritycmd"}),
		}
		reordered, doMove, _, _ := moveLastPriorityCmdToFront(flow)

		if !doMove {
			t.Error("should trigger move for priority command at end")
		}
		if reordered[0].MatchedPath()[0] != "prioritycmd" {
			t.Errorf("expected first command to be 'prioritycmd', got '%s'", reordered[0].MatchedPath()[0])
		}
	})
}

func TestMoveLastPriorityCmdToFront_CommandOrderPreservation(t *testing.T) {
	t.Run("args order should be preserved after tail command moved to front", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "arg1", false, false, []string{"arg1"}),
			createParsedCmd(tree, "arg2", false, false, []string{"arg2"}),
			createParsedCmd(tree, "arg3", false, false, []string{"arg3"}),
			createEmptyParsedCmd(),
			createParsedCmd(tree, "find", true, true, []string{"find"}),
		}
		reordered, _, _, _ := moveLastPriorityCmdToFront(flow)

		if len(reordered) != 4 {
			t.Fatalf("expected 4 commands, got %d", len(reordered))
		}

		if reordered[0].MatchedPath()[0] != "find" {
			t.Errorf("expected reordered[0] to be 'find', got '%s'", reordered[0].MatchedPath()[0])
		}
		if reordered[1].MatchedPath()[0] != "arg1" {
			t.Errorf("expected reordered[1] to be 'arg1', got '%s'", reordered[1].MatchedPath()[0])
		}
		if reordered[2].MatchedPath()[0] != "arg2" {
			t.Errorf("expected reordered[2] to be 'arg2', got '%s'", reordered[2].MatchedPath()[0])
		}
		if reordered[3].MatchedPath()[0] != "arg3" {
			t.Errorf("expected reordered[3] to be 'arg3', got '%s'", reordered[3].MatchedPath()[0])
		}
	})

	t.Run("tail mode flag should only be set on moved command", func(t *testing.T) {
		tree := newTestCmdTree()
		flow := []model.ParsedCmd{
			createParsedCmd(tree, "arg1", false, false, []string{"arg1"}),
			createEmptyParsedCmd(),
			createParsedCmd(tree, "find", true, true, []string{"find"}),
		}
		reordered, _, _, _ := moveLastPriorityCmdToFront(flow)

		if !reordered[0].TailMode {
			t.Error("moved command should have TailMode=true")
		}
		if reordered[1].TailMode {
			t.Error("non-moved command should have TailMode=false")
		}
	})
}
