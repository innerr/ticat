package builtin

import (
	"fmt"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func CleanSessions(
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

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)

	sessions := core.ListSessions(env, findStrs)
	for _, it := range sessions {
		if it.Cleaning {
			continue
		}
		dumpSession(it, env, cc.Screen)
	}
	return currCmdIdx, true
}

func LastSession(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	sessions := core.ListSessions(env, nil)
	if len(sessions) == 0 {
		display.PrintTipTitle(cc.Screen, env, "no executed sessions")
		return currCmdIdx, true
	}
	dumpSession(sessions[len(sessions)-1], env, cc.Screen)
	return currCmdIdx, true
}

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

func dumpSession(session core.SessionStatus, env *core.Env, screen core.Screen) {
	selfName := env.GetRaw("strs.self-name")

	screen.Print(display.ColorSession("["+session.DirName+"]\n", env))
	screen.Print(display.ColorProp("    cmd:\n", env))
	screen.Print(display.ColorFlow(fmt.Sprintf("        %s %s\n", selfName, session.Status), env))
	screen.Print(display.ColorProp("    start-at:\n", env))
	screen.Print(fmt.Sprintf("        %v\n", session.StartTs.Format(core.SessionDirTimeFormat)))
	screen.Print(display.ColorProp("    status:\n", env))
	if session.Running {
		screen.Print("        running\n")
		screen.Print(display.ColorProp("    pid:\n", env))
		screen.Print(fmt.Sprintf("        %v\n", session.Pid))
	} else {
		screen.Print("        done\n")
	}
}
