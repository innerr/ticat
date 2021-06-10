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
			line += MayQuoteStr(v)
			if printDef {
				if defV != v {
					line += "(def=" + MayQuoteStr(defV) + ")"
				} else {
					line += "(=def)"
				}
			}
		} else {
			line += MayQuoteStr(defV)
		}
		output = append(output, line)
	}
	return
}
