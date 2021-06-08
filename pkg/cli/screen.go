package cli

import (
	"fmt"
	"os"

	"github.com/pingcap/ticat/pkg/cli/core"
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

func (self *Screen) Error(text string) {
	fmt.Fprint(os.Stderr, text)
}

func (self *Screen) OutputNum() int {
	return self.outN
}

type CachedOutput struct {
	Text string
	IsError bool
}

type CacheScreen struct {
	data []CachedOutput
	outN int
}

func NewCacheScreen() *CacheScreen {
	return &CacheScreen{nil, 0}
}

func (self *CacheScreen) Print(text string) {
	self.data = append(self.data, CachedOutput{text, false})
	self.outN += 1
}

func (self *CacheScreen) Error(text string) {
	self.data = append(self.data, CachedOutput{text, true})
}

func (self *CacheScreen) OutputNum() int {
	return self.outN
}

func (self *CacheScreen) WriteToEx(
	screen core.Screen,
	transformer func(text string, isError bool)(string, bool)) {

	for _, it := range self.data {
		text, isError := transformer(it.Text, it.IsError)
		if isError {
			screen.Error(text)
		} else {
			screen.Print(text)
		}
	}
}

func (self *CacheScreen) WriteTo(screen core.Screen) {
	self.WriteToEx(screen, func(text string, isError bool)(string, bool) {
		return text, isError
	})
}
