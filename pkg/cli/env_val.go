package cli

import (
	"fmt"
	"strconv"
)

type EnvVal struct {
	Raw      string
	IsArg    bool
	ValCache interface{}
}

func (self *EnvVal) SetInt(val int) {
	self.Raw = fmt.Sprintf("%d", val)
}

func (self EnvVal) GetInt() int {
	val, err := strconv.ParseInt(self.Raw, 10, 64)
	if err != nil {
		panic(fmt.Errorf("[EnvVal.GetInt] strconv failed: %v", err))
	}
	return int(val)
}

func (self EnvVal) GetBool() bool {
	return StrToBool(self.Raw)
}
