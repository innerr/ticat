package display

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

// TODO: remove this, no use
func DumpArgs(args *core.Args, argv core.ArgVals, printDef bool) (output []string) {
	for _, k := range args.Names() {
		defV := args.DefVal(k)
		line := k + " = "
		v, provided := argv[k]
		if !provided || !v.Provided {
			line += mayQuoteStr(defV)
		} else {
			line += mayQuoteStr(v.Raw)
			if printDef {
				if defV != v.Raw {
					line += "(def=" + mayQuoteStr(defV) + ")"
				} else {
					line += "(=def)"
				}
			}
		}
		output = append(output, line)
	}
	return
}

func DumpProvidedArgs(env *core.Env, args *core.Args, argv core.ArgVals, colorize bool) (output []string) {
	for _, k := range args.Names() {
		v, provided := argv[k]
		if !provided || !v.Provided {
			continue
		}
		if colorize {
			line := ColorArg(k, env) + ColorSymbol(" = ", env) + mayQuoteStr(v.Raw)
			output = append(output, line)
		} else {
			line := k + " = " + mayQuoteStr(v.Raw)
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
	writtenKeys FlowWrittenKeys) (output []string) {

	for _, k := range args.Names() {
		defV := args.DefVal(k)
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
