package model

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/innerr/ticat/pkg/utils"
)

func SaveFlowToStr(flow *ParsedCmds, cmdPathSep string, trivialMark string, env *Env) (string, error) {
	w := bytes.NewBuffer(nil)
	err := SaveFlow(w, flow, cmdPathSep, trivialMark, env)
	return w.String(), err
}

func SaveFlow(w io.Writer, flow *ParsedCmds, cmdPathSep string, trivialMark string, env *Env) error {
	envPathSep := env.GetRaw("strs.env-path-sep")
	bracketLeft := env.GetRaw("strs.env-bracket-left")
	bracketRight := env.GetRaw("strs.env-bracket-right")
	envKeyValSep := env.GetRaw("strs.env-kv-sep")
	seqSep := env.GetRaw("strs.seq-sep")
	if len(envPathSep) == 0 || len(bracketLeft) == 0 || len(bracketRight) == 0 ||
		len(envKeyValSep) == 0 || len(seqSep) == 0 {
		// PANIC: Programming error - required env strings not found
		panic(fmt.Errorf("some predefined strs not found"))
	}

	for i, cmd := range flow.Cmds {
		if len(flow.Cmds) > 1 {
			if i == 0 {
				if flow.GlobalCmdIdx < 0 {
					if _, err := fmt.Fprint(w, seqSep+" "); err != nil {
						return fmt.Errorf("[SaveFlow] write failed: %w", err)
					}
				}
			} else {
				if _, err := fmt.Fprint(w, " "+seqSep+" "); err != nil {
					return fmt.Errorf("[SaveFlow] write failed: %w", err)
				}
			}
		}

		if cmd.ParseResult.Error != nil {
			if _, err := fmt.Fprint(w, strings.Join(cmd.ParseResult.Input, " ")); err != nil {
				return fmt.Errorf("[SaveFlow] write failed: %w", err)
			}
			continue
		}

		var path []string
		var prevSegHasNoCmd bool
		var cmdHasEnv bool

		for i := 0; i < cmd.TrivialLvl; i++ {
			if _, err := fmt.Fprint(w, trivialMark); err != nil {
				return fmt.Errorf("[SaveFlow] write failed: %w", err)
			}
		}

		for j, seg := range cmd.Segments {
			if len(cmd.Segments) > 1 && j != 0 && !prevSegHasNoCmd {
				if _, err := fmt.Fprint(w, cmdPathSep); err != nil {
					return fmt.Errorf("[SaveFlow] write failed: %w", err)
				}
			}
			if _, err := fmt.Fprint(w, seg.Matched.Name); err != nil {
				return fmt.Errorf("[SaveFlow] write failed: %w", err)
			}

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
	return nil
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
	if _, err := fmt.Fprintf(w, format, strings.Join(kvs, " ")); err != nil {
		// Runtime error: write failed - ignore since this is a display function
	}
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
