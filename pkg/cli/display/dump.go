package display

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

// TODO: clean code

func DumpFlow(cc *core.Cli, env *core.Env, flow []core.ParsedCmd, sep string, indentSize int) {
	dumpFlow(cc, env, flow, sep, indentSize, 0)
}

func DumpEnv(screen core.Screen, env *core.Env, indentSize int) {
	lines := dumpEnv(env, true, true, true, true, nil, indentSize)
	for _, line := range lines {
		screen.Print(line + "\n")
	}
}

func DumpCmds(
	cc *core.Cli,
	onlyNames bool,
	indentSize int,
	flatten bool,
	path string,
	findStrs ...string) {

	cmds := cc.Cmds
	if len(path) != 0 {
		cmds = cmds.GetSub(strings.Split(path, cc.Cmds.Strs.PathSep)...)
		if cmds == nil {
			panic(fmt.Errorf("[DumpCmds] can't find sub cmd tree by path '%s'", path))
		}
	}
	dumpCmd(cc.Screen, cmds, onlyNames, indentSize, true, flatten, -cmds.Depth(), findStrs...)
}

// TODO: dump more info, eg: full path
func DumpEnvAbbrs(cc *core.Cli, indentSize int) {
	dumpEnvAbbrs(cc.Screen, cc.EnvAbbrs, cc.Cmds.Strs.AbbrsSep, indentSize, 0)
}

func DumpEnvFlattenVals(screen core.Screen, env *core.Env, findStrs ...string) {
	flatten := env.Flatten(true, nil, true)
	var keys []string
	for k, _ := range flatten {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := flatten[k]
		if len(findStrs) != 0 {
			notMatched := false
			for _, findStr := range findStrs {
				if strings.Index(k, findStr) < 0 &&
					strings.Index(v, findStr) < 0 {
					notMatched = true
					break
				}
			}
			if notMatched {
				continue
			}
		}
		screen.Print(k + " = " + MayQuoteStr(v) + "\n")
	}
}

func DumpArgs(args *core.Args, argv core.ArgVals, printDef bool) (output []string) {
	for _, k := range args.Names() {
		defV := args.DefVal(k)
		line := k + " = "
		if argv != nil {
			v := argv[k].Raw
			line += MayQuoteStr(v)
			if printDef {
				if defV != v {
					line += "(def=" + MayQuoteStr(defV) + ")"
				} else {
					line += "(=def)"
				}
			}
		} else {
			line += MayQuoteStr(defV)
		}
		output = append(output, line)
	}
	return
}

func DumpEnvOpsCheckResult(screen core.Screen, result []core.EnvOpsCheckResult, sep string) {
	if len(result) == 0 {
		return
	}
	screen.Print("--------<unsatisfied env read>--------\n")

	fatals := newEnvOpsCheckResultAgg()
	risks := newEnvOpsCheckResultAgg()
	for _, it := range result {
		if it.ReadNotExist {
			fatals.Append(it)
		} else {
			risks.Append(it)
		}
	}

	prt0 := func(msg string) {
		screen.Print(msg + "\n")
	}
	prti := func(msg string, indent int) {
		screen.Print(strings.Repeat(" ", indent) + msg + "\n")
	}

	if len(risks.result) != 0 && len(fatals.result) == 0 {
		for _, it := range risks.result {
			screen.Print("\n")
			prt0("<RISK>  '" + it.Key + "'")
			if it.MayReadNotExist || it.MayReadMayWrite {
				prti("- may-read by:", 7)
			} else if it.ReadMayWrite {
				prti("- read by:", 7)
			}
			for _, cmd := range it.Cmds {
				prti("["+cmd+"]", 12)
			}
			if len(it.MayWriteCmdsBefore) != 0 && (it.ReadMayWrite || it.MayReadMayWrite) {
				prti("- but may not provided by:", 7)
				for _, cmd := range it.MayWriteCmdsBefore {
					prti("["+cmd.Matched.DisplayPath(sep, true)+"]", 12)
				}
			} else {
				if it.MayReadNotExist {
					prti("- but not provided", 7)
				} else {
					prti("- but may not provided", 7)
				}
			}
		}
	}

	if len(fatals.result) != 0 {
		for _, it := range fatals.result {
			screen.Print("\n")
			prt0("<FATAL> '" + it.Key + "'")
			prti("- read by:", 7)
			for _, cmd := range it.Cmds {
				prti("["+cmd+"]", 12)
			}
			prti("- but not provided", 7)
		}

		// User should know how to search commands, so no need to display the hint
		// screen.Print("\n<HINT>   to find key provider:\n")
		// prti("- ticat cmds.ls <key> write <other-find-str> <more-find-str> ..", 7)
	}
}

func dumpEnvAbbrs(
	screen core.Screen,
	abbrs *core.EnvAbbrs,
	abbrsSep string,
	indentSize int,
	indent int) {

	if abbrs == nil {
		return
	}
	prt := func(msg string) {
		if indent >= 0 {
			screen.Print(rpt(" ", indentSize*indent) + msg + "\n")
		}
	}

	name := strings.Join(abbrs.Abbrs(), abbrsSep)
	prt("[" + name + "]")

	for _, name := range abbrs.SubNames() {
		dumpEnvAbbrs(screen, abbrs.GetSub(name), abbrsSep, indentSize, indent+1)
	}
}

func dumpCmd(
	screen core.Screen,
	cmd *core.CmdTree,
	onlyNames bool,
	indentSize int,
	recursive bool,
	flatten bool,
	indentAdjust int,
	findStrs ...string) {

	if cmd == nil {
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
		name := strings.Join(cmd.Abbrs(), abbrsSep)
		if len(name) == 0 {
			name = cmd.DisplayName()
		}

		if !flatten || cic != nil {
			prt(0, "["+name+"]")
			if !onlyNames && cic != nil && len(cic.Help()) != 0 {
				prt(1, " '"+cic.Help()+"'")
			}
			if !onlyNames && cmd.Parent() != nil && cmd.Parent().Parent() != nil {
				prt(1, "- full-cmd:")
				full := cmd.DisplayPath()
				prt(2, full)
				abbrs := cmd.DisplayAbbrsPath()
				if len(abbrs) != 0 && abbrs != full {
					prt(1, "- full-abbrs:")
					prt(2, abbrs)
				}
			}
		}

		if !onlyNames && cic != nil {
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
			if len(cic.CmdLine()) != 0 && (cic.Type() == core.CmdTypeFile ||
				cic.Type() == core.CmdTypeDir || cic.Type() == core.CmdTypeFlow) {
				if cic.Type() == core.CmdTypeFlow {
					prt(1, "- flow:")
				} else {
					prt(1, "- executable:")
				}
				prt(2, cic.CmdLine())
			}
			if len(cic.Source()) != 0 && !strings.HasPrefix(cic.CmdLine(), cic.Source()) {
				prt(1, "- source:")
				if len(cic.Source()) == 0 {
					prt(2, "builtin")
				} else {
					prt(2, cic.Source())
				}
			}
			envOps := cic.EnvOps()
			envOpKeys := envOps.EnvKeys()
			if len(envOpKeys) != 0 {
				prt(1, "- env-ops:")
			}
			for _, k := range envOpKeys {
				prt(2, k+" = "+dumpEnvOps(envOps.Ops(k), envOpSep))
			}
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
		}
	}

	if recursive {
		for _, name := range cmd.SubNames() {
			dumpCmd(screen, cmd.GetSub(name), onlyNames, indentSize,
				recursive, flatten, indentAdjust, findStrs...)
		}
	}
}

func dumpFlow(cc *core.Cli, env *core.Env, flow []core.ParsedCmd, sep string, indentSize int, indentAdjust int) {
	for _, cmd := range flow {
		if len(cmd) != 0 {
			dumpFlowCmd(cc, env, cmd, sep, indentSize, indentAdjust)
		}
	}
}

func dumpFlowCmd(
	cc *core.Cli,
	env *core.Env,
	parsedCmd core.ParsedCmd,
	sep string,
	indentSize int,
	indentAdjust int) {

	cmd := parsedCmd[len(parsedCmd)-1].Cmd.Cmd
	if cmd == nil {
		return
	}

	envOpSep := " " + cmd.Strs.EnvOpSep + " "

	prt := func(indentLvl int, msg string) {
		indentLvl += indentAdjust
		padding := rpt(" ", indentSize*indentLvl)
		msg = autoPadNewLine(padding, msg)
		cc.Screen.Print(padding + msg + "\n")
	}

	cic := cmd.Cmd()
	if cic == nil {
		return
	}
	name := parsedCmd.DisplayPath(sep, true)
	prt(0, "["+name+"]")
	if cic != nil && len(cic.Help()) != 0 {
		prt(1, " '"+cic.Help()+"'")
	}

	cmdEnv := parsedCmd.GenEnv(env, cc.Cmds.Strs.EnvValDelMark, cc.Cmds.Strs.EnvValDelAllMark)
	args := parsedCmd.Args()
	argv := cmdEnv.GetArgv(parsedCmd.Path(), sep, args)
	argLines := DumpArgs(&args, argv, true)
	if len(argLines) != 0 {
		prt(1, "- args:")
	}
	for _, line := range argLines {
		prt(2, line)
	}

	// TODO: missed kvs in GlobalEnv
	cmdEnv = parsedCmd.GenEnv(core.NewEnv(), cc.Cmds.Strs.EnvValDelMark, cc.Cmds.Strs.EnvValDelAllMark)
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

	envOps := cic.EnvOps()
	envOpKeys := envOps.EnvKeys()
	if len(envOpKeys) != 0 {
		prt(1, "- env-ops:")
	}
	for _, k := range envOpKeys {
		prt(2, k+" = "+dumpEnvOps(envOps.Ops(k), envOpSep))
	}

	line := string(cic.Type())
	if cic.IsQuiet() {
		line += " (quiet)"
	}
	if cic.IsPriority() {
		line += " (priority)"
	}
	if cic.Type() != core.CmdTypeNormal || cic.IsQuiet() {
		if cic.Type() != core.CmdTypeFile && cic.Type() != core.CmdTypeDir &&
			cic.Type() != core.CmdTypeFlow {
			prt(1, "- cmd-type:")
			prt(2, line)
		}
	}

	if len(cic.Source()) != 0 && !strings.HasPrefix(cic.CmdLine(), cic.Source()) {
		prt(1, "- source:")
		prt(2, cic.Source())
	}

	if len(cic.CmdLine()) != 0 && (cic.Type() == core.CmdTypeFile ||
		cic.Type() == core.CmdTypeDir || cic.Type() == core.CmdTypeFlow) {
		if cic.Type() == core.CmdTypeFlow {
			prt(1, "- flow:")
		} else {
			prt(1, "- executable:")
		}
		prt(2, cic.CmdLine())
		if cic.Type() == core.CmdTypeFlow {
			prt(2, "--->>>")
			subFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, cic.Flow()...)
			dumpFlow(cc, env, subFlow.Cmds, sep, indentSize, indentAdjust+2)
			prt(2, "<<<---")
		}
	}
}

func dumpEnv(
	env *core.Env,
	printEnvLayer bool,
	printDefEnv bool,
	printRuntimeEnv bool,
	printEnvStrs bool,
	filterPrefixs []string,
	indentSize int) (res []string) {

	sep := env.Get("strs.env-path-sep").Raw
	if !printRuntimeEnv {
		sysPrefix := env.Get("strs.env-sys-path").Raw + sep
		filterPrefixs = append(filterPrefixs, sysPrefix)
	}
	if !printEnvStrs {
		strsPrefix := env.Get("strs.env-strs-path").Raw + sep
		filterPrefixs = append(filterPrefixs, strsPrefix)
	}

	if !printEnvLayer {
		flatten := env.Flatten(printDefEnv, filterPrefixs, true)
		var keys []string
		for k, _ := range flatten {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			res = append(res, k+" = "+flatten[k])
		}
	} else {
		dumpEnvLayer(env, printEnvLayer, printDefEnv, filterPrefixs, &res, indentSize, 0)
	}
	return
}

func dumpEnvLayer(
	env *core.Env,
	printEnvLayer bool,
	printDefEnv bool,
	filterPrefixs []string,
	res *[]string,
	indentSize int,
	depth int) {

	if env.LayerType() == core.EnvLayerDefault && !printDefEnv {
		return
	}
	var output []string
	indent := rpt(" ", depth*indentSize)
	keys, _ := env.Pairs()
	sort.Strings(keys)
	for _, k := range keys {
		v := env.Get(k)
		filtered := false
		for _, filterPrefix := range filterPrefixs {
			if len(filterPrefix) != 0 && strings.HasPrefix(k, filterPrefix) {
				filtered = true
				break
			}
		}
		if !filtered {
			output = append(output, indent+"- "+k+" = "+MayQuoteStr(v.Raw))
		}
	}
	if env.Parent() != nil {
		dumpEnvLayer(env.Parent(), printEnvLayer, printDefEnv, filterPrefixs, &output, indentSize, depth+1)
	}
	if len(output) != 0 {
		*res = append(*res, indent+"["+env.LayerTypeName()+"]")
		*res = append(*res, output...)
	}
}

func dumpEnvOps(ops []uint, sep string) (str string) {
	var strs []string
	for _, op := range ops {
		strs = append(strs, core.EnvOpStr(op))
	}
	return strings.Join(strs, sep)
}

// TODO: unused
func envDisplayKey(key string, val *core.ParsedEnvVal, sep string) string {
	fields := strings.Split(key, sep)
	if len(fields) != len(val.MatchedPath) {
		if len(val.MatchedPath) == 1 {
			return key
		}
		panic(fmt.Errorf("[envDisplayKey] internal error, key: '%v', matched: '%v'",
			key, val.MatchedPath))
	}
	var strs []string
	for i, field := range fields {
		abbr := val.MatchedPath[i]
		if abbr != field {
			field = abbr + "(=" + field + ")"
		}
		strs = append(strs, field)
	}
	return strings.Join(strs, sep)
}

func MayQuoteStr(origin string) string {
	trimed := strings.TrimSpace(origin)
	if len(trimed) == 0 || len(trimed) != len(origin) {
		return "'" + origin + "'"
	}
	fields := strings.Fields(origin)
	if len(fields) != 1 {
		return "'" + origin + "'"
	}
	return origin
}

func autoPadNewLine(padding string, msg string) string {
	msgNoPad := strings.TrimLeft(msg, "\t '\"")
	hiddenPad := rpt(" ", len(msg)-len(msgNoPad))
	msg = strings.ReplaceAll(msg, "\n", "\n"+padding+hiddenPad)
	return msg
}

type envOpsCheckResult struct {
	Cmds               []string
	Key                string
	MayWriteCmdsBefore []core.MayWriteCmd
	ReadMayWrite       bool
	MayReadMayWrite    bool
	MayReadNotExist    bool
	ReadNotExist       bool
	CmdMap             map[string]bool
}

type envOpsCheckResultAgg struct {
	result []envOpsCheckResult
	revIdx map[string]int
}

func newEnvOpsCheckResultAgg() *envOpsCheckResultAgg {
	return &envOpsCheckResultAgg{nil, map[string]int{}}
}

func (self *envOpsCheckResultAgg) Append(res core.EnvOpsCheckResult) {
	hashKey := fmt.Sprintf("%s_%v_%v_%v_%v", res.Key, res.ReadMayWrite,
		res.MayReadMayWrite, res.MayReadNotExist, res.ReadNotExist)
	idx, ok := self.revIdx[hashKey]
	if !ok {
		idx = len(self.result)
		self.result = append(self.result, envOpsCheckResult{
			[]string{res.CmdDisplayPath},
			res.Key,
			res.MayWriteCmdsBefore,
			res.ReadMayWrite,
			res.MayReadMayWrite,
			res.MayReadNotExist,
			res.ReadNotExist,
			map[string]bool{res.CmdDisplayPath: true},
		})
		self.revIdx[hashKey] = idx
	} else {
		old := self.result[idx]
		if !old.CmdMap[res.CmdDisplayPath] {
			old.Cmds = append(old.Cmds, res.CmdDisplayPath)
			old.CmdMap[res.CmdDisplayPath] = true
			// Discard res.MayWriteCmdsBefore, it's not important
			self.result[idx] = old
		}
	}
}
