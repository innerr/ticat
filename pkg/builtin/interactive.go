package builtin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterh/liner"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func InteractiveMode(cc *core.Cli, env *core.Env, exitStr string) {
	cc.Screen.Print(display.ColorExplain("(ctl-c to leave)\n", env))

	cc = cc.CopyForInteract()
	sessionEnv := env.GetLayer(core.EnvLayerSession)
	sessionEnv.SetBool("sys.interact.inside", true)

	seqSep := env.GetRaw("strs.seq-sep")
	sep := cc.Cmds.Strs.PathSep
	selfName := env.GetRaw("strs.self-name")

	lineReader := liner.NewLiner()
	defer lineReader.Close()
	lineReader.SetCtrlCAborts(true)

	hiddenCompletion := env.GetBool("display.completion.hidden")

	if !hiddenCompletion {
		lineReader.SetTabCompletionStyle(liner.TabPrints)
	}

	lineReader.SetCompleter(func(line string) (res []string) {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			return cc.Cmds.GatherSubNames()
		}
		field := strings.TrimLeft(fields[len(fields)-1], seqSep)
		prefix := line[0 : len(line)-len(field)]
		if len(field) == 0 {
			return
		}

		if field[len(field)-1:] == sep {
			parentPath := field[0 : len(field)-1]
			parent := cc.Cmds.GetSubByPath(parentPath, false)
			if parent == nil {
				return
			}
			for _, sub := range parent.GatherSubNames() {
				res = append(res, prefix+parentPath+sep+sub)
			}
			return
		}

		if hiddenCompletion {
			// Double it to let user understand this command exists
			if cc.Cmds.GetSubByPath(field, false) != nil {
				res = append(res, prefix+field)
				res = append(res, prefix+field)
			}
		}

		var parentPath []string
		parent := cc.Cmds
		brokePath := strings.Split(field, sep)
		if len(brokePath) > 1 {
			parentPath = brokePath[:len(brokePath)-1]
			parent = cc.Cmds.GetSub(parentPath...)
			if parent == nil {
				return
			}
		}
		brokeSub := brokePath[len(brokePath)-1]
		for _, sub := range parent.GatherSubNames() {
			if strings.HasPrefix(sub, brokeSub) {
				res = append(res, prefix+strings.Join(append(parentPath, sub), sep))
			}
		}
		return
	})

	historyDir := filepath.Join(os.TempDir(), ".ticat_interact_mode_cmds_history")
	if file, err := os.Open(historyDir); err == nil {
		lineReader.ReadHistory(file)
		file.Close()
	}

	for {
		if env.GetBool("sys.interact.leaving") {
			break
		}
		line, err := lineReader.Prompt(selfName + "> ")
		if err == liner.ErrPromptAborted {
			break
		}
		if err != nil {
			panic(fmt.Errorf("[InteractMode] read from stdin failed: %v", err))
		}

		lineReader.AppendHistory(line)
		executorSafeExecute("(interact)", cc, env, nil, core.FlowStrToStrs(line)...)
		//cc.Screen.Print(display.ColorExplain("(ctl-c to leave)\n", env))
	}

	sessionEnv.GetLayer(core.EnvLayerSession).Delete("sys.interact.inside")

	file, err := os.Create(historyDir)
	if err != nil {
		panic(fmt.Errorf("[InteractMode] error writing history file: %v", err))
	}
	lineReader.WriteHistory(file)
	file.Close()
}

func executorSafeExecute(caller string, cc *core.Cli, env *core.Env, masks []*core.ExecuteMask, input ...string) {
	env = env.GetLayer(core.EnvLayerSession)
	stackDepth := env.GetRaw("sys.stack-depth")
	stack := env.GetRaw("sys.stack")
	displayOne := env.GetRaw("display.one-cmd")
	env.SetBool("display.one-cmd", true)

	defer func() {
		env.Set("sys.stack-depth", stackDepth)
		env.Set("sys.stack", stack)
		env.Set("display.one-cmd", displayOne)

		if !env.GetBool("sys.panic.recover") {
			return
		}
		if r := recover(); r != nil {
			display.PrintError(cc, env, r.(error))
		}
	}()

	cc.Executor.Execute(caller, false, cc, env, masks, input...)
}
