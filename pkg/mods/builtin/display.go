package builtin

import (
	"runtime"
	"strings"

	"github.com/innerr/ticat/pkg/cli/display"
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

func AddEnvFilterPrefix(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	prefix := argv.GetRaw("prefix")
	if len(prefix) == 0 {
		return currCmdIdx, model.NewCmdError(flow.Cmds[currCmdIdx], "arg 'prefix' is empty")
	}

	env = env.GetLayer(model.EnvLayerSession)
	current := env.GetRaw("display.env.filter.prefix")
	listSep := env.GetRaw("strs.list-sep")

	if len(current) == 0 {
		env.Set("display.env.filter.prefix", prefix)
	} else {
		prefixes := strings.Split(current, listSep)
		for _, p := range prefixes {
			if p == prefix {
				_ = cc.Screen.Print(display.ColorWarn("prefix '"+prefix+"' already in filter list\n", env))
				return currCmdIdx, nil
			}
		}
		env.Set("display.env.filter.prefix", current+listSep+prefix)
	}

	_ = cc.Screen.Print(display.ColorKey("display.env.filter.prefix", env) +
		display.ColorSymbol(" = ", env) +
		display.ColorCmd(env.GetRaw("display.env.filter.prefix"), env) + "\n")
	return currCmdIdx, nil
}

func RemoveEnvFilterPrefix(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	prefix := argv.GetRaw("prefix")
	if len(prefix) == 0 {
		return currCmdIdx, model.NewCmdError(flow.Cmds[currCmdIdx], "arg 'prefix' is empty")
	}

	env = env.GetLayer(model.EnvLayerSession)
	current := env.GetRaw("display.env.filter.prefix")
	listSep := env.GetRaw("strs.list-sep")

	if len(current) == 0 {
		_ = cc.Screen.Print(display.ColorWarn("filter list is empty\n", env))
		return currCmdIdx, nil
	}

	prefixes := strings.Split(current, listSep)
	var newPrefixes []string
	found := false
	for _, p := range prefixes {
		if p == prefix {
			found = true
			continue
		}
		newPrefixes = append(newPrefixes, p)
	}

	if !found {
		_ = cc.Screen.Print(display.ColorWarn("prefix '"+prefix+"' not in filter list\n", env))
		return currCmdIdx, nil
	}

	if len(newPrefixes) == 0 {
		env.Delete("display.env.filter.prefix")
		_ = cc.Screen.Print(display.ColorExplain("filter list cleared\n", env))
	} else {
		env.Set("display.env.filter.prefix", strings.Join(newPrefixes, listSep))
		_ = cc.Screen.Print(display.ColorKey("display.env.filter.prefix", env) +
			display.ColorSymbol(" = ", env) +
			display.ColorCmd(env.GetRaw("display.env.filter.prefix"), env) + "\n")
	}
	return currCmdIdx, nil
}

func ListEnvFilterPrefixes(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	current := env.GetRaw("display.env.filter.prefix")
	if len(current) == 0 {
		_ = cc.Screen.Print(display.ColorExplain("(no filter prefixes set)\n", env))
		return currCmdIdx, nil
	}

	listSep := env.GetRaw("strs.list-sep")
	prefixes := strings.Split(current, listSep)
	for _, p := range prefixes {
		_ = cc.Screen.Print(display.ColorKey(p, env) + "\n")
	}
	return currCmdIdx, nil
}

func ClearEnvFilterPrefixes(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	env = env.GetLayer(model.EnvLayerSession)
	env.Delete("display.env.filter.prefix")
	_ = cc.Screen.Print(display.ColorExplain("all filter prefixes cleared\n", env))
	return currCmdIdx, nil
}
