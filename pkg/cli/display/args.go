package display

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

func DumpProvidedArgs(env *core.Env, args *core.Args, argv core.ArgVals, colorize bool) (output []string) {
	for _, k := range args.Names() {
		v, provided := argv[k]
		if !provided || !v.Provided {
			continue
		}
		val := mayQuoteStr(mayMaskSensitiveVal(env, k, v.Raw))
		if colorize {
			line := ColorArg(k, env) + ColorSymbol(" = ", env) + val
			output = append(output, line)
		} else {
			line := k + " = " + val
			output = append(output, line)
		}
	}
	return
}

func DumpSysArgs(env *core.Env, sysArgv core.SysArgVals, colorize bool) (output []string) {
	for k, v := range sysArgv {
		v = mayMaskSensitiveVal(env, k, v)
		if colorize {
			line := ColorExplain("[sys] ", env) + ColorArg(k, env) +
				ColorSymbol(" = ", env) + mayQuoteStr(v)
			output = append(output, line)
		} else {
			line := "[sys] " + k + " = " + mayQuoteStr(v)
			output = append(output, line)
		}
	}
	return
}

func DumpEffectedArgs(
	env *core.Env,
	arg2env *core.Arg2Env,
	args *core.Args,
	argv core.ArgVals,
	writtenKeys FlowWrittenKeys,
	stackDepth int) (output []string) {

	for _, k := range args.Names() {
		defV := args.DefVal(k, stackDepth)
		line := ColorArg(k, env) + " " + ColorSymbol("=", env) + " "
		v, provided := argv[k]
		if provided && v.Provided {
			line += mayQuoteStr(v.Raw)
		} else {
			if len(defV) == 0 {
				continue
			}
			key, hasMapping := arg2env.GetEnvKey(k)
			_, inEnv := env.GetEx(key)
			if hasMapping && inEnv {
				continue
			}
			if hasMapping && writtenKeys[key] {
				continue
			}
			line += mayQuoteStr(defV)
		}
		output = append(output, line)
	}
	return
}
