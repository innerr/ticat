package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
	"github.com/pingcap/ticat/pkg/utils"
)

func WaitForLatestBgTaskFinish(
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

	tid, task, ok := cc.BgTasks.GetLatestTask()
	if ok {
		errs := WaitBgTask(cc, env, tid, task)
		for _, err := range errs {
			display.PrintError(cc, env, err)
		}
	}
	return currCmdIdx, true
}

func WaitForBgTaskFinishByName(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	bgCmd := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "command")

	if utils.GoRoutineIdStr() != utils.GoRoutineIdStrMain {
		panic(core.NewCmdError(flow.Cmds[currCmdIdx],
			"must be in main thread to wait for other threads to finish"))
	}

	verifiedCmd := cc.NormalizeCmd(true, bgCmd)
	tid, task, ok := cc.BgTasks.GetTaskByCmd(verifiedCmd)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "command '"+bgCmd+"' not in background")
	} else {
		errs := WaitBgTask(cc, env, tid, task)
		for _, err := range errs {
			display.PrintError(cc, env, err)
		}
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
	errs := WaitBgTasks(cc, env, true)
	for _, err := range errs {
		display.PrintError(cc, env, err)
	}
	return currCmdIdx, true
}

func WaitBgTask(cc *core.Cli, env *core.Env, tid string, task *core.BgTask) (errs []error) {
	info := task.GetStat()
	display.PrintSwitchingThreadDisplay(utils.GoRoutineIdStr(), info, env, cc.Screen, true)

	cc.BgTasks.BringBgTaskToFront(tid, cc.CmdIO.CmdStdout)
	err := task.WaitForFinish()
	if err != nil {
		errs = append(errs, err)
	}
	cc.BgTasks.RemoveTask(tid)
	return
}

func WaitBgTasks(cc *core.Cli, env *core.Env, manual bool) (errs []error) {
	preTid := utils.GoRoutineIdStr()
	for {
		tid, task, ok := cc.BgTasks.GetEarliestTask()
		if !ok {
			break
		}
		info := task.GetStat()

		display.PrintSwitchingThreadDisplay(preTid, info, env, cc.Screen, manual)

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
