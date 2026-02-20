package display

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/innerr/ticat/pkg/core/model"
)

func createExecutedFlowWithCmds(cmdNames []string, results []model.ExecutedResult) *model.ExecutedFlow {
	flow := model.NewExecutedFlow("test-session-" + time.Now().Format("20060102-150405"))
	flow.Flow = strings.Join(cmdNames, " : ")
	for i, cmd := range cmdNames {
		executedCmd := model.NewExecutedCmd(cmd)
		if i < len(results) {
			executedCmd.Result = results[i]
		} else {
			executedCmd.Result = model.ExecutedResultSucceeded
		}
		executedCmd.StartTs = time.Now().Add(-time.Duration(i+1) * time.Second)
		executedCmd.FinishTs = time.Now().Add(-time.Duration(i) * time.Second)
		flow.Cmds = append(flow.Cmds, executedCmd)
	}
	return flow
}

func createMultiLayerExecutedFlow(layers int, failAtLayer int) *model.ExecutedFlow {
	sessionName := fmt.Sprintf("multilayer-L%d", layers)
	flow := model.NewExecutedFlow(sessionName)
	flow.Flow = "multilayer-flow"

	var cmds []*model.ExecutedCmd
	for i := 1; i <= layers; i++ {
		cmd := model.NewExecutedCmd(fmt.Sprintf("layer%d", i))
		cmd.StartTs = time.Now().Add(-time.Duration(layers-i+1) * time.Second)
		cmd.FinishTs = time.Now().Add(-time.Duration(layers-i) * time.Second)

		if failAtLayer > 0 && i == failAtLayer {
			cmd.Result = model.ExecutedResultError
			cmd.ErrStrs = []string{fmt.Sprintf("Error at layer %d", i)}
		} else if failAtLayer > 0 && i > failAtLayer {
			cmd.Result = model.ExecutedResultUnRun
		} else {
			cmd.Result = model.ExecutedResultSucceeded
		}
		cmds = append(cmds, cmd)
	}
	flow.Cmds = cmds

	if failAtLayer > 0 {
		flow.Result = model.ExecutedResultError
	} else {
		flow.Result = model.ExecutedResultSucceeded
	}

	return flow
}

func TestExecutedFlowSingleSucceeded(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("test command")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("cmd1", cmd1)},
	}
	args := NewDumpFlowArgs()
	args.ShowExecutedEnvFull = true

	executedFlow := model.NewExecutedFlow("session-001")
	executedCmd := model.NewExecutedCmd("cmd1")
	executedCmd.Result = model.ExecutedResultSucceeded
	executedCmd.StartTs = time.Now().Add(-time.Second)
	executedCmd.FinishTs = time.Now()
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1", "OK"})
}

func TestExecutedFlowSingleError(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	cmd1 := tree.AddSub("fail-cmd")
	cmd1.RegEmptyCmd("fail command")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("fail-cmd", cmd1)},
	}
	args := NewDumpFlowArgs()
	args.ShowExecutedEnvFull = true

	executedFlow := model.NewExecutedFlow("session-002")
	executedCmd := model.NewExecutedCmd("fail-cmd")
	executedCmd.Result = model.ExecutedResultError
	executedCmd.StartTs = time.Now().Add(-time.Second)
	executedCmd.FinishTs = time.Now()
	executedCmd.ErrStrs = []string{"Error: command failed", "Error: exit code 1"}
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"fail-cmd", "ERR", "command failed", "exit code 1"})
}

func TestExecutedFlowSingleSkipped(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	cmd1 := tree.AddSub("skip-cmd")
	cmd1.RegEmptyCmd("skip command")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("skip-cmd", cmd1)},
	}
	args := NewDumpFlowArgs()

	executedFlow := model.NewExecutedFlow("session-003")
	executedCmd := model.NewExecutedCmd("skip-cmd")
	executedCmd.Result = model.ExecutedResultSkipped
	executedCmd.StartTs = time.Now()
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"skip-cmd", "skipped"})
}

func TestExecutedFlowSingleIncompleted(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	cmd1 := tree.AddSub("incomplete-cmd")
	cmd1.RegEmptyCmd("incomplete command")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("incomplete-cmd", cmd1)},
	}
	args := NewDumpFlowArgs()

	executedFlow := model.NewExecutedFlow("session-004")
	executedCmd := model.NewExecutedCmd("incomplete-cmd")
	executedCmd.Result = model.ExecutedResultIncompleted
	executedCmd.StartTs = time.Now()
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"incomplete-cmd", "failed"})
}

func TestExecutedFlowWithUnRun(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("first command")
	cmd2 := tree.AddSub("cmd2")
	cmd2.RegEmptyCmd("second command")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			newParsedCmd("cmd1", cmd1),
			newParsedCmd("cmd2", cmd2),
		},
	}
	args := NewDumpFlowArgs()

	executedFlow := model.NewExecutedFlow("session-005")
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

func TestExecutedFlowRunning(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	cmd1 := tree.AddSub("running-cmd")
	cmd1.RegEmptyCmd("running command")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("running-cmd", cmd1)},
	}
	args := NewDumpFlowArgs()

	executedFlow := model.NewExecutedFlow("session-running")
	executedCmd := model.NewExecutedCmd("running-cmd")
	executedCmd.Result = model.ExecutedResultIncompleted
	executedCmd.StartTs = time.Now().Add(-5 * time.Second)
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, true, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"running-cmd", "running"})
}

func TestExecutedFlowTenCommands(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	var cmds []model.ParsedCmd
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("step%d", i)
		cmd := tree.AddSub(name)
		cmd.RegEmptyCmd(fmt.Sprintf("step %d command", i))
		cmds = append(cmds, newParsedCmd(name, cmd))
	}

	flow := &model.ParsedCmds{Cmds: cmds}
	args := NewDumpFlowArgs()

	executedFlow := model.NewExecutedFlow("session-ten")
	for i := 1; i <= 10; i++ {
		executedCmd := model.NewExecutedCmd(fmt.Sprintf("step%d", i))
		executedCmd.Result = model.ExecutedResultSucceeded
		executedCmd.StartTs = time.Now().Add(-time.Duration(11-i) * time.Second)
		executedCmd.FinishTs = time.Now().Add(-time.Duration(10-i) * time.Second)
		executedFlow.Cmds = append(executedFlow.Cmds, executedCmd)
	}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	for i := 1; i <= 10; i++ {
		assertOutputContains(t, output, []string{fmt.Sprintf("step%d", i)})
	}
}

func TestExecutedFlowTwentyCommands(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	var cmds []model.ParsedCmd
	for i := 1; i <= 20; i++ {
		name := fmt.Sprintf("cmd%02d", i)
		cmd := tree.AddSub(name)
		cmd.RegEmptyCmd(fmt.Sprintf("command %d", i))
		cmds = append(cmds, newParsedCmd(name, cmd))
	}

	flow := &model.ParsedCmds{Cmds: cmds}
	args := NewDumpFlowArgs()

	executedFlow := model.NewExecutedFlow("session-twenty")
	for i := 1; i <= 20; i++ {
		executedCmd := model.NewExecutedCmd(fmt.Sprintf("cmd%02d", i))
		executedCmd.Result = model.ExecutedResultSucceeded
		executedCmd.StartTs = time.Now().Add(-time.Duration(21-i) * time.Second)
		executedCmd.FinishTs = time.Now().Add(-time.Duration(20-i) * time.Second)
		executedFlow.Cmds = append(executedFlow.Cmds, executedCmd)
	}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	for i := 1; i <= 20; i++ {
		assertOutputContains(t, output, []string{fmt.Sprintf("cmd%02d", i)})
	}
}

func TestExecutedFlowFiftyCommands(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	var cmds []model.ParsedCmd
	for i := 1; i <= 50; i++ {
		name := fmt.Sprintf("task%02d", i)
		cmd := tree.AddSub(name)
		cmd.RegEmptyCmd(fmt.Sprintf("task %d", i))
		cmds = append(cmds, newParsedCmd(name, cmd))
	}

	flow := &model.ParsedCmds{Cmds: cmds}
	args := NewDumpFlowArgs()

	executedFlow := model.NewExecutedFlow("session-fifty")
	for i := 1; i <= 50; i++ {
		executedCmd := model.NewExecutedCmd(fmt.Sprintf("task%02d", i))
		executedCmd.Result = model.ExecutedResultSucceeded
		executedCmd.StartTs = time.Now().Add(-time.Duration(51-i) * time.Second)
		executedCmd.FinishTs = time.Now().Add(-time.Duration(50-i) * time.Second)
		executedFlow.Cmds = append(executedFlow.Cmds, executedCmd)
	}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	for i := 1; i <= 50; i++ {
		assertOutputContains(t, output, []string{fmt.Sprintf("task%02d", i)})
	}
}

func TestExecutedFlowNestedTwoLevels(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	parentCmd := tree.AddSub("parent")
	parentCmd.RegEmptyCmd("parent command")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("parent", parentCmd)},
	}
	args := NewDumpFlowArgs()
	args.MaxDepth = 32
	args.ShowExecutedEnvFull = true

	executedFlow := model.NewExecutedFlow("session-nested-2")
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

func TestExecutedFlowNestedThreeLevels(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	parentCmd := tree.AddSub("root")
	parentCmd.RegEmptyCmd("root command")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("root", parentCmd)},
	}
	args := NewDumpFlowArgs()
	args.MaxDepth = 32
	args.ShowExecutedEnvFull = true

	executedFlow := model.NewExecutedFlow("session-nested-3")
	executedRoot := model.NewExecutedCmd("root")
	executedRoot.Result = model.ExecutedResultSucceeded
	executedRoot.StartTs = time.Now().Add(-3 * time.Second)
	executedRoot.FinishTs = time.Now()

	subFlow1 := model.NewExecutedFlow("subflow-1")
	executedChild1 := model.NewExecutedCmd("level1")
	executedChild1.Result = model.ExecutedResultSucceeded
	executedChild1.StartTs = time.Now().Add(-2 * time.Second)
	executedChild1.FinishTs = time.Now()

	subFlow2 := model.NewExecutedFlow("subflow-2")
	executedChild2 := model.NewExecutedCmd("level2")
	executedChild2.Result = model.ExecutedResultSucceeded
	executedChild2.StartTs = time.Now().Add(-time.Second)
	executedChild2.FinishTs = time.Now()
	subFlow2.Cmds = []*model.ExecutedCmd{executedChild2}

	executedChild1.SubFlow = subFlow2
	subFlow1.Cmds = []*model.ExecutedCmd{executedChild1}
	executedRoot.SubFlow = subFlow1

	executedFlow.Cmds = []*model.ExecutedCmd{executedRoot}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"root", "OK"})
}

func TestExecutedFlowNestedWithSubFlowError(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	parentCmd := tree.AddSub("parent")
	parentCmd.RegEmptyCmd("parent command")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("parent", parentCmd)},
	}
	args := NewDumpFlowArgs()
	args.MaxDepth = 32
	args.ShowExecutedEnvFull = true

	executedFlow := model.NewExecutedFlow("session-nested-error")
	executedParent := model.NewExecutedCmd("parent")
	executedParent.Result = model.ExecutedResultSucceeded
	executedParent.StartTs = time.Now().Add(-2 * time.Second)
	executedParent.FinishTs = time.Now()

	subFlow := model.NewExecutedFlow("subflow")
	executedChild := model.NewExecutedCmd("child")
	executedChild.Result = model.ExecutedResultError
	executedChild.ErrStrs = []string{"Error: child command failed"}
	executedChild.StartTs = time.Now().Add(-time.Second)
	executedChild.FinishTs = time.Now()
	subFlow.Cmds = []*model.ExecutedCmd{executedChild}
	executedParent.SubFlow = subFlow

	executedFlow.Cmds = []*model.ExecutedCmd{executedParent}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"parent"})
}

func TestExecutedFlowMultiLayerWithFailure(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	var cmds []model.ParsedCmd
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("layer%d", i)
		cmd := tree.AddSub(name)
		cmd.RegEmptyCmd(fmt.Sprintf("layer %d command", i))
		cmds = append(cmds, newParsedCmd(name, cmd))
	}

	flow := &model.ParsedCmds{Cmds: cmds}
	args := NewDumpFlowArgs()

	executedFlow := createMultiLayerExecutedFlow(10, 5)

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"layer5", "ERR", "Error at layer 5"})
	assertOutputContains(t, output, []string{"layer6", "unrun"})
	assertOutputContains(t, output, []string{"layer4", "OK"})
}

func TestExecutedFlowMultiLayerAllSucceeded(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	var cmds []model.ParsedCmd
	for i := 1; i <= 25; i++ {
		name := fmt.Sprintf("layer%d", i)
		cmd := tree.AddSub(name)
		cmd.RegEmptyCmd(fmt.Sprintf("layer %d command", i))
		cmds = append(cmds, newParsedCmd(name, cmd))
	}

	flow := &model.ParsedCmds{Cmds: cmds}
	args := NewDumpFlowArgs()

	executedFlow := createMultiLayerExecutedFlow(25, 0)

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	for i := 1; i <= 25; i++ {
		assertOutputContains(t, output, []string{fmt.Sprintf("layer%d", i)})
	}
	assertOutputNotContains(t, output, []string{"ERR", "unrun"})
}

func TestExecutedFlowMultiLayerFiftyWithFailure(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	var cmds []model.ParsedCmd
	for i := 1; i <= 50; i++ {
		name := fmt.Sprintf("layer%d", i)
		cmd := tree.AddSub(name)
		cmd.RegEmptyCmd(fmt.Sprintf("layer %d command", i))
		cmds = append(cmds, newParsedCmd(name, cmd))
	}

	flow := &model.ParsedCmds{Cmds: cmds}
	args := NewDumpFlowArgs()

	executedFlow := createMultiLayerExecutedFlow(50, 25)

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"layer25", "ERR"})
	assertOutputContains(t, output, []string{"layer26", "unrun"})
	assertOutputContains(t, output, []string{"layer24", "OK"})
	assertOutputContains(t, output, []string{"layer1", "OK"})
}

func TestExecutedFlowSkeletonMode(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("command 1")
	cmd2 := tree.AddSub("cmd2")
	cmd2.RegEmptyCmd("command 2")
	cmd3 := tree.AddSub("cmd3")
	cmd3.RegEmptyCmd("command 3")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			newParsedCmd("cmd1", cmd1),
			newParsedCmd("cmd2", cmd2),
			newParsedCmd("cmd3", cmd3),
		},
	}
	args := NewDumpFlowArgs().SetSkeleton()

	executedFlow := model.NewExecutedFlow("session-skeleton")
	executedFlow.Cmds = []*model.ExecutedCmd{
		{Cmd: "cmd1", Result: model.ExecutedResultSucceeded},
		{Cmd: "cmd2", Result: model.ExecutedResultSucceeded},
		{Cmd: "cmd3", Result: model.ExecutedResultSucceeded},
	}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1", "cmd2", "cmd3"})
}

func TestExecutedFlowSimpleMode(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	cmd1 := tree.AddSub("cmd1")
	cmd1.RegEmptyCmd("command 1")
	cmd2 := tree.AddSub("cmd2")
	cmd2.RegEmptyCmd("command 2")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			newParsedCmd("cmd1", cmd1),
			newParsedCmd("cmd2", cmd2),
		},
	}
	args := NewDumpFlowArgs().SetSimple()

	executedFlow := model.NewExecutedFlow("session-simple")
	executedFlow.Cmds = []*model.ExecutedCmd{
		{Cmd: "cmd1", Result: model.ExecutedResultSucceeded},
		{Cmd: "cmd2", Result: model.ExecutedResultSucceeded},
	}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1", "cmd2"})
}

func TestExecutedFlowMonitorMode(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	cmd1 := tree.AddSub("monitor-cmd")
	cmd1.RegEmptyCmd("monitor command")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("monitor-cmd", cmd1)},
	}
	args := NewDumpFlowArgs().SetMonitorMode()

	executedFlow := model.NewExecutedFlow("monitor-session-001")
	executedCmd := model.NewExecutedCmd("monitor-cmd")
	executedCmd.Result = model.ExecutedResultSucceeded
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"monitor-session-001"})
}

func TestExecutedFlowMonitorModeRunning(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	cmd1 := tree.AddSub("running-cmd")
	cmd1.RegEmptyCmd("running command")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("running-cmd", cmd1)},
	}
	args := NewDumpFlowArgs().SetMonitorMode()

	executedFlow := model.NewExecutedFlow("monitor-running-001")
	executedCmd := model.NewExecutedCmd("running-cmd")
	executedCmd.Result = model.ExecutedResultIncompleted
	executedCmd.StartTs = time.Now().Add(-10 * time.Second)
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, true, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"monitor-running-001"})
}

func TestExecutedFlowMultipleErrors(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	cmd1 := tree.AddSub("fail-cmd")
	cmd1.RegEmptyCmd("fail command")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("fail-cmd", cmd1)},
	}
	args := NewDumpFlowArgs()
	args.ShowExecutedEnvFull = true

	executedFlow := model.NewExecutedFlow("session-multi-error")
	executedCmd := model.NewExecutedCmd("fail-cmd")
	executedCmd.Result = model.ExecutedResultError
	executedCmd.ErrStrs = []string{
		"Error: connection timeout",
		"Error: retry 1 failed",
		"Error: retry 2 failed",
		"Error: max retries exceeded",
	}
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{
		"fail-cmd", "ERR",
		"connection timeout",
		"retry 1 failed",
		"retry 2 failed",
		"max retries exceeded",
	})
}

func TestExecutedFlowMixedResults(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	var cmds []model.ParsedCmd
	for i := 1; i <= 6; i++ {
		name := fmt.Sprintf("cmd%d", i)
		cmd := tree.AddSub(name)
		cmd.RegEmptyCmd(fmt.Sprintf("command %d", i))
		cmds = append(cmds, newParsedCmd(name, cmd))
	}

	flow := &model.ParsedCmds{Cmds: cmds}
	args := NewDumpFlowArgs()
	args.ShowExecutedEnvFull = true

	executedFlow := model.NewExecutedFlow("session-mixed")
	results := []model.ExecutedResult{
		model.ExecutedResultSucceeded,
		model.ExecutedResultSkipped,
		model.ExecutedResultSucceeded,
		model.ExecutedResultError,
		model.ExecutedResultUnRun,
		model.ExecutedResultUnRun,
	}

	for i, result := range results {
		cmd := model.NewExecutedCmd(fmt.Sprintf("cmd%d", i+1))
		cmd.Result = result
		if result == model.ExecutedResultError {
			cmd.ErrStrs = []string{"Error: cmd4 failed"}
		}
		executedFlow.Cmds = append(executedFlow.Cmds, cmd)
	}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd1", "OK"})
	assertOutputContains(t, output, []string{"cmd2", "skipped"})
	assertOutputContains(t, output, []string{"cmd3", "OK"})
	assertOutputContains(t, output, []string{"cmd4", "ERR"})
	assertOutputContains(t, output, []string{"cmd5", "unrun"})
}

func TestExecutedFlowWithOnlyFailed(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	var cmds []model.ParsedCmd
	for i := 1; i <= 4; i++ {
		name := fmt.Sprintf("cmd%d", i)
		cmd := tree.AddSub(name)
		cmd.RegEmptyCmd(fmt.Sprintf("command %d", i))
		cmds = append(cmds, newParsedCmd(name, cmd))
	}

	flow := &model.ParsedCmds{Cmds: cmds}
	args := NewDumpFlowArgs()
	args.OnlyFailed = true
	args.Skeleton = true

	executedFlow := model.NewExecutedFlow("session-only-failed")
	executedFlow.Cmds = []*model.ExecutedCmd{
		{Cmd: "cmd1", Result: model.ExecutedResultSucceeded},
		{Cmd: "cmd2", Result: model.ExecutedResultSucceeded},
		{Cmd: "cmd3", Result: model.ExecutedResultError, ErrStrs: []string{"Error: cmd3 failed"}},
		{Cmd: "cmd4", Result: model.ExecutedResultUnRun},
	}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd3", "ERR"})
	assertOutputNotContains(t, output, []string{"cmd1", "cmd2"})
}

func TestExecutedFlowWithFilter(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	var cmds []model.ParsedCmd
	for _, name := range []string{"setup", "target-cmd", "cleanup", "target-verify"} {
		cmd := tree.AddSub(name)
		cmd.RegEmptyCmd(name + " command")
		cmds = append(cmds, newParsedCmd(name, cmd))
	}

	flow := &model.ParsedCmds{Cmds: cmds}
	args := NewDumpFlowArgs()
	args.FilterNames = []string{"target"}
	args.Skeleton = true

	executedFlow := model.NewExecutedFlow("session-filter")
	executedFlow.Cmds = []*model.ExecutedCmd{
		{Cmd: "setup", Result: model.ExecutedResultSucceeded},
		{Cmd: "target-cmd", Result: model.ExecutedResultSucceeded},
		{Cmd: "cleanup", Result: model.ExecutedResultSucceeded},
		{Cmd: "target-verify", Result: model.ExecutedResultSucceeded},
	}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"target-cmd", "target-verify"})
	assertOutputNotContains(t, output, []string{"setup", "cleanup"})
}

func TestExecutedFlowWithFilterMultiple(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	for _, name := range []string{"init", "db-connect", "api-call", "cleanup", "db-close"} {
		cmd := tree.AddSub(name)
		cmd.RegEmptyCmd(name + " command")
	}

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			newParsedCmd("init", tree.GetSub("init")),
			newParsedCmd("db-connect", tree.GetSub("db-connect")),
			newParsedCmd("api-call", tree.GetSub("api-call")),
			newParsedCmd("cleanup", tree.GetSub("cleanup")),
			newParsedCmd("db-close", tree.GetSub("db-close")),
		},
	}
	args := NewDumpFlowArgs()
	args.FilterNames = []string{"db", "api"}
	args.Skeleton = true

	executedFlow := model.NewExecutedFlow("session-filter-multi")
	executedFlow.Cmds = []*model.ExecutedCmd{
		{Cmd: "init", Result: model.ExecutedResultSucceeded},
		{Cmd: "db-connect", Result: model.ExecutedResultSucceeded},
		{Cmd: "api-call", Result: model.ExecutedResultSucceeded},
		{Cmd: "cleanup", Result: model.ExecutedResultSucceeded},
		{Cmd: "db-close", Result: model.ExecutedResultSucceeded},
	}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"db-connect", "api-call", "db-close"})
	assertOutputNotContains(t, output, []string{"init", "cleanup"})
}

func TestExecutedFlowLargeFlowWithError(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	var cmds []model.ParsedCmd
	for i := 1; i <= 100; i++ {
		name := fmt.Sprintf("step%d", i)
		cmd := tree.AddSub(name)
		cmd.RegEmptyCmd(fmt.Sprintf("step %d command", i))
		cmds = append(cmds, newParsedCmd(name, cmd))
	}

	flow := &model.ParsedCmds{Cmds: cmds}
	args := NewDumpFlowArgs()

	executedFlow := model.NewExecutedFlow("session-large-error")
	for i := 1; i <= 100; i++ {
		executedCmd := model.NewExecutedCmd(fmt.Sprintf("step%d", i))

		if i == 50 {
			executedCmd.Result = model.ExecutedResultError
			executedCmd.ErrStrs = []string{"Error: critical failure at step 50"}
		} else if i > 50 {
			executedCmd.Result = model.ExecutedResultUnRun
		} else {
			executedCmd.Result = model.ExecutedResultSucceeded
		}
		executedFlow.Cmds = append(executedFlow.Cmds, executedCmd)
	}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"step50", "ERR", "critical failure"})
	assertOutputContains(t, output, []string{"step49", "OK"})
	assertOutputContains(t, output, []string{"step51", "unrun"})
}

func TestExecutedFlowComplexScenario(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	setupCmd := tree.AddSub("setup")
	setupCmd.RegEmptyCmd("setup command")

	deployCmd := tree.AddSub("deploy")
	deployCmd.RegEmptyCmd("deploy command")

	testCmd := tree.AddSub("test")
	testCmd.RegEmptyCmd("test command")

	cleanupCmd := tree.AddSub("cleanup")
	cleanupCmd.RegEmptyCmd("cleanup command")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{
			newParsedCmd("setup", setupCmd),
			newParsedCmd("deploy", deployCmd),
			newParsedCmd("test", testCmd),
			newParsedCmd("cleanup", cleanupCmd),
		},
	}
	args := NewDumpFlowArgs()
	args.MaxDepth = 50
	args.ShowExecutedEnvFull = true

	executedFlow := model.NewExecutedFlow("session-complex")

	setupExecuted := model.NewExecutedCmd("setup")
	setupExecuted.Result = model.ExecutedResultSucceeded

	deployExecuted := model.NewExecutedCmd("deploy")
	deployExecuted.Result = model.ExecutedResultSucceeded

	deploySubFlow := model.NewExecutedFlow("deploy-subflow")
	for i := 1; i <= 5; i++ {
		subCmd := model.NewExecutedCmd(fmt.Sprintf("deploy-step%d", i))
		subCmd.Result = model.ExecutedResultSucceeded
		deploySubFlow.Cmds = append(deploySubFlow.Cmds, subCmd)
	}
	deployExecuted.SubFlow = deploySubFlow

	testExecuted := model.NewExecutedCmd("test")
	testExecuted.Result = model.ExecutedResultError
	testExecuted.ErrStrs = []string{"Error: test failed", "Error: assertion error"}

	cleanupExecuted := model.NewExecutedCmd("cleanup")
	cleanupExecuted.Result = model.ExecutedResultUnRun

	executedFlow.Cmds = []*model.ExecutedCmd{setupExecuted, deployExecuted, testExecuted, cleanupExecuted}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{
		"setup", "OK",
		"deploy",
		"test", "ERR", "test failed", "assertion error",
		"cleanup", "unrun",
	})
}

func TestExecutedFlowFlowNotMatched(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	cmd1 := tree.AddSub("expected-cmd")
	cmd1.RegEmptyCmd("expected command")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("expected-cmd", cmd1)},
	}
	args := NewDumpFlowArgs()

	executedFlow := model.NewExecutedFlow("session-mismatch")
	executedFlow.Flow = "different-flow-cmd"
	executedCmd := model.NewExecutedCmd("different-cmd")
	executedCmd.Result = model.ExecutedResultSucceeded
	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"flow not matched"})
}

func TestExecutedFlowWithEnvModified(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	cmd1 := tree.AddSub("env-cmd")
	cmd1.RegEmptyCmd("env modification command")

	flow := &model.ParsedCmds{
		Cmds: []model.ParsedCmd{newParsedCmd("env-cmd", cmd1)},
	}
	args := NewDumpFlowArgs()
	args.ShowExecutedModifiedEnv = true

	executedFlow := model.NewExecutedFlow("session-env-modified")
	executedCmd := model.NewExecutedCmd("env-cmd")
	executedCmd.Result = model.ExecutedResultSucceeded

	startEnv := model.NewEnvEx(model.EnvLayerSession)
	startEnv.Set("existing.key", "old-value")
	startEnv.Set("unchanged.key", "same-value")
	executedCmd.StartEnv = startEnv

	finishEnv := model.NewEnvEx(model.EnvLayerSession)
	finishEnv.Set("existing.key", "new-value")
	finishEnv.Set("unchanged.key", "same-value")
	finishEnv.Set("added.key", "added-value")
	executedCmd.FinishEnv = finishEnv

	executedFlow.Cmds = []*model.ExecutedCmd{executedCmd}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"env-cmd", "OK"})
}

func TestExecutedFlowMultiLayerFortyAllTypes(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	var cmds []model.ParsedCmd
	for i := 1; i <= 40; i++ {
		name := fmt.Sprintf("cmd%02d", i)
		cmd := tree.AddSub(name)
		cmd.RegEmptyCmd(fmt.Sprintf("command %d", i))
		cmds = append(cmds, newParsedCmd(name, cmd))
	}

	flow := &model.ParsedCmds{Cmds: cmds}
	args := NewDumpFlowArgs()
	args.ShowExecutedEnvFull = true

	executedFlow := model.NewExecutedFlow("session-forty-types")
	executedFlow.Flow = "forty command flow with all result types"

	for i := 1; i <= 40; i++ {
		cmd := model.NewExecutedCmd(fmt.Sprintf("cmd%02d", i))
		cmd.StartTs = time.Now().Add(-time.Duration(41-i) * time.Second)
		cmd.FinishTs = time.Now().Add(-time.Duration(40-i) * time.Second)

		switch {
		case i%10 == 0:
			cmd.Result = model.ExecutedResultSkipped
		case i%10 == 5:
			cmd.Result = model.ExecutedResultError
			cmd.ErrStrs = []string{fmt.Sprintf("Error at cmd%02d", i)}
		default:
			cmd.Result = model.ExecutedResultSucceeded
		}
		executedFlow.Cmds = append(executedFlow.Cmds, cmd)
	}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"cmd05", "ERR", "cmd10", "skipped", "cmd15", "ERR", "cmd01", "OK"})
}

func TestExecutedFlowHundredCommandsAllSucceeded(t *testing.T) {
	screen := &memoryScreen{}
	tree := createTestCmdTree()
	cc := createTestCli(screen, tree)
	env := createTestEnv()

	var cmds []model.ParsedCmd
	for i := 1; i <= 100; i++ {
		name := fmt.Sprintf("task%03d", i)
		cmd := tree.AddSub(name)
		cmd.RegEmptyCmd(fmt.Sprintf("task %d", i))
		cmds = append(cmds, newParsedCmd(name, cmd))
	}

	flow := &model.ParsedCmds{Cmds: cmds}
	args := NewDumpFlowArgs()

	executedFlow := model.NewExecutedFlow("session-hundred")
	for i := 1; i <= 100; i++ {
		executedCmd := model.NewExecutedCmd(fmt.Sprintf("task%03d", i))
		executedCmd.Result = model.ExecutedResultSucceeded
		executedFlow.Cmds = append(executedFlow.Cmds, executedCmd)
	}

	DumpFlowEx(cc, env, flow, 0, args, executedFlow, false, nil)

	output := screen.GetOutput()
	assertOutputContains(t, output, []string{"task001", "task050", "task100"})
}
