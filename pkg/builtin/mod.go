package builtin

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/proto/mod_meta"
)

func SetExtExec(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)
	env = env.GetLayer(core.EnvLayerDefault)
	env.Set("sys.ext.exec.bash", "bash")
	env.Set("sys.ext.exec.sh", "sh")
	env.Set("sys.ext.exec.py", "python")
	env.Set("sys.ext.exec.go", "go run")
	return currCmdIdx, true
}

func loadLocalMods(
	cc *core.Cli,
	root string,
	reposFileName string,
	metaExt string,
	flowExt string,
	helpExt string,
	abbrsSep string,
	envPathSep string,
	source string,
	panicRecover bool) {

	if len(root) > 0 && root[len(root)-1] == filepath.Separator {
		root = root[:len(root)-1]
	}

	// TODO: return filepath.SkipDir to avoid some non-sense scanning
	filepath.Walk(root, func(metaPath string, info fs.FileInfo, err error) error {
		if info != nil && info.IsDir() {
			// Skip hidden file or dir
			base := filepath.Base(metaPath)
			if len(base) > 0 && base[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}
		if metaPath == filepath.Join(root, reposFileName) {
			return nil
		}

		if strings.HasSuffix(metaPath, helpExt) {
			cc.Helps.RegHelpFile(metaPath)
			return nil
		}

		if strings.HasSuffix(metaPath, flowExt) {
			cmdPath := filepath.Base(metaPath[0 : len(metaPath)-len(flowExt)])
			cmdPaths := strings.Split(cmdPath, cc.Cmds.Strs.PathSep)
			mod_meta.RegMod(cc, metaPath, "", false, true, cmdPaths,
				abbrsSep, envPathSep, source, panicRecover)
			return nil
		}

		ext := filepath.Ext(metaPath)
		if ext != metaExt {
			return nil
		}
		targetPath := metaPath[0 : len(metaPath)-len(ext)]

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

		cmdPaths := strings.Split(cmdPath, string(filepath.Separator))
		mod_meta.RegMod(cc, metaPath, targetPath, isDir, false, cmdPaths,
			abbrsSep, envPathSep, source, panicRecover)
		return nil
	})
}
