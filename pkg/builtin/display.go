package builtin

import (
	"runtime"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/utils"
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
		env.Set("display.utf8.symbols.tip", " ☻ ")
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

func SetDisplayWidth(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	width := argv.GetInt("width")
	if width == 0 {
		_, width = utils.GetTerminalWidth(50, 100)
	}
	env = env.GetLayer(core.EnvLayerSession)
	env.SetInt("display.width", width)
	return currCmdIdx, true
}
