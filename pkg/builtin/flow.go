package builtin

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
	"github.com/pingcap/ticat/pkg/proto/flow_file"
	"github.com/pingcap/ticat/pkg/utils"
)

func ListFlows(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	flowExt := env.GetRaw("strs.flow-ext")
	root := env.GetRaw("sys.paths.flows")
	if len(root) == 0 {
		panic(fmt.Errorf("[ListFlows] env 'sys.paths.flows' is empty"))
	}

	screen := display.NewCacheScreen()
	findStrs := getFindStrsFromArgv(argv)

	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if path == root {
			return nil
		}
		if !strings.HasSuffix(path, flowExt) {
			return nil
		}

		cmdPath := getCmdPath(path, flowExt)
		flowStr, help, abbrsStr := flow_file.LoadFlowFile(path)

		matched := true
		for _, findStr := range findStrs {
			if len(findStr) == 0 {
				continue
			}
			if strings.Index(cmdPath, findStr) < 0 &&
				strings.Index(help, findStr) < 0 &&
				strings.Index(abbrsStr, findStr) < 0 &&
				strings.Index(flowStr, findStr) < 0 {
				matched = false
				break
			}
		}
		if !matched {
			return nil
		}

		screen.Print(fmt.Sprintf("[%s]\n", cmdPath))
		if len(help) != 0 {
			screen.Print(fmt.Sprintf("      '%s'\n", help))
		}
		if len(abbrsStr) != 0 {
			screen.Print("    - abbrs:\n")
			screen.Print(fmt.Sprintf("        %s\n", abbrsStr))
		}
		screen.Print("    - flow:\n")
		screen.Print(fmt.Sprintf("        %s\n", flowStr))
		screen.Print("    - executable:\n")
		screen.Print(fmt.Sprintf("        %s\n", path))
		return nil
	})

	if screen.OutputNum() > 0 {
		display.PrintTipTitle(cc.Screen, env, "all saved flows: (flows from added repos are not included)")
	} else {
		helpStr := []string{
			"there is no saved flows yet, save flow by:", "",
		}
		selfName := env.GetRaw("strs.self-name")
		helpStr = append(helpStr, display.SuggestStrsFlowAdd(selfName)...)
		display.PrintTipTitle(cc.Screen, env, helpStr...)
	}
	screen.WriteTo(cc.Screen)
	return true
}

func RemoveFlow(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	cmdPath, filePath := getFlowCmdPath(argv, cc, env, true, "cmd-path", "RemoveFlow")
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		panic(fmt.Errorf("[RemoveFlow] path '%s' does not exist", filePath))
	}
	err = os.Remove(filePath)
	if err != nil {
		panic(fmt.Errorf("[RemoveFlow] remove flow file '%s' failed: %v",
			filePath, err))
	}
	cc.Screen.Print(fmt.Sprintf("[%s] (removed)\n", cmdPath))
	cc.Screen.Print(fmt.Sprintf("    %s\n", filePath))
	return true
}

func RemoveAllFlows(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	flowExt := env.GetRaw("strs.flow-ext")
	root := env.GetRaw("sys.paths.flows")
	if len(root) == 0 {
		panic(fmt.Errorf("[ListFlows] env 'sys.paths.flows' is empty"))
	}

	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if path != root && strings.HasSuffix(path, flowExt) {
			err = os.Remove(path)
			if err != nil {
				panic(fmt.Errorf("[RemoveAllFlows] remove flow file '%s' failed: %v",
					path, err))
			}
			cmdPath := getCmdPath(path, flowExt)
			cc.Screen.Print(fmt.Sprintf("[%s] (removed)\n", cmdPath))
			cc.Screen.Print(fmt.Sprintf("    %s\n", path))
		}
		return nil
	})
	return true
}

func SaveFlow(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath, filePath := getFlowCmdPath(argv, cc, env, false, "to-cmd-path", "SaveFlow")

	_, err := os.Stat(filePath)
	if !os.IsNotExist(err) {
		if !env.GetBool("sys.interact") {
			panic(fmt.Errorf("[SaveFlow] path '%s' exists", filePath))
		} else {
			cc.Screen.Print(fmt.Sprintf("[confirm] flow file of '%s' exists, "+
				"press enter to overwrite\n", cmdPath))
			utils.UserConfirm()
		}
	}

	width := env.GetInt("display.width")

	w := bytes.NewBuffer(nil)
	flow.RemoveLeadingCmds(1)
	saveFlow(w, flow, cc.Cmds.Strs.PathSep, env)
	data := w.String()
	cc.Screen.Print(strings.Repeat("-", width) + "\n")
	cc.Screen.Print(data)
	cc.Screen.Print(strings.Repeat("-", width) + "\n")
	cc.Screen.Print(fmt.Sprintf("[%s]\n", cmdPath))
	cc.Screen.Print(fmt.Sprintf("    %s\n", filePath))

	dirPath := filepath.Dir(filePath)
	os.MkdirAll(dirPath, os.ModePerm)

	tmp := filePath + ".tmp"
	err = ioutil.WriteFile(tmp, []byte(data), 0644)
	if err != nil {
		panic(fmt.Errorf("[SaveFlow] write flow file '%s' failed: %v", tmp, err))
	}

	err = os.Rename(tmp, filePath)
	if err != nil {
		panic(fmt.Errorf("[SaveFlow] rename flow file from '%s' to '%s' failed: %v",
			tmp, filePath, err))
	}

	flow.Cmds = nil
	return 0, true
}

func SetFlowHelpStr(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	help := argv.GetRaw("help-str")
	_, path := getFlowCmdPath(argv, cc, env, true, "cmd-path", "SetFlowHelpStr")
	flowStr, _, abbrsStr := flow_file.LoadFlowFile(path)
	flow_file.SaveFlowFile(path, flowStr, help, abbrsStr)
	return true
}

func LoadFlows(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	root := env.GetRaw("sys.paths.flows")
	if len(root) == 0 {
		panic(fmt.Errorf("[LoadFlows] env 'sys.paths.flows' is empty"))
	}
	loadFlowsFromDir(root, cc, env, root)
	return true
}

func LoadFlowsFromDir(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	root := argv.GetRaw("path")
	loadFlowsFromDir(root, cc, env, root)
	return true
}

func loadFlowsFromDir(root string, cc *core.Cli, env *core.Env, source string) bool {
	if len(root) > 0 && root[len(root)-1] == filepath.Separator {
		root = root[:len(root)-1]
	}
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return true
		}
		panic(fmt.Errorf("[LoadLocalFlows] access flows dir '%s' failed: %v", root, err))
	}
	if !info.IsDir() {
		panic(fmt.Errorf("[LoadLocalFlows] flows dir '%s' is not dir", root))
	}

	flowExt := env.GetRaw("strs.flow-ext")

	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, flowExt) {
			return nil
		}
		loadFlow(cc, root, path, flowExt, source)
		return nil
	})
	return true
}

func loadFlow(cc *core.Cli, root string, path string, flowExt string, source string) {
	var cmdPathStr string
	defer func() {
		// TODO: configurable display
		if err := recover(); err != nil {
			cc.Screen.Error("======================================\n\n")
			cc.Screen.Error("[ERR] flow loading failed:\n")
			if len(cmdPathStr) != 0 {
				cc.Screen.Error("    - cmd:\n")
				cc.Screen.Error(fmt.Sprintf("        %s\n", cmdPathStr))
			}
			cc.Screen.Error("    - source:\n")
			cc.Screen.Error(fmt.Sprintf("        %s\n", source))
			cc.Screen.Error("    - error:\n")
			cc.Screen.Error(fmt.Sprintf("        %s\n", err.(error).Error()))
			cc.Screen.Error("\n======================================\n\n")
		}
	}()

	flowStr, help, abbrsStr := flow_file.LoadFlowFile(path)

	pathSep := cc.Cmds.Strs.PathSep
	abbrsSep := cc.Cmds.Strs.AbbrsSep
	var abbrs [][]string
	for _, abbrSeg := range strings.Split(abbrsStr, pathSep) {
		abbrList := strings.Split(abbrSeg, abbrsSep)
		abbrs = append(abbrs, abbrList)
	}

	cmdPathStr = getCmdPath(path, flowExt)
	cmdPath := strings.Split(cmdPathStr, pathSep)
	flow := cc.Cmds
	for i, cmd := range cmdPath {
		flow = flow.GetOrAddSub(cmd)
		if i < len(abbrs) {
			flow.AddAbbrs(abbrs[i]...)
		}
	}

	flow.RegFlowCmd(flowStr, help).SetSource(source)
}

func saveFlow(w io.Writer, flow *core.ParsedCmds, cmdPathSep string, env *core.Env) {
	envPathSep := env.GetRaw("strs.env-path-sep")
	bracketLeft := env.GetRaw("strs.env-bracket-left")
	bracketRight := env.GetRaw("strs.env-bracket-right")
	envKeyValSep := env.GetRaw("strs.env-kv-sep")
	seqSep := env.GetRaw("strs.seq-sep")
	if len(envPathSep) == 0 || len(bracketLeft) == 0 || len(bracketRight) == 0 ||
		len(envKeyValSep) == 0 || len(seqSep) == 0 {
		panic("[saveFlow] some predefined strs not found")
	}

	for i, cmd := range flow.Cmds {
		if len(flow.Cmds) > 1 {
			if i == 0 {
				if flow.GlobalCmdIdx < 0 {
					fmt.Fprint(w, seqSep+" ")
				}
			} else {
				fmt.Fprint(w, " "+seqSep+" ")
			}
		}

		var path []string
		var lastSegHasNoCmd bool
		var cmdHasEnv bool

		for j, seg := range cmd.Segments {
			if len(cmd.Segments) > 1 && j != 0 && !lastSegHasNoCmd {
				fmt.Fprint(w, cmdPathSep)
			}
			fmt.Fprint(w, seg.Matched.Name)

			if seg.Matched.Cmd != nil {
				path = append(path, seg.Matched.Cmd.Name())
			} else {
				path = append(path, seg.Matched.Name)
			}
			lastSegHasNoCmd = (seg.Matched.Cmd == nil)
			cmdHasEnv = cmdHasEnv || saveEnv(w, seg.Env, path, envPathSep,
				bracketLeft, bracketRight, envKeyValSep,
				!cmdHasEnv && j == len(cmd.Segments)-1)
		}
	}
	fmt.Fprintf(w, "\n")
}

func saveEnv(
	w io.Writer,
	env core.ParsedEnv,
	prefixPath []string,
	pathSep string,
	bracketLeft string,
	bracketRight string,
	envKeyValSep string,
	useArgsFmt bool) bool {

	if len(env) == 0 {
		return false
	}

	isAllArgs := true
	for _, v := range env {
		if !v.IsArg {
			isAllArgs = false
			break
		}
	}

	prefix := strings.Join(prefixPath, pathSep) + pathSep

	var kvs []string
	for k, v := range env {
		if strings.HasPrefix(k, prefix) && len(k) != len(prefix) {
			k = strings.Join(v.MatchedPath[len(prefixPath):], pathSep)
		}
		kvs = append(kvs, fmt.Sprintf("%v%s%v", k, envKeyValSep, quoteIfHasSpace(v.Val)))
	}

	format := bracketLeft + "%s" + bracketRight
	if isAllArgs && useArgsFmt {
		format = " %s"
	}
	fmt.Fprintf(w, format, strings.Join(kvs, " "))
	return true
}

func quoteIfHasSpace(str string) string {
	if strings.IndexAny(str, " \t\r\n") < 0 {
		return str
	}
	i := strings.Index(str, "\"")
	if i < 0 {
		return "\"" + str + "\""
	}
	i = strings.Index(str, "'")
	if i < 0 {
		return "'" + str + "'"
	}
	return str
}

func getFlowCmdPath(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	expectExists bool,
	argName string,
	funcName string) (cmdPath string, filePath string) {

	cmdPath = normalizeCmdPath(argv.GetRaw(argName),
		cc.Cmds.Strs.PathSep, cc.Cmds.Strs.PathAlterSeps)
	if len(cmdPath) == 0 {
		origin := argv.GetRaw(argName)
		if len(origin) == 0 {
			panic(fmt.Errorf("[%s] arg '%s' is empty", funcName, argName))
		} else {
			panic(fmt.Errorf("[%s] arg '%s' is empty after normalizing: %s -> %s",
				funcName, argName, origin, cmdPath))
		}
	}

	flowExt := env.GetRaw("strs.flow-ext")
	root := env.GetRaw("sys.paths.flows")
	if len(root) == 0 {
		panic(fmt.Errorf("[%s] env 'sys.paths.flows' is empty", funcName))
	}

	filePath = filepath.Join(root, cmdPath) + flowExt
	if !expectExists && fileExists(filePath) {
		if !env.GetBool("sys.interact") {
			panic(fmt.Errorf("[%s] flow '%s' file '%s' exists", funcName, cmdPath, filePath))
		} else {
			return
		}
	}
	if expectExists && !fileExists(filePath) {
		panic(fmt.Errorf("[%s] flow '%s' file '%s' not exists", funcName, cmdPath, filePath))
	}
	return
}

func normalizeCmdPath(path string, sep string, alterSeps string) string {
	var segs []string
	for len(path) > 0 {
		i := strings.IndexAny(path, alterSeps)
		if i < 0 {
			segs = append(segs, path)
			break
		} else if i == 0 {
			path = path[1:]
		} else {
			segs = append(segs, path[0:i])
			path = path[i+1:]
		}
	}
	return strings.Join(segs, sep)
}

func getCmdPath(path string, flowExt string) string {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, flowExt) {
		panic(fmt.Errorf("[getCmdPath] flow file '%s' ext not match '%s'", path, flowExt))
	}
	return base[:len(base)-len(flowExt)]
}
