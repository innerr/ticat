package execute

import (
	"github.com/pingcap/ticat/pkg/builtin"
	"github.com/pingcap/ticat/pkg/cli/core"
)

// TODO: better way to do this

// TODO: move to command property

func allowCheckEnvOpsFail(flow *core.ParsedCmds) bool {
	last := flow.Cmds[0].LastCmd()
	if last == nil {
		return false
	}
	allows := []interface{}{
		builtin.DumpCmdNoRecursive,
		builtin.SaveFlow,
		builtin.DumpTellTailCmd,
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

func doNotFilterEmptyCmds(flow *core.ParsedCmds) bool {
	last := flow.Cmds[len(flow.Cmds)-1].LastCmd()
	if last == nil {
		return false
	}
	allows := []interface{}{
		builtin.GlobalHelp,
		builtin.GlobalHelpMoreInfo,
		builtin.GlobalHelpLessInfo,
	}
	for _, allow := range allows {
		if last.IsTheSameFunc(allow) {
			return true
		}
	}
	return false
}

func isEndWithSearchCmd(flow *core.ParsedCmds) (isSearch, isLess, isMore bool) {
	if len(flow.Cmds) == 0 {
		return
	}
	cmd := flow.Cmds.LastCmd()
	last := cmd.LastCmd()
	if last == nil {
		return
	}
	if last.IsTheSameFunc(builtin.GlobalHelpMoreInfo) {
		isSearch = true
		isMore = true
	} else if last.IsTheSameFunc(builtin.GlobalHelpLessInfo) {
		isSearch = true
		isLess = true
	} else if last.IsTheSameFunc(builtin.DumpTellTailCmd) {
		isSearch = true
	}
	return
}
