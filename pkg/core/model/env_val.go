package model

import (
	"fmt"
	"strconv"
	"time"

	"github.com/innerr/ticat/pkg/utils"
)

func (self Env) GetRaw(name string) string {
	return self.Get(name).Raw
}

func (self Env) SetInt(name string, val int) {
	self.Set(name, fmt.Sprintf("%d", val))
}

func (self Env) SetUint64(name string, val uint64) {
	self.Set(name, fmt.Sprintf("%v", val))
}

func (self Env) GetInt(name string) int {
	val := self.Get(name).Raw
	if len(val) == 0 {
		return 0
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		panic(&EnvValErrWrongType{
			fmt.Sprintf("[EnvVal.GetInt] key '%s' = '%s' is not int: %v", name, val, err),
			name, val, "int", err,
		})
	}
	return int(intVal)
}

func (self Env) GetUint64(name string) uint64 {
	val := self.Get(name).Raw
	if len(val) == 0 {
		return 0
	}
	intVal, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		panic(&EnvValErrWrongType{
			fmt.Sprintf("[EnvVal.GetUint64] key '%s' = '%s' is not uint64: %v", name, val, err),
			name, val, "uint64", err,
		})
	}
	return uint64(intVal)
}

func (self Env) GetDur(name string) time.Duration {
	_, ok := self.GetEx(name)
	if !ok {
		panic(&EnvValErrNotFound{
			fmt.Sprintf("[EnvVal.GetDur] key '%s' not found in env", name),
			name,
		})
	}
	val := utils.NormalizeDurStr(self.Get(name).Raw)
	dur, err := time.ParseDuration(val)
	if err != nil {
		panic(&EnvValErrWrongType{
			fmt.Sprintf("[EnvVal.GetDur] key '%s' = '%s' is not duration format: %v", name, val, err),
			name, val, "Golang: time.Duration", err,
		})
	}
	return dur
}

func (self Env) SetDur(name string, val string) {
	val = utils.NormalizeDurStr(val)
	_, err := time.ParseDuration(val)
	if err != nil {
		panic(&EnvValErrWrongType{
			fmt.Sprintf("[EnvVal.SetDur] key '%s' = '%s' is not duration format: %v", name, val, err),
			name, val, "Golang: time.Duration", err,
		})
	}
	self.Set(name, val)
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
	Raw      string
	IsArg    bool
	IsSysArg bool
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

type EnvValErrNotFound struct {
	Err string
	Key string
}

func (self EnvValErrNotFound) Error() string {
	return self.Err
}
