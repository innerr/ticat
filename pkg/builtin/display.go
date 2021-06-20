package builtin

import (
	"runtime"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func LoadPlatformDisplay(_ core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	env = env.GetLayer(core.EnvLayerDefault)
	switch runtime.GOOS {
	}
	return true
}
