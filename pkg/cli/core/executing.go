package core

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

type ExecutingFlow struct {
	path  string
	level int
}

func NewExecutingFlow(path string, flow *ParsedCmds, env *Env) *ExecutingFlow {
	if len(path) > 0 && path[0] != '/' && path[0] != '\\' {
		panic(fmt.Errorf("[ExecutingFlow] status file '%s' invalid path", path))
	}

	executing := &ExecutingFlow{
		path:  path,
		level: 0,
	}
	executing.onFlowStart(flow, env)
	return executing
}

func (self *ExecutingFlow) onFlowStart(flow *ParsedCmds, env *Env) {
	trivialMark := env.GetRaw("strs.trivial-mark")
	cmdPathSep := env.GetRaw("strs.cmd-path-sep")
	buf := bytes.NewBuffer(nil)
	SaveFlow(buf, flow, 0, cmdPathSep, trivialMark, env)
	flowStr := buf.String()
	writeMarkedContent(self.path, "flow", 0, flowStr)
}

func (self *ExecutingFlow) OnCmdStart(flow *ParsedCmds, index int, env *Env) {
	indent := strings.Repeat(StatusFileIndent, self.level)
	buf := bytes.NewBuffer(nil)
	cmdPathSep := env.GetRaw("strs.cmd-path-sep")
	fprintf(buf, "%s\n%s%s\n%s\n",
		markStartStr("cmd", self.level),
		indent, flow.Cmds[index].DisplayPath(cmdPathSep, true),
		markFinishStr("cmd", self.level))
	writeCmdEnv(buf, env, "env-start", self.level)
	writeStatusContent(self.path, buf.String())
}

func (self *ExecutingFlow) OnAsyncTaskSchedule(flow *ParsedCmds, index int, env *Env, tid string) {
	indent := strings.Repeat(StatusFileIndent, self.level)
	buf := bytes.NewBuffer(nil)
	cmdPathSep := env.GetRaw("strs.cmd-path-sep")
	fprintf(buf, "%s\n%s%s\n%s\n",
		markStartStr("cmd", self.level),
		indent, flow.Cmds[index].DisplayPath(cmdPathSep, true),
		markFinishStr("cmd", self.level))
	fprintf(buf, "%s%s%s\n", markStartStr("scheduled", self.level), tid, markFinishStr("scheduled", 0))
	writeStatusContent(self.path, buf.String())
}

func (self *ExecutingFlow) OnCmdFinish(flow *ParsedCmds, index int, env *Env, succeeded bool, err error) {
	indent := strings.Repeat(StatusFileIndent, self.level)
	buf := bytes.NewBuffer(nil)
	writeCmdEnv(buf, env, "env-finish", self.level)
	fprintf(buf, "%s%v%s\n", markStartStr("result", self.level), succeeded, markFinishStr("result", 0))
	if err != nil {
		fprintf(buf, "%s\n", markStartStr("error", self.level))
		for _, line := range strings.Split(err.Error(), "\n") {
			fprintf(buf, "%s%s\n", indent, line)
		}
		fprintf(buf, "%s\n", markFinishStr("error", self.level))
	}
	writeStatusContent(self.path, buf.String())
}

func (self *ExecutingFlow) OnEnterSubFlow(flow string) {
	writeMarkStart(self.path, "subflow", self.level)
	self.level += 1
	writeMarkedContent(self.path, "flow", self.level, flow)
}

func (self *ExecutingFlow) OnLeaveSubFlow() {
	self.level -= 1
	writeMarkFinish(self.path, "subflow", self.level)
}

func (self *ExecutingFlow) OnFlowFinish() {
	writeStatusContent(self.path, StatusFileEOF+"\n")
}

func writeCmdEnv(w io.Writer, env *Env, mark string, level int) {
	envPathSep := env.GetRaw("strs.env-path-sep")
	// TODO: put these into config or env.key's prop
	filterPrefixs := []string{
		"session",
		"strs" + envPathSep,
		"display" + envPathSep,
		"sys" + envPathSep,
	}

	kvs := env.Flatten(false, filterPrefixs, true)
	buf := bytes.NewBuffer(nil)
	indent := strings.Repeat(StatusFileIndent, level)
	for k, v := range kvs {
		fprintf(buf, "%s%s=%s\n", indent, k, v)
	}
	if len(kvs) > 0 {
		fprintf(w, "%s\n%s%s\n", markStartStr(mark, level), buf.String(), markFinishStr(mark, level))
	} else {
		//fprintf(w, "%s\n", emptyMarkStr(mark, level))
	}
}

func writeMarkStart(path string, mark string, level int) {
	indent := strings.Repeat(StatusFileIndent, level)
	content := fmt.Sprintf("%s%s%s%s\n",
		indent, StatusFileMarkBracketLeft, mark, StatusFileMarkBracketRight)
	writeStatusContent(path, content)
}

func writeMarkFinish(path string, mark string, level int) {
	indent := strings.Repeat(StatusFileIndent, level)
	content := fmt.Sprintf("%s%s%s%s%s\n",
		indent, StatusFileMarkBracketLeft, StatusFileMarkFinishMark, "subflow", StatusFileMarkBracketRight)
	writeStatusContent(path, content)
}

func writeMarkedContent(path string, mark string, level int, lines ...string) {
	indent := strings.Repeat(StatusFileIndent, level)
	content := indent + StatusFileMarkBracketLeft + mark + StatusFileMarkBracketRight + "\n"
	for _, line := range lines {
		content += indent + line + "\n"
	}
	content += indent + StatusFileMarkBracketLeft + StatusFileMarkFinishMark + mark + StatusFileMarkBracketRight + "\n"
	writeStatusContent(path, content)
}

func writeStatusContent(path string, content string) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(fmt.Errorf("[ExecutingFlow] open executing status file '%s' failed: %v", path, err))
	}
	defer file.Close()
	_, err = file.Write([]byte(content))
	if err != nil {
		panic(fmt.Errorf("[ExecutingFlow] write executing status file '%s' failed: %v", path, err))
	}
}

func fprintf(w io.Writer, format string, a ...interface{}) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		panic(err)
	}
}

func markStartStr(mark string, level int) string {
	return strings.Repeat(StatusFileIndent, level) + StatusFileMarkBracketLeft + mark + StatusFileMarkBracketRight
}

func markFinishStr(mark string, level int) string {
	return strings.Repeat(StatusFileIndent, level) + StatusFileMarkBracketLeft +
		StatusFileMarkFinishMark + mark + StatusFileMarkBracketRight
}

func emptyMarkStr(mark string, level int) string {
	return strings.Repeat(StatusFileIndent, level) + StatusFileMarkBracketLeft + mark +
		StatusFileMarkFinishMark + StatusFileMarkBracketRight
}

const (
	StatusFileMarkBracketLeft  = "<"
	StatusFileMarkBracketRight = ">"
	StatusFileMarkFinishMark   = "/"
	StatusFileEOF              = StatusFileMarkBracketLeft + "EOF" + StatusFileMarkFinishMark + StatusFileMarkBracketRight
	StatusFileIndent           = "    "
)
