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
		panic(err)
	}
	return int(val)
}

func (self EnvVal) GetBool() bool {
	return self.Raw == "true" || self.Raw == "True" || self.Raw == "TRUE" || self.Raw == "1" ||
		self.Raw == "on" || self.Raw == "On" || self.Raw == "ON" || self.Raw == "y" || self.Raw == "Y"
}
