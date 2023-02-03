package builtin

import (
	"fmt"
	"strings"
	"time"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func SetSessionsKeepDur(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	env = env.GetLayer(core.EnvLayerSession)
	key := "sys.sessions.keep-status-duration"
	env.SetDur(key, argv.GetRaw("duration"))
	display.PrintTipTitle(cc.Screen, env, "each session status will be kept for '"+env.GetRaw(key)+"'")
	return currCmdIdx, true
}

func RemoveAllSessions(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	cleaned, runnings := core.CleanSessions(env)

	var line string
	if cleaned == 0 && runnings == 0 {
		line = "no executed sessions"
	} else if runnings != 0 {
		line = fmt.Sprintf("removed %v sessions, %v are still running and untouched", cleaned, runnings)
	} else {
		line = fmt.Sprintf("removed all %v sessions", cleaned)
	}
	display.PrintTipTitle(cc.Screen, env, line)
	return currCmdIdx, true
}

func SessionStatus(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	id := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "session-id")
	sessions, _ := findSessions(nil, id, cc, env, 1, true, true, true)
	if len(sessions) == 0 {
		return currCmdIdx, true
	}
	dumpSession(sessions[0], env, cc.Screen, "")
	return currCmdIdx, true
}

func ListSessions(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return listSessions(argv, cc, env, flow, currCmdIdx, true, true, true)
}

func ListSessionsError(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return listSessions(argv, cc, env, flow, currCmdIdx, true, false, false)
}

func ListSessionsDone(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return listSessions(argv, cc, env, flow, currCmdIdx, false, true, false)
}

func ListSessionsRunning(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return listSessions(argv, cc, env, flow, currCmdIdx, false, false, true)
}

func listSessions(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int,
	includeError bool,
	includeDone bool,
	includeRunning bool) (int, bool) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
	cntLimit := argv.GetInt("max-count")
	sessions, total := findSessions(findStrs, "", cc, env, cntLimit, includeError, includeDone, includeRunning)

	screen := display.NewCacheScreen()
	cnt := 0
	for _, it := range sessions {
		if it.Cleaning {
			dumpSession(it, env, screen, "expired")
		} else {
			dumpSession(it, env, screen, "")
		}
		cnt += 1
	}

	if cnt <= 0 {
		return clearFlow(flow)
	} else {
		cntStr := "1 session"
		if cnt > 1 {
			cntStr = fmt.Sprintf("%d sessions", cnt)
		}
		if cntLimit <= 0 {
			display.PrintTipTitle(cc.Screen, env, fmt.Sprintf("%s matched:", cntStr))
		} else {
			display.PrintTipTitle(cc.Screen, env, fmt.Sprintf("%s matched: (%d total, %d showed)", cntStr, total, cnt))
		}
	}

	screen.WriteTo(cc.Screen)
	if display.TooMuchOutput(env, screen) {
		limitStr := "use arg 'max-count'"
		display.PrintTipTitle(cc.Screen, env,
			fmt.Sprintf("too many sessions(%d total, %d showed), %s, or add more filter keywords", total, cnt, limitStr))
	}

	return clearFlow(flow)
}

func ErrorSessionDescLess(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	session, ok := getLastSession(env, true, false, false)
	if !ok {
		panic(fmt.Errorf("no executed error sessions"))
	}
	descSession(session, argv, cc, env, true, false, false, false)
	return currCmdIdx, true
}

func ErrorSessionDescMore(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	session, ok := getLastSession(env, true, false, false)
	if !ok {
		panic(fmt.Errorf("no executed error sessions"))
	}
	descSession(session, argv, cc, env, true, false, true, false)
	return currCmdIdx, true
}

func ErrorSessionDescFull(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	session, ok := getLastSession(env, true, false, false)
	if !ok {
		panic(fmt.Errorf("no executed error sessions"))
	}
	descSession(session, argv, cc, env, false, true, true, false)
	return currCmdIdx, true
}

func RunningSessionDescLess(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	session, ok := getLastSession(env, false, false, true)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no running sessions")
	} else {
		descSession(session, argv, cc, env, true, false, false, false)
	}
	return currCmdIdx, true
}

func DoneSessionDescLess(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	session, ok := getLastSession(env, false, true, false)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no executed sessions")
	} else {
		descSession(session, argv, cc, env, true, false, false, false)
	}
	return currCmdIdx, true
}

func DoneSessionDescMore(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	session, ok := getLastSession(env, false, true, false)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no executed sessions")
	} else {
		descSession(session, argv, cc, env, true, false, true, false)
	}
	return currCmdIdx, true
}

func DoneSessionDescFull(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	session, ok := getLastSession(env, false, true, false)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no executed sessions")
	} else {
		descSession(session, argv, cc, env, false, true, true, false)
	}
	return currCmdIdx, true
}

func RunningSessionDescMore(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	session, ok := getLastSession(env, false, false, true)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no running sessions")
	} else {
		descSession(session, argv, cc, env, true, false, true, false)
	}
	return currCmdIdx, true
}

func RunningSessionDescFull(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	session, ok := getLastSession(env, false, false, true)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no running sessions")
	} else {
		descSession(session, argv, cc, env, false, true, true, false)
	}
	return currCmdIdx, true
}

func RunningSessionDescMonitor(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	session, ok := getLastRunningSession(cc, env)
	if !ok {
		cc.Screen.Print(display.ColorTip("no running sessions", env) + "\n")
	} else {
		descSession(session, argv, cc, env, true, false, false, true)
	}
	return currCmdIdx, true
}

func RemoveSession(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	force := argv.GetBool("remove-running")
	id := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "session-id")
	sessions, _ := findSessions(nil, id, cc, env, 1, true, true, true)
	if len(sessions) == 0 {
		return currCmdIdx, true
	}
	removeAndDumpSession(sessions[0], cc.Screen, env, force)
	return currCmdIdx, true
}

func FindAndRemoveSessions(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	force := argv.GetBool("remove-running")

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
	sessions, _ := findSessions(findStrs, "", cc, env, 1, true, true, true)
	if len(sessions) == 0 {
		return currCmdIdx, true
	}

	cleaneds := 0
	for _, it := range sessions {
		cleaned := removeAndDumpSession(it, cc.Screen, env, force)
		if cleaned {
			cleaneds += 1
		}
	}

	if len(sessions) > 1 {
		display.PrintTipTitle(cc.Screen, env,
			fmt.Sprintf("%d of %d sessions removed", cleaneds, len(sessions)))
	}
	return clearFlow(flow)
}

func removeAndDumpSession(session core.SessionStatus, screen core.Screen, env *core.Env, force bool) bool {
	status := "removed"
	cleaned, running := core.CleanSession(session, env, force)
	if running {
		if cleaned {
			status = display.ColorWarn("running, force removed", env)
		} else {
			status = display.ColorWarn("running, untouched", env)
		}
	} else if !cleaned {
		status = display.ColorError("remove failed", env)
	}
	dumpSession(session, env, screen, status)
	return cleaned
}

func SessionDescLess(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	id := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "session-id")
	sessions, _ := findSessions(nil, id, cc, env, 1, true, true, true)
	if len(sessions) == 0 {
		return currCmdIdx, true
	}
	descSession(sessions[0], argv, cc, env, true, false, false, false)
	return currCmdIdx, true
}

func SessionDescMore(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	id := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "session-id")
	sessions, _ := findSessions(nil, id, cc, env, 1, true, true, true)
	if len(sessions) == 0 {
		return currCmdIdx, true
	}
	descSession(sessions[0], argv, cc, env, true, false, true, false)
	return currCmdIdx, true
}

func SessionDescMonitor(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	id := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "session-id")
	sessions, _ := findSessions(nil, id, cc, env, 1, true, true, true)
	if len(sessions) == 0 {
		return currCmdIdx, true
	}
	descSession(sessions[0], argv, cc, env, true, false, false, true)
	return currCmdIdx, true
}

func SessionDescFull(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	id := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "session-id")
	sessions, _ := findSessions(nil, id, cc, env, 1, true, true, true)
	if len(sessions) == 0 {
		return currCmdIdx, true
	}
	descSession(sessions[0], argv, cc, env, false, true, true, false)
	return currCmdIdx, true
}

func LastSession(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	session, ok := getLastSession(env, true, true, true)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no executed/running sessions")
	} else {
		dumpSession(session, env, cc.Screen, "")
	}
	return currCmdIdx, true
}

func LastSessionDescLess(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	session, ok := getLastSession(env, true, true, true)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no executed/running sessions")
	} else {
		descSession(session, argv, cc, env, true, false, false, false)
	}
	return currCmdIdx, true
}

func LastSessionDescMore(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	session, ok := getLastSession(env, true, true, true)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no executed/running sessions")
	} else {
		descSession(session, argv, cc, env, true, false, true, false)
	}
	return currCmdIdx, true
}

func LastSessionDescFull(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	session, ok := getLastSession(env, true, true, true)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no executed/running sessions")
	} else {
		descSession(session, argv, cc, env, false, true, true, false)
	}
	return currCmdIdx, true
}

func SessionRetry(argv core.ArgVals, cc *core.Cli, env *core.Env) (flow []string, masks []*core.ExecuteMask, ok bool) {
	id := argv.GetRaw("session-id")
	if len(id) == 0 {
		panic(fmt.Errorf("[SessionRetry] arg 'session-id' is empty"))
	}
	sessions, _ := findSessions(nil, id, cc, env, 1, true, true, true)
	if len(sessions) == 0 {
		return
	}
	if len(sessions) > 1 {
		panic(fmt.Errorf("[SessionRetry] should never happen"))
	}
	return retrySession(sessions[0], false)
}

func LastSessionRetry(argv core.ArgVals, cc *core.Cli, env *core.Env) (flow []string, masks []*core.ExecuteMask, ok bool) {
	session, ok := getLastSession(env, true, true, true)
	if !ok {
		panic(fmt.Errorf("no executed sessions"))
	}
	return retrySession(session, false)
}

func LastErrorSessionRetry(argv core.ArgVals, cc *core.Cli, env *core.Env) (flow []string, masks []*core.ExecuteMask, ok bool) {
	session, ok := getLastSession(env, true, false, false)
	if !ok {
		panic(fmt.Errorf("no executed sessions"))
	}
	return retrySession(session, false)
}

func findSessions(
	findStrs []string,
	id string,
	cc *core.Cli,
	env *core.Env,
	cntLimit int,
	includeError bool,
	includeDone bool,
	includeRunning bool) (sessions []core.SessionStatus, total int) {

	id = normalizeSid(id)

	sessions, total = core.ListSessions(env, findStrs, id, cntLimit, includeError, includeDone, includeRunning)
	if len(sessions) == 0 {
		if len(id) == 0 {
			if len(findStrs) > 0 {
				display.PrintTipTitle(cc.Screen, env, "no executed sessions found by '"+strings.Join(findStrs, " ")+"'")
			} else {
				display.PrintTipTitle(cc.Screen, env, "no executed sessions")
			}
		} else {
			if len(findStrs) > 0 {
				display.PrintTipTitle(cc.Screen, env,
					"no executed sessions found by '"+strings.Join(findStrs, " ")+"' with id = '"+id+"'")
			} else {
				display.PrintErrTitle(cc.Screen, env, "no executed sessiond with id = '"+id+"'")
			}
		}
	}
	return
}

func dumpSession(session core.SessionStatus, env *core.Env, screen core.Screen, status string) {
	selfName := env.GetRaw("strs.self-name")

	screen.Print(display.ColorSession("["+session.DirName+"]", env) + " " + display.ColorExplain(status, env) + "\n")

	screen.Print(display.ColorProp("    cmd:\n", env))
	screen.Print(display.ColorFlow(fmt.Sprintf("        %s %s\n", selfName, session.Status.Flow), env))

	if len(session.Status.Corrupted) != 0 {
		screen.Print(display.ColorProp("    corrupted-status:\n", env))
		screen.Print(display.ColorError("        [FOR DEBUG]\n", env))
		for _, line := range session.Status.Corrupted {
			screen.Print("        " + display.ColorExplain(line, env) + "\n")
		}
	}

	screen.Print(display.ColorProp("    start-at:\n", env))
	screen.Print(fmt.Sprintf("        %s\n", session.StartTs.Format(core.SessionTimeFormat)))
	screen.Print(fmt.Sprintf("        "+display.ColorExplain("%s ago\n", env),
		time.Now().Sub(session.StartTs).Round(time.Second).String()))

	if !session.Status.FinishTs.IsZero() {
		if !session.Running {
			screen.Print(display.ColorProp("    finish-at:\n", env))
			screen.Print(fmt.Sprintf("        %s\n", session.Status.FinishTs.Format(core.SessionTimeFormat)))
			incompletedDurStr := ""
			if !session.Running && session.Status.Result == core.ExecutedResultIncompleted {
				incompletedDurStr = "+?"
			}
			screen.Print(fmt.Sprintf("        "+display.ColorExplain("%s"+
				incompletedDurStr+" ago, elapsed %s"+incompletedDurStr, env)+"\n",
				time.Now().Sub(session.Status.FinishTs).Round(time.Second),
				session.Status.FinishTs.Sub(session.StartTs).Round(time.Second)))
		}
	}

	screen.Print(display.ColorProp("    status:\n", env))
	if session.Running {
		screen.Print("        " + display.ColorCmdCurr("running", env) + "\n")
		screen.Print(display.ColorProp("    pid:\n", env))
		screen.Print(fmt.Sprintf("        %v\n", session.Pid))
	} else if session.Status.Result == core.ExecutedResultSucceeded {
		screen.Print("        " + display.ColorCmdDone(string(session.Status.Result), env) + "\n")
	} else if session.Status.Result == core.ExecutedResultError {
		screen.Print("        " + display.ColorError(string(session.Status.Result), env) + "\n")
	} else if session.Status.Result == core.ExecutedResultIncompleted {
		screen.Print("        " + display.ColorWarn("failed\n", env))
	} else {
		screen.Print("        " + string(session.Status.Result) + "\n")
	}
}

func descSession(session core.SessionStatus, argv core.ArgVals, cc *core.Cli, env *core.Env,
	skeleton, showEnvFull bool, showModifiedEnv bool, monitorMode bool) {

	dumpArgs := display.NewDumpFlowArgs().SetMaxDepth(argv.GetIntEx("depth", 32)).SetMaxTrivial(argv.GetIntEx("unfold-trivial", 1))
	if skeleton {
		if !showEnvFull && !showModifiedEnv {
			dumpArgs.SetSkeleton()
		} else {
			dumpArgs.SetSimple()
		}
	}
	if showEnvFull {
		dumpArgs.SetShowExecutedEnvFull()
	}
	if showModifiedEnv {
		dumpArgs.SetShowExecutedModifiedEnv()
	}
	if monitorMode {
		dumpArgs.SetMonitorMode()
	}

	flow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, core.FlowStrToStrs(session.Status.Flow)...)
	display.DumpFlowEx(cc, env, flow, 0, dumpArgs, session.Status, session.Running, EnvOpCmds())
}

func getLastSession(env *core.Env, includeError bool, includeDone bool, includeRunning bool) (session core.SessionStatus, ok bool) {
	var sessions []core.SessionStatus
	currSession := env.GetRaw("sys.session.id")
	cntLimit := 2
	if len(currSession) == 0 {
		cntLimit = 1
	}
	sessions, _ = core.ListSessions(env, nil, "", cntLimit, includeError, includeDone, includeRunning)
	for i := len(sessions) - 1; i >= 0; i-- {
		session := sessions[i]
		if session.SessionId() != currSession {
			return session, true
		}
	}
	return
}

func getLastRunningSession(cc *core.Cli, env *core.Env) (session core.SessionStatus, ok bool) {
	var sessions []core.SessionStatus
	currSession := env.GetRaw("sys.session.id")

	sessions, _ = core.ListSessions(env, nil, "", 64, false, false, true)
	for i := len(sessions) - 1; i >= 0; i-- {
		session := sessions[i]
		if session.Status == nil {
			continue
		}
		if session.Status.Result == core.ExecutedResultSucceeded {
			continue
		}
		if len(session.Status.Cmds) == 0 {
			continue
		}
		lastCmdStr := session.Status.Cmds[len(session.Status.Cmds)-1].Cmd
		if parsedCmd, ok := cc.ParseCmd(false, core.FlowStrToStrs(lastCmdStr)...); ok {
			if parsedCmd.LastCmd() != nil && parsedCmd.LastCmd().IsHideInSessionsLast() {
				continue
			}
		}
		if session.SessionId() == currSession {
			continue
		}
		return session, true
	}
	return
}

func retrySession(session core.SessionStatus, errSessionOnly bool) (flow []string, masks []*core.ExecuteMask, ok bool) {
	if errSessionOnly && session.Status != nil && session.Status.Result == core.ExecutedResultSucceeded {
		panic(fmt.Errorf("session [%s] had succeeded, nothing to retry", session.DirName))
	}
	return []string{session.Status.Flow}, session.Status.GenExecMasks(), true
}

func normalizeSid(id string) string {
	// TODO: put brackets to env?
	return strings.TrimRight(strings.TrimLeft(id, "["), "]")
}
