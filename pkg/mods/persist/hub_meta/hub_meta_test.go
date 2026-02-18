package hub_meta

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteReposInfoFile_DoubleCloseFix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hub_meta_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hubDir := filepath.Join(tmpDir, "hub")
	metaPath := filepath.Join(hubDir, "repos.hub")

	infos := []RepoInfo{
		{
			Addr:      ParseRepoAddr("https://github.com/test/repo1"),
			AddReason: "test",
			Path:      "/path/to/repo1",
			HelpStr:   "test repo 1",
			OnOff:     "on",
		},
		{
			Addr:      ParseRepoAddr("https://github.com/test/repo2"),
			AddReason: "test",
			Path:      "/path/to/repo2",
			HelpStr:   "test repo 2",
			OnOff:     "off",
		},
	}

	WriteReposInfoFile(hubDir, metaPath, infos, "|")

	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Fatal("expected repos.hub file to be created")
	}

	readInfos, _, _ := ReadReposInfoFile(hubDir, metaPath, false, "|")

	if len(readInfos) != len(infos) {
		t.Errorf("expected %d infos, got %d", len(infos), len(readInfos))
	}

	for i, info := range readInfos {
		if info.Addr.Str() != infos[i].Addr.Str() {
			t.Errorf("expected addr %s, got %s", infos[i].Addr.Str(), info.Addr.Str())
		}
		if info.AddReason != infos[i].AddReason {
			t.Errorf("expected add reason %s, got %s", infos[i].AddReason, info.AddReason)
		}
		if info.HelpStr != infos[i].HelpStr {
			t.Errorf("expected help str %s, got %s", infos[i].HelpStr, info.HelpStr)
		}
		if info.OnOff != infos[i].OnOff {
			t.Errorf("expected on/off %s, got %s", infos[i].OnOff, info.OnOff)
		}
	}
}

func TestWriteReposInfoFile_EmptyList(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hub_meta_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hubDir := filepath.Join(tmpDir, "hub")
	metaPath := filepath.Join(hubDir, "repos.hub")

	WriteReposInfoFile(hubDir, metaPath, []RepoInfo{}, "|")

	readInfos, _, _ := ReadReposInfoFile(hubDir, metaPath, false, "|")

	if len(readInfos) != 0 {
		t.Errorf("expected 0 infos, got %d", len(readInfos))
	}
}

func TestWriteReposInfoFile_Overwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hub_meta_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hubDir := filepath.Join(tmpDir, "hub")
	metaPath := filepath.Join(hubDir, "repos.hub")

	infos1 := []RepoInfo{
		{
			Addr:      ParseRepoAddr("https://github.com/test/repo1"),
			AddReason: "test1",
			Path:      "/path/to/repo1",
			HelpStr:   "test repo 1",
			OnOff:     "on",
		},
	}

	WriteReposInfoFile(hubDir, metaPath, infos1, "|")

	infos2 := []RepoInfo{
		{
			Addr:      ParseRepoAddr("https://github.com/test/repo2"),
			AddReason: "test2",
			Path:      "/path/to/repo2",
			HelpStr:   "test repo 2",
			OnOff:     "off",
		},
	}

	WriteReposInfoFile(hubDir, metaPath, infos2, "|")

	readInfos, _, _ := ReadReposInfoFile(hubDir, metaPath, false, "|")

	if len(readInfos) != 1 {
		t.Fatalf("expected 1 info, got %d", len(readInfos))
	}

	if readInfos[0].Addr.Str() != "https://github.com/test/repo2" {
		t.Errorf("expected repo2, got %s", readInfos[0].Addr.Str())
	}
}

func TestReadReposInfoFile_NotExist(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hub_meta_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hubDir := filepath.Join(tmpDir, "hub")
	metaPath := filepath.Join(hubDir, "repos.hub")

	infos, _, _ := ReadReposInfoFile(hubDir, metaPath, true, "|")

	if len(infos) != 0 {
		t.Errorf("expected 0 infos for non-existent file, got %d", len(infos))
	}
}

func TestReadReposInfoFile_NotExistError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hub_meta_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hubDir := filepath.Join(tmpDir, "hub")
	metaPath := filepath.Join(hubDir, "repos.hub")

	_, _, err = ReadReposInfoFile(hubDir, metaPath, false, "|")
	if err == nil {
		t.Error("expected error for non-existent file when allowNotExist is false")
	}
}
