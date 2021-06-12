package display

import (
	"fmt"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func DumpCmdsByPath(cc *core.Cli, args *DumpCmdArgs, path string) {
	if len(path) == 0 && !args.Recursive {
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
	dumpCmd(cc.Screen, cmds, args, -cmds.Depth())
}

func DumpCmds(cmds *core.CmdTree, screen core.Screen, args *DumpCmdArgs) {
	dumpCmd(screen, cmds, args, -cmds.Depth())
}

type DumpCmdArgs struct {
	Skeleton   bool
	Flatten    bool
	Recursive  bool
	FindStrs   []string
	IndentSize int
}

func NewDumpCmdArgs() *DumpCmdArgs {
	return &DumpCmdArgs{false, true, true, nil, 4}
}

func (self *DumpCmdArgs) SetSkeleton() *DumpCmdArgs {
	self.Skeleton = true
	return self
}

func (self *DumpCmdArgs) NoFlatten() *DumpCmdArgs {
	self.Flatten = false
	return self
}

func (self *DumpCmdArgs) NoRecursive() *DumpCmdArgs {
	self.Recursive = false
	return self
}

func (self *DumpCmdArgs) AddFindStrs(findStrs ...string) *DumpCmdArgs {
	self.FindStrs = append(self.FindStrs, findStrs...)
	return self
}

func dumpCmd(
	screen core.Screen,
	cmd *core.CmdTree,
	args *DumpCmdArgs,
	indentAdjust int) {

	if cmd == nil || cmd.IsHidden() {
		return
	}

	builtinName := cmd.Strs.BuiltinDisplayName
	abbrsSep := cmd.Strs.AbbrsSep
	envOpSep := " " + cmd.Strs.EnvOpSep + " "
	indent := cmd.Depth() + indentAdjust

	prt := func(indentLvl int, msg string) {
		if !args.Flatten {
			indentLvl += indent
		}
		padding := rpt(" ", args.IndentSize*indentLvl)
		msg = autoPadNewLine(padding, msg)
		screen.Print(padding + msg + "\n")
	}

	if cmd.Parent() == nil || cmd.MatchFind(args.FindStrs...) {
		cic := cmd.Cmd()
		var name string
		if args.Flatten {
			name = cmd.DisplayPath()
		} else if !args.Skeleton {
			name = strings.Join(cmd.Abbrs(), abbrsSep)
		} else {
			name = cmd.DisplayName()
		}
		if len(name) == 0 {
			name = cmd.DisplayName()
		}

		if !args.Flatten || cic != nil {
			prt(0, "["+name+"]")
			if cic != nil {
				var helpStr string
				if !args.Skeleton {
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
				if !args.Skeleton && !args.Flatten {
					prt(1, "- full-cmd:")
					prt(2, full)
				}
			}
			if !args.Skeleton {
				abbrs := cmd.DisplayAbbrsPath()
				if len(abbrs) != 0 && abbrs != full {
					prt(1, "- full-abbrs:")
					prt(2, abbrs)
				}
			}
		}

		if !args.Skeleton && cic != nil {
			args := cic.Args()
			argNames := args.Names()
			if len(argNames) != 0 {
				prt(1, "- args:")
			}
			for _, name := range argNames {
				val := args.DefVal(name)
				nameStr := strings.Join(args.Abbrs(name), abbrsSep)
				prt(2, nameStr+" = "+mayQuoteStr(val))
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
					prt(2, builtinName)
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

	if args.Recursive {
		for _, name := range cmd.SubNames() {
			dumpCmd(screen, cmd.GetSub(name), args, indentAdjust)
		}
	}
}
