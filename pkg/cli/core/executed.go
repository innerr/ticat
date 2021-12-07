package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type ExecutedCmd struct {
	Cmd       string
	StartEnv  *Env
	FlowCmds  []*ExecutedCmd
	FinishEnv *Env
	Succeeded bool
	Err       []string
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
