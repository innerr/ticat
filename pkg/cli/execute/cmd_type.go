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
	// This list equals to funcs which will do 'clear-the-flow'
	allows := []interface{}{
		builtin.SaveFlow,
		builtin.GlobalFindCmds,
		builtin.GlobalFindCmdsWithUsage,
		builtin.GlobalFindCmdsWithDetails,
		builtin.DumpTailCmdWithUsage,
		builtin.DumpTailCmdWithDetails,
		//builtin.DumpTailCmdSub,
		//builtin.DumpTailCmdSubWithUsage,
		//builtin.DumpTailCmdSubWithDetails,
		builtin.DumpFlowAll,
		builtin.DumpFlowAllSimple,
		builtin.DumpFlow,
		builtin.DumpFlowSimple,
		builtin.DumpFlowDepends,
		builtin.DumpFlowSkeleton,
		builtin.DumpFlowEnvOpsCheckResult,
		builtin.DumpCmds,
		builtin.DumpCmdsWithUsage,
		builtin.DumpCmdsWithDetails,
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
		builtin.GlobalFindCmds,
		builtin.GlobalFindCmdsWithUsage,
		builtin.GlobalFindCmdsWithDetails,
		builtin.DumpCmds,
		builtin.DumpCmdsWithUsage,
		builtin.DumpCmdsWithDetails,
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

	if !cmd.IsBuiltin() {
		return false
	}

	funcs := []interface{}{
		builtin.DbgInteract,
		builtin.SessionRetry,
		//builtin.LastSessionRetry,
	}
	for _, it := range funcs {
		if cmd.Cmd() != nil && cmd.Cmd().IsTheSameFunc(it) {
			return false
		}
	}
	return true
}
