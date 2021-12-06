package core

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type ExecutedCmd struct {
	Cmd       string
	StartEnv  *Env
	FlowCmds  []*ExecutedCmd
	FinishEnv *Env
	Err       error
}

type ExecutedFlow struct {
	Flow     string
	DirName  string
	FlowCmds []*ExecutedCmd
	Done     bool
}

func ParseExecutedFlow(path string, dirName string, env *Env) *ExecutedFlow {
	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("[ParseExecutedFlow] open executed status file '%s' failed: %v", path, err))
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(fmt.Errorf("[ParseExecutedFlow] read executed status file '%s' failed: %v", path, err))
	}

	// TODO: it's slow
	lines := strings.Split(string(data), "\n")
	lines, flowStr := parseMarkedOneLineContent(lines, "flow")

	return &ExecutedFlow{
		Flow:    flowStr,
		DirName: dirName,
		Done:    len(lines) > 0 && lines[len(lines)-1] == StatusFileEOF,
	}
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

func tryParseMarkedOneLineContent(lines []string, mark string) (remain []string, content string, ok bool) {
	remain, res, ok := tryParseMarkedContent(lines, mark)
	if !ok {
		return
	}
	if len(res) != 1 {
		panic(fmt.Errorf("[ExecutedFlow.parseMarked] content '%s' should only have one line", mark))
	}
	return remain, res[0], ok
}

func parseMarkedOneLineContent(lines []string, mark string) (remain []string, content string) {
	remain, content, ok := tryParseMarkedOneLineContent(lines, mark)
	if !ok {
		panic(fmt.Errorf("[ExecutedFlow.parseMarked] content '%s' not found", mark))
	}
	return remain, content
}

type ExecutingFlow struct {
	path   string
	level  int
	indent string
}

func NewExecutingFlow(path string, flow *ParsedCmds, env *Env) *ExecutingFlow {
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
	writeMarkedContent(self.path, "flow", flowStr)
}

func (self *ExecutingFlow) OnCmdStart(flow *ParsedCmds, index int, env *Env) {
	buf := bytes.NewBuffer(nil)
	cmdPathSep := env.GetRaw("strs.cmd-path-sep")
	fprintf(buf, "%s%s\n%s%s\n%s%s\n",
		self.indent, markStartStr("cmd"),
		self.indent, flow.Cmds[index].DisplayPath(cmdPathSep, true),
		self.indent, markFinishStr("cmd"))
	writeCmdEnv(buf, env, "env-start", self.indent)
	writeStatusContent(self.path, buf.String())
}

func (self *ExecutingFlow) OnCmdFinish(flow *ParsedCmds, index int, env *Env, err error) {
	buf := bytes.NewBuffer(nil)
	writeCmdEnv(buf, env, "env-finish", self.indent)
	writeStatusContent(self.path, buf.String())
}

func (self *ExecutingFlow) OnEnterSubFlow() {
	writeMarkStart(self.path, "subflow", self.indent)
	self.level += 1
	self.indent = strings.Repeat(StatusFileIndent, self.level)
}

func (self *ExecutingFlow) OnLeaveSubFlow() {
	writeMarkFinish(self.path, "subflow", self.indent)
	self.level -= 1
	self.indent = strings.Repeat(StatusFileIndent, self.level)
}

func (self *ExecutingFlow) OnFlowFinish() {
	writeStatusContent(self.path, StatusFileEOF+"\n")
}

func writeCmdEnv(w io.Writer, env *Env, mark string, indent string) {
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
	for k, v := range kvs {
		fprintf(buf, "%s%s=%s\n", indent, k, v)
	}
	if len(kvs) > 0 {
		fprintf(w, "%s%s\n%s%s%s\n", indent, markStartStr(mark), buf.String(), indent, markFinishStr(mark))
	}
}

func writeMarkStart(path string, mark string, indent string) {
	content := fmt.Sprintf("%s%s%s%s\n",
		indent, StatusFileMarkBracketLeft, mark, StatusFileMarkBracketRight)
	writeStatusContent(path, content)
}

func writeMarkFinish(path string, mark string, indent string) {
	content := fmt.Sprintf("%s%s%s%s%s\n",
		indent, StatusFileMarkBracketLeft, StatusFileMarkFinishMark, "subflow", StatusFileMarkBracketRight)
	writeStatusContent(path, content)
}

func writeMarkedContent(path string, mark string, lines ...string) {
	content := fmt.Sprintf("%s\n%s\n%s\n",
		StatusFileMarkBracketLeft+mark+StatusFileMarkBracketRight,
		strings.Join(lines, "\n"),
		StatusFileMarkBracketLeft+StatusFileMarkFinishMark+mark+StatusFileMarkBracketRight)
	writeStatusContent(path, content)
}

func writeStatusContent(path string, content string) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(fmt.Errorf("[ExecutedFlow.write] open executing status file '%s' failed: %v", path, err))
	}
	defer file.Close()
	_, err = fmt.Fprintf(file, content)
	if err != nil {
		panic(fmt.Errorf("[ExecutedFlow.write] write executing status file '%s' failed: %v", path, err))
	}
}

func tryParseMarkedContent(lines []string, mark string) (remain []string, content []string, ok bool) {
	remain = lines
	if len(lines) < 2 {
		return
	}
	markStart := markStartStr(mark)
	if markStart != lines[0] {
		return
	}
	lines = lines[1:]
	markFinish := markFinishStr(mark)
	for i, line := range lines {
		if markFinish == line {
			return lines[i:], lines[0:i], true
		}
	}
	return
}

func fprintf(w io.Writer, format string, a ...interface{}) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		panic(err)
	}
}

func markStartStr(mark string) string {
	return StatusFileMarkBracketLeft + mark + StatusFileMarkBracketRight
}

func markFinishStr(mark string) string {
	return StatusFileMarkBracketLeft + StatusFileMarkFinishMark + mark + StatusFileMarkBracketRight
}

const (
	StatusFileMarkBracketLeft  = "[<"
	StatusFileMarkBracketRight = ">]"
	StatusFileMarkFinishMark   = "/"
	StatusFileEOF              = StatusFileMarkBracketLeft + "EOF" + StatusFileMarkFinishMark + StatusFileMarkBracketRight
	StatusFileIndent           = "    "
)
