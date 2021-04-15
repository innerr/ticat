package cli

import (
	"fmt"
)

type Screen struct {
}

func (self *Screen) PrintSeperatingHeader(text string) {
	fmt.Println("================")
	fmt.Printf("=> %s\n", text)
	fmt.Println("================")
}

func (self *Screen) Print(text string) {
	fmt.Println(text)
}
