package display

import (
	"sort"
	"strconv"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func DumpFlow(cc *core.Cli, env *core.Env, flow []core.ParsedCmd, sep string, indentSize int) {
	if len(flow) == 0 {
		return
	}
	cc.Screen.Print("[cmds:" + strconv.Itoa(len(flow)) + "]\n")

	abbrsSep := cc.Cmds.Strs.AbbrsSep

	for i, cmd := range flow {
		indents := strings.Repeat(" ", indentSize)
		indent2 := strings.Repeat(" ", indentSize*2)
		line := indents + "[cmd:" + strconv.Itoa(i) + "] " + getCmdPath(cmd, sep, true)
		cc.Screen.Print(line + "\n")
		cc.Screen.Print(indent2 + "'" + cmd.Help() + "'\n")

		cic := cmd.LastCmd()
		if cic != nil {
			envOps := cic.EnvOps()
			envOpKeys := envOps.EnvKeys()
			for _, k := range envOpKeys {
				cc.Screen.Print(indent2 + "- env-op: " + k + " = " +
					dumpEnvOps(envOps.OpSet(k), abbrsSep) + "\n")
			}
		}

		cmdEnv := cmd.GenEnv(env, cc.Cmds.Strs.EnvValDelMark, cc.Cmds.Strs.EnvValDelAllMark)
		args := cmd.Args()
		argv := cmdEnv.GetArgv(cmd.Path(), sep, args)
		for j, line := range DumpArgs(&args, argv, true) {
			cc.Screen.Print(indent2 + "- arg:" + strconv.Itoa(j) + " " + line + "\n")
		}
	}
}

func DumpEnv(screen core.Screen, env *core.Env, indentSize int) {
	lines := dumpEnv(env, true, true, true, true, nil, indentSize)
	for _, line := range lines {
		screen.Print(line + "\n")
	}
}

func DumpCmds(cc *core.Cli, indentSize int, flatten bool, findStr string) {
	dumpCmd(cc.Screen, cc.Cmds, indentSize, true, flatten, findStr)
}

func DumpEnvAbbrs(cc *core.Cli, indentSize int) {
	dumpEnvAbbrs(cc.Screen, cc.EnvAbbrs, cc.Cmds.Strs.AbbrsSep, indentSize, 0)
}

func DumpEnvFlattenVals(screen core.Screen, env *core.Env, findStr string) {
	flatten := env.Flatten(true, nil, true)
	var keys []string
	for k, _ := range flatten {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := flatten[k]
		if len(findStr) == 0 || strings.Index(k, findStr) >= 0 ||
			strings.Index(v, findStr) >= 0 {
			screen.Print(k + " = " + mayQuoteStr(v) + "\n")
		}
	}
}

func DumpArgs(args *core.Args, argv core.ArgVals, printDef bool) (output []string) {
	for _, k := range args.Names() {
		defV := mayQuoteStr(args.DefVal(k))
		line := k + " = "
		if argv != nil {
			v := argv[k].Raw
			line += mayQuoteStr(v)
			if printDef {
				if defV != v {
					line += " (def:" + mayQuoteStr(defV) + ")"
				} else {
					line += " (=def)"
				}
			}
		} else {
			line += mayQuoteStr(defV)
		}
		output = append(output, line)
	}
	return
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
	indentPrint := func(msg string) {
		if indent >= 0 {
			screen.Print(strings.Repeat(" ", indentSize*indent) + msg + "\n")
		}
	}

	name := strings.Join(append([]string{abbrs.DisplayName()}, abbrs.Abbrs()...), abbrsSep)
	indentPrint("[" + name + "]")

	for _, name := range abbrs.SubNames() {
		dumpEnvAbbrs(screen, abbrs.GetSub(name), abbrsSep, indentSize, indent+1)
	}
}

func dumpCmd(
	screen core.Screen,
	cmd *core.CmdTree,
	indentSize int,
	recursive bool,
	flatten bool,
	findStr string) {

	if cmd == nil {
		return
	}

	if cmd.Parent() == nil || cmd.MatchFind(findStr) {
		indent := cmd.Depth()
		indentPrint := func(msg string) {
			if flatten {
				if msg[0] == '[' {
					screen.Print(msg + "\n")
				} else {
					screen.Print(strings.Repeat(" ", indentSize) + msg + "\n")
				}
				return
			}
			if indent >= 0 {
				screen.Print(strings.Repeat(" ", indentSize*indent) + msg + "\n")
			}
		}
		abbrsSep := cmd.Strs.AbbrsSep
		cic := cmd.Cmd()
		name := strings.Join(append([]string{cmd.DisplayName()}, cmd.Abbrs()...), abbrsSep)

		if !flatten || cic != nil {
			indentPrint("[" + name + "]")
			if cmd.Parent() != nil && cmd.Parent().Parent() != nil {
				indentPrint("- full-cmd: " + cmd.DisplayPath())
			}
		}

		if cic != nil {
			indentPrint("- help: " + cic.Help())
			line := "- cmd-type: " + string(cic.Type())
			if cic.IsQuiet() {
				line += " (quiet)"
			}
			if cic.Type() != core.CmdTypeNormal || cic.IsQuiet() {
				indentPrint(line)
			}
			if cic.Type() == core.CmdTypeBash {
				indentPrint("- executable: " + cic.BashCmdLine())
			}
			envOps := cic.EnvOps()
			envOpKeys := envOps.EnvKeys()
			for _, k := range envOpKeys {
				indentPrint("- env-op: " + k + " = " + dumpEnvOps(envOps.OpSet(k), abbrsSep))
			}
			args := cic.Args()
			for i, name := range args.Names() {
				val := args.DefVal(name)
				names := append([]string{name}, args.Abbrs(name)...)
				nameStr := strings.Join(names, abbrsSep)
				indentPrint("- arg#" + strconv.Itoa(i) + " " + nameStr + " = " + mayQuoteStr(val))
			}
		}
	}

	if recursive {
		for _, name := range cmd.SubNames() {
			dumpCmd(screen, cmd.GetSub(name), indentSize, recursive, flatten, findStr)
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
	indent := strings.Repeat(" ", depth*indentSize)
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
			output = append(output, indent+"- "+k+" = "+mayQuoteStr(v.Raw))
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
		strs = append(strs, dumpEnvOp(op))
	}
	return strings.Join(strs, sep)
}

func dumpEnvOp(op uint) (str string) {
	switch op {
	case core.EnvOpTypeWrite:
		str = "write"
	case core.EnvOpTypeMayWrite:
		str = "may-write"
	case core.EnvOpTypeRead:
		str = "read"
	case core.EnvOpTypeMayRead:
		str = "may-read"
	default:
	}
	return
}

func getCmdPath(cmd core.ParsedCmd, sep string, printRealname bool) string {
	var path []string
	for _, seg := range cmd {
		if seg.Cmd.Cmd != nil {
			name := seg.Cmd.Name
			realname := seg.Cmd.Cmd.Name()
			if printRealname && name != realname {
				name += "(=" + realname + ")"
			}
			path = append(path, name)
		}
	}
	return strings.Join(path, sep)
}

func mayQuoteStr(origin string) string {
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
