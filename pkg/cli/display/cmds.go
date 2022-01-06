package display

import (
	"fmt"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func DumpCmdsWithTips(
	cmds *core.CmdTree,
	screen core.Screen,
	env *core.Env,
	args *DumpCmdArgs,
	displayCmdPath string,
	lessDetailCmd string,
	moreDetailCmd string) (allShown bool) {

	if !args.Recursive {
		panic(fmt.Errorf("should never happen, this func is not for non-recursive dumping display"))
	}

	prt := func(text ...interface{}) {
		PrintTipTitle(screen, env, text...)
	}

	buf := NewCacheScreen()
	allShown = dumpCmd(buf, env, cmds, args, -cmds.Depth(), 0)

	findStr := strings.Join(args.FindStrs, " ")

	padPropName := func(s string) string {
		return s + rpt(" ", 12-len(s))
	}

	filterLines := []string{}
	if len(displayCmdPath) != 0 {
		filterLines = append(filterLines, padPropName("- branch:")+displayCmdPath)
	}
	if len(args.Source) != 0 {
		filterLines = append(filterLines, padPropName("- source:")+args.Source)
	}
	if len(args.FindStrs) != 0 {
		if args.FindByTags {
			filterLines = append(filterLines, padPropName("- tags:")+findStr)
		} else {
			filterLines = append(filterLines, padPropName("- keywords:")+findStr)
		}
	}
	if len(filterLines) > 0 {
		filterLines = append([]string{""}, filterLines...)
	}

	var footerLines []string
	if allShown {
		if buf.OutputNum() <= 6 && args.Skeleton && len(moreDetailCmd) > 0 {
			footerLines = append(footerLines, "", "use '"+moreDetailCmd+"' to show more details")
		}
	} else {
		footerLines = append(footerLines, "", fmt.Sprintf("some may not shown by arg depth='%d'", args.MaxDepth))
	}

	title := ""
	if buf.OutputNum() > 0 {
		if len(filterLines) != 0 {
			title = "found commands matched:"
		} else {
			if allShown {
				title = "all commands:"
			} else {
				title = "commands:"
			}
		}
	} else {
		if len(filterLines) != 0 {
			title = "no commands matched:"
		} else {
			panic(fmt.Errorf("should never happen, no loaded commands"))
		}
	}

	prt(title, filterLines, footerLines)

	buf.WriteTo(screen)

	if TooMuchOutput(env, buf) {
		if (!args.Skeleton || args.ShowUsage) && len(lessDetailCmd) > 0 {
			prt("get a brief view by using '" + lessDetailCmd + "'")
		} else if !args.FindByTags {
			if len(args.FindStrs) != 0 {
				prt("narrow down results by adding more keywords.")
			} else {
				prt("narrow down results by using keywords.")
			}
		}
	}
	return
}

func DumpCmds(
	cmds *core.CmdTree,
	screen core.Screen,
	env *core.Env,
	args *DumpCmdArgs) (allShown bool) {

	return dumpCmd(screen, env, cmds, args, -cmds.Depth(), 0)
}

type DumpCmdArgs struct {
	Skeleton      bool
	ShowUsage     bool
	Flatten       bool
	Recursive     bool
	FindStrs      []string
	FindByTags    bool
	IndentSize    int
	MatchWriteKey string
	Source        string
	MaxDepth      int
}

func NewDumpCmdArgs() *DumpCmdArgs {
	return &DumpCmdArgs{false, true, true, true, nil, false, 4, "", "", 0}
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

func (self *DumpCmdArgs) SetMatchWriteKey(key string) *DumpCmdArgs {
	self.MatchWriteKey = key
	return self
}

func (self *DumpCmdArgs) SetSource(source string) *DumpCmdArgs {
	self.Source = source
	return self
}

func (self *DumpCmdArgs) SetMaxDepth(depth int) *DumpCmdArgs {
	self.MaxDepth = depth
	return self
}

func (self *DumpCmdArgs) MatchFind(cmd *core.CmdTree) bool {
	if len(self.MatchWriteKey) != 0 && !cmd.MatchWriteKey(self.MatchWriteKey) {
		return false
	}
	if len(self.FindStrs) != 0 {
		if self.FindByTags {
			if !cmd.MatchTags(self.FindStrs...) {
				return false
			}
		} else {
			if !cmd.MatchFind(self.FindStrs...) {
				return false
			}
		}
	}
	if len(self.Source) != 0 && !cmd.MatchSource(self.Source) {
		return false
	}

	return true
}

func dumpCmd(
	screen core.Screen,
	env *core.Env,
	cmd *core.CmdTree,
	args *DumpCmdArgs,
	indentAdjust int,
	depth int) (allShown bool) {

	if cmd == nil || cmd.IsHidden() {
		return true
	}

	builtinName := cmd.Strs.BuiltinDisplayName
	abbrsSep := cmd.Strs.AbbrsSep
	tagMark := cmd.Strs.TagMark
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

	if cmd.Parent() == nil || args.MatchFind(cmd) {
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

		//if !args.Flatten || cic != nil {
		if !args.Flatten || cmd.Parent() != nil {
			prt(0, cmdIdStr(cmd, name, env))

			if (!args.Skeleton || args.FindByTags) && len(cmd.Tags()) != 0 {
				prt(1, ColorTag(" "+tagMark+strings.Join(cmd.Tags(), " "+tagMark), env))
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
				val = mayMaskSensitiveVal(nameStr, val)
				prt(2, ColorArg(nameStr, env)+ColorSymbol(" = ", env)+mayQuoteStr(val))
			}
		}

		if !args.Skeleton && cic != nil {
			val2env := cic.GetVal2Env()
			if len(val2env.EnvKeys()) != 0 {
				prt(1, ColorProp("- env-direct-write:", env))
			}
			for _, k := range val2env.EnvKeys() {
				val := val2env.Val(k)
				val = mayMaskSensitiveVal(k, val)
				prt(2, ColorKey(k, env)+ColorSymbol(" = ", env)+mayQuoteStr(val))
			}

			arg2env := cic.GetArg2Env()
			if len(arg2env.EnvKeys()) != 0 {
				prt(1, ColorProp("- env-from-argv:", env))
			}
			for _, k := range arg2env.EnvKeys() {
				val := arg2env.GetArgName(cic, k, true)
				val = mayMaskSensitiveVal(k, val)
				prt(2, ColorKey(k, env)+ColorSymbol(" <- ", env)+
					ColorArg(mayQuoteStr(val), env))
			}

			envOps := cic.EnvOps()
			envOpKeys := envOps.RawEnvKeys()
			if len(envOpKeys) != 0 {
				prt(1, ColorProp("- env-ops:", env))
			}
			for _, k := range envOpKeys {
				prt(2, ColorKey(k, env)+ColorSymbol(" = ", env)+dumpEnvOps(envOps.Ops(k), envOpSep)+dumpIsAutoTimerKey(env, cic, k))
			}

			deps := cic.GetDepends()
			if len(deps) != 0 {
				prt(1, ColorProp("- os-cmd-dep:", env))
			}
			for _, dep := range deps {
				prt(2, ColorCmdDone(dep.OsCmd, env)+ColorSymbol(" = ", env)+dep.Reason)
			}

			// TODO: a bit messy
			//if !cic.HasSubFlow() && (cic.Type() != core.CmdTypeNormal || cic.IsQuiet()) {
			if cic.Type() != core.CmdTypeFlow || cic.Type() != core.CmdTypeAdHotFlow {
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

			// TODO: a bit messy
			if cic.Type() != core.CmdTypeNormal && cic.Type() != core.CmdTypePower {
				if len(cic.CmdLine()) != 0 || len(cic.FlowStrs()) != 0 {
					if cic.HasSubFlow() {
						prt(1, ColorProp("- flow:", env))
						for _, flowStr := range cic.FlowStrs() {
							prt(2, ColorFlow(flowStr, env))
						}
					}
					if len(cic.CmdLine()) != 0 {
						if cic.Type() == core.CmdTypeEmptyDir {
							prt(1, ColorProp("- dir:", env))
						} else if cic.Type() == core.CmdTypeFileNFlow {
							prt(1, ColorProp("- executable(after flow):", env))
						} else {
							prt(1, ColorProp("- executable:", env))
						}
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

	allShown = true
	if args.Recursive && (args.MaxDepth == 0 || depth < args.MaxDepth) {
		for _, name := range cmd.SubNames() {
			subShown := dumpCmd(screen, env, cmd.GetSub(name), args, indentAdjust, depth+1)
			allShown = allShown && subShown
		}
	} else {
		allShown = !cmd.HasSub()
	}
	return allShown
}

func dumpIsAutoTimerKey(env *core.Env, cmd *core.Cmd, key string) string {
	keys := cmd.GetAutoTimerKeys()
	if key == keys.Begin {
		return ColorSymbol(" <- ", env) + ColorExplain("(when running begins)", env)
	} else if key == keys.End {
		return ColorSymbol(" <- ", env) + ColorExplain("(when running ends)", env)
	} else if key == keys.Dur {
		return ColorSymbol(" <- ", env) + ColorExplain("(running elapsed secs)", env)
	}
	return ""
}

func cmdIdStr(cmd *core.CmdTree, name string, env *core.Env) string {
	frameColor := ColorCmd
	if !cmd.HasSub() {
		frameColor = ColorCmdEmpty
	}
	cmdColor := ColorCmd
	if cmd.Cmd() == nil || cmd.Cmd().IsNoExecutableCmd() {
		cmdColor = ColorCmdEmpty
	}
	return frameColor("[", env) +
		cmdColor(name, env) +
		frameColor("]", env)
}
