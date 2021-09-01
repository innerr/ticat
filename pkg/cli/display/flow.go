package display

import (
	"sort"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

// 'parsedGlobalEnv' + env in 'flow' = all env
func DumpFlow(
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	fromCmdIdx int,
	args *DumpFlowArgs,
	envOpCmds []core.EnvOpCmd) {

	if len(flow.Cmds) == 0 {
		return
	}

	// The env will be modified during dumping (so it could show the real value)
	// so we need to clone the env to protect it
	env = env.Clone()

	writtenKeys := FlowWrittenKeys{}

	PrintTipTitle(cc.Screen, env, "flow executing description:")
	cc.Screen.Print(ColorFlowing("--->>>", env) + "\n")
	dumpFlow(cc, env, envOpCmds, flow, fromCmdIdx, args, writtenKeys, args.MaxDepth, args.MaxTrivial, 0)
	cc.Screen.Print(ColorFlowing("<<<---", env) + "\n")
}

func dumpFlow(
	cc *core.Cli,
	env *core.Env,
	envOpCmds []core.EnvOpCmd,
	flow *core.ParsedCmds,
	fromCmdIdx int,
	args *DumpFlowArgs,
	writtenKeys FlowWrittenKeys,
	maxDepth int,
	maxTrivial int,
	indentAdjust int) {

	metFlows := map[string]bool{}
	for i, cmd := range flow.Cmds[fromCmdIdx:] {
		if cmd.IsEmpty() {
			continue
		}
		dumpFlowCmd(cc, cc.Screen, env, envOpCmds, flow, fromCmdIdx+i, args,
			maxDepth, maxTrivial, indentAdjust, metFlows, writtenKeys)
	}
}

func dumpFlowCmd(
	cc *core.Cli,
	screen core.Screen,
	env *core.Env,
	envOpCmds []core.EnvOpCmd,
	flow *core.ParsedCmds,
	currCmdIdx int,
	args *DumpFlowArgs,
	maxDepth int,
	maxTrivial int,
	indentAdjust int,
	metFlows map[string]bool,
	writtenKeys FlowWrittenKeys) {

	parsedCmd := flow.Cmds[currCmdIdx]
	cmd := parsedCmd.Last().Matched.Cmd
	if cmd == nil {
		return
	}

	trivialMark := env.GetRaw("strs.trivial-mark")

	sep := cmd.Strs.PathSep
	envOpSep := " " + cmd.Strs.EnvOpSep + " "

	prt := func(indentLvl int, msg string) {
		indentLvl += indentAdjust
		padding := rpt(" ", args.IndentSize*indentLvl)
		msg = autoPadNewLine(padding, msg)
		screen.Print(padding + msg + "\n")
	}

	cic := cmd.Cmd()
	if cic == nil {
		return
	}

	trivialDelta := cmd.Trivial() + parsedCmd.TrivialLvl

	if maxTrivial > 0 && maxDepth > 0 {
		var name string
		if args.Skeleton {
			name = strings.Join(parsedCmd.Path(), sep)
		} else {
			name = parsedCmd.DisplayPath(sep, true)
		}
		name = ColorCmd("["+name+"]", env)
		if maxTrivial == 1 && trivialDelta > 0 &&
			(cic.Type() == core.CmdTypeFlow || cic.Type() == core.CmdTypeFileNFlow) {
			name += ColorProp(trivialMark, env)
		}
		prt(0, name)

		if len(cic.Help()) != 0 {
			if !args.Skeleton {
				prt(1, " "+ColorHelp("'"+cic.Help()+"'", env))
			} else {
				prt(1, ColorHelp("'"+cic.Help()+"'", env))
			}
		}
	}

	// TODO: this is slow
	originEnv := env.Clone()
	cmdEnv, argv := parsedCmd.ApplyMappingGenEnvAndArgv(env, cc.Cmds.Strs.EnvValDelAllMark, sep)

	if maxTrivial > 0 && maxDepth > 0 {
		if !args.Skeleton {
			args := cic.Args()
			arg2env := cic.GetArg2Env()
			argLines := DumpEffectedArgs(originEnv, arg2env, &args, argv, writtenKeys)
			if len(argLines) != 0 {
				prt(1, ColorProp("- args:", env))
			}
			for _, line := range argLines {
				prt(2, line)
			}
		} else {
			for name, val := range argv {
				if !val.Provided {
					continue
				}
				prt(1, " "+ColorArg(name, env)+ColorSymbol(" = ", env)+val.Raw)
			}
		}
	}

	trivial := maxTrivial - trivialDelta

	if trivial <= 0 || maxDepth <= 0 {
		core.TryExeEnvOpCmds(argv, cc, cmdEnv, flow, currCmdIdx, envOpCmds, nil,
			"failed to execute env-op cmd in flow desc")

		// Thi is for render checking, even it's folded
		subFlow, rendered := cic.Flow(argv, cmdEnv, true)
		if rendered {
			parsedFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
			err := parsedFlow.FirstErr()
			if err != nil {
				panic(err.Error)
			}
			parsedFlow.GlobalEnv.WriteNotArgTo(env, cc.Cmds.Strs.EnvValDelAllMark)
			dumpFlow(cc, env, envOpCmds, parsedFlow, 0, args, writtenKeys,
				maxDepth-1, trivial, indentAdjust+2)
		}
		return
	}

	if !args.Skeleton {
		keys, kvs := dumpFlowEnv(cc, originEnv, flow.GlobalEnv, parsedCmd, cmd, argv, writtenKeys)
		if len(keys) != 0 {
			prt(1, ColorProp("- env-values:", env))
		}
		for _, k := range keys {
			v := kvs[k]
			prt(2, ColorKey(k, env)+ColorSymbol(" = ", env)+mayQuoteStr(v.Val)+" "+v.Source+"")
		}
	}
	writtenKeys.AddCmd(argv, env, cic)

	if !args.Skeleton {
		envOps := cic.EnvOps()
		envOpKeys, origins, _ := envOps.RenderedEnvKeys(argv, cmdEnv, cic, false)
		if len(envOpKeys) != 0 {
			prt(1, ColorProp("- env-ops:", env))
		}
		for i, k := range envOpKeys {
			prt(2, ColorKey(k, env)+ColorSymbol(" = ", env)+dumpEnvOps(envOps.Ops(origins[i]), envOpSep))
		}
	}

	if !args.Simple && !args.Skeleton {
		line := string(cic.Type())
		if cic.IsQuiet() {
			line += " quiet"
		}
		if cic.IsPriority() {
			line += " priority"
		}
		prt(1, ColorProp("- cmd-type:", env))
		prt(2, line)

		if len(cmd.Source()) != 0 && !strings.HasPrefix(cic.CmdLine(), cmd.Source()) {
			prt(1, ColorProp("- from:", env))
			prt(2, cmd.Source())
		}
	}

	if (len(cic.CmdLine()) != 0 || len(cic.FlowStrs()) != 0) &&
		cic.Type() != core.CmdTypeNormal && cic.Type() != core.CmdTypePower {
		metFlow := false
		if cic.Type() == core.CmdTypeFlow || cic.Type() == core.CmdTypeFileNFlow {
			flowStrs, _ := cic.RenderedFlowStrs(argv, cmdEnv, true)
			flowStr := strings.Join(flowStrs, " ")
			metFlow = metFlows[flowStr]
			if metFlow {
				if !args.Skeleton {
					prt(1, ColorProp("- flow (duplicated):", env))
				} else {
					prt(1, ColorProp("(duplicated sub flow)", env))
				}
			} else {
				metFlows[flowStr] = true
				if maxDepth <= 1 {
					if !args.Skeleton {
						prt(1, ColorProp("- flow (folded):", env))
					} else {
						prt(1, ColorProp("(folded flow)", env))
					}
				} else {
					if !args.Skeleton {
						prt(1, ColorProp("- flow:", env))
					}
				}
			}
			if !args.Skeleton {
				for _, flowStr := range flowStrs {
					prt(2, ColorFlow(flowStr, env))
				}
			}
		} else if !args.Simple && !args.Skeleton {
			if cic.Type() == core.CmdTypeEmptyDir {
				prt(1, ColorProp("- dir:", env))
				prt(2, cic.CmdLine())
			} else {
				prt(1, ColorProp("- executable:", env))
				prt(2, cic.CmdLine())
			}
			if len(cic.MetaFile()) != 0 {
				prt(1, ColorProp("- meta:", env))
				prt(2, cic.MetaFile())
			}
		}

		if cic.Type() == core.CmdTypeFlow || cic.Type() == core.CmdTypeFileNFlow {
			subFlow, rendered := cic.Flow(argv, cmdEnv, true)
			if rendered && len(subFlow) != 0 {
				if !metFlow {
					if maxDepth > 1 {
						prt(2, ColorFlowing("--->>>", env))
					}
					parsedFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
					err := parsedFlow.FirstErr()
					if err != nil {
						panic(err.Error)
					}

					// TODO: need it or not ?
					parsedFlow.GlobalEnv.WriteNotArgTo(env, cc.Cmds.Strs.EnvValDelAllMark)

					dumpFlow(cc, env, envOpCmds, parsedFlow, 0, args, writtenKeys,
						maxDepth-1, trivial, indentAdjust+2)
					exeMark := ""
					if cic.Type() == core.CmdTypeFileNFlow {
						exeMark = ColorCmd(" +", env)
					}
					if maxDepth > 1 {
						prt(2, ColorFlowing("<<<---", env)+exeMark)
					}
				}
			}
		}
	}

	core.TryExeEnvOpCmds(argv, cc, cmdEnv, flow, currCmdIdx, envOpCmds, nil,
		"failed to execute env-op cmd in flow desc")
}

type flowEnvVal struct {
	Val    string
	Source string
}

func dumpFlowEnv(
	cc *core.Cli,
	env *core.Env,
	parsedGlobalEnv core.ParsedEnv,
	parsedCmd core.ParsedCmd,
	cmd *core.CmdTree,
	argv core.ArgVals,
	writtenKeys FlowWrittenKeys) (keys []string, kvs map[string]flowEnvVal) {

	kvs = map[string]flowEnvVal{}
	cic := cmd.Cmd()

	tempEnv := core.NewEnv()
	parsedGlobalEnv.WriteNotArgTo(tempEnv, cc.Cmds.Strs.EnvValDelAllMark)
	cmdEssEnv := parsedCmd.GenCmdEnv(tempEnv, cc.Cmds.Strs.EnvValDelAllMark)
	val2env := cic.GetVal2Env()
	for _, k := range val2env.EnvKeys() {
		kvs[k] = flowEnvVal{val2env.Val(k), ColorSymbol("<- mod", env)}
	}

	flatten := cmdEssEnv.Flatten(true, nil, true)
	for k, v := range flatten {
		kvs[k] = flowEnvVal{v, ColorSymbol("<- flow", env)}
	}

	arg2env := cic.GetArg2Env()
	for name, val := range argv {
		if !val.Provided && len(val.Raw) == 0 {
			continue
		}
		key, hasMapping := arg2env.GetEnvKey(name)
		if !hasMapping {
			continue
		}
		_, inEnv := env.GetEx(key)
		if !val.Provided && inEnv {
			continue
		}
		if writtenKeys[key] {
			continue
		}
		kvs[key] = flowEnvVal{val.Raw, ColorSymbol("<- arg", env) +
			ColorArg(" '"+name+"'", env)}
	}

	for k, _ := range kvs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return
}

type DumpFlowArgs struct {
	Simple     bool
	Skeleton   bool
	IndentSize int
	MaxDepth   int
	MaxTrivial int
}

func NewDumpFlowArgs() *DumpFlowArgs {
	return &DumpFlowArgs{false, false, 4, 32, 1}
}

func (self *DumpFlowArgs) SetSimple() *DumpFlowArgs {
	self.Simple = true
	return self
}

func (self *DumpFlowArgs) SetMaxDepth(val int) *DumpFlowArgs {
	self.MaxDepth = val
	return self
}

func (self *DumpFlowArgs) SetMaxTrivial(val int) *DumpFlowArgs {
	self.MaxTrivial = val
	return self
}

func (self *DumpFlowArgs) SetSkeleton() *DumpFlowArgs {
	self.Simple = true
	self.Skeleton = true
	return self
}

type FlowWrittenKeys map[string]bool

func (self FlowWrittenKeys) AddCmd(argv core.ArgVals, env *core.Env, cic *core.Cmd) {
	if cic == nil {
		return
	}
	ops := cic.EnvOps()
	keys, _, _ := ops.RenderedEnvKeys(argv, env, cic, false)
	for _, k := range keys {
		// If is read-op, then the key must exists, so no need to check the op flags
		self[k] = true
	}
}
