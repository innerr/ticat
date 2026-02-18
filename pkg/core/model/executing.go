package model

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// TODO: this session status file format (and code) is bad and dirty, rewrite it

type ExecutingFlow struct {
	path  string
	level int
}

func NewExecutingFlow(path string, flow *ParsedCmds, env *Env) *ExecutingFlow {
	if len(path) > 0 && path[0] != '/' && path[0] != '\\' {
		// PANIC: Programming error - invalid status file path
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
	if env.GetBool("sys.unlog-status") {
		return
	}

	buf := bytes.NewBuffer(nil)

	trivialMark := env.GetRaw("strs.trivial-mark")
	cmdPathSep := env.GetRaw("strs.cmd-path-sep")
	flowStr, _ := SaveFlowToStr(flow, cmdPathSep, trivialMark, env)
	buf.Write([]byte(markedContent("flow", 0, flowStr)))

	now := time.Now().Format(SessionTimeFormat)
	buf.Write([]byte(markedOneLineContent("flow-start-time", self.level, now)))

	writeStatusContent(self.path, buf.String())
}

func (self *ExecutingFlow) OnCmdStart(flow *ParsedCmds, index int, env *Env, logFilePath string) {
	if env.GetBool("sys.unlog-status") {
		return
	}

	buf := bytes.NewBuffer(nil)

	cmdPathSep := env.GetRaw("strs.cmd-path-sep")
	cmdName := strings.Join(flow.Cmds[index].Path(), cmdPathSep)
	buf.Write([]byte(markedOneLineContent("cmd", self.level, cmdName)))

	now := time.Now().Format(SessionTimeFormat)
	buf.Write([]byte(markedOneLineContent("cmd-start-time", self.level, now)))

	if len(logFilePath) != 0 {
		buf.Write([]byte(markedOneLineContent("log", self.level, logFilePath)))
	}

	writeCmdEnv(buf, env, "env-start", self.level)

	writeStatusContent(self.path, buf.String())
}

func (self *ExecutingFlow) OnAsyncTaskSchedule(flow *ParsedCmds, index int, env *Env, tid string) {
	if env.GetBool("sys.unlog-status") {
		return
	}

	buf := bytes.NewBuffer(nil)

	cmdPathSep := env.GetRaw("strs.cmd-path-sep")
	cmdName := strings.Join(flow.Cmds[index].Path(), cmdPathSep)
	buf.Write([]byte(markedOneLineContent("cmd", self.level, cmdName)))

	now := time.Now().Format(SessionTimeFormat)
	buf.Write([]byte(markedOneLineContent("cmd-start-time", self.level, now)))

	buf.Write([]byte(markedOneLineContent("scheduled", self.level, tid)))

	writeStatusContent(self.path, buf.String())
}

func (self *ExecutingFlow) OnCmdFinish(flow *ParsedCmds, index int, env *Env, succeeded bool, err error, skipped bool) {
	if env.GetBool("sys.unlog-status") {
		return
	}

	buf := bytes.NewBuffer(nil)

	writeCmdEnv(buf, env, "env-finish", self.level)

	result := ExecutedResultError
	now := time.Now().Format(SessionTimeFormat)
	buf.Write([]byte(markedOneLineContent("cmd-finish-time", self.level, now)))

	if succeeded {
		if skipped {
			result = ExecutedResultSkipped
		} else {
			result = ExecutedResultSucceeded
		}
	}
	buf.Write([]byte(markedOneLineContent("cmd-result", self.level, string(result))))

	if err != nil {
		errLines := strings.Split(err.Error(), "\n")
		buf.Write([]byte(markedContent("error", self.level, errLines...)))
	}

	writeStatusContent(self.path, buf.String())
}

func (self *ExecutingFlow) OnSubFlowStart(env *Env, flow string) {
	if env.GetBool("sys.unlog-status") {
		return
	}

	content := markStartStr("subflow", self.level) + "\n"

	self.level += 1

	content += markedContent("flow", self.level, flow)

	now := time.Now().Format(SessionTimeFormat)
	content += markedOneLineContent("flow-start-time", self.level, now)

	writeStatusContent(self.path, content)
}

func (self *ExecutingFlow) OnSubFlowFinish(env *Env, succeeded bool, skipped bool) {
	if env.GetBool("sys.unlog-status") {
		return
	}

	buf := bytes.NewBuffer(nil)

	now := time.Now().Format(SessionTimeFormat)
	buf.Write([]byte(markedOneLineContent("flow-finish-time", self.level, now)))

	result := ExecutedResultError
	if succeeded {
		if skipped {
			result = ExecutedResultSkipped
		} else {
			result = ExecutedResultSucceeded
		}
	}

	buf.Write([]byte(markedOneLineContent("flow-result", self.level, string(result))))

	self.level -= 1
	buf.Write([]byte(markFinishStr("subflow", self.level) + "\n"))

	writeStatusContent(self.path, buf.String())
}

func (self *ExecutingFlow) OnFlowFinish(env *Env, succeeded bool) {
	if env.GetBool("sys.unlog-status") {
		return
	}

	buf := bytes.NewBuffer(nil)

	now := time.Now().Format(SessionTimeFormat)
	buf.Write([]byte(markedOneLineContent("flow-finish-time", self.level, now)))

	result := ExecutedResultError
	if succeeded {
		result = ExecutedResultSucceeded
	}
	buf.Write([]byte(markedOneLineContent("flow-result", self.level, string(result))))

	writeStatusContent(self.path, buf.String())
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
	}
}

func writeMarkStart(path string, mark string, level int) {
	indent := strings.Repeat(StatusFileIndent, level)
	content := indent + StatusFileMarkBracketLeft + mark + StatusFileMarkBracketRight + "\n"
	writeStatusContent(path, content)
}

func writeMarkedContent(path string, mark string, level int, lines ...string) {
	content := markedContent(mark, level, lines...)
	writeStatusContent(path, content)
}

func writeStatusContent(path string, content string) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		// Runtime error: cannot open status file - silently ignore
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Runtime error: cannot close status file - silently ignore
		}
	}()
	_, err = file.Write([]byte(content))
	if err != nil {
		// Runtime error: cannot write to status file - silently ignore
	}
}

func fprintf(w io.Writer, format string, a ...interface{}) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		// Runtime error: write failed - silently ignore
	}
}

func markedContent(mark string, level int, lines ...string) (content string) {
	indent := strings.Repeat(StatusFileIndent, level)
	content += indent + StatusFileMarkBracketLeft + mark + StatusFileMarkBracketRight + "\n"
	for _, line := range lines {
		content += indent + line + "\n"
	}
	content += indent + StatusFileMarkBracketLeft + StatusFileMarkFinishMark + mark + StatusFileMarkBracketRight + "\n"
	return
}

func markedOneLineContent(mark string, level int, line string) (content string) {
	indent := strings.Repeat(StatusFileIndent, level)
	content += indent + StatusFileMarkBracketLeft + mark + StatusFileMarkBracketRight
	content += line
	content += StatusFileMarkBracketLeft + StatusFileMarkFinishMark + mark + StatusFileMarkBracketRight + "\n"
	return
}

func markStartStr(mark string, level int) string {
	return strings.Repeat(StatusFileIndent, level) + StatusFileMarkBracketLeft + mark + StatusFileMarkBracketRight
}

func markFinishStr(mark string, level int) string {
	return strings.Repeat(StatusFileIndent, level) + StatusFileMarkBracketLeft +
		StatusFileMarkFinishMark + mark + StatusFileMarkBracketRight
}

const (
	StatusFileMarkBracketLeft  = "<"
	StatusFileMarkBracketRight = ">"
	StatusFileMarkFinishMark   = "/"
	StatusFileIndent           = "    "
)
