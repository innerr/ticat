package builtin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/innerr/ticat/pkg/core/model"
)

func TestLoadFlowAfterSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ticat-flow-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	flowDir := filepath.Join(tmpDir, "flows")
	if err := os.MkdirAll(flowDir, os.ModePerm); err != nil {
		t.Fatalf("failed to create flows dir: %v", err)
	}

	strs := model.CmdTreeStrsForTest()
	tree := model.NewCmdTree(strs)
	cc := &model.Cli{
		Screen:   &model.QuietScreen{},
		Cmds:     tree,
		EnvAbbrs: model.NewEnvAbbrs("test"),
	}

	env := model.NewEnvEx(model.EnvLayerDefault).NewLayer(model.EnvLayerSession)
	env.SetBool("sys.panic.recover", true)
	env.Set("sys.paths.flows", flowDir)
	env.Set("strs.flow-ext", ".tiflow")
	env.Set("strs.meta-ext", ".meta")
	env.Set("strs.abbrs-sep", "|")
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.self-name", "ticat")

	flowPath := filepath.Join(flowDir, "test_direct_load.tiflow")
	flowContent := `display.utf8.off : echo message=test`
	if err := os.WriteFile(flowPath, []byte(flowContent), 0644); err != nil {
		t.Fatalf("failed to write flow file: %v", err)
	}

	loadFlowAfterSave(cc, env, flowPath, "test_direct_load")

	loadedCmd := cc.Cmds.GetSub("test_direct_load")
	if loadedCmd == nil {
		t.Error("flow was not loaded after calling loadFlowAfterSave")
	}
}

func TestLoadFlowAfterSaveWithSubPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ticat-flow-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	flowDir := filepath.Join(tmpDir, "flows")
	if err := os.MkdirAll(flowDir, os.ModePerm); err != nil {
		t.Fatalf("failed to create flows dir: %v", err)
	}

	strs := model.CmdTreeStrsForTest()
	tree := model.NewCmdTree(strs)
	cc := &model.Cli{
		Screen:   &model.QuietScreen{},
		Cmds:     tree,
		EnvAbbrs: model.NewEnvAbbrs("test"),
	}

	env := model.NewEnvEx(model.EnvLayerDefault).NewLayer(model.EnvLayerSession)
	env.SetBool("sys.panic.recover", true)
	env.Set("sys.paths.flows", flowDir)
	env.Set("strs.flow-ext", ".tiflow")
	env.Set("strs.meta-ext", ".meta")
	env.Set("strs.abbrs-sep", "|")
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.self-name", "ticat")

	flowPath := filepath.Join(flowDir, "myns.mysub.flow.tiflow")
	flowContent := `display.utf8.off : echo message=test`
	if err := os.WriteFile(flowPath, []byte(flowContent), 0644); err != nil {
		t.Fatalf("failed to write flow file: %v", err)
	}

	loadFlowAfterSave(cc, env, flowPath, "myns.mysub.flow")

	mynsCmd := cc.Cmds.GetSub("myns")
	if mynsCmd == nil {
		t.Fatal("namespace 'myns' was not created")
	}

	mysubCmd := mynsCmd.GetSub("mysub")
	if mysubCmd == nil {
		t.Fatal("sub-namespace 'myns.mysub' was not created")
	}

	flowCmd := mysubCmd.GetSub("flow")
	if flowCmd == nil {
		t.Error("flow 'myns.mysub.flow' was not loaded")
	}
}

func TestLoadFlowAfterSaveOverwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ticat-flow-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	flowDir := filepath.Join(tmpDir, "flows")
	if err := os.MkdirAll(flowDir, os.ModePerm); err != nil {
		t.Fatalf("failed to create flows dir: %v", err)
	}

	strs := model.CmdTreeStrsForTest()
	tree := model.NewCmdTree(strs)
	cc := &model.Cli{
		Screen:   &model.QuietScreen{},
		Cmds:     tree,
		EnvAbbrs: model.NewEnvAbbrs("test"),
	}

	env := model.NewEnvEx(model.EnvLayerDefault).NewLayer(model.EnvLayerSession)
	env.SetBool("sys.panic.recover", true)
	env.Set("sys.paths.flows", flowDir)
	env.Set("strs.flow-ext", ".tiflow")
	env.Set("strs.meta-ext", ".meta")
	env.Set("strs.abbrs-sep", "|")
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.self-name", "ticat")

	flowPath := filepath.Join(flowDir, "overwrite_test.tiflow")
	flowContent1 := `echo message=first`
	if err := os.WriteFile(flowPath, []byte(flowContent1), 0644); err != nil {
		t.Fatalf("failed to write flow file: %v", err)
	}

	loadFlowAfterSave(cc, env, flowPath, "overwrite_test")

	loadedCmd := cc.Cmds.GetSub("overwrite_test")
	if loadedCmd == nil {
		t.Fatal("flow was not loaded on first call")
	}

	flowContent2 := `echo message=second`
	if err := os.WriteFile(flowPath, []byte(flowContent2), 0644); err != nil {
		t.Fatalf("failed to write updated flow file: %v", err)
	}

	loadFlowAfterSave(cc, env, flowPath, "overwrite_test")

	loadedCmd2 := cc.Cmds.GetSub("overwrite_test")
	if loadedCmd2 == nil {
		t.Error("flow was not loaded after overwrite")
	}
}

func TestLoadFlowAfterSaveWithHelp(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ticat-flow-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	flowDir := filepath.Join(tmpDir, "flows")
	if err := os.MkdirAll(flowDir, os.ModePerm); err != nil {
		t.Fatalf("failed to create flows dir: %v", err)
	}

	strs := model.CmdTreeStrsForTest()
	tree := model.NewCmdTree(strs)
	cc := &model.Cli{
		Screen:   &model.QuietScreen{},
		Cmds:     tree,
		EnvAbbrs: model.NewEnvAbbrs("test"),
	}

	env := model.NewEnvEx(model.EnvLayerDefault).NewLayer(model.EnvLayerSession)
	env.SetBool("sys.panic.recover", true)
	env.Set("sys.paths.flows", flowDir)
	env.Set("strs.flow-ext", ".tiflow")
	env.Set("strs.meta-ext", ".meta")
	env.Set("strs.abbrs-sep", "|")
	env.Set("strs.env-path-sep", ".")
	env.Set("strs.self-name", "ticat")

	flowPath := filepath.Join(flowDir, "with_help.tiflow")
	flowContent := `# help: this is a test flow
display.utf8.off : echo message=test`
	if err := os.WriteFile(flowPath, []byte(flowContent), 0644); err != nil {
		t.Fatalf("failed to write flow file: %v", err)
	}

	loadFlowAfterSave(cc, env, flowPath, "with_help")

	loadedCmd := cc.Cmds.GetSub("with_help")
	if loadedCmd == nil {
		t.Error("flow with help was not loaded")
	}

	if loadedCmd.Cmd() != nil && loadedCmd.Cmd().Help() != "this is a test flow" {
		t.Logf("flow help: %s", loadedCmd.Cmd().Help())
	}
}
