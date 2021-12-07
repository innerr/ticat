package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
	"github.com/pingcap/ticat/pkg/utils"
)

func WaitForAllBgTasksFinish(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	if utils.GoRoutineIdStr() != utils.GoRoutineIdStrMain {
		panic(core.NewCmdError(flow.Cmds[currCmdIdx],
			"must be in main thread to wait for other threads to finish"))
	}
	WaitAllBgTasks(cc)
	return currCmdIdx, true
}

func WaitAllBgTasks(cc *core.Cli) {
	preTid := utils.GoRoutineIdStr()
	for {
		tid, task, ok := cc.BgTasks.GetEarliestTask()
		if !ok {
			break
		}
		info := task.GetStat()

		display.PrintSwitchingThreadDisplay(preTid, info, cc.GlobalEnv, cc.Screen)

		cc.BgTasks.BringBgTaskToFront(tid, cc.CmdIO.CmdStdout)
		task.WaitForFinish()
		cc.BgTasks.RemoveTask(tid)
		preTid = tid
	}
}
