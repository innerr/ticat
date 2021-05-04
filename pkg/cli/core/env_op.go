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

func (self *EnvOps) EnvKeys() []string {
	return self.orderedNames
}

func (self *EnvOps) Ops(name string) (vals []uint) {
	val, ok := self.ops[name]
	if !ok || val == 0 {
		return
	}
	if (val & EnvOpTypeMayRead) != 0 {
		if (val & EnvOpTypeRead) != 0 {
			vals = append(vals, EnvOpTypeRead)
		} else {
			vals = append(vals, EnvOpTypeMayRead)
		}
	} else if (val & EnvOpTypeRead) != 0 {
		vals = append(vals, EnvOpTypeRead)
	}
	if (val & EnvOpTypeMayWrite) != 0 {
		if (val & EnvOpTypeWrite) != 0 {
			vals = append(vals, EnvOpTypeWrite)
		} else {
			vals = append(vals, EnvOpTypeMayWrite)
		}
	} else if (val & EnvOpTypeWrite) != 0 {
		vals = append(vals, EnvOpTypeWrite)
	}
	return vals
}
