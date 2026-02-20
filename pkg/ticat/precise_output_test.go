package ticat

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/innerr/ticat/pkg/core/model"
)

// ============================================================
// Helper Functions for Precise Output Verification
// ============================================================

// assertLineEquals checks if a line exactly matches expected content
func assertLineEquals(t *testing.T, actual, expected string, lineNum int) {
	t.Helper()
	if actual != expected {
		t.Errorf("line %d mismatch:\nexpected: %q\nactual:   %q", lineNum, expected, actual)
	}
}

// assertLineMatches checks if a line matches regex pattern
func assertLineMatches(t *testing.T, line, pattern string, lineNum int) {
	t.Helper()
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("invalid pattern %q: %v", pattern, err)
	}
	if !re.MatchString(line) {
		t.Errorf("line %d should match pattern %q but got: %q", lineNum, pattern, line)
	}
}

// assertLineHasExactIndent checks line has exact indent spaces before content
func assertLineHasExactIndent(t *testing.T, line string, expectedIndent int, expectedContent string) {
	t.Helper()
	indentStr := strings.Repeat(" ", expectedIndent)
	expected := indentStr + expectedContent
	if line != expected {
		t.Errorf("indent/content mismatch:\nexpected: %q\nactual:   %q", expected, line)
	}
}

// assertLineHasIndentRange checks line has indent between min and max spaces
func assertLineHasIndentRange(t *testing.T, line string, minIndent, maxIndent int) {
	t.Helper()
	indent := 0
	for _, c := range line {
		if c == ' ' {
			indent++
		} else {
			break
		}
	}
	if indent < minIndent || indent > maxIndent {
		t.Errorf("line indent %d not in range [%d, %d]: %q", indent, minIndent, maxIndent, line)
	}
}

// findLine finds a line containing substring and returns its index and content
func findLine(t *testing.T, output, contains string) (int, string) {
	t.Helper()
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if strings.Contains(line, contains) {
			return i, line
		}
	}
	t.Fatalf("line containing %q not found in output", contains)
	return -1, ""
}

// findLines finds all lines containing substring
func findLines(output, contains string) []string {
	var result []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, contains) {
			result = append(result, line)
		}
	}
	return result
}

// countLines counts lines matching pattern
func countLines(output, pattern string) int {
	re := regexp.MustCompile(pattern)
	count := 0
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if re.MatchString(line) {
			count++
		}
	}
	return count
}

// assertOutputContains checks output contains expected substring
func assertOutputContains(t *testing.T, output, expected string) {
	t.Helper()
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q\nfull output:\n%s", expected, output)
	}
}

// assertOutputNotContains checks output does not contain substring
func assertOutputNotContains(t *testing.T, output, notExpected string) {
	t.Helper()
	if strings.Contains(output, notExpected) {
		t.Errorf("output should not contain %q\nfull output:\n%s", notExpected, output)
	}
}

// assertOutputContainsExact checks output contains exact line
func assertOutputContainsExact(t *testing.T, output, exactLine string) {
	t.Helper()
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == exactLine {
			return
		}
	}
	t.Errorf("expected exact line %q not found in output", exactLine)
}

// assertSectionOrder verifies sections appear in correct order
func assertSectionOrder(t *testing.T, output string, sections []string) {
	t.Helper()
	lastIdx := -1
	for _, section := range sections {
		idx := strings.Index(output, section)
		if idx == -1 {
			t.Errorf("section %q not found in output", section)
			return
		}
		if idx <= lastIdx {
			t.Errorf("section %q appears before or at same position as previous section (at %d, previous at %d)", section, idx, lastIdx)
		}
		lastIdx = idx
	}
}

// getOutputLines splits output into non-empty lines
func getOutputLines(output string) []string {
	var result []string
	for _, line := range strings.Split(output, "\n") {
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

// ============================================================
// Precise DbgArgs Output Structure Tests
// ============================================================

func TestDbgArgsOutputStructurePrecise(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &argsCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	ok := tc.RunCli("dbg.args", "arg1=value1", "arg2=value2")
	if !ok {
		t.Fatal("command should succeed")
	}

	output := screen.GetOutput()
	lines := getOutputLines(output)

	var foundHeader, foundTailMode, foundTailModeCall, foundHasTailMode, foundCmdTailMode bool
	var foundArgsSection, foundRawInputSection, foundParsedEnvSection bool
	var headerIdx, argsIdx, rawIdx, envIdx int

	for i, line := range lines {
		switch line {
		case "=== DbgArgs Output ===":
			foundHeader = true
			headerIdx = i
		case "TailMode: false":
			foundTailMode = true
		case "TailModeCall: false":
			foundTailModeCall = true
		case "HasTailMode: false":
			foundHasTailMode = true
		case "Cmd.TailMode: false":
			foundCmdTailMode = true
		case "--- Arguments (argv) ---":
			foundArgsSection = true
			argsIdx = i
		case "--- Raw Input ---":
			foundRawInputSection = true
			rawIdx = i
		case "--- Parsed Env ---":
			foundParsedEnvSection = true
			envIdx = i
		}
	}

	if !foundHeader {
		t.Error("missing header '=== DbgArgs Output ==='")
	}
	if !foundTailMode {
		t.Error("missing 'TailMode: false'")
	}
	if !foundTailModeCall {
		t.Error("missing 'TailModeCall: false'")
	}
	if !foundHasTailMode {
		t.Error("missing 'HasTailMode: false'")
	}
	if !foundCmdTailMode {
		t.Error("missing 'Cmd.TailMode: false'")
	}
	if !foundArgsSection {
		t.Error("missing section '--- Arguments (argv) ---'")
	}
	if !foundRawInputSection {
		t.Error("missing section '--- Raw Input ---'")
	}
	if !foundParsedEnvSection {
		t.Error("missing section '--- Parsed Env ---'")
	}

	if !(headerIdx < argsIdx && argsIdx < rawIdx && rawIdx < envIdx) {
		t.Errorf("sections in wrong order: header=%d, args=%d, raw=%d, env=%d", headerIdx, argsIdx, rawIdx, envIdx)
	}
}

func TestDbgArgsOutputIndentPrecise(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &argsCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	ok := tc.RunCli("dbg.args", "arg1=test", "arg2=test2")
	if !ok {
		t.Fatal("command should succeed")
	}

	output := screen.GetOutput()
	lines := strings.Split(output, "\n")

	for i, line := range lines {
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "===") || strings.HasPrefix(line, "---") {
			assertLineHasExactIndent(t, line, 0, line)
		} else if strings.HasPrefix(line, "TailMode:") ||
			strings.HasPrefix(line, "TailModeCall:") ||
			strings.HasPrefix(line, "HasTailMode:") ||
			strings.HasPrefix(line, "Cmd.TailMode:") {
			assertLineHasExactIndent(t, line, 0, line)
		} else if strings.HasPrefix(line, "arg") ||
			strings.HasPrefix(line, "str-val") ||
			strings.HasPrefix(line, "int-val") ||
			strings.HasPrefix(line, "bool-val") ||
			strings.HasPrefix(line, "multi-abbr") {
			assertLineHasExactIndent(t, line, 2, line[2:])
		} else if regexp.MustCompile(`^\s*\[\d+\]:`).MatchString(line) {
			indent := getOutputLineIndent(line)
			if indent != 2 {
				t.Errorf("line %d: raw input should have 2-space indent, got %d: %q", i, indent, line)
			}
		} else if strings.HasPrefix(line, "env.") {
			indent := getOutputLineIndent(line)
			if indent != 2 {
				t.Errorf("line %d: parsed env should have 2-space indent, got %d: %q", i, indent, line)
			}
		}
	}
}

// getLineIndent returns number of leading spaces
func getOutputLineIndent(line string) int {
	count := 0
	for _, c := range line {
		if c == ' ' {
			count++
		} else {
			break
		}
	}
	return count
}

// ============================================================
// DbgArgs Value Format Tests
// ============================================================

func TestDbgArgsValueFormatPrecise(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		checkFn func(t *testing.T, output string)
	}{
		{
			name: "simple string value",
			args: []string{"dbg.args", "arg1=hello"},
			checkFn: func(t *testing.T, output string) {
				assertOutputContainsExact(t, output, "  arg1: raw=[hello] provided=true")
			},
		},
		{
			name: "value with equals sign",
			args: []string{"dbg.args", "arg1=db=test"},
			checkFn: func(t *testing.T, output string) {
				assertOutputContainsExact(t, output, "  arg1: raw=[db=test] provided=true")
			},
		},
		{
			name: "value with multiple equals",
			args: []string{"dbg.args", "arg1=a=b=c"},
			checkFn: func(t *testing.T, output string) {
				assertOutputContainsExact(t, output, "  arg1: raw=[a=b=c] provided=true")
			},
		},
		{
			name: "value with dots",
			args: []string{"dbg.args", "arg1=1.2.3.4"},
			checkFn: func(t *testing.T, output string) {
				assertOutputContainsExact(t, output, "  arg1: raw=[1.2.3.4] provided=true")
			},
		},
		{
			name: "value with path",
			args: []string{"dbg.args", "arg1=/path/to/file.conf"},
			checkFn: func(t *testing.T, output string) {
				assertOutputContainsExact(t, output, "  arg1: raw=[/path/to/file.conf] provided=true")
			},
		},
		{
			name: "empty value",
			args: []string{"dbg.args", "arg1="},
			checkFn: func(t *testing.T, output string) {
				assertOutputContainsExact(t, output, "  arg1: raw=[] provided=true")
			},
		},
		{
			name: "multiple args",
			args: []string{"dbg.args", "arg1=first", "arg2=second", "arg3=third"},
			checkFn: func(t *testing.T, output string) {
				assertOutputContainsExact(t, output, "  arg1: raw=[first] provided=true")
				assertOutputContainsExact(t, output, "  arg2: raw=[second] provided=true")
				assertOutputContainsExact(t, output, "  arg3: raw=[third] provided=true")
			},
		},
		{
			name: "default values shown",
			args: []string{"dbg.args"},
			checkFn: func(t *testing.T, output string) {
				assertOutputContainsExact(t, output, "  str-val: raw=[default-str] provided=false")
				assertOutputContainsExact(t, output, "  int-val: raw=[0] provided=false")
				assertOutputContainsExact(t, output, "  bool-val: raw=[false] provided=false")
			},
		},
		{
			name: "override default value",
			args: []string{"dbg.args", "str-val=custom"},
			checkFn: func(t *testing.T, output string) {
				assertOutputContainsExact(t, output, "  str-val: raw=[custom] provided=true")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ti := NewTiCatForTest()
			screen := &argsCaptureScreen{}
			ti.SetScreen(screen)
			ti.Env.SetBool("sys.panic.recover", false)

			ok := ti.RunCli(tc.args...)
			if !ok {
				t.Fatal("command should succeed")
			}

			tc.checkFn(t, screen.GetOutput())
		})
	}
}

// ============================================================
// DbgArgs Raw Input Format Tests
// ============================================================

func TestDbgArgsRawInputFormatPrecise(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "single arg",
			args:     []string{"dbg.args", "arg1=value1"},
			expected: []string{"  [0]: [dbg.args]", "  [1]: [arg1=value1]"},
		},
		{
			name:     "multiple args",
			args:     []string{"dbg.args", "arg1=a", "arg2=b"},
			expected: []string{"  [0]: [dbg.args]", "  [1]: [arg1=a]", "  [2]: [arg2=b]"},
		},
		{
			name:     "no args",
			args:     []string{"dbg.args"},
			expected: []string{"  [0]: [dbg.args]"},
		},
		{
			name:     "arg with special chars",
			args:     []string{"dbg.args", "arg1=db=test"},
			expected: []string{"  [0]: [dbg.args]", "  [1]: [arg1=db=test]"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ti := NewTiCatForTest()
			screen := &argsCaptureScreen{}
			ti.SetScreen(screen)
			ti.Env.SetBool("sys.panic.recover", false)

			ok := ti.RunCli(tc.args...)
			if !ok {
				t.Fatal("command should succeed")
			}

			output := screen.GetOutput()
			for _, expected := range tc.expected {
				assertOutputContainsExact(t, output, expected)
			}
		})
	}
}

// ============================================================
// DbgArgs Parsed Env Format Tests
// ============================================================

func TestDbgArgsParsedEnvFormatPrecise(t *testing.T) {
	ti := NewTiCatForTest()
	screen := &argsCaptureScreen{}
	ti.SetScreen(screen)
	ti.Env.SetBool("sys.panic.recover", false)

	ok := ti.RunCli("dbg.args", "arg1=value1", "arg2=value2")
	if !ok {
		t.Fatal("command should succeed")
	}

	output := screen.GetOutput()

	assertOutputContainsExact(t, output, "  env.dbg.args.arg1: [value1] isArg=true")
	assertOutputContainsExact(t, output, "  env.dbg.args.arg2: [value2] isArg=true")

	assertOutputContains(t, output, "--- Parsed Env ---")
	assertOutputContains(t, output, "env.dbg.args.arg1:")
	assertOutputContains(t, output, "env.dbg.args.arg2:")
}

// ============================================================
// DbgArgsTail Output Tests
// ============================================================

func TestDbgArgsTailOutputPrecise(t *testing.T) {
	t.Run("tail mode header structure", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("desc", ":", "dbg.args.tail", "arg1=test")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()

		assertOutputContainsExact(t, output, "=== DbgArgsTail Output ===")
		assertOutputContainsExact(t, output, "TailMode: true")
		assertOutputContainsExact(t, output, "TailModeCall: true")
		assertOutputContainsExact(t, output, "Cmd.TailMode: true")
		assertOutputContainsExact(t, output, "--- Arguments ---")
		assertOutputContainsExact(t, output, "--- Raw Input ---")
	})

	t.Run("tail mode arg format", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("desc", ":", "dbg.args.tail", "arg1=val1", "arg2=val2")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()

		assertOutputContainsExact(t, output, "  arg1: [val1] (provided=true)")
		assertOutputContainsExact(t, output, "  arg2: [val2] (provided=true)")
	})

	t.Run("tail mode raw input format", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("desc", ":", "dbg.args.tail", "arg1=test")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()

		matchingLines := findLines(output, "Input:")
		if len(matchingLines) > 0 {
			assertLineMatches(t, matchingLines[0], `^\s*Input: \[.*arg1=test.*\]$`, 0)
		}
	})

	t.Run("tail mode with special values", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("desc", ":", "dbg.args.tail", "arg1=db=x", "arg2=host=127.0.0.1")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()

		assertOutputContainsExact(t, output, "  arg1: [db=x] (provided=true)")
		assertOutputContainsExact(t, output, "  arg2: [host=127.0.0.1] (provided=true)")
	})
}

// ============================================================
// DbgArgsEnv Output Tests
// ============================================================

func TestDbgArgsEnvOutputPrecise(t *testing.T) {
	t.Run("header structure", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("dbg.args.env", "db=mysql")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()

		assertOutputContainsExact(t, output, "=== DbgArgsEnv Output ===")
		assertOutputContainsExact(t, output, "--- Arguments ---")
		assertOutputContainsExact(t, output, "--- Env Values (if auto-mapped) ---")
	})

	t.Run("arg format", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("dbg.args.env", "db=mysql", "host=localhost", "port=3306")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()

		assertOutputContainsExact(t, output, "  db: [mysql] (provided=true)")
		assertOutputContainsExact(t, output, "  host: [localhost] (provided=true)")
		assertOutputContainsExact(t, output, "  port: [3306] (provided=true)")
	})

	t.Run("section order", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("dbg.args.env", "db=test")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()

		assertSectionOrder(t, output, []string{
			"=== DbgArgsEnv Output ===",
			"--- Arguments ---",
			"--- Env Values (if auto-mapped) ---",
		})
	})
}

// ============================================================
// Echo Output Tests
// ============================================================

func TestEchoOutputPrecise(t *testing.T) {
	t.Run("simple message", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("echo", "message=hello world")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		assertOutputContainsExact(t, output, "hello world")
	})

	t.Run("message ends with newline", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("echo", "message=test")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		if !strings.HasSuffix(output, "\n") {
			t.Error("echo output should end with newline")
		}
	})

	t.Run("empty message", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("echo", "message=")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		assertOutputContainsExact(t, output, "")
	})

	t.Run("special characters", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("echo", "message=test=123")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		assertOutputContainsExact(t, output, "test=123")
	})
}

// ============================================================
// Dummy Output Tests
// ============================================================

func TestDummyOutputPrecise(t *testing.T) {
	ti := NewTiCatForTest()
	screen := &argsCaptureScreen{}
	ti.SetScreen(screen)
	ti.Env.SetBool("sys.panic.recover", false)

	ok := ti.RunCli("dummy")
	if !ok {
		t.Fatal("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContainsExact(t, output, "dummy command here")

	if !strings.HasSuffix(output, "\n") {
		t.Error("dummy output should end with newline")
	}
}

func TestNoopOutputPrecise(t *testing.T) {
	ti := NewTiCatForTest()
	screen := &argsCaptureScreen{}
	ti.SetScreen(screen)
	ti.Env.SetBool("sys.panic.recover", false)

	ok := ti.RunCli("noop")
	if !ok {
		t.Fatal("command should succeed")
	}

	output := screen.GetOutput()
	if output != "" {
		t.Errorf("noop should produce no output, got: %q", output)
	}
}

// ============================================================
// Flow Sequence Output Tests
// ============================================================

func TestFlowSequenceOutputPrecise(t *testing.T) {
	t.Run("sequence maintains output order", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("dbg.args", "arg1=first", ":", "dbg.args", "arg1=second", ":", "dbg.args", "arg1=third")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()

		firstIdx := strings.Index(output, "arg1: raw=[first]")
		secondIdx := strings.Index(output, "arg1: raw=[second]")
		thirdIdx := strings.Index(output, "arg1: raw=[third]")

		if firstIdx == -1 || secondIdx == -1 || thirdIdx == -1 {
			t.Fatal("not all values found in output")
		}

		if !(firstIdx < secondIdx && secondIdx < thirdIdx) {
			t.Errorf("values not in order: first=%d, second=%d, third=%d", firstIdx, secondIdx, thirdIdx)
		}
	})

	t.Run("sequence has three output blocks", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("dbg.args", "arg1=A", ":", "dbg.args", "arg1=B", ":", "dbg.args", "arg1=C")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()

		headerCount := strings.Count(output, "=== DbgArgs Output ===")
		if headerCount != 3 {
			t.Errorf("expected 3 header blocks, got %d", headerCount)
		}
	})

	t.Run("mixed command types in sequence", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("dbg.args", "arg1=first", ":", "dbg.args.env", "db=mysql", ":", "echo", "message=hello")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()

		assertOutputContains(t, output, "=== DbgArgs Output ===")
		assertOutputContains(t, output, "=== DbgArgsEnv Output ===")
		assertOutputContains(t, output, "hello")
	})
}

// ============================================================
// Desc Output Tests
// ============================================================

func TestDescOutputPrecise(t *testing.T) {
	t.Run("desc shows flow structure", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)
		ti.Env.SetBool("display.utf8", false)
		ti.Env.SetBool("display.color", false)

		ok := ti.RunCli("-", ":", "echo", "message=hello")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()

		assertOutputContains(t, output, "flow executing description:")
		assertOutputContains(t, output, "--->>>")
		assertOutputContains(t, output, "<<<---")
	})

	t.Run("desc shows command with brackets", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)
		ti.Env.SetBool("display.utf8", false)
		ti.Env.SetBool("display.color", false)

		ok := ti.RunCli("-", ":", "noop")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()

		assertOutputContains(t, output, "[noop]")
	})

	t.Run("desc multiple commands", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)
		ti.Env.SetBool("display.utf8", false)
		ti.Env.SetBool("display.color", false)

		ok := ti.RunCli("-", ":", "noop", ":", "noop", ":", "noop")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()

		noopCount := strings.Count(output, "[noop]")
		if noopCount < 3 {
			t.Errorf("expected at least 3 [noop] occurrences, got %d", noopCount)
		}
	})
}

// ============================================================
// Large Complex Integration Tests
// ============================================================

func TestComplexFlowOutputStructure(t *testing.T) {
	t.Run("deep nested flow with multiple arg types", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		args := []string{
			"dbg.args", "arg1=L1A", "arg2=L1B",
			":", "dbg.args.env", "db=mysql", "host=localhost",
			":", "dbg.args", "str-val=custom-string",
			":", "dbg.args.tail", "arg1=tail-value",
		}

		ok := ti.RunCli(args...)
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()

		assertOutputContains(t, output, "=== DbgArgs Output ===")
		assertOutputContains(t, output, "=== DbgArgsEnv Output ===")
		assertOutputContains(t, output, "=== DbgArgsTail Output ===")
	})

	t.Run("very deep flow with consistent structure", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		var args []string
		for i := 1; i <= 10; i++ {
			args = append(args, "dbg.args", fmt.Sprintf("arg1=step%d", i))
			if i < 10 {
				args = append(args, ":")
			}
		}

		ok := ti.RunCli(args...)
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()

		for i := 1; i <= 10; i++ {
			expected := fmt.Sprintf("arg1: raw=[step%d]", i)
			assertOutputContains(t, output, expected)
		}

		headerCount := strings.Count(output, "=== DbgArgs Output ===")
		if headerCount != 10 {
			t.Errorf("expected 10 header blocks, got %d", headerCount)
		}
	})

	t.Run("flow with env operations and verification", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli(
			"env.set", "key=test.key", "value=test-value",
			":", "dbg.args", "arg1=after-env-set",
		)
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		assertOutputContains(t, output, "arg1: raw=[after-env-set]")
	})
}

// ============================================================
// Edge Cases and Special Values
// ============================================================

func TestSpecialValuesOutputPrecise(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "unicode in value",
			args:     []string{"dbg.args", "arg1=hello世界"},
			expected: "arg1: raw=[hello世界]",
		},
		{
			name:     "very long value",
			args:     []string{"dbg.args", "arg1=" + strings.Repeat("x", 100)},
			expected: "arg1: raw=[",
		},
		{
			name:     "path with spaces in brackets",
			args:     []string{"dbg.args", "arg1={path with spaces}"},
			expected: "arg1:",
		},
		{
			name:     "url like value without port",
			args:     []string{"dbg.args", "arg1=http://localhost"},
			expected: "arg1: raw=[http://localhost]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ti := NewTiCatForTest()
			screen := &argsCaptureScreen{}
			ti.SetScreen(screen)
			ti.Env.SetBool("sys.panic.recover", false)

			ok := ti.RunCli(tc.args...)
			if !ok {
				t.Fatal("command should succeed")
			}

			output := screen.GetOutput()
			assertOutputContains(t, output, tc.expected)
		})
	}
}

// ============================================================
// Output Line Count Tests
// ============================================================

func TestOutputLineCountPrecise(t *testing.T) {
	t.Run("single dbg.args output lines", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("dbg.args", "arg1=value1")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		lines := getOutputLines(output)

		if len(lines) < 10 {
			t.Errorf("expected at least 10 lines for dbg.args output, got %d", len(lines))
		}
	})

	t.Run("sequence doubles output blocks", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("dbg.args", ":", "dbg.args")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		headerCount := strings.Count(output, "=== DbgArgs Output ===")

		if headerCount != 2 {
			t.Errorf("expected 2 header blocks for sequence, got %d", headerCount)
		}
	})
}

// ============================================================
// Whitespace and Formatting Tests
// ============================================================

func TestWhitespaceFormattingPrecise(t *testing.T) {
	t.Run("no trailing whitespace on lines", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("dbg.args", "arg1=test")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		lines := strings.Split(output, "\n")

		for i, line := range lines {
			if line != strings.TrimRight(line, " \t") {
				t.Errorf("line %d has trailing whitespace: %q", i, line)
			}
		}
	})

	t.Run("no consecutive blank lines", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("dbg.args", "arg1=test")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		lines := strings.Split(output, "\n")

		for i := 1; i < len(lines); i++ {
			if lines[i] == "" && lines[i-1] == "" {
				t.Errorf("consecutive blank lines at line %d", i)
			}
		}
	})

	t.Run("consistent indent within sections", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("dbg.args", "arg1=test", "arg2=test2", "arg3=test3")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		lines := strings.Split(output, "\n")

		var inArgsSection bool
		var expectedIndent int = -1

		for _, line := range lines {
			if line == "--- Arguments (argv) ---" {
				inArgsSection = true
				continue
			}
			if inArgsSection && strings.HasPrefix(line, "---") {
				inArgsSection = false
			}
			if inArgsSection && strings.HasPrefix(line, "  arg") {
				indent := getOutputLineIndent(line)
				if expectedIndent == -1 {
					expectedIndent = indent
				} else if indent != expectedIndent {
					t.Errorf("inconsistent indent in args section: expected %d, got %d in line: %q", expectedIndent, indent, line)
				}
			}
		}
	})
}

// ============================================================
// Abbreviation Tests
// ============================================================

func TestAbbreviationOutputPrecise(t *testing.T) {
	t.Run("arg abbreviation expands correctly", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("dbg.args", "a1=abbr-value")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		assertOutputContains(t, output, "arg1: raw=[abbr-value]")
	})

	t.Run("str-val abbreviation expands correctly", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("dbg.args", "str=string-value")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		assertOutputContains(t, output, "str-val: raw=[string-value]")
	})

	t.Run("multiple abbreviations in flow", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli("dbg.args", "a1=val1", ":", "dbg.args", "a2=val2")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		assertOutputContains(t, output, "arg1: raw=[val1]")
		assertOutputContains(t, output, "arg2: raw=[val2]")
	})
}

// ============================================================
// Env Propagation in Flow Tests
// ============================================================

func TestEnvPropagationOutputPrecise(t *testing.T) {
	t.Run("env set visible in next command", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli(
			"env.set", "key=propagation.test", "value=propagated-value",
			":", "env", "propagation.test",
		)
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		assertOutputContains(t, output, "propagated-value")
	})

	t.Run("multiple env sets in sequence", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)

		ok := ti.RunCli(
			"env.set", "key=multi.a", "value=1",
			":", "env.set", "key=multi.b", "value=2",
			":", "env.set", "key=multi.c", "value=3",
			":", "env", "multi",
		)
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		assertOutputContains(t, output, "multi.a")
		assertOutputContains(t, output, "multi.b")
		assertOutputContains(t, output, "multi.c")
	})
}

// ============================================================
// Global Env in Sequence Tests
// ============================================================

func TestGlobalEnvOutputPrecise(t *testing.T) {
	t.Run("global env shows in desc output", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)
		ti.Env.SetBool("display.utf8", false)
		ti.Env.SetBool("display.color", false)

		ok := ti.RunCli("-", ":", "echo", "message=test")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		assertOutputContains(t, output, "flow executing description:")
		assertOutputContains(t, output, "[echo]")
	})

	t.Run("global env with multiple commands", func(t *testing.T) {
		ti := NewTiCatForTest()
		screen := &argsCaptureScreen{}
		ti.SetScreen(screen)
		ti.Env.SetBool("sys.panic.recover", false)
		ti.Env.SetBool("display.utf8", false)
		ti.Env.SetBool("display.color", false)

		ok := ti.RunCli("-",
			":", "noop",
			":", "noop")
		if !ok {
			t.Fatal("command should succeed")
		}

		output := screen.GetOutput()
		assertOutputContains(t, output, "flow executing description:")
	})
}

// ============================================================
// Session Layer Persistence Tests
// ============================================================

func TestSessionLayerPersistence(t *testing.T) {
	ti := NewTiCatForTest()
	screen := &argsCaptureScreen{}
	ti.SetScreen(screen)
	ti.Env.SetBool("sys.panic.recover", false)

	ok := ti.RunCli("env.set", "key=session.persistent", "value=kept")
	if !ok {
		t.Fatal("command should succeed")
	}

	sessionEnv := ti.Env.GetLayer(model.EnvLayerSession)
	if sessionEnv.Get("session.persistent").Raw != "kept" {
		t.Error("session layer should contain the set value")
	}
}
