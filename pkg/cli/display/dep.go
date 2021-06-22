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

func DumpDepends(
	screen core.Screen,
	env *core.Env,
	deps Depends) (hasMissedOsCmd bool) {

	if len(deps) == 0 {
		return
	}

	foundOsCmds := map[string]bool{}
	var osCmds []string
	for osCmd, _ := range deps {
		exists := isOsCmdExists(osCmd)
		foundOsCmds[osCmd] = exists
		hasMissedOsCmd = hasMissedOsCmd || !exists
		osCmds = append(osCmds, osCmd)
	}
	sort.Strings(osCmds)

	sep := env.Get("strs.cmd-path-sep").Raw

	if !env.GetBool("display.flow.simplified") {
		if hasMissedOsCmd {
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
	} else {
		screen.Print(fmt.Sprintf("-------=<%s>=-------\n\n", "depended os-commands"))
	}

	for _, osCmd := range osCmds {
		if hasMissedOsCmd && foundOsCmds[osCmd] {
			continue
		}
		cmds := deps[osCmd]
		screen.Print(fmt.Sprintf("[%s]\n", osCmd))

		// TODO: sort cmds
		for _, info := range cmds {
			screen.Print(fmt.Sprintf("        '%s'\n", info.Reason))
			screen.Print(fmt.Sprintf("            [%s]\n", info.Cmd.DisplayPath(sep, true)))
		}
	}

	return
}

func CollectDepends(
	cc *core.Cli,
	env *core.Env,
	flow []core.ParsedCmd,
	res Depends,
	allowFlowTemplateRenderError bool) {

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
		cmdEnv, _ := it.GenEnvAndArgv(env, cc.Cmds.Strs.EnvValDelAllMark, cc.Cmds.Strs.PathSep)
		if cic.Type() != core.CmdTypeFlow {
			continue
		}
		subFlow, rendered := cic.Flow(cmdEnv, allowFlowTemplateRenderError)
		if rendered && len(subFlow) != 0 {
			parsedFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
			CollectDepends(cc, cmdEnv, parsedFlow.Cmds, res, allowFlowTemplateRenderError)
		}
	}
}

func isOsCmdExists(cmd string) bool {
	path, err := exec.LookPath(cmd)
	return err == nil && len(path) > 0
}
