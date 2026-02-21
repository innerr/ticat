package display

import (
	"testing"
)

func TestStripAnsi(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no_ansi",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "single_color_256",
			input:    "\x1b[38;5;135mkey\x1b[0m",
			expected: "key",
		},
		{
			name:     "multiple_colors",
			input:    "\x1b[38;5;135mkey\x1b[0m \x1b[38;5;76mcmd\x1b[0m",
			expected: "key cmd",
		},
		{
			name:     "simple_color_codes",
			input:    "\x1b[31mred\x1b[0m \x1b[32mgreen\x1b[0m",
			expected: "red green",
		},
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "only_ansi_codes",
			input:    "\x1b[31m\x1b[0m",
			expected: "",
		},
		{
			name:     "complex_ansi",
			input:    "\x1b[1;31;42mbold red on green\x1b[0m",
			expected: "bold red on green",
		},
		{
			name:     "realistic_colored_line",
			input:    "\x1b[38;5;124m<FATAL>\x1b[0m\x1b[38;5;135m 'test.key'\x1b[0m",
			expected: "<FATAL> 'test.key'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripAnsi(tt.input)
			if result != tt.expected {
				t.Errorf("StripAnsi(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "plain_text",
			input:    "hello",
			expected: 5,
		},
		{
			name:     "colored_text",
			input:    "\x1b[38;5;135mkey\x1b[0m",
			expected: 3,
		},
		{
			name:     "multiple_colored_segments",
			input:    "\x1b[38;5;135mkey\x1b[0m \x1b[38;5;76mcmd\x1b[0m",
			expected: 7,
		},
		{
			name:     "fatal_colored",
			input:    "\x1b[38;5;124m<FATAL>\x1b[0m\x1b[38;5;135m 'test.key'\x1b[0m",
			expected: 18,
		},
		{
			name:     "empty_string",
			input:    "",
			expected: 0,
		},
		{
			name:     "only_ansi_codes",
			input:    "\x1b[31m\x1b[0m",
			expected: 0,
		},
		{
			name:     "mixed_content",
			input:    "prefix \x1b[31mcolored\x1b[0m suffix",
			expected: 21,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DisplayWidth(tt.input)
			if result != tt.expected {
				t.Errorf("DisplayWidth(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
