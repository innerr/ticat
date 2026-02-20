package builtin

import (
	"strings"
	"testing"

	"github.com/innerr/ticat/pkg/core/model"
)

type mockScreen struct {
	lines []string
}

func (s *mockScreen) Print(msg string) error {
	s.lines = append(s.lines, strings.TrimSpace(msg))
	return nil
}

func (s *mockScreen) Error(msg string) error {
	s.lines = append(s.lines, "ERROR: "+strings.TrimSpace(msg))
	return nil
}

func (s *mockScreen) OutputtedLines() int {
	return len(s.lines)
}

func (s *mockScreen) WriteTo(dest model.Screen) {
	for _, line := range s.lines {
		_ = dest.Print(line + "\n")
	}
}

func newTestCliForFilter() *model.Cli {
	strs := model.CmdTreeStrsForTest()
	tree := model.NewCmdTree(strs)
	return &model.Cli{
		Screen:        &model.QuietScreen{},
		Cmds:          tree,
		EnvAbbrs:      model.NewEnvAbbrs("test"),
		TolerableErrs: model.NewTolerableErrs(),
	}
}

func newTestEnvForFilter() *model.Env {
	env := model.NewEnv()
	env.Set("strs.list-sep", ",")
	return env.NewLayer(model.EnvLayerSession)
}

func newTestParsedCmdForFilter(name string) model.ParsedCmd {
	return model.ParsedCmd{
		Segments: []model.ParsedCmdSeg{
			{
				Matched: model.MatchedCmd{
					Name: name,
				},
			},
		},
	}
}

func TestAddEnvFilterPrefix(t *testing.T) {
	tests := []struct {
		name        string
		existing    string
		prefixToAdd string
		expected    string
		expectError bool
	}{
		{
			name:        "add first prefix",
			existing:    "",
			prefixToAdd: "test.",
			expected:    "test.",
		},
		{
			name:        "add second prefix",
			existing:    "foo.",
			prefixToAdd: "bar.",
			expected:    "foo.,bar.",
		},
		{
			name:        "add duplicate prefix",
			existing:    "foo.,bar.",
			prefixToAdd: "foo.",
			expected:    "foo.,bar.",
		},
		{
			name:        "add third prefix",
			existing:    "foo.,bar.",
			prefixToAdd: "baz.",
			expected:    "foo.,bar.,baz.",
		},
		{
			name:        "empty prefix should error",
			existing:    "",
			prefixToAdd: "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cc := newTestCliForFilter()
			env := newTestEnvForFilter()

			if tc.existing != "" {
				env.Set("display.env.filter.prefix", tc.existing)
			}

			argv := model.ArgVals{
				"prefix": model.ArgVal{Raw: tc.prefixToAdd, Provided: true},
			}

			flow := &model.ParsedCmds{
				Cmds: []model.ParsedCmd{newTestParsedCmdForFilter("test")},
			}

			_, err := AddEnvFilterPrefix(argv, cc, env, flow, 0)

			if tc.expectError {
				if err == nil {
					t.Error("expected error for empty prefix")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			result := env.GetRaw("display.env.filter.prefix")
			if result != tc.expected {
				t.Errorf("expected [%s], got [%s]", tc.expected, result)
			}
		})
	}
}

func TestRemoveEnvFilterPrefix(t *testing.T) {
	tests := []struct {
		name           string
		existing       string
		prefixToRemove string
		expected       string
		expectError    bool
	}{
		{
			name:           "remove from empty list",
			existing:       "",
			prefixToRemove: "foo.",
			expected:       "",
		},
		{
			name:           "remove only prefix",
			existing:       "foo.",
			prefixToRemove: "foo.",
			expected:       "",
		},
		{
			name:           "remove first of two",
			existing:       "foo.,bar.",
			prefixToRemove: "foo.",
			expected:       "bar.",
		},
		{
			name:           "remove second of two",
			existing:       "foo.,bar.",
			prefixToRemove: "bar.",
			expected:       "foo.",
		},
		{
			name:           "remove middle of three",
			existing:       "foo.,bar.,baz.",
			prefixToRemove: "bar.",
			expected:       "foo.,baz.",
		},
		{
			name:           "remove non-existent prefix",
			existing:       "foo.,bar.",
			prefixToRemove: "notexist.",
			expected:       "foo.,bar.",
		},
		{
			name:           "empty prefix should error",
			existing:       "foo.",
			prefixToRemove: "",
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cc := newTestCliForFilter()
			env := newTestEnvForFilter()

			if tc.existing != "" {
				env.Set("display.env.filter.prefix", tc.existing)
			}

			argv := model.ArgVals{
				"prefix": model.ArgVal{Raw: tc.prefixToRemove, Provided: true},
			}

			flow := &model.ParsedCmds{
				Cmds: []model.ParsedCmd{newTestParsedCmdForFilter("test")},
			}

			_, err := RemoveEnvFilterPrefix(argv, cc, env, flow, 0)

			if tc.expectError {
				if err == nil {
					t.Error("expected error for empty prefix")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			result := env.GetRaw("display.env.filter.prefix")
			if result != tc.expected {
				t.Errorf("expected [%s], got [%s]", tc.expected, result)
			}
		})
	}
}

func TestListEnvFilterPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		expected []string
	}{
		{
			name:     "empty list",
			existing: "",
			expected: nil,
		},
		{
			name:     "single prefix",
			existing: "foo.",
			expected: []string{"foo."},
		},
		{
			name:     "multiple prefixes",
			existing: "foo.,bar.,baz.",
			expected: []string{"foo.", "bar.", "baz."},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cc := newTestCliForFilter()
			env := newTestEnvForFilter()

			if tc.existing != "" {
				env.Set("display.env.filter.prefix", tc.existing)
			}

			argv := model.ArgVals{}

			flow := &model.ParsedCmds{
				Cmds: []model.ParsedCmd{newTestParsedCmdForFilter("test")},
			}

			screen := &mockScreen{}
			cc.Screen = screen

			_, err := ListEnvFilterPrefixes(argv, cc, env, flow, 0)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tc.expected == nil {
				if len(screen.lines) != 1 || !strings.Contains(screen.lines[0], "no filter") {
					t.Errorf("expected 'no filter' message, got %v", screen.lines)
				}
				return
			}

			if len(screen.lines) != len(tc.expected) {
				t.Errorf("expected %d lines, got %d", len(tc.expected), len(screen.lines))
				return
			}

			for i, expected := range tc.expected {
				if !strings.Contains(screen.lines[i], expected) {
					t.Errorf("line %d: expected to contain [%s], got [%s]", i, expected, screen.lines[i])
				}
			}
		})
	}
}

func TestClearEnvFilterPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		existing string
	}{
		{
			name:     "clear empty list",
			existing: "",
		},
		{
			name:     "clear single prefix",
			existing: "foo.",
		},
		{
			name:     "clear multiple prefixes",
			existing: "foo.,bar.,baz.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cc := newTestCliForFilter()
			env := newTestEnvForFilter()

			if tc.existing != "" {
				env.Set("display.env.filter.prefix", tc.existing)
			}

			argv := model.ArgVals{}

			flow := &model.ParsedCmds{
				Cmds: []model.ParsedCmd{newTestParsedCmdForFilter("test")},
			}

			_, err := ClearEnvFilterPrefixes(argv, cc, env, flow, 0)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			result := env.GetRaw("display.env.filter.prefix")
			if result != "" {
				t.Errorf("expected empty after clear, got [%s]", result)
			}
		})
	}
}

func TestAddRemoveSequence(t *testing.T) {
	cc := newTestCliForFilter()
	env := newTestEnvForFilter()

	argv := model.ArgVals{
		"prefix": model.ArgVal{Raw: "foo.", Provided: true},
	}
	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newTestParsedCmdForFilter("test")},
	}

	AddEnvFilterPrefix(argv, cc, env, flow, 0)
	if env.GetRaw("display.env.filter.prefix") != "foo." {
		t.Errorf("expected [foo.], got [%s]", env.GetRaw("display.env.filter.prefix"))
	}

	argv = model.ArgVals{
		"prefix": model.ArgVal{Raw: "bar.", Provided: true},
	}
	AddEnvFilterPrefix(argv, cc, env, flow, 0)
	if env.GetRaw("display.env.filter.prefix") != "foo.,bar." {
		t.Errorf("expected [foo.,bar.], got [%s]", env.GetRaw("display.env.filter.prefix"))
	}

	argv = model.ArgVals{
		"prefix": model.ArgVal{Raw: "foo.", Provided: true},
	}
	RemoveEnvFilterPrefix(argv, cc, env, flow, 0)
	if env.GetRaw("display.env.filter.prefix") != "bar." {
		t.Errorf("expected [bar.], got [%s]", env.GetRaw("display.env.filter.prefix"))
	}

	argv = model.ArgVals{}
	ClearEnvFilterPrefixes(argv, cc, env, flow, 0)
	if env.GetRaw("display.env.filter.prefix") != "" {
		t.Errorf("expected empty, got [%s]", env.GetRaw("display.env.filter.prefix"))
	}
}

func TestAddDuplicatePrefix(t *testing.T) {
	cc := newTestCliForFilter()
	env := newTestEnvForFilter()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newTestParsedCmdForFilter("test")},
	}

	argv := model.ArgVals{
		"prefix": model.ArgVal{Raw: "foo.", Provided: true},
	}
	AddEnvFilterPrefix(argv, cc, env, flow, 0)

	argv = model.ArgVals{
		"prefix": model.ArgVal{Raw: "foo.", Provided: true},
	}
	AddEnvFilterPrefix(argv, cc, env, flow, 0)

	if env.GetRaw("display.env.filter.prefix") != "foo." {
		t.Errorf("duplicate should not be added, expected [foo.], got [%s]", env.GetRaw("display.env.filter.prefix"))
	}
}
