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
	selfName := env.GetRaw("strs.self-name")

	lineReader := liner.NewLiner()
	defer lineReader.Close()
	lineReader.SetCtrlCAborts(true)

	names := cc.Cmds.GatherNames()
	lineReader.SetCompleter(func(line string) (res []string) {
		fields := strings.Fields(line)
		field := strings.TrimLeft(fields[len(fields)-1], seqSep)
		prefix := line[0 : len(line)-len(field)]
		for _, name := range names {
			if strings.HasPrefix(name, field) {
				res = append(res, prefix+name)
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

	defer func() {
		env.Set("sys.stack-depth", stackDepth)
		env.Set("sys.stack", stack)

		if !env.GetBool("sys.panic.recover") {
			return
		}
		if r := recover(); r != nil {
			display.PrintError(cc, env, r.(error))
		}
	}()

	cc.Executor.Execute(caller, false, cc, env, masks, input...)
}
