package model

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
	ExecutedResultSkipped     ExecutedResult = "skipped"
	ExecutedResultError       ExecutedResult = "ERR"
	ExecutedResultIncompleted ExecutedResult = "incompleted"
	ExecutedResultUnRun       ExecutedResult = "unrun"
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

func (self *ExecutedCmd) CalResultInCaseIncompleted() {
	if self.Result != ExecutedResultSucceeded && self.Result != ExecutedResultIncompleted {
		return
	}
	if self.SubFlow == nil {
		return
	}
	self.SubFlow.CalResultInCaseIncompleted()
	self.Result = self.SubFlow.Result
}

func (self *ExecutedCmd) RoughFinishTs(running bool) time.Time {
	finishTs := self.FinishTs
	if self.Result == ExecutedResultIncompleted && running {
		finishTs = time.Now().Round(time.Second)
	}
	return finishTs
}

func (self *ExecutedCmd) RoughDuration(running bool) time.Duration {
	return self.RoughFinishTs(running).Sub(self.StartTs)
}

type ExecutedFlow struct {
	Flow     string
	DirName  string
	Cmds     []*ExecutedCmd
	StartTs  time.Time
	FinishTs time.Time

	// Result should never be skipped here
	Result ExecutedResult

	Corrupted []string
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

func SafeParseExecutedFlow(path ExecutedStatusFilePath) (executed *ExecutedFlow) {
	defer func() {
		if r := recover(); r != nil {
			println(r.(error).Error())
		}
	}()
	_, executed = ParseExecutedFlow(path)
	return
}

func ParseExecutedFlow(path ExecutedStatusFilePath) (lastActiveTs time.Time, executed *ExecutedFlow) {
	file, err := os.Open(path.Full())
	if err != nil {
		panic(fmt.Errorf("[ParseExecutedFlow] open executed status file '%s' failed: %v", path.Short(), err))
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(fmt.Errorf("[ParseExecutedFlow] close status file '%s' failed: %v", path.Short(), err))
		}
	}()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(fmt.Errorf("[ParseExecutedFlow] read executed status file '%s' failed: %v", path.Short(), err))
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Consider it's normal if ok=false but all lines are parsed
	executed, lines, lastActiveTs, ok := parseExecutedFlow(path, lines, 0)
	if !ok && len(lines) != 0 {
		// Tolerent the corrupted (by bug) status file
		sample := getSampleLines(lines)
		//panic(fmt.Errorf("[ParseExecutedFlow] bad executed status file '%s', has %v lines unparsed: %s",
		//	path.Short(), len(lines), sample))
		executed.Corrupted = sample
	}
	return lastActiveTs, executed
}

func (self *ExecutedFlow) GenExecMasks() (masks []*ExecuteMask) {
	for _, cmd := range self.Cmds {
		var subMasks []*ExecuteMask
		if cmd.SubFlow != nil {
			subMasks = cmd.SubFlow.GenExecMasks()
		}
		policy := ExecPolicyExec
		fileNFlowPolicy := ExecPolicyExec
		if cmd.Result == ExecutedResultSucceeded {
			if cmd.SubFlow == nil {
				policy = ExecPolicySkip
			} else {
				fileNFlowPolicy = ExecPolicySkip
			}
		}
		masks = append(masks, &ExecuteMask{
			cmd.Cmd,
			cmd.StartEnv,
			cmd.FinishEnv,
			policy,
			fileNFlowPolicy,
			subMasks,
			cmd.Result,
			cmd,
		})
	}
	return
}

func (self *ExecutedFlow) IsOneCmdSession(cmd string) bool {
	return len(self.Cmds) == 1 && self.Cmds[0].Cmd == cmd
}

func (self *ExecutedFlow) MatchFind(findStrs []string) bool {
	if len(findStrs) == 0 {
		return true
	}
	for _, it := range findStrs {
		if !strings.Contains(self.DirName, it) && !strings.Contains(self.Flow, it) {
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

func (self *ExecutedFlow) CalResultInCaseIncompleted() {
	if self.Result != ExecutedResultSucceeded && self.Result != ExecutedResultIncompleted {
		return
	}
	for _, cmd := range self.Cmds {
		if cmd.Result == ExecutedResultError {
			self.Result = cmd.Result
			break
		} else if cmd.Result != ExecutedResultError && cmd.Result == ExecutedResultIncompleted {
			self.Result = cmd.Result
			break
		}
	}
}

func parseExecutedFlow(path ExecutedStatusFilePath, lines []string,
	level int) (executed *ExecutedFlow, remain []string, lastActiveTs time.Time, ok bool) {

	executed, _, remain, lastActiveTs, ok = tryParseExecutedFlow(path, lines, level)
	if executed != nil && executed.FinishTs.IsZero() && !lastActiveTs.IsZero() {
		executed.FinishTs = lastActiveTs
	}
	return
}

func tryParseExecutedFlow(path ExecutedStatusFilePath, lines []string,
	level int) (executed *ExecutedFlow, hasDelayCmd bool, remain []string, lastActiveTs time.Time, ok bool) {

	executed = NewExecutedFlow(path.DirName)

	executed.Flow, lines, ok = parseMarkedOneLineContent(path, lines, "flow", level)
	if !ok {
		return executed, hasDelayCmd, lines, lastActiveTs, false
	}
	executed.StartTs, lines, ok = parseMarkedTime(path, lines, "flow-start-time", level)
	if !ok {
		return executed, hasDelayCmd, lines, lastActiveTs, false
	}
	lastActiveTs = executed.StartTs

	var cmdsLastActiveTs time.Time
	executed.Cmds, hasDelayCmd, lines, cmdsLastActiveTs, ok = parseExecutedCmds(path, lines, level)
	if !cmdsLastActiveTs.IsZero() {
		lastActiveTs = cmdsLastActiveTs
	}
	if !ok {
		return executed, hasDelayCmd, lines, lastActiveTs, false
	}

	executed.FinishTs, lines, ok = parseMarkedTime(path, lines, "flow-finish-time", level)
	if !ok {
		return executed, hasDelayCmd, lines, lastActiveTs, false
	}
	lastActiveTs = executed.FinishTs
	executed.Result, lines, ok = parseCmdOrFlowResult(path, lines, "flow-result", level, ExecutedResultIncompleted)
	return executed, hasDelayCmd, lines, lastActiveTs, ok
}

func parseExecutedCmds(path ExecutedStatusFilePath, lines []string,
	level int) (cmds []*ExecutedCmd, hasDelayCmd bool, remain []string, lastActiveTs time.Time, ok bool) {

	for len(lines) != 0 {
		var cmd *ExecutedCmd
		var cmdLastActiveTs time.Time
		cmd, lines, cmdLastActiveTs, ok = parseExecutedCmd(path, lines, level)
		if !cmdLastActiveTs.IsZero() {
			lastActiveTs = cmdLastActiveTs
		}
		if cmd != nil && cmd.FinishTs.IsZero() && !cmdLastActiveTs.IsZero() {
			cmd.FinishTs = cmdLastActiveTs
		}
		if !ok {
			if _, _, ok := parseMarkedOneLineContent(path, lines, "flow-finish-time", level); ok {
				break
			}
			if len(lines) != 0 {
				return cmds, hasDelayCmd, lines, lastActiveTs, false
			}
		}
		if cmd.IsDelay {
			hasDelayCmd = true
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return cmds, hasDelayCmd, lines, lastActiveTs, true
}

func parseExecutedCmd(path ExecutedStatusFilePath, lines []string,
	level int) (cmd *ExecutedCmd, remain []string, lastActiveTs time.Time, ok bool) {

	cmdStr, lines, ok := parseMarkedOneLineContent(path, lines, "cmd", level)
	if !ok {
		return nil, lines, lastActiveTs, false
	}
	cmd = NewExecutedCmd(strings.TrimSpace(cmdStr))

	cmd.StartTs, lines, ok = parseMarkedTime(path, lines, "cmd-start-time", level)
	if !ok {
		return cmd, lines, cmd.StartTs, false
	}
	lastActiveTs = cmd.StartTs

	logFilePath, lines, ok := parseMarkedOneLineContent(path, lines, "log", level)
	if ok {
		cmd.LogFilePath = strings.TrimSpace(logFilePath)
	}

	var cmdLastActiveTs time.Time
	lines, cmdLastActiveTs, ok = tryParseScheduledCmd(cmd, path, lines, level)
	if !cmdLastActiveTs.IsZero() {
		lastActiveTs = cmdLastActiveTs
	}
	if ok {
		return cmd, lines, lastActiveTs, true
	}

	startEnvLines, lines, ok := parseMarkedContent(path, lines, "env-start", level)
	if ok {
		cmd.StartEnv = parseEnvLines(path, startEnvLines, level)
	}

	subflowLines, lines, ok := parseMarkedContent(path, lines, "subflow", level)
	if len(subflowLines) > 0 {
		if !ok && len(lines) != 0 {
			return cmd, lines, lastActiveTs, false
		}
		var flowLastActiveTs time.Time
		cmd.SubFlow, subflowLines, flowLastActiveTs, ok = parseExecutedFlow(path, subflowLines, level+1)
		if !flowLastActiveTs.IsZero() {
			lastActiveTs = flowLastActiveTs
		}
		if !ok {
			if len(subflowLines) != 0 {
				// Tolerent the corrupted (by bug) status file
				return cmd, subflowLines, lastActiveTs, false
				//sample := getSampleLines(subflowLines)
				//cmd.SubFlow.Corrupted = sample
				//panic(fmt.Errorf("[ParseExecutedFlow] bad subflow lines in status file '%s', has %v lines unparsed: %s",
				//	path.Short(), len(subflowLines), sample))
			}
			return cmd, lines, lastActiveTs, false
		}
	}

	finishEnvLines, lines, ok := parseMarkedContent(path, lines, "env-finish", level)
	if ok {
		cmd.FinishEnv = parseEnvLines(path, finishEnvLines, level)
	}

	cmd.FinishTs, lines, ok = parseMarkedTime(path, lines, "cmd-finish-time", level)
	if !ok {
		return cmd, lines, lastActiveTs, false
	}
	lastActiveTs = cmd.FinishTs

	cmd.Result, lines, ok = parseCmdOrFlowResult(path, lines, "cmd-result", level, ExecutedResultError)
	cmd.CalResultInCaseIncompleted()
	if !ok {
		return cmd, lines, lastActiveTs, false
	}

	errLines, lines, ok := parseMarkedContent(path, lines, "error", level)
	if ok {
		for _, line := range errLines {
			cmd.ErrStrs = append(cmd.ErrStrs, strings.TrimSpace(line))
		}
	}

	return cmd, lines, lastActiveTs, true
}

func tryParseScheduledCmd(cmd *ExecutedCmd, path ExecutedStatusFilePath,
	lines []string, level int) (remain []string, lastActiveTs time.Time, ok bool) {

	tid, lines, ok := parseMarkedOneLineContent(path, lines, "scheduled", level)
	if !ok {
		return lines, lastActiveTs, false
	}
	bgSessionPath := ExecutedStatusFilePath{path.RootPath, filepath.Join(path.DirName, tid), path.FileName}
	lastActiveTs, subflow := ParseExecutedFlow(bgSessionPath)
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
	return lines, lastActiveTs, true
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

func getSampleLines(lines []string) []string {
	sample := []string{}
	for i, line := range lines {
		if i > 3 {
			break
		}
		sample = append(sample, strings.TrimSpace(line))
	}
	return sample
}
