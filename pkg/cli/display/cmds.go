package display

import (
	"fmt"
	"strings"

	"github.com/innerr/ticat/pkg/core/model"
)

func DumpCmdsWithTips(
	cmds *model.CmdTree,
	screen model.Screen,
	env *model.Env,
	args *DumpCmdArgs,
	displayCmdPath string,
	lessDetailCmd string,
	moreDetailCmd string) (allShown bool, err error) {

	if !args.Recursive {
		return false, fmt.Errorf("should never happen, this func is not for non-recursive dumping display")
	}

	prt := func(text ...interface{}) {
		PrintTipTitle(screen, env, text...)
	}

	buf := NewCacheScreen()
	stackDepth := env.GetInt("sys.stack-depth")
	allShown = dumpCmd(buf, env, cmds, args, -cmds.Depth(), 0, stackDepth)

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
		if buf.OutputtedLines() <= 6 && args.Skeleton && len(moreDetailCmd) > 0 {
			footerLines = append(footerLines, "", "use '"+moreDetailCmd+"' to show more details")
		}
	} else {
		footerLines = append(footerLines, "", fmt.Sprintf("some may not shown by arg depth='%d'", args.MaxDepth))
	}

	title := ""
	if buf.OutputtedLines() > 0 {
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
			return false, fmt.Errorf("should never happen, no loaded commands")
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
	return allShown, nil
}

func DumpCmds(
	cmds *model.CmdTree,
	screen model.Screen,
	env *model.Env,
	args *DumpCmdArgs) (allShown bool) {

	return dumpCmd(screen, env, cmds, args, -cmds.Depth(), 0, env.GetInt("sys.stack-depth"))
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

func (self *DumpCmdArgs) MatchFind(cmd *model.CmdTree) bool {
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
	screen model.Screen,
	env *model.Env,
	cmd *model.CmdTree,
	args *DumpCmdArgs,
	indentAdjust int,
	depth int,
	stackDepth int) (allShown bool) {

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

			// TODO: move 'help' from model.Cmd to model.CmdTree
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
				abbrs := displayCmdAbbrsPath(cmd, env)
				if len(abbrs) != 0 && abbrs != full {
					prt(1, ColorProp("- full-abbrs:", env))
					prt(2, abbrs)
				}
			}
		}

		if (!args.Skeleton || args.ShowUsage) && cic != nil {
			cicArgs := cic.Args()
			autoMapInfo := cic.GetArgsAutoMapStatus()
			argNames := cicArgs.Names()
			if len(argNames) != 0 {
				prt(1, ColorProp("- args:", env))
			}
			for _, name := range argNames {
				val := cicArgs.DefVal(name, stackDepth)
				var nameList []string
				for i, it := range cicArgs.Abbrs(name) {
					if i == 0 {
						it = ColorArg(it, env)
					} else {
						it = ColorExplain(it, env)
					}
					nameList = append(nameList, it)
				}
				nameStr := strings.Join(nameList, ColorAbbrSep(abbrsSep, env))
				val = mayMaskSensitiveVal(env, nameStr, val)
				line := nameStr + ColorSymbol(" = ", env) + mayQuoteStr(val)
				enums := cicArgs.EnumVals(name)
				if len(enums) != 0 {
					line += " " + ColorExplain("(enum: "+strings.Join(enums, cmd.Strs.ArgEnumSep)+")", env)
				}
				if !args.Skeleton {
					entry := autoMapInfo.GetMappedSource(name)
					if entry != nil {
						line += ColorExplain(" <- ", env) + ColorCmdLowKey("["+entry.SrcCmd.Owner().DisplayPath()+"]", env)
					}
				}
				prt(2, line)
			}
		}

		if !args.Skeleton && cic != nil {
			val2env := cic.GetVal2Env()
			if len(val2env.EnvKeys()) != 0 {
				prt(1, ColorProp("- env-direct-write:", env))
			}
			for _, k := range val2env.EnvKeys() {
				val := val2env.Val(k)
				val = mayMaskSensitiveVal(env, k, val)
				prt(2, ColorKey(k, env)+ColorSymbol(" = ", env)+mayQuoteStr(val))
			}

			arg2env := cic.GetArg2Env()
			if len(arg2env.EnvKeys()) != 0 {
				prt(1, ColorProp("- env-from-argv:", env))
			}
			for _, k := range arg2env.EnvKeys() {
				argName := arg2env.GetArgName(cic, k, true)
				prt(2, ColorKey(k, env)+ColorSymbol(" <- ", env)+
					ColorArg(mayQuoteStr(argName), env))
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
			//if !cic.HasSubFlow(false) && (cic.Type() != model.CmdTypeNormal || cic.IsQuiet()) {
			if cic.Type() != model.CmdTypeFlow || cic.Type() != model.CmdTypeAdHotFlow {
				line := string(cic.Type())
				if cic.IsQuiet() {
					line += " (quiet)"
				}
				if cic.IsPriority() {
					line += " (priority)"
				}
				if cic.AllowTailModeCall() {
					line += " (tail-mode)"
				}
				prt(1, ColorProp("- cmd-type:", env))
				prt(2, line)
			}

			if len(cmd.Source()) == 0 || !strings.HasPrefix(cic.CmdLine(), cmd.Source()) {
				prt(1, ColorProp("- from:", env))
				if len(cmd.Source()) == 0 {
					prt(2, builtinName)
				} else {
					prt(2, cmd.Source())
				}
			}

			// TODO: a bit messy
			if cic.Type() != model.CmdTypeNormal && cic.Type() != model.CmdTypePower {
				if len(cic.CmdLine()) != 0 || len(cic.FlowStrs()) != 0 {
					if cic.HasSubFlow(false) {
						prt(1, ColorProp("- flow:", env))
						for _, flowStr := range cic.FlowStrs() {
							prt(2, ColorFlow(flowStr, env))
						}
					}
					if len(cic.CmdLine()) != 0 {
						if cic.Type() == model.CmdTypeEmptyDir {
							prt(1, ColorProp("- dir:", env))
						} else if cic.Type() == model.CmdTypeFileNFlow {
							prt(1, ColorProp("- executable(after flow):", env))
						} else {
							prt(1, ColorProp("- executable:", env))
						}
						prt(2, cic.CmdLine())
					}
				}
			}

			if cic.Type() != model.CmdTypeNormal && cic.Type() != model.CmdTypePower {
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
			subShown := dumpCmd(screen, env, cmd.GetSub(name), args, indentAdjust, depth+1, stackDepth)
			allShown = allShown && subShown
		}
	} else {
		allShown = !cmd.HasSubs()
	}
	return allShown
}

func dumpIsAutoTimerKey(env *model.Env, cmd *model.Cmd, key string) string {
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

func cmdIdStr(cmd *model.CmdTree, name string, env *model.Env) string {
	frameColor := ColorCmd
	if !cmd.HasSubs() {
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

func getCmdAbbrsPath(cmd *model.CmdTree, env *model.Env) []string {
	if cmd.Parent() == nil {
		return nil
	}
	abbrs := cmd.Parent().SubAbbrs(cmd.Name())
	if len(abbrs) == 0 {
		return nil
	}
	sep := cmd.Strs.AbbrsSep
	//sep := ColorExplain(cmd.Strs.AbbrsSep, env)
	return append(getCmdAbbrsPath(cmd.Parent(), env), strings.Join(abbrs, sep))
}

func displayCmdAbbrsPath(cmd *model.CmdTree, env *model.Env) string {
	path := getCmdAbbrsPath(cmd, env)
	if len(path) == 0 {
		return ""
	} else {
		return strings.Join(path, cmd.Strs.PathSep)
	}
}
