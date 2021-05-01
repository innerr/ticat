package cli

import (
	"sort"
	"strconv"
	"strings"
)

func DumpCmdsEx(screen *Screen, env *Env, cmds []ParsedCmd, sep string) {
	if len(cmds) == 0 {
		return
	}
	indentSize := 4
	screen.Println("[cmds:" + strconv.Itoa(len(cmds)) + "]")
	for i, cmd := range cmds {
		line := strings.Repeat(" ", indentSize*1) + "[cmd:" + strconv.Itoa(i) + "] "
		line += getCmdPath(cmd, sep, true)
		screen.Println(line)
		args := cmd.Args()
		argv := cmd.GenEnv(env).GetArgv(cmd.Path(), sep, args)
		for j, line := range DumpArgs(&args, argv, true) {
			screen.Println(strings.Repeat(" ", indentSize*2) + "[arg:" + strconv.Itoa(j) + "] " + line)
		}
	}
}

func DumpCmds(cc *Cli, cmds []ParsedCmd) {
	DumpCmdsEx(cc.Screen, cc.GlobalEnv, cmds, cc.Parser.CmdPathSep())
}

func DumpEnv(screen *Screen, env *Env) {
	lines := dumpEnv(env, true, true, true, nil)
	for _, line := range lines {
		screen.Println(line)
	}
}

func DumpMods(cc *Cli) {
	dumpMod(cc.Screen, cc.Cmds, -1)
}

func dumpMod(screen *Screen, mod *CmdTree, indent int) {
	if mod == nil {
		return
	}
	indentPrint := func(msg string) {
		if indent >= 0 {
			screen.Println(strings.Repeat(" ", indent*4) + msg)
		}
	}
	name := strings.Join(append([]string{mod.DisplayName()}, mod.Abbrs()...), "|")
	indentPrint("[" + name + "]")
	if mod.cmd != nil {
		if mod.Parent() != nil && mod.Parent().Parent() != nil {
			indentPrint("- full-path: " + mod.DisplayPath())
		}
		indentPrint("- help: " + mod.cmd.Help())
		line := "- cmd-type: " + string(mod.cmd.Type())
		if mod.cmd.IsQuiet() {
			line += " (quiet)"
		}
		if mod.cmd.Type() != CmdTypeNormal || mod.cmd.IsQuiet() {
			indentPrint(line)
		}
		if mod.cmd.Type() == CmdTypeBash {
			indentPrint("- executable: " + mod.cmd.BashCmdLine())
		}
		args := mod.cmd.Args()
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

func dumpEnv(env *Env, printEnvLayer bool, printDefEnv bool,
	printRuntimeEnv bool, filterPrefixs []string) (res []string) {

	if !printRuntimeEnv {
		filterPrefixs = append(filterPrefixs, EnvRuntimeSysPrefix)
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

func dumpEnvLayer(env *Env, printEnvLayer bool, printDefEnv bool, filterPrefixs []string, res *[]string, depth int) {
	if env.LayerType() == EnvLayerDefault && !printDefEnv {
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

func getCmdPath(cmd ParsedCmd, sep string, printRealname bool) string {
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

func DumpArgs(args *Args, argv ArgVals, printDef bool) (output []string) {
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
