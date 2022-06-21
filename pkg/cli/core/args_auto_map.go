package core

type Arg2EnvAutoMapCmds map[*Cmd]bool

func (self Arg2EnvAutoMapCmds) Add(cmd *Cmd) {
	self[cmd] = true
}

func (self Arg2EnvAutoMapCmds) AutoMapArg2Env(
	cc *Cli,
	env *Env,
	envOpCmds []EnvOpCmd) {

	for cmd, unmapped := range self {
		if !unmapped {
			continue
		}
		targetCmdStack := NewArg2EnvAutoMapTargetCmdStack()
		targetCmdStack.Push(cmd)
		argv := env.GetArgv(cmd.Owner().Path(), env.GetRaw("strs.cmd-path-sep"), cmd.Args())
		autoMapArg2EnvForCmd(cc, env.Clone(), cmd, argv, envOpCmds, targetCmdStack)
		self[cmd] = false
	}
}

func autoMapArg2EnvForCmd(
	cc *Cli,
	env *Env,
	cmd *Cmd,
	argv ArgVals,
	envOpCmds []EnvOpCmd,
	targetCmdStack *Arg2EnvAutoMapTargetCmdStack) {

	if !cmd.HasSubFlow() {
		return
	}

	env = env.Clone().GetLayer(EnvLayerSession)
	ApplyVal2Env(env, cmd)
	ApplyArg2Env(env, cmd, argv)

	subFlow, _, rendered := cmd.Flow(argv, cc, env, true, true)
	if len(subFlow) == 0 || !rendered {
		return
	}

	parsedFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
	flowEnv := env.NewLayer(EnvLayerSubFlow)
	parsedFlow.GlobalEnv.WriteNotArgTo(flowEnv, cc.Cmds.Strs.EnvValDelAllMark)
	autoMapArg2EnvForCmdsInFlow(cc, flowEnv, parsedFlow, 0, envOpCmds, targetCmdStack)
}

func autoMapArg2EnvForCmdsInFlow(
	cc *Cli,
	env *Env,
	flow *ParsedCmds,
	currCmdIdx int,
	envOpCmds []EnvOpCmd,
	targetCmdStack *Arg2EnvAutoMapTargetCmdStack) {

	for i := currCmdIdx; i < len(flow.Cmds); i++ {
		it := flow.Cmds[i]
		cic := it.LastCmd()
		if cic == nil {
			continue
		}
		if it.ParseResult.Error != nil && !it.ParseResult.IsMinorErr {
			continue
		}

		cmdEnv, argv := it.ApplyMappingGenEnvAndArgv(env, cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep)

		//targetCmdStack.Push(cic)
		autoMapArg2EnvForCmd(cc, cmdEnv, cic, argv, envOpCmds, targetCmdStack)
		//targetCmdStack.Pop()

		TryExeEnvOpCmds(argv, cc, cmdEnv, flow, i, envOpCmds, nil,
			"failed to execute env op-cmd in depends collecting")

		targetCmdStack.MapCmdArg2EnvToTargets(cic)
	}
}

type Arg2EnvAutoMapTargetCmdStack struct {
	stack []*Cmd
}

func NewArg2EnvAutoMapTargetCmdStack() *Arg2EnvAutoMapTargetCmdStack {
	return &Arg2EnvAutoMapTargetCmdStack{nil}
}

func (self *Arg2EnvAutoMapTargetCmdStack) Push(cmd *Cmd) {
	self.stack = append(self.stack, cmd)
}

func (self *Arg2EnvAutoMapTargetCmdStack) Pop() {
	if len(self.stack) != 0 {
		self.stack = self.stack[:len(self.stack)-1]
	}
}

func (self *Arg2EnvAutoMapTargetCmdStack) MapCmdArg2EnvToTargets(src *Cmd) {
	for i := len(self.stack) - 1; i >= 0; i-- {
		target := self.stack[i]
		//println(src.Owner().DisplayPath(), " (arg2env) =>", target.Owner().DisplayPath())
		target.AddArg2EnvFromAnotherCmd(src)
	}
}

type ArgsAutoMapStatus struct {
	argList []string
	mapAll  bool
	argSet  map[string]bool

	// TODO: Check if all defined arg mapping names are mapped
	mapped map[string]bool
}

func NewArgsAutoMapStatus() *ArgsAutoMapStatus {
	return &ArgsAutoMapStatus{
		nil,
		false,
		map[string]bool{},
		map[string]bool{},
	}
}

func (self *ArgsAutoMapStatus) Add(args ...string) {
	for _, arg := range args {
		if arg == "*" {
			self.mapAll = true
			continue
		}
		if _, ok := self.argSet[arg]; !ok {
			self.argList = append(self.argList, arg)
			self.argSet[arg] = true
		}
	}
}

func (self *ArgsAutoMapStatus) IsEmpty() bool {
	return len(self.argList) == 0 && !self.mapAll
}

// TODO: Keep the order of the generated arg-list
func (self *ArgsAutoMapStatus) OrderedMappingArgList() []string {
	return self.argList
}

func (self *ArgsAutoMapStatus) ShouldMap(arg string) bool {
	if self.mapAll {
		return true
	}
	_, ok := self.argSet[arg]
	return ok
}
