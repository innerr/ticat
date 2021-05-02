package core

import (
	"fmt"
	"strconv"
)

type ArgVals map[string]ArgVal

type ArgVal struct {
	Raw string
}

func (self ArgVals) GetRaw(name string) (raw string) {
	val, ok := self[name]
	if !ok {
		return
	}
	return val.Raw
}

func (self ArgVals) GetInt(name string) int {
	val, ok := self[name]
	if !ok {
		panic(fmt.Errorf("[ArgVals.GetInt] arg '%s' not found", name))
	}
	intVal, err := strconv.ParseInt(val.Raw, 10, 64)
	if err != nil {
		panic(fmt.Errorf("[ArgVals.GetInt] arg '%s' = '%s' is not int: %v", name, val.Raw, err))
	}
	return int(intVal)
}

func (self ArgVals) GetBool(name string) bool {
	val, ok := self[name]
	if !ok {
		panic(fmt.Errorf("[ArgVals.GetBool] arg '%s' not found", name))
	}
	return StrToBool(val.Raw)
}
