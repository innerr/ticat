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

func TestParserTailModeSequence(t *testing.T) {
	root := newCmdTree()

	tailModeCmd := root.AddSub("find")
	tailModeCmd.RegEmptyCmd("find command").SetAllowTailModeCall().SetPriority()

	normalCmd := root.AddSub("cmd1")
	normalCmd.RegEmptyCmd("normal command")

	seqParser := NewSequenceParser(":", []string{"http", "HTTP"}, []string{"/"})
	envParser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"}
	cmdParser := NewCmdParser(envParser, ".", "./", "\t ", "<root>", "^", "/\\")

	parser := NewParser(seqParser, cmdParser)

	t.Run("double colon creates empty command in sequence", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1", ":", ":", "find")

		if len(parsed.Cmds) < 2 {
			t.Fatalf("expected at least 2 cmds, got %d", len(parsed.Cmds))
		}
	})

	t.Run("tail mode command with arg before double colon", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1", ":", ":", "find")

		if len(parsed.Cmds) < 2 {
			t.Fatalf("expected at least 2 cmds, got %d", len(parsed.Cmds))
		}
	})

	t.Run("single colon followed by command is normal sequence", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1", ":", "find")

		if parsed.HasTailMode {
			t.Error("HasTailMode should be false for single colon sequence")
		}
		if len(parsed.Cmds) != 2 {
			t.Fatalf("expected 2 cmds, got %d", len(parsed.Cmds))
		}
	})

	t.Run("global env with tail mode syntax", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "{key=val}", "cmd1", ":", ":", "find")

		if parsed.GlobalEnv == nil {
			t.Fatal("GlobalEnv should not be nil")
		}
		if parsed.GlobalEnv["key"].Val != "val" {
			t.Errorf("GlobalEnv should have key=val, got %v", parsed.GlobalEnv["key"])
		}
	})

	t.Run("multiple args before double colon", func(t *testing.T) {
		arg1 := root.AddSub("arg1")
		arg1.RegEmptyCmd("arg1 command")
		arg2 := root.AddSub("arg2")
		arg2.RegEmptyCmd("arg2 command")

		parsed := parser.Parse(root, nil, "arg1", "arg2", ":", ":", "find")

		if len(parsed.Cmds) < 3 {
			t.Fatalf("expected at least 3 cmds, got %d", len(parsed.Cmds))
		}
	})
}

func TestParserTailModeWithEnv(t *testing.T) {
	root := newCmdTree()

	tailModeCmd := root.AddSub("find")
	tailModeCmd.RegEmptyCmd("find command").SetAllowTailModeCall().SetPriority()

	normalCmd := root.AddSub("cmd1")
	normalCmd.RegEmptyCmd("normal command")

	seqParser := NewSequenceParser(":", []string{"http", "HTTP"}, []string{"/"})
	envParser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"}
	cmdParser := NewCmdParser(envParser, ".", "./", "\t ", "<root>", "^", "/\\")

	parser := NewParser(seqParser, cmdParser)

	t.Run("env before double colon should be in global env", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "{global=val}", "cmd1", ":", ":", "find")

		if parsed.GlobalEnv == nil {
			t.Fatal("GlobalEnv should not be nil")
		}
		if parsed.GlobalEnv["global"].Val != "val" {
			t.Errorf("GlobalEnv should have global=val, got %v", parsed.GlobalEnv["global"])
		}
	})

	t.Run("env with cmd-specific args and tail mode", func(t *testing.T) {
		cmdWithArgs := root.AddSub("cmdargs")
		cmdWithArgs.RegEmptyCmd("command with args").AddArg("message", "", "m")

		parsed := parser.Parse(root, nil, "{key=val}", "cmdargs", "hello", ":", ":", "find")

		if parsed.GlobalEnv == nil {
			t.Fatal("GlobalEnv should not be nil")
		}
		if parsed.GlobalEnv["key"].Val != "val" {
			t.Errorf("GlobalEnv should have key=val, got %v", parsed.GlobalEnv["key"])
		}
	})
}

func TestParserTailModeShortcuts(t *testing.T) {
	root := newCmdTree()

	shortcuts := []struct {
		name string
	}{
		{"/"},
		{"//"},
		{"///"},
		{"~"},
		{"~~"},
		{"~~~"},
		{"="},
		{"=="},
		{"-"},
		{"--"},
		{"+"},
		{"++"},
	}

	for _, sc := range shortcuts {
		cmd := root.AddSub(sc.name)
		cmd.RegEmptyCmd("shortcut command").SetAllowTailModeCall().SetPriority()
	}

	seqParser := NewSequenceParser(":", []string{"http", "HTTP"}, []string{"/"})
	envParser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"}
	cmdParser := NewCmdParser(envParser, ".", "./", "\t ", "<root>", "^", "/\\")

	parser := NewParser(seqParser, cmdParser)

	for _, sc := range shortcuts {
		t.Run("shortcut "+sc.name+" should be parseable in tail mode", func(t *testing.T) {
			argCmd := root.AddSub("arg_for_" + sc.name)
			argCmd.RegEmptyCmd("arg command")

			parsed := parser.Parse(root, nil, "arg_for_"+sc.name, ":", ":", sc.name)

			if len(parsed.Cmds) < 2 {
				t.Fatalf("expected at least 2 cmds for %s, got %d", sc.name, len(parsed.Cmds))
			}
		})
	}
}

func TestParserTailModeWithMultipleColons(t *testing.T) {
	root := newCmdTree()

	cmd1 := root.AddSub("cmd1")
	cmd1.RegEmptyCmd("command 1")

	cmd2 := root.AddSub("cmd2")
	cmd2.RegEmptyCmd("command 2")

	tailModeCmd := root.AddSub("find")
	tailModeCmd.RegEmptyCmd("find command").SetAllowTailModeCall().SetPriority()

	seqParser := NewSequenceParser(":", []string{"http", "HTTP"}, []string{"/"})
	envParser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"}
	cmdParser := NewCmdParser(envParser, ".", "./", "\t ", "<root>", "^", "/\\")

	parser := NewParser(seqParser, cmdParser)

	t.Run("triple colons should create two empty commands", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1", ":", ":", ":", "find")

		if len(parsed.Cmds) < 3 {
			t.Fatalf("expected at least 3 cmds, got %d", len(parsed.Cmds))
		}
	})

	t.Run("alternating commands and colons", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1", ":", "cmd2", ":", ":", "find")

		if len(parsed.Cmds) < 3 {
			t.Fatalf("expected at least 3 cmds, got %d", len(parsed.Cmds))
		}
	})
}

func TestParserHelpFlagTransform(t *testing.T) {
	root := newCmdTree()
	help := root.AddSub("help")
	help.RegEmptyCmd("show help")

	cmd := root.AddSub("cmd")
	cmd.RegEmptyCmd("show command info")
	cmdFullWithFlow := cmd.AddSub("full-with-flow")
	cmdFullWithFlow.RegEmptyCmd("show command info with flow")

	desc := root.AddSub("desc")
	descMore := desc.AddSub("more")
	descMore.RegEmptyCmd("show more description")

	cmd1 := root.AddSub("cmd1")
	cmd1.RegEmptyCmd("command 1")

	cmd2 := root.AddSub("cmd2")
	cmd2.RegEmptyCmd("command 2")

	cmd3 := root.AddSub("cmd3")
	cmd3.RegEmptyCmd("command 3")

	seqParser := NewSequenceParser(":", []string{"http", "HTTP"}, []string{"/"})
	envParser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"}
	cmdParser := NewCmdParser(envParser, ".", "./", "\t ", "<root>", "^", "/\\")

	parser := NewParser(seqParser, cmdParser)

	t.Run("-h alone transforms to help", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "-h")
		if len(parsed.Cmds) != 1 {
			t.Fatalf("expected 1 cmd, got %d", len(parsed.Cmds))
		}
		if len(parsed.Cmds[0].Segments) < 1 {
			t.Fatal("expected at least 1 segment")
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "help" {
			t.Errorf("expected command 'help', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("--help alone transforms to help", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "--help")
		if len(parsed.Cmds) != 1 {
			t.Fatalf("expected 1 cmd, got %d", len(parsed.Cmds))
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "help" {
			t.Errorf("expected command 'help', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("single cmd with -h transforms to cmd.full-with-flow {cmd}", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1", "-h")
		if len(parsed.Cmds) != 1 {
			t.Fatalf("expected 1 cmd, got %d", len(parsed.Cmds))
		}
		if len(parsed.Cmds[0].Segments) < 3 {
			t.Fatalf("expected at least 3 segments, got %d", len(parsed.Cmds[0].Segments))
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
			t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
		if parsed.Cmds[0].Segments[1].Matched.Name != "full-with-flow" {
			t.Errorf("expected second segment 'full-with-flow', got %q", parsed.Cmds[0].Segments[1].Matched.Name)
		}
		if parsed.Cmds[0].Segments[2].Matched.Name != "cmd1" {
			t.Errorf("expected third segment 'cmd1', got %q", parsed.Cmds[0].Segments[2].Matched.Name)
		}
	})

	t.Run("single cmd with --help transforms to cmd.full-with-flow {cmd}", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1", "--help")
		if len(parsed.Cmds) != 1 {
			t.Fatalf("expected 1 cmd, got %d", len(parsed.Cmds))
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
			t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
		if parsed.Cmds[0].Segments[1].Matched.Name != "full-with-flow" {
			t.Errorf("expected second segment 'full-with-flow', got %q", parsed.Cmds[0].Segments[1].Matched.Name)
		}
		if parsed.Cmds[0].Segments[2].Matched.Name != "cmd1" {
			t.Errorf("expected third segment 'cmd1', got %q", parsed.Cmds[0].Segments[2].Matched.Name)
		}
	})

	t.Run("two cmds with : and -h transforms to desc.more : {cmd1} : {cmd2}", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1", ":", "cmd2", "-h")
		if len(parsed.Cmds) != 3 {
			t.Fatalf("expected 3 cmds (desc.more, cmd1, cmd2), got %d", len(parsed.Cmds))
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
			t.Errorf("expected first cmd first segment 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
		if parsed.Cmds[0].Segments[1].Matched.Name != "more" {
			t.Errorf("expected first cmd second segment 'more', got %q", parsed.Cmds[0].Segments[1].Matched.Name)
		}
		if parsed.Cmds[1].Segments[0].Matched.Name != "cmd1" {
			t.Errorf("expected second cmd 'cmd1', got %q", parsed.Cmds[1].Segments[0].Matched.Name)
		}
		if parsed.Cmds[2].Segments[0].Matched.Name != "cmd2" {
			t.Errorf("expected third cmd 'cmd2', got %q", parsed.Cmds[2].Segments[0].Matched.Name)
		}
	})

	t.Run("two cmds with : and --help transforms to desc.more", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1", ":", "cmd2", "--help")
		if len(parsed.Cmds) != 3 {
			t.Fatalf("expected 3 cmds, got %d", len(parsed.Cmds))
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
			t.Errorf("expected first segment 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("three cmds with : and -h transforms to desc.more : cmd1 : cmd2 : cmd3", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1", ":", "cmd2", ":", "cmd3", "-h")
		if len(parsed.Cmds) != 4 {
			t.Fatalf("expected 4 cmds (desc.more, cmd1, cmd2, cmd3), got %d", len(parsed.Cmds))
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
			t.Errorf("expected first segment 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("three cmds with : and --help", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1", ":", "cmd2", ":", "cmd3", "--help")
		if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
			t.Errorf("expected first segment 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("-h in middle of args is not transformed", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1", "-h", "cmd2")
		if parsed.Cmds[0].Segments[0].Matched.Name != "cmd1" {
			t.Errorf("expected first segment 'cmd1', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("--help in middle of args is not transformed", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1", "--help", "cmd2")
		if parsed.Cmds[0].Segments[0].Matched.Name != "cmd1" {
			t.Errorf("expected first segment 'cmd1', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("empty input returns empty", func(t *testing.T) {
		parsed := parser.Parse(root, nil)
		if len(parsed.Cmds) != 1 {
			t.Fatalf("expected 1 cmd (empty), got %d", len(parsed.Cmds))
		}
	})

	t.Run("single cmd without help flag is not transformed", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1")
		if len(parsed.Cmds) != 1 {
			t.Fatalf("expected 1 cmd, got %d", len(parsed.Cmds))
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "cmd1" {
			t.Errorf("expected first segment 'cmd1', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("help flag with global env before cmd", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "{key=val}", "cmd1", "-h")
		if parsed.GlobalEnv["key"].Val != "val" {
			t.Errorf("expected global env key=val, got %v", parsed.GlobalEnv["key"])
		}
		if len(parsed.Cmds) < 1 {
			t.Fatalf("expected at least 1 cmd, got %d", len(parsed.Cmds))
		}
		found := false
		for _, seg := range parsed.Cmds[0].Segments {
			if seg.Matched.Name == "cmd" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find segment 'cmd' in parsed cmd")
		}
	})

	t.Run("help flag with cmd-specific env", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1", "{arg=value}", "-h")
		if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
			t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
		if parsed.Cmds[0].Segments[1].Matched.Name != "full-with-flow" {
			t.Errorf("expected second segment 'full-with-flow', got %q", parsed.Cmds[0].Segments[1].Matched.Name)
		}
	})

	t.Run("help flag with multiple env blocks", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "{a=1}", "{b=2}", "cmd1", "-h")
		if parsed.GlobalEnv["a"].Val != "1" {
			t.Errorf("expected global env a=1, got %v", parsed.GlobalEnv["a"])
		}
		if parsed.GlobalEnv["b"].Val != "2" {
			t.Errorf("expected global env b=2, got %v", parsed.GlobalEnv["b"])
		}
		if len(parsed.Cmds) < 1 {
			t.Fatalf("expected at least 1 cmd, got %d", len(parsed.Cmds))
		}
		found := false
		for _, seg := range parsed.Cmds[0].Segments {
			if seg.Matched.Name == "cmd" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find segment 'cmd' in parsed cmd")
		}
	})

	t.Run("subcommand with -h", func(t *testing.T) {
		subCmd := cmd1.AddSub("sub")
		subCmd.RegEmptyCmd("sub command")

		parsed := parser.Parse(root, nil, "cmd1.sub", "-h")
		if len(parsed.Cmds[0].Segments) < 2 {
			t.Fatalf("expected at least 2 segments, got %d", len(parsed.Cmds[0].Segments))
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
			t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
		if parsed.Cmds[0].Segments[1].Matched.Name != "full-with-flow" {
			t.Errorf("expected second segment 'full-with-flow', got %q", parsed.Cmds[0].Segments[1].Matched.Name)
		}
	})

	t.Run("-h at start is not transformed as help flag", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "-h", "cmd1")
		if len(parsed.Cmds) < 1 || len(parsed.Cmds[0].Segments) < 1 {
			t.Fatal("expected at least 1 cmd with 1 segment")
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "-h" {
			t.Errorf("expected first segment '-h', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("sequence with colon only at start converts to cmd.full-with-flow", func(t *testing.T) {
		parsed := parser.Parse(root, nil, ":", "cmd1", "-h")
		if len(parsed.Cmds) < 1 || len(parsed.Cmds[0].Segments) < 2 {
			t.Fatal("expected at least 1 cmd with 2 segments")
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
			t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
		if parsed.Cmds[0].Segments[1].Matched.Name != "full-with-flow" {
			t.Errorf("expected second segment 'full-with-flow', got %q", parsed.Cmds[0].Segments[1].Matched.Name)
		}
	})

	t.Run("sequence with double colon", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1", ":", ":", "cmd2", "-h")
		if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
			t.Errorf("expected first segment 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})
}

func TestParserHelpFlagNoSpace(t *testing.T) {
	root := newCmdTree()
	help := root.AddSub("help")
	help.RegEmptyCmd("show help")

	cmd := root.AddSub("cmd")
	cmd.RegEmptyCmd("show command info")

	desc := root.AddSub("desc")
	descMore := desc.AddSub("more")
	descMore.RegEmptyCmd("show more description")

	cmd1 := root.AddSub("cmd1")
	cmd1.RegEmptyCmd("command 1")

	cmd2 := root.AddSub("cmd2")
	cmd2.RegEmptyCmd("command 2")

	seqParser := NewSequenceParser(":", []string{"http", "HTTP"}, []string{"/"})
	envParser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"}
	cmdParser := NewCmdParser(envParser, ".", "./", "\t ", "<root>", "^", "/\\")

	parser := NewParser(seqParser, cmdParser)

	t.Run("no space separator - two cmds with --help", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1:cmd2", "--help")
		if len(parsed.Cmds) != 3 {
			t.Fatalf("expected 3 cmds (desc.more, cmd1, cmd2), got %d", len(parsed.Cmds))
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
			t.Errorf("expected first cmd 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
		if parsed.Cmds[1].Segments[0].Matched.Name != "cmd1" {
			t.Errorf("expected second cmd 'cmd1', got %q", parsed.Cmds[1].Segments[0].Matched.Name)
		}
		if parsed.Cmds[2].Segments[0].Matched.Name != "cmd2" {
			t.Errorf("expected third cmd 'cmd2', got %q", parsed.Cmds[2].Segments[0].Matched.Name)
		}
	})

	t.Run("no space separator - two cmds with -h", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1:cmd2", "-h")
		if len(parsed.Cmds) != 3 {
			t.Fatalf("expected 3 cmds, got %d", len(parsed.Cmds))
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
			t.Errorf("expected first segment 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("no space separator - three cmds with --help", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1:cmd2:cmd1", "--help")
		if len(parsed.Cmds) != 4 {
			t.Fatalf("expected 4 cmds, got %d", len(parsed.Cmds))
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
			t.Errorf("expected first segment 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("no space separator with -?", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1:cmd2", "-?")
		if len(parsed.Cmds) != 3 {
			t.Fatalf("expected 3 cmds, got %d", len(parsed.Cmds))
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
			t.Errorf("expected first cmd 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("no space separator with ?", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1:cmd2", "?")
		if len(parsed.Cmds) != 3 {
			t.Fatalf("expected 3 cmds, got %d", len(parsed.Cmds))
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
			t.Errorf("expected first cmd 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("no space separator with -help", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1:cmd2", "-help")
		if len(parsed.Cmds) != 3 {
			t.Fatalf("expected 3 cmds, got %d", len(parsed.Cmds))
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
			t.Errorf("expected first cmd 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("mixed space and no space separator", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1:cmd2", ":", "cmd1", "--help")
		if len(parsed.Cmds) != 4 {
			t.Fatalf("expected 4 cmds, got %d", len(parsed.Cmds))
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
			t.Errorf("expected first segment 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("single cmd with no space colon is not sequence", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "cmd1", "--help")
		if len(parsed.Cmds) != 1 {
			t.Fatalf("expected 1 cmd, got %d", len(parsed.Cmds))
		}
		if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
			t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
		if parsed.Cmds[0].Segments[1].Matched.Name != "full-with-flow" {
			t.Errorf("expected second segment 'full-with-flow', got %q", parsed.Cmds[0].Segments[1].Matched.Name)
		}
	})

	t.Run("no space separator with global env", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "{key=val}", "cmd1:cmd2", "--help")
		if parsed.GlobalEnv["key"].Val != "val" {
			t.Errorf("expected global env key=val, got %v", parsed.GlobalEnv["key"])
		}
		if len(parsed.Cmds[0].Segments) < 2 {
			t.Fatalf("expected at least 2 segments in first cmd, got %d", len(parsed.Cmds[0].Segments))
		}
		if parsed.Cmds[0].Segments[1].Matched.Name != "desc" {
			t.Errorf("expected second segment 'desc', got %q", parsed.Cmds[0].Segments[1].Matched.Name)
		}
	})
}

func TestParserHelpFlagTransformEdgeCases(t *testing.T) {
	root := newCmdTree()
	help := root.AddSub("help")
	help.RegEmptyCmd("show help")

	cmd := root.AddSub("cmd")
	cmd.RegEmptyCmd("show command info")

	desc := root.AddSub("desc")
	descMore := desc.AddSub("more")
	descMore.RegEmptyCmd("show more description")

	myCmd := root.AddSub("myCmd")
	myCmd.RegEmptyCmd("my command")

	seqParser := NewSequenceParser(":", []string{"http", "HTTP"}, []string{"/"})
	envParser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"}
	cmdParser := NewCmdParser(envParser, ".", "./", "\t ", "<root>", "^", "/\\")

	parser := NewParser(seqParser, cmdParser)

	t.Run("help-like value not at end is not transformed", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "myCmd", "-h", "value")
		if parsed.Cmds[0].Segments[0].Matched.Name != "myCmd" {
			t.Errorf("expected 'myCmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("partial help flag -he is not transformed", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "myCmd", "-he")
		if parsed.Cmds[0].Segments[0].Matched.Name != "myCmd" {
			t.Errorf("expected 'myCmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("help with equals --help=full is not transformed", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "myCmd", "--help=full")
		if parsed.Cmds[0].Segments[0].Matched.Name != "myCmd" {
			t.Errorf("expected 'myCmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("short flag -H (uppercase) is not transformed", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "myCmd", "-H")
		if parsed.Cmds[0].Segments[0].Matched.Name != "myCmd" {
			t.Errorf("expected 'myCmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("help flag with multiple colons in sequence", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "myCmd", ":", "myCmd", ":", "myCmd", "-h")
		if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
			t.Errorf("expected first segment 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("help with env containing colon", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "myCmd", "{url=http://example.com}", "-h")
		if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
			t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("help flag with spaces in args", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "myCmd", "  ", "-h")
		if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
			t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("transformHelpFlag is idempotent for non-help args", func(t *testing.T) {
		parsed1 := parser.Parse(root, nil, "myCmd", "arg1", "arg2")
		parsed2 := parser.Parse(root, nil, "myCmd", "arg1", "arg2")
		if parsed1.Cmds[0].Segments[0].Matched.Name != parsed2.Cmds[0].Segments[0].Matched.Name {
			t.Errorf("idempotency check failed")
		}
	})
}

func TestParserHelpFlagTransformIntegration(t *testing.T) {
	root := newCmdTree()
	help := root.AddSub("help")
	help.RegEmptyCmd("show help")

	cmd := root.AddSub("cmd")
	cmd.RegEmptyCmd("show command info")

	desc := root.AddSub("desc")
	descMore := desc.AddSub("more")
	descMore.RegEmptyCmd("show more description")

	echo := root.AddSub("echo")
	echo.RegEmptyCmd("print message").AddArg("message", "", "msg", "m")

	seqParser := NewSequenceParser(":", []string{"http", "HTTP"}, []string{"/"})
	envParser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"}
	cmdParser := NewCmdParser(envParser, ".", "./", "\t ", "<root>", "^", "/\\")

	parser := NewParser(seqParser, cmdParser)

	t.Run("help with cmd args", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "echo", "hello", "-h")
		if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
			t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("help with named args", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "echo", "{message=world}", "-h")
		if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
			t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("help with abbreviation args", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "echo", "m", "hello", "-h")
		if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
			t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("sequence with args and help", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "echo", "hello", ":", "echo", "world", "-h")
		if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
			t.Errorf("expected first segment 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
		}
	})

	t.Run("global env with help in sequence", func(t *testing.T) {
		parsed := parser.Parse(root, nil, "{debug=true}", "echo", ":", "echo", "-h")
		if parsed.GlobalEnv == nil || parsed.GlobalEnv["debug"].Val != "true" {
			t.Errorf("expected global env debug=true, got %v", parsed.GlobalEnv)
		}
		if len(parsed.Cmds[0].Segments) < 2 {
			t.Fatalf("expected at least 2 segments in first cmd, got %d", len(parsed.Cmds[0].Segments))
		}
		if parsed.Cmds[0].Segments[1].Matched.Name != "desc" {
			t.Errorf("expected second segment 'desc', got %q", parsed.Cmds[0].Segments[1].Matched.Name)
		}
	})
}

func TestParserHelpFlagVariants(t *testing.T) {
	root := newCmdTree()
	help := root.AddSub("help")
	help.RegEmptyCmd("show help")

	cmd := root.AddSub("cmd")
	cmd.RegEmptyCmd("show command info")
	cmdFullWithFlow := cmd.AddSub("full-with-flow")
	cmdFullWithFlow.RegEmptyCmd("show command info with flow")

	desc := root.AddSub("desc")
	descMore := desc.AddSub("more")
	descMore.RegEmptyCmd("show more description")

	cmd1 := root.AddSub("cmd1")
	cmd1.RegEmptyCmd("command 1")

	cmd2 := root.AddSub("cmd2")
	cmd2.RegEmptyCmd("command 2")

	seqParser := NewSequenceParser(":", []string{"http", "HTTP"}, []string{"/"})
	envParser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"}
	cmdParser := NewCmdParser(envParser, ".", "./", "\t ", "<root>", "^", "/\\")

	parser := NewParser(seqParser, cmdParser)

	helpFlags := []struct {
		name string
		flag string
	}{
		{"dash-question", "-?"},
		{"question", "?"},
		{"dash-help", "-help"},
	}

	for _, hf := range helpFlags {
		t.Run(hf.name+" alone transforms to help", func(t *testing.T) {
			parsed := parser.Parse(root, nil, hf.flag)
			if len(parsed.Cmds) != 1 {
				t.Fatalf("expected 1 cmd, got %d", len(parsed.Cmds))
			}
			if parsed.Cmds[0].Segments[0].Matched.Name != "help" {
				t.Errorf("expected command 'help', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
			}
		})

		t.Run(hf.name+" with single cmd transforms to cmd.full-with-flow {cmd}", func(t *testing.T) {
			parsed := parser.Parse(root, nil, "cmd1", hf.flag)
			if len(parsed.Cmds) != 1 {
				t.Fatalf("expected 1 cmd, got %d", len(parsed.Cmds))
			}
			if len(parsed.Cmds[0].Segments) < 3 {
				t.Fatalf("expected at least 3 segments, got %d", len(parsed.Cmds[0].Segments))
			}
			if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
				t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
			}
			if parsed.Cmds[0].Segments[1].Matched.Name != "full-with-flow" {
				t.Errorf("expected second segment 'full-with-flow', got %q", parsed.Cmds[0].Segments[1].Matched.Name)
			}
			if parsed.Cmds[0].Segments[2].Matched.Name != "cmd1" {
				t.Errorf("expected third segment 'cmd1', got %q", parsed.Cmds[0].Segments[2].Matched.Name)
			}
		})

		t.Run(hf.name+" with two cmds transforms to desc.more", func(t *testing.T) {
			parsed := parser.Parse(root, nil, "cmd1", ":", "cmd2", hf.flag)
			if len(parsed.Cmds) != 3 {
				t.Fatalf("expected 3 cmds (desc.more, cmd1, cmd2), got %d", len(parsed.Cmds))
			}
			if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
				t.Errorf("expected first cmd 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
			}
		})

		t.Run(hf.name+" with three cmds transforms to desc.more", func(t *testing.T) {
			parsed := parser.Parse(root, nil, "cmd1", ":", "cmd2", ":", "cmd1", hf.flag)
			if len(parsed.Cmds) != 4 {
				t.Fatalf("expected 4 cmds, got %d", len(parsed.Cmds))
			}
			if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
				t.Errorf("expected first segment 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
			}
		})

		t.Run(hf.name+" in middle of args is not transformed", func(t *testing.T) {
			parsed := parser.Parse(root, nil, "cmd1", hf.flag, "cmd2")
			if parsed.Cmds[0].Segments[0].Matched.Name != "cmd1" {
				t.Errorf("expected first segment 'cmd1', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
			}
		})

		t.Run(hf.name+" with global env before cmd", func(t *testing.T) {
			parsed := parser.Parse(root, nil, "{key=val}", "cmd1", hf.flag)
			if parsed.GlobalEnv["key"].Val != "val" {
				t.Errorf("expected global env key=val, got %v", parsed.GlobalEnv["key"])
			}
			if len(parsed.Cmds) < 1 {
				t.Fatalf("expected at least 1 cmd, got %d", len(parsed.Cmds))
			}
			found := false
			for _, seg := range parsed.Cmds[0].Segments {
				if seg.Matched.Name == "cmd" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected to find segment 'cmd' in parsed cmd")
			}
		})

		t.Run(hf.name+" with cmd-specific env", func(t *testing.T) {
			parsed := parser.Parse(root, nil, "cmd1", "{arg=value}", hf.flag)
			if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
				t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
			}
		})

		t.Run(hf.name+" with multiple env blocks", func(t *testing.T) {
			parsed := parser.Parse(root, nil, "{a=1}", "{b=2}", "cmd1", hf.flag)
			if parsed.GlobalEnv["a"].Val != "1" {
				t.Errorf("expected global env a=1, got %v", parsed.GlobalEnv["a"])
			}
			if parsed.GlobalEnv["b"].Val != "2" {
				t.Errorf("expected global env b=2, got %v", parsed.GlobalEnv["b"])
			}
			if len(parsed.Cmds[0].Segments) < 2 {
				t.Fatalf("expected at least 2 segments, got %d", len(parsed.Cmds[0].Segments))
			}
			if parsed.Cmds[0].Segments[1].Matched.Name != "cmd" {
				t.Errorf("expected second segment 'cmd', got %q", parsed.Cmds[0].Segments[1].Matched.Name)
			}
		})

		t.Run(hf.name+" with subcommand", func(t *testing.T) {
			subCmd := cmd1.AddSub("sub")
			subCmd.RegEmptyCmd("sub command")

			parsed := parser.Parse(root, nil, "cmd1.sub", hf.flag)
			if len(parsed.Cmds[0].Segments) < 2 {
				t.Fatalf("expected at least 2 segments, got %d", len(parsed.Cmds[0].Segments))
			}
			if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
				t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
			}
		})

		t.Run(hf.name+" with sequence with colon only at start", func(t *testing.T) {
			parsed := parser.Parse(root, nil, ":", "cmd1", hf.flag)
			if len(parsed.Cmds) < 1 || len(parsed.Cmds[0].Segments) < 1 {
				t.Fatal("expected at least 1 cmd")
			}
			if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
				t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
			}
		})

		t.Run(hf.name+" with sequence with double colon", func(t *testing.T) {
			parsed := parser.Parse(root, nil, "cmd1", ":", ":", "cmd2", hf.flag)
			if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
				t.Errorf("expected first segment 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
			}
		})
	}
}

func TestParserHelpFlagVariantsWithArgs(t *testing.T) {
	root := newCmdTree()
	help := root.AddSub("help")
	help.RegEmptyCmd("show help")

	cmd := root.AddSub("cmd")
	cmd.RegEmptyCmd("show command info")

	desc := root.AddSub("desc")
	descMore := desc.AddSub("more")
	descMore.RegEmptyCmd("show more description")

	echo := root.AddSub("echo")
	echo.RegEmptyCmd("print message").AddArg("message", "", "msg", "m")

	seqParser := NewSequenceParser(":", []string{"http", "HTTP"}, []string{"/"})
	envParser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"}
	cmdParser := NewCmdParser(envParser, ".", "./", "\t ", "<root>", "^", "/\\")

	parser := NewParser(seqParser, cmdParser)

	helpFlags := []struct {
		name string
		flag string
	}{
		{"dash-question", "-?"},
		{"question", "?"},
		{"dash-help", "-help"},
	}

	for _, hf := range helpFlags {
		t.Run(hf.name+" with cmd args", func(t *testing.T) {
			parsed := parser.Parse(root, nil, "echo", "hello", hf.flag)
			if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
				t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
			}
		})

		t.Run(hf.name+" with named args", func(t *testing.T) {
			parsed := parser.Parse(root, nil, "echo", "{message=world}", hf.flag)
			if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
				t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
			}
		})

		t.Run(hf.name+" with abbreviation args", func(t *testing.T) {
			parsed := parser.Parse(root, nil, "echo", "m", "hello", hf.flag)
			if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
				t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
			}
		})

		t.Run(hf.name+" with sequence with args and help", func(t *testing.T) {
			parsed := parser.Parse(root, nil, "echo", "hello", ":", "echo", "world", hf.flag)
			if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
				t.Errorf("expected first segment 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
			}
		})

		t.Run(hf.name+" with global env in sequence", func(t *testing.T) {
			parsed := parser.Parse(root, nil, "{debug=true}", "echo", ":", "echo", hf.flag)
			if parsed.GlobalEnv == nil || parsed.GlobalEnv["debug"].Val != "true" {
				t.Errorf("expected global env debug=true, got %v", parsed.GlobalEnv)
			}
			if len(parsed.Cmds[0].Segments) < 2 {
				t.Fatalf("expected at least 2 segments in first cmd, got %d", len(parsed.Cmds[0].Segments))
			}
			if parsed.Cmds[0].Segments[1].Matched.Name != "desc" {
				t.Errorf("expected second segment 'desc', got %q", parsed.Cmds[0].Segments[1].Matched.Name)
			}
		})

		t.Run(hf.name+" with env containing colon", func(t *testing.T) {
			parsed := parser.Parse(root, nil, "echo", "{url=http://example.com}", hf.flag)
			if parsed.Cmds[0].Segments[0].Matched.Name != "cmd" {
				t.Errorf("expected first segment 'cmd', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
			}
		})

		t.Run(hf.name+" with multiple colons in sequence", func(t *testing.T) {
			parsed := parser.Parse(root, nil, "echo", ":", "echo", ":", "echo", hf.flag)
			if parsed.Cmds[0].Segments[0].Matched.Name != "desc" {
				t.Errorf("expected first segment 'desc', got %q", parsed.Cmds[0].Segments[0].Matched.Name)
			}
		})
	}
}
