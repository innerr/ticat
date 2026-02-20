package display

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/innerr/ticat/pkg/core/model"
	"github.com/innerr/ticat/pkg/core/parser"
)

type memoryScreen struct {
	output bytes.Buffer
}

func (s *memoryScreen) Print(text string) error {
	s.output.WriteString(text)
	return nil
}

func (s *memoryScreen) Error(text string) error {
	s.output.WriteString(text)
	return nil
}

func (s *memoryScreen) OutputtedLines() int {
	return strings.Count(s.output.String(), "\n")
}

func (s *memoryScreen) GetOutput() string {
	return s.output.String()
}

func (s *memoryScreen) Reset() {
	s.output.Reset()
}

func setupTestEnv(env *model.Env) {
	env.SetBool("display.utf8", false)
	env.SetBool("display.color", false)
	env.SetBool("sys.panic.recover", true)
	env.SetInt("display.width", 120)
}

func createTestEnv() *model.Env {
	env := model.NewEnvEx(model.EnvLayerDefault).NewLayer(model.EnvLayerSession)
	env.SetBool("display.utf8", false)
	env.SetBool("display.color", false)
	env.SetBool("sys.panic.recover", true)
	env.SetInt("display.width", 120)
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.env-bracket-left", "{")
	env.Set("strs.env-bracket-right", "}")
	env.Set("strs.env-kv-sep", "=")
	env.Set("strs.seq-sep", ":")
	env.Set("strs.trivial-mark", "@")
	env.Set("strs.list-sep", ",")
	env.Set("strs.env-op-sep", "|")
	env.Set("strs.env-del-all-mark", "-")
	return env
}

func createTestCmdTree() *model.CmdTree {
	strs := model.CmdTreeStrsForTest()
	tree := model.NewCmdTree(strs)
	return tree
}

func createTestCli(screen model.Screen, tree *model.CmdTree) *model.Cli {
	envParser := parser.NewEnvParser(
		parser.Brackets{Left: "{", Right: "}"},
		" ",
		"=",
		".",
		"sys.",
	)
	cmdParser := parser.NewCmdParser(
		envParser,
		".",
		"",
		" ",
		"ticat",
		"@",
		"",
	)
	cliParser := parser.NewParser(
		parser.NewSequenceParser(":", []string{"http", "https"}, nil),
		cmdParser,
	)
	abbrs := model.NewEnvAbbrs("ticat")
	envKeysInfo := model.NewEnvKeysInfo()
	cmdIO := model.NewCmdIO(nil, nil, nil)
	return model.NewCli(screen, tree, cliParser, abbrs, cmdIO, envKeysInfo)
}

func newParsedCmdSeg(name string, cmd *model.CmdTree) model.ParsedCmdSeg {
	return model.ParsedCmdSeg{
		Matched: model.MatchedCmd{
			Name: name,
			Cmd:  cmd,
		},
	}
}

func newParsedCmd(name string, cmd *model.CmdTree) model.ParsedCmd {
	return model.ParsedCmd{
		Segments: []model.ParsedCmdSeg{newParsedCmdSeg(name, cmd)},
	}
}

func newParsedCmdWithEnv(name string, cmd *model.CmdTree, env model.ParsedEnv) model.ParsedCmd {
	return model.ParsedCmd{
		Segments: []model.ParsedCmdSeg{
			{
				Matched: model.MatchedCmd{
					Name: name,
					Cmd:  cmd,
				},
				Env: env,
			},
		},
	}
}

func assertOutputContains(t *testing.T, output string, expected []string) {
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected output to contain %q, got:\n%s", exp, output)
		}
	}
}

func assertOutputNotContains(t *testing.T, output string, notExpected []string) {
	for _, notExp := range notExpected {
		if strings.Contains(output, notExp) {
			t.Errorf("Output should NOT contain %q, but it does", notExp)
		}
	}
}

func TestDumpFlowEmpty(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{Cmds: []model.ParsedCmd{}}
	args := NewDumpFlowArgs()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	if len(output) > 0 {
		t.Errorf("Empty flow should produce no output, got: %s", output)
	}
}

func TestDumpFlowNilCmd(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "nil-cmd", Cmd: nil}}}},
		},
	}
	args := NewDumpFlowArgs()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	if strings.Contains(output, "[nil-cmd]") {
		t.Errorf("Nil cmd should not produce output, got: %s", output)
	}
}

func TestDumpFlowQuietCmd(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	quietCmd := tree.AddSub("quiet-cmd")
	quietCmd.RegEmptyCmd("a quiet command").SetQuiet()

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("quiet-cmd", quietCmd)},
	}
	args := NewDumpFlowArgs()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	if strings.Contains(output, "quiet-cmd") {
		t.Errorf("Quiet cmd should not be displayed by default, got: %s", output)
	}
}

func TestDumpFlowQuietCmdWithDisplayQuiet(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	quietCmd := tree.AddSub("quiet-cmd")
	quietCmd.RegEmptyCmd("a quiet command").SetQuiet()

	cc := createTestCli(screen, tree)
	env := createTestEnv()
	env.SetBool("display.mod.quiet", true)

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("quiet-cmd", quietCmd)},
	}
	args := NewDumpFlowArgs()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"quiet-cmd"})
}

func TestDumpFlowSimpleCommand(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command 1")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1"})
}

func TestDumpFlowSimpleSequence(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command 1")
	cmd2 := tree.AddSub("cmd2")
	cmd2.RegEmptyCmd("test command 2")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			newParsedCmd("cmd1", cmd1),
			newParsedCmd("cmd2", cmd2),
		},
	}
	args := NewDumpFlowArgs()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1", "cmd2"})
}

func TestDumpFlowThreeCommandSequence(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("first command")
	cmd2 := tree.AddSub("cmd2")
	cmd2.RegEmptyCmd("second command")
	cmd3 := tree.AddSub("cmd3")
	cmd3.RegEmptyCmd("third command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			newParsedCmd("cmd1", cmd1),
			newParsedCmd("cmd2", cmd2),
			newParsedCmd("cmd3", cmd3),
		},
	}
	args := NewDumpFlowArgs()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1", "cmd2", "cmd3"})

	idx1 := strings.Index(output, "cmd1")
	idx2 := strings.Index(output, "cmd2")
	idx3 := strings.Index(output, "cmd3")

	if !(idx1 < idx2 && idx2 < idx3) {
		t.Errorf("Commands should appear in order, got indices: %d, %d, %d", idx1, idx2, idx3)
	}
}

func TestDumpFlowWithGlobalEnv(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		GlobalEnv: model.ParsedEnv{
			"global.key": model.NewParsedEnvVal("global.key", "global-value"),
		},
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1"})
}

func TestDumpFlowWithCmdArgs(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	echoCmd := tree.AddSub("echo")
	echoCmd.RegEmptyCmd("print message").
		AddArg("message", "", "msg", "m").
		AddArg("color", "", "c")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			{
				Segments: []model.ParsedCmdSeg{
					{
						Matched: model.MatchedCmd{Name: "echo", Cmd: echoCmd},
						Env: model.ParsedEnv{
							"echo.message": model.NewParsedEnvArgv("echo.message", "hello-world"),
						},
					},
				},
			},
		},
	}
	args := NewDumpFlowArgs()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"echo", "hello-world"})
}

func TestDumpFlowSkeletonMode(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command with help")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs().SetSkeleton()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1"})
}

func TestDumpFlowSimpleMode(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs().SetSimple()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1"})
}

func TestDumpFlowExecutedSucceeded(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs()
	args.ShowExecutedEnvFull = true

	executedFlow := model.NewExecutedFlow("test-session")
	executedCmd := model.NewExecutedCmd("cmd1")
	executedCmd.Result = model.ExecutedResultSucceeded
	executedCmd.StartTs = time.Now().Add(-time.Second)
	executedCmd.FinishTs = time.Now()
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1", "OK"})
}

func TestDumpFlowExecutedError(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs()
	args.ShowExecutedEnvFull = true

	executedFlow := model.NewExecutedFlow("test-session")
	executedCmd := model.NewExecutedCmd("cmd1")
	executedCmd.Result = model.ExecutedResultError
	executedCmd.StartTs = time.Now().Add(-time.Second)
	executedCmd.FinishTs = time.Now()
	executedCmd.ErrStrs = []string{"something went wrong"}
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1", "ERR", "something went wrong"})
}

func TestDumpFlowExecutedSkipped(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs()

	executedFlow := model.NewExecutedFlow("test-session")
	executedCmd := model.NewExecutedCmd("cmd1")
	executedCmd.Result = model.ExecutedResultSkipped
	executedCmd.StartTs = time.Now()
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1", "skipped"})
}

func TestDumpFlowExecutedIncompleted(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs()

	executedFlow := model.NewExecutedFlow("test-session")
	executedCmd := model.NewExecutedCmd("cmd1")
	executedCmd.Result = model.ExecutedResultIncompleted
	executedCmd.StartTs = time.Now()
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1", "failed"})
}

func TestDumpFlowExecutedRunning(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs()

	executedFlow := model.NewExecutedFlow("test-session")
	executedCmd := model.NewExecutedCmd("cmd1")
	executedCmd.Result = model.ExecutedResultIncompleted
	executedCmd.StartTs = time.Now()
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, true, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1", "running"})
}

func TestDumpFlowExecutedUnRun(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")
	cmd2 := tree.AddSub("cmd2")
	cmd2.RegEmptyCmd("second command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			newParsedCmd("cmd1", cmd1),
			newParsedCmd("cmd2", cmd2),
		},
	}
	args := NewDumpFlowArgs()

	executedFlow := model.NewExecutedFlow("test-session")
	executedCmd1 := model.NewExecutedCmd("cmd1")
	executedCmd1.Result = model.ExecutedResultSucceeded
	executedCmd1.StartTs = time.Now().Add(-time.Second)
	executedCmd1.FinishTs = time.Now()
	executedCmd2 := model.NewExecutedCmd("cmd2")
	executedCmd2.Result = model.ExecutedResultUnRun
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd1, executedCmd2}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1", "OK", "cmd2", "unrun"})
}

func TestDumpFlowMonitorMode(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs().SetMonitorMode()

	executedFlow := model.NewExecutedFlow("monitor-session")
	executedCmd := model.NewExecutedCmd("cmd1")
	executedCmd.Result = model.ExecutedResultSucceeded
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"monitor-session"})
}

func TestDumpFlowDepthFolding(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs().SetMaxDepth(0)

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1"})
}

func TestDumpFlowMaxTrivial(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.SetTrivial(2)
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs().SetMaxTrivial(0)

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1"})
}

func TestDumpCmdHelp(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	helpCmd := tree.AddSub("help-cmd")
	helpCmd.RegEmptyCmd("this is a help message")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("help-cmd", helpCmd)},
	}
	args := NewDumpFlowArgs()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"help-cmd", "this is a help message"})
}

func TestDumpFlowNestedExecutedFlow(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	parentCmd := tree.AddSub("parent")
	parentCmd.RegEmptyCmd("parent command")
	childCmd := tree.AddSub("child")
	childCmd.RegEmptyCmd("child command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("parent", parentCmd)},
	}
	args := NewDumpFlowArgs()
	args.ShowExecutedEnvFull = true

	executedFlow := model.NewExecutedFlow("nested-session")
	executedParent := model.NewExecutedCmd("parent")
	executedParent.Result = model.ExecutedResultSucceeded
	executedParent.StartTs = time.Now().Add(-2 * time.Second)
	executedParent.FinishTs = time.Now()

	subFlow := model.NewExecutedFlow("subflow")
	executedChild := model.NewExecutedCmd("child")
	executedChild.Result = model.ExecutedResultSucceeded
	executedChild.StartTs = time.Now().Add(-time.Second)
	executedChild.FinishTs = time.Now()
	subFlow.Cmds = []*model.ExecutedCmd{executedChild}
	executedParent.SubFlow = subFlow

	executedFlow.Cmds = []*model.ExecutedCmd{executedParent}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"parent", "OK"})
}

func TestDumpFlowEnvModified(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs()
	args.ShowExecutedModifiedEnv = true

	executedFlow := model.NewExecutedFlow("env-session")
	executedCmd := model.NewExecutedCmd("cmd1")
	executedCmd.Result = model.ExecutedResultSucceeded
	executedCmd.StartTs = time.Now().Add(-time.Second)
	executedCmd.FinishTs = time.Now()

	startEnv := createTestEnv()
	startEnv.Set("test.key", "original-value")
	executedCmd.StartEnv = startEnv

	finishEnv := createTestEnv()
	finishEnv.Set("test.key", "modified-value")
	executedCmd.FinishEnv = finishEnv

	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1", "OK"})
}

func TestDumpFlowEnvAdded(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs()
	args.ShowExecutedModifiedEnv = true

	executedFlow := model.NewExecutedFlow("env-session")
	executedCmd := model.NewExecutedCmd("cmd1")
	executedCmd.Result = model.ExecutedResultSucceeded
	executedCmd.StartTs = time.Now().Add(-time.Second)
	executedCmd.FinishTs = time.Now()

	startEnv := createTestEnv()
	executedCmd.StartEnv = startEnv

	finishEnv := createTestEnv()
	finishEnv.Set("new.key", "added-value")
	executedCmd.FinishEnv = finishEnv

	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1", "OK"})
}

func TestDumpFlowEnvDeleted(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs()
	args.ShowExecutedModifiedEnv = true

	executedFlow := model.NewExecutedFlow("env-session")
	executedCmd := model.NewExecutedCmd("cmd1")
	executedCmd.Result = model.ExecutedResultSucceeded
	executedCmd.StartTs = time.Now().Add(-time.Second)
	executedCmd.FinishTs = time.Now()

	startEnv := createTestEnv()
	startEnv.Set("deleted.key", "will-be-deleted")
	executedCmd.StartEnv = startEnv

	finishEnv := createTestEnv()
	executedCmd.FinishEnv = finishEnv

	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1", "OK"})
}

func TestDumpFlowFlowNotMatched(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs()

	executedFlow := model.NewExecutedFlow("mismatch-session")
	executedCmd := model.NewExecutedCmd("different-cmd")
	executedCmd.Result = model.ExecutedResultSucceeded
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"flow not matched", "different-cmd"})
}

func TestDumpFlowExecutedWithErrorMsg(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs()

	executedFlow := model.NewExecutedFlow("error-session")
	executedCmd := model.NewExecutedCmd("cmd1")
	executedCmd.Result = model.ExecutedResultError
	executedCmd.StartTs = time.Now()
	executedCmd.ErrStrs = []string{"Error: connection failed", "Error: timeout"}
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1", "ERR", "connection failed", "timeout"})
}

func TestDumpFlowSensitiveValue(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		GlobalEnv: model.ParsedEnv{
			"db.password": model.NewParsedEnvVal("db.password", "secret123"),
		},
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1"})
}

func TestDumpFlowIndentation(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("first command")
	cmd2 := tree.AddSub("cmd2")
	cmd2.RegEmptyCmd("second command")
	cmd3 := tree.AddSub("cmd3")
	cmd3.RegEmptyCmd("third command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			newParsedCmd("cmd1", cmd1),
			newParsedCmd("cmd2", cmd2),
			newParsedCmd("cmd3", cmd3),
		},
	}
	args := NewDumpFlowArgs()
	args.IndentSize = 4

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	lines := strings.Split(output, "\n")
	var nonEmptyLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	if len(nonEmptyLines) < 3 {
		t.Errorf("Expected at least 3 non-empty lines, got %d", len(nonEmptyLines))
	}
}

func TestDumpFlowFromCmdIdx(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("first command")
	cmd2 := tree.AddSub("cmd2")
	cmd2.RegEmptyCmd("second command")
	cmd3 := tree.AddSub("cmd3")
	cmd3.RegEmptyCmd("third command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			newParsedCmd("cmd1", cmd1),
			newParsedCmd("cmd2", cmd2),
			newParsedCmd("cmd3", cmd3),
		},
	}
	args := NewDumpFlowArgs()

	DumpFlow(cc, env, flow, 1, args, nil)

	output := screen.GetOutput()
	assertOutputNotContains(t, output, []string{"cmd1"})
	assertOutputContains(t, output, []string{"cmd2", "cmd3"})
}

func TestDumpFlowTenCommandSequence(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	var cmds []*model.CmdTree
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("cmd%d", i)
		cmd := tree.AddSub(name)
		cmd.RegEmptyCmd(fmt.Sprintf("command %d", i))
		cmds = append(cmds, cmd)
	}

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	var parsedCmds []model.ParsedCmd
	for i, cmd := range cmds {
		parsedCmds = append(parsedCmds, newParsedCmd(fmt.Sprintf("cmd%d", i+1), cmd))
	}

	flow := &model.ParsedCmds{Cmds: parsedCmds}
	args := NewDumpFlowArgs()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	for i := 1; i <= 10; i++ {
		expected := fmt.Sprintf("cmd%d", i)
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q", expected)
		}
	}
}

func TestDumpFlowMultipleGlobalEnv(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		GlobalEnv: model.ParsedEnv{
			"global.key1": model.NewParsedEnvVal("global.key1", "value1"),
			"global.key2": model.NewParsedEnvVal("global.key2", "value2"),
			"global.key3": model.NewParsedEnvVal("global.key3", "value3"),
		},
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1"})
}

func TestDumpFlowDeepNestedStructure(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	var cmds []*model.CmdTree
	for i := 1; i <= 5; i++ {
		name := fmt.Sprintf("level%d", i)
		cmd := tree.AddSub(name)
		cmd.RegEmptyCmd(fmt.Sprintf("level %d command", i))
		cmds = append(cmds, cmd)
	}

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	var parsedCmds []model.ParsedCmd
	for i, cmd := range cmds {
		parsedCmds = append(parsedCmds, newParsedCmd(fmt.Sprintf("level%d", i+1), cmd))
	}

	flow := &model.ParsedCmds{Cmds: parsedCmds}
	args := NewDumpFlowArgs()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	for i := 1; i <= 5; i++ {
		expected := fmt.Sprintf("level%d", i)
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q", expected)
		}
	}
}

func TestDumpFlowTrivialMark(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("trivial-cmd")
	cmd1.SetTrivial(1)
	cmd1.RegEmptyCmd("a trivial command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			{TrivialLvl: 1, Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "trivial-cmd", Cmd: cmd1}}}},
		},
	}
	args := NewDumpFlowArgs().SetMaxTrivial(1)

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"trivial-cmd"})
}

func TestDumpFlowWithEmptySegment(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	cc := createTestCli(screen, tree)
	env := createTestEnv()

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			{Segments: []model.ParsedCmdSeg{{Matched: model.MatchedCmd{Name: "", Cmd: nil}}}},
			newParsedCmd("cmd1", cmd1),
		},
	}
	args := NewDumpFlowArgs()

	DumpFlow(cc, env, flow, 0, args, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1"})
}
