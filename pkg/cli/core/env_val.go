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
	val := self.Get(name).Raw
	intVal, err := strconv.Atoi(val)
	if err != nil {
		panic(EnvValErrWrongType{
			fmt.Sprintf("[EnvVal.GetInt] key '%s' = '%s' is not int: %v", name, val, err),
			name, val, "int", err,
		})
	}
	return int(intVal)
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

type EnvVal struct {
	Raw   string
	IsArg bool
}

type EnvValErrWrongType struct {
	Str        string
	Key        string
	Val        string
	ExpectType string
	ConvertErr error
}

func (self EnvValErrWrongType) Error() string {
	return self.Str
}
