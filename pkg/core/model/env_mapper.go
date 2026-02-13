package model

import (
	"fmt"
	"strings"
)

type Val2Env struct {
	orderedKeys []string
	pairs       map[string]string
}

func newVal2Env() *Val2Env {
	return &Val2Env{nil, map[string]string{}}
}

func (self *Val2Env) IsEmpty() bool {
	return len(self.orderedKeys) == 0
}

func (self *Val2Env) Add(envKey string, val string) {
	_, ok := self.pairs[envKey]
	if ok {
		panic(fmt.Errorf("[Val2Env.Add] duplicated key: %s", envKey))
	}
	self.orderedKeys = append(self.orderedKeys, envKey)
	self.pairs[envKey] = val
}

func (self *Val2Env) EnvKeys() []string {
	return self.orderedKeys
}

func (self *Val2Env) RenderedEnvKeys(
	argv ArgVals,
	env *Env,
	cmd *Cmd,
	allowError bool) (renderedKeys []string, origins []string, fullyRendered bool) {

	fullyRendered = true
	for _, name := range self.orderedKeys {
		keys, keyFullyRendered := renderTemplateStr(name, "map arg to env", cmd, argv, env, allowError)
		for _, key := range keys {
			renderedKeys = append(renderedKeys, key)
			origins = append(origins, name)
		}
		fullyRendered = fullyRendered && keyFullyRendered
	}
	return
}

func (self *Val2Env) Val(envKey string) string {
	return self.pairs[envKey]
}

func (self *Val2Env) RenderedVal(
	envKey string,
	argv ArgVals,
	env *Env,
	cmd *Cmd,
	allowError bool) string {

	val := self.pairs[envKey]
	lines, _ := renderTemplateStr(val, "render val in val2env", cmd, argv, env, allowError)
	return strings.Join(lines, "\n")
}

func (self *Val2Env) Has(envKey string) bool {
	_, ok := self.pairs[envKey]
	return ok
}

func (self *Val2Env) MatchFind(findStr string) bool {
	if strings.Index("write", findStr) >= 0 {
		return true
	}
	for _, envKey := range self.orderedKeys {
		if strings.Index(envKey, findStr) >= 0 {
			return true
		}
	}
	return false
}

func (self *Val2Env) Clone() *Val2Env {
	cloned := newVal2Env()
	cloned.orderedKeys = append([]string{}, self.orderedKeys...)
	for k, v := range self.pairs {
		cloned.pairs[k] = v
	}
	return cloned
}

type Arg2Env struct {
	orderedKeys []string
	keyNames    map[string]string
	nameKeys    map[string]string
}

func newArg2Env() *Arg2Env {
	return &Arg2Env{nil, map[string]string{}, map[string]string{}}
}

func (self *Arg2Env) Add(envKey string, argName string) {
	old, ok := self.keyNames[envKey]
	if ok {
		panic(fmt.Errorf("[Arg2Env.Add] multi args map to env '%s', old '%s', new '%s'",
			envKey, old, argName))
	}
	old, ok = self.nameKeys[argName]
	if ok {
		panic(fmt.Errorf("[Arg2Env.Add] duplicated arg name: '%s' (map to env: old '%s', new '%s')",
			argName, old, envKey))
	}
	self.orderedKeys = append(self.orderedKeys, envKey)
	self.keyNames[envKey] = argName
	self.nameKeys[argName] = envKey
}

func (self *Arg2Env) IsEmpty() bool {
	return len(self.orderedKeys) == 0
}

func (self *Arg2Env) EnvKeys() []string {
	return self.orderedKeys
}

// TODO: put this in use
func (self *Arg2Env) RenderedEnvKeys(
	argv ArgVals,
	env *Env,
	cmd *Cmd,
	allowError bool) (renderedKeys []string, origins []string, fullyRendered bool) {

	fullyRendered = true
	for _, name := range self.orderedKeys {
		keys, keyFullyRendered := renderTemplateStr(name, "map arg to env", cmd, argv, env, allowError)
		for _, key := range keys {
			renderedKeys = append(renderedKeys, key)
			origins = append(origins, name)
		}
		fullyRendered = fullyRendered && keyFullyRendered
	}
	return
}

func (self *Arg2Env) GetArgName(cmd *Cmd, envKey string, allowError bool) string {
	argName, ok := self.keyNames[envKey]
	if !ok {
		return ""
	}
	args := cmd.Args()
	if !args.Has(argName) {
		msg := fmt.Sprintf("module '%s' error: no arg %s for key %s", cmd.Owner().DisplayPath(), argName, envKey)
		if allowError {
			return "(" + msg + ")"
		} else {
			panic(fmt.Errorf("[Arg2Env.GetArgName] %s", msg))
		}
	}
	return argName
}

func (self *Arg2Env) GetEnvKey(argName string) (string, bool) {
	envKey, ok := self.nameKeys[argName]
	return envKey, ok
}

func (self *Arg2Env) MatchFind(findStr string) bool {
	if strings.Index("write", findStr) >= 0 {
		return true
	}
	for argName, envKey := range self.nameKeys {
		if strings.Index(argName, findStr) >= 0 {
			return true
		}
		if strings.Index(envKey, findStr) >= 0 {
			return true
		}
	}
	return false
}

func (self *Arg2Env) Has(envKey string) bool {
	_, ok := self.keyNames[envKey]
	return ok
}

func (self *Arg2Env) Clone() *Arg2Env {
	cloned := newArg2Env()
	cloned.orderedKeys = append([]string{}, self.orderedKeys...)
	for k, v := range self.keyNames {
		cloned.keyNames[k] = v
	}
	for k, v := range self.nameKeys {
		cloned.nameKeys[k] = v
	}
	return cloned
}
