package cli

import (
	"fmt"
	"strings"
)

type EnvLayerType string

const (
	EnvLayerDefault EnvLayerType = "default"
	EnvLayerPersisted = "persisted"
	EnvLayerSession = "session"
	EnvLayerCmd = "command"
)

func EnvLayerName(tp EnvLayerType) string {
	return string(tp)
}

type Env struct {
	pairs  map[string]EnvVal
	parent *Env
	tp   EnvLayerType
}

func NewEnv() *Env {
	return &Env{map[string]EnvVal{}, nil, EnvLayerDefault}
}

func (self *Env) NewLayer(tp EnvLayerType) *Env {
	env := NewEnv()
	env.parent = self
	env.tp = tp
	return env
}

func (self *Env) NewLayers(tp ...EnvLayerType) *Env {
	env := self
	for _, it := range tp {
		env = env.NewLayer(it)
	}
	return env
}

func (self *Env) GetLayer(tp EnvLayerType) *Env {
	if self.tp == tp {
		return self
	}
	if self.parent == nil {
		panic(fmt.Errorf("env layer '%s' not found", tp))
	}
	return self.parent.GetLayer(tp)
}

func (self Env) Get(name string) EnvVal {
	val, ok := self.pairs[name]
	if !ok && self.parent != nil {
		return self.parent.Get(name)
	}
	return val
}

func (self *Env) Set(name string, val string) (old EnvVal) {
	old = self.Get(name)
	if old.Raw == val {
		return
	}
	self.pairs[name] = EnvVal{val, nil}
	return
}

func (self *Env) Parent() *Env {
	return self.parent
}

func (self Env) Compact(includeDefault bool, filterPrefix string) map[string]string {
	res := map[string]string{}
	self.compact(includeDefault, filterPrefix, res)
	return res
}

func (self *Env) compact(includeDefault bool, filterPrefix string, res map[string]string) {
	if self.tp == EnvLayerDefault && !includeDefault {
		return
	}
	if self.parent != nil {
		self.parent.compact(includeDefault, filterPrefix, res)
	}
	for k, v := range self.pairs {
		if len(filterPrefix) != 0 && strings.HasPrefix(k, filterPrefix) {
			continue
		}
		res[k] = v.Raw
	}
}
