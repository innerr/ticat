package display

import (
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func DumpCmdsWithTips(
	cmds *core.CmdTree,
	screen core.Screen,
	env *core.Env,
	args *DumpCmdArgs,
	displayCmdPath string,
	isLessMore bool) {

	prt := func(text ...interface{}) {
		PrintTipTitle(screen, env, text...)
	}

	buf := NewCacheScreen()
	dumpCmd(buf, cmds, args, -cmds.Depth())

	findStr := strings.Join(args.FindStrs, " ")
	selfName := env.GetRaw("strs.self-name")

	if !args.Recursive {
		prt("command details:")
	} else if len(args.FindStrs) != 0 {
		tip := "search "
		matchStr := " commands matched '" + findStr + "'"
		if !cmds.IsRoot() {
			if buf.OutputNum() > 0 {
				prt(tip + "branch '" + displayCmdPath + "', found" + matchStr + ":")
			} else {
				prt(tip + "branch '" + displayCmdPath + "', no" + matchStr + ".")
			}
		} else {
			if buf.OutputNum() > 0 {
				if args.Skeleton && buf.OutputNum() <= 6 && isLessMore {
					prt(tip+"and found"+matchStr,
						"",
						"get more details by using '-' instead of '+'.")
				} else {
					prt(tip + "and found" + matchStr)
				}
			} else {
				prt(tip + "but no" + matchStr)
			}
		}
	} else {
		if !cmds.IsRoot() {
			if buf.OutputNum() > 0 {
				prt("branch '" + displayCmdPath + "' has commands:")
			} else {
				prt("branch '" + displayCmdPath + "' has no commands. (this should never happen)")
			}
		} else {
			if buf.OutputNum() > 0 {
				prt("all commands loaded to " + selfName + ":")
			} else {
				prt(selfName + " has no loaded commands. (this should never happen)")
			}
		}
	}

	buf.WriteTo(screen)

	if !args.Recursive || !TooMuchOutput(env, buf) {
		return
	}

	if !args.Flatten {
		prt(
			"locate exact commands by:",
			"",
			SuggestFindCmds(env))
	} else {
		if !isLessMore {
			if len(args.FindStrs) != 0 {
				prt("locate exact commands by adding more keywords.")
			} else {
				prt(
					"locate exact commands by:",
					"",
					SuggestFindCmds(env))
			}
		} else if !args.Skeleton {
			prt(
				"get a brief view by using '-' instead of '+'.",
				"",
				"or/and locate exact commands by adding more keywords:",
				"",
				SuggestFindCmds(env))
		} else {
			prt(
				"locate exact commands by adding more keywords:",
				"",
				SuggestFindCmdsLess(env))
		}
	}
}

func DumpCmds(
	cmds *core.CmdTree,
	screen core.Screen,
	env *core.Env,
	args *DumpCmdArgs) {

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

			val2env := cic.GetVal2Env()
			if len(val2env.EnvKeys()) != 0 {
				prt(1, "- env-ops: (write)")
			}
			for _, k := range val2env.EnvKeys() {
				prt(2, k+" = "+mayQuoteStr(val2env.Val(k)))
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
				prt(1, "- os-cmd-dep:")
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
				if len(cic.CmdLine()) != 0 || len(cic.FlowStrs()) != 0 {
					if cic.Type() == core.CmdTypeFlow {
						prt(1, "- flow:")
						for _, flowStr := range cic.FlowStrs() {
							prt(2, flowStr)
						}
					} else if cic.Type() == core.CmdTypeEmptyDir {
						prt(1, "- dir:")
						prt(2, cic.CmdLine())
					} else {
						prt(1, "- executable:")
						prt(2, cic.CmdLine())
					}
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
