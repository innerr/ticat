package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

func BlenderForestMode(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	if !cc.Blender.ForestMode.AtForestTopLvl(env) {
		cc.Blender.ForestMode.Push(core.GetLastStackFrame(env))
	}
	return currCmdIdx, true
}
