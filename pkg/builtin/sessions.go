package builtin

import (
	"fmt"
	"strings"
	"time"

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
	if len(sessions) == 0 {
		if len(findStrs) > 0 {
			display.PrintErrTitle(cc.Screen, env, "no executed sessions found by '"+strings.Join(findStrs, " ")+"'")
		} else {
			display.PrintErrTitle(cc.Screen, env, "no executed sessions")
		}
	}
	for _, it := range sessions {
		if it.Cleaning {
			continue
		}
		dumpSession(it, env, cc.Screen)
	}
	return currCmdIdx, true
}

func DescListedSession(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)

	sessions := core.ListSessions(env, findStrs)
	if len(sessions) == 0 {
		if len(findStrs) > 0 {
			display.PrintErrTitle(cc.Screen, env, "no executed sessions found by '"+strings.Join(findStrs, " ")+"'")
		} else {
			display.PrintErrTitle(cc.Screen, env, "no executed sessions")
		}
	} else if len(sessions) > 1 {
		descSession(sessions[len(sessions)-1], argv, cc, env)
		prefix := fmt.Sprintf("more than one sessions(%v), only display the last one, ", len(sessions))
		if len(findStrs) > 0 {
			display.PrintErrTitle(cc.Screen, env, prefix+"add more find-str to filter")
		} else {
			display.PrintErrTitle(cc.Screen, env, prefix+"pass find-str to filter")
		}
	} else {
		descSession(sessions[0], argv, cc, env)
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

func DescLastSession(
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
	descSession(sessions[len(sessions)-1], argv, cc, env)
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

func descSession(session core.SessionStatus, argv core.ArgVals, cc *core.Cli, env *core.Env) {
	dumpArgs := display.NewDumpFlowArgs().SetSkeleton().
		SetMaxDepth(argv.GetInt("depth")).SetMaxTrivial(argv.GetInt("trivial"))
	flow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, core.FlowStrToStrs(session.Status.Flow)...)
	display.DumpFlowEx(cc, env, flow, 0, dumpArgs, session.Status, EnvOpCmds())
}
