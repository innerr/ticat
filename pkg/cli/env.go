package cli

import (
	"fmt"
	"strconv"
	"strings"
)

type EnvLayerType uint

const (
	EnvLayerDefault EnvLayerType = iota
	EnvLayerPersisted
	EnvLayerSession
	EnvLayerMod
)

func EnvLayerName(tp EnvLayerType) string {
	switch tp {
	case EnvLayerDefault:
		return "default"
	case EnvLayerPersisted:
		return "persisted"
	case EnvLayerSession:
		return "session"
	case EnvLayerMod:
		return "module"
	default:
		panic(fmt.Errorf("unknown layer type, value: %v", tp))
	}
}

type Env struct {
	Pairs  map[string]EnvVal
	Parent *Env
	Type   EnvLayerType
}

func NewEnv() *Env {
	return &Env{map[string]EnvVal{}, nil, EnvLayerDefault}
}

func (self *Env) NewLayer(tp EnvLayerType) *Env {
	env := NewEnv()
	env.Parent = self
	env.Type = tp
	return env
}

func (self *Env) GetLayer(tp EnvLayerType) *Env {
	if self.Type == tp {
		return self
	}
	if self.Parent == nil {
		return nil
	}
	return self.Parent.GetLayer(tp)
}

func (self *Env) NewLayerIfTypeNotMatch(tp EnvLayerType) *Env {
	if self.Type != tp {
		return self.NewLayer(tp)
	}
	return self
}

func (self Env) Get(name string) EnvVal {
	val, ok := self.Pairs[name]
	if !ok && self.Parent != nil {
		return self.Parent.Get(name)
	}
	return val
}

func (self *Env) Set(name string, val string) EnvVal {
	old, _ := self.Pairs[name]
	self.Pairs[name] = EnvVal{val, nil}
	return old
}

func (self Env) Compact(includeDefault bool, filterPrefix string) map[string]string {
	res := map[string]string{}
	self.compact(includeDefault, filterPrefix, res)
	return res
}

func (self *Env) compact(includeDefault bool, filterPrefix string, res map[string]string) {
	if self.Type == EnvLayerDefault && !includeDefault {
		return
	}
	if self.Parent != nil {
		self.Parent.compact(includeDefault, filterPrefix, res)
	}
	for k, v := range self.Pairs {
		if len(filterPrefix) != 0 && strings.HasPrefix(k, filterPrefix) {
			continue
		}
		res[k] = v.Raw
	}
}

type Word struct {
	Val   string
	Abbrs []string
}

func NewWord(val string, abbrs ...string) *Word {
	return &Word{val, abbrs}
}

type EnvVal struct {
	Raw      string
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
		self.Raw == "on" || self.Raw == "On" || self.Raw == "ON"
}
