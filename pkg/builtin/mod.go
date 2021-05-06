package builtin

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/builtin/proto"
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

// TODO: mod's meta definition file (*.ticat) should have a formal manager

func regDirMod(cc *core.Cli, metaPath string, dirPath string, cmdPath string, abbrsSep string) {
	mod := cc.Cmds.GetOrAddSub(strings.Split(cmdPath, string(filepath.Separator))...)

	meta, err := proto.NewMeta(metaPath, "\n", "=")
	if err != nil {
		panic(fmt.Errorf("[LoadLocalMods.regDirMod] parse mod's meta file failed: %v", err))
	}
	abbrs := meta.SectionGet("", "abbrs")
	if len(abbrs) != 0 {
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

	meta, err := proto.NewMeta(metaPath, "\n", "=")
	if err != nil {
		panic(fmt.Errorf("[LoadLocalMods.regBashMod] parse mod's meta file failed: %v", err))
	}

	help := meta.SectionGet("", "help")
	cmd := mod.RegBashCmd(filePath, strings.TrimSpace(help))

	abbrs := meta.SectionGet("", "abbrs")
	if len(abbrs) != 0 {
		mod.AddAbbrs(strings.Split(abbrs, abbrsSep)...)
	}

	args, ok := meta.GetSection("args")
	if ok {
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

	envOps, ok := meta.GetSection("env")
	if ok {
		for _, names := range envOps.Keys() {
			op := envOps.Get(names)
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

func SetExtExec(_ core.ArgVals, cc *core.Cli, env *core.Env) bool {
	env = env.GetLayer(core.EnvLayerDefault)
	env.Set("sys.ext.exec.bash", "bash")
	env.Set("sys.ext.exec.sh", "sh")
	env.Set("sys.ext.exec.py", "python")
	env.Set("sys.ext.exec.go", "go run")
	return true
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && !info.IsDir()
}
