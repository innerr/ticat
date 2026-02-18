package model

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
)

type mockStatusFile struct {
	mu     sync.Mutex
	data   *bytes.Buffer
	closed bool
}

func newMockStatusFile() *mockStatusFile {
	return &mockStatusFile{
		data: bytes.NewBuffer(nil),
	}
}

func (f *mockStatusFile) Write(p []byte) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.data.Write(p)
}

func (f *mockStatusFile) ReadAll() []byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.data.Bytes()
}

func (f *mockStatusFile) String() string {
	return string(f.ReadAll())
}

func (f *mockStatusFile) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data = bytes.NewBuffer(nil)
}

type testFS struct {
	mu       sync.Mutex
	files    map[string]*mockStatusFile
	writeLog []string
}

func newTestFS() *testFS {
	return &testFS{
		files:    make(map[string]*mockStatusFile),
		writeLog: make([]string, 0),
	}
}

func (fs *testFS) open(path string, content string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.files[path] == nil {
		fs.files[path] = newMockStatusFile()
	}
	_, _ = fs.files[path].Write([]byte(content))
	fs.writeLog = append(fs.writeLog, path)
}

func (fs *testFS) GetContent(path string) string {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if f, ok := fs.files[path]; ok {
		return f.String()
	}
	return ""
}

func (fs *testFS) GetWriteCount(path string) int {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	count := 0
	for _, p := range fs.writeLog {
		if p == path {
			count++
		}
	}
	return count
}

func (fs *testFS) GetAllWriteCount() int {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return len(fs.writeLog)
}

func setupTestFS() *testFS {
	fs := newTestFS()
	SetStatusFileOpener(fs.open)
	return fs
}

func teardownTestFS() {
	ResetStatusFileOpener()
}

func newTestEnv() *Env {
	env := NewEnv()
	env.Set("strs.trivial-mark", "@")
	env.Set("strs.cmd-path-sep", ".")
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.env-bracket-left", "{")
	env.Set("strs.env-bracket-right", "}")
	env.Set("strs.env-kv-sep", "=")
	env.Set("strs.seq-sep", ":")
	return env.NewLayer(EnvLayerSession)
}

func newTestFlow(cmdNames ...string) *ParsedCmds {
	strs := CmdTreeStrsForTest()
	root := NewCmdTree(strs)

	var cmds []ParsedCmd
	for _, name := range cmdNames {
		sub := root.GetOrAddSub(name)
		cmds = append(cmds, ParsedCmd{
			Segments: []ParsedCmdSeg{
				{
					Matched: MatchedCmd{
						Name: name,
						Cmd:  sub,
					},
				},
			},
		})
	}
	return &ParsedCmds{Cmds: cmds}
}

func TestExecutingFlow_OnFlowStart(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	if executing == nil {
		t.Fatal("NewExecutingFlow returned nil")
	}

	content := fs.GetContent(path)
	if content == "" {
		t.Error("Status file should have content after flow start")
	}

	if !strings.Contains(content, "<flow>") {
		t.Error("Status file should contain flow marker")
	}

	if !strings.Contains(content, "<flow-start-time>") {
		t.Error("Status file should contain flow-start-time marker")
	}
}

func TestExecutingFlow_OnCmdStart(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)
	executing.OnCmdStart(flow, 0, env, "/log/path.txt")

	content := fs.GetContent(path)
	if !strings.Contains(content, "<cmd>cmd1</cmd>") {
		t.Error("Status file should contain cmd marker with command name")
	}

	if !strings.Contains(content, "<cmd-start-time>") {
		t.Error("Status file should contain cmd-start-time marker")
	}

	if !strings.Contains(content, "<log>/log/path.txt</log>") {
		t.Error("Status file should contain log path")
	}
}

func TestExecutingFlow_OnCmdFinish_Success(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)
	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	content := fs.GetContent(path)
	if !strings.Contains(content, "<cmd-result>OK</cmd-result>") {
		t.Errorf("Status file should contain OK result, got: %s", content)
	}

	if !strings.Contains(content, "<cmd-finish-time>") {
		t.Error("Status file should contain cmd-finish-time marker")
	}
}

func TestExecutingFlow_OnCmdFinish_Error(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)
	executing.OnCmdStart(flow, 0, env, "")
	testErr := fmt.Errorf("test error occurred")
	executing.OnCmdFinish(flow, 0, env, false, testErr, false)

	content := fs.GetContent(path)
	if !strings.Contains(content, "<cmd-result>ERR</cmd-result>") {
		t.Errorf("Status file should contain ERR result, got: %s", content)
	}

	if !strings.Contains(content, "test error occurred") {
		t.Error("Status file should contain error message")
	}
}

func TestExecutingFlow_OnCmdFinish_Skipped(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)
	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, true)

	content := fs.GetContent(path)
	if !strings.Contains(content, "<cmd-result>skipped</cmd-result>") {
		t.Errorf("Status file should contain skipped result, got: %s", content)
	}
}

func TestExecutingFlow_OnFlowFinish_Success(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)
	executing.OnFlowFinish(env, true)

	content := fs.GetContent(path)
	if !strings.Contains(content, "<flow-result>OK</flow-result>") {
		t.Errorf("Status file should contain OK flow result, got: %s", content)
	}
}

func TestExecutingFlow_OnFlowFinish_Error(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)
	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Errorf("Status file should contain ERR flow result, got: %s", content)
	}
}

func TestExecutingFlow_OnSubFlow(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)
	executing.OnSubFlowStart(env, "subflow content")
	executing.OnSubFlowFinish(env, true, false)

	content := fs.GetContent(path)
	if !strings.Contains(content, "<subflow>") {
		t.Error("Status file should contain subflow start marker")
	}

	if !strings.Contains(content, "</subflow>") {
		t.Error("Status file should contain subflow end marker")
	}
}

func TestExecutingFlow_CompleteFlow(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	for i := 0; i < len(flow.Cmds); i++ {
		executing.OnCmdStart(flow, i, env, "")
		executing.OnCmdFinish(flow, i, env, true, nil, false)
	}

	executing.OnFlowFinish(env, true)

	content := fs.GetContent(path)

	for _, cmd := range []string{"cmd1", "cmd2", "cmd3"} {
		if !strings.Contains(content, "<cmd>"+cmd+"</cmd>") {
			t.Errorf("Status file should contain %s", cmd)
		}
	}

	if !strings.Contains(content, "<flow-result>OK</flow-result>") {
		t.Error("Status file should end with OK result")
	}
}

func TestExecutingFlow_FlowWithError(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 1, env, "")
	executing.OnCmdFinish(flow, 1, env, false, fmt.Errorf("cmd2 failed"), false)

	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "<cmd-result>OK</cmd-result>") {
		t.Error("cmd1 should have OK result")
	}

	if !strings.Contains(content, "<cmd-result>ERR</cmd-result>") {
		t.Error("cmd2 should have ERR result")
	}

	if !strings.Contains(content, "cmd2 failed") {
		t.Error("Status file should contain error message")
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_UnlogStatus(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	env.SetBool("sys.unlog-status", true)
	flow := newTestFlow("cmd1")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	content := fs.GetContent(path)
	if content != "" {
		t.Error("Status file should be empty when sys.unlog-status is true")
	}

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)
	executing.OnFlowFinish(env, true)

	content = fs.GetContent(path)
	if content != "" {
		t.Error("Status file should remain empty when sys.unlog-status is true")
	}
}

func TestExecutingFlow_ConcurrentWrites(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3", "cmd4", "cmd5")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	var wg sync.WaitGroup
	for i := 0; i < len(flow.Cmds); i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			executing.OnCmdStart(flow, idx, env, "")
			executing.OnCmdFinish(flow, idx, env, true, nil, false)
		}(i)
	}
	wg.Wait()

	content := fs.GetContent(path)

	if strings.Contains(content, "\x00") {
		t.Error("Status file should not contain null bytes from concurrent writes")
	}
}

func TestExecutingFlow_MultiLineError(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)
	executing.OnCmdStart(flow, 0, env, "")

	multiLineErr := fmt.Errorf("error line 1\nerror line 2\nerror line 3")
	executing.OnCmdFinish(flow, 0, env, false, multiLineErr, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "error line 1") {
		t.Error("Status file should contain first error line")
	}
	if !strings.Contains(content, "error line 2") {
		t.Error("Status file should contain second error line")
	}
	if !strings.Contains(content, "error line 3") {
		t.Error("Status file should contain third error line")
	}
}

func TestExecutingFlow_WriteOrderIntegrity(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)
	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)
	executing.OnFlowFinish(env, true)

	content := fs.GetContent(path)

	flowIdx := strings.Index(content, "<flow>")
	flowStartIdx := strings.Index(content, "<flow-start-time>")
	cmdIdx := strings.Index(content, "<cmd>")
	cmdStartIdx := strings.Index(content, "<cmd-start-time>")
	cmdFinishIdx := strings.Index(content, "<cmd-finish-time>")
	cmdResultIdx := strings.Index(content, "<cmd-result>")
	flowFinishIdx := strings.Index(content, "<flow-finish-time>")
	flowResultIdx := strings.Index(content, "<flow-result>")

	if flowIdx == -1 || flowStartIdx == -1 || cmdIdx == -1 || cmdStartIdx == -1 ||
		cmdFinishIdx == -1 || cmdResultIdx == -1 || flowFinishIdx == -1 || flowResultIdx == -1 {
		t.Fatal("Missing expected markers in status file")
	}

	if !(flowIdx < flowStartIdx && flowStartIdx < cmdIdx && cmdIdx < cmdStartIdx &&
		cmdStartIdx < cmdFinishIdx && cmdFinishIdx < cmdResultIdx &&
		cmdResultIdx < flowFinishIdx && flowFinishIdx < flowResultIdx) {
		t.Error("Markers in status file are not in expected order")
	}
}

func TestParseExecutedFlow_CompleteFlow(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)
	for i := 0; i < len(flow.Cmds); i++ {
		executing.OnCmdStart(flow, i, env, "")
		executing.OnCmdFinish(flow, i, env, true, nil, false)
	}
	executing.OnFlowFinish(env, true)

	content := fs.GetContent(path)
	lines := strings.Split(content, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	parsed, remain, _, ok := parseExecutedFlow(
		ExecutedStatusFilePath{RootPath: "/test", DirName: "test", FileName: "status.txt"},
		lines, 0)

	if !ok {
		t.Errorf("Failed to parse executed flow, remaining lines: %v", remain)
	}

	if parsed == nil {
		t.Fatal("Parsed flow is nil")
	}

	if len(parsed.Cmds) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(parsed.Cmds))
	}

	if parsed.Result != ExecutedResultSucceeded {
		t.Errorf("Expected OK result, got %s", parsed.Result)
	}
}

func TestParseExecutedFlow_CorruptedRecovery(t *testing.T) {
	content := `<flow>cmd1 cmd2</flow>
<flow-start-time>2023-01-01 12:00:00</flow-start-time>
<cmd>cmd1</cmd>
<cmd-start-time>2023-01-01 12:00:01</cmd-start-time>
corrupted line without markers
another corrupted line
`

	lines := strings.Split(content, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	parsed, remain, _, ok := parseExecutedFlow(
		ExecutedStatusFilePath{RootPath: "/test", DirName: "test", FileName: "status.txt"},
		lines, 0)

	if ok {
		t.Error("Expected parsing to fail for corrupted content")
	}

	if parsed == nil {
		t.Fatal("Parsed flow should not be nil even for corrupted content")
	}

	if len(remain) == 0 {
		t.Error("Should have remaining corrupted lines")
	}

	if len(remain) < 2 {
		t.Errorf("Expected at least 2 remaining lines, got %d: %v", len(remain), remain)
	}
}

func TestParseExecutedFlow_PartialCmd(t *testing.T) {
	content := `<flow>cmd1 cmd2</flow>
<flow-start-time>2023-01-01 12:00:00</flow-start-time>
<cmd>cmd1</cmd>
<cmd-start-time>2023-01-01 12:00:01</cmd-start-time>
`

	lines := strings.Split(content, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	parsed, _, _, ok := parseExecutedFlow(
		ExecutedStatusFilePath{RootPath: "/test", DirName: "test", FileName: "status.txt"},
		lines, 0)

	if ok {
		t.Error("Expected parsing to fail for incomplete cmd")
	}

	if parsed == nil {
		t.Fatal("Parsed flow should not be nil")
	}

	if len(parsed.Cmds) != 1 {
		t.Errorf("Expected 1 command, got %d", len(parsed.Cmds))
	}

	if parsed.Cmds[0].Result != ExecutedResultIncompleted {
		t.Errorf("Expected incomplete result, got %s", parsed.Cmds[0].Result)
	}
}

func TestParseExecutedFlow_EmptyFile(t *testing.T) {
	content := ""

	lines := strings.Split(content, "\n")

	parsed, _, _, ok := parseExecutedFlow(
		ExecutedStatusFilePath{RootPath: "/test", DirName: "test", FileName: "status.txt"},
		lines, 0)

	if ok {
		t.Error("Expected parsing to fail for empty content")
	}

	if parsed == nil {
		t.Fatal("Parsed flow should not be nil even for empty content")
	}
}

func TestExecutingFlow_NestedSubFlow(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")

	executing.OnSubFlowStart(env, "subflow1")
	executing.OnSubFlowStart(env, "subflow2")

	executing.OnSubFlowFinish(env, true, false)
	executing.OnSubFlowFinish(env, true, false)

	executing.OnCmdFinish(flow, 0, env, true, nil, false)
	executing.OnFlowFinish(env, true)

	content := fs.GetContent(path)

	subflowStarts := strings.Count(content, "<subflow>")
	subflowEnds := strings.Count(content, "</subflow>")

	if subflowStarts != 2 {
		t.Errorf("Expected 2 subflow starts, got %d", subflowStarts)
	}

	if subflowEnds != 2 {
		t.Errorf("Expected 2 subflow ends, got %d", subflowEnds)
	}
}

func TestExecutingFlow_SpecialCharactersInError(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)
	executing.OnCmdStart(flow, 0, env, "")

	specialErr := fmt.Errorf("error with <special> & \"quotes\" and 'apostrophes'")
	executing.OnCmdFinish(flow, 0, env, false, specialErr, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "<special>") {
		t.Error("Special characters should be preserved in error message")
	}
}

func TestExecutingFlow_LongFlow(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()

	var cmdNames []string
	for i := 0; i < 100; i++ {
		cmdNames = append(cmdNames, fmt.Sprintf("cmd%d", i))
	}
	flow := newTestFlow(cmdNames...)
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	for i := 0; i < len(flow.Cmds); i++ {
		executing.OnCmdStart(flow, i, env, "")
		executing.OnCmdFinish(flow, i, env, true, nil, false)
	}
	executing.OnFlowFinish(env, true)

	content := fs.GetContent(path)

	for i := 0; i < 100; i++ {
		expected := fmt.Sprintf("<cmd>cmd%d</cmd>", i)
		if !strings.Contains(content, expected) {
			t.Errorf("Status file should contain %s", expected)
		}
	}
}

func TestExecutingFlow_ErrorRecovery(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 1, env, "")
	executing.OnCmdFinish(flow, 1, env, false, fmt.Errorf("cmd2 error"), false)

	func() {
		defer func() {
			_ = recover()
		}()
		executing.OnCmdStart(flow, 2, env, "")
		executing.OnCmdFinish(flow, 2, env, true, nil, false)
	}()

	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "<cmd>cmd1</cmd>") {
		t.Error("Status file should contain cmd1")
	}

	if !strings.Contains(content, "<cmd>cmd2</cmd>") {
		t.Error("Status file should contain cmd2")
	}
}

func TestExecutingFlow_AsyncTaskSchedule(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)
	executing.OnAsyncTaskSchedule(flow, 0, env, "thread-123")

	content := fs.GetContent(path)

	if !strings.Contains(content, "<scheduled>thread-123</scheduled>") {
		t.Error("Status file should contain scheduled marker with thread id")
	}
}

func TestStatusFile_IntegrationWithParser(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("test.cmd")
	path := "/test/session.123/status.txt"

	executing := NewExecutingFlow(path, flow, env)
	executing.OnCmdStart(flow, 0, env, "/var/log/cmd.log")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)
	executing.OnFlowFinish(env, true)

	content := fs.GetContent(path)
	lines := strings.Split(content, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	parsed, remain, _, ok := parseExecutedFlow(
		ExecutedStatusFilePath{RootPath: "/test/session.123", DirName: "session.123", FileName: "status.txt"},
		lines, 0)

	if !ok {
		t.Errorf("Failed to parse, remaining: %v", remain)
	}

	if len(parsed.Cmds) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(parsed.Cmds))
	}

	if parsed.Cmds[0].Cmd != "test.cmd" {
		t.Errorf("Expected cmd name 'test.cmd', got '%s'", parsed.Cmds[0].Cmd)
	}

	if parsed.Cmds[0].Result != ExecutedResultSucceeded {
		t.Errorf("Expected succeeded result, got %s", parsed.Cmds[0].Result)
	}

	if parsed.Cmds[0].LogFilePath != "/var/log/cmd.log" {
		t.Errorf("Expected log path '/var/log/cmd.log', got '%s'", parsed.Cmds[0].LogFilePath)
	}
}

func TestStatusFile_PanicDuringWrite(t *testing.T) {
	panicFS := &panicTestFS{
		files:     make(map[string]*mockStatusFile),
		writeLog:  make([]string, 0),
		panicOn:   2,
		writeCall: 0,
	}
	SetStatusFileOpener(panicFS.open)
	defer ResetStatusFileOpener()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	for i := 0; i < len(flow.Cmds); i++ {
		func() {
			defer func() {
				_ = recover()
			}()
			executing.OnCmdStart(flow, i, env, "")
			executing.OnCmdFinish(flow, i, env, true, nil, false)
		}()
	}

	content := panicFS.GetContent(path)

	if strings.Contains(content, "\x00") || strings.Contains(content, "panic") {
		t.Error("Status file should not contain corrupted data after panic")
	}
}

type panicTestFS struct {
	mu        sync.Mutex
	files     map[string]*mockStatusFile
	writeLog  []string
	panicOn   int
	writeCall int
}

func (fs *panicTestFS) open(path string, content string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.writeCall++
	if fs.writeCall == fs.panicOn {
		panic("simulated panic during write")
	}
	if fs.files[path] == nil {
		fs.files[path] = newMockStatusFile()
	}
	_, _ = fs.files[path].Write([]byte(content))
	fs.writeLog = append(fs.writeLog, path)
}

func (fs *panicTestFS) GetContent(path string) string {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if f, ok := fs.files[path]; ok {
		return f.String()
	}
	return ""
}

func TestStatusFile_AbsolutePathRequired(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for relative path")
		}
	}()

	env := newTestEnv()
	flow := newTestFlow("cmd1")

	NewExecutingFlow("relative/path/status.txt", flow, env)
}

func TestMarkedContent(t *testing.T) {
	tests := []struct {
		name     string
		mark     string
		level    int
		lines    []string
		expected []string
	}{
		{
			name:     "single_line",
			mark:     "test",
			level:    0,
			lines:    []string{"content1"},
			expected: []string{"<test>", "content1", "</test>"},
		},
		{
			name:     "multi_line",
			mark:     "test",
			level:    0,
			lines:    []string{"line1", "line2", "line3"},
			expected: []string{"<test>", "line1", "line2", "line3", "</test>"},
		},
		{
			name:     "with_indent",
			mark:     "test",
			level:    2,
			lines:    []string{"content"},
			expected: []string{"        <test>", "        content", "        </test>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := markedContent(tt.mark, tt.level, tt.lines...)
			resultLines := strings.Split(strings.TrimSuffix(result, "\n"), "\n")

			if len(resultLines) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(resultLines))
				return
			}

			for i, expected := range tt.expected {
				if resultLines[i] != expected {
					t.Errorf("Line %d: expected '%s', got '%s'", i, expected, resultLines[i])
				}
			}
		})
	}
}

func TestMarkedOneLineContent(t *testing.T) {
	tests := []struct {
		name     string
		mark     string
		level    int
		line     string
		expected string
	}{
		{
			name:     "basic",
			mark:     "cmd",
			level:    0,
			line:     "test.cmd",
			expected: "<cmd>test.cmd</cmd>",
		},
		{
			name:     "with_indent",
			mark:     "cmd",
			level:    1,
			line:     "test.cmd",
			expected: "    <cmd>test.cmd</cmd>",
		},
		{
			name:     "result_value",
			mark:     "cmd-result",
			level:    0,
			line:     "OK",
			expected: "<cmd-result>OK</cmd-result>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := markedOneLineContent(tt.mark, tt.level, tt.line)
			result = strings.TrimSuffix(result, "\n")

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestParseMarkedContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		mark        string
		level       int
		expectOk    bool
		expectLines []string
	}{
		{
			name:        "single_line_oneline_marker",
			content:     "<cmd>test.cmd</cmd>\n<cmd-start-time>12:00:00</cmd-start-time>",
			mark:        "cmd",
			level:       0,
			expectOk:    true,
			expectLines: []string{"test.cmd"},
		},
		{
			name:        "multi_line_marker",
			content:     "<error>\nline1\nline2\n</error>\n<cmd-result>ERR</cmd-result>",
			mark:        "error",
			level:       0,
			expectOk:    true,
			expectLines: []string{"line1", "line2"},
		},
		{
			name:        "not_found",
			content:     "<cmd>test</cmd>",
			mark:        "other",
			level:       0,
			expectOk:    false,
			expectLines: nil,
		},
		{
			name:        "with_indent",
			content:     "    <env-start>\n    key=value\n    </env-start>\n<cmd-result>OK</cmd-result>",
			mark:        "env-start",
			level:       1,
			expectOk:    true,
			expectLines: []string{"    key=value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			content, remain, ok := parseMarkedContent(
				ExecutedStatusFilePath{RootPath: "/test", DirName: "test", FileName: "status.txt"},
				lines,
				tt.mark,
				tt.level,
			)

			if ok != tt.expectOk {
				t.Errorf("Expected ok=%v, got %v", tt.expectOk, ok)
				return
			}

			if !tt.expectOk {
				return
			}

			if len(content) != len(tt.expectLines) {
				t.Errorf("Expected %d lines, got %d: %v", len(tt.expectLines), len(content), content)
				return
			}

			for i, expected := range tt.expectLines {
				if content[i] != expected {
					t.Errorf("Line %d: expected '%s', got '%s'", i, expected, content[i])
				}
			}

			if tt.expectOk && len(remain) == 0 && len(lines) > 1 {
				t.Error("Remaining lines should not be empty when there's more content")
			}
		})
	}
}

func TestWriteCmdEnv(t *testing.T) {
	env := newTestEnv()
	env.Set("my.key", "my.value")
	env.Set("another.key", "another.value")

	var buf bytes.Buffer
	writeCmdEnv(&buf, env, "env-start", 0)

	result := buf.String()

	if !strings.Contains(result, "<env-start>") {
		t.Error("Should contain env-start marker")
	}

	if !strings.Contains(result, "my.key=my.value") {
		t.Error("Should contain my.key=my.value")
	}

	if !strings.Contains(result, "another.key=another.value") {
		t.Error("Should contain another.key=another.value")
	}
}

func TestWriteCmdEnv_FiltersSystemKeys(t *testing.T) {
	var buf bytes.Buffer
	env := newTestEnv()
	env.Set("session.id", "123")
	env.Set("sys.path", "/usr/bin")
	env.Set("strs.sep", ".")
	env.Set("display.mode", "on")
	env.Set("user.key", "value")

	writeCmdEnv(&buf, env, "env-start", 0)

	result := buf.String()

	if strings.Contains(result, "session.id") {
		t.Error("Should filter session prefix")
	}
	if strings.Contains(result, "sys.path") {
		t.Error("Should filter sys prefix")
	}
	if strings.Contains(result, "strs.sep") {
		t.Error("Should filter strs prefix")
	}
	if strings.Contains(result, "display.mode") {
		t.Error("Should filter display prefix")
	}
	if !strings.Contains(result, "user.key=value") {
		t.Error("Should contain user.key=value")
	}
}

func BenchmarkExecutingFlow_CompleteFlow(b *testing.B) {
	fs := newTestFS()
	SetStatusFileOpener(fs.open)
	defer ResetStatusFileOpener()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3", "cmd4", "cmd5")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("/test/status_%d.txt", i)
		executing := NewExecutingFlow(path, flow, env)
		for j := 0; j < len(flow.Cmds); j++ {
			executing.OnCmdStart(flow, j, env, "")
			executing.OnCmdFinish(flow, j, env, true, nil, false)
		}
		executing.OnFlowFinish(env, true)
	}
}

func BenchmarkParseExecutedFlow_CompleteFlow(b *testing.B) {
	fs := newTestFS()
	SetStatusFileOpener(fs.open)
	defer ResetStatusFileOpener()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3", "cmd4", "cmd5")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)
	for j := 0; j < len(flow.Cmds); j++ {
		executing.OnCmdStart(flow, j, env, "")
		executing.OnCmdFinish(flow, j, env, true, nil, false)
	}
	executing.OnFlowFinish(env, true)

	content := fs.GetContent(path)
	lines := strings.Split(content, "\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = parseExecutedFlow(
			ExecutedStatusFilePath{RootPath: "/test", DirName: "test", FileName: "status.txt"},
			lines, 0)
	}
}

func TestExecutingFlow_MultiLevelSubFlow_AllSuccess(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("parent.cmd")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level1.cmd1 : level1.cmd2")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level2.cmd1 : level2.cmd2")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)
	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnSubFlowFinish(env, true, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnSubFlowFinish(env, true, false)

	executing.OnCmdFinish(flow, 0, env, true, nil, false)
	executing.OnFlowFinish(env, true)

	content := fs.GetContent(path)

	subflowCount := strings.Count(content, "<subflow>")
	subflowEndCount := strings.Count(content, "</subflow>")
	if subflowCount != 2 {
		t.Errorf("Expected 2 subflow starts, got %d", subflowCount)
	}
	if subflowEndCount != 2 {
		t.Errorf("Expected 2 subflow ends, got %d", subflowEndCount)
	}

	if !strings.Contains(content, "<flow-result>OK</flow-result>") {
		t.Error("Flow should end with OK result")
	}

	level1Idx := strings.Index(content, "level1.cmd1")
	level2Idx := strings.Index(content, "level2.cmd1")
	if level1Idx == -1 || level2Idx == -1 {
		t.Error("Both level1 and level2 commands should be in status file")
	}
	if level2Idx < level1Idx {
		t.Error("Level2 should appear after level1 (nested inside)")
	}
}

func TestExecutingFlow_MultiLevelSubFlow_ErrorInDeepestLevel(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("parent.cmd")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level1.cmd")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level2.cmd")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("deepest level error"), false)

	executing.OnSubFlowFinish(env, false, false)

	executing.OnCmdFinish(flow, 0, env, false, nil, false)

	executing.OnSubFlowFinish(env, false, false)

	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "deepest level error") {
		t.Error("Status file should contain error from deepest level")
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}

	errCount := strings.Count(content, "<error>")
	if errCount != 1 {
		t.Errorf("Expected 1 error block, got %d", errCount)
	}
}

func TestExecutingFlow_MultiLevelSubFlow_ErrorInMiddleLevel(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("parent.cmd")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level1.cmd")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level2.cmd")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnSubFlowFinish(env, true, false)

	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("middle level error"), false)

	executing.OnSubFlowFinish(env, false, false)

	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "middle level error") {
		t.Error("Status file should contain error from middle level")
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}

	subflowCount := strings.Count(content, "</subflow>")
	if subflowCount != 2 {
		t.Errorf("Expected 2 subflow ends, got %d", subflowCount)
	}
}

func TestExecutingFlow_MultiLevelSubFlow_ErrorInTopLevel(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("parent.cmd")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level1.cmd")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level2.cmd")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnSubFlowFinish(env, true, false)

	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnSubFlowFinish(env, true, false)

	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("top level error"), false)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "top level error") {
		t.Error("Status file should contain error from top level")
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}

	level2OkIdx := strings.Index(content, "level2")
	topErrIdx := strings.Index(content, "top level error")
	if level2OkIdx > topErrIdx {
		t.Error("Top level error should appear after level2 content")
	}
}

func TestExecutingFlow_PanicInCmd_DeepestLevel(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("parent.cmd")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level1.cmd")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level2.cmd")

	executing.OnCmdStart(flow, 0, env, "")

	var panicErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				switch v := r.(type) {
				case error:
					panicErr = v
				case string:
					panicErr = fmt.Errorf("%s", v)
				default:
					panicErr = fmt.Errorf("panic: %v", r)
				}
			}
		}()
		panic("simulated panic in deepest cmd")
	}()

	if panicErr == nil {
		t.Fatal("Expected panic to be caught")
	}

	executing.OnCmdFinish(flow, 0, env, false, panicErr, false)
	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "simulated panic") {
		t.Error("Status file should contain panic error message")
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result after panic")
	}
}

func TestExecutingFlow_PanicInCmd_MiddleLevel(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("parent.cmd")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level1.cmd")

	var middlePanicErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				middlePanicErr = fmt.Errorf("panic: %v", r)
			}
		}()
		panic("panic in middle level")
	}()

	executing.OnCmdFinish(flow, 0, env, false, middlePanicErr, false)

	executing.OnSubFlowFinish(env, false, false)

	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "panic in middle level") {
		t.Error("Status file should contain panic message from middle level")
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}

	subflowCount := strings.Count(content, "</subflow>")
	if subflowCount != 1 {
		t.Errorf("Expected 1 subflow end, got %d", subflowCount)
	}
}

func TestExecutingFlow_PanicRecovery_StatusFileIntegrity(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 1, env, "")

	var panicErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr = fmt.Errorf("panic: %v", r)
			}
		}()
		panic("panic during cmd2")
	}()

	executing.OnCmdFinish(flow, 1, env, false, panicErr, false)

	func() {
		defer func() {
			_ = recover()
		}()
		executing.OnCmdStart(flow, 2, env, "")
		executing.OnCmdFinish(flow, 2, env, true, nil, false)
	}()

	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	markers := []string{
		"<flow>",
		"<cmd>cmd1</cmd>",
		"<cmd-result>OK</cmd-result>",
		"<cmd>cmd2</cmd>",
		"panic during cmd2",
		"<flow-result>ERR</flow-result>",
	}

	for _, marker := range markers {
		if !strings.Contains(content, marker) {
			t.Errorf("Status file should contain '%s'", marker)
		}
	}

	if strings.Contains(content, "\x00") {
		t.Error("Status file should not contain null bytes")
	}

	lines := strings.Split(content, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	parsed, _, _, ok := parseExecutedFlow(
		ExecutedStatusFilePath{RootPath: "/test", DirName: "test", FileName: "status.txt"},
		lines, 0)

	if !ok {
		t.Error("Status file should be parseable even after panic")
	}

	if parsed == nil {
		t.Fatal("Parsed flow should not be nil")
	}

	if len(parsed.Cmds) < 2 {
		t.Errorf("Expected at least 2 commands, got %d", len(parsed.Cmds))
	}
}

func TestExecutingFlow_MultiLevelSubFlow_PartialCompletion(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("parent.cmd")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level1.cmd1 : level1.cmd2")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("level1.cmd2 failed"), false)

	executing.OnSubFlowFinish(env, false, false)

	executing.OnCmdFinish(flow, 0, env, false, nil, false)

	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "level1.cmd2 failed") {
		t.Error("Status file should contain error from level1.cmd2")
	}

	lines := strings.Split(content, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	parsed, _, _, ok := parseExecutedFlow(
		ExecutedStatusFilePath{RootPath: "/test", DirName: "test", FileName: "status.txt"},
		lines, 0)

	if !ok {
		t.Error("Status file should be parseable")
	}

	if parsed == nil {
		t.Fatal("Parsed flow should not be nil")
	}

	if parsed.Result != ExecutedResultError {
		t.Errorf("Expected ERR result, got %s", parsed.Result)
	}
}

func TestExecutingFlow_MultiLevelSubFlow_ThreeLevelsDeep(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("root")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level1")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level2")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level3")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnSubFlowFinish(env, true, false)
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnSubFlowFinish(env, true, false)
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnSubFlowFinish(env, true, false)
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnFlowFinish(env, true)

	content := fs.GetContent(path)

	subflowCount := strings.Count(content, "<subflow>")
	if subflowCount != 3 {
		t.Errorf("Expected 3 subflow starts, got %d", subflowCount)
	}

	subflowEndCount := strings.Count(content, "</subflow>")
	if subflowEndCount != 3 {
		t.Errorf("Expected 3 subflow ends, got %d", subflowEndCount)
	}

	if !strings.Contains(content, "<flow-result>OK</flow-result>") {
		t.Error("Flow should end with OK result")
	}

	lines := strings.Split(content, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	parsed, _, _, ok := parseExecutedFlow(
		ExecutedStatusFilePath{RootPath: "/test", DirName: "test", FileName: "status.txt"},
		lines, 0)

	if !ok {
		t.Error("Status file should be parseable")
	}

	if parsed == nil {
		t.Fatal("Parsed flow should not be nil")
	}
}

func TestExecutingFlow_MultiLevelSubFlow_MixedSuccessAndError(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("parent")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "subflow1")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnSubFlowFinish(env, true, false)
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "subflow2")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("subflow2.cmd2 error"), false)

	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "subflow2.cmd2 error") {
		t.Error("Status file should contain error from subflow2.cmd2")
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}

	lines := strings.Split(content, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	parsed, _, _, ok := parseExecutedFlow(
		ExecutedStatusFilePath{RootPath: "/test", DirName: "test", FileName: "status.txt"},
		lines, 0)

	if !ok {
		t.Error("Status file should be parseable")
	}

	if parsed == nil {
		t.Fatal("Parsed flow should not be nil")
	}

	if len(parsed.Cmds) < 2 {
		t.Errorf("Expected at least 2 commands in parsed flow, got %d", len(parsed.Cmds))
	}

	if parsed.Result != ExecutedResultError {
		t.Errorf("Expected ERR result, got %s", parsed.Result)
	}
}

func TestExecutingFlow_PanicWithStack_Trace(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")

	var panicErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr = fmt.Errorf("runtime error: nil pointer dereference\n    at func1()\n    at func2()\n    at main()")
			}
		}()
		panic("nil pointer")
	}()

	executing.OnCmdFinish(flow, 0, env, false, panicErr, false)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "nil pointer dereference") {
		t.Error("Status file should contain error message")
	}

	if !strings.Contains(content, "at func1()") {
		t.Error("Status file should contain stack trace line 1")
	}

	if !strings.Contains(content, "at func2()") {
		t.Error("Status file should contain stack trace line 2")
	}

	if !strings.Contains(content, "at main()") {
		t.Error("Status file should contain stack trace line 3")
	}
}

func TestExecutingFlow_MultipleErrorsInDifferentLevels(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("parent")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level1")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("error in level1 cmd1"), false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("error in level1 cmd2"), false)

	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)

	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "error in level1 cmd1") {
		t.Error("Status file should contain first error")
	}

	if !strings.Contains(content, "error in level1 cmd2") {
		t.Error("Status file should contain second error")
	}

	errorCount := strings.Count(content, "<error>")
	if errorCount != 2 {
		t.Errorf("Expected 2 error blocks, got %d", errorCount)
	}
}

func TestExecutingFlow_SingleLevel_FirstCmdError(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("first command failed"), false)

	executing.OnCmdStart(flow, 1, env, "")
	executing.OnCmdFinish(flow, 1, env, true, nil, false)

	executing.OnCmdStart(flow, 2, env, "")
	executing.OnCmdFinish(flow, 2, env, true, nil, false)

	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "<cmd>cmd1</cmd>") {
		t.Error("Status file should contain cmd1")
	}
	if !strings.Contains(content, "<cmd-result>ERR</cmd-result>") {
		t.Error("cmd1 should have ERR result")
	}
	if !strings.Contains(content, "first command failed") {
		t.Error("Status file should contain error message")
	}
	if !strings.Contains(content, "<cmd>cmd2</cmd>") {
		t.Error("Status file should contain cmd2 (subsequent commands still logged)")
	}
	if !strings.Contains(content, "<cmd>cmd3</cmd>") {
		t.Error("Status file should contain cmd3 (subsequent commands still logged)")
	}
	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_SingleLevel_MiddleCmdError(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3", "cmd4")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 1, env, "")
	executing.OnCmdFinish(flow, 1, env, false, fmt.Errorf("middle command failed"), false)

	executing.OnCmdStart(flow, 2, env, "")
	executing.OnCmdFinish(flow, 2, env, true, nil, false)

	executing.OnCmdStart(flow, 3, env, "")
	executing.OnCmdFinish(flow, 3, env, true, nil, false)

	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "<cmd>cmd1</cmd>") {
		t.Error("Status file should contain cmd1")
	}
	cmd1Result := strings.Index(content, "<cmd>cmd1</cmd>")
	cmd1Ok := strings.Index(content[cmd1Result:], "<cmd-result>OK</cmd-result>")
	if cmd1Ok == -1 {
		t.Error("cmd1 should have OK result")
	}

	if !strings.Contains(content, "<cmd>cmd2</cmd>") {
		t.Error("Status file should contain cmd2")
	}
	if !strings.Contains(content, "middle command failed") {
		t.Error("Status file should contain error message")
	}

	if !strings.Contains(content, "<cmd>cmd3</cmd>") {
		t.Error("Status file should contain cmd3")
	}
	if !strings.Contains(content, "<cmd>cmd4</cmd>") {
		t.Error("Status file should contain cmd4")
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_SingleLevel_LastCmdError(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 1, env, "")
	executing.OnCmdFinish(flow, 1, env, true, nil, false)

	executing.OnCmdStart(flow, 2, env, "")
	executing.OnCmdFinish(flow, 2, env, false, fmt.Errorf("last command failed"), false)

	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "<cmd>cmd1</cmd>") {
		t.Error("Status file should contain cmd1")
	}
	if !strings.Contains(content, "<cmd>cmd2</cmd>") {
		t.Error("Status file should contain cmd2")
	}
	if !strings.Contains(content, "<cmd>cmd3</cmd>") {
		t.Error("Status file should contain cmd3")
	}
	if !strings.Contains(content, "last command failed") {
		t.Error("Status file should contain error message")
	}

	cmd1Idx := strings.Index(content, "<cmd>cmd1</cmd>")
	cmd2Idx := strings.Index(content, "<cmd>cmd2</cmd>")
	cmd3Idx := strings.Index(content, "<cmd>cmd3</cmd>")
	errIdx := strings.Index(content, "last command failed")

	if !(cmd1Idx < cmd2Idx && cmd2Idx < cmd3Idx && cmd3Idx < errIdx) {
		t.Error("Commands and error should be in order")
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_SingleLevel_FirstCmdPanic(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")

	var panicErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr = fmt.Errorf("panic at first command: %v", r)
			}
		}()
		panic("early panic")
	}()

	executing.OnCmdFinish(flow, 0, env, false, panicErr, false)

	executing.OnCmdStart(flow, 1, env, "")
	executing.OnCmdFinish(flow, 1, env, true, nil, false)

	executing.OnCmdStart(flow, 2, env, "")
	executing.OnCmdFinish(flow, 2, env, true, nil, false)

	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "panic at first command") {
		t.Error("Status file should contain panic message")
	}

	if !strings.Contains(content, "<cmd>cmd2</cmd>") {
		t.Error("Status file should contain cmd2")
	}
	if !strings.Contains(content, "<cmd>cmd3</cmd>") {
		t.Error("Status file should contain cmd3")
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_SingleLevel_MiddleCmdPanic(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3", "cmd4")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 1, env, "")
	executing.OnCmdFinish(flow, 1, env, true, nil, false)

	executing.OnCmdStart(flow, 2, env, "")

	var panicErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr = fmt.Errorf("panic at middle command: %v", r)
			}
		}()
		panic("middle panic")
	}()

	executing.OnCmdFinish(flow, 2, env, false, panicErr, false)

	executing.OnCmdStart(flow, 3, env, "")
	executing.OnCmdFinish(flow, 3, env, true, nil, false)

	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	cmd1OkIdx := strings.Index(content, "<cmd-result>OK</cmd-result>")
	cmd2OkIdx := strings.Index(content[cmd1OkIdx+1:], "<cmd-result>OK</cmd-result>") + cmd1OkIdx + 1
	panicIdx := strings.Index(content, "panic at middle command")

	if cmd1OkIdx == -1 || cmd2OkIdx == -1 || panicIdx == -1 {
		t.Error("Expected cmd1 OK, cmd2 OK, and panic message in order")
	}

	if !(cmd1OkIdx < cmd2OkIdx && cmd2OkIdx < panicIdx) {
		t.Error("Commands should be in order: cmd1 OK, cmd2 OK, then panic")
	}

	if !strings.Contains(content, "<cmd>cmd4</cmd>") {
		t.Error("Status file should contain cmd4 (after panic)")
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_SingleLevel_LastCmdPanic(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 1, env, "")
	executing.OnCmdFinish(flow, 1, env, true, nil, false)

	executing.OnCmdStart(flow, 2, env, "")

	var panicErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr = fmt.Errorf("panic at last command: %v", r)
			}
		}()
		panic("final panic")
	}()

	executing.OnCmdFinish(flow, 2, env, false, panicErr, false)

	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "panic at last command") {
		t.Error("Status file should contain panic message")
	}

	okCount := strings.Count(content, "<cmd-result>OK</cmd-result>")
	if okCount != 2 {
		t.Errorf("Expected 2 OK results (cmd1 and cmd2), got %d", okCount)
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_SubLevel_FirstCmdError(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("parent")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "subcmd1 : subcmd2 : subcmd3")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("first subcmd failed"), false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "first subcmd failed") {
		t.Error("Status file should contain error from first subcmd")
	}

	if !strings.Contains(content, "<subflow>") {
		t.Error("Status file should contain subflow marker")
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_SubLevel_MiddleCmdError(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("parent")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "subcmd1 : subcmd2 : subcmd3 : subcmd4")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("middle subcmd failed"), false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "middle subcmd failed") {
		t.Error("Status file should contain error from middle subcmd")
	}

	subOkCount := strings.Count(content, "<cmd-result>OK</cmd-result>")
	if subOkCount < 1 {
		t.Error("Should have at least one OK result in subflow")
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_SubLevel_LastCmdError(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("parent")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "subcmd1 : subcmd2 : subcmd3")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("last subcmd failed"), false)

	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "last subcmd failed") {
		t.Error("Status file should contain error from last subcmd")
	}

	subOkCount := strings.Count(content, "<cmd-result>OK</cmd-result>")
	if subOkCount != 2 {
		t.Errorf("Expected 2 OK results in subflow, got %d", subOkCount)
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_SubLevel_FirstCmdPanic(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("parent")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "subcmd1 : subcmd2")

	executing.OnCmdStart(flow, 0, env, "")

	var panicErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr = fmt.Errorf("panic in first subcmd: %v", r)
			}
		}()
		panic("first subcmd panic")
	}()

	executing.OnCmdFinish(flow, 0, env, false, panicErr, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "panic in first subcmd") {
		t.Error("Status file should contain panic message from first subcmd")
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_SubLevel_MiddleCmdPanic(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("parent")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "subcmd1 : subcmd2 : subcmd3 : subcmd4")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 0, env, "")

	var panicErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr = fmt.Errorf("panic in middle subcmd: %v", r)
			}
		}()
		panic("middle subcmd panic")
	}()

	executing.OnCmdFinish(flow, 0, env, false, panicErr, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "panic in middle subcmd") {
		t.Error("Status file should contain panic message from middle subcmd")
	}

	subOkCount := strings.Count(content, "<cmd-result>OK</cmd-result>")
	if subOkCount < 3 {
		t.Errorf("Expected at least 3 OK results (2 before + 1 after panic), got %d", subOkCount)
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_SubLevel_LastCmdPanic(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("parent")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "subcmd1 : subcmd2 : subcmd3")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 0, env, "")

	var panicErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr = fmt.Errorf("panic in last subcmd: %v", r)
			}
		}()
		panic("last subcmd panic")
	}()

	executing.OnCmdFinish(flow, 0, env, false, panicErr, false)

	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	if !strings.Contains(content, "panic in last subcmd") {
		t.Error("Status file should contain panic message from last subcmd")
	}

	subOkCount := strings.Count(content, "<cmd-result>OK</cmd-result>")
	if subOkCount != 2 {
		t.Errorf("Expected 2 OK results before panic, got %d", subOkCount)
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_MultiLevel_FirstCmdErrorAtEachLevel(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("root")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("root level first cmd error"), false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level1")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("level1 first cmd error"), false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level2")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("level2 first cmd error"), false)

	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	errors := []string{
		"root level first cmd error",
		"level1 first cmd error",
		"level2 first cmd error",
	}

	for _, err := range errors {
		if !strings.Contains(content, err) {
			t.Errorf("Status file should contain error: %s", err)
		}
	}

	errorCount := strings.Count(content, "<error>")
	if errorCount != 3 {
		t.Errorf("Expected 3 error blocks, got %d", errorCount)
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_MultiLevel_LastCmdErrorAtEachLevel(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("root")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)
	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("root level last cmd error"), false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level1")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)
	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("level1 last cmd error"), false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level2")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)
	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("level2 last cmd error"), false)

	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	errors := []string{
		"root level last cmd error",
		"level1 last cmd error",
		"level2 last cmd error",
	}

	for _, err := range errors {
		if !strings.Contains(content, err) {
			t.Errorf("Status file should contain error: %s", err)
		}
	}

	errorCount := strings.Count(content, "<error>")
	if errorCount != 3 {
		t.Errorf("Expected 3 error blocks, got %d", errorCount)
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_MultiLevel_FirstCmdPanicAtEachLevel(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("root")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	var panicErr0 error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr0 = fmt.Errorf("root panic: %v", r)
			}
		}()
		panic("root first panic")
	}()
	executing.OnCmdFinish(flow, 0, env, false, panicErr0, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level1")

	executing.OnCmdStart(flow, 0, env, "")
	var panicErr1 error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr1 = fmt.Errorf("level1 panic: %v", r)
			}
		}()
		panic("level1 first panic")
	}()
	executing.OnCmdFinish(flow, 0, env, false, panicErr1, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level2")

	executing.OnCmdStart(flow, 0, env, "")
	var panicErr2 error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr2 = fmt.Errorf("level2 panic: %v", r)
			}
		}()
		panic("level2 first panic")
	}()
	executing.OnCmdFinish(flow, 0, env, false, panicErr2, false)

	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	panics := []string{
		"root panic",
		"level1 panic",
		"level2 panic",
	}

	for _, p := range panics {
		if !strings.Contains(content, p) {
			t.Errorf("Status file should contain panic: %s", p)
		}
	}

	errorCount := strings.Count(content, "<error>")
	if errorCount != 3 {
		t.Errorf("Expected 3 error blocks, got %d", errorCount)
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_MultiLevel_LastCmdPanicAtEachLevel(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("root")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)
	executing.OnCmdStart(flow, 0, env, "")
	var panicErr0 error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr0 = fmt.Errorf("root last panic: %v", r)
			}
		}()
		panic("root last panic")
	}()
	executing.OnCmdFinish(flow, 0, env, false, panicErr0, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level1")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)
	executing.OnCmdStart(flow, 0, env, "")
	var panicErr1 error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr1 = fmt.Errorf("level1 last panic: %v", r)
			}
		}()
		panic("level1 last panic")
	}()
	executing.OnCmdFinish(flow, 0, env, false, panicErr1, false)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnSubFlowStart(env, "level2")

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)
	executing.OnCmdStart(flow, 0, env, "")
	var panicErr2 error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr2 = fmt.Errorf("level2 last panic: %v", r)
			}
		}()
		panic("level2 last panic")
	}()
	executing.OnCmdFinish(flow, 0, env, false, panicErr2, false)

	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnSubFlowFinish(env, false, false)
	executing.OnCmdFinish(flow, 0, env, false, nil, false)
	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	panics := []string{
		"root last panic",
		"level1 last panic",
		"level2 last panic",
	}

	for _, p := range panics {
		if !strings.Contains(content, p) {
			t.Errorf("Status file should contain panic: %s", p)
		}
	}

	errorCount := strings.Count(content, "<error>")
	if errorCount != 3 {
		t.Errorf("Expected 3 error blocks, got %d", errorCount)
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_SingleLevel_AllCmdsError(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, false, fmt.Errorf("cmd1 error"), false)

	executing.OnCmdStart(flow, 1, env, "")
	executing.OnCmdFinish(flow, 1, env, false, fmt.Errorf("cmd2 error"), false)

	executing.OnCmdStart(flow, 2, env, "")
	executing.OnCmdFinish(flow, 2, env, false, fmt.Errorf("cmd3 error"), false)

	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	errors := []string{"cmd1 error", "cmd2 error", "cmd3 error"}
	for _, err := range errors {
		if !strings.Contains(content, err) {
			t.Errorf("Status file should contain error: %s", err)
		}
	}

	errorCount := strings.Count(content, "<error>")
	if errorCount != 3 {
		t.Errorf("Expected 3 error blocks, got %d", errorCount)
	}

	errResultCount := strings.Count(content, "<cmd-result>ERR</cmd-result>")
	if errResultCount != 3 {
		t.Errorf("Expected 3 ERR results, got %d", errResultCount)
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_SingleLevel_AllCmdsPanic(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	var panicErr0 error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr0 = fmt.Errorf("panic cmd1: %v", r)
			}
		}()
		panic("panic1")
	}()
	executing.OnCmdFinish(flow, 0, env, false, panicErr0, false)

	executing.OnCmdStart(flow, 1, env, "")
	var panicErr1 error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr1 = fmt.Errorf("panic cmd2: %v", r)
			}
		}()
		panic("panic2")
	}()
	executing.OnCmdFinish(flow, 1, env, false, panicErr1, false)

	executing.OnCmdStart(flow, 2, env, "")
	var panicErr2 error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr2 = fmt.Errorf("panic cmd3: %v", r)
			}
		}()
		panic("panic3")
	}()
	executing.OnCmdFinish(flow, 2, env, false, panicErr2, false)

	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	panics := []string{"panic cmd1", "panic cmd2", "panic cmd3"}
	for _, p := range panics {
		if !strings.Contains(content, p) {
			t.Errorf("Status file should contain panic: %s", p)
		}
	}

	errorCount := strings.Count(content, "<error>")
	if errorCount != 3 {
		t.Errorf("Expected 3 error blocks, got %d", errorCount)
	}

	if !strings.Contains(content, "<flow-result>ERR</flow-result>") {
		t.Error("Flow should end with ERR result")
	}
}

func TestExecutingFlow_SingleLevel_AlternatingSuccessError(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3", "cmd4", "cmd5")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 1, env, "")
	executing.OnCmdFinish(flow, 1, env, false, fmt.Errorf("cmd2 error"), false)

	executing.OnCmdStart(flow, 2, env, "")
	executing.OnCmdFinish(flow, 2, env, true, nil, false)

	executing.OnCmdStart(flow, 3, env, "")
	executing.OnCmdFinish(flow, 3, env, false, fmt.Errorf("cmd4 error"), false)

	executing.OnCmdStart(flow, 4, env, "")
	executing.OnCmdFinish(flow, 4, env, true, nil, false)

	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	okCount := strings.Count(content, "<cmd-result>OK</cmd-result>")
	errCount := strings.Count(content, "<cmd-result>ERR</cmd-result>")

	if okCount != 3 {
		t.Errorf("Expected 3 OK results (cmd1, cmd3, cmd5), got %d", okCount)
	}
	if errCount != 2 {
		t.Errorf("Expected 2 ERR results (cmd2, cmd4), got %d", errCount)
	}

	if !strings.Contains(content, "cmd2 error") {
		t.Error("Status file should contain cmd2 error")
	}
	if !strings.Contains(content, "cmd4 error") {
		t.Error("Status file should contain cmd4 error")
	}
}

func TestExecutingFlow_SingleLevel_AlternatingSuccessPanic(t *testing.T) {
	fs := setupTestFS()
	defer teardownTestFS()

	env := newTestEnv()
	flow := newTestFlow("cmd1", "cmd2", "cmd3", "cmd4", "cmd5")
	path := "/test/status.txt"

	executing := NewExecutingFlow(path, flow, env)

	executing.OnCmdStart(flow, 0, env, "")
	executing.OnCmdFinish(flow, 0, env, true, nil, false)

	executing.OnCmdStart(flow, 1, env, "")
	var panicErr1 error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr1 = fmt.Errorf("cmd2 panic: %v", r)
			}
		}()
		panic("panic2")
	}()
	executing.OnCmdFinish(flow, 1, env, false, panicErr1, false)

	executing.OnCmdStart(flow, 2, env, "")
	executing.OnCmdFinish(flow, 2, env, true, nil, false)

	executing.OnCmdStart(flow, 3, env, "")
	var panicErr3 error
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr3 = fmt.Errorf("cmd4 panic: %v", r)
			}
		}()
		panic("panic4")
	}()
	executing.OnCmdFinish(flow, 3, env, false, panicErr3, false)

	executing.OnCmdStart(flow, 4, env, "")
	executing.OnCmdFinish(flow, 4, env, true, nil, false)

	executing.OnFlowFinish(env, false)

	content := fs.GetContent(path)

	okCount := strings.Count(content, "<cmd-result>OK</cmd-result>")
	errCount := strings.Count(content, "<cmd-result>ERR</cmd-result>")

	if okCount != 3 {
		t.Errorf("Expected 3 OK results, got %d", okCount)
	}
	if errCount != 2 {
		t.Errorf("Expected 2 ERR results, got %d", errCount)
	}

	if !strings.Contains(content, "cmd2 panic") {
		t.Error("Status file should contain cmd2 panic")
	}
	if !strings.Contains(content, "cmd4 panic") {
		t.Error("Status file should contain cmd4 panic")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
