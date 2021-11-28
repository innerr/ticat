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
	buffer *bytes.Buffer
	lock   sync.Mutex
}

func NewBgStdout() *BgStdout {
	return &BgStdout{
		buffer: bytes.NewBuffer(nil),
	}
}

func (self *BgStdout) Write(p []byte) (n int, err error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	return self.buffer.Write(p)
}

type BgTask struct {
	tid            string
	finishNotifier chan interface{}
}

func (self *BgTask) OnFinish() {
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

func (self *BgTasks) GetOrAddTask(tid string) *BgTask {
	self.lock.Lock()
	defer self.lock.Unlock()
	task, ok := self.tasks[tid]
	if ok {
		return task
	}
	self.tids = append(self.tids, tid)
	task = &BgTask{tid, make(chan interface{})}
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

func (self *BgTasks) BrightBgTaskToFront(cmdIO CmdIO) {
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
