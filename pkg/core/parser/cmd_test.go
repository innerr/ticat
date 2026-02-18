package parser

import (
	"testing"

	"github.com/innerr/ticat/pkg/core/model"
)

func newCmdTree() *model.CmdTree {
	return model.NewCmdTree(model.CmdTreeStrsForTest())
}

func newTestParser() *CmdParser {
	return &CmdParser{
		&EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"},
		".", "./", "\t ", "<root>", "^", map[byte]bool{'/': true, '\\': true},
	}
}

func TestCmdParserParseBasic(t *testing.T) {
	root := newCmdTree()
	l2 := root.AddSub("X")
	l2.AddSub("21", "twenty-one")
	parser := newTestParser()

	tests := []struct {
		name     string
		input    []string
		wantCmds []string
		wantErr  bool
	}{
		{"empty input", nil, nil, false},
		{"empty slice", []string{}, nil, false},
		{"single command", []string{"X"}, []string{"X"}, false},
		{"dot separated", []string{"X.21"}, []string{"X", "21"}, false},
		{"slash separated", []string{"X/21"}, []string{"X", "21"}, false},
		{"mixed separators", []string{"X/", ".21"}, []string{"X", "21"}, false},
		{"with empty env block", []string{"X{}21"}, []string{"X", "21"}, false},
		{"with spaces", []string{"X / . / 21"}, []string{"X", "21"}, false},
		{"with abbreviation", []string{"X.twenty-one"}, []string{"X", "twenty-one"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := parser.Parse(root, nil, tt.input)

			if tt.wantErr {
				if parsed.ParseResult.Error == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if len(parsed.Segments) != len(tt.wantCmds) {
				t.Fatalf("expected %d segments, got %d", len(tt.wantCmds), len(parsed.Segments))
			}

			for i, wantCmd := range tt.wantCmds {
				if parsed.Segments[i].Matched.Name != wantCmd {
					t.Errorf("segment %d: expected %q, got %q", i, wantCmd, parsed.Segments[i].Matched.Name)
				}
			}
		})
	}
}

func TestCmdParserParseWithEnv(t *testing.T) {
	root := newCmdTree()
	l2 := root.AddSub("X")
	l2.AddSub("21", "twenty-one")
	parser := newTestParser()

	t.Run("env before command", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"{a=V}", "X", "/", "21"})
		if len(parsed.Segments) != 3 {
			t.Fatalf("expected 3 segments, got %d", len(parsed.Segments))
		}
		if parsed.Segments[0].Env == nil || parsed.Segments[0].Env["a"].Val != "V" {
			t.Error("expected env 'a=V' in first segment")
		}
		if parsed.Segments[1].Matched.Name != "X" {
			t.Errorf("expected X in second segment, got %s", parsed.Segments[1].Matched.Name)
		}
	})

	t.Run("env after command", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"X.", "21.", "{a=V}"})
		if len(parsed.Segments) != 2 {
			t.Fatalf("expected 2 segments, got %d", len(parsed.Segments))
		}
		if parsed.Segments[1].Env == nil || parsed.Segments[1].Env["X.21.a"].Val != "V" {
			t.Error("expected env 'X.21.a=V' in last segment")
		}
	})

	t.Run("env between commands", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"X", ".{a=V}.", "21"})
		if len(parsed.Segments) != 2 {
			t.Fatalf("expected 2 segments, got %d", len(parsed.Segments))
		}
	})

	t.Run("env attached to command", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"X{a=V}.21"})
		if len(parsed.Segments) != 2 {
			t.Fatalf("expected 2 segments, got %d", len(parsed.Segments))
		}
		if parsed.Segments[0].Env == nil || parsed.Segments[0].Env["X.a"].Val != "V" {
			t.Error("expected env 'X.a=V' attached to command X")
		}
	})

	t.Run("multiple env blocks", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"{a=V}{b=V}X{c=V}21{d=V}{e=V}"})
		if len(parsed.Segments) != 3 {
			t.Fatalf("expected 3 segments, got %d", len(parsed.Segments))
		}
	})

	t.Run("empty env blocks ignored", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"{}{}X{}"})
		if len(parsed.Segments) != 1 {
			t.Fatalf("expected 1 segment (empty env blocks ignored), got %d", len(parsed.Segments))
		}
	})
}

func TestCmdParserParseWithCmdPath(t *testing.T) {
	root := newCmdTree()
	l2 := root.AddSub("X")
	l2.AddSub("21", "twenty-one")
	parser := newTestParser()

	t.Run("env gets command path prefix", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"X", "{a=V}", "21"})
		if len(parsed.Segments) != 2 {
			t.Fatalf("expected 2 segments, got %d", len(parsed.Segments))
		}
		if parsed.Segments[0].Env == nil {
			t.Fatal("expected env in first segment")
		}
		if _, ok := parsed.Segments[0].Env["X.a"]; !ok {
			t.Errorf("expected env key 'X.a', got keys: %v", getEnvKeys(parsed.Segments[0].Env))
		}
	})

	t.Run("env after full path gets path prefix", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"X", ".", "21", "{a=V}"})
		if len(parsed.Segments) < 2 {
			t.Fatalf("expected at least 2 segments, got %d", len(parsed.Segments))
		}
		lastSeg := parsed.Segments[len(parsed.Segments)-1]
		if lastSeg.Env == nil {
			t.Fatal("expected env in last segment")
		}
		if _, ok := lastSeg.Env["X.21.a"]; !ok {
			t.Errorf("expected env key 'X.21.a', got keys: %v", getEnvKeys(lastSeg.Env))
		}
	})
}

func getEnvKeys(env model.ParsedEnv) []string {
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	return keys
}

func TestCmdParserParseWithArgs(t *testing.T) {
	root := newCmdTree()
	echo := root.AddSub("echo")
	echo.RegEmptyCmd("print message").
		AddArg("message", "", "msg", "m").
		AddArg("color", "", "colour", "clr", "c")
	parser := newTestParser()

	t.Run("positional arg", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"echo", "hello"})
		assertParseResult(t, parsed, 1, "echo")
		assertEnvValue(t, parsed.Segments[0].Env, "echo.message", "hello")
	})

	t.Run("named arg", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"echo", "{message=world}"})
		assertParseResult(t, parsed, 1, "echo")
		assertEnvValue(t, parsed.Segments[0].Env, "echo.message", "world")
	})

	t.Run("two positional args", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"echo", "hi", "red"})
		assertParseResult(t, parsed, 1, "echo")
		assertEnvValue(t, parsed.Segments[0].Env, "echo.message", "hi")
		assertEnvValue(t, parsed.Segments[0].Env, "echo.color", "red")
	})

	t.Run("abbreviated arg name", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"echo", "m", "hello"})
		assertParseResult(t, parsed, 1, "echo")
		assertEnvValue(t, parsed.Segments[0].Env, "echo.message", "hello")
	})

	t.Run("all abbreviated names", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"echo", "msg", "test", "clr", "blue"})
		assertParseResult(t, parsed, 1, "echo")
		assertEnvValue(t, parsed.Segments[0].Env, "echo.message", "test")
		assertEnvValue(t, parsed.Segments[0].Env, "echo.color", "blue")
	})

	t.Run("arg value with dots", func(t *testing.T) {
		envrm := root.AddSub("envrm")
		envrm.RegEmptyCmd("env remove").AddArg("key", "")
		parsed := parser.Parse(root, nil, []string{"envrm", "bench.workload"})
		assertParseResult(t, parsed, 1, "envrm")
		assertEnvValue(t, parsed.Segments[0].Env, "envrm.key", "bench.workload")
	})

	t.Run("arg value with multiple dots", func(t *testing.T) {
		envset := root.AddSub("envset")
		envset.RegEmptyCmd("env set").AddArg("key", "").AddArg("val", "")
		parsed := parser.Parse(root, nil, []string{"envset", "db.host", "192.168.1.1"})
		assertParseResult(t, parsed, 1, "envset")
		assertEnvValue(t, parsed.Segments[0].Env, "envset.key", "db.host")
		assertEnvValue(t, parsed.Segments[0].Env, "envset.val", "192.168.1.1")
	})
}

func TestCmdParserParseWithDotsInValue(t *testing.T) {
	root := newCmdTree()
	deploy := root.AddSub("deploy")
	deploy.RegEmptyCmd("deploy command")
	parser := newTestParser()

	t.Run("dots in value (IP addresses)", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"deploy", "{hosts=172.16.5.34,172.16.5.37}"})
		assertParseResult(t, parsed, 1, "deploy")
		assertEnvValue(t, parsed.Segments[0].Env, "deploy.hosts", "172.16.5.34,172.16.5.37")
	})

	t.Run("dots in key", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"deploy", "{server.hosts=172.16.5.34}"})
		assertParseResult(t, parsed, 1, "deploy")
		assertEnvValue(t, parsed.Segments[0].Env, "deploy.server.hosts", "172.16.5.34")
	})

	t.Run("equals in value", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"deploy", "{key=a=b=c}"})
		assertParseResult(t, parsed, 1, "deploy")
		assertEnvValue(t, parsed.Segments[0].Env, "deploy.key", "a=b=c")
	})

	t.Run("URL with query params", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"deploy", "{url=http://example.com?a=1&b=2}"})
		assertParseResult(t, parsed, 1, "deploy")
		assertEnvValue(t, parsed.Segments[0].Env, "deploy.url", "http://example.com?a=1&b=2")
	})

	t.Run("single IP address", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"deploy", "{ip=192.168.1.1}"})
		assertParseResult(t, parsed, 1, "deploy")
		assertEnvValue(t, parsed.Segments[0].Env, "deploy.ip", "192.168.1.1")
	})

	t.Run("arg with IP value", func(t *testing.T) {
		echo := root.AddSub("echo")
		echo.RegEmptyCmd("echo command").AddArg("ip", "")
		parsed := parser.Parse(root, nil, []string{"echo", "{ip=172.16.5.34}"})
		assertParseResult(t, parsed, 1, "echo")
		assertEnvValue(t, parsed.Segments[0].Env, "echo.ip", "172.16.5.34")
	})
}

func TestCmdParserParseEnvEdgeCases(t *testing.T) {
	root := newCmdTree()
	deploy := root.AddSub("deploy")
	deploy.RegEmptyCmd("deploy command")
	parser := newTestParser()

	tests := []struct {
		name     string
		input    []string
		envKey   string
		envVal   string
		checkNil bool
	}{
		{"multiple comma separated IPs", []string{"deploy", "{hosts=10.0.0.1,10.0.0.2,10.0.0.3}"}, "deploy.hosts", "10.0.0.1,10.0.0.2,10.0.0.3", false},
		{"comma and equals in value", []string{"deploy", "{config=a=1,b=2}"}, "deploy.config", "a=1,b=2", false},
		{"spaces around equals", []string{"deploy", "{key = value}"}, "deploy.key", "value", false},
		{"empty value", []string{"deploy", "{key=}"}, "", "", true},
		{"colons in value", []string{"deploy", "{time=12:30:45}"}, "deploy.time", "12:30:45", false},
		{"semicolons in value", []string{"deploy", "{path=/usr/bin;/usr/local/bin}"}, "deploy.path", "/usr/bin;/usr/local/bin", false},
		{"URL with special chars", []string{"deploy", "{url=https://user:pass@host:8080/path?query=1&other=2#fragment}"}, "deploy.url", "https://user:pass@host:8080/path?query=1&other=2#fragment", false},
		{"JSON-like content", []string{"deploy", "{data={\"key\":\"value\"}}"}, "deploy.data", "{\"key\":\"value\"}", false},
		{"deeply nested key path", []string{"deploy", "{a.b.c.d.e=deep}"}, "deploy.a.b.c.d.e", "deep", false},
		{"hyphens and underscores", []string{"deploy", "{key=my-value_name}"}, "deploy.key", "my-value_name", false},
		{"numeric value", []string{"deploy", "{port=8080}"}, "deploy.port", "8080", false},
		{"slash in value", []string{"deploy", "{path=/home/user/file.txt}"}, "deploy.path", "/home/user/file.txt", false},
		{"Windows path", []string{"deploy", "{path=C:\\Users\\test\\file.txt}"}, "deploy.path", "C:\\Users\\test\\file.txt", false},
		{"IPv6 localhost", []string{"deploy", "{ip=::1}"}, "deploy.ip", "::1", false},
		{"IPv6 full address", []string{"deploy", "{ip=2001:0db8:85a3:0000:0000:8a2e:0370:7334}"}, "deploy.ip", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"email value", []string{"deploy", "{email=user@example.com}"}, "deploy.email", "user@example.com", false},
		{"database connection string", []string{"deploy", "{dsn=mysql://user:pass@localhost:3306/db?charset=utf8}"}, "deploy.dsn", "mysql://user:pass@localhost:3306/db?charset=utf8", false},
		{"math expression", []string{"deploy", "{formula=a+b=c*d}"}, "deploy.formula", "a+b=c*d", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := parser.Parse(root, nil, tt.input)
			assertParseResult(t, parsed, 1, "deploy")
			if tt.checkNil {
				if parsed.Segments[0].Env != nil && len(parsed.Segments[0].Env) > 0 {
					t.Errorf("expected no env for empty value case, got %v", parsed.Segments[0].Env)
				}
			} else {
				assertEnvValue(t, parsed.Segments[0].Env, tt.envKey, tt.envVal)
			}
		})
	}
}

func TestCmdParserParseEnvAttachedToCommand(t *testing.T) {
	root := newCmdTree()
	deploy := root.AddSub("deploy")
	deploy.RegEmptyCmd("deploy command")
	parser := newTestParser()

	t.Run("env attached without space", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"deploy{key=value}"})
		assertParseResult(t, parsed, 1, "deploy")
		assertEnvValue(t, parsed.Segments[0].Env, "deploy.key", "value")
	})

	t.Run("env after subcommand", func(t *testing.T) {
		sub := deploy.AddSub("sub")
		sub.RegEmptyCmd("sub command")
		parsed := parser.Parse(root, nil, []string{"deploy.sub{key=value}"})
		if len(parsed.Segments) != 2 {
			t.Fatalf("expected 2 segments, got %d", len(parsed.Segments))
		}
		assertEnvValue(t, parsed.Segments[1].Env, "deploy.sub.key", "value")
	})

	t.Run("attached env then separate env", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"deploy{a=1}", "{b=2}"})
		assertParseResult(t, parsed, 1, "deploy")
		if len(parsed.Segments[0].Env) != 2 {
			t.Fatalf("expected 2 env vars, got %d", len(parsed.Segments[0].Env))
		}
	})
}

func TestCmdParserParseMultipleEnvBlocks(t *testing.T) {
	root := newCmdTree()
	deploy := root.AddSub("deploy")
	deploy.RegEmptyCmd("deploy command")
	parser := newTestParser()

	t.Run("two separate env blocks", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"deploy", "{a=1}", "{b=2}"})
		assertParseResult(t, parsed, 1, "deploy")
		if len(parsed.Segments[0].Env) != 2 {
			t.Fatalf("expected 2 env vars, got %d", len(parsed.Segments[0].Env))
		}
		assertEnvValue(t, parsed.Segments[0].Env, "deploy.a", "1")
		assertEnvValue(t, parsed.Segments[0].Env, "deploy.b", "2")
	})

	t.Run("env before and after command", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"{a=1}", "deploy", "{b=2}"})
		if len(parsed.Segments) != 2 {
			t.Fatalf("expected 2 segments, got %d", len(parsed.Segments))
		}
	})

	t.Run("env with spaces in value", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"deploy", "{msg=hello world}"})
		assertParseResult(t, parsed, 1, "deploy")
		assertEnvValue(t, parsed.Segments[0].Env, "deploy.msg", "hello world")
	})
}

func TestCmdParserParseErrors(t *testing.T) {
	root := newCmdTree()
	root.AddSub("cmd")
	parser := newTestParser()

	t.Run("unknown command at root level", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"unknown"})
		if parsed.ParseResult.Error == nil {
			t.Error("expected error for unknown command")
		}
	})

	t.Run("command not found after env", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"{key=val}", "nonexistent"})
		if parsed.ParseResult.Error == nil {
			t.Error("expected error for nonexistent command after env")
		}
	})
}

func assertParseResult(t *testing.T, parsed model.ParsedCmd, wantSegments int, wantCmd string) {
	t.Helper()
	if len(parsed.Segments) != wantSegments {
		t.Fatalf("expected %d segments, got %d", wantSegments, len(parsed.Segments))
	}
	if wantSegments > 0 && wantCmd != "" {
		if parsed.Segments[0].Matched.Name != wantCmd {
			t.Errorf("expected command %q, got %q", wantCmd, parsed.Segments[0].Matched.Name)
		}
	}
}

func assertEnvValue(t *testing.T, env model.ParsedEnv, key, value string) {
	t.Helper()
	if env == nil {
		t.Fatalf("expected env to be set, got nil")
	}
	if env[key].Val != value {
		t.Errorf("expected env[%s]=%q, got %q", key, value, env[key].Val)
	}
}
