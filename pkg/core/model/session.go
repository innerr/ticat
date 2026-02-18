package model

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/innerr/ticat/pkg/utils"
)

type SessionStatus struct {
	DirName  string
	Pid      int
	StartTs  time.Time
	Running  bool
	Cleaning bool
	Status   *ExecutedFlow
}

func (self SessionStatus) SessionId() string {
	return self.DirName
}

// TODO: use iterator to avoid extra parsing?
func ListSessions(
	env *Env,
	findStrs []string,
	mustMatchDirName string,
	cntLimit int,
	includeError bool,
	includeDone bool,
	includeRunning bool) (sessions []SessionStatus, total int) {

	sessionsRoot := env.GetRaw("sys.paths.sessions")
	if len(sessionsRoot) == 0 {
		// PANIC: Programming error - sessions root path not configured
		panic(fmt.Errorf("[ListSessions] can't get sessions' root path\n"))
	}

	entrys, err := os.ReadDir(sessionsRoot)
	if err != nil {
		// PANIC: Runtime error - cannot read sessions directory
		panic(fmt.Sprintf("[ListSessions] can't read sessions' root path '%s'\n", sessionsRoot))
	}
	dirs := make([]string, len(entrys))
	for i, it := range entrys {
		dirs[i] = it.Name()
	}
	total = len(dirs)
	sort.Sort(sort.Reverse(sort.StringSlice(dirs)))

	keepDur := env.GetDur("sys.sessions.keep-status-duration")
	statusFileName := env.GetRaw("strs.session-status-file")

	now := time.Now()
	cnt := 0

	for _, dir := range dirs {
		oldSessionPid, oldSessionStartTs, ok := parseSessionDirName(dir)
		if !ok {
			continue
		}

		if len(mustMatchDirName) != 0 && dir != mustMatchDirName {
			continue
		}

		status := SafeParseExecutedFlow(ExecutedStatusFilePath{sessionsRoot, dir, statusFileName})
		// TODO: better error handling
		if status == nil {
			continue
		}

		if !status.MatchFind(findStrs) {
			continue
		}

		running := utils.IsPidRunning(oldSessionPid)
		if running && !includeRunning {
			continue
		}

		if !running {
			if status.Result == ExecutedResultSucceeded && !includeDone {
				continue
			}
			if (status.Result == ExecutedResultIncompleted || status.Result == ExecutedResultError) && !includeError {
				continue
			}
		}

		if running && (status.StartTs == status.FinishTs) {
			status.FinishTs = time.Now()
		}

		cleaning := oldSessionStartTs.Add(keepDur).Before(now)
		session := SessionStatus{dir, oldSessionPid, oldSessionStartTs, running, cleaning, status}
		sessions = append([]SessionStatus{session}, sessions...)
		cnt += 1
		if cntLimit > 0 && cnt >= cntLimit {
			break
		}
	}

	return
}

func CleanSessions(env *Env) (cleaned uint, runnings uint) {
	sessionsRoot := env.GetRaw("sys.paths.sessions")
	if len(sessionsRoot) == 0 {
		// PANIC: Programming error - sessions root path not configured
		panic(fmt.Errorf("[ListSessions] can't get sessions' root path\n"))
	}

	dirs, err := os.ReadDir(sessionsRoot)
	if err != nil {
		// PANIC: Runtime error - cannot read sessions directory
		panic(fmt.Sprintf("[ListSessions] can't read sessions' root path '%s'\n", sessionsRoot))
	}

	for _, dir := range dirs {
		oldSessionPid, _, ok := parseSessionDirName(dir.Name())
		if !ok {
			continue
		}
		running := utils.IsPidRunning(oldSessionPid)
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

func CleanSession(session SessionStatus, env *Env, force bool) (cleaned bool, running bool) {
	sessionsRoot := env.GetRaw("sys.paths.sessions")
	if len(sessionsRoot) == 0 {
		// PANIC: Programming error - sessions root path not configured
		panic(fmt.Errorf("[CleanSession] can't get sessions' root path"))
	}

	running = utils.IsPidRunning(session.Pid)
	if running && !force {
		return false, running
	}

	sessionDir := filepath.Join(sessionsRoot, session.DirName)
	err := os.RemoveAll(sessionDir)
	if err != nil {
		//panic(fmt.Sprintf("[CleanSession] can't remove session dir '%s'", sessionDir))
		return false, running
	}
	return true, running
}

func SessionSetId(env *Env) {
	_, _, id := GenSessionId()
	env = env.GetLayer(EnvLayerSession)
	env.Set("sys.session.id", id)
	env.Set("sys.session.id.full", id+"@"+env.GetRaw("sys.session.id.ip"))
}

func SessionInit(cc *Cli, flow *ParsedCmds, env *Env, sessionFileName string,
	statusFileName string) (flowStatus *ExecutingFlow, crossProcessInnerCall bool, ok bool) {

	sessionsRoot := env.GetRaw("sys.paths.sessions")
	if len(sessionsRoot) == 0 {
		_ = cc.Screen.Print("[sessionInit] can't get sessions' root path\n")
		return nil, false, false
	}

	sessionDir := env.GetRaw("session")

	sessionPath := filepath.Join(sessionDir, sessionFileName)
	if len(sessionDir) != 0 {
		_ = LoadEnvFromFile(env, sessionPath, cc.Cmds.Strs.EnvKeyValSep, cc.Cmds.Strs.EnvValDelAllMark)
		// NOTE: treat recursive ticat call as non-ticat scripts, not record the executing status
		//statusPath := filepath.Join(sessionDir, statusFileName)
		//return NewExecutingFlow(statusPath, flow, env), true
		return nil, true, true
	}

	keepDur := env.GetDur("sys.sessions.keep-status-duration")

	if err := os.MkdirAll(sessionsRoot, os.ModePerm); err != nil {
		_ = cc.Screen.Print(fmt.Sprintf("[sessionInit] can't create sessions root path '%s': %v\n",
			sessionsRoot, err))
		return nil, false, false
	}
	dirs, err := os.ReadDir(sessionsRoot)
	if err != nil {
		_ = cc.Screen.Print(fmt.Sprintf("[sessionInit] can't read sessions' root path '%s'\n",
			sessionsRoot))
		return nil, false, false
	}

	_, now, id := GenSessionId()
	dirName := id

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
			if err := os.RemoveAll(filepath.Join(sessionsRoot, dir.Name())); err != nil {
				_ = cc.Screen.Print(fmt.Sprintf("[sessionInit] can't remove old session '%s': %v\n",
					dir.Name(), err))
			}
		}
	}

	sessionDir = filepath.Join(sessionsRoot, dirName)
	err = os.MkdirAll(sessionDir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		_ = cc.Screen.Print(fmt.Sprintf("[sessionInit] can't create session dir '%s'\n",
			sessionDir))
		return nil, false, false
	}

	sessionEnv := env.GetLayer(EnvLayerSession)
	sessionEnv.Set("session", sessionDir)
	sessionEnv.Set("sys.session.id", id)
	sessionEnv.Set("sys.session.id.full", id+"@"+env.GetRaw("sys.session.id.ip"))

	statusPath := filepath.Join(sessionDir, statusFileName)
	return NewExecutingFlow(statusPath, flow, env), false, true
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
*/

func GenSessionId() (pid int, now time.Time, dirName string) {
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
const SessionTimeShortFormat = "02-15:04:05"
