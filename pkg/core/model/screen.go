package model

import (
	"io"
	"sync"
)

type QuietScreen struct {
	outN int
}

func (self *QuietScreen) Print(text string) error {
	self.outN += 1
	return nil
}

func (self *QuietScreen) Error(text string) error {
	return nil
}

func (self *QuietScreen) OutputtedLines() int {
	return self.outN
}

type StdScreen struct {
	stdout io.Writer
	stderr io.Writer
	outN   int
}

func NewStdScreen(stdout io.Writer, stderr io.Writer) *StdScreen {
	return &StdScreen{
		stdout: stdout,
		stderr: stderr,
	}
}

func (self *StdScreen) Print(text string) error {
	self.outN += 1
	if self.stdout == nil {
		return nil
	}
	_, err := io.WriteString(self.stdout, text)
	return err
}

func (self *StdScreen) Error(text string) error {
	if self.stderr == nil {
		return nil
	}
	_, err := io.WriteString(self.stderr, text)
	return err
}

func (self *StdScreen) OutputtedLines() int {
	return self.outN
}

type BgTaskScreen struct {
	basic  *StdScreen
	stdout *BgStdout
	lock   sync.Mutex
}

func NewBgTaskScreen() *BgTaskScreen {
	stdout := NewBgStdout()
	return &BgTaskScreen{
		basic:  NewStdScreen(stdout, stdout),
		stdout: stdout,
	}
}

func (self *BgTaskScreen) Print(text string) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	return self.basic.Print(text)
}

func (self *BgTaskScreen) Error(text string) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	return self.basic.Error(text)
}

func (self *BgTaskScreen) OutputtedLines() int {
	self.lock.Lock()
	defer self.lock.Unlock()
	return self.basic.OutputtedLines()
}

func (self *BgTaskScreen) GetBgStdout() *BgStdout {
	self.lock.Lock()
	defer self.lock.Unlock()
	return self.stdout
}
