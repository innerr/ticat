package core

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

func (self *Val2Env) Val(envKey string) string {
	return self.pairs[envKey]
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

func (self *Arg2Env) EnvKeys() []string {
	return self.orderedKeys
}

func (self *Arg2Env) ArgName(envKey string) string {
	return self.keyNames[envKey]
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
