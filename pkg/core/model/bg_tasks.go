package model

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

type CmdIO struct {
	CmdStdin  io.Reader
	CmdStdout io.Writer
	CmdStderr io.Writer
}

func NewCmdIO(stdio io.Reader, stdout io.Writer, stderr io.Writer) *CmdIO {
	return &CmdIO{stdio, stdout, stderr}
}

func (self *CmdIO) SetupForExec(cmd *exec.Cmd, logFilePath string) (logger io.WriteCloser) {
	if len(logFilePath) != 0 {
		file, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_SYNC, 0644)
		if err != nil {
			// Runtime error: cannot open log file - return nil logger
			return nil
		}
		logger = file
	}

	if self.CmdStdin != nil {
		cmd.Stdin = self.CmdStdin
	}

	if logger == nil {
		if self.CmdStdout != nil {
			cmd.Stdout = self.CmdStdout
		}
		if self.CmdStderr != nil {
			cmd.Stderr = self.CmdStderr
		}
	} else {
		if self.CmdStdout != nil {
			cmd.Stdout = io.MultiWriter(self.CmdStdout, logger)
		} else {
			cmd.Stdout = logger
		}
		if self.CmdStderr != nil {
			cmd.Stderr = io.MultiWriter(self.CmdStderr, logger)
		} else {
			cmd.Stderr = self.CmdStderr
		}
	}
	return
}

type BgStdout struct {
	bg   *bytes.Buffer
	fg   io.Writer
	lock sync.Mutex
}

func NewBgStdout() *BgStdout {
	return &BgStdout{
		bg: bytes.NewBuffer(nil),
	}
}

func (self *BgStdout) BringToFront(fg io.Writer) {
	self.lock.Lock()
	defer self.lock.Unlock()
	_, err := io.Copy(fg, self.bg)
	if err != nil {
		// Runtime error: cannot copy buffer to foreground - silently ignore
	}
	self.fg = fg
	self.bg = nil
}

func (self *BgStdout) Write(p []byte) (n int, err error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if self.fg != nil {
		return self.fg.Write(p)
	}
	return self.bg.Write(p)
}

type BgTaskInfo struct {
	Tid      string
	Cmd      string
	Started  bool
	Finished bool
	Err      error
}

type BgTask struct {
	info           BgTaskInfo
	realCmd        string
	stdout         *BgStdout
	finishNotifier chan error
	lock           sync.Mutex
}

func NewBgTask(tid string, cmd string, realCmd string, stdout *BgStdout) *BgTask {
	return &BgTask{
		info: BgTaskInfo{
			Tid: tid,
			Cmd: cmd,
		},
		realCmd:        realCmd,
		stdout:         stdout,
		finishNotifier: make(chan error),
	}
}

func (self *BgTask) OnStart() {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.info.Started = true
}

func (self *BgTask) GetStat() BgTaskInfo {
	self.lock.Lock()
	defer self.lock.Unlock()
	return self.info
}

func (self *BgTask) OnFinish(err error) {
	self.lock.Lock()
	self.info.Finished = true
	self.info.Err = err
	self.lock.Unlock()
	self.finishNotifier <- err
}

func (self *BgTask) WaitForFinish() error {
	return <-self.finishNotifier
}

type BgTasks struct {
	tids      []string
	tasks     map[string]*BgTask
	name2task map[string]*BgTask
	lock      sync.Mutex
}

func NewBgTasks() *BgTasks {
	return &BgTasks{
		tids:      []string{},
		tasks:     map[string]*BgTask{},
		name2task: map[string]*BgTask{},
	}
}

func (self *BgTasks) GetOrAddTask(tid string, displayName string, realName string, stdout *BgStdout) *BgTask {
	self.lock.Lock()
	defer self.lock.Unlock()
	task, ok := self.tasks[tid]
	if ok {
		return task
	}
	self.tids = append(self.tids, tid)
	task = NewBgTask(tid, displayName, realName, stdout)
	self.tasks[tid] = task
	println(realName)
	self.name2task[realName] = task
	return task
}

func (self *BgTasks) GetEarliestTask() (tid string, task *BgTask, ok bool) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if len(self.tids) == 0 {
		return
	}
	tid = self.tids[0]
	task, ok = self.tasks[tid]
	return
}

func (self *BgTasks) GetTaskByCmd(cmd string) (tid string, task *BgTask, ok bool) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if len(self.tids) == 0 {
		return
	}
	task, ok = self.name2task[cmd]
	if ok {
		tid = task.info.Tid
	}
	return
}

func (self *BgTasks) GetLatestTask() (tid string, task *BgTask, ok bool) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if len(self.tids) == 0 {
		return
	}
	tid = self.tids[len(self.tids)-1]
	task, ok = self.tasks[tid]
	return
}

func (self *BgTasks) GetStat() []BgTaskInfo {
	self.lock.Lock()
	defer self.lock.Unlock()
	infos := make([]BgTaskInfo, len(self.tids))
	for i, tid := range self.tids {
		infos[i] = self.tasks[tid].GetStat()
	}
	return infos
}

func (self *BgTasks) BringBgTaskToFront(tid string, stdout io.Writer) {
	self.lock.Lock()
	defer self.lock.Unlock()
	task, ok := self.tasks[tid]
	if !ok {
		// PANIC: Programming error - task not found
		panic(fmt.Errorf("[BgTasks.BringBgTaskToFront] task '%s' not found", tid))
	}
	task.stdout.BringToFront(stdout)
}

func (self *BgTasks) RemoveTask(tid string) {
	self.lock.Lock()
	defer self.lock.Unlock()
	task, ok := self.tasks[tid]
	if len(self.tids) == 0 || !ok {
		// PANIC: Programming error - task not found or empty task list
		panic(fmt.Errorf("[BgTasks.RemoveTask] task '%s' not found", tid))
	}
	if self.tids[0] != tid {
		// PANIC: Programming error - trying to remove non-earliest task
		panic(fmt.Errorf("[BgTasks.RemoveTask] removing task '%s' is not the earliest", tid))
	}
	self.tids = self.tids[1:]
	delete(self.tasks, tid)
	delete(self.name2task, task.realCmd)
}
