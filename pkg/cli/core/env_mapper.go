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
