package display

import (
	"fmt"
	"os/exec"
	"sort"

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

	foundOsCmds := map[string]bool{}
	hasMissedOsCmd := false
	var osCmds []string
	for osCmd, _ := range deps {
		exists := isOsCmdExists(osCmd)
		foundOsCmds[osCmd] = exists
		hasMissedOsCmd = hasMissedOsCmd || !exists
		osCmds = append(osCmds, osCmd)
	}
	sort.Strings(osCmds)

	sep := env.Get("strs.cmd-path-sep").Raw
	notFoundStr := "(not found)"
	if env.GetBool("display.utf8.symbols") {
		notFoundStr = "⛔"
	}

	if !env.GetBool("display.flow.simplified") {
		tipFunc := PrintTipTitle
		if hasMissedOsCmd {
			tipFunc = PrintErrTitle
		}
		tipFunc(cc.Screen, env,
			"depended os commands.",
			"",
			"this flow need the os commands below to execute,",
			"make sure they are all installed.")
	} else {
		cc.Screen.Print(fmt.Sprintf("-------=<%s>=-------\n\n", "depended os commands"))
	}

	for _, osCmd := range osCmds {
		cmds := deps[osCmd]
		statusStr := ""
		if !foundOsCmds[osCmd] {
			statusStr = notFoundStr
		}
		cc.Screen.Print(fmt.Sprintf("[%s]%s\n", osCmd, statusStr))
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

func isOsCmdExists(cmd string) bool {
	path, err := exec.LookPath(cmd)
	return err == nil && len(path) > 0
}
