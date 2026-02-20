package ticat

import (
	"fmt"
	"strings"
	"testing"
)

type previewCaptureScreen struct {
	output strings.Builder
}

func (s *previewCaptureScreen) Print(text string) error {
	s.output.WriteString(text)
	return nil
}

func (s *previewCaptureScreen) Error(text string) error {
	s.output.WriteString(text)
	return nil
}

func (s *previewCaptureScreen) OutputtedLines() int {
	return strings.Count(s.output.String(), "\n")
}

func (s *previewCaptureScreen) GetOutput() string {
	return s.output.String()
}

func (s *previewCaptureScreen) Reset() {
	s.output.Reset()
}

func buildComplexMultiLayerFlow() []string {
	return []string{
		"env.set", "key=layer1.config", "value=config1",
		":", "echo", "message=layer1-start",
		":", "env.set", "key=layer2.data", "value=data2",
		":", "echo", "message=layer2-middle",
		":", "env.update", "key=layer1.config", "value=config2",
		":", "echo", "message=layer3-end",
		":", "env.set", "key=layer4.final", "value=done",
	}
}

func buildVeryDeepFlow(depth int) []string {
	args := []string{}
	for i := 1; i <= depth; i++ {
		args = append(args, "env.set", fmt.Sprintf("key=deep.level%d", i), fmt.Sprintf("value=val%d", i))
		if i < depth {
			args = append(args, ":")
		}
	}
	return args
}

func buildFlowWithMixedCommands() []string {
	return []string{
		"env.set", "key=mixed.start", "value=init",
		":", "noop",
		":", "echo", "message=step1",
		":", "env.update", "key=mixed.start", "value=updated",
		":", "noop",
		":", "echo", "message=step2",
		":", "env.set", "key=mixed.end", "value=final",
	}
}

func buildTenLevelFlow() []string {
	args := []string{}
	for i := 1; i <= 10; i++ {
		args = append(args, "env.set", fmt.Sprintf("key=level%d.key", i), fmt.Sprintf("value=%d", i))
		if i < 10 {
			args = append(args, ":")
		}
	}
	return args
}

func TestPreviewDashSimpleFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{"echo", "message=hello", ":", "-", ":", "dbg.args.tail", "arg1=test"}
	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertPreviewOutputContains(t, output, []string{"echo", "hello", "test"})
}

func TestPreviewDashComplexFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := buildComplexMultiLayerFlow()
	flow = append(flow, ":", "-", ":", "dbg.args.tail", "arg1=preview-test")

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertPreviewOutputContains(t, output, []string{
		"layer1.config",
		"layer1-start",
		"layer2-middle",
		"layer3-end",
		"layer4.final",
		"preview-test",
	})
}

func TestPreviewDashVeryDeepFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := buildVeryDeepFlow(15)
	flow = append(flow, ":", "-", ":", "dbg.args.tail", "arg1=deep-preview")

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	for i := 1; i <= 15; i++ {
		expected := fmt.Sprintf("deep.level%d", i)
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
	assertOutputContains(t, output, "deep-preview")
}

func TestPreviewDashWithMixedCommands(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := buildFlowWithMixedCommands()
	flow = append(flow, ":", "-", ":", "dbg.args.tail", "arg1=mixed-test")

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertPreviewOutputContains(t, output, []string{
		"mixed.start",
		"step1",
		"step2",
		"mixed.end",
		"mixed-test",
	})
}

func TestPreviewDashDashSimpleFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{"echo", "message=test", ":", "noop", ":", "--", ":", "dbg.args.tail", "arg1=dash-dash"}
	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertPreviewOutputContains(t, output, []string{"echo", "test", "dash-dash"})
}

func TestPreviewDashDashComplexFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := buildComplexMultiLayerFlow()
	flow = append(flow, ":", "--", ":", "dbg.args.tail", "arg1=dd-preview")

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertPreviewOutputContains(t, output, []string{
		"layer1-start",
		"layer2-middle",
		"layer3-end",
		"dd-preview",
	})
}

func TestPreviewDashDashDeepFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := buildTenLevelFlow()
	flow = append(flow, ":", "--", ":", "dbg.args.tail", "arg1=ten-level-dd")

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	for i := 1; i <= 10; i++ {
		expected := fmt.Sprintf("level%d.key", i)
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
	assertOutputContains(t, output, "ten-level-dd")
}

func TestPreviewDashWithDepthZero(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := buildComplexMultiLayerFlow()
	flow = append(flow, ":", "-", "d=0", ":", "dbg.args.tail", "arg1=depth-zero")

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "depth-zero")
	assertPreviewOutputContains(t, output, []string{
		"layer1-start",
		"layer2-middle",
	})
}

func TestPreviewDashWithDepthZeroDeepFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := buildVeryDeepFlow(20)
	flow = append(flow, ":", "-", "d=0", ":", "dbg.args.tail", "arg1=very-deep-d0")

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "very-deep-d0")
	for i := 1; i <= 20; i++ {
		expected := fmt.Sprintf("deep.level%d", i)
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
}

func TestPreviewPlusSimpleFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{"echo", "message=plus-test", ":", "+", ":", "dbg.args.tail", "arg1=plus"}
	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertPreviewOutputContains(t, output, []string{"echo", "plus-test", "plus"})
}

func TestPreviewPlusComplexFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := buildComplexMultiLayerFlow()
	flow = append(flow, ":", "+", ":", "dbg.args.tail", "arg1=plus-preview")

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertPreviewOutputContains(t, output, []string{
		"layer1-start",
		"layer2-middle",
		"layer3-end",
		"plus-preview",
	})
}

func TestPreviewPlusDeepFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := buildTenLevelFlow()
	flow = append(flow, ":", "+", ":", "dbg.args.tail", "arg1=ten-plus")

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "ten-plus")
	for i := 1; i <= 10; i++ {
		expected := fmt.Sprintf("level%d.key", i)
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
}

func TestPreviewPlusWithMixedCommands(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := buildFlowWithMixedCommands()
	flow = append(flow, ":", "+", ":", "dbg.args.tail", "arg1=plus-mixed")

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertPreviewOutputContains(t, output, []string{
		"mixed.start",
		"step1",
		"step2",
		"plus-mixed",
	})
}

func TestPreviewPlusPlusSimpleFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{"echo", "message=plusplus", ":", "noop", ":", "++", ":", "dbg.args.tail", "arg1=plusplus"}
	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertPreviewOutputContains(t, output, []string{"echo", "plusplus"})
}

func TestPreviewPlusPlusComplexFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := buildComplexMultiLayerFlow()
	flow = append(flow, ":", "++", ":", "dbg.args.tail", "arg1=plusplus-preview")

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertPreviewOutputContains(t, output, []string{
		"layer1-start",
		"layer2-middle",
		"layer3-end",
		"plusplus-preview",
	})
}

func TestPreviewPlusPlusDeepFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := buildVeryDeepFlow(15)
	flow = append(flow, ":", "++", ":", "dbg.args.tail", "arg1=deep-pp")

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "deep-pp")
	for i := 1; i <= 15; i++ {
		expected := fmt.Sprintf("deep.level%d", i)
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
}

func TestPreviewPlusWithDepthZero(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := buildComplexMultiLayerFlow()
	flow = append(flow, ":", "+", "d=0", ":", "dbg.args.tail", "arg1=plus-d0")

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "plus-d0")
	assertPreviewOutputContains(t, output, []string{
		"layer1-start",
		"layer2-middle",
	})
}

func TestPreviewPlusWithDepthZeroVeryDeepFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := buildVeryDeepFlow(25)
	flow = append(flow, ":", "+", "d=0", ":", "dbg.args.tail", "arg1=very-deep-plus-d0")

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "very-deep-plus-d0")
	for i := 1; i <= 25; i++ {
		expected := fmt.Sprintf("deep.level%d", i)
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
}

func TestPreviewDashVsDashDash(t *testing.T) {
	tc1 := NewTiCatForTest()
	screen1 := &previewCaptureScreen{}
	tc1.SetScreen(screen1)
	tc1.Env.SetBool("sys.panic.recover", false)

	flow := buildComplexMultiLayerFlow()
	flow1 := append(flow, ":", "-", ":", "dbg.args.tail", "arg1=dash")
	tc1.RunCli(flow1...)
	output1 := screen1.GetOutput()

	tc2 := NewTiCatForTest()
	screen2 := &previewCaptureScreen{}
	tc2.SetScreen(screen2)
	tc2.Env.SetBool("sys.panic.recover", false)

	flow2 := append(buildComplexMultiLayerFlow(), ":", "--", ":", "dbg.args.tail", "arg1=dashdash")
	tc2.RunCli(flow2...)
	output2 := screen2.GetOutput()

	assertOutputContains(t, output1, "dash")
	assertOutputContains(t, output2, "dashdash")
}

func TestPreviewPlusVsPlusPlus(t *testing.T) {
	tc1 := NewTiCatForTest()
	screen1 := &previewCaptureScreen{}
	tc1.SetScreen(screen1)
	tc1.Env.SetBool("sys.panic.recover", false)

	flow := buildComplexMultiLayerFlow()
	flow1 := append(flow, ":", "+", ":", "dbg.args.tail", "arg1=plus")
	tc1.RunCli(flow1...)
	output1 := screen1.GetOutput()

	tc2 := NewTiCatForTest()
	screen2 := &previewCaptureScreen{}
	tc2.SetScreen(screen2)
	tc2.Env.SetBool("sys.panic.recover", false)

	flow2 := append(buildComplexMultiLayerFlow(), ":", "++", ":", "dbg.args.tail", "arg1=plusplus")
	tc2.RunCli(flow2...)
	output2 := screen2.GetOutput()

	assertOutputContains(t, output1, "plus")
	assertOutputContains(t, output2, "plusplus")
}

func TestPreviewAllModesWithSameFlow(t *testing.T) {
	flow := buildComplexMultiLayerFlow()

	modes := []struct {
		name    string
		preview string
		tailArg string
	}{
		{"dash", "-", "preview-dash"},
		{"dashdash", "--", "preview-dashdash"},
		{"plus", "+", "preview-plus"},
		{"plusplus", "++", "preview-plusplus"},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			tc := NewTiCatForTest()
			screen := &previewCaptureScreen{}
			tc.SetScreen(screen)
			tc.Env.SetBool("sys.panic.recover", false)

			testFlow := make([]string, len(flow))
			copy(testFlow, flow)
			testFlow = append(testFlow, ":", mode.preview, ":", "dbg.args.tail", "arg1="+mode.tailArg)

			ok := tc.RunCli(testFlow...)
			if !ok {
				t.Errorf("%s preview should succeed", mode.name)
			}

			output := screen.GetOutput()
			assertOutputContains(t, output, mode.tailArg)
			assertPreviewOutputContains(t, output, []string{
				"layer1-start",
				"layer2-middle",
				"layer3-end",
			})
		})
	}
}

func TestPreviewAllModesWithDepthZero(t *testing.T) {
	flow := buildComplexMultiLayerFlow()

	modes := []struct {
		name    string
		preview string
		tailArg string
	}{
		{"dash-d0", "-", "preview-dash-d0"},
		{"plus-d0", "+", "preview-plus-d0"},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			tc := NewTiCatForTest()
			screen := &previewCaptureScreen{}
			tc.SetScreen(screen)
			tc.Env.SetBool("sys.panic.recover", false)

			testFlow := make([]string, len(flow))
			copy(testFlow, flow)
			testFlow = append(testFlow, ":", mode.preview, "d=0", ":", "dbg.args.tail", "arg1="+mode.tailArg)

			ok := tc.RunCli(testFlow...)
			if !ok {
				t.Errorf("%s preview with depth=0 should succeed", mode.name)
			}

			output := screen.GetOutput()
			assertOutputContains(t, output, mode.tailArg)
		})
	}
}

func TestPreviewWithGlobalEnv(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"{global.test=global-value}",
		":", "echo", "message=with-global",
		":", "-", ":", "dbg.args.tail", "arg1=global-test",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertPreviewOutputContains(t, output, []string{"with-global", "global-test"})
}

func TestPreviewWithCommandEnv(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"echo", "message=cmd-env", "{local.env=local-value}",
		":", "-", ":", "dbg.args.tail", "arg1=cmd-env-test",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertPreviewOutputContains(t, output, []string{"cmd-env", "cmd-env-test"})
}

func TestPreviewTwentyLevelFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	args := []string{}
	for i := 1; i <= 20; i++ {
		args = append(args, "env.set", fmt.Sprintf("key=l%d.k", i), fmt.Sprintf("value=%d", i))
		if i < 20 {
			args = append(args, ":")
		}
	}
	args = append(args, ":", "-", ":", "dbg.args.tail", "arg1=twenty")

	ok := tc.RunCli(args...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "twenty")
	for i := 1; i <= 20; i++ {
		expected := fmt.Sprintf("l%d.k", i)
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
}

func TestPreviewThirtyLevelFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	args := []string{}
	for i := 1; i <= 30; i++ {
		args = append(args, "env.set", fmt.Sprintf("key=level%d", i), fmt.Sprintf("value=%d", i))
		if i < 30 {
			args = append(args, ":")
		}
	}
	args = append(args, ":", "-", ":", "dbg.args.tail", "arg1=thirty")

	ok := tc.RunCli(args...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "thirty")
}

func TestPreviewWithEchoOnly(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"echo", "message=a",
		":", "echo", "message=b",
		":", "echo", "message=c",
		":", "echo", "message=d",
		":", "echo", "message=e",
		":", "-", ":", "dbg.args.tail", "arg1=echo-only",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertPreviewOutputContains(t, output, []string{"echo-only", "a", "b", "c", "d", "e"})
}

func TestPreviewWithEnvOperationsOnly(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"env.set", "key=a", "value=1",
		":", "env.set", "key=b", "value=2",
		":", "env.set", "key=c", "value=3",
		":", "env.update", "key=a", "value=11",
		":", "-", ":", "dbg.args.tail", "arg1=env-only",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "env-only")
}

func TestPreviewPlusShowsEnvOps(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"env.set", "key=check.key", "value=check-value",
		":", "+", ":", "dbg.args.tail", "arg1=plus-checks",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "plus-checks")
}

func TestPreviewDashVsPlusOutput(t *testing.T) {
	tc1 := NewTiCatForTest()
	screen1 := &previewCaptureScreen{}
	tc1.SetScreen(screen1)
	tc1.Env.SetBool("sys.panic.recover", false)

	flow := buildComplexMultiLayerFlow()
	flow1 := append(flow, ":", "-", ":", "dbg.args.tail", "arg1=test")
	tc1.RunCli(flow1...)
	output1 := screen1.GetOutput()

	tc2 := NewTiCatForTest()
	screen2 := &previewCaptureScreen{}
	tc2.SetScreen(screen2)
	tc2.Env.SetBool("sys.panic.recover", false)

	flow2 := append(buildComplexMultiLayerFlow(), ":", "+", ":", "dbg.args.tail", "arg1=test")
	tc2.RunCli(flow2...)
	output2 := screen2.GetOutput()

	assertOutputContains(t, output1, "test")
	assertOutputContains(t, output2, "test")
}

func TestPreviewMultiplePreviewsInSequence(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"echo", "message=start",
		":", "-", ":", "dbg.args.tail", "arg1=first-preview",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertPreviewOutputContains(t, output, []string{"start", "first-preview"})
}

func TestPreviewWithSpecialChars(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"env.set", "key=path", "value=/path/to/file.conf",
		":", "env.set", "key=host", "value=127.0.0.1",
		":", "-", ":", "dbg.args.tail", "arg1=special",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "special")
}

func TestPreviewWithEmptyValues(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"env.set", "key=empty", "value=",
		":", "-", ":", "dbg.args.tail", "arg1=empty-test",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "empty-test")
}

func TestPreviewNestedEnvModification(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"env.set", "key=nested.val", "value=initial",
		":", "env.update", "key=nested.val", "value=modified1",
		":", "env.update", "key=nested.val", "value=modified2",
		":", "env.update", "key=nested.val", "value=final",
		":", "-", ":", "dbg.args.tail", "arg1=nested-mod",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "nested-mod")
}

func TestPreviewFiftyLevelFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	args := []string{}
	for i := 1; i <= 50; i++ {
		args = append(args, "env.set", fmt.Sprintf("key=fifty.l%d", i), fmt.Sprintf("value=%d", i))
		if i < 50 {
			args = append(args, ":")
		}
	}
	args = append(args, ":", "-", ":", "dbg.args.tail", "arg1=fifty")

	ok := tc.RunCli(args...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "fifty")
}

func TestPreviewOrderPreservation(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"echo", "message=first",
		":", "echo", "message=second",
		":", "echo", "message=third",
		":", "-", ":", "dbg.args.tail", "arg1=order-test",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "order-test")

	idx1 := strings.Index(output, "first")
	idx2 := strings.Index(output, "second")
	idx3 := strings.Index(output, "third")

	if !(idx1 < idx2 && idx2 < idx3) {
		t.Errorf("commands should appear in order, got indices: %d, %d, %d", idx1, idx2, idx3)
	}
}

func TestFilterSingleMatch(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"echo", "message=a",
		":", "echo", "message=b",
		":", "echo", "message=c",
		":", "-", "filter=echo", ":", "dbg.args.tail", "arg1=filter-test",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "filter-test")
	assertOutputContains(t, output, "echo")
}

func TestFilterMultipleMatches(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"env.set", "key=a", "value=1",
		":", "env.set", "key=b", "value=2",
		":", "echo", "message=c",
		":", "-", "filter=env", ":", "dbg.args.tail", "arg1=multi-filter",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "multi-filter")
	assertOutputContains(t, output, "env.set")
	assertOutputContains(t, output, "[env.set]")
	if strings.Contains(output, "[echo]") {
		t.Errorf("output should NOT contain [echo] after filtering")
	}
}

func TestFilterCommaSeparated(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"env.set", "key=a", "value=1",
		":", "echo", "message=b",
		":", "noop",
		":", "-", "filter=env,echo", ":", "dbg.args.tail", "arg1=comma-filter",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "comma-filter")
	assertOutputContains(t, output, "env.set")
	assertOutputContains(t, output, "[echo]")
	if strings.Contains(output, "[noop]") {
		t.Errorf("output should NOT contain [noop] after filtering")
	}
}

func TestFilterNoMatch(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"echo", "message=a",
		":", "echo", "message=b",
		":", "-", "filter=nonexistent", ":", "dbg.args.tail", "arg1=no-match",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "no-match")
}

func TestFilterPlusMode(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"env.set", "key=a", "value=1",
		":", "echo", "message=b",
		":", "+", "filter=echo", ":", "dbg.args.tail", "arg1=plus-filter",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "plus-filter")
	assertOutputContains(t, output, "[echo]")
	if strings.Contains(output, "[env.set]") {
		t.Errorf("output should NOT contain [env.set] after filtering")
	}
}

func TestFilterDashDashMode(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"env.set", "key=a", "value=1",
		":", "echo", "message=b",
		":", "noop",
		":", "--", "filter=env", ":", "dbg.args.tail", "arg1=dd-filter",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "dd-filter")
	assertOutputContains(t, output, "[env.set]")
	if strings.Contains(output, "[echo]") {
		t.Errorf("output should NOT contain [echo] after filtering")
	}
	if strings.Contains(output, "[noop]") {
		t.Errorf("output should NOT contain [noop] after filtering")
	}
}

func TestFilterTwentyLevelFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	args := []string{}
	for i := 1; i <= 20; i++ {
		if i%3 == 0 {
			args = append(args, "echo", fmt.Sprintf("message=echo%d", i))
		} else {
			args = append(args, "env.set", fmt.Sprintf("key=level%d", i), fmt.Sprintf("value=%d", i))
		}
		if i < 20 {
			args = append(args, ":")
		}
	}
	args = append(args, ":", "-", "filter=echo", ":", "dbg.args.tail", "arg1=deep-filter")

	ok := tc.RunCli(args...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "deep-filter")
	assertOutputContains(t, output, "[echo]")
	if strings.Contains(output, "[env.set]") {
		t.Errorf("output should NOT contain [env.set] after filtering")
	}
}

func TestFilterThirtyLevelFlowMixed(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	args := []string{}
	for i := 1; i <= 30; i++ {
		if i%5 == 0 {
			args = append(args, "noop")
		} else if i%3 == 0 {
			args = append(args, "echo", fmt.Sprintf("message=msg%d", i))
		} else {
			args = append(args, "env.set", fmt.Sprintf("key=k%d", i), fmt.Sprintf("value=%d", i))
		}
		if i < 30 {
			args = append(args, ":")
		}
	}
	args = append(args, ":", "-", "filter=echo,noop", ":", "dbg.args.tail", "arg1=thirty-filter")

	ok := tc.RunCli(args...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "thirty-filter")
	assertOutputContains(t, output, "[echo]")
	assertOutputContains(t, output, "[noop]")
	if strings.Contains(output, "[env.set]") {
		t.Errorf("output should NOT contain [env.set] after filtering")
	}
}

func TestFilterFiftyLevelFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	args := []string{}
	for i := 1; i <= 50; i++ {
		if i == 25 || i == 50 {
			args = append(args, "echo", fmt.Sprintf("message=marker%d", i))
		} else {
			args = append(args, "env.set", fmt.Sprintf("key=f%d", i), fmt.Sprintf("value=%d", i))
		}
		if i < 50 {
			args = append(args, ":")
		}
	}
	args = append(args, ":", "-", "filter=echo", ":", "dbg.args.tail", "arg1=fifty-filter")

	ok := tc.RunCli(args...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "fifty-filter")
	assertOutputContains(t, output, "echo")
	assertOutputContains(t, output, "marker25")
	assertOutputContains(t, output, "marker50")
}

func TestFilterPartialMatch(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"env.set", "key=a", "value=1",
		":", "env.update", "key=a", "value=2",
		":", "echo", "message=b",
		":", "-", "filter=env", ":", "dbg.args.tail", "arg1=partial",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "partial")
	assertOutputContains(t, output, "[env.set]")
	assertOutputContains(t, output, "[env.update]")
	if strings.Contains(output, "[echo]") {
		t.Errorf("output should NOT contain [echo] after filtering")
	}
}

func TestFilterWithDepthArg(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"env.set", "key=a", "value=1",
		":", "echo", "message=b",
		":", "env.set", "key=c", "value=3",
		":", "-", "filter=env", "d=0", ":", "dbg.args.tail", "arg1=filter-depth",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "filter-depth")
	assertOutputContains(t, output, "[env.set]")
	if strings.Contains(output, "[echo]") {
		t.Errorf("output should NOT contain [echo] after filtering")
	}
}

func TestFilterEmptyValue(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"echo", "message=a",
		":", "echo", "message=b",
		":", "-", "filter=", ":", "dbg.args.tail", "arg1=empty-filter",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "empty-filter")
}

func TestFilterWhitespaceInValue(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"echo", "message=a",
		":", "echo", "message=b",
		":", "-", "filter={echo , env}", ":", "dbg.args.tail", "arg1=ws-filter",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "ws-filter")
}

func TestFilterPlusPlusMode(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := []string{
		"env.set", "key=a", "value=1",
		":", "echo", "message=b",
		":", "noop",
		":", "++", "filter=echo", ":", "dbg.args.tail", "arg1=pp-filter",
	}

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "pp-filter")
	assertOutputContains(t, output, "[echo]")
	if strings.Contains(output, "[env.set]") {
		t.Errorf("output should NOT contain [env.set] after filtering")
	}
	if strings.Contains(output, "[noop]") {
		t.Errorf("output should NOT contain [noop] after filtering")
	}
}

func TestFilterComplexFlow(t *testing.T) {
	tc := NewTiCatForTest()
	screen := &previewCaptureScreen{}
	tc.SetScreen(screen)
	tc.Env.SetBool("sys.panic.recover", false)

	flow := buildComplexMultiLayerFlow()
	flow = append(flow, ":", "-", "filter=echo", ":", "dbg.args.tail", "arg1=complex-filter")

	ok := tc.RunCli(flow...)
	if !ok {
		t.Error("command should succeed")
	}

	output := screen.GetOutput()
	assertOutputContains(t, output, "complex-filter")
	assertOutputContains(t, output, "[echo]")
	if strings.Contains(output, "[env.set]") {
		t.Errorf("output should NOT contain [env.set] after filtering")
	}
}

func TestFilterAllModesConsistency(t *testing.T) {
	flow := []string{
		"env.set", "key=a", "value=1",
		":", "echo", "message=b",
		":", "noop",
	}

	modes := []struct {
		name    string
		preview string
	}{
		{"dash", "-"},
		{"dashdash", "--"},
		{"plus", "+"},
		{"plusplus", "++"},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			tc := NewTiCatForTest()
			screen := &previewCaptureScreen{}
			tc.SetScreen(screen)
			tc.Env.SetBool("sys.panic.recover", false)

			testFlow := make([]string, len(flow))
			copy(testFlow, flow)
			testFlow = append(testFlow, ":", mode.preview, "filter=echo", ":", "dbg.args.tail", "arg1=test")

			ok := tc.RunCli(testFlow...)
			if !ok {
				t.Errorf("%s filter should succeed", mode.name)
			}

			output := screen.GetOutput()
			assertOutputContains(t, output, "test")
			assertOutputContains(t, output, "[echo]")
			if strings.Contains(output, "[env.set]") {
				t.Errorf("%s: output should NOT contain [env.set] after filtering", mode.name)
			}
			if strings.Contains(output, "[noop]") {
				t.Errorf("%s: output should NOT contain [noop] after filtering", mode.name)
			}
		})
	}
}

func assertPreviewOutputContains(t *testing.T, output string, expected []string) {
	t.Helper()
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain %q\nfull output:\n%s", exp, output)
		}
	}
}
