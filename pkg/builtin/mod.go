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
	bashExt := "." + env.Get("strs.proto-bash-ext").Raw

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
		ext = filepath.Ext(target)
		if ext == bashExt {
			regBashMod(cc, path, target, target[len(root)+1:len(target)-len(ext)], abbrsSep)
		}
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

func regBashMod(cc *core.Cli, metaPath string, filePath string, cmdPath string, abbrsSep string) {
	mod := cc.Cmds.GetOrAddSub(strings.Split(cmdPath, string(filepath.Separator))...)

	ini := goini.New()
	err := ini.ParseFile(metaPath)
	if err != nil {
		panic(fmt.Errorf("[LoadLocalMods.regBashMod] parse mod's meta file failed: %v", err))
	}
	help, _ := ini.SectionGet("", "help")
	cmd := mod.RegBashCmd(filePath, help)
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
			name := nameAndAbbrs[0]
			argAbbrs := nameAndAbbrs[1:]
			cmd.AddArg(name, defVal, argAbbrs...)
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
