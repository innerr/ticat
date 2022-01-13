package core

import (
	"fmt"
	"strings"
)

type ForestMode struct {
	stack []string
}

func (self ForestMode) AtForestTopLvl(env *Env) bool {
	return len(self.stack) != 0 && self.stack[len(self.stack)-1] == GetLastStackFrame(env)
}

func (self *ForestMode) Pop(env *Env) {
	if len(self.stack) == 0 {
		panic(fmt.Errorf("[BlenderForest] should never happen: pop on empty"))
	}
	last1 := GetLastStackFrame(env)
	last2 := self.stack[len(self.stack)-1]
	if last1 != last2 {
		panic(fmt.Errorf("[BlenderForest] should never happen: pop on wrong frame: %s != %s", last1, last2))
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

type Blender struct {
	ForestMode *ForestMode
}

func NewBlender() *Blender {
	return &Blender{&ForestMode{[]string{}}}
}

func (self *Blender) Clone() *Blender {
	return &Blender{self.ForestMode.Clone()}
}

// TODO: should not in core package
func GetLastStackFrame(env *Env) string {
	stackStr := env.GetRaw("sys.stack")
	if len(stackStr) == 0 {
		panic(fmt.Errorf("[BlenderForestMode] should never happen"))
	}
	listSep := env.GetRaw("strs.list-sep")
	stack := strings.Split(stackStr, listSep)
	return stack[len(stack)-1]
}
