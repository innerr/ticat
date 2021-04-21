package cli

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
		panic(fmt.Errorf("arg '%s' not found", name))
	}
	intVal, err := strconv.ParseInt(val.Raw, 10, 64)
	if err != nil {
		panic(err)
	}
	return int(intVal)
}
