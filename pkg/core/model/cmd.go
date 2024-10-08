package model

import (
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
	CmdTypeUninited CmdType = "uninited"
	// For internal test
	CmdTypeEmpty      CmdType = "empty"
	CmdTypeMetaOnly   CmdType = "no-executable"
	CmdTypeNormal     CmdType = "normal"
	CmdTypePower      CmdType = "power"
	CmdTypeFlow       CmdType = "flow"
	CmdTypeFileNFlow  CmdType = "executable-file+flow"
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

func (self AutoTimerKeys) IsEmpty() bool {
	return len(self.Begin) == 0 &&
		len(self.End) == 0 &&
		len(self.Dur) == 0
}

type CmdFlags struct {
	quiet               bool
	priority            bool
	allowTailModeCall   bool
	unLog               bool
	blender             bool
	quietError          bool
	noSession           bool
	hideInSessionsLast  bool
	quietSubFlow        bool
	exeInExecuted       bool
	ignoreFollowingDeps bool
	unbreakFileNFlow    bool
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
	orderedMacros []string
	macros        map[string][]string
	argsAutoMap   *ArgsAutoMapStatus
}

func defaultCmd(owner *CmdTree, help string) *Cmd {
	return &Cmd{
		owner:         owner,
		help:          help,
		ty:            CmdTypeUninited,
		flags:         &CmdFlags{},
		args:          newArgs(),
		normal:        nil,
		power:         nil,
		adhotFlow:     nil,
		cmdLine:       "",
		flow:          nil,
		envOps:        newEnvOps(),
		depends:       nil,
		metaFilePath:  "",
		val2env:       newVal2Env(),
		arg2env:       newArg2Env(),
		orderedMacros: []string{},
		macros:        map[string][]string{},
		argsAutoMap:   NewArgsAutoMapStatus(),
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

func NewMetaOnlyCmd(owner *CmdTree, help string) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypeMetaOnly
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
	allowError bool,
	currCmdIdx int,
	tryBreakInsideFileNFlow func(*Cli, *Env, *Cmd) bool) (newCurrCmdIdx int, ok bool) {

	if self.MustExeInExecuted() {
		mask = NewExecuteMask(self.owner.DisplayPath())
	}

	begin := time.Now()
	if len(self.autoTimerKeys.Begin) != 0 {
		env.GetLayer(EnvLayerSession).SetInt(self.autoTimerKeys.Begin, int(begin.Unix()))
	}

	newCurrCmdIdx, ok = self.execute(argv, cc, env, mask, flow, allowError, currCmdIdx, tryBreakInsideFileNFlow)
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

	return newCurrCmdIdx, ok || allowError
}

func (self *Cmd) execute(
	argv ArgVals,
	cc *Cli,
	env *Env,
	mask *ExecuteMask,
	flow *ParsedCmds,
	allowError bool,
	currCmdIdx int,
	tryBreakInsideFileNFlow func(*Cli, *Env, *Cmd) bool) (newCurrCmdIdx int, succeeded bool) {

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
				cc.FlowStatus.OnCmdFinish(flow, currCmdIdx, env, succeeded, err, !shouldExecByMask(mask) && !executedAndSucceeded(mask))
				cc.HandledErrors[r] = true
			}
			if r != nil && !self.flags.quietError && !allowError {
				panic(r)
			}
		}()
	}

	if mask != nil && mask.OverWriteFinishEnv != nil {
		p := env
		for p != nil && p.LayerType() != EnvLayerSession {
			p.CleanCurrLayer()
			p = p.Parent()
		}
		if p != nil {
			mask.OverWriteFinishEnv.WriteCurrLayerTo(p)
		}
	}

	disableQuietKey := "display.executor"
	envSession := env.GetLayer(EnvLayerSession)
	originQuiet := envSession.GetRaw(disableQuietKey)
	shouldQuietSubFlow := self.HasSubFlow(true) && self.flags.quietSubFlow
	if shouldQuietSubFlow {
		envSession.SetBool(disableQuietKey, false)
	}

	newCurrCmdIdx, succeeded = self.executeByType(argv, cc, env, mask, flow,
		allowError, currCmdIdx, logFilePath, tryBreakInsideFileNFlow)

	if shouldQuietSubFlow {
		envSession.Set(disableQuietKey, originQuiet)
	}
	return
}

func (self *Cmd) executeByType(
	argv ArgVals,
	cc *Cli,
	env *Env,
	mask *ExecuteMask,
	flow *ParsedCmds,
	allowError bool,
	currCmdIdx int,
	logFilePath string,
	tryBreakInsideFileNFlow func(*Cli, *Env, *Cmd) bool) (int, bool) {

	if !shouldExecByMask(mask) {
		// TODO: print this outside core pkg, so it can be colorize
		cc.Screen.Print("(skipped executing)\n")
		return currCmdIdx, true
	}

	switch self.ty {
	case CmdTypePower:
		return self.executePowerCmd(argv, cc, env, flow, currCmdIdx)
	case CmdTypeNormal:
		return currCmdIdx, self.normal(argv, cc, env, flow.Cmds[currCmdIdx:])
	case CmdTypeFile:
		return currCmdIdx, self.executeFile(argv, cc, env, allowError, flow.Cmds[currCmdIdx], logFilePath)
	case CmdTypeEmptyDir:
		return currCmdIdx, true
	case CmdTypeDirWithCmd:
		return currCmdIdx, self.executeFile(argv, cc, env, allowError, flow.Cmds[currCmdIdx], logFilePath)
	case CmdTypeFlow:
		return currCmdIdx, self.executeFlow(argv, cc, env, mask)
	case CmdTypeFileNFlow:
		return currCmdIdx, self.executeFileNFlow(argv, cc, env, allowError, flow.Cmds[currCmdIdx], logFilePath, mask, tryBreakInsideFileNFlow)
	case CmdTypeAdHotFlow:
		return currCmdIdx, self.executeFlow(argv, cc, env, mask)
	case CmdTypeEmpty:
		return currCmdIdx, true
	case CmdTypeMetaOnly:
		return currCmdIdx, true
	default:
		panic(NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("[Cmd.Execute] unknown cmd executable type: %v", self.ty)))
	}
}

func (self *Cmd) HasSubFlow(includeQuietSubFlow bool) bool {
	if !includeQuietSubFlow && self.flags.quietSubFlow {
		return false
	}
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
		//arg := self.arg2env.GetArgName(self, key, false)
		//return len(self.args.DefVal(arg)) != 0
		return true
	}
	return false
}

func (self *Cmd) SetIsApi() *Cmd {
	self.owner.SetIsApi()
	return self
}

func (self *Cmd) AddArg(name string, defVal string, abbrs ...string) *Cmd {
	self.args.AddArg(self.owner, name, defVal, abbrs...)
	return self
}

func (self *Cmd) SetArgEnums(name string, vals ...string) *Cmd {
	self.args.SetArgEnums(self.owner, name, vals...)
	return self
}

func (self *Cmd) AddAutoMapAllArg(name string, defVal string, abbrs ...string) *Cmd {
	self.args.AddAutoMapAllArg(self.owner, name, defVal, abbrs...)
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

func (self *Cmd) AddMacro(name string, flow []string) {
	self.orderedMacros = append(self.orderedMacros, name)
	self.macros[name] = flow
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

func (self *Cmd) SetIgnoreFollowingDeps() *Cmd {
	self.flags.ignoreFollowingDeps = true
	return self
}

func (self *Cmd) SetQuiet() *Cmd {
	self.flags.quiet = true
	return self
}

func (self *Cmd) SetExeInExecuted() *Cmd {
	self.flags.exeInExecuted = true
	return self
}

func (self *Cmd) SetNoSession() *Cmd {
	self.flags.noSession = true
	return self
}

func (self *Cmd) SetHideInSessionsLast() *Cmd {
	self.flags.hideInSessionsLast = true
	return self
}

func (self *Cmd) SetQuietSubFlow() *Cmd {
	self.flags.quietSubFlow = true
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

func (self *Cmd) SetArg2EnvAutoMap(names []string) *Cmd {
	self.argsAutoMap.AddDefinitions(self, names...)
	return self
}

func (self *Cmd) GetArgsAutoMapStatus() *ArgsAutoMapStatus {
	return self.argsAutoMap
}

func (self *Cmd) ReorderArgs(names []string) {
	self.args.Reorder(self, names)
}

func (self *Cmd) IsBlenderCmd() bool {
	return self.flags.blender
}

func (self *Cmd) SetUnLog() *Cmd {
	self.flags.unLog = true
	return self
}

func (self *Cmd) SetQuietError() *Cmd {
	self.flags.quietError = true
	return self
}

func (self *Cmd) SetUnbreakFileNFlow() *Cmd {
	self.flags.unbreakFileNFlow = true
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

func (self *Cmd) AddArg2EnvFromAnotherCmd(src *Cmd) {
	if self.argsAutoMap.IsEmpty() {
		return
	}
	srcMapper := src.GetArg2Env()
	srcArgs := src.Args()

	// TODO: EnvKeys => RenderedEnvKeys?
	for _, key := range srcMapper.EnvKeys() {
		srcArgName := srcMapper.GetArgName(src, key, false)
		if self.arg2env.Has(key) {
			continue
		}
		srcArgAbbrs := srcArgs.Abbrs(srcArgName)
		mapArgName, shouldMap, shouldMarkMapped := self.argsAutoMap.ShouldMapByDefinition(self, src, srcArgAbbrs)
		if !shouldMap {
			continue
		}

		defVal, newAbbrs, ok := self.checkCanAddArgFromAnotherArg(srcArgs, mapArgName)
		if ok {
			self.argsAutoMap.MarkAndCacheMapping(src, key, mapArgName, defVal, newAbbrs, shouldMarkMapped)
		}
	}
}

func (self *Cmd) FinishArg2EnvAutoMap(cc *Cli) {
	self.argsAutoMap.FlushCache(self)
	if !self.argsAutoMap.FullyMappedOrMapAll() {
		err := fmt.Errorf("args of cmd '%s' can't be fully mapped: %s",
			self.owner.DisplayPath(), strings.Join(self.argsAutoMap.GetUnmappedArgs(), ","))
		cc.TolerableErrs.OnErr(err, self.owner.Source(), self.metaFilePath, "arg2env auto mapping failed")
	}
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
	return self.ty == CmdTypeNormal || self.ty == CmdTypeEmpty || self.ty == CmdTypePower
}

func (self *Cmd) HasCmdLine() bool {
	return len(self.cmdLine) != 0
}

func (self *Cmd) IsTotallyEmpty() bool {
	return (self.ty == CmdTypeEmptyDir || self.ty == CmdTypeEmpty || self.ty == CmdTypeMetaOnly) &&
		self.args.IsEmpty() &&
		self.arg2env.IsEmpty() &&
		self.val2env.IsEmpty() &&
		self.autoTimerKeys.IsEmpty() &&
		self.envOps.IsEmpty() &&
		self.argsAutoMap.IsEmpty() &&
		len(self.macros) == 0 &&
		len(self.orderedMacros) == 0 &&
		len(self.cmdLine) == 0 &&
		len(self.flow) == 0 &&
		len(self.depends) == 0 &&
		self.adhotFlow == nil &&
		self.power == nil &&
		self.normal == nil
}

func (self *Cmd) IsNoExecutableCmd() bool {
	if len(self.val2env.EnvKeys()) > 0 {
		return false
	}
	if !self.arg2env.IsEmpty() {
		return false
	}
	return self.ty == CmdTypeUninited || self.ty == CmdTypeEmpty || self.ty == CmdTypeMetaOnly || self.ty == CmdTypeEmptyDir
}

func (self *Cmd) IsPowerCmd() bool {
	return self.ty == CmdTypePower
}

func (self *Cmd) AllowTailModeCall() bool {
	return self.flags.allowTailModeCall
}

func (self *Cmd) MustExeInExecuted() bool {
	return self.flags.exeInExecuted
}

func (self *Cmd) ShouldIgnoreFollowingDeps() bool {
	return self.flags.ignoreFollowingDeps
}

func (self *Cmd) IsQuiet() bool {
	return self.flags.quiet
}

func (self *Cmd) IsNoSessionCmd() bool {
	return self.flags.noSession
}

func (self *Cmd) IsHideInSessionsLast() bool {
	return self.flags.hideInSessionsLast
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

func (self *Cmd) ShouldUnbreakFileNFlow() bool {
	return self.flags.unbreakFileNFlow
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
	if self.flags.unLog || self.flags.noSession {
		return false
	}
	return self.ty == CmdTypeFile ||
		self.ty == CmdTypeDirWithCmd ||
		self.ty == CmdTypeFileNFlow
}

func (self *Cmd) genLogFileName() string {
	name := self.owner.DisplayPath()
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%s_%s_%v", time.Now().Format("20060102-150405"), name, rand.Uint32())
}

func (self *Cmd) renderMacros(
	argv ArgVals,
	env *Env,
	allowFlowTemplateRenderError bool) (envWithMacros *Env, fullyRendered bool) {

	env = env.NewLayer(EnvLayerTmp)
	fullyRendered = true
	sep := env.GetRaw("strs.seq-sep")
	for _, macro := range self.orderedMacros {
		macroStrs := self.macros[macro]
		macroFlow, macroFullyRendered := RenderTemplateStrLines(macroStrs, "macro", self, argv, env, allowFlowTemplateRenderError)
		macroFlow = StripFlowForExecute(macroFlow, sep)
		env.Set(macro, FlowStrsToStr(macroFlow))
		fullyRendered = fullyRendered && macroFullyRendered
	}
	return env, fullyRendered
}

// TODO: move to parser ?
// TODO: read and parse session file too many times for CmdTypeAdHotFlow, cache it
func (self *Cmd) RenderedFlowStrs(
	argv ArgVals,
	cc *Cli,
	env *Env,
	allowFlowTemplateRenderError bool,
	forChecking bool) (flow []string, masks []*ExecuteMask, fullyRendered bool) {

	if forChecking {
		cc = cc.CloneForChecking()
	}

	// Render macros into tmp env
	env, macrosFullyRendered := self.renderMacros(argv, env, allowFlowTemplateRenderError)

	flowStrs := self.flow
	if self.ty == CmdTypeAdHotFlow {
		if len(flowStrs) != 0 {
			panic(fmt.Errorf("[Cmd.RenderedFlowStrs] should never happend: ad-hot flow has fixed flow-strings"))
		}
		flowStrs, masks, _ = self.adhotFlow(argv, cc, env)
	}

	flow, flowFullyRendered := RenderTemplateStrLines(flowStrs, "flow", self, argv, env, allowFlowTemplateRenderError)
	fullyRendered = macrosFullyRendered && flowFullyRendered
	if fullyRendered {
		if masks != nil && !cc.Blender.IsEmpty() {
			panic(fmt.Errorf("[Cmd.RenderedFlowStrs] can't use blender on a masked flow"))
		}
		flow = self.invokeBlender(cc, env, flow)
	}
	return
}

// TODO: extra parse/save for a flow, slow and ugly
func (self *Cmd) invokeBlender(cc *Cli, env *Env, input []string) []string {
	input = normalizeInput(input, env.GetRaw("strs.seq-sep"))
	flow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, input...)
	cc.Blender.Invoke(cc, env, flow)
	trivialMark := env.GetRaw("strs.trivial-mark")
	flowStr := SaveFlowToStr(flow, cc.Cmds.Strs.PathSep, trivialMark, env)
	return []string{flowStr}
}

// TODO: this is a bit confusing, when we need this and when we dont, fix it
func normalizeInput(input []string, sequenceSep string) []string {
	input = StripFlowForExecute(input, sequenceSep)
	return FlowStrToStrs(FlowStrsToStr(input))
}

func StripFlowForExecute(flow []string, sequenceSep string) []string {
	var output []string
	for _, line := range flow {
		strings.TrimSpace(line)
		if len(line) <= 0 {
			continue
		}
		output = append(output, line)
	}

	flow = output
	output = []string{}
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

func (self *Cmd) Flow(argv ArgVals, cc *Cli, env *Env,
	allowFlowTemplateRenderError bool, forChecking bool) (flow []string, masks []*ExecuteMask, rendered bool) {

	flow, masks, rendered = self.RenderedFlowStrs(argv, cc, env, allowFlowTemplateRenderError, forChecking)
	if len(flow) == 0 {
		return
	}
	if !rendered {
		//return
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
	flow, masks, _ := self.Flow(argv, cc, env, false, false)
	flowStr := FlowStrsToStr(flow)
	flowEnv := env.NewLayer(EnvLayerSubFlow)
	skipped := false

	flowStrEnvVal := flowStr
	// TODO: calculate the min width properly
	if len(flowStrEnvVal) > 70 {
		flowStrEnvVal = flowStrEnvVal[:33] + "...." + flowStrEnvVal[len(flowStrEnvVal)-34:]
	}
	flowEnv.Set("sys.subflow", flowStrEnvVal)

	if cc.FlowStatus != nil {
		cc.FlowStatus.OnSubFlowStart(flowEnv, flowStr)
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
		// TODO: print this outside core pkg, so it can be colorize
		cc.Screen.Print("(skipped subflow\n")
		succeeded = true
		skipped = !executedAndSucceeded(mask)
	}
	return
}

func (self *Cmd) executeFileNFlow(argv ArgVals, cc *Cli, env *Env, allowError bool, parsedCmd ParsedCmd,
	logFilePath string, mask *ExecuteMask, tryBreakInsideFileNFlow func(*Cli, *Env, *Cmd) bool) (succeeded bool) {

	succeeded = self.executeFlow(argv, cc, env, mask)
	if !succeeded {
		return false
	}

	// TODO: user will feel a bit weird when FileNFlowExecPolicy is skip
	if tryBreakInsideFileNFlow == nil || tryBreakInsideFileNFlow(cc, env, self) {
		if mask == nil || mask.FileNFlowExecPolicy == ExecPolicyExec {
			return self.executeFile(argv, cc, env, allowError, parsedCmd, logFilePath)
		}
	}

	// TODO: print this outside core pkg, so it can be colorize
	cc.Screen.Print("(skipped executing after subflow)\n")
	return true
}

func (self *Cmd) executeFile(argv ArgVals, cc *Cli, env *Env, allowError bool, parsedCmd ParsedCmd, logFilePath string) bool {
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

	var sessionDir string
	var sessionPath string
	if !self.flags.noSession {
		sessionDir, sessionPath = saveEnvToSessionFile(cc, env, parsedCmd, false)
	}

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
	if err != nil && !allowError {
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

	if len(sessionPath) != 0 {
		LoadEnvFromFile(env.GetLayer(EnvLayerSession), sessionPath, sep, delMark)
	}
	return true
}

func (self *Cmd) checkCanAddArgFromAnotherArg(srcArgs Args, name string) (defVal string, abbrs []string, ok bool) {
	if srcArgs.IsFromAutoMapAll(name) {
		return
	}
	if self.args.Has(name) {
		return
	}
	if self.args.HasArgOrAbbr(name) {
		return
	}
	var newAbbrs []string
	realname := srcArgs.Realname(name)
	for _, abbr := range srcArgs.Abbrs(realname) {
		if abbr == name {
			continue
		}
		if self.args.Has(abbr) {
			continue
		}
		abbrReal := self.args.Realname(abbr)
		if len(abbrReal) != 0 {
			continue
		}
		newAbbrs = append(newAbbrs, abbr)
	}
	return srcArgs.RawDefVal(realname), newAbbrs, true
}

func shouldExecByMask(mask *ExecuteMask) bool {
	return (mask == nil || mask.ExecPolicy == ExecPolicyExec)
}

func executedAndSucceeded(mask *ExecuteMask) bool {
	return (mask != nil || mask.ResultIfExecuted == ExecutedResultSucceeded)
}
