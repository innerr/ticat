package core

const (
	EnvOpTypeRead uint = 0x01
	EnvOpTypeWrite uint = 0x02
	EnvOpTypeMayRead uint = 0x04
	EnvOpTypeMayWrite uint = 0x08
)

type EnvOps struct {
	orderedNames []string
	ops map[string]uint
}

func newEnvOps() EnvOps {
	return EnvOps{nil, map[string]uint{}}
}

func (self *EnvOps) AddOp(name string, op uint) {
	old, ok := self.ops[name]
	if !ok {
		self.orderedNames = append(self.orderedNames, name)
	}
	self.ops[name] = old | op
}
