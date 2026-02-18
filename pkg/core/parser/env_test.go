package parser

import (
	"testing"

	"github.com/innerr/ticat/pkg/core/model"
)

func newTestEnvParser() *EnvParser {
	return &EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"}
}

func parsedEnvVal(v string) model.ParsedEnvVal {
	return model.ParsedEnvVal{Val: v, IsArg: false, IsSysArg: false, MatchedPath: nil, MatchedPathStr: ""}
}

func parsedEnvArg(v string) model.ParsedEnvVal {
	return model.ParsedEnvVal{Val: v, IsArg: true, IsSysArg: false, MatchedPath: nil, MatchedPathStr: ""}
}

func TestEnvParserTryParseRawBasic(t *testing.T) {
	root := newCmdTree()
	parser := newTestEnvParser()

	tests := []struct {
		name     string
		input    []string
		wantEnv  model.ParsedEnv
		wantRest []string
	}{
		{"empty input", nil, nil, nil},
		{"empty slice", []string{}, nil, nil},
		{"single key-value", []string{"a=A"}, model.ParsedEnv{"a": parsedEnvVal("A")}, nil},
		{"two key-values", []string{"a=A", "b=B"}, model.ParsedEnv{"a": parsedEnvVal("A"), "b": parsedEnvVal("B")}, nil},
		{"key-value then plain", []string{"a=A", "bB"}, model.ParsedEnv{"a": parsedEnvVal("A")}, []string{"bB"}},
		{"spaces around equals", []string{" a = A "}, model.ParsedEnv{"a": parsedEnvVal("A")}, nil},
		{"separated tokens", []string{"a", "=", "A"}, model.ParsedEnv{"a": parsedEnvVal("A")}, nil},
		{"split at equals", []string{"a=", "A"}, model.ParsedEnv{"a": parsedEnvVal("A")}, nil},
		{"split after equals", []string{"a", "=A"}, model.ParsedEnv{"a": parsedEnvVal("A")}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEnv, gotRest := parser.TryParseRaw(root, nil, tt.input)
			if !tt.wantEnv.Equal(gotEnv) {
				t.Errorf("env mismatch: expected %v, got %v", tt.wantEnv, gotEnv)
			}
			if !sliceEqual(gotRest, tt.wantRest) {
				t.Errorf("rest mismatch: expected %v, got %v", tt.wantRest, gotRest)
			}
		})
	}
}

func TestEnvParserTryParseRawWithArgs(t *testing.T) {
	root := newCmdTree()
	dummy := func(model.ArgVals, *model.Cli, *model.Env, []model.ParsedCmd) error { return nil }
	cmd := root.RegCmd(dummy, "", "")
	cmd.AddArg("aa", "da")
	cmd.AddArg("bb", "db", "BB")

	parser := newTestEnvParser()

	t.Run("arg by name", func(t *testing.T) {
		env, _ := parser.TryParseRaw(root, nil, []string{"aa=A"})
		if env["aa"].Val != "A" || !env["aa"].IsArg {
			t.Errorf("expected arg 'aa=A', got %v", env)
		}
	})

	t.Run("arg by abbreviation", func(t *testing.T) {
		env, _ := parser.TryParseRaw(root, nil, []string{"BB=B"})
		if env["bb"].Val != "B" {
			t.Errorf("expected arg 'bb=B' via abbreviation, got %v", env)
		}
	})

	t.Run("two args by position", func(t *testing.T) {
		env, _ := parser.TryParseRaw(root, nil, []string{"A", "B"})
		if env["aa"].Val != "A" || env["bb"].Val != "B" {
			t.Errorf("expected positional args, got %v", env)
		}
	})
}

func TestEnvParserFindLeft(t *testing.T) {
	parser := newTestEnvParser()

	tests := []struct {
		name      string
		input     []string
		wantRest  []string
		wantFound bool
		wantAgain bool
	}{
		{"empty input", nil, nil, false, false},
		{"no bracket", []string{"aaa"}, []string{"aaa"}, false, false},
		{"just bracket", []string{"{"}, nil, true, false},
		{"bracket with content", []string{"{aaa"}, []string{"aaa"}, true, false},
		{"bracket not at start", []string{"aaa", "{", "bbb"}, []string{"aaa", "{", "bbb"}, false, false},
		{"bracket in middle", []string{"aa{a", "bbb"}, []string{"aa", "{", "a", "bbb"}, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRest, gotFound, gotAgain := parser.findLeft(tt.input)
			if !sliceEqual(gotRest, tt.wantRest) || gotFound != tt.wantFound || gotAgain != tt.wantAgain {
				t.Errorf("expected (%v, %v, %v), got (%v, %v, %v)",
					tt.wantRest, tt.wantFound, tt.wantAgain, gotRest, gotFound, gotAgain)
			}
		})
	}
}

func TestEnvParserFindRight(t *testing.T) {
	parser := newTestEnvParser()

	tests := []struct {
		name      string
		input     []string
		wantEnv   []string
		wantRest  []string
		wantFound bool
	}{
		{"no closing bracket", []string{"aaa"}, nil, nil, false},
		{"just closing bracket", []string{"}"}, nil, nil, true},
		{"closing bracket with rest", []string{"}", "aaa"}, nil, []string{"aaa"}, true},
		{"content before bracket", []string{"aaa", "}"}, []string{"aaa"}, nil, true},
		{"mixed content", []string{"a=A", "b=B}"}, []string{"a=A", "b=B"}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEnv, gotRest, gotFound := parser.findRight(tt.input)
			if !sliceEqual(gotEnv, tt.wantEnv) || !sliceEqual(gotRest, tt.wantRest) || gotFound != tt.wantFound {
				t.Errorf("expected (%v, %v, %v), got (%v, %v, %v)",
					tt.wantEnv, tt.wantRest, tt.wantFound, gotEnv, gotRest, gotFound)
			}
		})
	}
}

func TestEnvParserTryParse(t *testing.T) {
	root := newCmdTree()
	parser := newTestEnvParser()

	tests := []struct {
		name      string
		input     []string
		wantEnv   model.ParsedEnv
		wantRest  []string
		wantFound bool
		wantErr   bool
	}{
		{"empty", nil, nil, nil, false, false},
		{"empty brackets", []string{"{}"}, nil, nil, true, false},
		{"simple key-value", []string{"{a=A}"}, model.ParsedEnv{"a": parsedEnvVal("A")}, nil, true, false},
		{"key-value with rest", []string{"{a=A}", "bb"}, model.ParsedEnv{"a": parsedEnvVal("A")}, []string{"bb"}, true, false},
		{"two key-values", []string{"{a=A", "b=B}"}, model.ParsedEnv{"a": parsedEnvVal("A"), "b": parsedEnvVal("B")}, nil, true, false},
		{"spaces in value", []string{"{ a = A }"}, model.ParsedEnv{"a": parsedEnvVal("A")}, nil, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEnv, gotRest, gotFound, gotErr := parser.TryParse(root, nil, tt.input)
			if !tt.wantEnv.Equal(gotEnv) {
				t.Errorf("env mismatch: expected %v, got %v", tt.wantEnv, gotEnv)
			}
			if !sliceEqual(gotRest, tt.wantRest) {
				t.Errorf("rest mismatch: expected %v, got %v", tt.wantRest, gotRest)
			}
			if gotFound != tt.wantFound {
				t.Errorf("found mismatch: expected %v, got %v", tt.wantFound, gotFound)
			}
			if tt.wantErr != (gotErr != nil) {
				t.Errorf("error mismatch: expected error=%v, got %v", tt.wantErr, gotErr)
			}
		})
	}
}

func TestEnvParserSpecialCases(t *testing.T) {
	root := newCmdTree()
	parser := newTestEnvParser()

	t.Run("dots in value", func(t *testing.T) {
		env, _, found, _ := parser.TryParse(root, nil, []string{"{ip=192.168.1.1}"})
		if !found {
			t.Fatal("expected to find env block")
		}
		if env["ip"].Val != "192.168.1.1" {
			t.Errorf("expected '192.168.1.1', got %q", env["ip"].Val)
		}
	})

	t.Run("equals in value", func(t *testing.T) {
		env, _, found, _ := parser.TryParse(root, nil, []string{"{key=a=b}"})
		if !found {
			t.Fatal("expected to find env block")
		}
		if env["key"].Val != "a=b" {
			t.Errorf("expected 'a=b', got %q", env["key"].Val)
		}
	})

	t.Run("colon in value", func(t *testing.T) {
		env, _, found, _ := parser.TryParse(root, nil, []string{"{time=12:30:45}"})
		if !found {
			t.Fatal("expected to find env block")
		}
		if env["time"].Val != "12:30:45" {
			t.Errorf("expected '12:30:45', got %q", env["time"].Val)
		}
	})

	t.Run("comma in value", func(t *testing.T) {
		env, _, found, _ := parser.TryParse(root, nil, []string{"{hosts=1.1.1.1,2.2.2.2}"})
		if !found {
			t.Fatal("expected to find env block")
		}
		if env["hosts"].Val != "1.1.1.1,2.2.2.2" {
			t.Errorf("expected '1.1.1.1,2.2.2.2', got %q", env["hosts"].Val)
		}
	})
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
