package mod_meta

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func RegMod(
	cc *core.Cli,
	metaPath string,
	path string,
	isDir bool,
	cmdPath string,
	abbrsSep string,
	envPathSep string,
	source string) {

	mod := cc.Cmds.GetOrAddSub(strings.Split(cmdPath, string(filepath.Separator))...)

	meta := NewModMeta(metaPath)

	// 'cmd' should be a relative path base on this file
	cmdLine := meta.Get("cmd")
	help := meta.Get("help")
	if len(help) == 0 && (!isDir || len(cmdLine) != 0) {
		panic(fmt.Errorf("[regMod] cmd '%s' has no help string in '%s'",
			cmdPath, metaPath))
	}

	if len(cmdLine) != 0 {
		if !isDir {
			path = filepath.Dir(path)
		}
		var err error
		path, err = filepath.Abs(filepath.Join(path, cmdLine))
		if err != nil {
			panic(fmt.Errorf("[regMod] cmd '%s' get abs path of '%s' failed",
				cmdPath, path))
		}
		if !fileExists(path) {
			panic(fmt.Errorf("[regMod] cmd '%s' point to a not existed file '%s'",
				cmdPath, path))
		}
	} else if isDir {
		path = ""
	}

	var cmd *core.Cmd
	if isDir {
		cmd = mod.RegDirCmd(path, strings.TrimSpace(help)).SetSource(source)
	} else {
		cmd = mod.RegFileCmd(path, strings.TrimSpace(help)).SetSource(source)
	}

	abbrs := meta.Get("abbrs")
	if len(abbrs) == 0 {
		abbrs = meta.Get("abbr")
	}
	if len(abbrs) != 0 {
		mod.AddAbbrs(strings.Split(abbrs, abbrsSep)...)
	}

	args := meta.GetSession("args")
	if args == nil {
		args = meta.GetSession("arg")
	}
	if args != nil {
		for _, names := range args.Keys() {
			defVal := args.Get(names)
			nameAndAbbrs := strings.Split(names, abbrsSep)
			name := strings.TrimSpace(nameAndAbbrs[0])
			var argAbbrs []string
			for _, abbr := range nameAndAbbrs[1:] {
				argAbbrs = append(argAbbrs, strings.TrimSpace(abbr))
			}
			cmd.AddArg(name, defVal, argAbbrs...)
		}
	}

	envOps := meta.GetSession("env")
	if envOps != nil {
		for _, names := range envOps.Keys() {
			op := envOps.Get(names)
			segs := strings.Split(names, envPathSep)
			envAbbrs := cc.EnvAbbrs
			var path []string
			for _, seg := range segs {
				var abbrs []string
				// TODO: just use ":"
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

			key := strings.Join(path, envPathSep)
			opFields := strings.Split(op, abbrsSep)
			// TODO: change all "|" to ":" in envOps
			if len(opFields) == 1 {
				opFields = strings.Split(op, ":")
			}
			for _, it := range opFields {
				field := strings.ToLower(it)
				may := strings.Index(field, "may") >= 0 || strings.Index(field, "opt") >= 0
				write := strings.Index(field, "w") >= 0
				read := strings.Index(field, "rd") >= 0 ||
					strings.Index(field, "read") >= 0 || field == "r"
				if write && read {
					panic(fmt.Errorf("[LoadLocalMods.regMod] "+
						"parse env r|w definition failed: %v", it))
				}
				if write {
					if may {
						cmd.AddEnvOp(key, core.EnvOpTypeMayWrite)
					} else {
						cmd.AddEnvOp(key, core.EnvOpTypeWrite)
					}
				}
				if read {
					if may {
						cmd.AddEnvOp(key, core.EnvOpTypeMayRead)
					} else {
						cmd.AddEnvOp(key, core.EnvOpTypeRead)
					}
				}
			}
		}
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && !info.IsDir()
}
