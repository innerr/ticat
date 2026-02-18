package parser

import (
	"testing"
)

func TestParserGlobalEnvInSequence(t *testing.T) {
	root := newCmdTree()
	cmd1 := root.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command 1")
	cmd2 := root.AddSub("cmd2")
	cmd2.RegEmptyCmd("test command 2")

	seqParser := NewSequenceParser(":", []string{"http", "HTTP"}, []string{"/"})
	envParser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"}
	cmdParser := NewCmdParser(envParser, ".", "./", "\t ", "<root>", "^", "/\\")

	parser := NewParser(seqParser, cmdParser)

	t.Run("global env before first cmd is accessible to all cmds", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "{mock-key=mock-value}", "cmd1", ":", "cmd2")

		if parsed.GlobalEnv == nil {
			t.Fatal("GlobalEnv should not be nil")
		}
		if parsed.GlobalEnv["mock-key"].Val != "mock-value" {
			t.Errorf("GlobalEnv should have mock-key=mock-value, got %v", parsed.GlobalEnv["mock-key"])
		}
		if len(parsed.Cmds) != 2 {
			t.Fatalf("expected 2 cmds, got %d", len(parsed.Cmds))
		}
	})

	t.Run("global env with multiple separate brackets", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "{key1=val1}", "{key2=val2}", "cmd1", ":", "cmd2")

		if parsed.GlobalEnv == nil {
			t.Fatal("GlobalEnv should not be nil")
		}
		if parsed.GlobalEnv["key1"].Val != "val1" {
			t.Errorf("GlobalEnv should have key1=val1, got %v", parsed.GlobalEnv["key1"])
		}
		if parsed.GlobalEnv["key2"].Val != "val2" {
			t.Errorf("GlobalEnv should have key2=val2, got %v", parsed.GlobalEnv["key2"])
		}
	})

	t.Run("sequence starting with colon has no global env", func(t *testing.T) {
		parsed := parser.Parse(root, nil, ":", "cmd1", ":", "cmd2")

		if len(parsed.GlobalEnv) != 0 {
			t.Errorf("GlobalEnv should be empty when starting with colon, got %v", parsed.GlobalEnv)
		}
		if len(parsed.Cmds) != 3 {
			t.Fatalf("expected 3 cmds (including empty first), got %d", len(parsed.Cmds))
		}
	})
}

func TestParserGlobalEnvWithThreeCmds(t *testing.T) {
	root := newCmdTree()
	cmd1 := root.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command 1")
	cmd2 := root.AddSub("cmd2")
	cmd2.RegEmptyCmd("test command 2")
	cmd3 := root.AddSub("cmd3")
	cmd3.RegEmptyCmd("test command 3")

	seqParser := NewSequenceParser(":", []string{"http", "HTTP"}, []string{"/"})
	envParser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"}
	cmdParser := NewCmdParser(envParser, ".", "./", "\t ", "<root>", "^", "/\\")

	parser := NewParser(seqParser, cmdParser)

	parsed := parser.Parse(root, nil, "{global-key=global-val}", "cmd1", ":", "cmd2", ":", "cmd3")

	if parsed.GlobalEnv["global-key"].Val != "global-val" {
		t.Errorf("GlobalEnv should have global-key=global-val, got %v", parsed.GlobalEnv["global-key"])
	}
	if len(parsed.Cmds) != 3 {
		t.Fatalf("expected 3 cmds, got %d", len(parsed.Cmds))
	}
}

func TestParserGlobalEnvWithCmdArgs(t *testing.T) {
	root := newCmdTree()
	echo := root.AddSub("echo")
	echo.RegEmptyCmd("print message").
		AddArg("message", "", "msg", "m")

	seqParser := NewSequenceParser(":", []string{"http", "HTTP"}, []string{"/"})
	envParser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"}
	cmdParser := NewCmdParser(envParser, ".", "./", "\t ", "<root>", "^", "/\\")

	parser := NewParser(seqParser, cmdParser)

	t.Run("global env is separate from cmd args", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "{debug=true}", "echo", "hello", ":", "echo", "world")

		if parsed.GlobalEnv["debug"].Val != "true" {
			t.Errorf("GlobalEnv should have debug=true, got %v", parsed.GlobalEnv["debug"])
		}
		if len(parsed.Cmds) != 2 {
			t.Fatalf("expected 2 cmds, got %d", len(parsed.Cmds))
		}

		if parsed.Cmds[0].Segments[1].Env == nil {
			t.Fatal("first cmd segment should have env for args")
		}
		if parsed.Cmds[0].Segments[1].Env["echo.message"].Val != "hello" {
			t.Errorf("first cmd should have echo.message=hello, got %v",
				parsed.Cmds[0].Segments[1].Env["echo.message"])
		}

		if parsed.Cmds[1].Segments[0].Env == nil {
			t.Fatal("second cmd segment should have env for args")
		}
		if parsed.Cmds[1].Segments[0].Env["echo.message"].Val != "world" {
			t.Errorf("second cmd should have echo.message=world, got %v",
				parsed.Cmds[1].Segments[0].Env["echo.message"])
		}
	})
}

func TestParserGlobalEnvCmdNames(t *testing.T) {
	root := newCmdTree()
	cmd1 := root.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command 1")
	cmd2 := root.AddSub("cmd2")
	cmd2.RegEmptyCmd("test command 2")

	seqParser := NewSequenceParser(":", []string{"http", "HTTP"}, []string{"/"})
	envParser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"}
	cmdParser := NewCmdParser(envParser, ".", "./", "\t ", "<root>", "^", "/\\")

	parser := NewParser(seqParser, cmdParser)

	parsed := parser.Parse(root, nil, "{key=val}", "cmd1", ":", "cmd2")

	if len(parsed.Cmds) != 2 {
		t.Fatalf("expected 2 cmds, got %d", len(parsed.Cmds))
	}

	if len(parsed.Cmds[0].Segments) < 2 {
		t.Fatalf("first cmd should have at least 2 segments (env + cmd), got %d", len(parsed.Cmds[0].Segments))
	}

	cmdSeg := parsed.Cmds[0].Segments[1]
	if cmdSeg.Matched.Name != "cmd1" {
		t.Errorf("second segment of first cmd should be cmd1, got %s", cmdSeg.Matched.Name)
	}

	if len(parsed.Cmds[1].Segments) < 1 {
		t.Fatalf("second cmd should have at least 1 segment, got %d", len(parsed.Cmds[1].Segments))
	}

	if parsed.Cmds[1].Segments[0].Matched.Name != "cmd2" {
		t.Errorf("first segment of second cmd should be cmd2, got %s", parsed.Cmds[1].Segments[0].Matched.Name)
	}
}

func TestParserGlobalEnvParsedCmdsStructure(t *testing.T) {
	root := newCmdTree()
	cmd1 := root.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command 1")

	seqParser := NewSequenceParser(":", []string{"http", "HTTP"}, []string{"/"})
	envParser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"}
	cmdParser := NewCmdParser(envParser, ".", "./", "\t ", "<root>", "^", "/\\")

	parser := NewParser(seqParser, cmdParser)

	t.Run("GlobalCmdIdx should point to global area", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "{key=val}", "cmd1")

		if parsed.GlobalCmdIdx != 0 {
			t.Errorf("GlobalCmdIdx should be 0, got %d", parsed.GlobalCmdIdx)
		}
	})

	t.Run("HasTailMode should be false for normal call", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "{key=val}", "cmd1")

		if parsed.HasTailMode {
			t.Error("HasTailMode should be false")
		}
	})
}
