package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestReadLogFileLastLines(t *testing.T) {
	t.Run("file not exists", func(t *testing.T) {
		lines, err := ReadLogFileLastLines("/nonexistent/path/to/file.log", 1024, 10)
		if err != nil {
			t.Errorf("expected nil error for non-existent file, got: %v", err)
		}
		if lines != nil {
			t.Errorf("expected nil lines for non-existent file, got: %v", lines)
		}
	})

	t.Run("read last lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.log")
		content := "line1\nline2\nline3\nline4\nline5\n"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}

		lines, err := ReadLogFileLastLines(tmpFile, 1024, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lines) != 3 {
			t.Errorf("expected 3 lines, got %d", len(lines))
		}
		if len(lines) >= 3 && lines[2] != "line5" {
			t.Errorf("expected last line 'line5', got '%s'", lines[2])
		}
	})

	t.Run("buf size larger than file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "small.log")
		content := "line1\nline2\n"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}

		lines, err := ReadLogFileLastLines(tmpFile, 10000, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lines) != 2 {
			t.Errorf("expected 2 lines, got %d: %v", len(lines), lines)
		}
	})

	t.Run("empty lines skipped", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "empty.log")
		content := "line1\n\n\nline2\n   \n"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}

		lines, err := ReadLogFileLastLines(tmpFile, 1024, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				t.Errorf("empty line should be skipped: '%s'", line)
			}
		}
	})
}

func TestMoveFile(t *testing.T) {
	t.Run("move within same filesystem", func(t *testing.T) {
		tmpDir := t.TempDir()
		src := filepath.Join(tmpDir, "src.txt")
		dst := filepath.Join(tmpDir, "dst.txt")
		content := []byte("test content")
		if err := os.WriteFile(src, content, 0644); err != nil {
			t.Fatalf("failed to write src file: %v", err)
		}

		if err := MoveFile(src, dst); err != nil {
			t.Fatalf("MoveFile failed: %v", err)
		}

		if _, err := os.Stat(src); !os.IsNotExist(err) {
			t.Error("source file should not exist after move")
		}
		data, err := os.ReadFile(dst)
		if err != nil {
			t.Fatalf("failed to read dst file: %v", err)
		}
		if string(data) != string(content) {
			t.Errorf("content mismatch: got '%s'", string(data))
		}
	})

	t.Run("source not exists", func(t *testing.T) {
		err := MoveFile("/nonexistent/src.txt", "/tmp/dst.txt")
		if err == nil {
			t.Error("expected error for non-existent source")
		}
	})
}

func TestGoRoutineId(t *testing.T) {
	id, err := GoRoutineId()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive goroutine id, got %d", id)
	}
}

func TestGoRoutineIdStr(t *testing.T) {
	idStr := GoRoutineIdStr()
	if idStr == "" {
		t.Error("expected non-empty goroutine id string")
	}
	if idStr == "main" {
		t.Log("running in main goroutine")
	}
}

func TestRandomName(t *testing.T) {
	name1 := RandomName(10)
	name2 := RandomName(10)

	if len(name1) != 10 {
		t.Errorf("expected length 10, got %d", len(name1))
	}
	for _, c := range name1 {
		if !strings.ContainsRune(Chars, c) {
			t.Errorf("unexpected char '%c' in random name", c)
		}
	}
	if name1 == name2 {
		t.Log("warning: two random names are equal (very unlikely)")
	}
}

func TestNormalizeDurStr(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"30", "30s"},
		{"1.5", "1.5s"},
		{"100ms", "100ms"},
		{"2h", "2h"},
		{"5m30s", "5m30s"},
	}

	for _, tt := range tests {
		result := NormalizeDurStr(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeDurStr(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestQuoteStrIfHasSpace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"nospace", "nospace"},
		{"has space", `"has space"`},
		{`has "quote"`, `'has "quote"'`},
		{`has 'single' and "double"`, `has 'single' and "double"`},
	}

	for _, tt := range tests {
		result := QuoteStrIfHasSpace(tt.input)
		if result != tt.expected {
			t.Errorf("QuoteStrIfHasSpace(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestIsPidExists(t *testing.T) {
	t.Run("current process", func(t *testing.T) {
		pid := os.Getpid()
		exists, err := IsPidExists(pid)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Error("current process should exist")
		}
	})

	t.Run("non existent pid", func(t *testing.T) {
		exists, err := IsPidExists(9999999)
		if err != nil {
			t.Logf("IsPidExists returned error (expected on some systems): %v", err)
		}
		if exists {
			t.Error("non-existent pid should not exist")
		}
	})
}

func TestIsPidRunning(t *testing.T) {
	pid := os.Getpid()
	if !IsPidRunning(pid) {
		t.Error("current process should be running")
	}
}

func TestIsOsCmdExists(t *testing.T) {
	t.Run("existing command", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			if !IsOsCmdExists("cmd") {
				t.Error("cmd should exist on Windows")
			}
		} else {
			if !IsOsCmdExists("ls") {
				t.Error("ls should exist on Unix-like systems")
			}
		}
	})

	t.Run("non existent command", func(t *testing.T) {
		if IsOsCmdExists("nonexistent_command_12345") {
			t.Error("nonexistent command should not exist")
		}
	})
}

func TestFindPython(t *testing.T) {
	path := FindPython()
	if path != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("FindPython returned non-existent path: %s", path)
		}
	}
}

func TestStdoutIsPipe(t *testing.T) {
	result := StdoutIsPipe()
	if result {
		t.Log("stdout is a pipe (test may be running in piped environment)")
	} else {
		t.Log("stdout is not a pipe")
	}
}

func TestGetTerminalWidth(t *testing.T) {
	row, col := GetTerminalWidth(24, 80)
	if row <= 0 {
		t.Log("GetTerminalWidth returned default row value")
	}
	if col <= 0 {
		t.Log("GetTerminalWidth returned default col value")
	}
	t.Logf("Terminal size: row=%d, col=%d", row, col)
}

func TestIpId(t *testing.T) {
	ip := IpId()
	if ip == IpIdOnError {
		t.Log("IpId returned error state (may be expected in some environments)")
	} else if ip == IpIdOnNoNetwork {
		t.Log("IpId returned no network state")
	} else {
		parts := strings.Split(ip, ".")
		if len(parts) != 4 {
			t.Errorf("expected IPv4 address format, got: %s", ip)
		}
		for _, part := range parts {
			n, err := strconv.Atoi(part)
			if err != nil || n < 0 || n > 255 {
				t.Errorf("invalid IPv4 octet in '%s'", ip)
			}
		}
	}
}

func TestConstants(t *testing.T) {
	if GoRoutineIdStrMain != "main" {
		t.Errorf("GoRoutineIdStrMain = %q, want 'main'", GoRoutineIdStrMain)
	}
	if len(Chars) != 62 {
		t.Errorf("len(Chars) = %d, want 62", len(Chars))
	}
	if IpIdOnError == "" {
		t.Error("IpIdOnError should not be empty")
	}
	if IpIdOnNoNetwork == "" {
		t.Error("IpIdOnNoNetwork should not be empty")
	}
}
