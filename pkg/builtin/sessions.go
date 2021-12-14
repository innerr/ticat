package builtin

import (
	"fmt"
	"os"
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
		descSession(sessions[len(sessions)-1], argv, cc, env, skeleton, showEnvFull, showModifiedEnv)
		prefix := fmt.Sprintf("more than one sessions(%v), only display the last one, ", len(sessions))
		findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
		if len(findStrs) > 0 {
			display.PrintErrTitle(cc.Screen, env, prefix+"add more find-strs to filter, or specify the arg 'id'")
		} else {
			display.PrintErrTitle(cc.Screen, env, prefix+"pass find-strs to filter, or specify the arg 'id'")
		}
	}

	allSessions := sessions
	if len(sessions) > 1 {
		sessions = filterRetrySessions(sessions)
	}
	if len(sessions) > 1 {
		handleTooMany()
	} else if len(sessions) != 0 {
		descSession(sessions[0], argv, cc, env, skeleton, showEnvFull, showModifiedEnv)
	} else if len(allSessions) != len(sessions) {
		handleTooMany()
	}
	return currCmdIdx, true
}

func filterRetrySessions(sessions []core.SessionStatus) (res []core.SessionStatus) {
	for _, session := range sessions {
		if !session.Status.IsRetry {
			res = append(res, session)
		}
	}
	return
}

func findSessionsByStrsAndId(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (sessions []core.SessionStatus) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)
	id := argv.GetRaw("session-id")
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

func LastSession(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

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

	return lastSessionDesc(argv, cc, env, flow, currCmdIdx, true, false, false)
}

func LastSessionDescMore(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return lastSessionDesc(argv, cc, env, flow, currCmdIdx, true, false, true)
}

func LastSessionDescFull(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

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

func ListSessionRetry(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	sessions := findSessionsByStrsAndId(argv, cc, env, flow, currCmdIdx)
	if len(sessions) == 0 {
		return currCmdIdx, true
	}

	sessions = filterRetrySessions(sessions)
	if len(sessions) == 0 {
		display.PrintErrTitle(cc.Screen, env, "matched sessions are all retry-session, can't re-retry")
		return currCmdIdx, true
	}

	if len(sessions) > 1 {
		panic(fmt.Errorf("[ListSessionRetry] should never happen"))
	} else {
		return currCmdIdx, retrySession(sessions[0], argv, cc, env, flow, currCmdIdx)
	}
	return currCmdIdx, true
}

func SessionRetry(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	id := getAndCheckArg(argv, flow.Cmds[currCmdIdx], "session-id")
	id = strings.TrimRight(strings.TrimLeft(id, "["), "]")
	sessions := core.ListSessions(env, nil, id)
	if len(sessions) == 0 {
		display.PrintErrTitle(cc.Screen, env, "no session with id = '"+id+"'")
		return currCmdIdx, true
	} else if len(sessions) > 1 {
		panic(fmt.Errorf("[SessionRetry] should never happen"))
	}

	session := sessions[0]
	if session.Status.IsRetry {
		display.PrintErrTitle(cc.Screen, env, "no session with id = '"+id+"', is retry-session, can't re-retry")
		return currCmdIdx, true
	}

	return currCmdIdx, retrySession(session, argv, cc, env, flow, currCmdIdx)
}

func LastSessionRetry(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	sessions := core.ListSessions(env, nil, "")
	if len(sessions) == 0 {
		display.PrintTipTitle(cc.Screen, env, "no executed sessions")
		return currCmdIdx, true
	}

	sessions = filterRetrySessions(sessions)
	if len(sessions) == 0 {
		display.PrintErrTitle(cc.Screen, env, "all sessions are all retry-session, can't re-retry")
		return currCmdIdx, true
	}

	ok := retrySession(sessions[len(sessions)-1], argv, cc, env, flow, currCmdIdx)
	return currCmdIdx, ok
}

func retrySession(
	session core.SessionStatus,
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (succeeded bool) {

	/*
		if session.Status != nil && session.Status.Executed {
			cc.Screen.Print("last session succeeded, nothing to retry\n")
			return true
		}
	*/

	// TODO: forbit retry retry-sesion

	sessionFlow := core.FlowStrToStrs(session.Status.Flow)
	flowEnv := env.NewLayer(core.EnvLayerTmp)

	if cc.FlowStatus != nil {
		cc.FlowStatus.OnSubFlowStart(core.FlowStrsToStr(sessionFlow))
		defer func() {
			if succeeded {
				cc.FlowStatus.OnSubFlowFinish(flowEnv)
			}
			// Avoid double recover
			if r := recover(); r != nil {
				display.PrintError(cc, env, r.(error))
				os.Exit(-1)
			}
		}()
	}
	succeeded = cc.Executor.Execute("retry:"+session.DirName, cc, flowEnv, sessionFlow...)
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

	dumpArgs := display.NewDumpFlowArgs().SetMaxDepth(argv.GetInt("depth")).SetMaxTrivial(argv.GetInt("trivial"))
	if skeleton {
		if !showEnvFull && !showModifiedEnv {
			dumpArgs.SetSkeleton()
		} else {
			//dumpArgs.SetSimple()
			dumpArgs.SetSkeleton()
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
