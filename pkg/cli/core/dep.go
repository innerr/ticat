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

	collectDepends(cc, env.Clone(), flow, currCmdIdx, res, allowFlowTemplateRenderError, envOpCmds, 0)
}

func collectDepends(
	cc *Cli,
	env *Env,
	flow *ParsedCmds,
	currCmdIdx int,
	res Depends,
	allowFlowTemplateRenderError bool,
	envOpCmds []EnvOpCmd,
	depth int) {

	for i := currCmdIdx; i < len(flow.Cmds); i++ {
		it := flow.Cmds[i]
		cic := it.LastCmd()
		if cic == nil {
			continue
		}

		cmdEnv, argv := it.ApplyMappingGenEnvAndArgv(env, cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep, depth+1)

		if cic.Type() == CmdTypeFileNFlow {
			subFlow, _, rendered := cic.Flow(argv, cc, cmdEnv, allowFlowTemplateRenderError, true)
			if rendered && len(subFlow) != 0 {
				parsedFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
				flowEnv := cmdEnv.NewLayer(EnvLayerSubFlow)
				parsedFlow.GlobalEnv.WriteNotArgTo(flowEnv, cc.Cmds.Strs.EnvValDelAllMark)
				// Allow parse errors here
				collectDepends(cc, flowEnv, parsedFlow, 0, res, allowFlowTemplateRenderError, envOpCmds, depth+1)
			}
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

		TryExeEnvOpCmds(argv, cc, cmdEnv, flow, i, envOpCmds, nil,
			"failed to execute env op-cmd in depends collecting")

		if !cic.HasSubFlow(true) {
			continue
		}

		subFlow, _, rendered := cic.Flow(argv, cc, cmdEnv, allowFlowTemplateRenderError, true)
		if rendered && len(subFlow) != 0 {
			parsedFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
			flowEnv := cmdEnv.NewLayer(EnvLayerSubFlow)
			parsedFlow.GlobalEnv.WriteNotArgTo(flowEnv, cc.Cmds.Strs.EnvValDelAllMark)
			// Allow parse errors here
			collectDepends(cc, flowEnv, parsedFlow, 0, res, allowFlowTemplateRenderError, envOpCmds, depth+1)
		}
	}
}
