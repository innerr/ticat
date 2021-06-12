package display

import (
	"fmt"

	"github.com/pingcap/ticat/pkg/cli/core"
)

type DependInfo struct {
	Reason string
	Cmd    core.ParsedCmd
}

type Depends map[string]map[*core.Cmd]DependInfo

func DumpDepends(cc *core.Cli, env *core.Env, deps Depends) {
	if len(deps) == 0 {
		return
	}

	sep := env.Get("strs.cmd-path-sep").Raw

	if !env.GetBool("display.flow.simplified") {
		PrintTipTitle(cc.Screen, env,
			"depended os commands.",
			"",
			"this flow need the os commands below to execute,",
			"make sure they are all installed.")
	} else {
		cc.Screen.Print(fmt.Sprintf("-------=<%s>=-------\n\n", "depended os commands"))
	}

	for osCmd, cmds := range deps {
		cc.Screen.Print(fmt.Sprintf("[%s]\n", osCmd))
		for _, info := range cmds {
			cc.Screen.Print(fmt.Sprintf("        '%s'\n", info.Reason))
			cc.Screen.Print(fmt.Sprintf("            [%s]\n", info.Cmd.DisplayPath(sep, true)))
		}
	}
}

func CollectDepends(cc *core.Cli, flow []core.ParsedCmd, res Depends) {
	for _, it := range flow {
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
				res[dep.OsCmd] = map[*core.Cmd]DependInfo{cic: DependInfo{dep.Reason, it}}
			}
		}
		if cic.Type() != core.CmdTypeFlow {
			continue
		}
		subFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, cic.Flow()...)
		CollectDepends(cc, subFlow.Cmds, res)
	}
}
