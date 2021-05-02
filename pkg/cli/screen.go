package cli

import (
	"fmt"
)

type Screen struct {
}

func NewScreen() *Screen {
	return &Screen{}
}

func (self *Screen) Print(text string) {
	fmt.Print(text)
}
