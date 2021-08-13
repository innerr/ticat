package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mattn/go-shellwords"
)

type CmdType string

const (
	CmdTypeUninited   CmdType = "uninited"
	CmdTypeNormal     CmdType = "normal"
	CmdTypePower      CmdType = "power"
	CmdTypeFlow       CmdType = "flow"
	CmdTypeEmpty      CmdType = "no-executable"
	CmdTypeFile       CmdType = "executable-file"
	CmdTypeEmptyDir   CmdType = "dir-with-no-executable"
	CmdTypeDirWithCmd CmdType = "dir-with-executable-file"
)

type NormalCmd func(argv ArgVals, cc *Cli, env *Env, flow []ParsedCmd) (succeeded bool)
type PowerCmd func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds,
	currCmdIdx int) (newCurrCmdIdx int, succeeded bool)

type Cmd struct {
	owner        *CmdTree
	help         string
	ty           CmdType
	quiet        bool
	priority     bool
	args         Args
	normal       NormalCmd
	power        PowerCmd
	cmdLine      string
	flow         []string
	envOps       EnvOps
	depends      []Depend
	metaFilePath string
	val2env      *Val2Env
	arg2env      *Arg2Env
}

func defaultCmd(owner *CmdTree, help string) *Cmd {
	return &Cmd{
		owner:        owner,
		help:         help,
		ty:           CmdTypeUninited,
		quiet:        false,
		priority:     false,
		args:         newArgs(),
		normal:       nil,
		power:        nil,
		cmdLine:      "",
		flow:         nil,
		envOps:       newEnvOps(),
		depends:      nil,
		metaFilePath: "",
		val2env:      newVal2Env(),
		arg2env:      newArg2Env(),
	}
}

func NewCmd(owner *CmdTree, help string, cmd NormalCmd) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypeNormal
	c.normal = cmd
	return c
}

func NewEmptyCmd(owner *CmdTree, help string) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypeEmpty
	return c
}

func NewPowerCmd(owner *CmdTree, help string, cmd PowerCmd) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypePower
	c.power = cmd
	return c
}

func NewFileCmd(owner *CmdTree, help string, cmd string) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypeFile
	c.cmdLine = cmd
	return c
}

func NewEmptyDirCmd(owner *CmdTree, help string, dir string) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypeEmptyDir
	c.cmdLine = dir
	return c
}

func NewDirWithCmd(owner *CmdTree, help string, cmd string) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypeDirWithCmd
	c.cmdLine = cmd
	return c
}

func NewFlowCmd(owner *CmdTree, help string, flow []string) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypeFlow
	c.flow = flow
	return c
}

func (self *Cmd) Execute(
	argv ArgVals,
	cc *Cli,
	env *Env,
	flow *ParsedCmds,
	currCmdIdx int) (int, bool) {

	newCurrCmdIdx, ok := self.execute(argv, cc, env, flow, currCmdIdx)
	if !ok {
		panic(NewCmdError(flow.Cmds[currCmdIdx], "command failed without detail info"))
	}
	return newCurrCmdIdx, ok
}

func (self *Cmd) execute(
	argv ArgVals,
	cc *Cli,
	env *Env,
	flow *ParsedCmds,
	currCmdIdx int) (int, bool) {

	switch self.ty {
	case CmdTypePower:
		currCmdIdx, succeeded := self.power(argv, cc, env, flow, currCmdIdx)
		if flow.TailMode {
			currCmdIdx = 0
			flow.Cmds = nil
		}
		return currCmdIdx, succeeded
	case CmdTypeNormal:
		return currCmdIdx, self.normal(argv, cc, env, flow.Cmds[currCmdIdx:])
	case CmdTypeFile:
		return currCmdIdx, self.executeFile(argv, cc, env, flow.Cmds[currCmdIdx])
	case CmdTypeEmptyDir:
		return currCmdIdx, true
	case CmdTypeDirWithCmd:
		return currCmdIdx, self.executeFile(argv, cc, env, flow.Cmds[currCmdIdx])
	case CmdTypeFlow:
		return currCmdIdx, self.executeFlow(argv, cc, env)
	case CmdTypeEmpty:
		return currCmdIdx, true
	default:
		panic(NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("[Cmd.Execute] unknown cmd executable type: %v", self.ty)))
	}
}

func (self *Cmd) MatchFind(findStr string) bool {
	if strings.Index(self.owner.DisplayPath(), findStr) >= 0 {
		return true
	}
	if strings.Index(self.help, findStr) >= 0 {
		return true
	}
	if strings.Index(self.cmdLine, findStr) >= 0 {
		return true
	}
	if self.args.MatchFind(findStr) {
		return true
	}
	if self.val2env.MatchFind(findStr) {
		return true
	}
	if self.arg2env.MatchFind(findStr) {
		return true
	}
	if self.envOps.MatchFind(findStr) {
		return true
	}
	if strings.Index(string(self.ty), findStr) >= 0 {
		return true
	}
	for _, dep := range self.depends {
		if strings.Index(dep.OsCmd, findStr) >= 0 {
			return true
		}
		if strings.Index(dep.Reason, findStr) >= 0 {
			return true
		}
	}
	if self.quiet && strings.Index("quiet", findStr) >= 0 {
		return true
	}
	if self.ty == CmdTypePower && strings.Index("power", findStr) >= 0 {
		return true
	}
	if self.priority && strings.Index("priority", findStr) >= 0 {
		return true
	}
	return false
}

func (self *Cmd) MatchWriteKey(key string) bool {
	if self.envOps.MatchWriteKey(key) {
		return true
	}
	return false
}

func (self *Cmd) AddArg(name string, defVal string, abbrs ...string) *Cmd {
	self.args.AddArg(self.owner, name, defVal, abbrs...)
	return self
}

func (self *Cmd) AddEnvOp(name string, op uint) *Cmd {
	self.envOps.AddOp(name, op)
	return self
}

func (self *Cmd) AddSub(name string, abbrs ...string) *CmdTree {
	return self.owner.AddSub(name, abbrs...)
}

func (self *Cmd) SetMetaFile(path string) *Cmd {
	self.metaFilePath = path
	return self
}

func (self *Cmd) AddDepend(dep string, reason string) *Cmd {
	self.depends = append(self.depends, Depend{dep, reason})
	return self
}

func (self *Cmd) SetQuiet() *Cmd {
	self.quiet = true
	return self
}

func (self *Cmd) SetPriority() *Cmd {
	self.priority = true
	return self
}

func (self *Cmd) AddVal2Env(envKey string, val string) *Cmd {
	self.val2env.Add(envKey, val)
	return self
}

func (self *Cmd) AddArg2Env(envKey string, argName string) *Cmd {
	self.arg2env.Add(envKey, argName)
	return self
}

func (self *Cmd) GetVal2Env() *Val2Env {
	return self.val2env
}

func (self *Cmd) GetArg2Env() *Arg2Env {
	return self.arg2env
}

func (self *Cmd) GetDepends() []Depend {
	return self.depends
}

func (self *Cmd) MetaFile() string {
	return self.metaFilePath
}

func (self *Cmd) Owner() *CmdTree {
	return self.owner
}

func (self *Cmd) Help() string {
	return self.help
}

func (self *Cmd) DisplayHelpStr() string {
	if len(self.help) == 0 && self.ty == CmdTypeFlow {
		return self.cmdLine
	}
	return self.help
}

func (self *Cmd) IsNoExecutableCmd() bool {
	if len(self.val2env.EnvKeys()) > 0 {
		return false
	}
	if len(self.arg2env.EnvKeys()) > 0 {
		return false
	}
	return self.ty == CmdTypeUninited || self.ty == CmdTypeEmpty || self.ty == CmdTypeEmptyDir
}

func (self *Cmd) IsPowerCmd() bool {
	return self.ty == CmdTypePower
}

func (self *Cmd) IsQuiet() bool {
	return self.quiet
}

func (self *Cmd) IsPriority() bool {
	return self.priority
}

func (self *Cmd) Type() CmdType {
	return self.ty
}

func (self *Cmd) CmdLine() string {
	return self.cmdLine
}

func (self *Cmd) Args() Args {
	return self.args
}

func (self *Cmd) EnvOps() EnvOps {
	return self.envOps
}

func (self *Cmd) FlowStrs() []string {
	return self.flow
}

func (self *Cmd) IsTheSameFunc(fun interface{}) bool {
	fr1 := reflect.ValueOf(fun)
	if self.power != nil {
		fr2 := reflect.ValueOf(self.power)
		if fr1.Pointer() == fr2.Pointer() {
			return true
		}
	}
	if self.normal != nil {
		fr2 := reflect.ValueOf(self.normal)
		if fr1.Pointer() == fr2.Pointer() {
			return true
		}
	}
	return false
}

// TODO: move to parser ?
func (self *Cmd) RenderedFlowStrs(
	argv ArgVals,
	env *Env,
	allowFlowTemplateRenderError bool) (flow []string, fullyRendered bool) {

	templBracketLeft := self.owner.Strs.FlowTemplateBracketLeft
	templBracketRight := self.owner.Strs.FlowTemplateBracketRight
	templMultiplyMark := self.owner.Strs.FlowTemplateMultiplyMark
	hasError := false

	renderLineAndAddToFlow := func(it string) {
		findPos := 0
		for {
			str := it[findPos:]
			i := strings.Index(str, templBracketLeft)
			if i < 0 {
				break
			}
			tail := str[i+len(templBracketLeft):]
			j := strings.Index(tail, templBracketRight)
			if j < 0 {
				break
			}
			key := tail[0:j]
			if env == nil {
				// TODO: remove this, not allow env is nil
				// return self.flow, false
				panic(fmt.Errorf("legacy code, should never happen. TODO: remove this"))
			}
			var valStr string
			val, ok := env.GetEx(key)
			valStr = val.Raw
			if !ok {
				val, inArg := argv[key]
				valStr = val.Raw
				ok = inArg && len(valStr) != 0
			}
			if !ok {
				if allowFlowTemplateRenderError {
					hasError = true
					findPos += j + len(templBracketRight)
					continue
				}
				self.flowTemplateRenderPanic(key, false)
			}
			it = it[:findPos] + str[0:i] + valStr + tail[j+len(templBracketRight):]
		}
		flow = append(flow, it)
	}

	for _, it := range self.flow {
		var lines []string
		lines, hasError = self.tryRenderMultiply(argv, env, it, templBracketLeft, templBracketRight,
			templMultiplyMark, allowFlowTemplateRenderError)
		for _, line := range lines {
			if hasError {
				flow = append(flow, line)
				continue
			} else {
				renderLineAndAddToFlow(line)
			}
		}
	}
	fullyRendered = !hasError
	return
}

func (self *Cmd) Flow(argv ArgVals, env *Env, allowFlowTemplateRenderError bool) (flow []string, rendered bool) {
	flow, rendered = self.RenderedFlowStrs(argv, env, allowFlowTemplateRenderError)
	if !rendered || len(flow) == 0 {
		return
	}
	flowStr := strings.Join(flow, " ")
	flow, err := shellwords.Parse(flowStr)
	if err != nil {
		// TODO: better display
		panic(fmt.Errorf("[Cmd.executeFlow] parse '%s' failed: %v",
			self.cmdLine, err))
	}
	return
}

// TODO:
//
// The env in the sub flow is cc.GlobalEnv:
//   1. which will loss values in EnvLayerCmd, currently we think is OK
//      (also consider remove the concept of EnvLayerCmd)
//   2. if we support async or parallel commands one day, this is not fit
//   3. (consider remove concept cc.GlobalEnv)
//
func (self *Cmd) executeFlow(argv ArgVals, cc *Cli, env *Env) bool {
	flow, _ := self.Flow(argv, env, false)
	return cc.Executor.Execute(self.owner.DisplayPath(), cc, flow...)
}

func (self *Cmd) executeFile(argv ArgVals, cc *Cli, env *Env, parsedCmd ParsedCmd) bool {
	if len(self.cmdLine) == 0 {
		return true
	}

	for _, dep := range self.depends {
		_, err := exec.LookPath(dep.OsCmd)
		if err != nil {
			// TODO: better display
			panic(NewCmdError(parsedCmd,
				fmt.Sprintf("[Cmd.executeFile] %s", err)))
		}
	}

	var bin string
	var args []string
	ext := filepath.Ext(self.cmdLine)

	// TODO: move this code block out?
	runner := env.Get("sys.ext.exec" + ext).Raw
	if len(runner) != 0 {
		fields := strings.Fields(runner)
		if len(fields) == 1 {
			bin = runner
		} else {
			bin = fields[0]
			args = append(args, fields[1:]...)
		}
	} else {
		bin = "bash"
	}

	sep := cc.Cmds.Strs.EnvKeyValSep

	sessionDir, sessionPath := saveEnvToSessionFile(cc, env, parsedCmd)

	args = append(args, self.cmdLine)
	args = append(args, sessionDir)
	for _, k := range self.args.Names() {
		args = append(args, argv[k].Raw)
	}
	cmd := exec.Command(bin, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		err = RunCmdFileFailed{
			err.Error(),
			parsedCmd,
			argv,
			bin,
			sessionPath,
		}
		panic(err)
	}

	LoadEnvFromFile(env.GetLayer(EnvLayerSession), sessionPath, sep)
	return true
}

// It's slow and ugly, but it should works fine
// only support one multiply definition in a line
func (self *Cmd) tryRenderMultiply(
	argv ArgVals,
	env *Env,
	in string,
	templBracketLeft string,
	templBracketRight string,
	templMultiplyMark string,
	allowFlowTemplateRenderError bool) (out []string, hasError bool) {

	out = []string{in}
	mm := templMultiplyMark + templMultiplyMark
	ml := templBracketLeft + templMultiplyMark
	mr := templMultiplyMark + templBracketRight

	valBegin := strings.Index(in, ml)
	if valBegin <= 0 {
		return
	}

	valEnd := strings.Index(in[valBegin:], mr)
	if valEnd <= 0 {
		return
	}
	valEnd += valBegin

	tempBegin := strings.Index(in[:valBegin], mm)
	if tempBegin < 0 {
		return
	}

	tempEnd := strings.Index(in[valEnd:], mm)
	if tempEnd < 0 {
		return
	}
	tempEnd += valEnd

	key := strings.TrimSpace(in[valBegin+len(ml) : valEnd])

	// TODO: duplicated code with normal flow template rendering
	var valStr string
	val, ok := env.GetEx(key)
	valStr = val.Raw
	if !ok {
		val, inArg := argv[key]
		valStr = val.Raw
		ok = inArg && len(valStr) != 0
	}
	if !ok {
		if allowFlowTemplateRenderError {
			hasError = true
			return
		}
		self.flowTemplateRenderPanic(key, true)
	}

	out = nil
	if tempBegin != 0 {
		out = append(out, strings.TrimSpace(in[:tempBegin]))
	}

	head := in[tempBegin+len(mm) : valBegin]
	tail := in[valEnd+len(mr) : tempEnd]

	listSep := self.owner.Strs.ListSep

	vals := strings.Split(valStr, listSep)
	for _, val := range vals {
		val = strings.TrimSpace(val)
		out = append(out, strings.TrimSpace(head+val+tail))
	}

	if tempEnd+len(mm) != len(in) {
		out = append(out, strings.TrimSpace(in[tempEnd+len(mm):]))
	}
	return
}

func (self *Cmd) flowTemplateRenderPanic(key string, isMultiply bool) {
	multiply := ""
	if isMultiply {
		multiply = "multiply "
	}

	findArgIdx := func(name string) int {
		idx := -1
		if len(name) == 0 {
			return idx
		}
		for i, it := range self.args.Names() {
			if it == name {
				idx = i
			}
		}
		return idx
	}

	if self.args.Has(key) {
		err := CmdMissedArgValWhenRenderFlow{
			"render flow " + multiply + "template failed, arg value missed.",
			self.owner.DisplayPath(),
			self.metaFilePath,
			self.owner.Source(),
			self,
			key,
			findArgIdx(key),
		}
		panic(err)
	} else {
		argName := self.arg2env.GetArgName(key)
		argIdx := findArgIdx(argName)
		err := CmdMissedEnvValWhenRenderFlow{
			"render flow " + multiply + "template failed, env value missed.",
			self.owner.DisplayPath(),
			self.metaFilePath,
			self.owner.Source(),
			key,
			self,
			argName,
			argIdx,
		}
		panic(err)
	}
}

func saveEnvToSessionFile(cc *Cli, env *Env, parsedCmd ParsedCmd) (sessionDir string, sessionPath string) {
	sep := cc.Cmds.Strs.EnvKeyValSep

	sessionDir = env.GetRaw("session")
	if len(sessionDir) == 0 {
		panic(NewCmdError(parsedCmd, "[Cmd.executeFile] session dir not found in env"))
	}
	sessionFileName := env.GetRaw("strs.session-env-file")
	if len(sessionFileName) == 0 {
		panic(NewCmdError(parsedCmd, "[Cmd.executeFile] session env file name not found in env"))
	}
	sessionPath = filepath.Join(sessionDir, sessionFileName)
	SaveEnvToFile(env.GetLayer(EnvLayerSession), sessionPath, sep)
	return
}

func mayQuoteStr(origin string) string {
	trimed := strings.TrimSpace(origin)
	if len(trimed) == 0 || len(trimed) != len(origin) {
		return "'" + origin + "'"
	}
	fields := strings.Fields(origin)
	if len(fields) != 1 {
		return "'" + origin + "'"
	}
	return origin
}

type Depend struct {
	OsCmd  string
	Reason string
}

type CmdError struct {
	Cmd ParsedCmd
	Err error
}

func WrapCmdError(cmd ParsedCmd, err error) *CmdError {
	return &CmdError{cmd, err}
}

func NewCmdError(cmd ParsedCmd, err string) *CmdError {
	return &CmdError{cmd, fmt.Errorf(err)}
}

func (self CmdError) Error() string {
	return self.Err.Error()
}
