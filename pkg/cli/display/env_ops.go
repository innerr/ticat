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

	fatals := newEnvOpsCheckResultAgg()
	risks := newEnvOpsCheckResultAgg()
	for _, it := range result {
		if it.ReadNotExist {
			fatals.Append(it)
		} else {
			risks.Append(it)
		}
	}

	var cic *core.Cmd
	var arg2env *core.Arg2Env
	var matchedCmdPath string
	isArg2EnvCanFixAllFatals := (len(cmds) == 1 && len(fatals.result) != 0)
	if isArg2EnvCanFixAllFatals {
		matchedCmdPath = cmds[0].DisplayPath(sep, false)
		cmd := cmds[0].Last().Matched.Cmd
		if cmd == nil || cmd.Cmd() == nil {
			isArg2EnvCanFixAllFatals = false
		} else {
			cic = cmd.Cmd()
			arg2env = cic.GetArg2Env()
			for _, it := range fatals.result {
				if !arg2env.Has(it.Key) {
					isArg2EnvCanFixAllFatals = false
					break
				}
			}
		}
	}

	if !env.GetBool("display.flow.simplified") {
		if len(fatals.result) != 0 {
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
				"",
			}
			if isArg2EnvCanFixAllFatals {
				helpStr = append(helpStr,
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
	} else {
		screen.Print(fmt.Sprintf("-------=<%s>=-------\n\n", "unsatisfied env read"))
	}

	prt0 := func(msg string) {
		screen.Print(msg + "\n")
	}
	prti := func(msg string, indent int) {
		screen.Print(strings.Repeat(" ", indent) + msg + "\n")
	}

	if len(risks.result) != 0 && len(fatals.result) == 0 {
		for i, it := range risks.result {
			if i != 0 {
				screen.Print("\n")
			}
			prt0("<risk>  '" + it.Key + "'")
			if it.MayReadNotExist || it.MayReadMayWrite {
				prti("- may-read by:", 7)
			} else if it.ReadMayWrite {
				prti("- read by:", 7)
			}
			for _, cmd := range it.Cmds {
				prti("["+cmd+"]", 12)
			}
			if len(it.MayWriteCmdsBefore) != 0 && (it.ReadMayWrite || it.MayReadMayWrite) {
				prti("- but may not provided by:", 7)
				for _, cmd := range it.MayWriteCmdsBefore {
					prti("["+cmd.Matched.DisplayPath(sep, true)+"]", 12)
				}
			} else {
				if it.MayReadNotExist {
					prti("- but not provided.", 7)
				} else {
					prti("- but may not provided.", 7)
				}
			}
		}
	}

	abbrsSep := env.GetRaw("strs.abbrs-sep")

	if len(fatals.result) != 0 {
		for i, it := range fatals.result {
			if i != 0 {
				screen.Print("\n")
			}
			prt0("<FATAL> '" + it.Key + "'")
			prti("- read by:", 7)
			for _, cmd := range it.Cmds {
				prti("["+cmd+"]", 12)
			}
			prti("- but not provided.", 7)
			if arg2env != nil && arg2env.Has(it.Key) {
				prti("- an arg of ["+matchedCmdPath+"] is mapped to this key, pass it to solve the error:", 7)
				argName := arg2env.GetArgName(it.Key)
				argInfo := "'" + argName + "'"
				args := cic.Args()
				argInfo = fmt.Sprintf("%s #%d", argInfo, args.Index(argName))
				abbrs := args.Abbrs(argName)
				if len(abbrs) > 1 {
					abbrTerm := "abbr"
					if len(abbrs) > 2 {
						abbrTerm = "abbrs"
					}
					argInfo += " (" + abbrTerm + ": " + strings.Join(abbrs[1:], abbrsSep) + ")"
				}
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
	Cmds               []string
	Key                string
	MayWriteCmdsBefore []core.MayWriteCmd
	ReadMayWrite       bool
	MayReadMayWrite    bool
	MayReadNotExist    bool
	ReadNotExist       bool
	CmdMap             map[string]bool
}

type envOpsCheckResultAgg struct {
	result []envOpsCheckResult
	revIdx map[string]int
}

func newEnvOpsCheckResultAgg() *envOpsCheckResultAgg {
	return &envOpsCheckResultAgg{nil, map[string]int{}}
}

func (self *envOpsCheckResultAgg) Append(res core.EnvOpsCheckResult) {
	hashKey := fmt.Sprintf("%s_%v_%v_%v_%v", res.Key, res.ReadMayWrite,
		res.MayReadMayWrite, res.MayReadNotExist, res.ReadNotExist)
	idx, ok := self.revIdx[hashKey]
	if !ok {
		idx = len(self.result)
		self.result = append(self.result, envOpsCheckResult{
			[]string{res.CmdDisplayPath},
			res.Key,
			res.MayWriteCmdsBefore,
			res.ReadMayWrite,
			res.MayReadMayWrite,
			res.MayReadNotExist,
			res.ReadNotExist,
			map[string]bool{res.CmdDisplayPath: true},
		})
		self.revIdx[hashKey] = idx
	} else {
		old := self.result[idx]
		if !old.CmdMap[res.CmdDisplayPath] {
			old.Cmds = append(old.Cmds, res.CmdDisplayPath)
			old.CmdMap[res.CmdDisplayPath] = true
			// Discard res.MayWriteCmdsBefore, it's not important
			self.result[idx] = old
		}
	}
}
