package model

import (
	"fmt"
)

type ForestMode struct {
	stack []string
}

func NewForestMode() *ForestMode {
	return &ForestMode{[]string{}}
}

func (self ForestMode) AtForestTopLvl(env *Env) bool {
	return len(self.stack) != 0 && self.stack[len(self.stack)-1] == GetLastStackFrame(env)
}

func (self *ForestMode) Pop(env *Env) {
	if len(self.stack) == 0 {
		panic(fmt.Errorf("[Forest] should never happen: pop on empty"))
	}
	last1 := GetLastStackFrame(env)
	last2 := self.stack[len(self.stack)-1]
	if last1 != last2 {
		panic(fmt.Errorf("[Forest] should never happen: pop on wrong frame: %s != %s", last1, last2))
	}
	self.stack = self.stack[0 : len(self.stack)-1]
}

func (self *ForestMode) Push(frame string) {
	self.stack = append(self.stack, frame)
}

func (self *ForestMode) Clone() *ForestMode {
	stack := []string{}
	for _, it := range self.stack {
		stack = append(stack, it)
	}
	return &ForestMode{stack}
}
