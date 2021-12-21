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
	key := "sys.session.keep-status-duration"
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

func ListSessions(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	sessions := findSessionsByStrsAndId(argv, cc, env, flow, currCmdIdx)
	for _, it := range sessions {
		if it.Cleaning {
			dumpSession(it, env, cc.Screen, "expired")
		} else {
			dumpSession(it, env, cc.Screen, "")
		}
	}
	return currCmdIdx, true
}

func FindAndRemoveSessions(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	sessions := findSessionsByStrsAndId(argv, cc, env, flow, currCmdIdx)
	for _, it := range sessions {
		status := "removed"
		cleaned, running := core.CleanSession(it, env)
		if running {
			status = "running, untouched"
		} else if !cleaned {
			status = "remove failed"
		}
		dumpSession(it, env, cc.Screen, status)
	}
	return currCmdIdx, true
}

func ListedSessionDescLess(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return listedSessionDesc(argv, cc, env, flow, currCmdIdx, true, false, false)
}

func ListedSessionDescMore(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return listedSessionDesc(argv, cc, env, flow, currCmdIdx, true, false, true)
}

func ListedSessionDescFull(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return listedSessionDesc(argv, cc, env, flow, currCmdIdx, false, true, true)
}

func listedSessionDesc(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int,
	skeleton bool,
	showEnvFull bool,
	showModifiedEnv bool) (int, bool) {

	sessions := findSessionsByStrsAndId(argv, cc, env, flow, currCmdIdx)

	handleTooMany := func() {
		descSession(sessions[0], argv, cc, env, skeleton, showEnvFull, showModifiedEnv)
		prefix := fmt.Sprintf("more than one sessions(%v) matched, only display the first one, ", len(sessions))
		findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
		if len(findStrs) > 0 {
			display.PrintErrTitle(cc.Screen, env, prefix+"add more find-strs to filter, or specify arg 'id'")
		} else {
			display.PrintErrTitle(cc.Screen, env, prefix+"pass find-strs to filter, or specify arg 'id'")
		}
	}

	allSessions := sessions
	if len(sessions) > 1 {
		handleTooMany()
	} else if len(sessions) != 0 {
		descSession(sessions[0], argv, cc, env, skeleton, showEnvFull, showModifiedEnv)
	} else if len(allSessions) != len(sessions) {
		handleTooMany()
	}
	return currCmdIdx, true
}

func LastSession(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	sessions := core.ListSessions(env, nil, "")
	if len(sessions) == 0 {
		display.PrintTipTitle(cc.Screen, env, "no executed sessions")
		return currCmdIdx, true
	}
	dumpSession(sessions[len(sessions)-1], env, cc.Screen, "")
	return currCmdIdx, true
}

func LastSessionDescLess(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	return lastSessionDesc(argv, cc, env, flow, currCmdIdx, true, false, false)
}

func LastSessionDescMore(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	return lastSessionDesc(argv, cc, env, flow, currCmdIdx, true, false, true)
}

func LastSessionDescFull(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	return lastSessionDesc(argv, cc, env, flow, currCmdIdx, false, true, true)
}

func lastSessionDesc(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int,
	skeleton bool,
	showEnvFull bool,
	showModifiedEnv bool) (int, bool) {

	sessions := core.ListSessions(env, nil, "")
	if len(sessions) == 0 {
		display.PrintTipTitle(cc.Screen, env, "no executed sessions")
		return currCmdIdx, true
	}
	descSession(sessions[len(sessions)-1], argv, cc, env, skeleton, showEnvFull, showModifiedEnv)
	return currCmdIdx, true
}

func SessionRetry(argv core.ArgVals, cc *core.Cli, env *core.Env) (flow []string, masks []*core.ExecuteMask, ok bool) {
	id := argv.GetRaw("session-id")
	if len(id) == 0 {
		panic(fmt.Errorf("[SessionRetry] arg 'session-id' is empty"))
	}

	sessions := findSessions(nil, id, cc, env)
	if len(sessions) == 0 {
		return
	}
	if len(sessions) > 1 {
		panic(fmt.Errorf("[SessionRetry] should never happen"))
	}

	return retrySession(sessions[0])
}

/*
func LastSessionRetry(argv core.ArgVals, cc *core.Cli, env *core.Env) (flow []string, masks []*core.ExecuteMask, ok bool) {
	sessions := core.ListSessions(env, nil, "")
	// if there is one, it's this session itself
	if len(sessions) <= 1 {
		panic(fmt.Errorf("no executed sessions"))
	}

	sessions = filterRetryLastSessions(sessions)
	if len(sessions) == 0 {
		panic(fmt.Errorf("all sessions are all retry-session, can't re-retry, use 'sessions.list.retry' to force retry"))
	}

	return retrySession(sessions[len(sessions)-1])
}

func filterRetryLastSessions(sessions []core.SessionStatus) (res []core.SessionStatus) {
	// TODO: this is bad, this function should not know what it's registered cmd-path
	retryLast := "sessions.last.retry"
	retrySession := "sessions.list.retry"
	for _, session := range sessions {
		if !session.Status.IsOneCmdSession(retryLast) &&
			!session.Status.IsOneCmdSession(retrySession) {
			res = append(res, session)
		}
	}
	return
}
*/

func findSessionsByStrsAndId(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (sessions []core.SessionStatus) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
	id := argv.GetRaw("session-id")
	return findSessions(findStrs, id, cc, env)
}

func findSessions(findStrs []string, id string, cc *core.Cli, env *core.Env) (sessions []core.SessionStatus) {
	// TODO: put brackets to env?
	id = strings.TrimRight(strings.TrimLeft(id, "["), "]")

	sessions = core.ListSessions(env, findStrs, id)
	if len(sessions) == 0 {
		if len(id) == 0 {
			if len(findStrs) > 0 {
				display.PrintErrTitle(cc.Screen, env, "no executed sessions found by '"+strings.Join(findStrs, " ")+"'")
			} else {
				display.PrintErrTitle(cc.Screen, env, "no executed sessions")
			}
		} else {
			if len(findStrs) > 0 {
				display.PrintErrTitle(cc.Screen, env,
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

	screen.Print(display.ColorProp("    start-at:\n", env))
	screen.Print(fmt.Sprintf("        %s\n", session.StartTs.Format(core.SessionTimeFormat)))
	screen.Print(fmt.Sprintf("        "+display.ColorExplain("%s ago", env)+"\n",
		time.Now().Sub(session.StartTs).Round(time.Second).String()))

	if session.Status.Executed {
		screen.Print(display.ColorProp("    finish-at:\n", env))
		screen.Print(fmt.Sprintf("        %s\n", session.Status.FinishTs.Format(core.SessionTimeFormat)))
		screen.Print(fmt.Sprintf("        "+display.ColorExplain("%s ago, elapsed %s", env)+"\n",
			time.Now().Sub(session.Status.FinishTs).Round(time.Second),
			session.Status.FinishTs.Sub(session.StartTs).Round(time.Second)))
	}

	screen.Print(display.ColorProp("    status:\n", env))
	if session.Running {
		screen.Print("        " + display.ColorCmdCurr("running", env) + "\n")
		screen.Print(display.ColorProp("    pid:\n", env))
		screen.Print(fmt.Sprintf("        %v\n", session.Pid))
	} else if session.Status.Executed {
		screen.Print(display.ColorCmdDone("        done\n", env))
	} else {
		screen.Print(display.ColorError("        ERR\n", env))
	}
}

func descSession(session core.SessionStatus, argv core.ArgVals, cc *core.Cli, env *core.Env,
	skeleton, showEnvFull bool, showModifiedEnv bool) {

	dumpArgs := display.NewDumpFlowArgs().SetMaxDepth(argv.GetInt("depth")).SetMaxTrivial(argv.GetInt("unfold-trivial"))
	if skeleton {
		if !showEnvFull && !showModifiedEnv {
			dumpArgs.SetSkeleton()
		} else {
			dumpArgs.SetSimple()
			//dumpArgs.SetSkeleton()
		}
	}
	if showEnvFull {
		dumpArgs.SetShowExecutedEnvFull()
	}
	if showModifiedEnv {
		dumpArgs.SetShowExecutedModifiedEnv()
	}

	flow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, core.FlowStrToStrs(session.Status.Flow)...)
	display.DumpFlowEx(cc, env, flow, 0, dumpArgs, session.Status, session.Running, EnvOpCmds())
}

func retrySession(session core.SessionStatus) (flow []string, masks []*core.ExecuteMask, ok bool) {
	if session.Status.Executed {
		panic(fmt.Errorf("session [%s] had succeeded, nothing to retry", session.DirName))
	}
	return []string{session.Status.Flow}, session.Status.GenExecMasks(), true
}
