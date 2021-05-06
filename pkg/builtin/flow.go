package builtin

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func SaveFlow(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	cmds []core.ParsedCmd,
	currCmdIdx int,
	input []string) ([]core.ParsedCmd, int, bool) {

	cmdPath := argv.GetRaw("to-cmd-path")
	if len(cmdPath) == 0 {
		panic(fmt.Errorf("[SaveFlow] arg 'to-cmd-path' is empty"))
	}
	root := env.Get("sys.paths.mods").Raw
	if len(cmdPath) == 0 {
		panic(fmt.Errorf("[SaveFlow] env 'sys.paths.mods' is empty"))
	}

	filePath := filepath.Join(root, cmdPath)
	if fileExists(filePath) {
		panic(fmt.Errorf("[SaveFlow] cmd path '%s' is not empty", cmdPath))
	}

	_, err := os.Stat(filePath)
	if !os.IsNotExist(err) {
		panic(fmt.Errorf("[SaveFlow] path '%s' exists", filePath))
	}

	dirPath := filepath.Dir(filePath)
	os.MkdirAll(dirPath, os.ModePerm)

	tmp := filePath + ".tmp"
	file, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(fmt.Errorf("[SaveFlow] open flow file '%s' failed: %v", tmp, err))
	}

	w := bufio.NewWriter(file)

	for i, it := range input {
		// TODO: better quoting, check if it's already quoted
		if strings.IndexAny(it, " \t\n\r") >= 0 {
			it = "'" + it + "'"
		}
		// TODO: wrap if it's too long
		if i != len(input)-1 {
			it += " "
		}
		fmt.Fprint(w, it)
	}
	fmt.Fprintln(w)
	w.Flush()

	err = file.Close()
	if err != nil {
		panic(fmt.Errorf("[SaveFlow] close flow file '%s' failed: %v", tmp, err))
	}

	err = os.Rename(tmp, filePath)
	if err != nil {
		panic(fmt.Errorf("[SaveFlow] rename flow file from '%s' to '%s' failed: %v",
			tmp, filePath, err))
	}
	return nil, 0, true
}
