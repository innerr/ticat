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

func LoadLocalMods(_ core.ArgVals, cc *core.Cli, env *core.Env) bool {
	root := env.Get("sys.paths.mods").Raw
	metaExt := "." + env.Get("strs.meta-ext").Raw
	abbrsSep := env.Get("strs.abbrs-sep").Raw
	envPathSep := env.Get("strs.env-path-sep").Raw

	if root[len(root)-1] == filepath.Separator {
		root = root[:len(root)-1]
	}
	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		ext := filepath.Ext(path)
		if ext != metaExt {
			return nil
		}
		target := path[0 : len(path)-len(ext)]

		if dirExists(target) {
			regDirMod(cc, path, target, target[len(root)+1:], abbrsSep)
			return nil
		}
		if !fileExists(target) {
			return nil
		}

		cmdPath := target[len(root)+1:]
		for {
			ext := filepath.Ext(cmdPath)
			if len(ext) == 0 {
				break
			} else {
				cmdPath = cmdPath[0 : len(cmdPath)-len(ext)]
			}
		}
		regBashMod(cc, path, target, cmdPath, abbrsSep, envPathSep)
		return nil
	})
	return true
}

// TODO: mod's meta definition file (*.ticat) should have a formal formal manager

func regDirMod(cc *core.Cli, metaPath string, dirPath string, cmdPath string, abbrsSep string) {
	mod := cc.Cmds.GetOrAddSub(strings.Split(cmdPath, string(filepath.Separator))...)

	ini := goini.New()
	err := ini.ParseFile(metaPath)
	if err != nil {
		panic(fmt.Errorf("[LoadLocalMods.regDirMod] parse mod's meta file failed: %v", err))
	}
	abbrs, ok := ini.SectionGet("", "abbrs")
	if ok {
		abbrs = strings.Trim(abbrs, "'\"")
		mod.AddAbbrs(strings.Split(abbrs, abbrsSep)...)
	}
}

func regBashMod(
	cc *core.Cli,
	metaPath string,
	filePath string,
	cmdPath string,
	abbrsSep string,
	envPathSep string) {

	mod := cc.Cmds.GetOrAddSub(strings.Split(cmdPath, string(filepath.Separator))...)

	ini := goini.New()
	err := ini.ParseFile(metaPath)
	if err != nil {
		panic(fmt.Errorf("[LoadLocalMods.regBashMod] parse mod's meta file failed: %v", err))
	}

	help, _ := ini.SectionGet("", "help")
	cmd := mod.RegBashCmd(filePath, strings.TrimSpace(strings.Trim(help, "'\"")))

	abbrs, ok := ini.SectionGet("", "abbrs")
	if ok {
		abbrs = strings.Trim(abbrs, "'\"")
		mod.AddAbbrs(strings.Split(abbrs, abbrsSep)...)
	}

	args, ok := ini.GetKvmap("args")
	if ok {
		for names, defVal := range args {
			names = strings.Trim(names, "'\"")
			defVal = strings.Trim(defVal, "'\"")
			nameAndAbbrs := strings.Split(names, abbrsSep)
			name := strings.TrimSpace(nameAndAbbrs[0])
			var argAbbrs []string
			for _, abbr := range nameAndAbbrs[1:] {
				argAbbrs = append(argAbbrs, strings.TrimSpace(abbr))
			}
			cmd.AddArg(name, defVal, argAbbrs...)
		}
	}

	envOps, ok := ini.GetKvmap("env")
	if ok {
		for names, op := range envOps {
			segs := strings.Split(strings.Trim(names, "'\""), envPathSep)
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
			opFields := strings.Split(strings.Trim(op, "'\""), abbrsSep)
			for _, it := range opFields {
				field := strings.ToLower(it)
				may := strings.Index(field, "may") >= 0 || strings.Index(field, "opt") >= 0
				write := strings.Index(field, "w") >= 0
				read := strings.Index(field, "rd") >= 0 ||
					strings.Index(field, "read") >= 0 || field == "r"
				if write && read {
					panic(fmt.Errorf("[LoadLocalMods.regBashMod] "+
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

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && !info.IsDir()
}
