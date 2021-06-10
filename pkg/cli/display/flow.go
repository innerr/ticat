package display

import (
	"sort"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func DumpFlow(
	cc *core.Cli,
	env *core.Env,
	flow []core.ParsedCmd,
	sep string,
	indentSize int,
	simple bool,
	skeleton bool) {

	depth := env.GetInt("display.flow.depth")

	if skeleton {
		simple = true
	}

	if len(flow) == 0 {
		return
	}

	PrintTipTitle(cc.Screen, env, "flow executing description:")
	cc.Screen.Print("--->>>\n")
	dumpFlow(cc, env, flow, depth, sep, indentSize, simple, skeleton, 0)
	cc.Screen.Print("<<<---\n")
}

func dumpFlow(
	cc *core.Cli,
	env *core.Env,
	flow []core.ParsedCmd,
	depth int,
	sep string,
	indentSize int,
	simple bool,
	skeleton bool,
	indentAdjust int) {

	for _, cmd := range flow {
		if !cmd.IsEmpty() {
			dumpFlowCmd(cc, cc.Screen, env, cmd, depth, sep, indentSize,
				simple, skeleton, indentAdjust)
		}
	}
}

func dumpFlowCmd(
	cc *core.Cli,
	screen core.Screen,
	env *core.Env,
	parsedCmd core.ParsedCmd,
	depth int,
	sep string,
	indentSize int,
	simple bool,
	skeleton bool,
	indentAdjust int) {

	cmd := parsedCmd.Last().Matched.Cmd
	if cmd == nil {
		return
	}

	envOpSep := " " + cmd.Strs.EnvOpSep + " "

	prt := func(indentLvl int, msg string) {
		indentLvl += indentAdjust
		padding := rpt(" ", indentSize*indentLvl)
		msg = autoPadNewLine(padding, msg)
		screen.Print(padding + msg + "\n")
	}

	cic := cmd.Cmd()
	if cic == nil {
		return
	}
	var name string
	if skeleton {
		name = strings.Join(parsedCmd.Path(), sep)
	} else {
		name = parsedCmd.DisplayPath(sep, true)
	}
	prt(0, "["+name+"]")
	if len(cic.Help()) != 0 {
		prt(1, " '"+cic.Help()+"'")
	}

	cmdEnv := parsedCmd.GenEnv(env, cc.Cmds.Strs.EnvValDelAllMark)
	if !skeleton {
		args := parsedCmd.Args()
		argv := cmdEnv.GetArgv(parsedCmd.Path(), sep, args)
		argLines := DumpArgs(&args, argv, true)
		if len(argLines) != 0 {
			prt(1, "- args:")
		}
		for _, line := range argLines {
			prt(2, line)
		}
	}

	if !skeleton {
		// TODO: missed kvs in GlobalEnv
		cmdEnv = parsedCmd.GenEnv(core.NewEnv(), cc.Cmds.Strs.EnvValDelAllMark)
		flatten := cmdEnv.Flatten(false, nil, true)
		if len(flatten) != 0 {
			prt(1, "- env-values:")
			var keys []string
			for k, _ := range flatten {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				prt(2, k+" = "+flatten[k])
			}
		}
	}

	if !skeleton {
		envOps := cic.EnvOps()
		envOpKeys := envOps.EnvKeys()
		if len(envOpKeys) != 0 {
			prt(1, "- env-ops:")
		}
		for _, k := range envOpKeys {
			prt(2, k+" = "+dumpEnvOps(envOps.Ops(k), envOpSep))
		}
	}

	if !simple && !skeleton {
		line := string(cic.Type())
		if cic.IsQuiet() {
			line += " (quiet)"
		}
		if cic.IsPriority() {
			line += " (priority)"
		}
		prt(1, "- cmd-type:")
		prt(2, line)

		if len(cic.Source()) != 0 && !strings.HasPrefix(cic.CmdLine(), cic.Source()) {
			prt(1, "- from:")
			prt(2, cic.Source())
		}
	}

	if len(cic.CmdLine()) != 0 && cic.Type() != core.CmdTypeNormal &&
		cic.Type() != core.CmdTypePower {
		if cic.Type() == core.CmdTypeFlow && !skeleton {
			prt(1, "- flow:")
			prt(2, cic.CmdLine())
		} else if !simple && !skeleton {
			if cic.Type() == core.CmdTypeEmptyDir {
				prt(1, "- dir:")
				prt(2, cic.CmdLine())
			} else {
				prt(1, "- executable:")
				prt(2, cic.CmdLine())
			}
			if len(cic.MetaFile()) != 0 {
				prt(1, "- meta:")
				prt(2, cic.MetaFile())
			}
		}
		if cic.Type() == core.CmdTypeFlow && depth > 1 {
			prt(2, "--->>>")
			subFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, cic.Flow()...)
			dumpFlow(cc, env, subFlow.Cmds, depth-1, sep, indentSize, simple, skeleton, indentAdjust+2)
			prt(2, "<<<---")
		}
	}
}
