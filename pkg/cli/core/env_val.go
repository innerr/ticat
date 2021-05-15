package core

import (
	"fmt"
	"strconv"
)

func (self Env) GetRaw(name string) string {
	return self.Get(name).Raw
}

func (self Env) SetInt(name string, val int) {
	self.Set(name, fmt.Sprintf("%d", val))
}

func (self Env) GetInt(name string) int {
	val, err := strconv.Atoi(self.Get(name).Raw)
	if err != nil {
		panic(fmt.Errorf("[EnvVal.GetInt] strconv failed: %v", err))
	}
	return int(val)
}

func (self Env) PlusInt(name string, val int) {
	self.SetInt(name, self.GetInt(name)+val)
}

func (self Env) SetBool(name string, val bool) bool {
	old := StrToBool(self.Get(name).Raw)
	self.Set(name, fmt.Sprintf("%v", val))
	return old
}

func (self Env) GetBool(name string) bool {
	return StrToBool(self.Get(name).Raw)
}
