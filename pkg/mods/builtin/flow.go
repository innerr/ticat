package builtin

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/innerr/ticat/pkg/cli/display"
	"github.com/innerr/ticat/pkg/core/model"
	"github.com/innerr/ticat/pkg/mods/persist/flow_file"
	"github.com/innerr/ticat/pkg/mods/persist/mod_meta"
	"github.com/innerr/ticat/pkg/utils"
)

func ListFlows(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	flowExt := env.GetRaw("strs.flow-ext")
	root := getFlowRoot(env, flow.Cmds[currCmdIdx])

	screen := display.NewCacheScreen()
	findStrs := getFindStrsFromArgvAndFlow(flow, currCmdIdx, argv)

	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if path == root {
			return nil
		}
		if !strings.HasSuffix(path, flowExt) {
			return nil
		}

		cmdPath, _ := getCmdPath(path, flowExt, flow.Cmds[currCmdIdx])
		flowStrs, help, abbrsStr, _, _, _, err := flow_file.LoadFlowFile(path)
		if err != nil {
			return nil
		}
		flowStr := strings.Join(flowStrs, " ")

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

		screen.Print(fmt.Sprintf(display.ColorCmd("[%s]\n", env), cmdPath))
		if len(help) != 0 {
			screen.Print("     " + display.ColorHelp("'"+help+"'", env) + "\n")
		}
		if len(abbrsStr) != 0 {
			screen.Print("    " + display.ColorProp("- abbrs:", env) + "\n")
			screen.Print(fmt.Sprintf("        %s\n", abbrsStr))
		}
		screen.Print("    " + display.ColorProp("- flow:", env) + "\n")
		for _, flowStr := range flowStrs {
			screen.Print("        " + display.ColorFlow(flowStr, env) + "\n")
		}
		screen.Print("    " + display.ColorProp("- executable:", env) + "\n")
		screen.Print(fmt.Sprintf("        %s\n", path))
		return nil
	})

	if screen.OutputtedLines() > 0 {
		display.PrintTipTitle(cc.Screen, env,
			"all saved flows: (flows from added repos are not included)")
	} else {
		display.PrintTipTitle(cc.Screen, env,
			"there is no saved flows yet, save flow by:",
			"",
			display.SuggestFlowAdd(env))
	}
	screen.WriteTo(cc.Screen)
	if display.TooMuchOutput(env, screen) {
		display.PrintTipTitle(cc.Screen, env,
			"filter flows by keywords if there are too many:",
			"",
			display.SuggestFlowsFilter(env))
	}
	return currCmdIdx, nil
}

func RenameFlow(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	argSrcCmdPath, srcCmdPath, srcFilePath, _ := getFlowCmdPath(flow, currCmdIdx, true, "", argv, cc, env, true, "src")
	argDestCmdPath, destCmdPath, destFilePath, _ := getFlowCmdPath(flow, currCmdIdx, true, "", argv, cc, env, false, "dest")

	_, err := os.Stat(srcFilePath)
	if os.IsNotExist(err) {
		return currCmdIdx, model.NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("path '%s' does not exist", srcFilePath))
	}

	err = os.Rename(srcFilePath, destFilePath)
	if err != nil {
		return currCmdIdx, model.NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("move flow file '%s' to '%s' failed: %v", srcFilePath, destFilePath, err))
	}

	realSrcCmdStr := ""
	if argSrcCmdPath != srcCmdPath {
		realSrcCmdStr = display.ColorExplain("("+srcCmdPath+")", env)
	}
	realDestCmdStr := ""
	if argDestCmdPath != destCmdPath {
		realDestCmdStr = display.ColorExplain("("+destCmdPath+")", env)
	}
	cc.Screen.Print(display.ColorCmd("["+argSrcCmdPath+"]", env) + realSrcCmdStr +
		display.ColorSymbol(" -> ", env) + display.ColorCmd("["+argDestCmdPath+"]", env) + realDestCmdStr + "\n")
	return currCmdIdx, nil
}

func RemoveFlow(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	argCmdPath, cmdPath, filePath, _ := getFlowCmdPath(flow, currCmdIdx, true, "", argv, cc, env, true, "cmd-path")
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return currCmdIdx, model.NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("path '%s' does not exist", filePath))
	}
	err = os.Remove(filePath)
	if err != nil {
		return currCmdIdx, model.NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("remove flow file '%s' failed: %v", filePath, err))
	}

	realCmdStr := ""
	if argCmdPath != cmdPath {
		realCmdStr = "(" + cmdPath + ")"
	}
	display.PrintTipTitle(cc.Screen, env,
		"flow '"+argCmdPath+"'"+realCmdStr+"  is removed")
	cc.Screen.Print(fmt.Sprintf(display.ColorCmd("[%s]", env)+
		display.ColorDisabled(" (removed)", env)+"\n", cmdPath))
	cc.Screen.Print(fmt.Sprintf("    %s\n", filePath))
	return currCmdIdx, nil
}

func RemoveAllFlows(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	flowExt := env.GetRaw("strs.flow-ext")
	root := getFlowRoot(env, flow.Cmds[currCmdIdx])
	screen := display.NewCacheScreen()

	var walkErr error
	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if path != root && strings.HasSuffix(path, flowExt) {
			err = os.Remove(path)
			if err != nil {
				walkErr = model.NewCmdError(flow.Cmds[currCmdIdx],
					fmt.Sprintf("remove flow file '%s' failed: %v", path, err))
				return err
			}
			cmdPath, _ := getCmdPath(path, flowExt, flow.Cmds[currCmdIdx])
			screen.Print(fmt.Sprintf(display.ColorCmd("[%s]", env)+
				display.ColorDisabled(" (removed)", env)+"\n", cmdPath))
			screen.Print(fmt.Sprintf("    %s\n", path))
		}
		return nil
	})

	if walkErr != nil {
		return currCmdIdx, walkErr
	}

	if screen.OutputtedLines() > 0 {
		display.PrintTipTitle(cc.Screen, env, "all saved flows are removed")
	} else {
		display.PrintTipTitle(cc.Screen, env,
			"there is no saved flows yet, nothing to do.",
			"",
			"save flow by:",
			"",
			display.SuggestFlowAdd(env))
	}
	screen.WriteTo(cc.Screen)
	return currCmdIdx, nil
}

func SaveFlow(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	quietOverwrite := argv.GetBool("quiet-overwrite")
	help := argv.GetRaw("help-str")
	trivialLvl := argv.GetRaw("unfold-trivial")
	autoArgs := argv.GetRaw("auto-args")
	toDir := argv.GetRaw("to-dir")
	packSub := argv.GetRaw("pack-subflow")

	argCmdPath, cmdPath, filePath, cmdExists := getFlowCmdPath(flow, currCmdIdx, false, toDir, argv, cc, env, false, "new-cmd-path")
	realCmdStr := ""
	if argCmdPath != cmdPath {
		realCmdStr = "(" + cmdPath + ")"
	}
	screen := display.NewCacheScreen()

	_, err := os.Stat(filePath)
	if !os.IsNotExist(err) {
		if quietOverwrite {
			// do nothing
		} else {
			if !env.GetBool("sys.confirm.ask") {
				return currCmdIdx, model.NewCmdError(flow.Cmds[currCmdIdx],
					fmt.Sprintf("path '%s' exists", filePath))
			} else {
				cc.Screen.Print(fmt.Sprintf(display.ColorTip("[confirm]", env)+
					" flow file of '%s'"+realCmdStr+" exists, "+
					"type "+display.ColorWarn("'y'", env)+" and press enter to "+
					display.ColorWarn("overwrite:", env)+"\n", argCmdPath))
				utils.UserConfirm()
			}
		}
	} else if cmdExists {
		display.PrintErrTitle(cc.Screen, env,
			"cmd '"+argCmdPath+"'"+realCmdStr+" exists but it is not a saved flow in default place.",
			"", "so can not be overwrited by 'flow.save', recommand to use 'cmd.edit' to modify it")
		return currCmdIdx, fmt.Errorf("cmd '%s' exists but is not a saved flow", argCmdPath)
	}

	flow.RemoveLeadingCmds(1)

	if !checkAndConfirmIfFlowHasParseError(cc.Screen, flow, env) {
		return currCmdIdx, fmt.Errorf("flow has parse error")
	}

	trivialMark := env.GetRaw("strs.trivial-mark")

	// TODO: wrap line if too long
	flowStr, _ := model.SaveFlowToStr(flow, cc.Cmds.Strs.PathSep, trivialMark, env)

	screen.Print(fmt.Sprintf(display.ColorCmd("[%s]", env)+"\n", cmdPath))
	screen.Print("    " + display.ColorProp("- flow:", env) + "\n")
	screen.Print("        " + display.ColorFlow(flowStr, env) + "\n")
	screen.Print("    " + display.ColorProp("- executable:", env) + "\n")
	screen.Print(fmt.Sprintf("        %s\n", filePath))

	dirPath := filepath.Dir(filePath)
	os.MkdirAll(dirPath, os.ModePerm)

	if err := flow_file.SaveFlowFile(filePath, []string{flowStr}, help, "", trivialLvl, autoArgs, packSub); err != nil {
		return currCmdIdx, err
	}

	display.PrintTipTitle(cc.Screen, env,
		"flow '"+argCmdPath+"'"+realCmdStr+" is saved, can be used as a command")
	screen.WriteTo(cc.Screen)
	return clearFlow(flow)
}

func SetFlowHelpStr(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}

	help := argv.GetRaw("help-str")
	argCmdPath, cmdPath, filePath, _ := getFlowCmdPath(flow, currCmdIdx, true, "", argv, cc, env, true, "cmd-path")
	flowStrs, oldHelp, abbrsStr, trivial, autoArgs, packSub, err := flow_file.LoadFlowFile(filePath)
	if err != nil {
		return currCmdIdx, err
	}
	if err := flow_file.SaveFlowFile(filePath, flowStrs, help, abbrsStr, trivial, autoArgs, packSub); err != nil {
		return currCmdIdx, err
	}

	realCmdStr := ""
	if argCmdPath != cmdPath {
		realCmdStr = "(" + cmdPath + ")"
	}
	display.PrintTipTitle(cc.Screen, env,
		"help string of flow '"+argCmdPath+"'"+realCmdStr+" is saved")

	cc.Screen.Print(display.ColorCmd(fmt.Sprintf("[%s]", cmdPath), env) + "\n")
	cc.Screen.Print("     " + display.ColorHelp("'"+help+"'", env) + "\n")
	cc.Screen.Print("    " + display.ColorProp("- flow:", env) + "\n")
	for _, flowStr := range flowStrs {
		cc.Screen.Print("        " + display.ColorFlow(flowStr, env) + "\n")
	}
	cc.Screen.Print("    " + display.ColorProp("- executable:", env) + "\n")
	cc.Screen.Print(fmt.Sprintf("        %s\n", filePath))
	if len(oldHelp) != 0 {
		cc.Screen.Print("    " + display.ColorProp("- old-help:", env) + "\n")
		cc.Screen.Print("       " + display.ColorHelp("'"+help+"'", env) + "\n")
	}
	return currCmdIdx, nil
}

func LoadFlows(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	if err := assertNotTailMode(flow, currCmdIdx); err != nil {
		return currCmdIdx, err
	}
	root := getFlowRoot(env, flow.Cmds[currCmdIdx])
	loadFlowsFromDir(flow, currCmdIdx, root, cc, env, root)
	return currCmdIdx, nil
}

func LoadFlowsFromDir(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, error) {

	path, err := tailModeCallArg(flow, currCmdIdx, argv, "path")
	if err != nil {
		return currCmdIdx, err
	}
	loadFlowsFromDir(flow, currCmdIdx, path, cc, env, path)
	display.PrintTipTitle(cc.Screen, env,
		"flows from '"+path+"' is loaded")
	return currCmdIdx, nil
}

func loadFlowsFromDir(
	flow *model.ParsedCmds,
	currCmdIdx int,
	root string,
	cc *model.Cli,
	env *model.Env,
	source string) error {

	if len(root) > 0 && root[len(root)-1] == filepath.Separator {
		root = root[:len(root)-1]
	}
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return model.NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("access flows dir '%s' failed: %v", root, err))
	}
	if !info.IsDir() {
		return model.NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("flows dir '%s' is not dir", root))
	}

	flowExt := env.GetRaw("strs.flow-ext")
	envPathSep := env.GetRaw("strs.env-path-sep")
	panicRecover := env.GetBool("sys.panic.recover")

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
		cmdPath := filepath.Base(path[0 : len(path)-len(flowExt)])
		cmdPaths := strings.Split(cmdPath, cc.Cmds.Strs.PathSep)
		if err := mod_meta.RegMod(cc, path, "", false, true, cmdPaths, flowExt,
			cc.Cmds.Strs.AbbrsSep, envPathSep, source, panicRecover); err != nil {
			if panicRecover {
				return nil
			}
			return err
		}
		return nil
	})
	return nil
}

func getCmdRealPath(
	flow *model.ParsedCmds,
	currCmdIdx int,
	cc *model.Cli,
	cmdPath string) (newCmdPath string, exists bool, err error) {

	if len(cmdPath) == 0 {
		return "", false, model.NewCmdError(flow.Cmds[currCmdIdx], "cmd path is empty")
	}

	var realSegs []string
	sep := cc.Cmds.Strs.PathSep
	cmdSegs := strings.Split(cmdPath, sep)
	currNode := cc.Cmds
	exists = true
	for i, cmdSeg := range cmdSegs {
		sub := currNode.GetSub(cmdSeg)
		if sub == nil {
			exists = false
			realSegs = append(realSegs, cmdSegs[i:]...)
			break
		}
		realSegs = append(realSegs, sub.Name())
		currNode = sub
	}

	return strings.Join(realSegs, sep), exists, nil
}

func getFlowCmdPath(
	flow *model.ParsedCmds,
	currCmdIdx int,
	getArgFromFlow bool,
	inDir string,
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	expectExists bool,
	argName string) (originCmd string, cmdPath string, filePath string, cmdExists bool) {

	var arg string
	var err error
	if !getArgFromFlow {
		arg = argv.GetRaw(argName)
		if len(arg) == 0 {
			return "", "", "", false
		}
	} else {
		arg, err = tailModeCallArg(flow, currCmdIdx, argv, argName)
		if err != nil {
			return "", "", "", false
		}
	}
	originCmd = arg

	cmdPath = normalizeCmdPath(arg,
		cc.Cmds.Strs.PathSep, cc.Cmds.Strs.PathAlterSeps)
	cmdPath, cmdExists, _ = getCmdRealPath(flow, currCmdIdx, cc, cmdPath)

	if len(cmdPath) == 0 {
		return "", "", "", false
	}

	if expectExists && !cmdExists {
		return "", "", "", false
	}

	var root string
	if len(inDir) == 0 {
		root = getFlowRoot(env, flow.Cmds[currCmdIdx])
	} else {
		if !dirExists(inDir) {
			return "", "", "", false
		}
		root = inDir
	}

	flowExt := env.GetRaw("strs.flow-ext")

	filePath = filepath.Join(root, cmdPath) + flowExt
	if !expectExists && fileExists(filePath) {
		if !env.GetBool("sys.confirm.ask") {
			return "", "", "", false
		} else {
			return
		}
	}
	if expectExists && !fileExists(filePath) {
		return "", "", "", false
	}
	return
}

func checkAndConfirmIfFlowHasParseError(screen model.Screen, flow *model.ParsedCmds, env *model.Env) bool {
	for _, cmd := range flow.Cmds {
		if cmd.ParseResult.Error == nil {
			continue
		}
		screen.Print(display.ColorTip("[confirm]", env) + " flow has parse error, " +
			"type " + display.ColorWarn("'y'", env) + " and press enter to force save:\n")
		utils.UserConfirm()
		break
	}
	return true
}

func getFlowRoot(env *model.Env, cmd model.ParsedCmd) string {
	root := env.GetRaw("sys.paths.flows")
	if len(root) == 0 {
		return ""
	}
	return root
}
