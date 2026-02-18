package model

import (
	"fmt"
	"sort"
	"strings"
)

type Arg2EnvAutoMapCmds map[*Cmd]bool

func (self Arg2EnvAutoMapCmds) AddAutoMapTarget(cmd *Cmd) {
	self[cmd] = true
}

func (self Arg2EnvAutoMapCmds) AutoMapArg2Env(cc *Cli, env *Env, envOpCmds []EnvOpCmd, stackDepth int) {
	sep := env.GetRaw("strs.cmd-path-sep")
	for cmd := range self {
		argv := env.GetArgv(cmd.Owner().Path(), sep, stackDepth, cmd.Args())
		autoMapArg2EnvForCmd(cc, env.Clone(), cmd, argv, envOpCmds, cmd, stackDepth)
		cmd.FinishArg2EnvAutoMap(cc)
	}
}

func autoMapArg2EnvForCmd(
	cc *Cli,
	env *Env,
	srcCmd *Cmd,
	argv ArgVals,
	envOpCmds []EnvOpCmd,
	targetCmd *Cmd,
	stackDepth int) (done bool) {

	targetCmd.GetArgsAutoMapStatus().MarkMetWithArgv(srcCmd, argv)
	if !srcCmd.HasSubFlow(true) {
		return false
	}

	env = env.Clone().GetLayer(EnvLayerSession)
	ApplyVal2Env(env, srcCmd, argv)
	ApplyArg2Env(env, srcCmd, argv)

	subFlow, _, _ := srcCmd.Flow(argv, cc, env, true, true)
	if len(subFlow) == 0 {
		// TODO: careful test: handle not fully renderd subflow
		//return false
		_ = subFlow
	}

	parsedFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
	flowEnv := env.NewLayer(EnvLayerSubFlow)
	parsedFlow.GlobalEnv.WriteNotArgTo(flowEnv, cc.Cmds.Strs.EnvValDelAllMark)

	for key := range parsedFlow.GlobalEnv {
		targetCmd.GetArgsAutoMapStatus().RecordGlobalEnvKey(key)
	}

	return autoMapArg2EnvForCmdsInFlow(cc, flowEnv, parsedFlow, 0, envOpCmds, targetCmd, stackDepth)
}

func autoMapArg2EnvForCmdsInFlow(
	cc *Cli,
	env *Env,
	flow *ParsedCmds,
	currCmdIdx int,
	envOpCmds []EnvOpCmd,
	targetCmd *Cmd,
	stackDepth int) (done bool) {

	for i := currCmdIdx; i < len(flow.Cmds); i++ {
		it := flow.Cmds[i]
		cic := it.LastCmd()
		if cic == nil {
			continue
		}
		if it.ParseResult.Error != nil && !it.ParseResult.IsMinorErr {
			targetCmd.GetArgsAutoMapStatus().MarkMet(cic)
			continue
		}
		if targetCmd.GetArgsAutoMapStatus().ShouldSkip(cic) {
			continue
		}

		cmdEnv, argv := it.ApplyMappingGenEnvAndArgv(env, cc.Cmds.Strs.EnvValDelAllMark,
			cc.Cmds.Strs.PathSep, stackDepth)

		if autoMapArg2EnvForCmd(cc, cmdEnv, cic, argv, envOpCmds, targetCmd, stackDepth+1) {
			return true
		}

		var checker EnvOpsChecker
		checker = EnvOpsChecker{}
		TryExeEnvOpCmds(argv, cc, cmdEnv, flow, i, envOpCmds, &checker,
			"failed to execute env op-cmd in depends collecting")

		for key := range checker {
			targetCmd.GetArgsAutoMapStatus().RecordGlobalEnvKey(key)
		}

		//println(cic.Owner().DisplayPath(), " (arg2env) =>", targetCmd.Owner().DisplayPath())
		targetCmd.AddArg2EnvFromAnotherCmd(cic)
		if targetCmd.GetArgsAutoMapStatus().FullyMapped() {
			return true
		}
	}
	return false
}

type ArgsAutoMapStatus struct {
	argList       []string
	mapNoProvider bool
	mapAll        bool
	argSet        map[string]bool
	mapped        map[string]bool
	metCmds       map[*Cmd]bool
	cache         map[string]Arg2EnvMappingEntry
	providedKeys  map[string]bool
	resultArgs    []string
	resultData    map[string]Arg2EnvMappingEntry
	notAutoArgs   map[int]bool
	originArgCnt  int
	reorderedArgs []string
}

type Arg2EnvMappingEntry struct {
	SrcCmd  *Cmd
	Key     string
	ArgName string
	DefVal  string
	Abbrs   []string
}

func NewArgsAutoMapStatus() *ArgsAutoMapStatus {
	return &ArgsAutoMapStatus{
		nil,
		false,
		false,
		map[string]bool{},
		map[string]bool{},
		map[*Cmd]bool{},
		map[string]Arg2EnvMappingEntry{},
		map[string]bool{},
		nil,
		map[string]Arg2EnvMappingEntry{},
		map[int]bool{},
		0,
		nil,
	}
}

// TODO: move the parsing coding out of core package
func (self *ArgsAutoMapStatus) AddDefinitions(owner *Cmd, args ...string) error {
	ownerArgs := owner.Args()
	self.originArgCnt = len(ownerArgs.Names())

	keyValSep := owner.Owner().Strs.EnvKeyValSep

	for i, argDefinition := range args {
		argDefinition = strings.TrimSpace(argDefinition)
		fields := strings.Split(argDefinition, keyValSep)
		if len(fields) == 1 {
			if argDefinition == "*" {
				self.mapNoProvider = true
			} else if argDefinition == "**" {
				self.mapAll = true
			} else {
				// Will change `self.reorderedArgs[origin_size+i]` later on mapping finish
				self.reorderedArgs = append(self.reorderedArgs, argDefinition)
			}
		} else {
			self.notAutoArgs[i] = true
			argName, err := self.addNotAutoArg(owner, argDefinition)
			if err != nil {
				return err
			}
			self.reorderedArgs = append(self.reorderedArgs, argName)
		}
		if _, ok := self.argSet[argDefinition]; !ok {
			self.argList = append(self.argList, argDefinition)
			self.argSet[argDefinition] = true
		}
	}
	return nil
}

func (self *ArgsAutoMapStatus) IsEmpty() bool {
	return len(self.argList) == 0 && !self.mapAll && !self.mapNoProvider
}

// TODO: Keep the order of the generated arg-list
func (self *ArgsAutoMapStatus) OrderedMappingArgList() []string {
	return self.argList
}

func (self *ArgsAutoMapStatus) ShouldMapByDefinition(
	owner *Cmd, srcCmd *Cmd, argNameAndAbbrs []string) (matchName string, shouldMap bool, shouldMarkMapped bool) {

	if len(argNameAndAbbrs) == 0 {
		return
	}
	for _, name := range argNameAndAbbrs {
		if _, ok := self.mapped[name]; ok {
			continue
		}
		_, ok := self.argSet[name]
		if ok {
			return name, true, true
		}
	}
	if self.mapAll || self.mapNoProvider {
		return argNameAndAbbrs[0], true, false
	}

	return
}

func (self *ArgsAutoMapStatus) ShouldSkip(srcCmd *Cmd) bool {
	_, ok := self.metCmds[srcCmd]
	return ok
}

func (self *ArgsAutoMapStatus) MarkMet(srcCmd *Cmd) {
	if _, ok := self.metCmds[srcCmd]; ok {
		return
	}
	self.metCmds[srcCmd] = true
	self.recordProvidedKeys(srcCmd)
}

func (self *ArgsAutoMapStatus) MarkMetWithArgv(srcCmd *Cmd, argv ArgVals) {
	if _, ok := self.metCmds[srcCmd]; ok {
		return
	}
	self.metCmds[srcCmd] = true
	self.recordProvidedKeysWithArgv(srcCmd, argv)
}

func (self *ArgsAutoMapStatus) FullyMappedOrMapAll() bool {
	if self.mapAll || self.mapNoProvider {
		return true
	}
	return len(self.mapped)+len(self.notAutoArgs) == len(self.argSet)
}

func (self *ArgsAutoMapStatus) GetUnmappedArgs() (unmapped []string) {
	for _, argName := range self.argList {
		if _, ok := self.mapped[argName]; !ok {
			unmapped = append(unmapped, argName)
		}
	}
	return
}

func (self *ArgsAutoMapStatus) FullyMapped() bool {
	if self.mapAll || self.mapNoProvider {
		return false
	}
	return len(self.mapped) == len(self.argSet)
}

func (self *ArgsAutoMapStatus) MarkAndCacheMapping(srcCmd *Cmd, key string, argName string, defVal string, abbrs []string, shouldMarkMapped bool) {
	self.MarkMet(srcCmd)
	self.cache[argName] = Arg2EnvMappingEntry{srcCmd, key, argName, defVal, abbrs}
	if shouldMarkMapped {
		self.mapped[argName] = true
	}
}

func (self *ArgsAutoMapStatus) GetMappedSource(argName string) *Arg2EnvMappingEntry {
	entry, ok := self.resultData[argName]
	if !ok {
		return nil
	}
	return &entry
}

func (self *ArgsAutoMapStatus) FlushCache(owner *Cmd) error {
	for i, argName := range self.argList {
		if argName == "*" || argName == "**" {
			if i+1 != len(self.argList) {
				return fmt.Errorf("[%s] '*' or '**' can only at the end of args auto mapping definition", owner.Owner().DisplayPath())
			}
			var args []string
			for arg := range self.cache {
				args = append(args, arg)
			}
			sort.Strings(args)
			for _, arg := range args {
				newArgName := self.flushCacheEntry(owner, self.cache[arg], true)
				if len(newArgName) != 0 {
					self.reorderedArgs = append(self.reorderedArgs, newArgName)
				}
			}
			break
		}
		if self.notAutoArgs[i] {
			continue
		}
		if entry, ok := self.cache[argName]; ok {
			newArgName := self.flushCacheEntry(owner, entry, false)
			self.reorderedArgs[i] = newArgName
			delete(self.cache, argName)
		}
	}

	self.reorderArgs(owner)
	return nil
}

func (self *ArgsAutoMapStatus) reorderArgs(owner *Cmd) {
	ownerArgs := owner.Args()
	origin := ownerArgs.Names()[:self.originArgCnt]
	var reorderedArgs []string
	for _, it := range self.reorderedArgs {
		if ownerArgs.Has(it) {
			reorderedArgs = append(reorderedArgs, it)
		}
	}
	reorderedArgs = append(origin, reorderedArgs...)
	owner.ReorderArgs(reorderedArgs)
}

func (self *ArgsAutoMapStatus) addNotAutoArg(owner *Cmd, argDefinition string) (string, error) {
	keyValSep := owner.Owner().Strs.EnvKeyValSep
	i := strings.Index(argDefinition, keyValSep)
	if i <= 0 {
		return "", fmt.Errorf("[%s] bad not-auto arg definition: %s", owner.Owner().DisplayPath(), argDefinition)
	}
	argName := argDefinition[:i]
	defVal := strings.TrimSpace(argDefinition[len(argName)+len(keyValSep):])
	defVal = strings.Trim(strings.Trim(defVal, "'"), "\"")
	nameAndAbbrs := strings.Split(argName, owner.Owner().Strs.AbbrsSep)
	name := strings.TrimSpace(nameAndAbbrs[0])
	var argAbbrs []string
	for _, abbr := range nameAndAbbrs[1:] {
		argAbbrs = append(argAbbrs, strings.TrimSpace(abbr))
	}
	owner.AddArg(name, defVal, argAbbrs...)
	return name, nil
}

func (self *ArgsAutoMapStatus) flushCacheEntry(owner *Cmd, entry Arg2EnvMappingEntry, isAutoMapAll bool) (argName string) {
	if !self.mapAll && self.mapNoProvider {
		_, userSpecify := self.argSet[entry.ArgName]
		if !userSpecify {
			if _, ok := self.providedKeys[entry.Key]; ok {
				return
			}
		}
	}
	arg2env := owner.GetArg2Env()
	if arg2env.Has(entry.Key) {
		return
	}
	args := owner.Args()
	if args.HasArgOrAbbr(entry.ArgName) {
		return
	}

	srcArgs := entry.SrcCmd.Args()
	realArgNameInSrc := srcArgs.Realname(entry.ArgName)

	argName = entry.ArgName
	if !args.HasArgOrAbbr(realArgNameInSrc) {
		argName = realArgNameInSrc
		entry.Abbrs = append(entry.Abbrs, entry.ArgName)
	}
	var newAbbrs []string
	for _, abbr := range entry.Abbrs {
		if abbr == argName {
			continue
		}
		if args.HasArgOrAbbr(abbr) {
			continue
		}
		newAbbrs = append(newAbbrs, abbr)
	}

	if isAutoMapAll {
		owner.AddAutoMapAllArg(argName, entry.DefVal, newAbbrs...)
	} else {
		owner.AddArg(argName, entry.DefVal, newAbbrs...)
	}
	owner.AddArg2Env(entry.Key, argName)
	self.resultArgs = append(self.resultArgs, argName)
	self.resultData[argName] = Arg2EnvMappingEntry{entry.SrcCmd, entry.Key, argName, entry.DefVal, newAbbrs}
	return argName
}

func (self *ArgsAutoMapStatus) recordProvidedKeys(srcCmd *Cmd) {
	envOps := srcCmd.EnvOps()
	for _, key := range envOps.AllWriteKeys() {
		self.providedKeys[key] = true
	}

	val2env := srcCmd.GetVal2Env()
	for _, key := range val2env.EnvKeys() {
		self.providedKeys[key] = true
	}
}

func (self *ArgsAutoMapStatus) recordProvidedKeysWithArgv(srcCmd *Cmd, argv ArgVals) {
	envOps := srcCmd.EnvOps()
	for _, key := range envOps.AllWriteKeys() {
		self.providedKeys[key] = true
	}

	val2env := srcCmd.GetVal2Env()
	for _, key := range val2env.EnvKeys() {
		self.providedKeys[key] = true
	}

	arg2env := srcCmd.GetArg2Env()
	for _, envKey := range arg2env.EnvKeys() {
		argName := arg2env.GetArgName(srcCmd, envKey, true)
		if val, ok := argv[argName]; ok && val.Provided {
			self.providedKeys[envKey] = true
		}
	}
}

func (self *ArgsAutoMapStatus) RecordGlobalEnvKey(key string) {
	self.providedKeys[key] = true
}
