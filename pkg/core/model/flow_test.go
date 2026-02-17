package model

import (
	"bytes"
	"strings"
	"testing"
)

func TestSaveFlowToStr(t *testing.T) {
	flow := &ParsedCmds{
		Cmds: []ParsedCmd{
			{
				Segments: []ParsedCmdSeg{
					{
						Matched: MatchedCmd{Name: "cmd1"},
					},
				},
			},
		},
	}

	env := NewEnv()
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.env-bracket-left", "{")
	env.Set("strs.env-bracket-right", "}")
	env.Set("strs.env-kv-sep", "=")
	env.Set("strs.seq-sep", ":")

	result := SaveFlowToStr(flow, ".", "@", env)
	if !strings.Contains(result, "cmd1") {
		t.Errorf("Expected flow string to contain 'cmd1', got: %s", result)
	}
}

func TestSaveFlow(t *testing.T) {
	flow := &ParsedCmds{
		Cmds: []ParsedCmd{
			{
				Segments: []ParsedCmdSeg{
					{
						Matched: MatchedCmd{Name: "cmd1"},
					},
					{
						Matched: MatchedCmd{Name: "cmd2"},
					},
				},
			},
		},
	}

	env := NewEnv()
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.env-bracket-left", "{")
	env.Set("strs.env-bracket-right", "}")
	env.Set("strs.env-kv-sep", "=")
	env.Set("strs.seq-sep", ":")

	w := bytes.NewBuffer(nil)
	SaveFlow(w, flow, ".", "@", env)

	result := w.String()
	if !strings.Contains(result, "cmd1") || !strings.Contains(result, "cmd2") {
		t.Errorf("Expected flow string to contain 'cmd1' and 'cmd2', got: %s", result)
	}
}

func TestSaveFlowWithEnv(t *testing.T) {
	env := NewEnv()
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.env-bracket-left", "{")
	env.Set("strs.env-bracket-right", "}")
	env.Set("strs.env-kv-sep", "=")
	env.Set("strs.seq-sep", ":")

	parsedEnv := ParsedEnv{
		"key1": NewParsedEnvVal("key1", "value1"),
	}

	flow := &ParsedCmds{
		Cmds: []ParsedCmd{
			{
				Segments: []ParsedCmdSeg{
					{
						Matched: MatchedCmd{Name: "cmd1"},
						Env:     parsedEnv,
					},
				},
			},
		},
	}

	w := bytes.NewBuffer(nil)
	SaveFlow(w, flow, ".", "@", env)

	result := w.String()
	if !strings.Contains(result, "key1") || !strings.Contains(result, "value1") {
		t.Errorf("Expected flow string to contain env key-value, got: %s", result)
	}
}

func TestSaveFlowMultipleCmds(t *testing.T) {
	flow := &ParsedCmds{
		Cmds: []ParsedCmd{
			{
				Segments: []ParsedCmdSeg{
					{
						Matched: MatchedCmd{Name: "cmd1"},
					},
				},
			},
			{
				Segments: []ParsedCmdSeg{
					{
						Matched: MatchedCmd{Name: "cmd2"},
					},
				},
			},
		},
	}

	env := NewEnv()
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.env-bracket-left", "{")
	env.Set("strs.env-bracket-right", "}")
	env.Set("strs.env-kv-sep", "=")
	env.Set("strs.seq-sep", ":")

	w := bytes.NewBuffer(nil)
	SaveFlow(w, flow, ".", "@", env)

	result := w.String()
	expectedSubstr := "cmd1" + " " + ":" + " " + "cmd2"
	if !strings.Contains(result, expectedSubstr) {
		t.Errorf("Expected flow string to contain '%s', got: %s", expectedSubstr, result)
	}
}

func TestSaveFlowEnv(t *testing.T) {
	env := NewEnv()
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.env-bracket-left", "{")
	env.Set("strs.env-bracket-right", "}")
	env.Set("strs.env-kv-sep", "=")
	env.Set("strs.seq-sep", ":")

	parsedEnv := ParsedEnv{
		"key1": NewParsedEnvVal("key1", "value1"),
		"key2": NewParsedEnvVal("key2", "value2"),
	}

	w := bytes.NewBuffer(nil)
	result := SaveFlowEnv(w, parsedEnv, []string{"cmd"}, ".", ":", "{", "}", "=", true)

	if !result {
		t.Error("SaveFlowEnv should return true when env is not empty")
	}

	output := w.String()
	if !strings.Contains(output, "key1") || !strings.Contains(output, "key2") {
		t.Errorf("Expected output to contain env keys, got: %s", output)
	}
}

func TestSaveFlowEnvEmpty(t *testing.T) {
	w := bytes.NewBuffer(nil)
	result := SaveFlowEnv(w, nil, []string{"cmd"}, ".", ":", "{", "}", "=", true)

	if result {
		t.Error("SaveFlowEnv should return false when env is empty")
	}
}

func TestNormalizeEnvVal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		seqSep   string
		expected string
	}{
		{"no separator", "simple", ":", "simple"},
		{"single separator", "value:with", ":", "value\\\\:with"},
		{"multiple separators", "a:b:c", ":", "a\\\\:b\\\\:c"},
		{"different separator", "value|with", "|", "value\\\\|with"},
		{"already escaped", "value\\::with", ":", "value\\:\\\\:with"},
		{"empty string", "", ":", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := normalizeEnvVal(test.input, test.seqSep)
			if result != test.expected {
				t.Errorf("normalizeEnvVal(%q, %q) = %q, want %q", test.input, test.seqSep, result, test.expected)
			}
		})
	}
}

func TestNormalizeEnvValNoEscape(t *testing.T) {
	input := "simple_value"
	result := normalizeEnvVal(input, ":")
	if result != input {
		t.Errorf("Expected %s, got %s", input, result)
	}
}

func TestNormalizeEnvValWithSeparator(t *testing.T) {
	input := "value:with:separator"
	result := normalizeEnvVal(input, ":")
	expected := "value\\\\:with\\\\:separator"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFlowStrsToStr(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{"cmd1", ":", "cmd2"}, "cmd1 : cmd2"},
		{[]string{"cmd1"}, "cmd1"},
		{[]string{}, ""},
		{[]string{"cmd1", ":", "cmd2", ":", "cmd3"}, "cmd1 : cmd2 : cmd3"},
	}

	for _, test := range tests {
		result := FlowStrsToStr(test.input)
		if result != test.expected {
			t.Errorf("FlowStrsToStr(%v) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestFlowStrToStrs(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"cmd1 : cmd2 : cmd3", []string{"cmd1", ":", "cmd2", ":", "cmd3"}},
		{"cmd1", []string{"cmd1"}},
		{"", []string{}},
		{"cmd1 : cmd2", []string{"cmd1", ":", "cmd2"}},
	}

	for _, test := range tests {
		result := FlowStrToStrs(test.input)
		if len(result) != len(test.expected) {
			t.Errorf("FlowStrToStrs(%q) returned %d elements, want %d", test.input, len(result), len(test.expected))
			continue
		}
		for i, v := range result {
			if v != test.expected[i] {
				t.Errorf("FlowStrToStrs(%q)[%d] = %q, want %q", test.input, i, v, test.expected[i])
			}
		}
	}
}

func TestStripFlowForExecute(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		seqSep   string
		expected []string
	}{
		{
			name:     "basic flow",
			input:    []string{"cmd1", ":", "cmd2"},
			seqSep:   ":",
			expected: []string{"cmd1 :", ":", "cmd2"},
		},
		{
			name:     "with spaces",
			input:    []string{"  cmd1  ", ":", "  cmd2  "},
			seqSep:   ":",
			expected: []string{"  cmd1   :", ":", "  cmd2  "},
		},
		{
			name:     "with comments",
			input:    []string{"# comment", "cmd1", ":", "cmd2"},
			seqSep:   ":",
			expected: []string{"cmd1 :", ":", "cmd2"},
		},
		{
			name:     "with empty lines",
			input:    []string{"", "cmd1", "", ":", "", "cmd2", ""},
			seqSep:   ":",
			expected: []string{"cmd1 :", ":", "cmd2"},
		},
		{
			name:     "empty input",
			input:    []string{},
			seqSep:   ":",
			expected: []string{},
		},
		{
			name:     "only comments",
			input:    []string{"# comment1", "# comment2"},
			seqSep:   ":",
			expected: []string{},
		},
		{
			name:     "cmd without separator at end",
			input:    []string{"cmd1", "cmd2"},
			seqSep:   ":",
			expected: []string{"cmd1 :", "cmd2"},
		},
		{
			name:     "single cmd",
			input:    []string{"cmd1"},
			seqSep:   ":",
			expected: []string{"cmd1"},
		},
		{
			name:     "cmd ending with separator",
			input:    []string{"cmd1:", "cmd2"},
			seqSep:   ":",
			expected: []string{"cmd1:", "cmd2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := StripFlowForExecute(test.input, test.seqSep)
			if len(result) != len(test.expected) {
				t.Errorf("StripFlowForExecute(%v, %q) returned %d elements, want %d: %v",
					test.input, test.seqSep, len(result), len(test.expected), result)
				return
			}
			for i, v := range result {
				if v != test.expected[i] {
					t.Errorf("StripFlowForExecute(%v, %q)[%d] = %q, want %q",
						test.input, test.seqSep, i, v, test.expected[i])
				}
			}
		})
	}
}

func TestStripFlowForExecuteComments(t *testing.T) {
	input := []string{"# This is a comment", "cmd1", ":", "cmd2"}
	result := StripFlowForExecute(input, ":")

	for _, s := range result {
		if strings.HasPrefix(s, "#") {
			t.Errorf("Comments should be stripped, but found: %s", s)
		}
	}
}

func TestStripFlowForExecuteEmptyLines(t *testing.T) {
	input := []string{"", "cmd1", "", ":", "", "cmd2", ""}
	result := StripFlowForExecute(input, ":")

	for _, s := range result {
		if s == "" {
			t.Errorf("Empty lines should be stripped, but found empty string")
		}
	}
}

func TestSaveFlowWithTrivialLevel(t *testing.T) {
	flow := &ParsedCmds{
		Cmds: []ParsedCmd{
			{
				TrivialLvl: 2,
				Segments: []ParsedCmdSeg{
					{
						Matched: MatchedCmd{Name: "cmd1"},
					},
				},
			},
		},
	}

	env := NewEnv()
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.env-bracket-left", "{")
	env.Set("strs.env-bracket-right", "}")
	env.Set("strs.env-kv-sep", "=")
	env.Set("strs.seq-sep", ":")

	w := bytes.NewBuffer(nil)
	SaveFlow(w, flow, ".", "@", env)

	result := w.String()
	if !strings.Contains(result, "@@") {
		t.Errorf("Expected flow string to contain trivial marks '@@', got: %s", result)
	}
}

func TestSaveFlowWithParseError(t *testing.T) {
	flow := &ParsedCmds{
		Cmds: []ParsedCmd{
			{
				ParseResult: ParseResult{
					Input: []string{"cmd1"},
					Error: nil,
				},
				Segments: []ParsedCmdSeg{
					{
						Matched: MatchedCmd{Name: "cmd1"},
					},
				},
			},
		},
	}

	env := NewEnv()
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.env-bracket-left", "{")
	env.Set("strs.env-bracket-right", "}")
	env.Set("strs.env-kv-sep", "=")
	env.Set("strs.seq-sep", ":")

	w := bytes.NewBuffer(nil)
	SaveFlow(w, flow, ".", "@", env)

	result := w.String()
	if !strings.Contains(result, "cmd1") {
		t.Errorf("Expected flow string to contain 'cmd1', got: %s", result)
	}
}

func TestSaveFlowEmptyCmds(t *testing.T) {
	flow := &ParsedCmds{
		Cmds: []ParsedCmd{},
	}

	env := NewEnv()
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.env-bracket-left", "{")
	env.Set("strs.env-bracket-right", "}")
	env.Set("strs.env-kv-sep", "=")
	env.Set("strs.seq-sep", ":")

	w := bytes.NewBuffer(nil)
	SaveFlow(w, flow, ".", "@", env)

	result := w.String()
	if result != "" {
		t.Errorf("Expected empty string for empty flow, got: %q", result)
	}
}

func TestSaveFlowEnvWithSpaces(t *testing.T) {
	env := NewEnv()
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.env-bracket-left", "{")
	env.Set("strs.env-bracket-right", "}")
	env.Set("strs.env-kv-sep", "=")
	env.Set("strs.seq-sep", ":")

	parsedEnv := ParsedEnv{
		"key1": NewParsedEnvVal("key1", "value with spaces"),
	}

	w := bytes.NewBuffer(nil)
	SaveFlowEnv(w, parsedEnv, []string{"cmd"}, ".", ":", "{", "}", "=", true)

	output := w.String()
	// Values with spaces should be quoted
	if !strings.Contains(output, "\"value with spaces\"") {
		t.Errorf("Expected quoted value for spaces, got: %s", output)
	}
}

func TestSaveFlowEnvNonArg(t *testing.T) {
	env := NewEnv()
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.env-bracket-left", "{")
	env.Set("strs.env-bracket-right", "}")
	env.Set("strs.env-kv-sep", "=")
	env.Set("strs.seq-sep", ":")

	// Non-arg env value
	parsedEnv := ParsedEnv{
		"key1": NewParsedEnvVal("key1", "value1"),
	}
	// Create new value with IsArg = false
	parsedEnv["key1"] = ParsedEnvVal{
		Val:            "value1",
		IsArg:          false,
		IsSysArg:       false,
		MatchedPath:    []string{"key1"},
		MatchedPathStr: "key1",
	}

	w := bytes.NewBuffer(nil)
	result := SaveFlowEnv(w, parsedEnv, []string{"cmd"}, ".", ":", "{", "}", "=", true)

	if !result {
		t.Error("SaveFlowEnv should return true for non-arg env")
	}

	output := w.String()
	// Non-arg should use bracket format
	if !strings.Contains(output, "{") || !strings.Contains(output, "}") {
		t.Errorf("Expected bracket format for non-arg env, got: %s", output)
	}
}
