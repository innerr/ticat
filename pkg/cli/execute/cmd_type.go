package execute

import (
	"github.com/pingcap/ticat/pkg/builtin"
	"github.com/pingcap/ticat/pkg/cli/core"
)

// TODO: move to command property
func allowCheckEnvOpsFail(flow *core.ParsedCmds) bool {
	last := flow.Cmds[0].LastCmd()
	if last == nil {
		return false
	}
	// Equal to funcs which clear flow
	allows := []interface{}{
		builtin.SaveFlow,
		builtin.DumpCmdNoRecursive,
		builtin.DumpTailCmdInfo,
		builtin.DumpTailCmdSub,
		builtin.DumpTailCmdSubUsage,
		builtin.DumpTailCmdSubDetails,
		builtin.DumpTailCmdUsage,
		builtin.GlobalHelpMoreInfo,
		builtin.GlobalHelpLessInfo,
		builtin.DumpFlowAll,
		builtin.DumpFlowAllSimple,
		builtin.DumpFlow,
		builtin.DumpFlowSimple,
		builtin.DumpFlowDepends,
		builtin.DumpFlowSkeleton,
		builtin.DumpFlowEnvOpsCheckResult,
	}
	for _, allow := range allows {
		if last.IsTheSameFunc(allow) {
			return true
		}
	}
	return false
}

// TODO: try to remove this, it just for better display
func isStartWithSearchCmd(flow *core.ParsedCmds) (isSearch bool) {
	if len(flow.Cmds) == 0 {
		return
	}
	last := flow.Cmds[0].LastCmd()
	if last == nil {
		return
	}
	funcs := []interface{}{
		builtin.GlobalHelpMoreInfo,
		builtin.GlobalHelpLessInfo,
		builtin.DumpTailCmdInfo,
		builtin.DumpTailCmdSub,
		builtin.DumpTailCmdSubUsage,
		builtin.DumpTailCmdSubDetails,
	}
	for _, it := range funcs {
		if last.IsTheSameFunc(it) {
			return true
		}
	}
	return false
}

func allowParseError(flow *core.ParsedCmds) bool {
	if len(flow.Cmds) == 0 {
		return false
	}
	if flow.TailModeCall {
		return flow.Cmds[0].ParseResult.Error == nil
	}
	last := flow.Cmds[0].LastCmd()
	if last == nil {
		return false
	}
	funcs := []interface{}{
		builtin.SaveFlow,
	}
	for _, it := range funcs {
		if last.IsTheSameFunc(it) {
			return true
		}
	}
	return false
}

func noSessionCmds(flow *core.ParsedCmds) bool {
	if flow.HasTailMode {
		return true
	}
	if len(flow.Cmds) != 1 {
		return false
	}
	cmd := flow.Cmds[0].LastCmdNode()
	// No source == builtin command
	if cmd.IsBuiltin() {
		return true
	}
	return false
}
