package model

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// TODO: It's slow and ugly, but it should works fine

func RenderTemplateStrLines(
	in []string,
	targetName string,
	cmd *Cmd,
	argv ArgVals,
	env *Env,
	allowError bool) (out []string, fullyRendered bool) {

	fullyRendered = true
	for _, line := range in {
		// TODO: put # into env.strs
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		rendereds, lineFullyRendered := renderTemplateStr(line, targetName, cmd, argv, env, allowError)
		for _, rendered := range rendereds {
			out = append(out, rendered)
		}
		fullyRendered = fullyRendered && lineFullyRendered
	}
	return
}

func renderTemplateStr(
	in string,
	targetName string,
	cmd *Cmd,
	argv ArgVals,
	env *Env,
	allowError bool) (rendered []string, fullyRendered bool) {

	fullyRendered = true

	lines, hasError := tryRenderMultiplyTemplate(in, targetName, cmd, argv, env, allowError)
	for _, line := range lines {
		if !hasError {
			line, hasError = tryRenderTemplate(line, targetName, cmd, argv, env, allowError)
		}
		fullyRendered = fullyRendered && !hasError
		rendered = append(rendered, line)
	}
	return
}

// only support one multiply definition in a line
func tryRenderMultiplyTemplate(
	in string,
	targetName string,
	cmd *Cmd,
	argv ArgVals,
	env *Env,
	allowError bool) (out []string, hasError bool) {

	templBracketLeft := cmd.owner.Strs.FlowTemplateBracketLeft
	templBracketRight := cmd.owner.Strs.FlowTemplateBracketRight
	templMultiplyMark := cmd.owner.Strs.FlowTemplateMultiplyMark
	listSep := cmd.owner.Strs.ListSep

	out = []string{in}
	mm := templMultiplyMark + templMultiplyMark
	ml := templBracketLeft + templMultiplyMark
	mr := templMultiplyMark + templBracketRight

	valBegin := strings.Index(in, ml)
	if valBegin <= 0 {
		return
	}

	valEnd := strings.Index(in[valBegin:], mr)
	if valEnd <= 0 {
		return
	}
	valEnd += valBegin

	tempBegin := strings.Index(in[:valBegin], mm)
	if tempBegin < 0 {
		return
	}

	tempEnd := strings.Index(in[valEnd:], mm)
	if tempEnd < 0 {
		return
	}
	tempEnd += valEnd

	key := strings.TrimSpace(in[valBegin+len(ml) : valEnd])

	var valStr string
	val, ok := env.GetEx(key)
	valStr = val.Raw
	if !ok {
		val, inArg := argv[key]
		valStr = val.Raw
		ok = inArg && len(valStr) != 0
	}
	if !ok {
		hasError = true
		if allowError {
			return
		}
		templateRenderPanic(in, cmd, targetName, key, true)
	}

	out = nil
	if tempBegin != 0 {
		out = append(out, strings.TrimSpace(in[:tempBegin]))
	}

	head := in[tempBegin+len(mm) : valBegin]
	tail := in[valEnd+len(mr) : tempEnd]

	vals := strings.Split(valStr, listSep)
	for _, val := range vals {
		val = strings.TrimSpace(val)
		out = append(out, strings.TrimSpace(head+val+tail))
	}

	if tempEnd+len(mm) != len(in) {
		out = append(out, strings.TrimSpace(in[tempEnd+len(mm):]))
	}
	return
}

func tryRenderTemplate(
	in string,
	targetName string,
	cmd *Cmd,
	argv ArgVals,
	env *Env,
	allowError bool) (out string, hasError bool) {

	templBracketLeft := cmd.owner.Strs.FlowTemplateBracketLeft
	templBracketRight := cmd.owner.Strs.FlowTemplateBracketRight

	findPos := 0
	for {
		str := in[findPos:]
		i := strings.Index(str, templBracketLeft)
		if i < 0 {
			break
		}
		tail := str[i+len(templBracketLeft):]
		j := strings.Index(tail, templBracketRight)
		if j < 0 {
			break
		}
		key := tail[0:j]
		var valStr string
		val, ok := env.GetEx(key)
		valStr = val.Raw
		if !ok {
			val, inArg := argv[key]
			valStr = val.Raw
			ok = inArg && len(valStr) != 0
		}
		if !ok {
			valStr, ok = tryRenderSystemVars(key)
		}
		if !ok {
			hasError = true
			if allowError {
				findPos += j + len(templBracketRight)
				continue
			}
			templateRenderPanic(in, cmd, targetName, key, false)
		}
		in = in[:findPos] + str[0:i] + valStr + tail[j+len(templBracketRight):]
	}
	return in, hasError
}

func tryRenderSystemVars(key string) (val string, ok bool) {
	rand.Seed(time.Now().UnixNano())
	val, ok = map[string]string{
		"RANDOM": fmt.Sprintf("%d", rand.Int()),
	}[key]
	return
}

func templateRenderPanic(in string, cmd *Cmd, targetName string, key string, isMultiply bool) {
	multiply := ""
	if isMultiply {
		multiply = "multiply "
	}

	findArgIdx := func(name string) int {
		idx := -1
		if len(name) == 0 {
			return idx
		}
		for i, it := range cmd.args.Names() {
			if it == name {
				idx = i
			}
		}
		return idx
	}

	if cmd.args.Has(key) {
		err := &CmdMissedArgValWhenRenderFlow{
			"render " + targetName + " template " + multiply + "failed, arg value missed.",
			cmd.owner.DisplayPath(),
			cmd.metaFilePath,
			cmd.owner.Source(),
			in,
			cmd,
			key,
			findArgIdx(key),
		}
		panic(err)
	} else {
		argName := cmd.arg2env.GetArgName(cmd, key, true)
		argIdx := findArgIdx(argName)
		err := &CmdMissedEnvValWhenRenderFlow{
			"render " + targetName + " template " + multiply + "failed, env value missed.",
			cmd.owner.DisplayPath(),
			cmd.metaFilePath,
			cmd.owner.Source(),
			key,
			in,
			cmd,
			argName,
			argIdx,
		}
		panic(err)
	}
}
