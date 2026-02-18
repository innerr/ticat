package builtin

import (
	"fmt"
	"strings"
	"time"

	"github.com/innerr/ticat/pkg/cli/display"
	"github.com/innerr/ticat/pkg/core/model"
)

func SetSessionsKeepDur(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	env = env.GetLayer(model.EnvLayerSession)
	key := "sys.sessions.keep-status-duration"
	env.SetDur(key, argv.GetRaw("duration"))
	display.PrintTipTitle(cc.Screen, env, "each session status will be kept for '"+env.GetRaw(key)+"'")
	return currCmdIdx, nil
}

func RemoveAllSessions(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	cleaned, runnings := model.CleanSessions(env)

	var line string
	if cleaned == 0 && runnings == 0 {
		line = "no executed sessions"
	} else if runnings != 0 {
		line = fmt.Sprintf("removed %v sessions, %v are still running and untouched", cleaned, runnings)
	} else {
		line = fmt.Sprintf("removed all %v sessions", cleaned)
	}
	display.PrintTipTitle(cc.Screen, env, line)
	return currCmdIdx, nil
}

func SessionStatus(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	id, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "session-id")
	if err != nil {
		return currCmdIdx, err
	}
	sessions, _ := findSessions(nil, id, cc, env, 1, true, true, true)
	if len(sessions) == 0 {
		return currCmdIdx, nil
	}
	dumpSession(sessions[0], env, cc.Screen, "")
	return currCmdIdx, nil
}

func ListSessions(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	return listSessions(argv, cc, env, flow, currCmdIdx, true, true, true)
}

func ListSessionsError(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	return listSessions(argv, cc, env, flow, currCmdIdx, true, false, false)
}

func ListSessionsDone(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	return listSessions(argv, cc, env, flow, currCmdIdx, false, true, false)
}

func ListSessionsRunning(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	return listSessions(argv, cc, env, flow, currCmdIdx, false, false, true)
}

func listSessions(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int,
	includeError bool,
	includeDone bool,
	includeRunning bool) (int, error) {

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
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	session, ok := getLastSession(cc, env, true, false, false)
	if !ok {
		return currCmdIdx, fmt.Errorf("no executed error sessions")
	}
	descSession(session, argv, cc, env, true, false, false, false)
	return currCmdIdx, nil
}

func ErrorSessionDescMore(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	session, ok := getLastSession(cc, env, true, false, false)
	if !ok {
		return currCmdIdx, fmt.Errorf("no executed error sessions")
	}
	descSession(session, argv, cc, env, true, false, true, false)
	return currCmdIdx, nil
}

func ErrorSessionDescFull(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	session, ok := getLastSession(cc, env, true, false, false)
	if !ok {
		return currCmdIdx, fmt.Errorf("no executed error sessions")
	}
	descSession(session, argv, cc, env, false, true, true, false)
	return currCmdIdx, nil
}

func RunningSessionDescLess(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	session, ok := getLastSession(cc, env, false, false, true)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no running sessions")
	} else {
		descSession(session, argv, cc, env, true, false, false, false)
	}
	return currCmdIdx, nil
}

func DoneSessionDescLess(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	session, ok := getLastSession(cc, env, false, true, false)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no executed sessions")
	} else {
		descSession(session, argv, cc, env, true, false, false, false)
	}
	return currCmdIdx, nil
}

func DoneSessionDescMore(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	session, ok := getLastSession(cc, env, false, true, false)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no executed sessions")
	} else {
		descSession(session, argv, cc, env, true, false, true, false)
	}
	return currCmdIdx, nil
}

func DoneSessionDescFull(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	session, ok := getLastSession(cc, env, false, true, false)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no executed sessions")
	} else {
		descSession(session, argv, cc, env, false, true, true, false)
	}
	return currCmdIdx, nil
}

func RunningSessionDescMore(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	session, ok := getLastSession(cc, env, false, false, true)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no running sessions")
	} else {
		descSession(session, argv, cc, env, true, false, true, false)
	}
	return currCmdIdx, nil
}

func RunningSessionDescFull(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	session, ok := getLastSession(cc, env, false, false, true)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no running sessions")
	} else {
		descSession(session, argv, cc, env, false, true, true, false)
	}
	return currCmdIdx, nil
}

func RunningSessionDescMonitor(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	session, ok := getLastSession(cc, env, false, false, true)
	if !ok {
		cc.Screen.Print(display.ColorTip("no running sessions", env) + "\n")
	} else {
		descSession(session, argv, cc, env, true, false, false, true)
	}
	return currCmdIdx, nil
}

func RemoveSession(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	force := argv.GetBool("remove-running")
	id, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "session-id")
	if err != nil {
		return currCmdIdx, err
	}
	sessions, _ := findSessions(nil, id, cc, env, 1, true, true, true)
	if len(sessions) == 0 {
		return currCmdIdx, nil
	}
	removeAndDumpSession(sessions[0], cc.Screen, env, force)
	return currCmdIdx, nil
}

func FindAndRemoveSessions(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	force := argv.GetBool("remove-running")

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
	sessions, _ := findSessions(findStrs, "", cc, env, 1, true, true, true)
	if len(sessions) == 0 {
		return currCmdIdx, nil
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

func removeAndDumpSession(session model.SessionStatus, screen model.Screen, env *model.Env, force bool) bool {
	status := "removed"
	cleaned, running := model.CleanSession(session, env, force)
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
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	id, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "session-id")
	if err != nil {
		return currCmdIdx, err
	}
	sessions, _ := findSessions(nil, id, cc, env, 1, true, true, true)
	if len(sessions) == 0 {
		return currCmdIdx, nil
	}
	descSession(sessions[0], argv, cc, env, true, false, false, false)
	return currCmdIdx, nil
}

func SessionDescMore(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	id, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "session-id")
	if err != nil {
		return currCmdIdx, err
	}
	sessions, _ := findSessions(nil, id, cc, env, 1, true, true, true)
	if len(sessions) == 0 {
		return currCmdIdx, nil
	}
	descSession(sessions[0], argv, cc, env, true, false, true, false)
	return currCmdIdx, nil
}

func SessionDescMonitor(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	id, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "session-id")
	if err != nil {
		return currCmdIdx, err
	}
	sessions, _ := findSessions(nil, id, cc, env, 1, true, true, true)
	if len(sessions) == 0 {
		return currCmdIdx, nil
	}
	descSession(sessions[0], argv, cc, env, true, false, false, true)
	return currCmdIdx, nil
}

func SessionDescFull(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	id, err := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "session-id")
	if err != nil {
		return currCmdIdx, err
	}
	sessions, _ := findSessions(nil, id, cc, env, 1, true, true, true)
	if len(sessions) == 0 {
		return currCmdIdx, nil
	}
	descSession(sessions[0], argv, cc, env, false, true, true, false)
	return currCmdIdx, nil
}

func LastSession(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	session, ok := getLastSession(cc, env, true, true, true)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no executed/running sessions")
	} else {
		dumpSession(session, env, cc.Screen, "")
	}
	return currCmdIdx, nil
}

func LastSessionDescLess(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	session, ok := getLastSession(cc, env, true, true, true)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no executed/running sessions")
	} else {
		descSession(session, argv, cc, env, true, false, false, false)
	}
	return currCmdIdx, nil
}

func LastSessionDescMore(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	session, ok := getLastSession(cc, env, true, true, true)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no executed/running sessions")
	} else {
		descSession(session, argv, cc, env, true, false, true, false)
	}
	return currCmdIdx, nil
}

func LastSessionDescFull(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	session, ok := getLastSession(cc, env, true, true, true)
	if !ok {
		display.PrintTipTitle(cc.Screen, env, "no executed/running sessions")
	} else {
		descSession(session, argv, cc, env, false, true, true, false)
	}
	return currCmdIdx, nil
}

func SessionRetry(argv model.ArgVals, cc *model.Cli, env *model.Env) (flow []string, masks []*model.ExecuteMask, err error) {
	id := argv.GetRaw("session-id")
	if len(id) == 0 {
		return nil, nil, fmt.Errorf("[SessionRetry] arg 'session-id' is empty")
	}
	sessions, _ := findSessions(nil, id, cc, env, 1, true, true, true)
	if len(sessions) == 0 {
		return
	}
	if len(sessions) > 1 {
		return nil, nil, fmt.Errorf("[SessionRetry] should never happen")
	}
	return retrySession(sessions[0], false)
}

func LastSessionRetry(argv model.ArgVals, cc *model.Cli, env *model.Env) (flow []string, masks []*model.ExecuteMask, err error) {
	session, ok := getLastSession(cc, env, true, true, true)
	if !ok {
		return nil, nil, fmt.Errorf("no executed sessions")
	}
	return retrySession(session, false)
}

func LastErrorSessionRetry(argv model.ArgVals, cc *model.Cli, env *model.Env) (flow []string, masks []*model.ExecuteMask, err error) {
	session, ok := getLastSession(cc, env, true, false, false)
	if !ok {
		return nil, nil, fmt.Errorf("no executed sessions")
	}
	return retrySession(session, false)
}

func findSessions(
	findStrs []string,
	id string,
	cc *model.Cli,
	env *model.Env,
	cntLimit int,
	includeError bool,
	includeDone bool,
	includeRunning bool) (sessions []model.SessionStatus, total int) {

	id = normalizeSid(id)

	sessions, total = model.ListSessions(env, findStrs, id, cntLimit, includeError, includeDone, includeRunning)
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

func dumpSession(session model.SessionStatus, env *model.Env, screen model.Screen, status string) {
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
	screen.Print(fmt.Sprintf("        %s\n", session.StartTs.Format(model.SessionTimeFormat)))
	screen.Print(fmt.Sprintf("        "+display.ColorExplain("%s ago\n", env),
		time.Now().Sub(session.StartTs).Round(time.Second).String()))

	if !session.Status.FinishTs.IsZero() {
		if !session.Running {
			screen.Print(display.ColorProp("    finish-at:\n", env))
			screen.Print(fmt.Sprintf("        %s\n", session.Status.FinishTs.Format(model.SessionTimeFormat)))
			incompletedDurStr := ""
			if !session.Running && session.Status.Result == model.ExecutedResultIncompleted {
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
	} else if session.Status.Result == model.ExecutedResultSucceeded {
		screen.Print("        " + display.ColorCmdDone(string(session.Status.Result), env) + "\n")
	} else if session.Status.Result == model.ExecutedResultError {
		screen.Print("        " + display.ColorError(string(session.Status.Result), env) + "\n")
	} else if session.Status.Result == model.ExecutedResultIncompleted {
		screen.Print("        " + display.ColorWarn("failed\n", env))
	} else {
		screen.Print("        " + string(session.Status.Result) + "\n")
	}
}

func descSession(session model.SessionStatus, argv model.ArgVals, cc *model.Cli, env *model.Env,
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

	flow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, model.FlowStrToStrs(session.Status.Flow)...)
	display.DumpFlowEx(cc, env, flow, 0, dumpArgs, session.Status, session.Running, EnvOpCmds())
}

func getLastSession(cc *model.Cli, env *model.Env, includeError bool, includeDone bool,
	includeRunning bool) (session model.SessionStatus, ok bool) {

	var sessions []model.SessionStatus
	currSession := env.GetRaw("sys.session.id")
	cntLimit := 16
	if len(currSession) == 0 {
		cntLimit = 1
	}
	sessions, _ = model.ListSessions(env, nil, "", cntLimit, includeError, includeDone, includeRunning)
	for i := len(sessions) - 1; i >= 0; i-- {
		session := sessions[i]
		if session.Status == nil {
			continue
		}
		if len(session.Status.Cmds) == 0 {
			continue
		}
		if session.SessionId() == currSession {
			continue
		}
		if isSessionShouldHideInLast(cc, session) {
			continue
		}
		return session, true
	}
	return
}

func isSessionShouldHideInLast(cc *model.Cli, session model.SessionStatus) bool {
	if session.Status == nil {
		return true
	}
	lastCmdStr := session.Status.Cmds[len(session.Status.Cmds)-1].Cmd
	if parsedCmd, ok := cc.ParseCmd(false, model.FlowStrToStrs(lastCmdStr)...); ok {
		if parsedCmd.LastCmd() != nil && parsedCmd.LastCmd().IsHideInSessionsLast() {
			return true
		}
	}
	return false
}

func retrySession(session model.SessionStatus, errSessionOnly bool) (flow []string, masks []*model.ExecuteMask, err error) {
	if errSessionOnly && session.Status != nil && session.Status.Result == model.ExecutedResultSucceeded {
		return nil, nil, fmt.Errorf("session [%s] had succeeded, nothing to retry", session.DirName)
	}
	return []string{session.Status.Flow}, session.Status.GenExecMasks(), nil
}

func normalizeSid(id string) string {
	// TODO: put brackets to env?
	return strings.TrimRight(strings.TrimLeft(id, "["), "]")
}
