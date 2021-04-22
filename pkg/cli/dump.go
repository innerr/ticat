package cli

import (
	"fmt"
	"strings"
	"strconv"
)

func DumpCmdsEx(screen *Screen, env *Env, cmds []ParsedCmd, sep string) {
	if len(cmds) == 0 {
		return
	}
	screen.Println("[cmds:" + strconv.Itoa(len(cmds)) + "]")
	for i, cmd := range cmds {
		line := strings.Repeat(" ", 4) + "[cmd:" + strconv.Itoa(i) + "] "
		line += getCmdPath(cmd, sep, true)
		screen.Println(line)
		args := cmd.Args()
		if args == nil {
			continue
		}
		argv := cmd.GenEnv(env).GetArgv(cmd.Path(), sep, args)
		for j, k := range args.List() {
			defV := args.DefVal(k)
			v := argv[k].Raw
			line = strings.Repeat(" ", 8) + "[arg:" + strconv.Itoa(j) + "] " + k + " = " + v
			if defV != v {
				line += " (def:" + defV + ")"
			}
			screen.Println(line)
		}
	}
}

func DumpCmds(cc *Cli, cmds []ParsedCmd) {
	DumpCmdsEx(cc.Screen, cc.GlobalEnv, cmds, cc.Parser.CmdPathSep())
}

func DumpEnv(env *Env) {
	lines := dumpEnv(env, true, true, true, nil)
	for _, line := range lines {
		fmt.Println(line)
	}
}
