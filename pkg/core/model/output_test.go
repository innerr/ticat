package model

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestIsJsonOutputMode(t *testing.T) {
	env := NewEnvEx(EnvLayerDefault)
	env.Set("sys.output.format", "text")
	if IsJsonOutputMode(env) {
		t.Error("expected text mode")
	}
	env.Set("sys.output.format", "json")
	if !IsJsonOutputMode(env) {
		t.Error("expected json mode")
	}
}

func TestOutputJson(t *testing.T) {
	var stdout, stderr bytes.Buffer
	screen := NewStdScreen(&stdout, &stderr)
	env := NewEnvEx(EnvLayerDefault)
	env.Set("sys.output.format", "json")
	cc := &Cli{Screen: screen}

	data := map[string]string{"key": "value", "status": "ok"}
	err := Output(cc, env, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v, got: %s", err, output)
	}
	if result["key"] != "value" || result["status"] != "ok" {
		t.Errorf("unexpected JSON content: %v", result)
	}
}

func TestOutputText(t *testing.T) {
	var stdout, stderr bytes.Buffer
	screen := NewStdScreen(&stdout, &stderr)
	env := NewEnvEx(EnvLayerDefault)
	env.Set("sys.output.format", "text")
	cc := &Cli{Screen: screen}

	data := "hello world"
	err := Output(cc, env, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if output != "hello world" {
		t.Errorf("expected 'hello world', got: %s", output)
	}
}

type testTextFormatter struct {
	val string
}

func (t testTextFormatter) FormatText() string {
	return "formatted: " + t.val + "\n"
}

func TestOutputTextFormatter(t *testing.T) {
	var stdout, stderr bytes.Buffer
	screen := NewStdScreen(&stdout, &stderr)
	env := NewEnvEx(EnvLayerDefault)
	env.Set("sys.output.format", "text")
	cc := &Cli{Screen: screen}

	data := testTextFormatter{val: "test"}
	err := Output(cc, env, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if output != "formatted: test" {
		t.Errorf("expected 'formatted: test', got: %s", output)
	}
}

func TestOutputErrorJson(t *testing.T) {
	var stdout, stderr bytes.Buffer
	screen := NewStdScreen(&stdout, &stderr)
	env := NewEnvEx(EnvLayerDefault)
	env.Set("sys.output.format", "json")
	cc := &Cli{Screen: screen}

	handled := OutputError(cc, env, "test_error", &testErr{"something failed"}, map[string]string{
		"command": "test.cmd",
	})
	if !handled {
		t.Error("expected error to be handled in json mode")
	}

	output := strings.TrimSpace(stderr.String())
	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v, got: %s", err, output)
	}
	if result["error"] != "something failed" {
		t.Errorf("unexpected error field: %v", result["error"])
	}
	if result["type"] != "test_error" {
		t.Errorf("unexpected type field: %v", result["type"])
	}
}

func TestOutputErrorTextNotHandled(t *testing.T) {
	env := NewEnvEx(EnvLayerDefault)
	env.Set("sys.output.format", "text")
	cc := &Cli{Screen: &QuietScreen{}}

	handled := OutputError(cc, env, "test_error", &testErr{"something failed"}, nil)
	if handled {
		t.Error("expected error NOT to be handled in text mode")
	}
}

func TestOutputJsonMarshalFailure(t *testing.T) {
	var stdout, stderr bytes.Buffer
	screen := NewStdScreen(&stdout, &stderr)
	env := NewEnvEx(EnvLayerDefault)
	env.Set("sys.output.format", "json")
	cc := &Cli{Screen: screen}

	// Channels cannot be marshaled to JSON
	ch := make(chan int)
	err := Output(cc, env, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have written a valid JSON error to stderr
	output := strings.TrimSpace(stderr.String())
	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("marshal-failure fallback is not valid JSON: %v, got: %s", err, output)
	}
	errorMsg, ok := result["error"].(string)
	if !ok || !strings.Contains(errorMsg, "marshal failed") {
		t.Errorf("expected 'marshal failed' in error, got: %v", result["error"])
	}

	// stdout should be empty (no data written)
	if stdout.Len() != 0 {
		t.Errorf("expected empty stdout on marshal failure, got: %s", stdout.String())
	}
}

func TestOutputErrorJsonNilDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	screen := NewStdScreen(&stdout, &stderr)
	env := NewEnvEx(EnvLayerDefault)
	env.Set("sys.output.format", "json")
	cc := &Cli{Screen: screen}

	handled := OutputError(cc, env, "unknown", &testErr{"some error"}, nil)
	if !handled {
		t.Error("expected error to be handled")
	}

	output := strings.TrimSpace(stderr.String())
	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v, got: %s", err, output)
	}
	// detail should not be present when nil is passed
	if _, exists := result["detail"]; exists {
		t.Errorf("expected no 'detail' field when nil, got: %v", result["detail"])
	}
}

type testErr struct {
	msg string
}

func (e *testErr) Error() string {
	return e.msg
}
