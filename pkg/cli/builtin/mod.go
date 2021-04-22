package builtin

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli"
	"github.com/zieckey/goini"
)

func LoadLocalMods(_ cli.ArgVals, cc *cli.Cli, env *cli.Env) bool {
	root := env.Get("runtime.sys.paths.mods").Raw
	if root[len(root)-1] == filepath.Separator {
		root = root[:len(root)-1]
	}
	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		ext := filepath.Ext(path)
		if ext != "."+cli.SelfName {
			return nil
		}
		target := path[0 : len(path)-len(ext)]

		if dirExists(target) {
			regDirMod(cc, path, target, target[len(root)+1:])
			return nil
		}
		if !fileExists(target) {
			return nil
		}
		ext = filepath.Ext(target)
		if ext == ".bash" {
			regBashMod(cc, path, target, target[len(root)+1:len(target)])
		}
		return nil
	})
	return true
}

// TODO: mod's meta definition file (*.ticat) should have a formal formal manager

func regDirMod(cc *cli.Cli, metaPath string, dirPath string, cmdPath string) {
	mod := cc.Cmds.GetOrAddSub(strings.Split(cmdPath, string(filepath.Separator))...)

	ini := goini.New()
	err := ini.ParseFile(metaPath)
	if err != nil {
		panic(fmt.Errorf("[LoadLocalMods.regDirMod] parse mod's meta file failed: %v", err))
	}
	abbrs, ok := ini.SectionGet("", "abbrs")
	if ok {
		mod.AddAbbr(strings.Split(abbrs, cli.AbbrSep)...)
	}
}

func regBashMod(cc *cli.Cli, metaPath string, filePath string, cmdPath string) {
	mod := cc.Cmds.GetOrAddSub(strings.Split(cmdPath, string(filepath.Separator))...)
	cmd := mod.RegBashCmd(filePath)

	ini := goini.New()
	err := ini.ParseFile(metaPath)
	if err != nil {
		panic(fmt.Errorf("[LoadLocalMods.regBashMod] parse mod's meta file failed: %v", err))
	}
	abbrs, ok := ini.SectionGet("", "abbrs")
	if ok {
		mod.AddAbbr(strings.Split(abbrs, cli.AbbrSep)...)
	}
	args, ok := ini.GetKvmap("args")
	if ok {
		for names, defVal := range args {
			nameAndAbbrs := strings.Split(names, cli.AbbrSep)
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
