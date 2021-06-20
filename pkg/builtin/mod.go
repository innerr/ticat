package builtin

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/proto/mod_meta"
)

func SetExtExec(_ core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	env = env.GetLayer(core.EnvLayerDefault)
	env.Set("sys.ext.exec.bash", "bash")
	env.Set("sys.ext.exec.sh", "sh")
	env.Set("sys.ext.exec.py", "python")
	env.Set("sys.ext.exec.go", "go run")
	return true
}

func loadLocalMods(
	cc *core.Cli,
	root string,
	reposFileName string,
	metaExt string,
	flowExt string,
	abbrsSep string,
	envPathSep string,
	source string) {

	if len(root) > 0 && root[len(root)-1] == filepath.Separator {
		root = root[:len(root)-1]
	}

	// TODO: return filepath.SkipDir to avoid some non-sense scanning
	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if info != nil && info.IsDir() {
			// Skip hidden file or dir
			base := filepath.Base(path)
			if len(base) > 0 && base[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Base(path) == reposFileName {
			return nil
		}

		if strings.HasSuffix(path, flowExt) {
			loadFlow(cc, root, path, flowExt, source)
			return nil
		}
		ext := filepath.Ext(path)
		if ext != metaExt {
			return nil
		}
		targetPath := path[0 : len(path)-len(ext)]
		metaPath := path

		// Note: strip all ext(s) from cmd-path
		cmdPath := targetPath[len(root)+1:]
		for {
			ext := filepath.Ext(cmdPath)
			if len(ext) == 0 {
				break
			} else {
				cmdPath = cmdPath[0 : len(cmdPath)-len(ext)]
			}
		}

		isDir := false
		info, err = os.Stat(targetPath)
		if os.IsNotExist(err) {
			targetPath = ""
		} else if err == nil {
			isDir = info.IsDir()
		}

		mod_meta.RegMod(cc, metaPath, targetPath, isDir, cmdPath, abbrsSep, envPathSep, source)
		return nil
	})
}
