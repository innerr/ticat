package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type SessionStatus struct {
	DirName  string
	Pid      int
	StartTs  time.Time
	Running  bool
	Cleaning bool
	Status   *ExecutedFlow
}

// TODO: use iterator to avoid extra parsing
func ListSessions(env *Env, findStrs []string, mustMatchDirName string) (sessions []SessionStatus) {
	sessionsRoot := env.GetRaw("sys.paths.sessions")
	if len(sessionsRoot) == 0 {
		panic(fmt.Errorf("[ListSessions] can't get sessions' root path\n"))
	}

	entrys, err := os.ReadDir(sessionsRoot)
	if err != nil {
		panic(fmt.Sprintf("[ListSessions] can't read sessions' root path '%s'\n", sessionsRoot))
	}
	dirs := make([]string, len(entrys))
	for i, it := range entrys {
		dirs[i] = it.Name()
	}
	sort.Strings(dirs)

	keepDur := env.GetDur("sys.session.keep-status-duration")
	statusFileName := env.GetRaw("strs.session-status-file")

	now := time.Now()

	for _, dir := range dirs {
		oldSessionPid, oldSessionStartTs, ok := parseSessionDirName(dir)
		if !ok {
			continue
		}

		if len(mustMatchDirName) != 0 && dir != mustMatchDirName {
			continue
		}

		status := ParseExecutedFlow(ExecutedStatusFilePath{sessionsRoot, dir, statusFileName})
		if !status.MatchFind(findStrs) {
			continue
		}
		err = syscall.Kill(oldSessionPid, syscall.Signal(0))
		running := !(err != nil && err == syscall.ESRCH)
		cleaning := oldSessionStartTs.Add(keepDur).Before(now)
		sessions = append(sessions, SessionStatus{dir, oldSessionPid, oldSessionStartTs, running, cleaning, status})
	}

	return
}

func CleanSessions(env *Env) (cleaned uint, runnings uint) {
	sessionsRoot := env.GetRaw("sys.paths.sessions")
	if len(sessionsRoot) == 0 {
		panic(fmt.Errorf("[ListSessions] can't get sessions' root path\n"))
	}

	dirs, err := os.ReadDir(sessionsRoot)
	if err != nil {
		panic(fmt.Sprintf("[ListSessions] can't read sessions' root path '%s'\n", sessionsRoot))
	}

	for _, dir := range dirs {
		oldSessionPid, _, ok := parseSessionDirName(dir.Name())
		if !ok {
			continue
		}
		err = syscall.Kill(oldSessionPid, syscall.Signal(0))
		running := !(err != nil && err == syscall.ESRCH)
		if !running {
			err = os.RemoveAll(filepath.Join(sessionsRoot, dir.Name()))
			if err == nil {
				cleaned += 1
			}
		} else {
			runnings += 1
		}
	}
	return
}

func CleanSession(session SessionStatus, env *Env) (cleaned bool, running bool) {
	sessionsRoot := env.GetRaw("sys.paths.sessions")
	if len(sessionsRoot) == 0 {
		panic(fmt.Errorf("[CleanSession] can't get sessions' root path"))
	}

	err := syscall.Kill(session.Pid, syscall.Signal(0))
	running = !(err != nil && err == syscall.ESRCH)
	if running {
		return false, true
	}

	sessionDir := filepath.Join(sessionsRoot, session.DirName)
	err = os.RemoveAll(sessionDir)
	if err != nil {
		//panic(fmt.Sprintf("[CleanSession] can't remove session dir '%s'", sessionDir))
		return false, false
	}
	return true, false
}

func SessionInit(cc *Cli, flow *ParsedCmds, env *Env, sessionFileName string,
	statusFileName string) (flowStatus *ExecutingFlow, ok bool) {

	sessionsRoot := env.GetRaw("sys.paths.sessions")
	if len(sessionsRoot) == 0 {
		cc.Screen.Print("[sessionInit] can't get sessions' root path\n")
		return nil, false
	}

	sessionDir := env.GetRaw("session")

	sessionPath := filepath.Join(sessionDir, sessionFileName)
	if len(sessionDir) != 0 {
		LoadEnvFromFile(env, sessionPath, cc.Cmds.Strs.EnvKeyValSep)
		// NOTE: treat recursive ticat call as non-ticat scripts, not record the executing status
		//statusPath := filepath.Join(sessionDir, statusFileName)
		//return NewExecutingFlow(statusPath, flow, env), true
		return nil, true
	}

	keepDur := env.GetDur("sys.session.keep-status-duration")

	os.MkdirAll(sessionsRoot, os.ModePerm)
	dirs, err := os.ReadDir(sessionsRoot)
	if err != nil {
		cc.Screen.Print(fmt.Sprintf("[sessionInit] can't read sessions' root path '%s'\n",
			sessionsRoot))
		return nil, false
	}

	_, now, dirName := genSessionDirName()

	// TODO: move this cleaning code to another function
	for _, dir := range dirs {
		oldSessionPid, oldSessionStartTs, ok := parseSessionDirName(dir.Name())
		if !ok {
			// TODO: print a warning msg
			continue
		}
		err = syscall.Kill(oldSessionPid, syscall.Signal(0))
		if !(err != nil && err == syscall.ESRCH) {
			continue
		}
		added := oldSessionStartTs.Add(keepDur)
		if added.Before(now) {
			os.RemoveAll(filepath.Join(sessionsRoot, dir.Name()))
		}
	}

	sessionDir = filepath.Join(sessionsRoot, dirName)
	err = os.MkdirAll(sessionDir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		cc.Screen.Print(fmt.Sprintf("[sessionInit] can't create session dir '%s'\n",
			sessionDir))
		return nil, false
	}

	env.GetLayer(EnvLayerSession).Set("session", sessionDir)

	statusPath := filepath.Join(sessionDir, statusFileName)
	return NewExecutingFlow(statusPath, flow, env), true
}

// TODO: clean it
// Seems not very useful, no user now.
/*
func SessionFinish(cc *Cli, flow *ParsedCmds, env *Env, sessionFileName string) bool {
	sessionDir := env.GetRaw("session")
	if len(sessionDir) == 0 {
		return true
	}
	path := filepath.Join(sessionDir, sessionFileName)
	SaveEnvToFile(env, path, cc.Cmds.Strs.EnvKeyValSep)
	return true
}
*/

func isNoSessionCmd(flow *ParsedCmds, noSessionCmds []interface{}) bool {
	if len(flow.Cmds) != 1 {
		return false
	}
	cmd := flow.Cmds[0].LastCmd()
	for _, it := range noSessionCmds {
		if cmd.IsTheSameFunc(it) {
			return true
		}
	}
	return false
}

func genSessionDirName() (pid int, now time.Time, dirName string) {
	pid = os.Getpid()
	now = time.Now()
	return pid, now, fmt.Sprintf("%s.%d", now.Format(SessionDirTimeFormat), os.Getpid())
}

func parseSessionDirName(dirName string) (pid int, startTs time.Time, ok bool) {
	fields := strings.Split(dirName, ".")
	if len(fields) != 2 {
		return
	}
	startTs, err := time.ParseInLocation(SessionDirTimeFormat, fields[0], time.Local)
	if err != nil {
		return
	}
	pid, err = strconv.Atoi(fields[1])
	if err != nil {
		return
	}
	return pid, startTs, true
}

const SessionDirTimeFormat = "20060102-150405"
const SessionTimeFormat = "2006-01-02 15:04:05"
