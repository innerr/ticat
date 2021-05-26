package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func GlobalHelp(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	if len(argv.GetRaw("1st-str")) != 0 {
		ok := FindAny(argv, cc, env)
		flow.Cmds = nil
		return 0, ok
	}

	if len(flow.Cmds) > 2 {
		return DumpFlow(argv, cc, env, flow, currCmdIdx)
	} else if len(flow.Cmds) == 2 {
		display.DumpCmds(cc, false, 4, false,
			flow.Cmds[1].DisplayPath(cc.Cmds.Strs.PathSep, false))
		flow.Cmds = nil
		return 0, true
	}

	pln := func(text string) {
		cc.Screen.Print(text + "\n")
	}

	pln("usages:")
	pln("    list all cmds:                 - ticat cmd.tree")
	pln("    find cmds or env KVs:          - ticat find example")
	pln("                                   - ticat find example golang")
	pln("                                   - ticat find str1 str2 str3")
	pln("    list all env KVs:              - ticat env.tree")
	pln("    execute a cmd with args:       - ticat example.golang arg1 arg2")
	pln("                                     ticat example.golang {arg1 arg2}")
	pln("                                     ticat example.golang {a=arg1 b=arg2}")
	pln("    execute a list of cmd:         - ticat cmd1 : cmd2 : cmd3")
	pln("    check and desc cmd list:       - ticat cmd1 : cmd2 : cmd3 : desc")
	pln("    set env KVs when executing:    - ticat cmd1 : {display.style=ascii} cmd2")
	pln("    set session-global env KVs:    - ticat {display.width=120} : cmd1 : cmd2")
	pln("    set env KVs and save to local: - ticat {display.width=120} : env.save")
	pln("    use abbrs in cmd:              - ticat exam.go arg1 arg2")
	pln("    use abbrs in env KVs setting:  - ticat {disp.w=120} : cmd1 : cmd2")
	pln("                                   - ticat {disp.w=120} : e.s")

	flow.Cmds = nil
	return 0, true
}

func FindAny(argv core.ArgVals, cc *core.Cli, env *core.Env) bool {
	findStrs := getFindStrsFromArgv(argv)
	if len(findStrs) == 0 {
		return true
	}
	display.DumpEnvFlattenVals(cc.Screen, env, findStrs...)
	display.DumpCmds(cc, false, 4, true, "", findStrs...)
	return true
}
