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
	CmdTypeFileNFlow  CmdType = "executable-file+flow"
	CmdTypeEmpty      CmdType = "no-executable"
	CmdTypeFile       CmdType = "executable-file"
	CmdTypeEmptyDir   CmdType = "dir-with-no-executable"
	CmdTypeDirWithCmd CmdType = "dir-with-executable-file"
)

type NormalCmd func(argv ArgVals, cc *Cli, env *Env, flow []ParsedCmd) (succeeded bool)
type PowerCmd func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds,
	currCmdIdx int) (newCurrCmdIdx int, succeeded bool)

type Depend struct {
	OsCmd  string
	Reason string
}

type Cmd struct {
	owner             *CmdTree
	help              string
	ty                CmdType
	quiet             bool
	priority          bool
	allowTailModeCall bool
	args              Args
	normal            NormalCmd
	power             PowerCmd
	cmdLine           string
	flow              []string
	envOps            EnvOps
	depends           []Depend
	metaFilePath      string
	val2env           *Val2Env
	arg2env           *Arg2Env
}

func defaultCmd(owner *CmdTree, help string) *Cmd {
	return &Cmd{
		owner:             owner,
		help:              help,
		ty:                CmdTypeUninited,
		quiet:             false,
		priority:          false,
		allowTailModeCall: false,
		args:              newArgs(),
		normal:            nil,
		power:             nil,
		cmdLine:           "",
		flow:              nil,
		envOps:            newEnvOps(),
		depends:           nil,
		metaFilePath:      "",
		val2env:           newVal2Env(),
		arg2env:           newArg2Env(),
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

func NewFileNFlowCmd(owner *CmdTree, help string, cmd string, flow []string) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypeFileNFlow
	c.cmdLine = cmd
	c.flow = flow
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
		// Normally the command should print info before return false, so no need to panic
		// panic(NewCmdError(flow.Cmds[currCmdIdx], "command failed without detail info"))
	}
	return newCurrCmdIdx, ok
}

func (self *Cmd) executePowerCmd(
	argv ArgVals,
	cc *Cli,
	env *Env,
	flow *ParsedCmds,
	currCmdIdx int) (int, bool) {

	currCmdIdx, succeeded := self.power(argv, cc, env, flow, currCmdIdx)
	// Let commands manually clear it when it's tail-mode flow(not call),
	// in that we could run tail-mode recursively
	if flow.TailModeCall {
		currCmdIdx = 0
		flow.Cmds = nil
	}
	return currCmdIdx, succeeded
}

func (self *Cmd) execute(
	argv ArgVals,
	cc *Cli,
	env *Env,
	flow *ParsedCmds,
	currCmdIdx int) (int, bool) {

	switch self.ty {
	case CmdTypePower:
		return self.executePowerCmd(argv, cc, env, flow, currCmdIdx)
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
	case CmdTypeFileNFlow:
		succeeded := self.executeFlow(argv, cc, env)
		if succeeded {
			succeeded = self.executeFile(argv, cc, env, flow.Cmds[currCmdIdx])
		}
		return currCmdIdx, succeeded
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

// TODO: more flags to filter the result, too much
func (self *Cmd) MatchWriteKey(key string) bool {
	if self.envOps.MatchWriteKey(key) {
		return true
	}
	if self.val2env.Has(key) {
		return true
	}
	if self.arg2env.Has(key) {
		arg := self.arg2env.GetArgName(self, key, false)
		return len(self.args.DefVal(arg)) != 0
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

func (self *Cmd) SetAllowTailModeCall() *Cmd {
	self.allowTailModeCall = true
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

func (self *Cmd) AllowTailModeCall() bool {
	return self.allowTailModeCall
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

	fullyRendered = true

	for _, line := range self.flow {
		rendereds, lineFullyRendered := renderTemplateStr(line, "flow", self, argv, env, allowFlowTemplateRenderError)
		for _, rendered := range rendereds {
			flow = append(flow, rendered)
		}
		fullyRendered = fullyRendered && lineFullyRendered
	}
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
