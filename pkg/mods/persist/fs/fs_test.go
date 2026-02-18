package fs

import (
	"os"
	"testing"
)

func TestMemFS_WriteAndReadFile(t *testing.T) {
	fs := NewMemFS()

	err := fs.WriteFile("test.txt", []byte("hello world"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	data, err := fs.ReadFile("test.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(data))
	}
}

func TestMemFS_MkdirAll(t *testing.T) {
	fs := NewMemFS()

	err := fs.MkdirAll("a/b/c", 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	if !fs.Exists("a/b/c") {
		t.Error("directory should exist")
	}
}

func TestMemFS_WriteFileCreatesDirs(t *testing.T) {
	fs := NewMemFS()

	err := fs.WriteFile("a/b/c/test.txt", []byte("data"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if !fs.Exists("a/b/c") {
		t.Error("parent directories should be created")
	}
}

func TestMemFS_Remove(t *testing.T) {
	fs := NewMemFS()

	fs.WriteFile("test.txt", []byte("data"), 0644)

	err := fs.Remove("test.txt")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if fs.Exists("test.txt") {
		t.Error("file should be removed")
	}
}

func TestMemFS_RemoveAll(t *testing.T) {
	fs := NewMemFS()

	fs.WriteFile("a/b/c/test.txt", []byte("data"), 0644)
	fs.WriteFile("a/b/d/test.txt", []byte("data"), 0644)

	err := fs.RemoveAll("a/b")
	if err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}

	if fs.Exists("a/b/c/test.txt") {
		t.Error("file should be removed")
	}
	if fs.Exists("a/b") {
		t.Error("directory should be removed")
	}
}

func TestMemFS_Rename(t *testing.T) {
	fs := NewMemFS()

	fs.WriteFile("old.txt", []byte("data"), 0644)

	err := fs.Rename("old.txt", "new.txt")
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	if fs.Exists("old.txt") {
		t.Error("old file should not exist")
	}
	if !fs.Exists("new.txt") {
		t.Error("new file should exist")
	}

	data, _ := fs.ReadFile("new.txt")
	if string(data) != "data" {
		t.Errorf("content mismatch: %q", string(data))
	}
}

func TestMemFS_OpenFile_Truncate(t *testing.T) {
	fs := NewMemFS()

	fs.WriteFile("test.txt", []byte("old content"), 0644)

	f, err := fs.OpenFile("test.txt", os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	f.Write([]byte("new"))
	f.Close()

	data, _ := fs.ReadFile("test.txt")
	if string(data) != "new" {
		t.Errorf("expected 'new', got %q", string(data))
	}
}

func TestMemFS_Stat(t *testing.T) {
	fs := NewMemFS()

	fs.WriteFile("test.txt", []byte("hello"), 0644)

	info, err := fs.Stat("test.txt")
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if info.Name() != "test.txt" {
		t.Errorf("expected name 'test.txt', got %q", info.Name())
	}

	if info.Size() != 5 {
		t.Errorf("expected size 5, got %d", info.Size())
	}

	if info.IsDir() {
		t.Error("should not be a directory")
	}
}

func TestMemFS_StatDir(t *testing.T) {
	fs := NewMemFS()

	fs.MkdirAll("a/b/c", 0755)

	info, err := fs.Stat("a/b/c")
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if !info.IsDir() {
		t.Error("should be a directory")
	}
}

func TestMemFS_IsNotExist(t *testing.T) {
	fs := NewMemFS()

	_, err := fs.Stat("nonexistent")
	if !fs.IsNotExist(err) {
		t.Error("IsNotExist should return true for nonexistent file")
	}
}

func TestMemFS_OpenReader(t *testing.T) {
	fs := NewMemFS()

	fs.WriteFile("test.txt", []byte("hello"), 0644)

	f, err := fs.Open("test.txt")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()

	buf := make([]byte, 5)
	n, err := f.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if n != 5 || string(buf) != "hello" {
		t.Errorf("expected 'hello', got %q", string(buf[:n]))
	}
}

func TestMemFS_Open_NonExistent(t *testing.T) {
	fs := NewMemFS()

	_, err := fs.Open("nonexistent.txt")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestMemFS_ListFiles(t *testing.T) {
	fs := NewMemFS()

	fs.WriteFile("a/1.txt", []byte("1"), 0644)
	fs.WriteFile("a/2.txt", []byte("2"), 0644)
	fs.WriteFile("b/3.txt", []byte("3"), 0644)

	files := fs.ListFiles("a/")
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(files), files)
	}
}
