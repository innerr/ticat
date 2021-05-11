package builtin

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/zieckey/goini"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func SetExtExec(_ core.ArgVals, cc *core.Cli, env *core.Env) bool {
	env = env.GetLayer(core.EnvLayerDefault)
	env.Set("sys.ext.exec.bash", "bash")
	env.Set("sys.ext.exec.sh", "sh")
	env.Set("sys.ext.exec.py", "python")
	env.Set("sys.ext.exec.go", "go run")
	return true
}

func LoadLocalMods(_ core.ArgVals, cc *core.Cli, env *core.Env) bool {
	root := env.Get("sys.paths.mods").Raw
	metaExt := "." + env.Get("strs.meta-ext").Raw
	abbrsSep := env.Get("strs.abbrs-sep").Raw
	envPathSep := env.Get("strs.env-path-sep").Raw
	loadLocalMods(cc, root, metaExt, abbrsSep, envPathSep)
	return true
}

func loadLocalMods(
	cc *core.Cli,
	root string,
	metaExt string,
	abbrsSep string,
	envPathSep string) {

	if len(root) > 0 && root[len(root)-1] == filepath.Separator {
		root = root[:len(root)-1]
	}

	// TODO: return filepath.SkipDir to avoid non-sense scanning
	filepath.Walk(root, func(metaPath string, info fs.FileInfo, err error) error {
		if len(metaPath) > 0 && metaPath[0] == filepath.Separator {
			return filepath.SkipDir
		}
		ext := filepath.Ext(metaPath)
		if ext != metaExt {
			return nil
		}
		target := metaPath[0 : len(metaPath)-len(ext)]

		info, err = os.Stat(target)
		if os.IsNotExist(err) {
			panic(fmt.Errorf("[LoadLocalMods] target '%s' of meta file '%s' not exists",
				target, metaPath))
			return nil
		}

		// Strip all ext from cmd-path
		cmdPath := target[len(root)+1:]
		for {
			ext := filepath.Ext(cmdPath)
			if len(ext) == 0 {
				break
			} else {
				cmdPath = cmdPath[0 : len(cmdPath)-len(ext)]
			}
		}

		regMod(cc, metaPath, target, info.IsDir(), cmdPath, abbrsSep, envPathSep)
		return nil
	})
}

// TODO: mod's meta definition file (*.ticat) should have a formal manager

func regMod(
	cc *core.Cli,
	metaPath string,
	path string,
	isDir bool,
	cmdPath string,
	abbrsSep string,
	envPathSep string) {

	mod := cc.Cmds.GetOrAddSub(strings.Split(cmdPath, string(filepath.Separator))...)

	meta := goini.New()
	meta.SetTrimQuotes(true)
	err := meta.ParseFile(metaPath)
	if err != nil {
		panic(fmt.Errorf("[LoadLocalMods.regMod] parse mod's meta file failed: %v", err))
	}

	// 'cmd' should be a relative path base on this file
	cmdLine, _ := meta.Get("cmd")
	help, _ := meta.Get("help")
	if len(help) == 0 && (!isDir || len(cmdLine) != 0) {
		panic(fmt.Errorf("[LoadLocalMods.regMod] cmd '%s' has no help string in '%s'",
			cmdPath, metaPath))
	}

	if len(cmdLine) != 0 {
		path, err = filepath.Abs(filepath.Join(path, cmdLine))
		if err != nil {
			panic(fmt.Errorf("[LoadLocalMods.regMod] cmd '%s' get abs path of '%s' failed",
				cmdPath, path))
		}
		if !fileExists(path) {
			panic(fmt.Errorf("[LoadLocalMods.regMod] cmd '%s' point to a not existed file '%s'",
				cmdPath, path))
		}
	}
	var cmd *core.Cmd
	if isDir {
		cmd = mod.RegDirCmd(path, strings.TrimSpace(help))
	} else {
		cmd = mod.RegFileCmd(path, strings.TrimSpace(help))
	}

	abbrs, _ := meta.Get("abbrs")
	if len(abbrs) != 0 {
		mod.AddAbbrs(strings.Split(abbrs, abbrsSep)...)
	}

	args, ok := meta.GetKvmap("args")
	if ok {
		for names, defVal := range args {
			nameAndAbbrs := strings.Split(names, abbrsSep)
			name := strings.TrimSpace(nameAndAbbrs[0])
			var argAbbrs []string
			for _, abbr := range nameAndAbbrs[1:] {
				argAbbrs = append(argAbbrs, strings.TrimSpace(abbr))
			}
			cmd.AddArg(name, defVal, argAbbrs...)
		}
	}

	envOps, ok := meta.GetKvmap("env")
	if ok {
		for names, op := range envOps {
			segs := strings.Split(names, envPathSep)
			envAbbrs := cc.EnvAbbrs
			var path []string
			for _, seg := range segs {
				var abbrs []string
				for _, abbr := range strings.Split(seg, abbrsSep) {
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
