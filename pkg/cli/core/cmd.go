package core

import (
	"bytes"
	"fmt"
	"math/rand"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"time"

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
	CmdTypeAdHotFlow  CmdType = "adhot-flow"
)

type NormalCmd func(argv ArgVals, cc *Cli, env *Env, flow []ParsedCmd) (succeeded bool)

type PowerCmd func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds,
	currCmdIdx int) (newCurrCmdIdx int, succeeded bool)

type AdHotFlowCmd func(argv ArgVals, cc *Cli, env *Env) (flow []string, masks []*ExecuteMask, succeeded bool)

type Depend struct {
	OsCmd  string
	Reason string
}

type AutoTimerKeys struct {
	Begin string
	End   string
	Dur   string
}

type CmdFlags struct {
	quiet             bool
	priority          bool
	allowTailModeCall bool
	unLog             bool
	blender           bool
}

type Cmd struct {
	owner         *CmdTree
	help          string
	ty            CmdType
	flags         *CmdFlags
	args          Args
	normal        NormalCmd
	power         PowerCmd
	adhotFlow     AdHotFlowCmd
	cmdLine       string
	flow          []string
	envOps        EnvOps
	depends       []Depend
	metaFilePath  string
	val2env       *Val2Env
	arg2env       *Arg2Env
	autoTimerKeys AutoTimerKeys
}

func defaultCmd(owner *CmdTree, help string) *Cmd {
	return &Cmd{
		owner:        owner,
		help:         help,
		ty:           CmdTypeUninited,
		flags:        &CmdFlags{},
		args:         newArgs(),
		normal:       nil,
		power:        nil,
		adhotFlow:    nil,
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

func NewAdHotFlowCmd(owner *CmdTree, help string, adhotFlow AdHotFlowCmd) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypeAdHotFlow
	c.adhotFlow = adhotFlow
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
	mask *ExecuteMask,
	flow *ParsedCmds,
	currCmdIdx int) (newCurrCmdIdx int, ok bool) {

	begin := time.Now()
	if len(self.autoTimerKeys.Begin) != 0 {
		env.GetLayer(EnvLayerSession).SetInt(self.autoTimerKeys.Begin, int(begin.Unix()))
	}

	newCurrCmdIdx, ok = self.execute(argv, cc, env, mask, flow, currCmdIdx)
	if !ok {
		// Normally the command should print info before return false, so no need to panic
		// panic(NewCmdError(flow.Cmds[currCmdIdx], "command failed without detail info"))
	}

	end := time.Now()
	if len(self.autoTimerKeys.End) != 0 {
		env.GetLayer(EnvLayerSession).SetInt(self.autoTimerKeys.End, int(end.Unix()))
	}
	if len(self.autoTimerKeys.Dur) != 0 {
		env.GetLayer(EnvLayerSession).SetInt(self.autoTimerKeys.Dur, int(end.Sub(begin)/time.Second))
	}

	return newCurrCmdIdx, ok
}

func (self *Cmd) execute(
	argv ArgVals,
	cc *Cli,
	env *Env,
	mask *ExecuteMask,
	flow *ParsedCmds,
	currCmdIdx int) (newCurrCmdIdx int, succeeded bool) {

	// TODO: this logic should be in upper layer
	if mask != nil && mask.OverWriteStartEnv != nil {
		p := env
		for p != nil && p.LayerType() != EnvLayerSession {
			p.CleanCurrLayer()
			p = p.Parent()
		}
		if p != nil {
			mask.OverWriteStartEnv.WriteCurrLayerTo(p)
		}
	}

	logFilePath := self.genLogFilePath(env)

	if cc.FlowStatus != nil {
		cc.FlowStatus.OnCmdStart(flow, currCmdIdx, env, logFilePath)
		defer func() {
			r := recover()
			handledErr := false
			var err error
			isAbort := false
			if r != nil {
				handledErr = cc.HandledErrors[r]
				err = r.(error)
				_, isAbort = err.(*AbortByUserErr)
			}

			if (r == nil || !handledErr) && !isAbort {
				cc.FlowStatus.OnCmdFinish(flow, currCmdIdx, env, succeeded, err, !shouldExecByMask(mask))
				cc.HandledErrors[r] = true
			}
			if r != nil {
				panic(r)
			}
		}()
	}

	if !shouldExecByMask(mask) {
		// TODO: print this outside core pkg, so it can be colorize
		cc.Screen.Print("(skipped)\n")
		newCurrCmdIdx, succeeded = currCmdIdx, true
	} else {
		newCurrCmdIdx, succeeded = self.executeByType(argv, cc, env, mask, flow, currCmdIdx, logFilePath)
	}
	return
}

func (self *Cmd) executeByType(
	argv ArgVals,
	cc *Cli,
	env *Env,
	mask *ExecuteMask,
	flow *ParsedCmds,
	currCmdIdx int,
	logFilePath string) (int, bool) {

	switch self.ty {
	case CmdTypePower:
		return self.executePowerCmd(argv, cc, env, flow, currCmdIdx)
	case CmdTypeNormal:
		return currCmdIdx, self.normal(argv, cc, env, flow.Cmds[currCmdIdx:])
	case CmdTypeFile:
		return currCmdIdx, self.executeFile(argv, cc, env, flow.Cmds[currCmdIdx], logFilePath)
	case CmdTypeEmptyDir:
		return currCmdIdx, true
	case CmdTypeDirWithCmd:
		return currCmdIdx, self.executeFile(argv, cc, env, flow.Cmds[currCmdIdx], logFilePath)
	case CmdTypeFlow:
		return currCmdIdx, self.executeFlow(argv, cc, env, mask)
	case CmdTypeFileNFlow:
		succeeded := self.executeFlow(argv, cc, env, mask)
		if succeeded && shouldExecByMask(mask) {
			succeeded = self.executeFile(argv, cc, env, flow.Cmds[currCmdIdx], logFilePath)
		}
		return currCmdIdx, succeeded
	case CmdTypeAdHotFlow:
		return currCmdIdx, self.executeFlow(argv, cc, env, mask)
	case CmdTypeEmpty:
		return currCmdIdx, true
	default:
		panic(NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("[Cmd.Execute] unknown cmd executable type: %v", self.ty)))
	}
}

func (self *Cmd) HasSubFlow() bool {
	return self.ty == CmdTypeFlow || self.ty == CmdTypeFileNFlow || self.ty == CmdTypeAdHotFlow
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
	if self.flags.quiet && strings.Index("quiet", findStr) >= 0 {
		return true
	}
	if self.ty == CmdTypePower && strings.Index("power", findStr) >= 0 {
		return true
	}
	if self.flags.priority && strings.Index("priority", findStr) >= 0 {
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

func (self *Cmd) RegAutoTimerBeginKey(key string) {
	self.autoTimerKeys.Begin = key
	self.AddEnvOp(key, EnvOpTypeWrite)
}

func (self *Cmd) RegAutoTimerEndKey(key string) {
	self.autoTimerKeys.End = key
	self.AddEnvOp(key, EnvOpTypeWrite)
}

func (self *Cmd) RegAutoTimerDurKey(key string) {
	self.autoTimerKeys.Dur = key
	self.AddEnvOp(key, EnvOpTypeWrite)
}

func (self *Cmd) GetAutoTimerKeys() AutoTimerKeys {
	return self.autoTimerKeys
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
	self.flags.quiet = true
	return self
}

func (self *Cmd) SetAllowTailModeCall() *Cmd {
	self.flags.allowTailModeCall = true
	return self
}

func (self *Cmd) SetIsBlenderCmd() *Cmd {
	self.flags.blender = true
	self.flags.unLog = true
	return self
}

func (self *Cmd) IsBlenderCmd() bool {
	return self.flags.blender
}

func (self *Cmd) SetUnLog() *Cmd {
	self.flags.unLog = true
	return self
}

func (self *Cmd) SetPriority() *Cmd {
	self.flags.priority = true
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
	if len(self.help) == 0 && (self.ty == CmdTypeFlow || self.ty == CmdTypeFileNFlow) {
		return self.cmdLine
	}
	return self.help
}

func (self *Cmd) IsBuiltinCmd() bool {
	return self.ty == CmdTypeNormal || self.ty == CmdTypePower
}

func (self *Cmd) HasCmdLine() bool {
	return len(self.cmdLine) != 0
}

func (self *Cmd) IsNoExecutableCmd() bool {
	if len(self.val2env.EnvKeys()) > 0 {
		return false
	}
	if !self.arg2env.IsEmpty() {
		return false
	}
	return self.ty == CmdTypeUninited || self.ty == CmdTypeEmpty || self.ty == CmdTypeEmptyDir
}

func (self *Cmd) IsPowerCmd() bool {
	return self.ty == CmdTypePower
}

func (self *Cmd) AllowTailModeCall() bool {
	return self.flags.allowTailModeCall
}

func (self *Cmd) IsQuiet() bool {
	return self.flags.quiet
}

func (self *Cmd) IsPriority() bool {
	return self.flags.priority
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
	if self.adhotFlow != nil {
		fr2 := reflect.ValueOf(self.adhotFlow)
		if fr1.Pointer() == fr2.Pointer() {
			return true
		}
	}
	return false
}

func (self *Cmd) genLogFilePath(env *Env) string {
	if !self.shouldWriteLogFile() {
		return ""
	}
	// TODO: move this logic to upper layer(executor)
	if env.GetBool("sys.interact.inside") {
		return ""
	}
	sessionDir := env.GetRaw("session")
	if len(sessionDir) == 0 {
		panic(fmt.Errorf("[Cmd.genLogFilePath] session dir not found in env"))
	}
	fileName := self.genLogFileName()
	return filepath.Join(sessionDir, fileName)
}

func (self *Cmd) shouldWriteLogFile() bool {
	if self.flags.unLog {
		return false
	}
	return self.ty == CmdTypeFile ||
		self.ty == CmdTypeDirWithCmd ||
		self.ty == CmdTypeFileNFlow
}

func (self *Cmd) genLogFileName() string {
	name := self.owner.DisplayPath()
	return fmt.Sprintf("%s_%s_%v", time.Now().Format("20060102-150405"), name, rand.Uint32())
}

// TODO: move to parser ?
// TODO: read and parse session file too many times for CmdTypeAdHotFlow, cache it
func (self *Cmd) RenderedFlowStrs(
	argv ArgVals,
	cc *Cli,
	env *Env,
	allowFlowTemplateRenderError bool) (flow []string, masks []*ExecuteMask, fullyRendered bool) {

	fullyRendered = true

	flowStrs := self.flow
	if self.ty == CmdTypeAdHotFlow {
		flowStrs, masks, _ = self.adhotFlow(argv, cc, env)
	}

	for _, line := range flowStrs {
		rendereds, lineFullyRendered := renderTemplateStr(line, "flow", self, argv, env, allowFlowTemplateRenderError)
		for _, rendered := range rendereds {
			flow = append(flow, rendered)
		}
		fullyRendered = fullyRendered && lineFullyRendered
	}

	if fullyRendered {
		if masks != nil {
			panic(fmt.Errorf("[Cmd.RenderedFlowStrs] can't use blender on a masked flow"))
		}
		flow = self.invokeBlender(cc, env, flow)
	}
	return
}

// TODO: extra parse/save for a flow, slow and ugly
func (self *Cmd) invokeBlender(cc *Cli, env *Env, input []string) []string {
	flow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, input...)
	cc.Blender.Invoke(cc, env, flow)
	trivialMark := env.GetRaw("strs.trivial-mark")
	w := bytes.NewBuffer(nil)
	SaveFlow(w, flow, cc.Cmds.Strs.PathSep, trivialMark, env)
	flowStr := w.String()
	return FlowStrToStrs(flowStr)
}

func StripFlowForExecute(flow []string, sequenceSep string) []string {
	var output []string
	for i, line := range flow {
		strings.TrimSpace(line)
		if len(line) <= 0 {
			continue
		}
		// TODO: put # into env.strs
		if line[0] == '#' {
			continue
		}
		if line[len(line)-1:] != sequenceSep && i != len(flow)-1 {
			line += " " + sequenceSep
		}
		output = append(output, line)
	}
	return output
}

func (self *Cmd) Flow(argv ArgVals, cc *Cli, env *Env, allowFlowTemplateRenderError bool) (flow []string, masks []*ExecuteMask, rendered bool) {
	flow, masks, rendered = self.RenderedFlowStrs(argv, cc, env, allowFlowTemplateRenderError)
	if !rendered || len(flow) == 0 {
		return
	}
	flow = StripFlowForExecute(flow, env.GetRaw("strs.seq-sep"))
	flowStr := FlowStrsToStr(flow)
	flow = FlowStrToStrs(flowStr)
	return
}

func FlowStrToStrs(flowStr string) []string {
	flowStrs, err := shellwords.Parse(flowStr)
	if err != nil {
		// TODO: better display
		panic(fmt.Errorf("[shellwords] parse '%s' failed: %v",
			flowStr, err))
	}
	return flowStrs
}

func FlowStrsToStr(flowStrs []string) string {
	return strings.Join(flowStrs, " ")
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

// TODO: flow must not have argv, is it OK?
func (self *Cmd) executeFlow(argv ArgVals, cc *Cli, env *Env, mask *ExecuteMask) (succeeded bool) {
	flow, masks, _ := self.Flow(argv, cc, env, false)
	flowEnv := env.NewLayer(EnvLayerSubFlow)
	skipped := false

	if cc.FlowStatus != nil {
		cc.FlowStatus.OnSubFlowStart(FlowStrsToStr(flow))
		defer func() {
			if succeeded {
				cc.FlowStatus.OnSubFlowFinish(flowEnv, succeeded, skipped)
			}
		}()
	}
	if mask != nil && len(mask.SubFlow) != 0 {
		if len(masks) != 0 {
			// TODO: handle masks confliction
		}
		masks = mask.SubFlow
	}
	if shouldExecByMask(mask) {
		succeeded = cc.Executor.Execute(self.owner.DisplayPath(), true, cc, flowEnv, masks, flow...)
	} else {
		cc.Screen.Print("(skipped+)\n")
		succeeded = true
		skipped = true
	}
	return
}

func (self *Cmd) executeFile(argv ArgVals, cc *Cli, env *Env, parsedCmd ParsedCmd, logFilePath string) bool {
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
	delMark := cc.Cmds.Strs.EnvValDelAllMark

	sessionDir, sessionPath := saveEnvToSessionFile(cc, env, parsedCmd, false)

	args = append(args, self.cmdLine)
	args = append(args, sessionDir)
	for _, k := range self.args.Names() {
		args = append(args, argv[k].Raw)
	}
	cmd := exec.Command(bin, args...)
	cmd.Dir = filepath.Dir(self.cmdLine)

	logger := cc.CmdIO.SetupForExec(cmd, logFilePath)
	if logger != nil {
		defer logger.Close()
	}

	err := cmd.Run()
	if err != nil {
		err = &RunCmdFileFailed{
			err.Error(),
			parsedCmd,
			argv,
			bin,
			sessionPath,
			logFilePath,
		}
		if logger != nil {
			logger.Close()
		}
		panic(err)
	}

	LoadEnvFromFile(env.GetLayer(EnvLayerSession), sessionPath, sep, delMark)
	return true
}

func shouldExecByMask(mask *ExecuteMask) bool {
	return (mask == nil || mask.ExecPolicy == ExecPolicyExec)
}
