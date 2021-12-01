package core

import (
	"fmt"
	"io"
	"strings"

	"github.com/pingcap/ticat/pkg/utils"
)

func SaveFlow(w io.Writer, flow *ParsedCmds, currCmdIdx int, cmdPathSep string, trivialMark string, env *Env) {
	envPathSep := env.GetRaw("strs.env-path-sep")
	bracketLeft := env.GetRaw("strs.env-bracket-left")
	bracketRight := env.GetRaw("strs.env-bracket-right")
	envKeyValSep := env.GetRaw("strs.env-kv-sep")
	seqSep := env.GetRaw("strs.seq-sep")
	if len(envPathSep) == 0 || len(bracketLeft) == 0 || len(bracketRight) == 0 ||
		len(envKeyValSep) == 0 || len(seqSep) == 0 {
		panic(NewCmdError(flow.Cmds[currCmdIdx], "some predefined strs not found"))
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

		if cmd.ParseResult.Error != nil {
			fmt.Fprint(w, strings.Join(cmd.ParseResult.Input, " "))
			continue
		}

		var path []string
		var lastSegHasNoCmd bool
		var cmdHasEnv bool

		for i := 0; i < cmd.TrivialLvl; i++ {
			fmt.Fprint(w, trivialMark)
		}

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
			cmdHasEnv = cmdHasEnv || SaveFlowEnv(w, seg.Env, path, envPathSep,
				bracketLeft, bracketRight, envKeyValSep,
				!cmdHasEnv && j == len(cmd.Segments)-1)
		}
	}
}

func SaveFlowEnv(
	w io.Writer,
	env ParsedEnv,
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
		kvs = append(kvs, fmt.Sprintf("%v%s%v", k, envKeyValSep, utils.QuoteStrIfHasSpace(v.Val)))
	}

	format := bracketLeft + "%s" + bracketRight
	if isAllArgs && useArgsFmt {
		format = " %s"
	}
	fmt.Fprintf(w, format, strings.Join(kvs, " "))
	return true
}
