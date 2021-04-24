package cli

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
)

func EnvLayerName(tp EnvLayerType) string {
	return string(tp)
}

type EnvVal struct {
	Raw      string
	IsArg    bool
	ValCache interface{}
}

type Env struct {
	pairs  map[string]EnvVal
	parent *Env
	tp     EnvLayerType
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
		panic(fmt.Errorf("[Env.GetLayer] env layer '%s' not found", tp))
	}
	return self.parent.GetLayer(tp)
}

func (self Env) DeleteSelf(name string) {
	delete(self.pairs, name)
}

func (self Env) Delete(name string) {
	delete(self.pairs, name)
	if self.parent != nil {
		self.parent.Delete(name)
	}
}

func (self Env) DeleteExt(name string, stopLayer EnvLayerType) {
	if self.tp == stopLayer {
		return
	}
	delete(self.pairs, name)
	if self.parent != nil {
		self.parent.DeleteExt(name, stopLayer)
	}
}

func (self Env) Get(name string) EnvVal {
	val, ok := self.pairs[name]
	if !ok && self.parent != nil {
		return self.parent.Get(name)
	}
	return val
}

func (self Env) GetExt(name string) (EnvVal, bool) {
	val, ok := self.pairs[name]
	if !ok && self.parent != nil {
		return self.parent.GetExt(name)
	}
	return val, ok
}

func (self *Env) Merge(x *Env) {
	for k, v := range x.pairs {
		self.pairs[k] = EnvVal{v.Raw, false, nil}
	}
}

func (self *Env) Deduplicate() {
	if self.parent == nil {
		return
	}
	for k, _ := range self.pairs {
		_, ok := self.parent.GetExt(k)
		if ok {
			delete(self.pairs, k)
		}
	}
}

func (self *Env) Set(name string, val string) (old EnvVal) {
	return self.SetExt(name, val, false)
}

func (self *Env) SetAsArg(name string, val string) (old EnvVal) {
	return self.SetExt(name, val, true)
}

func (self *Env) SetExt(name string, val string, isArg bool) (old EnvVal) {
	var exists bool
	old, exists = self.GetExt(name)
	if exists && old.Raw == val {
		return
	}
	self.pairs[name] = EnvVal{val, isArg, nil}
	return
}

func (self *Env) Parent() *Env {
	return self.parent
}

func (self *Env) GetArgv(path []string, sep string, args *Args) ArgVals {
	if args == nil {
		return nil
	}
	argv := ArgVals{}
	list := args.List()
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

func (self Env) Compact(includeDefault bool, filterPrefixs []string) map[string]string {
	res := map[string]string{}
	self.compact(includeDefault, filterPrefixs, res)
	return res
}

func (self *Env) compact(includeDefault bool, filterPrefixs []string, res map[string]string) {
	if self.tp == EnvLayerDefault && !includeDefault {
		return
	}
	if self.parent != nil {
		self.parent.compact(includeDefault, filterPrefixs, res)
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
			res[k] = v.Raw
		}
	}
}
