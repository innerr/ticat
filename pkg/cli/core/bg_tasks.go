package core

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

type CmdIO struct {
	CmdStdin  io.Reader
	CmdStdout io.Writer
	CmdStderr io.Writer
}

func (self CmdIO) SetupCmd(cmd *exec.Cmd) {
	if self.CmdStdin != nil {
		cmd.Stdin = self.CmdStdin
	}
	if self.CmdStdout != nil {
		cmd.Stdout = self.CmdStdout
	}
	if self.CmdStderr != nil {
		cmd.Stderr = self.CmdStderr
	}
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
		panic(err)
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
}

type BgTask struct {
	info           BgTaskInfo
	stdout         *BgStdout
	finishNotifier chan interface{}
	lock           sync.Mutex
}

func NewBgTask(tid string, cmd string, stdout *BgStdout) *BgTask {
	return &BgTask{
		info: BgTaskInfo{
			Tid: tid,
			Cmd: cmd,
		},
		stdout:         stdout,
		finishNotifier: make(chan interface{}),
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

func (self *BgTask) OnFinish() {
	self.lock.Lock()
	self.info.Finished = true
	self.lock.Unlock()
	self.finishNotifier <- nil
}

func (self *BgTask) WaitForFinish() {
	<-self.finishNotifier
}

type BgTasks struct {
	tids  []string
	tasks map[string]*BgTask
	lock  sync.Mutex
}

func NewBgTasks() *BgTasks {
	return &BgTasks{
		tids:  []string{},
		tasks: map[string]*BgTask{},
	}
}

func (self *BgTasks) GetOrAddTask(tid string, cmd string, stdout *BgStdout) *BgTask {
	self.lock.Lock()
	defer self.lock.Unlock()
	task, ok := self.tasks[tid]
	if ok {
		return task
	}
	self.tids = append(self.tids, tid)
	task = NewBgTask(tid, cmd, stdout)
	self.tasks[tid] = task
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
		panic(fmt.Errorf("[BgTasks.BringBgTaskToFront] task '%s' not found", tid))
	}
	task.stdout.BringToFront(stdout)
}

func (self *BgTasks) RemoveTask(tid string) {
	self.lock.Lock()
	defer self.lock.Unlock()
	_, ok := self.tasks[tid]
	if len(self.tids) == 0 || !ok {
		panic(fmt.Errorf("[BgTasks.RemoveTask] task '%s' not found", tid))
	}
	if self.tids[0] != tid {
		panic(fmt.Errorf("[BgTasks.RemoveTask] removing task '%s' is not the earliest", tid))
	}
	self.tids = self.tids[1:]
	delete(self.tasks, tid)
}
