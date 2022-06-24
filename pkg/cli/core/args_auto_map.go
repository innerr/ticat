package core

import (
	"sort"
)

type Arg2EnvAutoMapCmds map[*Cmd]bool

func (self Arg2EnvAutoMapCmds) Add(cmd *Cmd) {
	self[cmd] = true
}

func (self Arg2EnvAutoMapCmds) AutoMapArg2Env(cc *Cli, env *Env, envOpCmds []EnvOpCmd) {
	for cmd, _ := range self {
		argv := env.GetArgv(cmd.Owner().Path(), env.GetRaw("strs.cmd-path-sep"), cmd.Args())
		autoMapArg2EnvForCmd(cc, env.Clone(), cmd, argv, envOpCmds, cmd)
		cmd.FinishArg2EnvAutoMap(cc)
	}
}

func autoMapArg2EnvForCmd(
	cc *Cli,
	env *Env,
	srcCmd *Cmd,
	argv ArgVals,
	envOpCmds []EnvOpCmd,
	targetCmd *Cmd) (done bool) {

	targetCmd.GetArgsAutoMapStatus().MarkMet(srcCmd)
	if !srcCmd.HasSubFlow() {
		return false
	}

	env = env.Clone().GetLayer(EnvLayerSession)
	ApplyVal2Env(env, srcCmd)
	ApplyArg2Env(env, srcCmd, argv)

	subFlow, _, rendered := srcCmd.Flow(argv, cc, env, true, true)
	if len(subFlow) == 0 || !rendered {
		// TODO: careful test: handle not fully renderd subflow
		//return false
	}

	parsedFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
	flowEnv := env.NewLayer(EnvLayerSubFlow)
	parsedFlow.GlobalEnv.WriteNotArgTo(flowEnv, cc.Cmds.Strs.EnvValDelAllMark)

	return autoMapArg2EnvForCmdsInFlow(cc, flowEnv, parsedFlow, 0, envOpCmds, targetCmd)
}

func autoMapArg2EnvForCmdsInFlow(
	cc *Cli,
	env *Env,
	flow *ParsedCmds,
	currCmdIdx int,
	envOpCmds []EnvOpCmd,
	targetCmd *Cmd) (done bool) {

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

		cmdEnv, argv := it.ApplyMappingGenEnvAndArgv(env, cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep)

		if autoMapArg2EnvForCmd(cc, cmdEnv, cic, argv, envOpCmds, targetCmd) {
			return true
		}

		TryExeEnvOpCmds(argv, cc, cmdEnv, flow, i, envOpCmds, nil,
			"failed to execute env op-cmd in depends collecting")

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
	writeKeys     map[string]bool
	resultArgs    []string
	resultData    map[string]Arg2EnvMappingEntry
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
	}
}

func (self *ArgsAutoMapStatus) Add(args ...string) {
	for _, arg := range args {
		if arg == "*" {
			self.mapNoProvider = true
		}
		if arg == "**" {
			self.mapAll = true
		}
		if _, ok := self.argSet[arg]; !ok {
			self.argList = append(self.argList, arg)
			self.argSet[arg] = true
		}
	}
}

func (self *ArgsAutoMapStatus) IsEmpty() bool {
	return len(self.argList) == 0 && !self.mapAll && !self.mapNoProvider
}

// TODO: Keep the order of the generated arg-list
func (self *ArgsAutoMapStatus) OrderedMappingArgList() []string {
	return self.argList
}

func (self *ArgsAutoMapStatus) ShouldMapByDefinition(
	cmd *Cmd, srcCmd *Cmd, argNameAndAbbrs []string) (matchName string, shouldMap bool, shouldMarkMapped bool) {

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

func (self *ArgsAutoMapStatus) MarkMet(cmd *Cmd) {
	if _, ok := self.metCmds[cmd]; ok {
		return
	}
	self.metCmds[cmd] = true
	self.recordWriteKeys(cmd)
}

func (self *ArgsAutoMapStatus) FullyMappedOrMapAll() bool {
	if self.mapAll || self.mapNoProvider {
		return true
	}
	return len(self.mapped) == len(self.argSet)
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

func (self *ArgsAutoMapStatus) FlushCache(cmd *Cmd) {
	for _, argName := range self.argList {
		if argName == "*" || argName == "**" {
			var args []string
			for arg, _ := range self.cache {
				args = append(args, arg)
			}
			sort.Strings(args)
			for _, arg := range args {
				self.flushCacheEntry(cmd, self.cache[arg])
			}
			return
		}
		if entry, ok := self.cache[argName]; ok {
			self.flushCacheEntry(cmd, entry)
			delete(self.cache, argName)
		}
	}
}

func (self *ArgsAutoMapStatus) flushCacheEntry(cmd *Cmd, entry Arg2EnvMappingEntry) {
	if !self.mapAll && self.mapNoProvider {
		_, userSpecify := self.argSet[entry.ArgName]
		if !userSpecify {
			if _, ok := self.writeKeys[entry.Key]; ok {
				return
			}
		}
	}
	arg2env := cmd.GetArg2Env()
	if arg2env.Has(entry.Key) {
		return
	}
	args := cmd.Args()
	if args.HasArgOrAbbr(entry.ArgName) {
		return
	}

	srcArgs := entry.SrcCmd.Args()
	realArgNameInSrc := srcArgs.Realname(entry.ArgName)

	argName := entry.ArgName
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

	cmd.AddArg(argName, entry.DefVal, newAbbrs...)
	cmd.AddArg2Env(entry.Key, argName)
	self.resultArgs = append(self.resultArgs, argName)
	self.resultData[argName] = Arg2EnvMappingEntry{entry.SrcCmd, entry.Key, argName, entry.DefVal, newAbbrs}
}

func (self *ArgsAutoMapStatus) recordWriteKeys(cmd *Cmd) {
	envOps := cmd.EnvOps()
	for _, key := range envOps.AllWriteKeys() {
		self.writeKeys[key] = true
	}
}
