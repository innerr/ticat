package cli

import (
	"fmt"
)

type Screen struct {
}

func NewScreen() *Screen {
	return &Screen{}
}

func (self *Screen) Println(text string) {
	fmt.Println(text)
}

func (self *Screen) Print(text string) {
	fmt.Print(text)
}
