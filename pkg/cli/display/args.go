package display

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

func DumpArgs(args *core.Args, argv core.ArgVals, printDef bool) (output []string) {
	for _, k := range args.Names() {
		defV := args.DefVal(k)
		line := k + " = "
		if argv != nil {
			v := argv[k].Raw
			line += mayQuoteStr(v)
			if printDef {
				if defV != v {
					line += "(def=" + mayQuoteStr(defV) + ")"
				} else {
					line += "(=def)"
				}
			}
		} else {
			line += mayQuoteStr(defV)
		}
		output = append(output, line)
	}
	return
}

func DumpProvidedArgs(args *core.Args, argv core.ArgVals) (output []string) {
	if argv == nil {
		return
	}
	for _, k := range args.Names() {
		v := argv[k]
		if !v.Provided {
			continue
		}
		line := k + " = " + mayQuoteStr(v.Raw)
		output = append(output, line)
	}
	return
}
