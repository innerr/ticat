package builtin

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/proto/mod_meta"
)

func SetExtExec(_ core.ArgVals, cc *core.Cli, env *core.Env) bool {
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
		if strings.HasSuffix(path, flowExt) {
			loadFlow(cc, root, path, flowExt)
			return nil
		}
		ext := filepath.Ext(path)
		if ext != metaExt {
			return nil
		}
		target := path[0 : len(path)-len(ext)]
		metaPath := path

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

		mod_meta.RegMod(cc, metaPath, target, info.IsDir(), cmdPath, abbrsSep, envPathSep, source)
		return nil
	})
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && !info.IsDir()
}
