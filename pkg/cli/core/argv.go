package core

import (
	"fmt"
	"strconv"
)

type ArgVals map[string]ArgVal

type ArgVal struct {
	Raw      string
	Provided bool
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
		panic(ArgValErrNotFound{
			fmt.Sprintf("[ArgVals.GetInt] arg '%s' not found", name),
			name,
		})
	}
	intVal, err := strconv.Atoi(val.Raw)
	if err != nil {
		panic(ArgValErrWrongType{
			fmt.Sprintf("[ArgVals.GetInt] arg '%s' = '%s' is not int: %v", name, val.Raw, err),
			name, val.Raw, "int", err,
		})
	}
	return int(intVal)
}

func (self ArgVals) GetBool(name string) bool {
	val, ok := self[name]
	if !ok {
		panic(ArgValErrNotFound{
			fmt.Sprintf("[ArgVals.GetBool] arg '%s' not found", name),
			name,
		})
	}
	return StrToBool(val.Raw)
}

type ArgValErrNotFound struct {
	Str     string
	ArgName string
}

func (self ArgValErrNotFound) Error() string {
	return self.Str
}

type ArgValErrWrongType struct {
	Str        string
	ArgName    string
	Val        string
	ExpectType string
	ConvertErr error
}

func (self ArgValErrWrongType) Error() string {
	return self.Str
}
