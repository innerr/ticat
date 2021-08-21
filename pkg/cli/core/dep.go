package core

type DependInfo struct {
	Reason string
	Cmd    ParsedCmd
}

type Depends map[string]map[*Cmd]DependInfo

func CollectDepends(
	cc *Cli,
	env *Env,
	flow *ParsedCmds,
	currCmdIdx int,
	res Depends,
	allowFlowTemplateRenderError bool,
	envOpCmds []EnvOpCmd) {

	collectDepends(cc, env.Clone(), flow, currCmdIdx, res, allowFlowTemplateRenderError, envOpCmds)
}

func collectDepends(
	cc *Cli,
	env *Env,
	flow *ParsedCmds,
	currCmdIdx int,
	res Depends,
	allowFlowTemplateRenderError bool,
	envOpCmds []EnvOpCmd) {

	for i := currCmdIdx; i < len(flow.Cmds); i++ {
		it := flow.Cmds[i]
		cic := it.LastCmd()
		if cic == nil {
			continue
		}
		deps := cic.GetDepends()
		for _, dep := range deps {
			cmds, ok := res[dep.OsCmd]
			if ok {
				cmds[cic] = DependInfo{dep.Reason, it}
			} else {
				res[dep.OsCmd] = map[*Cmd]DependInfo{cic: DependInfo{dep.Reason, it}}
			}
		}
		cmdEnv, argv := it.ApplyMappingGenEnvAndArgv(env, cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep)
		TryExeEnvOpCmds(argv, cc, cmdEnv, flow, i, envOpCmds, nil,
			"failed to execute env op-cmd in depends collecting")
		if cic.Type() != CmdTypeFlow {
			continue
		}
		subFlow, rendered := cic.Flow(argv, cmdEnv, allowFlowTemplateRenderError)
		if rendered && len(subFlow) != 0 {
			parsedFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
			parsedFlow.GlobalEnv.WriteNotArgTo(env, cc.Cmds.Strs.EnvValDelAllMark)
			// Allow parse errors here
			collectDepends(cc, env, parsedFlow, 0, res, allowFlowTemplateRenderError, envOpCmds)
		}
	}
}

func TryExeEnvOpCmds(
	argv ArgVals,
	cc *Cli,
	env *Env,
	flow *ParsedCmds,
	currCmdIdx int,
	envOpCmds []EnvOpCmd,
	checker *EnvOpsChecker,
	errString string) {

	cmd := flow.Cmds[currCmdIdx].LastCmd()
	for _, it := range envOpCmds {
		if !cmd.IsTheSameFunc(it.Func) {
			continue
		}
		newCC := cc.Clone()
		newCC.Screen = &QuietScreen{}
		_, succeeded := cmd.Execute(argv, newCC, env, flow, currCmdIdx)
		if !succeeded {
			panic(NewCmdError(flow.Cmds[currCmdIdx], errString))
		}
		if checker != nil {
			it.Action(checker, argv)
		}
		break
	}
}
