package display

import (
	"fmt"
	"os/exec"
	"sort"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func GatherOsCmdsExistingInfo(deps core.Depends) (foundOsCmds map[string]bool, osCmds []string, missedOsCmds int) {
	foundOsCmds = map[string]bool{}
	for osCmd, _ := range deps {
		exists := isOsCmdExists(osCmd)
		foundOsCmds[osCmd] = exists
		if !exists {
			missedOsCmds += 1
		}
		osCmds = append(osCmds, osCmd)
	}
	sort.Strings(osCmds)
	return
}

func DumpDepends(
	screen core.Screen,
	env *core.Env,
	deps core.Depends) (hasMissedOsCmd bool) {

	if len(deps) == 0 {
		return
	}

	foundOsCmds, osCmds, missedOsCmds := GatherOsCmdsExistingInfo(deps)

	sep := env.Get("strs.cmd-path-sep").Raw

	if missedOsCmds > 0 {
		PrintErrTitle(screen, env,
			"missed depended os-commands.",
			"",
			"the needed os-commands below are not installed:")
	} else {
		PrintTipTitle(screen, env,
			"depended os-commands are all installed.",
			"",
			"this flow need these os-commands below to execute:")
	}

	for _, osCmd := range osCmds {
		if missedOsCmds > 0 && foundOsCmds[osCmd] {
			continue
		}
		cmds := deps[osCmd]
		screen.Print(ColorCmd(fmt.Sprintf("[%s]\n", osCmd), env))

		// TODO: sort cmds
		for _, info := range cmds {
			screen.Print("        " + ColorHelp(fmt.Sprintf("'%s'\n", info.Reason), env))
			screen.Print("            " + ColorCmd(fmt.Sprintf("[%s]\n", info.Cmd.DisplayPath(sep, true)), env))
		}
	}

	return missedOsCmds > 0
}

func isOsCmdExists(cmd string) bool {
	path, err := exec.LookPath(cmd)
	return err == nil && len(path) > 0
}
