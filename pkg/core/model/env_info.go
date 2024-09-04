package model

type EnvKeyInfo struct {
	DisplayLen       int
	InvisibleDisplay string
}

type EnvKeysInfo map[string]*EnvKeyInfo

func NewEnvKeysInfo() *EnvKeysInfo {
	return &EnvKeysInfo{}
}

func (self *EnvKeysInfo) GetOrAdd(key string) *EnvKeyInfo {
	val, ok := (*self)[key]
	if !ok {
		val = &EnvKeyInfo{}
		(*self)[key] = val
	}
	return val
}

func (self *EnvKeysInfo) Get(key string) *EnvKeyInfo {
	return (*self)[key]
}
