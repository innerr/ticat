package mod_meta

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pingcap/ticat/pkg/core/model"
	"github.com/pingcap/ticat/pkg/mods/persist/meta_file"
)

func RegMod(
	cc *model.Cli,
	// Meta file full path
	metaPath string,
	// Related executable file full path
	executablePath string,
	isDir bool,
	// Has flow ext
	isFlow bool,
	// Cmd path defined by meta file base name
	cmdPath []string,
	flowExt string,
	abbrsSep string,
	envPathSep string,
	source string,
	panicRecover bool) {

	defer func() {
		if !panicRecover {
			return
		}
		if err := recover(); err != nil {
			cc.TolerableErrs.OnErr(err, source, metaPath, "module loading failed")
		}
	}()

	metas := meta_file.NewMetaFile(metaPath)
	if len(metas) == 0 {
		panic(fmt.Errorf("[RegMod] should never happen: NewMetaFile return no meta"))
	}

	if len(metas) == 1 && metas[0].NotVirtual {
		regModFile(metas[0].Meta, cc, metaPath, executablePath, isDir, isFlow, cmdPath,
			abbrsSep, envPathSep, source, panicRecover)
		return
	}

	if len(executablePath) != 0 || isDir || !isFlow {
		panic(fmt.Errorf("[RegMod] should never happen: a combined flow file not meet the expection"))
	}

	// Discard the cmdPath from real file name
	for _, meta := range metas {
		path := getVirtualFileCmdPath(meta.VirtualPath, flowExt, cc.Cmds.Strs.PathSep)
		regModFile(meta.Meta, cc, metaPath, "", false, true, path,
			abbrsSep, envPathSep, source, panicRecover)
	}
}

func getVirtualFileCmdPath(path string, flowExt string, pathSep string) []string {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, flowExt) {
		panic(fmt.Errorf("virtual flow file '%s' ext not match '%s'", path, flowExt))
	}
	raw := base[:len(base)-len(flowExt)]
	return strings.Split(raw, pathSep)
}

func regModFile(
	meta *meta_file.MetaFile,
	cc *model.Cli,
	metaPath string,
	executablePath string,
	isDir bool,
	isFlow bool,
	cmdPath []string,
	abbrsSep string,
	envPathSep string,
	source string,
	panicRecover bool) {

	mod := cc.Cmds.GetOrAddSubEx(source, cmdPath...)
	cmd := regMod(meta, mod, executablePath, isDir, source)
	mod.SetSource(source)
	cmd.SetMetaFile(metaPath)

	// Reg by isFlow, not 'cmd.Type()'
	if isFlow {
		regFlowAbbrs(meta, cc.Cmds, cmdPath)
	} else {
		regModAbbrs(meta, mod)
	}

	regMacro(meta, cmd)
	regTrivial(meta, mod)
	regUnLog(meta, cmd)
	regQuietError(meta, cmd)
	regNoSession(meta, cmd)
	regQuietCmd(meta, cmd)
	regHideInSessionsLast(meta, cmd)
	regQuietSubFlow(meta, cmd)
	regUnbreakFileNFlow(meta, cmd)

	regAutoTimer(meta, cmd)
	regTags(meta, mod)

	// AutoMap must after regular args are registered
	regArgs(meta, cmd, abbrsSep)
	regArg2EnvAutoMap(cc, meta, cmd)

	regDeps(meta, cmd)
	regEnvOps(cc.EnvAbbrs, meta, cmd, abbrsSep, envPathSep)
	regVal2Env(cc.EnvAbbrs, meta, cmd, abbrsSep, envPathSep)
	regArg2Env(cc.EnvAbbrs, meta, cmd, abbrsSep, envPathSep)
}

func regAutoTimer(meta *meta_file.MetaFile, cmd *model.Cmd) {
	key := meta.Get("begin-ts-key")
	if len(key) != 0 {
		cmd.RegAutoTimerBeginKey(key)
	} else {
		key := meta.Get("begin-key")
		if len(key) != 0 {
			cmd.RegAutoTimerBeginKey(key)
		}
	}

	key = meta.Get("end-ts-key")
	if len(key) != 0 {
		cmd.RegAutoTimerEndKey(key)
	} else {
		key := meta.Get("end-key")
		if len(key) != 0 {
			cmd.RegAutoTimerEndKey(key)
		}
	}

	key = meta.Get("duration-key")
	if len(key) != 0 {
		cmd.RegAutoTimerDurKey(key)
	} else {
		key = meta.Get("dur-key")
		if len(key) != 0 {
			cmd.RegAutoTimerDurKey(key)
		}
	}
}

func regMacro(meta *meta_file.MetaFile, cmd *model.Cmd) {
	candidates := meta.KeysWithPrefix("[[")
	origins := map[string]string{}
	var macros []string
	for _, origin := range candidates {
		candidate := origin[2:]
		if !strings.HasSuffix(candidate, "]]") {
			continue
		}
		candidate = candidate[:len(candidate)-2]
		if len(candidate) == 0 {
			continue
		}
		macro := strings.TrimSpace(candidate)
		macros = append(macros, macro)
		origins[macro] = origin
	}

	globalSection := meta.GetGlobalSection()
	for _, macro := range macros {
		macroContent := globalSection.GetMultiLineVal(origins[macro], false)
		if len(macroContent) != 0 {
			cmd.AddMacro(macro, macroContent)
		}
	}
}

func regTrivial(meta *meta_file.MetaFile, mod *model.CmdTree) {
	val := meta.Get("trivial")
	if len(val) == 0 {
		return
	}
	trivial, err := strconv.Atoi(val)
	if err != nil {
		panic(fmt.Errorf("[regTrivial] trivial string '%s' is not int: '%v'", val, err))
	}
	mod.SetTrivial(trivial)
}

func regUnLog(meta *meta_file.MetaFile, cmd *model.Cmd) {
	for _, key := range []string{"nolog", "unlog", "interact", "interactive"} {
		val := meta.Get(key)
		if len(val) == 0 {
			continue
		}
		unLog, err := strconv.ParseBool(val)
		if err != nil {
			panic(fmt.Errorf("[regUnLog] unlog value string '%s' is not bool: '%v'", val, err))
		}
		if unLog {
			cmd.SetUnLog()
		}
		return
	}
}

func regQuietError(meta *meta_file.MetaFile, cmd *model.Cmd) {
	for _, key := range []string{"quiet-error", "quiet-err", "silent-error", "silent-err"} {
		val := meta.Get(key)
		if len(val) == 0 {
			continue
		}
		quiet, err := strconv.ParseBool(val)
		if err != nil {
			panic(fmt.Errorf("[regQuietError] quiet-error value string '%s' is not bool: '%v'", val, err))
		}
		if quiet {
			cmd.SetQuietError()
		}
		return
	}
}

func regNoSession(meta *meta_file.MetaFile, cmd *model.Cmd) {
	for _, key := range []string{"no-session"} {
		val := meta.Get(key)
		if len(val) == 0 {
			continue
		}
		no, err := strconv.ParseBool(val)
		if err != nil {
			panic(fmt.Errorf("[regNoSession] no-session value string '%s' is not bool: '%v'", val, err))
		}
		if no {
			cmd.SetNoSession()
		}
		return
	}
}

func regQuietCmd(meta *meta_file.MetaFile, cmd *model.Cmd) {
	for _, key := range []string{"quiet"} {
		val := meta.Get(key)
		if len(val) == 0 {
			continue
		}
		quiet, err := strconv.ParseBool(val)
		if err != nil {
			panic(fmt.Errorf("[regQuietCmd] quiet value string '%s' is not bool: '%v'", val, err))
		}
		if quiet {
			cmd.SetQuiet()
		}
		return
	}
}

func regHideInSessionsLast(meta *meta_file.MetaFile, cmd *model.Cmd) {
	for _, key := range []string{"hide-in-sessions-last", "hide-in-session-last", "hide-in-last"} {
		val := meta.Get(key)
		if len(val) == 0 {
			continue
		}
		hide, err := strconv.ParseBool(val)
		if err != nil {
			panic(fmt.Errorf("[regHideInSessionsLast] hide-in-last value string '%s' is not bool: '%v'", val, err))
		}
		if hide {
			cmd.SetHideInSessionsLast()
		}
		return
	}
}

func regQuietSubFlow(meta *meta_file.MetaFile, cmd *model.Cmd) {
	for _, key := range []string{"pack-subflow", "pack-sub"} {
		val := meta.Get(key)
		if len(val) == 0 {
			continue
		}
		quiet, err := strconv.ParseBool(val)
		if err != nil {
			panic(fmt.Errorf("[regQuietSubFlow] quiet-subflow value string '%s' is not bool: '%v'", val, err))
		}
		if quiet {
			cmd.SetQuietSubFlow()
		}
		return
	}
}

func regUnbreakFileNFlow(meta *meta_file.MetaFile, cmd *model.Cmd) {
	for _, key := range []string{"unbreak-file-n-flow", "unbreak-file&flow", "unbreak-file+flow", "unbreak-filenflow", "unbreak-fnf"} {
		val := meta.Get(key)
		if len(val) == 0 {
			continue
		}
		quiet, err := strconv.ParseBool(val)
		if err != nil {
			panic(fmt.Errorf("[regUnbreakFileNFlow] unbreak-file-n-flow value string '%s' is not bool: '%v'", val, err))
		}
		if quiet {
			cmd.SetUnbreakFileNFlow()
		}
		return
	}
}

func regArg2EnvAutoMap(cc *model.Cli, meta *meta_file.MetaFile, cmd *model.Cmd) {
	globalSection := meta.GetGlobalSection()
	var names []string
	for _, key := range []string{"args.auto", "arg.auto", "arg2env.auto-map", "arg2env.auto", "arg2env.map"} {
		lines := globalSection.GetMultiLineVal(key, false)
		if len(lines) == 0 {
			continue
		}
		for _, line := range lines {
			// TODO: get sep from env
			for _, name := range strings.Split(line, ",") {
				name = strings.TrimSpace(name)
				if len(name) != 0 {
					names = append(names, name)
				}
			}
		}
		break
	}
	if len(names) != 0 {
		cmd.SetArg2EnvAutoMap(names)
		cc.Arg2EnvAutoMapCmds.AddAutoMapTarget(cmd)
	}
}

func regTags(meta *meta_file.MetaFile, mod *model.CmdTree) {
	tags := meta.Get("tags")
	if len(tags) == 0 {
		tags = meta.Get("tag")
	}
	if len(tags) == 0 {
		return
	}
	mod.AddTags(strings.Fields(tags)...)
}

func regMod(
	meta *meta_file.MetaFile,
	mod *model.CmdTree,
	executablePath string,
	isDir bool,
	source string) *model.Cmd {

	cmdPath := mod.DisplayPath()

	globalSection := meta.GetGlobalSection()
	flow := globalSection.GetMultiLineVal("flow", false)

	// 'cmd' should be a relative path base on this file when 'isDir'
	cmdLine := meta.Get("cmd")

	help := meta.Get("help")
	// If has executable file, it need to have help string, a flow can have not
	// if len(help) == 0 && (!isDir && len(flow) == 0 || len(cmdLine) != 0) {
	// 	panic(fmt.Errorf("[regMod] cmd '%s' has no help string in '%s'",
	//		cmdPath, meta.Path()))
	//}

	// Even if 'isFlow' is true, if it does not have 'flow' content, it can't reg as flow
	if len(flow) != 0 && len(cmdLine) == 0 && len(executablePath) == 0 {
		return mod.RegFlowCmd(flow, help, source)
	}

	if len(executablePath) == 0 {
		return mod.RegMetaOnlyCmd(help, source)
	}

	// Adjust 'executablePath'
	if len(cmdLine) != 0 {
		if !isDir {
			executablePath = filepath.Dir(executablePath)
		}
		var err error
		executablePath, err = filepath.Abs(filepath.Join(executablePath, cmdLine))
		if err != nil {
			panic(fmt.Errorf("[regMod] cmd '%s' get abs path of '%s' failed",
				cmdPath, executablePath))
		}
		if !fileExists(executablePath) {
			panic(fmt.Errorf("[regMod] cmd '%s' point to a not existed file '%s'",
				cmdPath, executablePath))
		}
	}

	// TOOD: a bit messy

	if isDir {
		if len(cmdLine) != 0 {
			if len(flow) != 0 {
				panic(fmt.Errorf("[regMod] cmd '%s' has both command-line '%s' and flow",
					cmdPath, cmdLine))
			}
			return mod.RegDirWithCmd(executablePath, help, source)
		} else {
			if len(flow) != 0 {
				return mod.RegFlowCmd(flow, help, source)
			}
			return mod.RegEmptyDirCmd(executablePath, help)
		}
	} else if len(flow) != 0 {
		return mod.RegFileNFlowCmd(flow, executablePath, help, source)
	} else {
		return mod.RegFileCmd(executablePath, help, source)
	}
}

func regModAbbrs(meta *meta_file.MetaFile, mod *model.CmdTree) {
	abbrs := meta.Get("abbrs")
	if len(abbrs) == 0 {
		abbrs = meta.Get("abbr")
	}
	if len(abbrs) == 0 {
		return
	}
	abbrsSep := mod.Strs.AbbrsSep
	mod.AddAbbrs(strings.Split(abbrs, abbrsSep)...)
}

func regFlowAbbrs(meta *meta_file.MetaFile, cmds *model.CmdTree, cmdPath []string) {
	abbrsStr := meta.Get("abbrs")
	if len(abbrsStr) == 0 {
		abbrsStr = meta.Get("abbr")
	}
	if len(abbrsStr) == 0 {
		return
	}

	pathSep := cmds.Strs.PathSep
	abbrsSep := cmds.Strs.AbbrsSep

	var abbrs [][]string
	for _, abbrSeg := range strings.Split(abbrsStr, pathSep) {
		abbrList := strings.Split(abbrSeg, abbrsSep)
		abbrs = append(abbrs, abbrList)
	}

	mod := cmds
	for i, cmd := range cmdPath {
		mod = mod.GetOrAddSub(cmd)
		if i < len(abbrs) {
			mod.AddAbbrs(abbrs[i]...)
		}
	}
}

func regArgs(meta *meta_file.MetaFile, cmd *model.Cmd, abbrsSep string) {
	args := meta.GetSection("args")
	if args == nil {
		args = meta.GetSection("arg")
	}
	if args == nil {
		return
	}
	enumSep := cmd.Owner().Strs.ArgEnumSep
	for _, names := range args.Keys() {
		nameAndAbbrs := strings.Split(names, abbrsSep)
		name := strings.TrimSpace(nameAndAbbrs[0])
		var argAbbrs []string
		for _, abbr := range nameAndAbbrs[1:] {
			argAbbrs = append(argAbbrs, strings.TrimSpace(abbr))
		}
		defVal := args.Get(names)
		var enums []string
		if len(defVal) > 0 && defVal[len(defVal)-1] == ')' {
			i := strings.LastIndex(defVal, "(")
			if i >= 0 {
				enumsStr := strings.TrimSpace(defVal[i+1 : len(defVal)-1])
				if strings.HasPrefix(enumsStr, "enum:") {
					enumsStr = strings.TrimSpace(enumsStr[5:])
				}
				enums = strings.Split(enumsStr, enumSep)
				defVal = strings.TrimSpace(defVal[:i])
			}
		}
		cmd.AddArg(name, defVal, argAbbrs...)
		if len(enums) != 0 {
			cmd.SetArgEnums(name, enums...)
		}
	}
}

func regDeps(meta *meta_file.MetaFile, cmd *model.Cmd) {
	deps := meta.GetSection("deps")
	if deps == nil {
		deps = meta.GetSection("dep")
	}
	if deps != nil {
		for _, dep := range deps.Keys() {
			reason := deps.GetUnTrim(dep)
			cmd.AddDepend(dep, reason)
		}
	}
}

func regEnvOps(
	envAbbrs *model.EnvAbbrs,
	meta *meta_file.MetaFile,
	cmd *model.Cmd,
	abbrsSep string,
	envPathSep string) {

	envOps := meta.GetSection("env")
	if envOps == nil {
		return
	}

	for _, envKey := range envOps.Keys() {
		op := envOps.Get(envKey)
		key := regEnvKeyAbbrs(envAbbrs, envKey, abbrsSep, envPathSep)
		opFields := strings.Split(op, abbrsSep)
		if len(opFields) == 1 {
			opFields = strings.Split(op, ":")
		}
		for _, it := range opFields {
			regEnvOp(cmd, key, it)
		}
	}
}

func regVal2Env(
	envAbbrs *model.EnvAbbrs,
	meta *meta_file.MetaFile,
	cmd *model.Cmd,
	abbrsSep string,
	envPathSep string) {

	writes := meta.GetSection("env.write")
	if writes == nil {
		writes = meta.GetSection("val2env")
	}
	if writes == nil {
		return
	}

	for _, envKey := range writes.Keys() {
		val := writes.Get(envKey)
		key := regEnvKeyAbbrs(envAbbrs, envKey, abbrsSep, envPathSep)
		cmd.AddVal2Env(key, val)
	}
}

func regArg2Env(
	envAbbrs *model.EnvAbbrs,
	meta *meta_file.MetaFile,
	cmd *model.Cmd,
	abbrsSep string,
	envPathSep string) {

	writes := meta.GetSection("env.from-arg")
	if writes == nil {
		writes = meta.GetSection("env.arg")
	}
	if writes == nil {
		writes = meta.GetSection("arg2env")
	}
	if writes == nil {
		return
	}

	for _, envKey := range writes.Keys() {
		argName := writes.Get(envKey)
		key := regEnvKeyAbbrs(envAbbrs, envKey, abbrsSep, envPathSep)
		cmd.AddArg2Env(key, argName)
	}
}

func regEnvKeyAbbrs(
	envAbbrs *model.EnvAbbrs,
	envKeyWithAbbrs string,
	abbrsSep string,
	envPathSep string) (key string) {

	var path []string
	segs := strings.Split(envKeyWithAbbrs, envPathSep)
	for _, seg := range segs {
		var abbrs []string
		fields := strings.Split(seg, abbrsSep)
		if len(fields) == 1 {
			fields = strings.Split(seg, ":")
		}
		for _, abbr := range fields {
			abbrs = append(abbrs, strings.TrimSpace(abbr))
		}
		name := abbrs[0]
		abbrs = abbrs[1:]
		subEnvAbbrs := envAbbrs.GetOrAddSub(name)
		if len(abbrs) > 0 {
			envAbbrs.AddSubAbbrs(name, abbrs...)
		}
		envAbbrs = subEnvAbbrs
		path = append(path, name)
	}

	return strings.Join(path, envPathSep)
}

func regEnvOp(cmd *model.Cmd, key string, opOrigin string) {
	op := strings.ToLower(opOrigin)
	may := strings.Index(op, "may") >= 0 || strings.Index(op, "opt") >= 0
	write := strings.Index(op, "w") >= 0
	read := strings.Index(op, "rd") >= 0 ||
		strings.Index(op, "read") >= 0 || op == "r"
	if write && read {
		panic(fmt.Errorf("[LoadLocalMods.regEnvOp] "+
			"parse env r|w definition failed: %v", opOrigin))
	}
	if write {
		if may {
			cmd.AddEnvOp(key, model.EnvOpTypeMayWrite)
		} else {
			cmd.AddEnvOp(key, model.EnvOpTypeWrite)
		}
	}
	if read {
		if may {
			cmd.AddEnvOp(key, model.EnvOpTypeMayRead)
		} else {
			cmd.AddEnvOp(key, model.EnvOpTypeRead)
		}
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && !info.IsDir()
}
