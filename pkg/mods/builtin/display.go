package builtin

import (
	"runtime"

	"github.com/pingcap/ticat/pkg/core/model"
	"github.com/pingcap/ticat/pkg/utils"
)

func LoadPlatformDisplay(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	env = env.GetLayer(model.EnvLayerDefault)
	switch runtime.GOOS {
	case "linux":
		env.Set("display.utf8.symbols.tip", " ☻ ")
	}
	return currCmdIdx, true
}

func SetDisplayStyle(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	style := argv.GetRaw("style")
	env = env.GetLayer(model.EnvLayerSession)
	env.Set("display.style", style)
	return currCmdIdx, true
}

func SetDisplayWidth(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	width := argv.GetInt("width")
	if width == 0 {
		_, width = utils.GetTerminalWidth(50, 100)
	}
	env = env.GetLayer(model.EnvLayerSession)
	env.SetInt("display.width", width)
	return currCmdIdx, true
}
