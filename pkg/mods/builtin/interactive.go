package builtin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterh/liner"

	"github.com/innerr/ticat/pkg/cli/display"
	"github.com/innerr/ticat/pkg/core/model"
)

func InteractiveMode(cc *model.Cli, env *model.Env, exitStr string) error {
	cc = cc.CopyForInteract()
	sessionEnv := env.GetLayer(model.EnvLayerSession)
	sessionEnv.SetBool("sys.interact.inside", true)

	seqSep := env.GetRaw("strs.seq-sep")
	kvSep := env.GetRaw("strs.env-kv-sep")
	sep := cc.Cmds.Strs.PathSep
	selfName := env.GetRaw("strs.self-name")
	hiddenCompletion := env.GetBool("display.completion.hidden")
	abbrCompletion := env.GetBool("display.completion.abbr")
	shortcutCompletion := env.GetBool("display.completion.shortcut")

	if cc.TestingHook != nil {
		for {
			cc.Screen.Print(display.ColorExplain("(ctl-c to leave)\n", env))

			if env.GetBool("sys.interact.leaving") {
				break
			}

			line, hasInput := cc.TestingHook.OnInteractPrompt(selfName + "> ")
			if !hasInput {
				break
			}

			executorSafeExecute("(interact)", cc, env, nil, model.FlowStrToStrs(line)...)
		}

		sessionEnv.GetLayer(model.EnvLayerSession).Delete("sys.interact.inside")
		return nil
	}

	lineReader := liner.NewLiner()
	defer func() {
		if err := lineReader.Close(); err != nil {
			fmt.Errorf("[RunInteractive] close line reader failed: %v", err)
		}
	}()

	lineReader.SetCtrlCAborts(true)

	// TODO: this is a mess, but not in top priority, to solve this we need a better parser with original pos info
	lineReader.SetCompleter(func(line string) (res []string) {
		parsed := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, model.FlowStrToStrs(line)...)

		if hiddenCompletion || len(parsed.Cmds) > 1 {
			lineReader.SetTabCompletionStyle(liner.TabCircular)
		} else {
			lineReader.SetTabCompletionStyle(liner.TabPrints)
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			return cc.Cmds.GatherSubNames(abbrCompletion, shortcutCompletion)
		}
		tailBlanks := line[len(strings.TrimRight(line, " \t")):]
		last := strings.TrimLeft(fields[len(fields)-1], seqSep)

		hasTailSeqSep := len(last) != len(fields[len(fields)-1])
		if hasTailSeqSep {
			lineReader.SetTabCompletionStyle(liner.TabCircular)
		}

		prefix := line[0 : len(line)-len(last)-len(tailBlanks)]
		if len(last) == 0 {
			return
		}

		if last[len(last)-1:] == sep {
			parentPath := last[0 : len(last)-1]
			parent := cc.Cmds.GetSubByPath(parentPath, false)
			if parent == nil {
				return
			}
			for _, sub := range parent.GatherSubNames(abbrCompletion, shortcutCompletion) {
				res = append(res, prefix+parentPath+sep+sub)
			}
			return
		}

		// TODO: this is very ugly, should use parser here
		if len(tailBlanks) > 0 {
			var tailCmd *model.CmdTree
			var argsInUse []string
			for i := len(fields) - 1; i >= 0; i-- {
				if strings.Index(fields[i], seqSep) >= 0 {
					break
				}
				name := strings.Trim(fields[i], seqSep)
				if len(name) == 0 {
					continue
				}
				if strings.Index(fields[i], kvSep) > 0 {
					argsInUse = append(argsInUse, fields[i])
					continue
				}
				tailCmd = cc.Cmds.GetSubByPath(name, false)
			}
			if tailCmd != nil {
				lineReader.SetTabCompletionStyle(liner.TabCircular)
				args := tailCmd.Args()
				for _, arg := range args.Names() {
					inUse := false
					for _, argInUse := range argsInUse {
						if strings.HasPrefix(argInUse, arg) {
							inUse = true
							break
						}
					}
					if !inUse {
						res = append(res, prefix+last+tailBlanks+arg+kvSep)
					}
				}
			}
			return
		}

		tailCmd := cc.Cmds.GetSubByPath(last, false)
		if tailCmd != nil {
			if hiddenCompletion {
				// Double it to let user understand this command exists
				res = append(res, prefix+last)
				res = append(res, prefix+last)
			}
		}

		var parentPath []string
		parent := cc.Cmds
		brokePath := strings.Split(last, sep)
		if len(brokePath) > 1 {
			parentPath = brokePath[:len(brokePath)-1]
			parent = cc.Cmds.GetSub(parentPath...)
			if parent == nil {
				return
			}
		}
		brokeSub := brokePath[len(brokePath)-1]
		for _, sub := range parent.GatherSubNames(abbrCompletion, shortcutCompletion) {
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
		cc.Screen.Print(display.ColorExplain("(ctl-c to leave)\n", env))

		if env.GetBool("sys.interact.leaving") {
			break
		}
		line, err := lineReader.Prompt(selfName + "> ")
		if err == liner.ErrPromptAborted {
			break
		}
		if err != nil {
			return fmt.Errorf("[InteractMode] read from stdin failed: %v", err)
		}

		lineReader.AppendHistory(line)
		executorSafeExecute("(interact)", cc, env, nil, model.FlowStrToStrs(line)...)
		//cc.Screen.Print(display.ColorExplain("(ctl-c to leave)\n", env))
	}

	sessionEnv.GetLayer(model.EnvLayerSession).Delete("sys.interact.inside")

	file, err := os.Create(historyDir)
	if err != nil {
		return fmt.Errorf("[InteractMode] error writing history file: %v", err)
	}
	lineReader.WriteHistory(file)
	file.Close()
	return nil
}

func executorSafeExecute(caller string, cc *model.Cli, env *model.Env, masks []*model.ExecuteMask, input ...string) {
	env = env.GetLayer(model.EnvLayerSession)
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
