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
	dumpCmd(buf, env, cmds, args, -cmds.Depth())

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

	dumpCmd(screen, env, cmds, args, -cmds.Depth())
}

type DumpCmdArgs struct {
	Skeleton   bool
	ShowUsage  bool
	Flatten    bool
	Recursive  bool
	FindStrs   []string
	FindByTags bool
	IndentSize int
}

func NewDumpCmdArgs() *DumpCmdArgs {
	return &DumpCmdArgs{false, true, true, true, nil, false, 4}
}

func (self *DumpCmdArgs) NoShowShowUsage() *DumpCmdArgs {
	self.ShowUsage = false
	return self
}

func (self *DumpCmdArgs) SetShowUsage() *DumpCmdArgs {
	self.ShowUsage = true
	return self
}

func (self *DumpCmdArgs) SetSkeleton() *DumpCmdArgs {
	self.Skeleton = true
	self.ShowUsage = false
	return self
}

func (self *DumpCmdArgs) SetFindByTags() *DumpCmdArgs {
	self.FindByTags = true
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
	env *core.Env,
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

	if cmd.Parent() == nil ||
		(!args.FindByTags && cmd.MatchFind(args.FindStrs...) ||
			cmd.MatchTags(args.FindStrs...)) {

		cic := cmd.Cmd()
		var name string
		abbrs := cmd.Abbrs()
		if args.Flatten {
			name = cmd.DisplayPath()
		} else if !args.Skeleton && len(abbrs) > 1 {
			name = strings.Join(abbrs, abbrsSep)
		} else {
			name = cmd.DisplayPath()
		}
		if len(name) == 0 {
			name = cmd.DisplayPath()
		}

		if !args.Flatten || cic != nil {
			prt(0, ColorCmd("["+name+"]", env))

			if (!args.Skeleton || args.FindByTags) && len(cmd.Tags()) != 0 {
				prt(1, ColorTag(" @"+strings.Join(cmd.Tags(), " @"), env))
			}

			// TODO: move 'help' from core.Cmd to core.CmdTree
			if cic != nil {
				var helpStr string
				if !args.Skeleton {
					helpStr = cic.Help()
				} else {
					helpStr = cic.DisplayHelpStr()
				}
				if len(helpStr) != 0 {
					prt(1, " "+ColorHelp("'"+helpStr+"'", env))
				}
			}

			full := cmd.DisplayPath()
			if cmd.Parent() != nil && cmd.Parent().Parent() != nil {
				if !args.Skeleton && !args.Flatten && full != name {
					prt(1, ColorProp("- full-cmd:", env))
					prt(2, full)
				}
			}
			if !args.Skeleton || args.ShowUsage {
				abbrs := cmd.DisplayAbbrsPath()
				if len(abbrs) != 0 && abbrs != full {
					prt(1, ColorProp("- full-abbrs:", env))
					prt(2, abbrs)
				}
			}
		}

		if (!args.Skeleton || args.ShowUsage) && cic != nil {
			args := cic.Args()
			argNames := args.Names()
			if len(argNames) != 0 {
				prt(1, ColorProp("- args:", env))
			}
			for _, name := range argNames {
				val := args.DefVal(name)
				nameStr := strings.Join(args.Abbrs(name), abbrsSep)
				prt(2, ColorArg(nameStr, env)+ColorSymbol(" = ", env)+mayQuoteStr(val))
			}
		}

		if !args.Skeleton && cic != nil {
			val2env := cic.GetVal2Env()
			if len(val2env.EnvKeys()) != 0 {
				prt(1, ColorProp("- env-direct-write:", env))
			}
			for _, k := range val2env.EnvKeys() {
				prt(2, ColorKey(k, env)+ColorSymbol(" = ", env)+mayQuoteStr(val2env.Val(k)))
			}

			arg2env := cic.GetArg2Env()
			if len(arg2env.EnvKeys()) != 0 {
				prt(1, ColorProp("- env-from-argv:", env))
			}
			for _, k := range arg2env.EnvKeys() {
				prt(2, ColorKey(k, env)+ColorSymbol(" <- ", env)+ColorArg(mayQuoteStr(arg2env.GetArgName(k)), env))
			}

			envOps := cic.EnvOps()
			envOpKeys := envOps.EnvKeys()
			if len(envOpKeys) != 0 {
				prt(1, ColorProp("- env-ops:", env))
			}
			for _, k := range envOpKeys {
				prt(2, ColorKey(k, env)+ColorSymbol(" = ", env)+dumpEnvOps(envOps.Ops(k), envOpSep))
			}

			deps := cic.GetDepends()
			if len(deps) != 0 {
				prt(1, ColorProp("- os-cmd-dep:", env))
			}
			for _, dep := range deps {
				prt(2, ColorCmdDone(dep.OsCmd, env)+ColorSymbol(" = ", env)+dep.Reason)
			}

			if cic.Type() != core.CmdTypeFlow && (cic.Type() != core.CmdTypeNormal || cic.IsQuiet()) {
				line := string(cic.Type())
				if cic.IsQuiet() {
					line += " (quiet)"
				}
				if cic.IsPriority() {
					line += " (priority)"
				}
				prt(1, ColorProp("- cmd-type:", env))
				prt(2, line)
			}

			if cic.Type() != core.CmdTypeNormal && cic.Type() != core.CmdTypePower {
				if len(cic.CmdLine()) != 0 || len(cic.FlowStrs()) != 0 {
					if cic.Type() == core.CmdTypeFlow {
						prt(1, ColorProp("- flow:", env))
						for _, flowStr := range cic.FlowStrs() {
							prt(2, ColorFlow(flowStr, env))
						}
					} else if cic.Type() == core.CmdTypeEmptyDir {
						prt(1, ColorProp("- dir:", env))
						prt(2, cic.CmdLine())
					} else {
						prt(1, ColorProp("- executable:", env))
						prt(2, cic.CmdLine())
					}
				}
			}

			if len(cmd.Source()) == 0 || !strings.HasPrefix(cic.CmdLine(), cmd.Source()) {
				prt(1, ColorProp("- from:", env))
				if len(cmd.Source()) == 0 {
					prt(2, builtinName)
				} else {
					prt(2, cmd.Source())
				}
			}

			if cic.Type() != core.CmdTypeNormal && cic.Type() != core.CmdTypePower {
				if len(cic.MetaFile()) != 0 {
					prt(1, ColorProp("- meta:", env))
					prt(2, cic.MetaFile())
				}
			}
		}
	}

	if args.Recursive {
		for _, name := range cmd.SubNames() {
			dumpCmd(screen, env, cmd.GetSub(name), args, indentAdjust)
		}
	}
}
