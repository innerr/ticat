package display

import (
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

// TODO: dump more info, eg: full path
func DumpEnvAbbrs(cc *core.Cli, env *core.Env, indentSize int) {
	dumpEnvAbbrs(cc.Screen, cc.EnvAbbrs, env, cc.Cmds.Strs.AbbrsSep, indentSize, 0)
}

func dumpEnvAbbrs(
	screen core.Screen,
	abbrs *core.EnvAbbrs,
	env *core.Env,
	abbrsSep string,
	indentSize int,
	indent int) {

	if abbrs == nil {
		return
	}
	prt := func(msg string) {
		if indent >= 0 {
			screen.Print(rpt(" ", indentSize*indent) + msg + "\n")
		}
	}

	name := strings.Join(abbrs.Abbrs(), abbrsSep)
	prt(ColorKey("["+name+"]", env))

	for _, name := range abbrs.SubNames() {
		dumpEnvAbbrs(screen, abbrs.GetSub(name), env, abbrsSep, indentSize, indent+1)
	}
}
