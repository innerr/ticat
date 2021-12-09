package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ExecutedStatusFilePath struct {
	RootPath string
	DirName  string
	FileName string
}

type ExecutedCmd struct {
	Cmd        string
	IsDelay    bool
	StartEnv   *Env
	SubFlow    *ExecutedFlow
	FinishEnv  *Env
	Unexecuted bool
	Succeeded  bool
	Err        []string
}

type ExecutedFlow struct {
	Flow     string
	DirName  string
	Cmds     []*ExecutedCmd
	Executed bool
	FinishTs time.Time
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
	return parseExecutedFlow(path, lines)
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

func parseExecutedFlow(path ExecutedStatusFilePath, lines []string) (executed *ExecutedFlow) {
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	executed = &ExecutedFlow{
		DirName: path.DirName,
	}

	flowStr, lines, ok := parseMarkedOneLineContent(path, lines, "flow", 0)
	if !ok {
		return
	}
	executed.Flow = flowStr

	cmds, lines, ok := parseExecutedCmds(path, lines, 0)
	if !ok {
		return
	}
	executed.Cmds = cmds

	executed.FinishTs, executed.Executed = parseStatusFileEOF(path, lines)
	return
}

func parseStatusFileEOF(path ExecutedStatusFilePath, lines []string) (finishTs time.Time, ok bool) {
	if len(lines) != 1 {
		return
	}
	var finishTsStr string
	finishTsStr, _, ok = parseMarkedOneLineContent(path, lines, StatusFileEOF, 0)
	if !ok {
		return
	}
	finishTs, err := time.ParseInLocation(SessionTimeFormat, finishTsStr, time.Local)
	if err != nil {
		panic(fmt.Errorf("[ParseExecutedFlow] bad finish-ts format '%s' in status file '%s'", finishTsStr, path.Short()))
	}
	return finishTs, ok
}

func parseExecutedCmds(path ExecutedStatusFilePath, lines []string, level int) (cmds []*ExecutedCmd, remain []string, ok bool) {
	for len(lines) != 0 {
		var cmd *ExecutedCmd
		cmd, lines, ok = parseExecutedCmd(path, lines, level)
		if !ok {
			if _, ok := parseStatusFileEOF(path, lines); ok {
				break
			}
			if len(lines) != 0 {
				panic(fmt.Errorf("[ParseExecutedFlow] bad executed status file '%s', has extra %v lines",
					path.Short(), len(lines)))
			}
			return cmds, nil, false
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
	cmd = &ExecutedCmd{
		Cmd: strings.TrimSpace(cmdStr),
	}

	tid, lines, ok := parseMarkedOneLineContent(path, lines, "scheduled", level)
	if ok {
		bgSessionPath := ExecutedStatusFilePath{path.RootPath, filepath.Join(path.DirName, tid), path.FileName}
		subflow := ParseExecutedFlow(bgSessionPath)
		if len(subflow.Cmds) == 0 {
			subflow.Cmds = append(subflow.Cmds, &ExecutedCmd{Cmd: cmd.Cmd, Unexecuted: true})
		} else if len(subflow.Cmds) != 1 {
			panic(fmt.Errorf("[ParseExecutedFlow] expect only one cmd in delayed task"))
		}
		cmd.SubFlow = subflow
		cmd.IsDelay = true
		cmd.Succeeded = true
		return cmd, lines, true
	}

	startEnvLines, lines, ok := parseMarkedContent(path, lines, "env-start", level)
	if ok {
		cmd.StartEnv = parseEnvLines(path, startEnvLines, level)
	}

	subflowLines, lines, ok := parseMarkedContent(path, lines, "subflow", level)
	if len(subflowLines) > 0 {
		if !ok && len(lines) != 0 {
			panic(fmt.Errorf("[ParseExecutedFlow] bad subflow in executed status file '%s', has extra %v lines",
				path.Short(), len(lines)))
		}
		cmd.Succeeded = ok
		subflow := &ExecutedFlow{}
		flowStr, subflowLines, ok := parseMarkedOneLineContent(path, subflowLines, "flow", level+1)
		if ok {
			subflow.Flow = flowStr
		} else {
			panic(fmt.Errorf("[ParseExecutedFlow] expect 'flow' mark"))
		}
		cmds, subflowRemain, ok := parseExecutedCmds(path, subflowLines, level+1)
		if len(subflowRemain) != 0 {
			panic(fmt.Errorf("[ParseExecutedFlow] bad lines of subflow in status file '%s'", path.Short()))
		}
		if !ok {
			panic(fmt.Errorf("[ParseExecutedFlow] parse subflow failed in status file '%s'", path.Short()))
		}
		subflow.Cmds = cmds
		cmd.SubFlow = subflow
	}

	finishEnvLines, lines, ok := parseMarkedContent(path, lines, "env-finish", level)
	if ok {
		cmd.FinishEnv = parseEnvLines(path, finishEnvLines, level)
	}

	succeeded, lines, ok := parseMarkedOneLineContent(path, lines, "result", level)
	if !ok {
		if cmd.SubFlow != nil {
			return cmd, lines, true
		} else {
			return nil, lines, false
		}
	}
	cmd.Succeeded = (succeeded == "true")

	errLines, lines, ok := parseMarkedContent(path, lines, "error", level)
	if ok {
		for _, line := range errLines {
			cmd.Err = append(cmd.Err, strings.TrimSpace(line))
		}
	}

	return cmd, lines, true
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

	remain = lines
	if len(lines) == 0 {
		return
	}

	if lines[0] == emptyMarkStr(mark, level) {
		return nil, lines[1:], true
	}

	markStart := markStartStr(mark, level)
	markFinish := markFinishStr(mark, 0)
	if strings.HasPrefix(lines[0], markStart) && strings.HasSuffix(lines[0], markFinish) {
		return []string{lines[0][len(markStart) : len(lines[0])-len(markFinish)]}, lines[1:], true
	}

	if lines[0] != markStart {
		return nil, remain, false
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
