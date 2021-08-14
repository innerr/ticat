package builtin

import (
	"runtime"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func LoadPlatformDisplay(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	env = env.GetLayer(core.EnvLayerDefault)
	switch runtime.GOOS {
	case "linux":
	}
	return currCmdIdx, true
}

func SetDisplayStyle(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	style := argv.GetRaw("style")
	env = env.GetLayer(core.EnvLayerSession)
	env.Set("display.style", style)
	return currCmdIdx, true
}
