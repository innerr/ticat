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
		if args == nil {
			continue
		}
		argv := cmd.GenEnv(env).GetArgv(cmd.Path(), sep, args)
		for j, line := range args.Dump(argv, true) {
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

func DumpMods(cc *Cli, printAlias bool) {
	dumpMod(cc.Screen, cc.Cmds, printAlias, 0)
}

func dumpMod(screen *Screen, mod *CmdTree, printAlias bool, indent int) {
	if mod == nil {
		return
	}
	indentPrint := func(msg string) {
		screen.Println(strings.Repeat(" ", indent*4) + msg)
	}
	indentPrint("[" + mod.DisplayName() + "]")
	if mod.parent != nil {
		indentPrint("- parent: " + mod.parent.DisplayName())
	}
	if mod.cmd != nil {
		line := "- cmd-type: " + string(mod.cmd.ty)
		if mod.cmd.quiet {
			line += " (quiet)"
		}
		indentPrint(line)
		if mod.cmd.ty == CmdTypeBash {
			indentPrint("  - executable: " + mod.cmd.bash)
		}
		args := mod.cmd.args
		for i, name := range args.list {
			val := args.defVals[name]
			indentPrint("  - arg:" + strconv.Itoa(i) + " " + name + " = '" + val + "'")
		}
		if printAlias {
			for k, v := range args.abbrsRevIdx {
				if k != v {
					indentPrint("  - arg-alias: " + k + " = " + v)
				}
			}
		}
	}
	if printAlias {
		for k, v := range mod.subAbbrsRevIdx {
			if k != v {
				indentPrint("- alias: " + k + " = " + v)
			}
		}
	}
	for _, it := range mod.sub {
		dumpMod(screen, it, printAlias, indent+1)
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
	if env.tp == EnvLayerDefault && !printDefEnv {
		return
	}
	var output []string
	indent := strings.Repeat(" ", depth*4)
	var keys []string
	for k, _ := range env.pairs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := env.pairs[k]
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
	if env.parent != nil {
		dumpEnvLayer(env.parent, printEnvLayer, printDefEnv, filterPrefixs, &output, depth+1)
	}
	if len(output) != 0 {
		*res = append(*res, indent+"["+EnvLayerName(env.tp)+"]")
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
