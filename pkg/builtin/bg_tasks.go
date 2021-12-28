package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
	"github.com/pingcap/ticat/pkg/utils"
)

func WaitForBgTaskFinish(
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

	tid := argv.GetRaw("thread")
	if len(tid) == 0 {
		return WaitForAllBgTasksFinish(argv, cc, env, flow, currCmdIdx)
	}

	errs := WaitBgTasks(cc, env, tid)
	for _, err := range errs {
		display.PrintError(cc, env, err)
	}
	return currCmdIdx, true
}

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
	errs := WaitBgTasks(cc, env, "")
	for _, err := range errs {
		display.PrintError(cc, env, err)
	}
	return currCmdIdx, true
}

func WaitBgTasks(cc *core.Cli, env *core.Env, matchTid string) (errs []error) {
	preTid := utils.GoRoutineIdStr()
	for {
		tid, task, ok := cc.BgTasks.GetEarliestTask()
		if !ok {
			break
		}
		if len(matchTid) != 0 && matchTid != tid {
			continue
		}
		info := task.GetStat()

		display.PrintSwitchingThreadDisplay(preTid, info, env, cc.Screen)

		cc.BgTasks.BringBgTaskToFront(tid, cc.CmdIO.CmdStdout)
		err := task.WaitForFinish()
		if err != nil {
			errs = append(errs, err)
		}
		cc.BgTasks.RemoveTask(tid)
		preTid = tid
	}
	return
}
