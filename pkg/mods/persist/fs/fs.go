package fs

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type FS interface {
	Open(name string) (File, error)
	Create(name string) (File, error)
	OpenFile(name string, flag int, perm os.FileMode) (File, error)
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)
	Remove(name string) error
	RemoveAll(path string) error
	Rename(oldpath, newpath string) error
	Exists(name string) bool
	IsNotExist(err error) bool
}

type File interface {
	io.Reader
	io.Writer
	io.Closer
	Name() string
	Stat() (os.FileInfo, error)
}

type RealFS struct{}

func NewRealFS() *RealFS {
	return &RealFS{}
}

func (f *RealFS) Open(name string) (File, error) {
	return os.Open(name)
}

func (f *RealFS) Create(name string) (File, error) {
	return os.Create(name)
}

func (f *RealFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return os.OpenFile(name, flag, perm)
}

func (f *RealFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (f *RealFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (f *RealFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (f *RealFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (f *RealFS) Remove(name string) error {
	return os.Remove(name)
}

func (f *RealFS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (f *RealFS) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (f *RealFS) Exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func (f *RealFS) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

type MemFS struct {
	mu    sync.RWMutex
	files map[string]*memFile
	dirs  map[string]bool
}

func NewMemFS() *MemFS {
	return &MemFS{
		files: make(map[string]*memFile),
		dirs:  make(map[string]bool),
	}
}

func (f *MemFS) Open(name string) (File, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	name = normalizePath(name)
	file, ok := f.files[name]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}
	return &memFileReader{memFile: file, name: name}, nil
}

func (f *MemFS) Create(name string) (File, error) {
	return f.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
}

func (f *MemFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	name = normalizePath(name)

	if flag&os.O_CREATE != 0 {
		f.ensureDir(filepath.Dir(name))
	}

	file, ok := f.files[name]
	if !ok {
		if flag&os.O_CREATE == 0 {
			return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
		}
		file = &memFile{
			name:    name,
			data:    &bytes.Buffer{},
			modTime: time.Now(),
			mode:    perm,
		}
		f.files[name] = file
	}

	if flag&os.O_TRUNC != 0 {
		file.data = &bytes.Buffer{}
		file.modTime = time.Now()
	}

	return &memFileWriter{memFile: file, name: name, fs: f}, nil
}

func (f *MemFS) ReadFile(name string) ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	name = normalizePath(name)
	file, ok := f.files[name]
	if !ok {
		return nil, &os.PathError{Op: "readfile", Path: name, Err: os.ErrNotExist}
	}
	return file.data.Bytes(), nil
}

func (f *MemFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	name = normalizePath(name)
	f.ensureDir(filepath.Dir(name))
	f.files[name] = &memFile{
		name:    name,
		data:    bytes.NewBuffer(data),
		modTime: time.Now(),
		mode:    perm,
	}
	return nil
}

func (f *MemFS) MkdirAll(path string, perm os.FileMode) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	path = normalizePath(path)
	f.ensureDir(path)
	return nil
}

func (f *MemFS) ensureDir(path string) {
	path = normalizePath(path)
	if path == "" || path == "." {
		return
	}
	for p := path; p != ""; p = filepath.Dir(p) {
		f.dirs[p] = true
		if p == filepath.Dir(p) {
			break
		}
	}
}

func (f *MemFS) Stat(name string) (os.FileInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	name = normalizePath(name)
	if file, ok := f.files[name]; ok {
		return &memFileInfo{name: name, file: file}, nil
	}
	if f.dirs[name] {
		return &memDirInfo{name: name}, nil
	}
	return nil, &os.PathError{Op: "stat", Path: name, Err: os.ErrNotExist}
}

func (f *MemFS) Remove(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	name = normalizePath(name)
	if _, ok := f.files[name]; ok {
		delete(f.files, name)
		return nil
	}
	if f.dirs[name] {
		for p := range f.files {
			if strings.HasPrefix(p, name+string(filepath.Separator)) {
				return &os.PathError{Op: "remove", Path: name, Err: fmt.Errorf("directory not empty")}
			}
		}
		delete(f.dirs, name)
		return nil
	}
	return &os.PathError{Op: "remove", Path: name, Err: os.ErrNotExist}
}

func (f *MemFS) RemoveAll(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	path = normalizePath(path)
	for name := range f.files {
		if strings.HasPrefix(name, path+string(filepath.Separator)) || name == path {
			delete(f.files, name)
		}
	}
	for name := range f.dirs {
		if strings.HasPrefix(name, path+string(filepath.Separator)) || name == path {
			delete(f.dirs, name)
		}
	}
	delete(f.dirs, path)
	return nil
}

func (f *MemFS) Rename(oldpath, newpath string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	oldpath = normalizePath(oldpath)
	newpath = normalizePath(newpath)

	file, ok := f.files[oldpath]
	if !ok {
		return &os.PathError{Op: "rename", Path: oldpath, Err: os.ErrNotExist}
	}

	f.ensureDir(filepath.Dir(newpath))
	file.name = newpath
	file.modTime = time.Now()
	f.files[newpath] = file
	delete(f.files, oldpath)
	return nil
}

func (f *MemFS) Exists(name string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	name = normalizePath(name)
	_, ok := f.files[name]
	if ok {
		return true
	}
	return f.dirs[name]
}

func (f *MemFS) IsNotExist(err error) bool {
	if err == nil {
		return false
	}
	if pathErr, ok := err.(*os.PathError); ok {
		return pathErr.Err == os.ErrNotExist
	}
	return false
}

func normalizePath(path string) string {
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, "/")
	return path
}

type memFile struct {
	name    string
	data    *bytes.Buffer
	modTime time.Time
	mode    os.FileMode
}

type memFileReader struct {
	*memFile
	name string
	pos  int
}

func (f *memFileReader) Read(p []byte) (n int, err error) {
	data := f.memFile.data.Bytes()
	if f.pos >= len(data) {
		return 0, io.EOF
	}
	n = copy(p, data[f.pos:])
	f.pos += n
	return n, nil
}

func (f *memFileReader) Close() error {
	return nil
}

func (f *memFileReader) Name() string {
	return f.name
}

func (f *memFileReader) Stat() (os.FileInfo, error) {
	return &memFileInfo{name: f.name, file: f.memFile}, nil
}

func (f *memFileReader) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read-only file")
}

type memFileWriter struct {
	*memFile
	name string
	fs   *MemFS
}

func (f *memFileWriter) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("write-only file")
}

func (f *memFileWriter) Write(p []byte) (n int, err error) {
	n, err = f.memFile.data.Write(p)
	f.memFile.modTime = time.Now()
	return
}

func (f *memFileWriter) Close() error {
	return nil
}

func (f *memFileWriter) Name() string {
	return f.name
}

func (f *memFileWriter) Stat() (os.FileInfo, error) {
	return &memFileInfo{name: f.name, file: f.memFile}, nil
}

type memFileInfo struct {
	name string
	file *memFile
}

func (i *memFileInfo) Name() string {
	return filepath.Base(i.name)
}

func (i *memFileInfo) Size() int64 {
	return int64(i.file.data.Len())
}

func (i *memFileInfo) Mode() os.FileMode {
	return i.file.mode
}

func (i *memFileInfo) ModTime() time.Time {
	return i.file.modTime
}

func (i *memFileInfo) IsDir() bool {
	return false
}

func (i *memFileInfo) Sys() interface{} {
	return nil
}

type memDirInfo struct {
	name string
}

func (i *memDirInfo) Name() string {
	return filepath.Base(i.name)
}

func (i *memDirInfo) Size() int64 {
	return 0
}

func (i *memDirInfo) Mode() os.FileMode {
	return os.ModeDir | 0755
}

func (i *memDirInfo) ModTime() time.Time {
	return time.Now()
}

func (i *memDirInfo) IsDir() bool {
	return true
}

func (i *memDirInfo) Sys() interface{} {
	return nil
}

func (f *MemFS) ListFiles(prefix string) []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var result []string
	for name := range f.files {
		if prefix == "" || strings.HasPrefix(name, prefix) {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result
}
