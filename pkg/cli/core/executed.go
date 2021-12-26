package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ExecutedResult string

const (
	ExecutedResultSucceeded   ExecutedResult = "OK"
	ExecutedResultError       ExecutedResult = "ERR"
	ExecutedResultSkipped     ExecutedResult = "skipped"
	ExecutedResultIncompleted ExecutedResult = "incompleted"
	ExecutedResultUnRun       ExecutedResult = "un-run"
)

type ExecutedStatusFilePath struct {
	RootPath string
	DirName  string
	FileName string
}

type ExecutedCmd struct {
	Cmd         string
	LogFilePath string
	IsDelay     bool
	StartEnv    *Env
	SubFlow     *ExecutedFlow
	FinishEnv   *Env
	StartTs     time.Time
	FinishTs    time.Time
	Result      ExecutedResult
	ErrStrs     []string
}

func NewExecutedCmd(cmd string) *ExecutedCmd {
	return &ExecutedCmd{Cmd: cmd, Result: ExecutedResultIncompleted}
}

type ExecutedFlow struct {
	Flow     string
	DirName  string
	Cmds     []*ExecutedCmd
	StartTs  time.Time
	FinishTs time.Time
	// Result should never be skipped here
	Result ExecutedResult
}

func NewExecutedFlow(dirName string) *ExecutedFlow {
	return &ExecutedFlow{DirName: dirName, Result: ExecutedResultIncompleted}
}

func (self ExecutedStatusFilePath) Full() string {
	return filepath.Join(self.RootPath, self.DirName, self.FileName)
}

func (self ExecutedStatusFilePath) Short() string {
	return filepath.Join("...", self.DirName, self.FileName)
}

func ParseExecutedFlow(path ExecutedStatusFilePath) *ExecutedFlow {
	file, err := os.Open(path.Full())
	if err != nil {
		panic(fmt.Errorf("[ParseExecutedFlow] open executed status file '%s' failed: %v", path.Short(), err))
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(fmt.Errorf("[ParseExecutedFlow] read executed status file '%s' failed: %v", path.Short(), err))
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Consider it's normal if ok=false but all lines are parsed
	executed, lines, ok := parseExecutedFlow(path, lines, 0)
	if !ok && len(lines) != 0 {
		sample := getSampleLines(lines)
		panic(fmt.Errorf("[ParseExecutedFlow] bad executed status file '%s', has %v lines unparsed: %s",
			path.Short(), len(lines), sample))
	}
	return executed
}

func (self *ExecutedFlow) GenExecMasks() (masks []*ExecuteMask) {
	for _, cmd := range self.Cmds {
		var subMasks []*ExecuteMask
		if cmd.SubFlow != nil {
			subMasks = cmd.SubFlow.GenExecMasks()
		}
		policy := ExecPolicyExec
		if cmd.Result == ExecutedResultSucceeded {
			policy = ExecPolicySkip
		}
		masks = append(masks, &ExecuteMask{cmd.Cmd, cmd.StartEnv, policy, subMasks})
	}
	return
}

func (self *ExecutedFlow) IsOneCmdSession(cmd string) bool {
	return len(self.Cmds) == 1 && self.Cmds[0].Cmd == cmd
}

func (self *ExecutedFlow) CalResultWhenIncompleted() {
	if self.Result != ExecutedResultIncompleted {
		return
	}
	for _, cmd := range self.Cmds {
		if cmd.Result == ExecutedResultError {
			self.Result = cmd.Result
			break
		}
	}
}

func (self *ExecutedFlow) MatchFind(findStrs []string) bool {
	if len(findStrs) == 0 {
		return true
	}
	for _, it := range findStrs {
		if strings.Index(self.DirName, it) < 0 && strings.Index(self.Flow, it) < 0 {
			return false
		}
	}
	return true
}

func (self *ExecutedFlow) GetCmd(idx int) *ExecutedCmd {
	if idx >= len(self.Cmds) {
		return nil
	}
	cmd := self.Cmds[idx]
	if !cmd.IsDelay {
		return cmd
	}
	if len(cmd.SubFlow.Cmds) == 0 {
		return nil
	}
	return cmd.SubFlow.Cmds[0]
}

func parseExecutedFlow(path ExecutedStatusFilePath, lines []string,
	level int) (executed *ExecutedFlow, remain []string, ok bool) {

	executed, remain, ok = tryParseExecutedFlow(path, lines, level)
	executed.CalResultWhenIncompleted()
	return
}

func tryParseExecutedFlow(path ExecutedStatusFilePath, lines []string,
	level int) (executed *ExecutedFlow, remain []string, ok bool) {

	executed = NewExecutedFlow(path.DirName)

	executed.Flow, lines, ok = parseMarkedOneLineContent(path, lines, "flow", level)
	if !ok {
		return executed, lines, false
	}
	executed.StartTs, lines, ok = parseMarkedTime(path, lines, "flow-start-time", level)
	if !ok {
		return executed, lines, false
	}
	executed.Cmds, lines, ok = parseExecutedCmds(path, lines, level)
	if !ok {
		return executed, lines, false
	}
	executed.FinishTs, lines, ok = parseMarkedTime(path, lines, "flow-finish-time", level)
	if !ok {
		return executed, lines, false
	}
	executed.Result, lines, ok = parseCmdOrFlowResult(path, lines, "flow-result", level, ExecutedResultIncompleted)
	return executed, lines, ok
}

func parseExecutedCmds(path ExecutedStatusFilePath, lines []string, level int) (cmds []*ExecutedCmd, remain []string, ok bool) {
	for len(lines) != 0 {
		var cmd *ExecutedCmd
		cmd, lines, ok = parseExecutedCmd(path, lines, level)
		if !ok {
			if _, _, ok := parseMarkedOneLineContent(path, lines, "flow-finish-time", level); ok {
				break
			}
			if len(lines) != 0 {
				return cmds, lines, false
			}
		}
		cmds = append(cmds, cmd)
	}
	return cmds, lines, true
}

func parseExecutedCmd(path ExecutedStatusFilePath, lines []string, level int) (cmd *ExecutedCmd, remain []string, ok bool) {
	cmdStr, lines, ok := parseMarkedOneLineContent(path, lines, "cmd", level)
	if !ok {
		return nil, lines, false
	}
	cmd = NewExecutedCmd(strings.TrimSpace(cmdStr))

	cmd.StartTs, lines, ok = parseMarkedTime(path, lines, "cmd-start-time", level)
	if !ok {
		return cmd, lines, false
	}

	logFilePath, lines, ok := parseMarkedOneLineContent(path, lines, "log", level)
	if ok {
		cmd.LogFilePath = strings.TrimSpace(logFilePath)
	}

	lines, ok = tryParseScheduledCmd(cmd, path, lines, level)
	if ok {
		return cmd, lines, true
	}

	startEnvLines, lines, ok := parseMarkedContent(path, lines, "env-start", level)
	if ok {
		cmd.StartEnv = parseEnvLines(path, startEnvLines, level)
	}

	subflowLines, lines, ok := parseMarkedContent(path, lines, "subflow", level)
	if len(subflowLines) > 0 {
		if !ok && len(lines) != 0 {
			return cmd, lines, false
		}
		cmd.SubFlow, subflowLines, ok = parseExecutedFlow(path, subflowLines, level+1)
		if !ok {
			if len(subflowLines) != 0 {
				sample := getSampleLines(subflowLines)
				panic(fmt.Errorf("[ParseExecutedFlow] bad subflow lines in status file '%s', has %v lines unparsed: %s",
					path.Short(), len(subflowLines), sample))
			}
			return cmd, lines, false
		}
	}

	finishEnvLines, lines, ok := parseMarkedContent(path, lines, "env-finish", level)
	if ok {
		cmd.FinishEnv = parseEnvLines(path, finishEnvLines, level)
	}

	cmd.FinishTs, lines, ok = parseMarkedTime(path, lines, "cmd-finish-time", level)
	if !ok {
		return cmd, lines, false
	}

	cmd.Result, lines, ok = parseCmdOrFlowResult(path, lines, "cmd-result", level, ExecutedResultError)
	if !ok {
		return cmd, lines, false
	}

	errLines, lines, ok := parseMarkedContent(path, lines, "error", level)
	if ok {
		for _, line := range errLines {
			cmd.ErrStrs = append(cmd.ErrStrs, strings.TrimSpace(line))
		}
	}

	return cmd, lines, true
}

func tryParseScheduledCmd(cmd *ExecutedCmd, path ExecutedStatusFilePath, lines []string, level int) (remain []string, ok bool) {
	tid, lines, ok := parseMarkedOneLineContent(path, lines, "scheduled", level)
	if !ok {
		return lines, false
	}
	bgSessionPath := ExecutedStatusFilePath{path.RootPath, filepath.Join(path.DirName, tid), path.FileName}
	subflow := ParseExecutedFlow(bgSessionPath)
	if len(subflow.Cmds) == 0 {
		executedCmd := NewExecutedCmd(cmd.Cmd)
		executedCmd.Result = ExecutedResultIncompleted
		subflow.Cmds = append(subflow.Cmds, executedCmd)
	} else if len(subflow.Cmds) != 1 {
		panic(fmt.Errorf("[ParseExecutedFlow] expect only one cmd in delayed task"))
	}
	cmd.IsDelay = true
	cmd.SubFlow = subflow
	// The schedule-result is not important, use the execute-result
	cmd.Result = subflow.Result
	return lines, true
}

func parseCmdOrFlowResult(path ExecutedStatusFilePath, lines []string, mark string,
	level int, defVal ExecutedResult) (result ExecutedResult, remain []string, ok bool) {

	resultStr, lines, ok := parseMarkedOneLineContent(path, lines, mark, level)
	if !ok {
		return defVal, lines, false
	}
	return ExecutedResult(resultStr), lines, true
}

func parseEnvLines(path ExecutedStatusFilePath, lines []string, level int) (env *Env) {
	env = NewEnvEx(EnvLayerSession)
	indent := strings.Repeat(StatusFileIndent, level)
	for _, line := range lines {
		if !strings.HasPrefix(line, indent) {
			panic(fmt.Errorf("[ParseExecutedFlow] bad indent of env line '%s' in status file '%s'", line, path.Short()))
		}
		line = line[len(indent):]
		i := strings.Index(line, "=")
		if i <= 0 {
			panic(fmt.Errorf("[ParseExecutedFlow] bad format of env line '%s' in status file '%s'", line, path.Short()))
		}
		env.Set(line[0:i], line[i+1:])
	}
	return env
}

func parseMarkedTime(path ExecutedStatusFilePath, lines []string, mark string, level int) (ts time.Time, remain []string, ok bool) {
	var tsStr string
	tsStr, remain, ok = parseMarkedOneLineContent(path, lines, mark, level)
	if !ok {
		return
	}
	var err error
	ts, err = time.ParseInLocation(SessionTimeFormat, tsStr, time.Local)
	if err != nil {
		panic(fmt.Errorf("[ParseExecutedFlow] bad ts format '%s' with mark '%s' in status file '%s', err: %s",
			tsStr, mark, path.Short(), err))
	}
	return
}

func parseMarkedOneLineContent(path ExecutedStatusFilePath, lines []string,
	mark string, level int) (content string, remain []string, ok bool) {

	res, remain, ok := parseMarkedContent(path, lines, mark, level)
	if !ok {
		return
	}
	if len(res) != 1 {
		panic(fmt.Errorf("[ParseExecutedFlow] expect only one line for mark '%s' in status file '%s'", mark, path.Short()))
	}
	return res[0], remain, ok
}

func parseMarkedContent(path ExecutedStatusFilePath, lines []string,
	mark string, level int) (content []string, remain []string, ok bool) {

	if len(lines) == 0 {
		return nil, lines, false
	}

	markStart := markStartStr(mark, level)
	markFinish := markFinishStr(mark, 0)
	if strings.HasPrefix(lines[0], markStart) && strings.HasSuffix(lines[0], markFinish) {
		return []string{lines[0][len(markStart) : len(lines[0])-len(markFinish)]}, lines[1:], true
	}

	if lines[0] != markStart {
		return nil, lines, false
	}
	lines = lines[1:]

	depth := 0

	markFinish = markFinishStr(mark, level)
	for i, line := range lines {
		if markStart == line {
			depth += 1
		} else if markFinish == line {
			if depth == 0 {
				return lines[0:i], lines[i+1:], true
			} else {
				depth -= 1
				if depth < 0 {
					panic(fmt.Errorf("[ParseExecutedFlow] bad recusive mark '%s' in status file '%s'", mark, path.Short()))
				}
			}
		}
	}

	return lines, nil, false
}

func getSampleLines(lines []string) string {
	sample := []string{}
	for i, line := range lines {
		if i > 3 {
			break
		}
		sample = append(sample, strings.TrimSpace(line))
	}
	return strings.Join(sample, " ")
}
