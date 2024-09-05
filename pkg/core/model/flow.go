package model

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/innerr/ticat/pkg/utils"
)

func SaveFlowToStr(flow *ParsedCmds, cmdPathSep string, trivialMark string, env *Env) string {
	w := bytes.NewBuffer(nil)
	SaveFlow(w, flow, cmdPathSep, trivialMark, env)
	return w.String()
}

func SaveFlow(w io.Writer, flow *ParsedCmds, cmdPathSep string, trivialMark string, env *Env) {
	envPathSep := env.GetRaw("strs.env-path-sep")
	bracketLeft := env.GetRaw("strs.env-bracket-left")
	bracketRight := env.GetRaw("strs.env-bracket-right")
	envKeyValSep := env.GetRaw("strs.env-kv-sep")
	seqSep := env.GetRaw("strs.seq-sep")
	if len(envPathSep) == 0 || len(bracketLeft) == 0 || len(bracketRight) == 0 ||
		len(envKeyValSep) == 0 || len(seqSep) == 0 {
		panic(fmt.Errorf("some predefined strs not found"))
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
		var prevSegHasNoCmd bool
		var cmdHasEnv bool

		for i := 0; i < cmd.TrivialLvl; i++ {
			fmt.Fprint(w, trivialMark)
		}

		for j, seg := range cmd.Segments {
			if len(cmd.Segments) > 1 && j != 0 && !prevSegHasNoCmd {
				fmt.Fprint(w, cmdPathSep)
			}
			fmt.Fprint(w, seg.Matched.Name)

			if seg.Matched.Cmd != nil {
				path = append(path, seg.Matched.Cmd.Name())
			} else if len(seg.Matched.Name) != 0 {
				path = append(path, seg.Matched.Name)
			}

			prevSegHasNoCmd = (seg.Matched.Cmd == nil)

			savedEnv := false
			useArgsFmt := (j == len(cmd.Segments)-1)
			savedEnv = SaveFlowEnv(w, seg.Env, path, envPathSep, seqSep,
				bracketLeft, bracketRight, envKeyValSep, useArgsFmt)
			cmdHasEnv = cmdHasEnv || savedEnv
		}
	}
}

func SaveFlowEnv(
	w io.Writer,
	env ParsedEnv,
	prefixPath []string,
	pathSep string,
	seqSep string,
	bracketLeft string,
	bracketRight string,
	envKeyValSep string,
	useArgsFmt bool) bool {

	if len(env) == 0 {
		return false
	}

	isAllArgs := true
	var keys []string
	for k, v := range env {
		if !v.IsArg {
			isAllArgs = false
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	prefix := strings.Join(prefixPath, pathSep) + pathSep

	var kvs []string
	for _, k := range keys {
		v := env[k]
		if strings.HasPrefix(k, prefix) && len(k) != len(prefix) {
			k = strings.Join(v.MatchedPath[len(prefixPath):], pathSep)
		}
		val := normalizeEnvVal(v.Val, seqSep)
		kv := fmt.Sprintf("%v%s%v", k, envKeyValSep, utils.QuoteStrIfHasSpace(val))
		kvs = append(kvs, kv)
	}

	format := bracketLeft + "%s" + bracketRight
	if isAllArgs && useArgsFmt {
		format = " %s"
	}
	fmt.Fprintf(w, format, strings.Join(kvs, " "))
	return true
}

func normalizeEnvVal(v string, seqSep string) string {
	i := 0
	for i < len(v) {
		origin := i
		i = strings.Index(v[i:], seqSep)
		if i < 0 {
			break
		}
		i += origin
		if i > len(BackSlash) && v[i-len(BackSlash):i] == BackSlash {
			i += len(seqSep)
			continue
		}
		// NOTE: shellwords in 'cmd.go#FlowStrToStrs' will eat the single '\', so double '\\' is needed
		v = v[0:i] + BackSlash + BackSlash + seqSep + v[i+len(seqSep):]
		i += len(BackSlash)*2 + len(seqSep)
	}
	return v
}

// TODO: put '\' to env
// TODO: too many related logic: here, parser, shellwords
const BackSlash = "\\"
