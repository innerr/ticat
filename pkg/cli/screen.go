package cli

import (
	"fmt"
)

// TODO: stdin/stderr

type Screen struct {
	outN int
}

func NewScreen() *Screen {
	return &Screen{}
}

func (self *Screen) Print(text string) {
	fmt.Print(text)
	self.outN += 1
}

func (self *Screen) OutputNum() int {
	return self.outN
}
