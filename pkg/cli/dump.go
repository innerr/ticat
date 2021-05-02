package cli

import (
	"sort"
	"strconv"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func DumpCmdsEx(screen core.Screen, env *core.Env, cmds []core.ParsedCmd, sep string) {
	if len(cmds) == 0 {
		return
	}
	indentSize := 4
	screen.Print("[cmds:" + strconv.Itoa(len(cmds)) + "]\n")
	for i, cmd := range cmds {
		line := strings.Repeat(" ", indentSize*1) + "[cmd:" + strconv.Itoa(i) + "] "
		line += getCmdPath(cmd, sep, true)
		screen.Print(line + "\n")
		args := cmd.Args()
		// TODO: XX
		argv := cmd.GenEnv(env, "-", "--").GetArgv(cmd.Path(), sep, args)
		for j, line := range DumpArgs(&args, argv, true) {
			screen.Print(strings.Repeat(" ", indentSize*2) + "[arg:" + strconv.Itoa(j) + "] " + line + "\n")
		}
	}
}

func DumpCmds(cc *core.Cli, cmds []core.ParsedCmd) {
	DumpCmdsEx(cc.Screen, cc.GlobalEnv, cmds, cc.Cmds.Strs.PathSep)
}

func DumpEnv(screen core.Screen, env *core.Env) {
	lines := dumpEnv(env, true, true, true, nil)
	for _, line := range lines {
		screen.Print(line + "\n")
	}
}

func DumpMods(cc *core.Cli) {
	dumpMod(cc.Screen, cc.Cmds, -1)
}

func dumpMod(screen core.Screen, mod *core.CmdTree, indent int) {
	if mod == nil {
		return
	}
	indentPrint := func(msg string) {
		if indent >= 0 {
			screen.Print(strings.Repeat(" ", indent*4) + msg + "\n")
		}
	}
	name := strings.Join(append([]string{mod.DisplayName()}, mod.Abbrs()...), "|")
	indentPrint("[" + name + "]")
	cmd := mod.Cmd()
	if cmd != nil {
		if mod.Parent() != nil && mod.Parent().Parent() != nil {
			indentPrint("- full-path: " + mod.DisplayPath())
		}
		indentPrint("- help: " + cmd.Help())
		line := "- cmd-type: " + string(cmd.Type())
		if cmd.IsQuiet() {
			line += " (quiet)"
		}
		if cmd.Type() != core.CmdTypeNormal || cmd.IsQuiet() {
			indentPrint(line)
		}
		if cmd.Type() == core.CmdTypeBash {
			indentPrint("- executable: " + cmd.BashCmdLine())
		}
		args := cmd.Args()
		for i, name := range args.Names() {
			val := args.DefVal(name)
			names := append([]string{name}, args.Abbrs(name)...)
			indentPrint("- arg#" + strconv.Itoa(i) + " " + strings.Join(names, "|") + " = '" + val + "'")
		}
	}
	for _, name := range mod.SubNames() {
		dumpMod(screen, mod.GetSub(name), indent+1)
	}
}

func dumpEnv(env *core.Env, printEnvLayer bool, printDefEnv bool,
	printRuntimeEnv bool, filterPrefixs []string) (res []string) {

	if !printRuntimeEnv {
		sep := env.Get("strs.env-path-sep").Raw
		sysPrefix := env.Get("strs.env-sys-path").Raw + sep
		filterPrefixs = append(filterPrefixs, sysPrefix, "strs"+sep)
	}
	if !printEnvLayer {
		compacted := env.Compact(printDefEnv, filterPrefixs)
		var keys []string
		for k, _ := range compacted {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			res = append(res, k+" = "+compacted[k])
		}
	} else {
		dumpEnvLayer(env, printEnvLayer, printDefEnv, filterPrefixs, &res, 0)
	}
	return
}

func dumpEnvLayer(env *core.Env, printEnvLayer bool, printDefEnv bool, filterPrefixs []string, res *[]string, depth int) {
	if env.LayerType() == core.EnvLayerDefault && !printDefEnv {
		return
	}
	var output []string
	indent := strings.Repeat(" ", depth*4)
	keys, vals := env.Pairs()
	sort.Strings(keys)
	for i, k := range keys {
		v := vals[i]
		filtered := false
		for _, filterPrefix := range filterPrefixs {
			if len(filterPrefix) != 0 && strings.HasPrefix(k, filterPrefix) {
				filtered = true
				break
			}
		}
		if !filtered {
			output = append(output, indent+"- "+k+" = "+v.Raw)
		}
	}
	if env.Parent() != nil {
		dumpEnvLayer(env.Parent(), printEnvLayer, printDefEnv, filterPrefixs, &output, depth+1)
	}
	if len(output) != 0 {
		*res = append(*res, indent+"["+env.LayerTypeName()+"]")
		*res = append(*res, output...)
	}
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

func DumpArgs(args *core.Args, argv core.ArgVals, printDef bool) (output []string) {
	for _, k := range args.Names() {
		defV := args.DefVal(k)
		if len(defV) == 0 {
			defV = "'" + defV + "'"
		}
		line := k + " = "
		if argv != nil {
			v := argv[k].Raw
			if len(v) == 0 {
				v = "'" + v + "'"
			}
			line += v
			if printDef {
				if defV != v {
					line += " (def:" + defV + ")"
				} else {
					line += " (=def)"
				}
			}
		} else {
			line += defV + "'"
		}
		output = append(output, line)
	}
	return
}
