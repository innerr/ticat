package display

import (
	"fmt"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func DumpEnvOpsCheckResult(
	screen core.Screen,
	cmds []core.ParsedCmd,
	env *core.Env,
	result []core.EnvOpsCheckResult,
	sep string) {

	if len(result) == 0 {
		return
	}

	fatals, risks, isArg2EnvCanFixAllFatals := AggEnvOpsCheckResult(result)

	if len(fatals.Result) != 0 {
		helpStr := []interface{}{
			"this flow has 'read before write' on env keys, so it can't execute.",
			"",
			"search which commands write these keys and concate them in front of the flow:",
			"",
			SuggestFindProvider(env),
			"",
			//"some configuring-flows will provide a batch env keys by calling providing commands,",
			//"use these two tags to find them:",
			//"",
			//SuggestFindConfigFlows(env),
			//"",
			"or provide keys by putting '{key=value}' in front of the flow.",
		}
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
		screen.Print(msg + "\n")
	}
	prti := func(msg string, indent int) {
		screen.Print(strings.Repeat(" ", indent) + msg + "\n")
	}

	prefix := ColorProp("- ", env)

	if len(risks.Result) != 0 && len(fatals.Result) == 0 {
		for i, it := range risks.Result {
			if i != 0 {
				screen.Print("\n")
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
				screen.Print("\n")
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
		strs = append(strs, core.EnvOpStr(op))
	}
	return strings.Join(strs, sep)
}

type envOpsCheckResult struct {
	Key                string
	Cmds               []string
	FirstArg2Env       *core.ParsedCmd
	MayWriteCmdsBefore []core.MayWriteCmd
	ReadMayWrite       bool
	MayReadMayWrite    bool
	MayReadNotExist    bool
	ReadNotExist       bool
	CmdMap             map[string]bool
}

func AggEnvOpsCheckResult(result []core.EnvOpsCheckResult) (fatals *EnvOpsCheckResultAgg,
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

func (self *EnvOpsCheckResultAgg) Append(res core.EnvOpsCheckResult) {
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

func getMissedMapperArgInfo(env *core.Env, cic *core.Cmd, key string) string {
	arg2env := cic.GetArg2Env()
	argName := arg2env.GetArgName(key)
	return getArgInfoLine(env, cic, argName)
}

func getArgInfoLine(env *core.Env, cic *core.Cmd, argName string) string {
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
