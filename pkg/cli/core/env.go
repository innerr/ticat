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
	EnvLayerSubFlow                = "subflow"
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

func NewEnvEx(ty EnvLayerType) *Env {
	return &Env{map[string]EnvVal{}, nil, ty}
}

// TODO: COW ?
func (self *Env) Clone() (env *Env) {
	pairs := map[string]EnvVal{}
	for k, v := range self.pairs {
		pairs[k] = EnvVal{v.Raw, v.IsArg, v.IsSysArg}
	}
	var parent *Env
	if self.parent != nil {
		parent = self.parent.Clone()
	}
	return &Env{pairs, parent, self.ty}
}

func (self *Env) Clear(recursive bool) {
	pairs := map[string]EnvVal{}
	for k, v := range self.pairs {
		// TODO: put all these special key path in one place
		if strings.HasPrefix(k, "sys.") || k == "session" {
			pairs[k] = EnvVal{v.Raw, v.IsArg, v.IsSysArg}
		}
	}
	self.pairs = pairs
	if recursive && self.parent != nil {
		self.parent.Clear(recursive)
	}
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

func (self *Env) GetOneOfLayers(tys ...EnvLayerType) (env *Env) {
	names := []string{}
	for _, ty := range tys {
		env = self.getLayer(ty)
		if env != nil {
			break
		}
		names = append(names, string(ty))
	}
	if env == nil {
		panic(fmt.Errorf("[Env.GetOneOfLayers] env layers '%s' not found", strings.Join(names, " ")))
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

func (self Env) Has(name string) bool {
	_, ok := self.pairs[name]
	if ok {
		return true
	}
	if self.parent == nil {
		return false
	}
	return self.parent.Has(name)
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
	// TODO: why we discard arg flags from x?
	for k, v := range x.pairs {
		self.pairs[k] = EnvVal{v.Raw, false, false}
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
	return self.SetEx(name, val, false, false)
}

func (self *Env) SetIfEmpty(name string, val string) (old EnvVal) {
	var exists bool
	old, exists = self.GetEx(name)
	if exists {
		return
	}
	self.pairs[name] = EnvVal{val, false, false}
	return
}

//func (self *Env) SetAsArg(name string, val string, isSysArg bool) (old EnvVal) {
//	return self.SetEx(name, val, true, isSysArg)
//}

func (self *Env) SetEx(name string, val string, isArg bool, isSysArg bool) (old EnvVal) {
	var exists bool
	old, exists = self.GetEx(name)
	if exists && old.Raw == val {
		return
	}
	self.pairs[name] = EnvVal{val, isArg, isSysArg}
	return
}

func (self *Env) Parent() *Env {
	return self.parent
}

func (self *Env) GetArgv(cmdPath []string, sep string, args Args) ArgVals {
	argv := ArgVals{}
	list := args.Names()
	for i, it := range list {
		key := strings.Join(append(cmdPath, it), sep)
		val, ok := self.GetEx(key)
		if ok {
			argv[it] = ArgVal{val.Raw, true, i}
		} else {
			argv[it] = ArgVal{args.DefVal(it), false, i}
		}
	}
	return argv
}

func (self *Env) GetSysArgv(cmdPath []string, sep string) SysArgVals {
	sysArgv := SysArgVals{}
	keyPrefix := strings.Join(cmdPath, sep) + sep
	sysArgPrefix := self.GetRaw("strs.sys-arg-prefix")
	self.getSysArgv(keyPrefix, sysArgPrefix, sysArgv)
	return sysArgv
}

func (self *Env) getSysArgv(keyPrefix string, sysArgPrefix string, sysArgv SysArgVals) {
	prefixLen := len(keyPrefix) + len(sysArgPrefix)
	for k, v := range self.pairs {
		if !v.IsSysArg || len(k) <= prefixLen || !strings.HasPrefix(k, keyPrefix) {
			continue
		}
		key := k[prefixLen:]
		if _, ok := sysArgv[key]; !ok {
			sysArgv[key] = v.Raw
		}
	}
	if self.parent != nil {
		self.parent.getSysArgv(keyPrefix, sysArgPrefix, sysArgv)
	}
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

func (self Env) FlattenAll() map[string]string {
	return self.Flatten(true, nil, false)
}

func (self Env) Flatten(
	includeDefault bool,
	filterPrefixs []string,
	filterArgs bool) map[string]string {

	res := map[string]string{}
	self.flatten(includeDefault, filterPrefixs, res, filterArgs)
	return res
}

func (self Env) WriteCurrLayerTo(env *Env) {
	for k, v := range self.pairs {
		env.Set(k, v.Raw)
	}
}

func (self *Env) CleanCurrLayer() {
	self.pairs = map[string]EnvVal{}
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

func IsSensitiveKeyVal(key string, val string) bool {
	sensitives := []string{
		"pwd",
		"password",
	}
	key = strings.ToLower(key)
	for _, it := range sensitives {
		if strings.Index(key, it) >= 0 && !StrToTrue(val) && !StrToFalse(val) {
			return true
		}
	}
	return false
}
