package model

import (
	"fmt"
	"strconv"
)

type ArgVals map[string]ArgVal

type ArgVal struct {
	Raw      string
	Provided bool
	Index    int
}

func (self ArgVals) GetRaw(name string) (raw string) {
	val, ok := self[name]
	if !ok {
		return
	}
	return val.Raw
}

func (self ArgVals) GetRawEx(name string, defVal string) (raw string) {
	val, ok := self[name]
	if !ok {
		return defVal
	}
	return val.Raw
}

func (self ArgVals) GetUint64(name string) uint64 {
	val, ok := self[name]
	if !ok {
		panic(&ArgValErrNotFound{
			fmt.Sprintf("[ArgVals.GetUint64] arg '%s' not found", name),
			name,
		})
	}
	if len(val.Raw) == 0 {
		return 0
	}
	intVal, err := strconv.ParseUint(val.Raw, 10, 64)
	if err != nil {
		panic(&ArgValErrWrongType{
			fmt.Sprintf("[ArgVals.GetUint64] arg '%s' = '%s' is not uint64: %v", name, val.Raw, err),
			name, val.Raw, "uint64", err,
		})
	}
	return uint64(intVal)
}

func (self ArgVals) GetUint64Ex(name string, defVal uint64) uint64 {
	_, ok := self[name]
	if !ok {
		return defVal
	}
	return self.GetUint64(name)
}

func (self ArgVals) GetInt(name string) int {
	val, ok := self[name]
	if !ok {
		panic(&ArgValErrNotFound{
			fmt.Sprintf("[ArgVals.GetInt] arg '%s' not found", name),
			name,
		})
	}
	if len(val.Raw) == 0 {
		return 0
	}
	intVal, err := strconv.Atoi(val.Raw)
	if err != nil {
		panic(&ArgValErrWrongType{
			fmt.Sprintf("[ArgVals.GetInt] arg '%s' = '%s' is not int: %v", name, val.Raw, err),
			name, val.Raw, "int", err,
		})
	}
	return int(intVal)
}

func (self ArgVals) GetIntEx(name string, defVal int) int {
	_, ok := self[name]
	if !ok {
		return defVal
	}
	return self.GetInt(name)
}

func (self ArgVals) GetBool(name string) bool {
	val, ok := self[name]
	if !ok {
		panic(&ArgValErrNotFound{
			fmt.Sprintf("[ArgVals.GetBool] arg '%s' not found", name),
			name,
		})
	}
	return StrToBool(val.Raw)
}

func (self ArgVals) GetBoolEx(name string, defVal bool) bool {
	val, ok := self[name]
	if !ok {
		return defVal
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
