package builtin

import (
	"runtime"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func LoadPlatformDisplay(_ core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	env = env.GetLayer(core.EnvLayerDefault)
	switch runtime.GOOS {
	case "linux":
	}
	return true
}

func SetDisplayStyle(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	style := argv.GetRaw("style")
	env = env.GetLayer(core.EnvLayerSession)
	env.Set("display.style", style)
	return true
}
