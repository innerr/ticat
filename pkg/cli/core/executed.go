package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type ExecutedCmd struct {
	Cmd       string
	IsDelay   bool
	StartEnv  *Env
	Cmds      []*ExecutedCmd
	FinishEnv *Env
	Succeeded bool
	Err       []string
}

type ExecutedFlow struct {
	Flow     string
	DirName  string
	Cmds     []*ExecutedCmd
	Executed bool
}

func ParseExecutedFlow(path string, dirName string) *ExecutedFlow {
	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("[ParseExecutedFlow] open executed status file '%s' failed: %v", path, err))
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(fmt.Errorf("[ParseExecutedFlow] read executed status file '%s' failed: %v", path, err))
	}
	lines := strings.Split(string(data), "\n")
	return parseExecutedFlow(path, lines, dirName)
}

func (self *ExecutedFlow) MatchFind(findStrs []string) bool {
	if len(findStrs) == 0 {
		return true
	}
	for _, it := range findStrs {
		if len(self.DirName) > 0 && strings.Index(self.DirName, it) >= 0 || strings.Index(self.Flow, it) >= 0 {
			return true
		}
	}
	return false
}

func parseExecutedFlow(path string, lines []string, dirName string) (executed *ExecutedFlow) {
	executed = &ExecutedFlow{
		DirName: dirName,
	}

	flowStrs, lines, ok := parseMarkedContent(path, lines, "flow", 0)
	if !ok {
		return
	}
	executed.Flow = strings.Join(flowStrs, " ")

	cmds, lines, ok := parseExecutedCmds(path, dirName, lines, 0)
	if !ok {
		return
	}
	executed.Cmds = cmds

	executed.Executed = parseStatusFileEOF(lines)
	return
}

func parseStatusFileEOF(lines []string) bool {
	return len(lines) == 1 && lines[0] == StatusFileEOF
}

func parseExecutedCmds(path string, dirName string, lines []string, level int) (cmds []*ExecutedCmd, remain []string, ok bool) {
	for len(lines) != 0 {
		var cmd *ExecutedCmd
		cmd, lines, ok = parseExecutedCmd(path, dirName, lines, level)
		if !ok {
			if !parseStatusFileEOF(lines) {
				panic(fmt.Errorf("[ParseExecutedFlow] bad executed status file '%s', has extra %v lines", path, len(lines)))
			}
			return cmds, nil, false
		}
		cmds = append(cmds, cmd)
	}
	return cmds, lines, true
}

func parseExecutedCmd(path string, dirName string, lines []string, level int) (cmd *ExecutedCmd, remain []string, ok bool) {
	cmdStr, lines, ok := parseMarkedOneLineContent(path, lines, "cmd", level)
	if !ok {
		return nil, lines, false
	}
	cmd = &ExecutedCmd{
		Cmd: cmdStr,
	}

	tid, lines, ok := parseMarkedOneLineContent(path, lines, "scheduled", level)
	if ok {
		bgSessionPath := filepath.Join(path, tid)
		bgFlow := ParseExecutedFlow(bgSessionPath, filepath.Join(dirName, tid))
		cmd.Cmds = bgFlow.Cmds
		cmd.IsDelay = true
		return cmd, lines, true
	}

	startEnvLines, lines, ok := parseMarkedContent(path, lines, "env-start", level)
	if !ok {
		return nil, lines, false
	}
	cmd.StartEnv = parseEnvLines(path, startEnvLines, level)

	subflowLines, lines, ok := parseMarkedContent(path, lines, "subflow", level)
	if ok {
		cmds, subflowRemain, ok := parseExecutedCmds(path, dirName, subflowLines, level+1)
		if len(subflowRemain) != 0 {
			panic(fmt.Errorf("[ParseExecutedFlow] bad lines of subflow in status file '%s'", path))
		}
		if !ok {
			panic(fmt.Errorf("[ParseExecutedFlow] parse subflow failed in status file '%s'", path))
		}
		cmd.Cmds = cmds
	}

	finishEnvLines, lines, ok := parseMarkedContent(path, lines, "env-finish", level)
	if ok {
		cmd.FinishEnv = parseEnvLines(path, finishEnvLines, level)
	}

	succeeded, lines, ok := parseMarkedOneLineContent(path, lines, "result", level)
	if !ok {
		return nil, lines, false
	}
	cmd.Succeeded = (succeeded == "true")

	errLines, lines, ok := parseMarkedContent(path, lines, "error", level)
	if ok {
		cmd.Err = errLines
	}

	return cmd, lines, true
}

func parseEnvLines(path string, lines []string, level int) (env *Env) {
	env = NewEnvEx(EnvLayerSession)
	indent := strings.Repeat(StatusFileIndent, level)
	for _, line := range lines {
		if !strings.HasPrefix(line, indent) {
			panic(fmt.Errorf("[ParseExecutedFlow] bad indent of env line '%s' in status file '%s'", line, path))
		}
		line = line[len(indent):]
		i := strings.Index(line, "=")
		if i <= 0 {
			panic(fmt.Errorf("[ParseExecutedFlow] bad format of env line '%s' in status file '%s'", line, path))
		}
		env.Set(line[0:i], line[i+1:])
	}
	return env
}

func parseMarkedOneLineContent(path string, lines []string, mark string, level int) (content string, remain []string, ok bool) {
	res, remain, ok := parseMarkedContent(path, lines, mark, level)
	if !ok {
		return
	}
	if len(res) != 1 {
		panic(fmt.Errorf("[ParseExecutedFlow] expect only one line for mark '%s' in status file '%s'", mark, path))
	}
	return res[0], remain, ok
}

func parseMarkedContent(path string, lines []string, mark string, level int) (content []string, remain []string, ok bool) {
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
					panic(fmt.Errorf("[ParseExecutedFlow] bad recusive mark '%s' in status file '%s'", mark, path))
				}
			}
		}
	}

	return lines, nil, false
}
