package cli

import (
	"fmt"
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

type EnvVal struct {
	Raw string
	Val interface{}
}

type Env struct {
	pairs  map[string]*EnvVal
	parent *Env
	Type   EnvLayerType
}

func NewEnv() *Env {
	return &Env{map[string]*EnvVal{}, nil, EnvLayerDefault}
}

func (self *Env) NewLayer(tp EnvLayerType) *Env {
	env := NewEnv()
	env.parent = self
	env.Type = tp
	return env
}

func (self *Env) NewLayerIfTypeNotMatch(tp EnvLayerType) *Env {
	if self.Type != tp {
		return self.NewLayer(tp)
	}
	return self
}

func (self *Env) Get(name string) *EnvVal {
	val, ok := self.pairs[name]
	if !ok && self.parent != nil {
		return self.parent.Get(name)
	}
	return val
}

func (self *Env) Set(name string, val *EnvVal) *EnvVal {
	old, _ := self.pairs[name]
	self.pairs[name] = val
	return old
}

func (self *Env) ParseAndSet(strs []string) {
}
