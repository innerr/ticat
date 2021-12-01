package core

import (
	"fmt"
	"io/ioutil"
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
}

func ListSessions(env *Env) (sessions []SessionStatus) {
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

	now := time.Now()
	keepDur := env.GetDur("sys.session.keep-status-duration")

	for _, dir := range dirs {
		oldSessionPid, oldSessionStartTs, ok := parseSessionDirName(dir)
		if !ok {
			continue
		}
		err = syscall.Kill(oldSessionPid, syscall.Signal(0))
		running := !(err != nil && err == syscall.ESRCH)
		cleaning := oldSessionStartTs.Add(keepDur).Before(now)
		sessions = append(sessions, SessionStatus{dir, oldSessionPid, oldSessionStartTs, running, cleaning})
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
			os.RemoveAll(filepath.Join(sessionsRoot, dir.Name()))
			cleaned += 1
		} else {
			runnings += 1
		}
	}
	return
}

func SessionInit(cc *Cli, flow *ParsedCmds, env *Env, sessionFileName string, statusFileName string) bool {
	sessionDir := env.GetRaw("session")
	sessionPath := filepath.Join(sessionDir, sessionFileName)
	if len(sessionDir) != 0 {
		LoadEnvFromFile(env, sessionPath, cc.Cmds.Strs.EnvKeyValSep)
		return true
	}

	keepDur := env.GetDur("sys.session.keep-status-duration")

	sessionsRoot := env.GetRaw("sys.paths.sessions")
	if len(sessionsRoot) == 0 {
		cc.Screen.Print("[sessionInit] can't get sessions' root path\n")
		return false
	}

	os.MkdirAll(sessionsRoot, os.ModePerm)
	dirs, err := os.ReadDir(sessionsRoot)
	if err != nil {
		cc.Screen.Print(fmt.Sprintf("[sessionInit] can't read sessions' root path '%s'\n",
			sessionsRoot))
		return false
	}

	_, now, dirName := genSessionDirName()

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
		return false
	}

	env.GetLayer(EnvLayerSession).Set("session", sessionDir)

	statusPath := filepath.Join(sessionDir, statusFileName)
	SaveSessionStatus(statusPath, flow, env)
	return true
}

func SaveSessionStatus(statusPath string, flow *ParsedCmds, env *Env) {
	// TODO: save session full status to this file
	file, err := os.OpenFile(statusPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(fmt.Errorf("[SaveSessionStatus] open session status file '%s' failed: %v", statusPath, err))
	}
	defer file.Close()

	trivialMark := env.GetRaw("strs.trivial-mark")
	cmdPathSep := env.GetRaw("strs.cmd-path-sep")
	SaveFlow(file, flow, 0, cmdPathSep, trivialMark, env)
}

func LoadSessionStatus(statusPath string, env *Env) string {
	data, err := ioutil.ReadFile(statusPath)
	if err != nil {
		panic(fmt.Errorf("[LoadSessionStatus] open session status file '%s' failed: %v", statusPath, err))
	}
	return string(data)
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
