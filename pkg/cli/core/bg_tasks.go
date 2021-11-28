package core

import (
	"bytes"
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
	return self.buffer.Write(p)
}

type BgTask struct {
}

type BgTasks struct {
	tasks map[string]BgTask
}

func NewBgTasks() *BgTasks {
	return &BgTasks{}
}

func (self *BgTasks) GetOrAddTask(id string) *BgTask {
	return nil
}

func (self *BgTasks) TrySwitchToBgTaskScreen() {
}
