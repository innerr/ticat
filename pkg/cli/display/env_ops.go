package display

import (
	"fmt"
	"strings"

	"github.com/innerr/ticat/pkg/core/model"
)

func DumpEnvOpsCheckResult(
	screen model.Screen,
	cmds []model.ParsedCmd,
	env *model.Env,
	result []model.EnvOpsCheckResult,
	sep string,
	cmdTree *model.CmdTree) {

	if len(result) == 0 {
		return
	}

	fatals, risks, isArg2EnvCanFixAllFatals := AggEnvOpsCheckResult(result)

	if len(fatals.Result) != 0 {
		helpStr := []interface{}{
			"this flow has 'read before write' on env keys, so it can't execute.",
			"",
		}

		if cmdTree != nil {
			suggestions := findCmdsForFatalKeys(cmdTree, env, fatals.Result)
			if len(suggestions) > 0 {
				helpStr = append(helpStr,
					"commands which can provide these keys:",
					"")
				helpStr = append(helpStr, suggestions)
				helpStr = append(helpStr, "")
			}
		}

		helpStr = append(helpStr,
			"search which commands write these keys:",
			"",
			SuggestFindProvider(env),
			"",
			"or provide keys by putting '{key=value}' in front of the flow.",
		)
		if isArg2EnvCanFixAllFatals {
			helpStr = append(helpStr,
				"",
				"pass args properly to commands could solve all errors.")
		}
		PrintErrTitle(screen, env, helpStr...)
	} else {
		PrintTipTitle(screen, env,
			"this flow has 'read before write' risks on env keys.",
			"",
			"risks are caused by 'may-read' or 'may-write' on env keys,",
			"normally modules declair these uncertain behaviors will handle them, don't worry too much.")
	}

	prt0 := func(msg string) {
		_ = screen.Print(msg + "\n")
	}
	prti := func(msg string, indent int) {
		_ = screen.Print(strings.Repeat(" ", indent) + msg + "\n")
	}

	prefix := ColorProp("- ", env)

	if len(risks.Result) != 0 && len(fatals.Result) == 0 {
		for i, it := range risks.Result {
			if i != 0 {
				_ = screen.Print("\n")
			}
			prt0(ColorWarn("<risk>", env) + "  " + ColorKey("'"+it.Key+"'", env))
			if it.MayReadNotExist || it.MayReadMayWrite {
				prti(prefix+"may-read by:", 7)
			} else if it.ReadMayWrite {
				prti(prefix+"read by:", 7)
			}
			for _, cmd := range it.Cmds {
				prti(ColorCmd("["+cmd+"]", env), 12)
			}
			if len(it.MayWriteCmdsBefore) != 0 && (it.ReadMayWrite || it.MayReadMayWrite) {
				prti(prefix+" but may not provided by:", 7)
				for _, cmd := range it.MayWriteCmdsBefore {
					prti(ColorCmd("["+cmd.Matched.DisplayPath(sep, true)+"]", env), 12)
				}
			} else {
				if it.MayReadNotExist {
					prti(prefix+"but not provided.", 7)
				} else {
					prti(prefix+"but may not provided.", 7)
				}
			}
		}
	}

	if len(fatals.Result) != 0 {
		for i, it := range fatals.Result {
			if i != 0 {
				_ = screen.Print("\n")
			}
			prt0(ColorError("<FATAL>", env) + ColorKey(" '"+it.Key+"'", env))
			prti(prefix+"read by:", 7)
			for _, cmd := range it.Cmds {
				prti(ColorCmd("["+cmd+"]", env), 12)
			}
			prti(prefix+"but not provided.", 7)

			if it.FirstArg2Env != nil {
				matched := it.FirstArg2Env
				matchedCmdPath := matched.DisplayPath(sep, false)
				prti(prefix+"an arg of "+ColorCmd("["+matchedCmdPath+"]", env)+
					" is mapped to this key, pass it to solve the error:", 7)
				cic := matched.LastCmd()
				argInfo := getMissedMapperArgInfo(env, cic, it.Key)
				prti(argInfo, 12)
			}
		}

		// TODO: hint ?
	}
}

func dumpEnvOps(ops []uint, sep string) (str string) {
	var strs []string
	for _, op := range ops {
		strs = append(strs, model.EnvOpStr(op))
	}
	return strings.Join(strs, sep)
}

type envOpsCheckResult struct {
	Key                string
	Cmds               []string
	FirstArg2Env       *model.ParsedCmd
	MayWriteCmdsBefore []model.MayWriteCmd
	ReadMayWrite       bool
	MayReadMayWrite    bool
	MayReadNotExist    bool
	ReadNotExist       bool
	CmdMap             map[string]bool
}

func AggEnvOpsCheckResult(result []model.EnvOpsCheckResult) (fatals *EnvOpsCheckResultAgg,
	risks *EnvOpsCheckResultAgg, isArg2EnvCanFixAllFatals bool) {

	fatals = newEnvOpsCheckResultAgg()
	risks = newEnvOpsCheckResultAgg()
	isArg2EnvCanFixAllFatals = true

	for _, it := range result {
		if it.ReadNotExist {
			fatals.Append(it)
			if it.FirstArg2Env == nil {
				isArg2EnvCanFixAllFatals = false
			}
		} else {
			risks.Append(it)
		}
	}
	return
}

type EnvOpsCheckResultAgg struct {
	Result []envOpsCheckResult
	revIdx map[string]int
}

func newEnvOpsCheckResultAgg() *EnvOpsCheckResultAgg {
	return &EnvOpsCheckResultAgg{nil, map[string]int{}}
}

func (self *EnvOpsCheckResultAgg) Append(res model.EnvOpsCheckResult) {
	hashKey := fmt.Sprintf("%s_%v_%v_%v_%v", res.Key, res.ReadMayWrite,
		res.MayReadMayWrite, res.MayReadNotExist, res.ReadNotExist)
	idx, ok := self.revIdx[hashKey]
	if !ok {
		idx = len(self.Result)
		self.Result = append(self.Result, envOpsCheckResult{
			res.Key,
			[]string{res.CmdDisplayPath},
			res.FirstArg2Env,
			res.MayWriteCmdsBefore,
			res.ReadMayWrite,
			res.MayReadMayWrite,
			res.MayReadNotExist,
			res.ReadNotExist,
			map[string]bool{res.CmdDisplayPath: true},
		})
		self.revIdx[hashKey] = idx
	} else {
		old := self.Result[idx]
		if !old.CmdMap[res.CmdDisplayPath] {
			old.Cmds = append(old.Cmds, res.CmdDisplayPath)
			old.CmdMap[res.CmdDisplayPath] = true
			// Discard res.MayWriteCmdsBefore, it's not important
			self.Result[idx] = old
		}
	}
}

func getMissedMapperArgInfo(env *model.Env, cic *model.Cmd, key string) string {
	arg2env := cic.GetArg2Env()
	argName := arg2env.GetArgName(cic, key, true)
	return getArgInfoLine(env, cic, argName)
}

func getArgInfoLine(env *model.Env, cic *model.Cmd, argName string) string {
	argInfo := "'" + argName + "'"
	args := cic.Args()
	argInfo = ColorArg(argInfo, env) + " " + ColorSymbol(fmt.Sprintf("#%d", args.Index(argName)), env)
	abbrs := args.Abbrs(argName)
	if len(abbrs) > 1 {
		abbrTerm := "abbr"
		if len(abbrs) > 2 {
			abbrTerm = "abbrs"
		}
		abbrsSep := env.GetRaw("strs.abbrs-sep")
		argInfo += ColorArg(" ("+abbrTerm+": "+strings.Join(abbrs[1:], abbrsSep)+")", env)
	}
	return argInfo
}

func findCmdsForFatalKeys(cmdTree *model.CmdTree, env *model.Env, fatals []envOpsCheckResult) []string {
	selfName := env.GetRaw("strs.self-name")
	maxSuggestions := env.GetInt("display.env.suggest.max-cmds")
	if maxSuggestions == 0 {
		maxSuggestions = 3
	}

	failedCmds := make(map[string]bool)
	for _, fatal := range fatals {
		for _, cmd := range fatal.Cmds {
			failedCmds[cmd] = true
		}
	}

	var suggestions []string
	seenCmds := make(map[string]bool)

	for _, fatal := range fatals {
		dumpArgs := NewDumpCmdArgs().SetSkeleton().SetMatchWriteKey(fatal.Key)
		cacheScreen := NewCacheScreen()
		searchEnv := env.Clone()
		searchEnv.SetBool("display.color", false)
		DumpCmds(cmdTree, cacheScreen, searchEnv, dumpArgs)

		lines := cacheScreen.Lines()

		var cmdLines []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
				cmdLines = append(cmdLines, line)
			}
		}

		if len(cmdLines) == 0 {
			continue
		}

		suggestions = append(suggestions, "  "+ColorKey(fatal.Key, env)+":")
		count := 0
		for _, line := range cmdLines {
			cmdPath := strings.Trim(line, "[]")
			cmdPath = strings.TrimSpace(cmdPath)
			if len(cmdPath) == 0 {
				continue
			}
			if seenCmds[cmdPath] {
				continue
			}
			seenCmds[cmdPath] = true
			if failedCmds[cmdPath] {
				continue
			}

			count++
			if count > maxSuggestions {
				break
			}

			suggestions = append(suggestions, "    "+ColorCmd("["+cmdPath+"]", env)+" -> "+
				ColorCmd(selfName+" "+cmdPath+" : ...", env))
		}

		if len(cmdLines) > maxSuggestions {
			suggestions = append(suggestions, "    "+ColorExplain(
				fmt.Sprintf("  ... and %d more (use '%s env.who-write %s' to see all)",
					len(cmdLines)-maxSuggestions, selfName, fatal.Key), env))
		}
	}

	return suggestions
}
