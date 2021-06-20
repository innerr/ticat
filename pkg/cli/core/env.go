package core

import (
	"fmt"
	"strings"
)

type EnvLayerType string

const (
	EnvLayerDefault   EnvLayerType = "default"
	EnvLayerPersisted              = "persisted"
	EnvLayerSession                = "session"
	EnvLayerCmd                    = "command"
	EnvLayerTmp                    = "temporary"
)

func EnvLayerName(ty EnvLayerType) string {
	return string(ty)
}

type Env struct {
	pairs  map[string]EnvVal
	parent *Env
	ty     EnvLayerType
}

func NewEnv() *Env {
	return &Env{map[string]EnvVal{}, nil, EnvLayerDefault}
}

func (self *Env) NewLayer(ty EnvLayerType) *Env {
	env := NewEnv()
	env.parent = self
	env.ty = ty
	return env
}

func (self *Env) NewLayers(ty ...EnvLayerType) *Env {
	env := self
	for _, it := range ty {
		env = env.NewLayer(it)
	}
	return env
}

func (self *Env) GetLayer(ty EnvLayerType) *Env {
	env := self.getLayer(ty)
	if env == nil {
		panic(fmt.Errorf("[Env.GetLayer] env layer '%s' not found", ty))
	}
	return env
}

func (self *Env) GetOrNewLayer(ty EnvLayerType) *Env {
	env := self.getLayer(ty)
	if env == nil {
		return self.NewLayer(ty)
	}
	return env
}

func (self *Env) getLayer(ty EnvLayerType) *Env {
	if self.ty == ty {
		return self
	}
	if self.parent == nil {
		return nil
	}
	return self.parent.getLayer(ty)
}

func (self Env) DeleteInSelfLayer(name string) {
	delete(self.pairs, name)
}

func (self Env) Delete(name string) {
	delete(self.pairs, name)
	if self.parent != nil {
		self.parent.Delete(name)
	}
}

func (self Env) DeleteEx(name string, stopLayer EnvLayerType) {
	if self.ty == stopLayer {
		return
	}
	delete(self.pairs, name)
	if self.parent != nil {
		self.parent.DeleteEx(name, stopLayer)
	}
}

func (self *Env) Merge(x *Env) {
	for k, v := range x.pairs {
		self.pairs[k] = EnvVal{v.Raw, false}
	}
}

func (self *Env) Deduplicate() {
	if self.parent == nil {
		return
	}
	for k, v := range self.pairs {
		old, ok := self.parent.GetEx(k)
		if ok && old.Raw == v.Raw {
			delete(self.pairs, k)
		}
	}
}

func (self *Env) Set(name string, val string) (old EnvVal) {
	return self.SetEx(name, val, false)
}

func (self *Env) SetIfEmpty(name string, val string) (old EnvVal) {
	var exists bool
	old, exists = self.GetEx(name)
	if exists {
		return
	}
	self.pairs[name] = EnvVal{val, false}
	return
}

func (self *Env) SetAsArg(name string, val string) (old EnvVal) {
	return self.SetEx(name, val, true)
}

func (self *Env) SetEx(name string, val string, isArg bool) (old EnvVal) {
	var exists bool
	old, exists = self.GetEx(name)
	if exists && old.Raw == val {
		return
	}
	self.pairs[name] = EnvVal{val, isArg}
	return
}

func (self *Env) Parent() *Env {
	return self.parent
}

func (self *Env) GetArgv(path []string, sep string, args Args) ArgVals {
	argv := ArgVals{}
	list := args.Names()
	for _, it := range list {
		key := strings.Join(append(path, it), sep)
		val := self.Get(key)
		if len(val.Raw) != 0 {
			argv[it] = ArgVal{val.Raw}
		} else {
			argv[it] = ArgVal{args.DefVal(it)}
		}
	}
	if len(argv) == 0 {
		return nil
	}
	return argv
}

func (self Env) Get(name string) EnvVal {
	val, ok := self.pairs[name]
	if !ok && self.parent != nil {
		return self.parent.Get(name)
	}
	return val
}

func (self Env) GetEx(name string) (EnvVal, bool) {
	val, ok := self.pairs[name]
	if !ok && self.parent != nil {
		return self.parent.GetEx(name)
	}
	return val, ok
}

func (self Env) Pairs() (keys []string, vals []EnvVal) {
	for k, v := range self.pairs {
		keys = append(keys, k)
		vals = append(vals, v)
	}
	return
}

func (self Env) LayerType() EnvLayerType {
	return self.ty
}

func (self Env) LayerTypeName() string {
	return EnvLayerName(self.ty)
}

func (self Env) Flatten(
	includeDefault bool,
	filterPrefixs []string,
	filterArgs bool) map[string]string {

	res := map[string]string{}
	self.flatten(includeDefault, filterPrefixs, res, filterArgs)
	return res
}

func (self *Env) flatten(
	includeDefault bool,
	filterPrefixs []string,
	res map[string]string,
	filterArgs bool) {

	if self.ty == EnvLayerDefault && !includeDefault {
		return
	}
	if self.parent != nil {
		self.parent.flatten(includeDefault, filterPrefixs, res, filterArgs)
	}
	for k, v := range self.pairs {
		filtered := false
		for _, filterPrefix := range filterPrefixs {
			if len(filterPrefix) != 0 && strings.HasPrefix(k, filterPrefix) {
				filtered = true
				break
			}
		}
		if !filtered {
			if !filterArgs || !v.IsArg {
				res[k] = v.Raw
			}
		}
	}
}
