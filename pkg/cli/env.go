package cli

type EnvLayerType uint

const (
	EnvLayerDefault EnvLayerType = iota
	EnvLayerPersisted
	EnvLayerSession
	EnvLayerCmdArg
)

func EnvLayerName(tp EnvLayerType) string {
	switch tp {
	case EnvLayerDefault:
		return "default"
	case EnvLayerPersisted:
		return "persisted"
	case EnvLayerSession:
		return "session"
	case EnvLayerCmdArg:
		return "arg"
	default:
		panic(fmt.Errorf("unknown layer type, value: %v", tp))
	}
}

type EnvVal struct {
	Raw string
	Val interface{}
}

type Env struct {
	pairs map[string]*EnvVal
	parent *Env
}

func NewEnv() *Env {
	return &Env{ map[string]*EnvVal{}, nil }
}

func (self *Env) NewLayer(tp EnLayerType) *Env {
	env := NewEnv()
	env.parent = self
	return env
}

func (self *Env) Get(name string) *EnvVal {
	val, ok := self.pairs[name]
	if !ok && self.parent != nil {
		return self.parent.Get(name)
	}
	return val
}

func (self *Env) Set(name string, val *EnvVal) *EnvVal {
	old, _ := self.pairs[name]
	self.pairs[name] = val
	return old
}
