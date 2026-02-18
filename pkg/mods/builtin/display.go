package builtin

import (
	"runtime"

	"github.com/innerr/ticat/pkg/core/model"
	"github.com/innerr/ticat/pkg/utils"
)

func LoadPlatformDisplay(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	env = env.GetLayer(model.EnvLayerDefault)
	switch runtime.GOOS {
	case "linux":
		env.Set("display.utf8.symbols.tip", " ☻ ")
	}
	return currCmdIdx, nil
}

func SetDisplayStyle(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	style := argv.GetRaw("style")
	env = env.GetLayer(model.EnvLayerSession)
	env.Set("display.style", style)
	return currCmdIdx, nil
}

func SetDisplayWidth(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	width := argv.GetInt("width")
	if width == 0 {
		_, width = utils.GetTerminalWidth(50, 100)
	}
	env = env.GetLayer(model.EnvLayerSession)
	env.SetInt("display.width", width)
	return currCmdIdx, nil
}
