package builtin

import (
	"fmt"
	"path/filepath"

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

	sessions := core.ListSessions(env)
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

	sessions := core.ListSessions(env)
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
	env.SetDur("sys.session.keep-status-duration", argv.GetRaw("duration"))
	return currCmdIdx, true
}

func dumpSession(session core.SessionStatus, env *core.Env, screen core.Screen) {
	sessionsRoot := env.GetRaw("sys.paths.sessions")
	if len(sessionsRoot) == 0 {
		panic(fmt.Errorf("[dumpSession] can't get sessions' root path"))
	}

	statusFileName := env.GetRaw("strs.session-status-file")
	selfName := env.GetRaw("strs.self-name")

	sessionDir := filepath.Join(sessionsRoot, session.DirName)
	statusPath := filepath.Join(sessionDir, statusFileName)
	status := core.LoadSessionStatus(statusPath, env)
	screen.Print(display.ColorSession("["+session.DirName+"]\n", env))
	screen.Print(display.ColorProp("    cmd:\n", env))
	screen.Print(display.ColorFlow(fmt.Sprintf("        %s %s\n", selfName, status), env))
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
