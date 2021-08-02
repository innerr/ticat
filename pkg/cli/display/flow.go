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
	parsedGlobalEnv core.ParsedEnv,
	flow []core.ParsedCmd,
	args *DumpFlowArgs) {

	if len(flow) == 0 {
		return
	}

	// The env will be modified during dumping (so it could show the real value)
	// so we need to clone the env to protect it
	env = env.Clone()
	maxDepth := env.GetInt("display.flow.depth")

	PrintTipTitle(cc.Screen, env, "flow executing description:")
	cc.Screen.Print("--->>>\n")
	dumpFlow(cc, env, parsedGlobalEnv, flow, args, maxDepth, 0)
	cc.Screen.Print("<<<---\n")
}

func dumpFlow(
	cc *core.Cli,
	env *core.Env,
	parsedGlobalEnv core.ParsedEnv,
	flow []core.ParsedCmd,
	args *DumpFlowArgs,
	maxDepth int,
	indentAdjust int) {

	metFlows := map[string]bool{}
	for _, cmd := range flow {
		if !cmd.IsEmpty() {
			dumpFlowCmd(cc, cc.Screen, env, parsedGlobalEnv, cmd, args,
				maxDepth, indentAdjust, metFlows)
		}
	}
}

func dumpFlowCmd(
	cc *core.Cli,
	screen core.Screen,
	env *core.Env,
	parsedGlobalEnv core.ParsedEnv,
	parsedCmd core.ParsedCmd,
	args *DumpFlowArgs,
	maxDepth int,
	indentAdjust int,
	metFlows map[string]bool) {

	cmd := parsedCmd.Last().Matched.Cmd
	if cmd == nil {
		return
	}

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
	var name string
	if args.Skeleton {
		name = strings.Join(parsedCmd.Path(), sep)
	} else {
		name = parsedCmd.DisplayPath(sep, true)
	}
	prt(0, "["+name+"]")
	if len(cic.Help()) != 0 {
		prt(1, " '"+cic.Help()+"'")
	}

	// TODO: this is slow
	originEnv := env.Clone()
	cmdEnv, argv := parsedCmd.ApplyMappingGenEnvAndArgv(env, cc.Cmds.Strs.EnvValDelAllMark, sep)

	if !args.Skeleton {
		args := parsedCmd.Args()
		arg2env := cic.GetArg2Env()
		argLines := DumpEffectedArgs(originEnv, arg2env, &args, argv)
		if len(argLines) != 0 {
			prt(1, "- args:")
		}
		for _, line := range argLines {
			prt(2, line)
		}
	}

	if !args.Skeleton {
		keys, kvs := dumpFlowEnv(cc, originEnv, parsedGlobalEnv, parsedCmd, cmd, argv)
		if len(keys) != 0 {
			prt(1, "- env-values:")
		}
		for _, k := range keys {
			v := kvs[k]
			prt(2, k+" = "+mayQuoteStr(v.Val)+" "+v.Source+"")
		}
	}

	if !args.Skeleton {
		envOps := cic.EnvOps()
		envOpKeys := envOps.EnvKeys()
		if len(envOpKeys) != 0 {
			prt(1, "- env-ops:")
		}
		for _, k := range envOpKeys {
			prt(2, k+" = "+dumpEnvOps(envOps.Ops(k), envOpSep))
		}
	}

	if !args.Simple && !args.Skeleton {
		line := string(cic.Type())
		if cic.IsQuiet() {
			line += " (quiet)"
		}
		if cic.IsPriority() {
			line += " (priority)"
		}
		prt(1, "- cmd-type:")
		prt(2, line)

		if len(cmd.Source()) != 0 && !strings.HasPrefix(cic.CmdLine(), cmd.Source()) {
			prt(1, "- from:")
			prt(2, cmd.Source())
		}
	}

	if (len(cic.CmdLine()) != 0 || len(cic.FlowStrs()) != 0) &&
		cic.Type() != core.CmdTypeNormal && cic.Type() != core.CmdTypePower {
		metFlow := false
		if cic.Type() == core.CmdTypeFlow {
			flowStrs, _ := cic.RenderedFlowStrs(argv, cmdEnv, true)
			flowStr := strings.Join(flowStrs, " ")
			metFlow = metFlows[flowStr]
			if metFlow {
				prt(1, "- flow (duplicated):")
			} else {
				metFlows[flowStr] = true
				prt(1, "- flow:")
			}
			for _, flowStr := range flowStrs {
				prt(2, flowStr)
			}
		} else if !args.Simple && !args.Skeleton {
			if cic.Type() == core.CmdTypeEmptyDir {
				prt(1, "- dir:")
				prt(2, cic.CmdLine())
			} else {
				prt(1, "- executable:")
				prt(2, cic.CmdLine())
			}
			if len(cic.MetaFile()) != 0 {
				prt(1, "- meta:")
				prt(2, cic.MetaFile())
			}
		}
		if cic.Type() == core.CmdTypeFlow && maxDepth > 1 {
			subFlow, rendered := cic.Flow(argv, cmdEnv, true)
			if rendered && len(subFlow) != 0 {
				if !metFlow {
					prt(2, "--->>>")
					parsedFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
					err := parsedFlow.FirstErr()
					if err != nil {
						panic(err.Error)
					}
					parsedFlow.GlobalEnv.WriteNotArgTo(env, cc.Cmds.Strs.EnvValDelAllMark)
					dumpFlow(cc, env, parsedGlobalEnv, parsedFlow.Cmds, args, maxDepth-1, indentAdjust+2)
					prt(2, "<<<---")
				}
			}
		}
	}
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
	argv core.ArgVals) (keys []string, kvs map[string]flowEnvVal) {

	kvs = map[string]flowEnvVal{}
	cic := cmd.Cmd()

	tempEnv := core.NewEnv()
	parsedGlobalEnv.WriteNotArgTo(tempEnv, cc.Cmds.Strs.EnvValDelAllMark)
	cmdEssEnv := parsedCmd.GenCmdEnv(tempEnv, cc.Cmds.Strs.EnvValDelAllMark)
	val2env := cic.GetVal2Env()
	for _, k := range val2env.EnvKeys() {
		kvs[k] = flowEnvVal{val2env.Val(k), "<- mod"}
	}

	flatten := cmdEssEnv.Flatten(true, nil, true)
	for k, v := range flatten {
		kvs[k] = flowEnvVal{v, "<- flow"}
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
		kvs[key] = flowEnvVal{val.Raw, "<- arg '" + name + "'"}
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
}

func NewDumpFlowArgs() *DumpFlowArgs {
	return &DumpFlowArgs{false, false, 4}
}

func (self *DumpFlowArgs) SetSimple() *DumpFlowArgs {
	self.Simple = true
	return self
}

func (self *DumpFlowArgs) SetSkeleton() *DumpFlowArgs {
	self.Simple = true
	self.Skeleton = true
	return self
}
