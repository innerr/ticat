package display

import (
	"strings"

	"github.com/pingcap/ticat/pkg/core/model"
)

// TODO: dump more info, eg: full path
func DumpEnvAbbrs(cc *model.Cli, env *model.Env, indentSize int) {
	dumpEnvAbbrs(cc.Screen, cc.EnvAbbrs, env, cc.Cmds.Strs.AbbrsSep, indentSize, 0)
}

func dumpEnvAbbrs(
	screen model.Screen,
	abbrs *model.EnvAbbrs,
	env *model.Env,
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
