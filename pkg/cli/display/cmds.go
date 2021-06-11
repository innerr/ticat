package display

import (
	"fmt"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func DumpSubCmds(
	screen core.Screen,
	cmd *core.CmdTree,
	indentSize int,
	findStrs ...string) {

	dumpCmd(screen, cmd, true, indentSize,
		true, false, -cmd.Depth(), findStrs...)
}

func DumpAllCmds(
	cmds *core.CmdTree,
	screen core.Screen,
	skeleton bool,
	indentSize int,
	flatten bool,
	recursive bool,
	findStrs ...string) {

	dumpCmd(screen, cmds, skeleton, indentSize,
		recursive, flatten, -cmds.Depth(), findStrs...)
}

// TODO: remove this
func DumpCmds(
	cc *core.Cli,
	skeleton bool,
	indentSize int,
	flatten bool,
	recursive bool,
	path string,
	findStrs ...string) {

	if len(path) == 0 && !recursive {
		return
	}

	cmds := cc.Cmds
	if len(path) != 0 {
		cmds = cmds.GetSub(strings.Split(path, cc.Cmds.Strs.PathSep)...)
		if cmds == nil {
			// TODO: better display
			panic(fmt.Errorf("[DumpCmds] can't find sub cmd tree by path '%s'", path))
		}
	}
	dumpCmd(cc.Screen, cmds, skeleton, indentSize,
		recursive, flatten, -cmds.Depth(), findStrs...)
}

func dumpCmd(
	screen core.Screen,
	cmd *core.CmdTree,
	skeleton bool,
	indentSize int,
	recursive bool,
	flatten bool,
	indentAdjust int,
	findStrs ...string) {

	if cmd == nil || cmd.IsHidden() {
		return
	}

	abbrsSep := cmd.Strs.AbbrsSep
	envOpSep := " " + cmd.Strs.EnvOpSep + " "
	indent := cmd.Depth() + indentAdjust

	prt := func(indentLvl int, msg string) {
		if !flatten {
			indentLvl += indent
		}
		padding := rpt(" ", indentSize*indentLvl)
		msg = autoPadNewLine(padding, msg)
		screen.Print(padding + msg + "\n")
	}

	if cmd.Parent() == nil || cmd.MatchFind(findStrs...) {
		cic := cmd.Cmd()
		var name string
		if flatten {
			name = cmd.DisplayPath()
		} else if !skeleton {
			name = strings.Join(cmd.Abbrs(), abbrsSep)
		} else {
			name = cmd.DisplayName()
		}
		if len(name) == 0 {
			name = cmd.DisplayName()
		}

		if !flatten || cic != nil {
			prt(0, "["+name+"]")
			if cic != nil {
				var helpStr string
				if !skeleton {
					helpStr = cic.Help()
				} else {
					helpStr = cic.DisplayHelpStr()
				}
				if len(helpStr) != 0 {
					prt(1, " '"+helpStr+"'")
				}
			}
			full := cmd.DisplayPath()
			if cmd.Parent() != nil && cmd.Parent().Parent() != nil {
				if !skeleton && !flatten {
					prt(1, "- full-cmd:")
					prt(2, full)
				}
			}
			if !skeleton {
				abbrs := cmd.DisplayAbbrsPath()
				if len(abbrs) != 0 && abbrs != full {
					prt(1, "- full-abbrs:")
					prt(2, abbrs)
				}
			}
		}

		if !skeleton && cic != nil {
			args := cic.Args()
			argNames := args.Names()
			if len(argNames) != 0 {
				prt(1, "- args:")
			}
			for _, name := range argNames {
				val := args.DefVal(name)
				nameStr := strings.Join(args.Abbrs(name), abbrsSep)
				prt(2, nameStr+" = "+MayQuoteStr(val))
			}

			envOps := cic.EnvOps()
			envOpKeys := envOps.EnvKeys()
			if len(envOpKeys) != 0 {
				prt(1, "- env-ops:")
			}
			for _, k := range envOpKeys {
				prt(2, k+" = "+dumpEnvOps(envOps.Ops(k), envOpSep))
			}

			deps := cic.GetDepends()
			if len(deps) != 0 {
				prt(1, "- deps:")
			}
			for _, dep := range deps {
				prt(2, dep.OsCmd+" = '"+dep.Reason+"'")
			}

			if cic.Type() != core.CmdTypeFlow && (cic.Type() != core.CmdTypeNormal || cic.IsQuiet()) {
				line := string(cic.Type())
				if cic.IsQuiet() {
					line += " (quiet)"
				}
				if cic.IsPriority() {
					line += " (priority)"
				}
				prt(1, "- cmd-type:")
				prt(2, line)
			}

			if len(cic.Source()) == 0 || !strings.HasPrefix(cic.CmdLine(), cic.Source()) {
				prt(1, "- from:")
				if len(cic.Source()) == 0 {
					prt(2, "builtin")
				} else {
					prt(2, cic.Source())
				}
			}

			if cic.Type() != core.CmdTypeNormal && cic.Type() != core.CmdTypePower {
				if len(cic.CmdLine()) != 0 {
					if cic.Type() == core.CmdTypeFlow {
						prt(1, "- flow:")
					} else if cic.Type() == core.CmdTypeEmptyDir {
						prt(1, "- dir:")
					} else {
						prt(1, "- executable:")
					}
					prt(2, cic.CmdLine())
				}
				if len(cic.MetaFile()) != 0 {
					prt(1, "- meta:")
					prt(2, cic.MetaFile())
				}
			}
		}
	}

	if recursive {
		for _, name := range cmd.SubNames() {
			dumpCmd(screen, cmd.GetSub(name), skeleton, indentSize,
				recursive, flatten, indentAdjust, findStrs...)
		}
	}
}
