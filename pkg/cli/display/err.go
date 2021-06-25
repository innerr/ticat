package display

import (
	"fmt"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func PrintEmptyDirCmdHint(screen core.Screen, env *core.Env, cmd core.ParsedCmd) {
	sep := env.GetRaw("strs.cmd-path-sep")
	name := cmd.DisplayPath(sep, true)
	last := cmd.LastCmdNode()
	if !last.HasSub() {
		PrintTipTitle(screen, env,
			fmt.Sprintf("'%v' is not executable and has no commands on this branch.", name))
		return
	}
	PrintTipTitle(screen, env,
		fmt.Sprintf("'%v' is not executable, but has commands on this branch:", name))
	dumpArgs := NewDumpCmdArgs().SetSkeleton()
	DumpCmds(last, screen, env, dumpArgs)
}

func PrintError(cc *core.Cli, env *core.Env, err error) {
	switch err.(type) {
	case core.CmdMissedEnvValWhenRenderFlow:
		e := err.(core.CmdMissedEnvValWhenRenderFlow)
		PrintErrTitle(cc.Screen, env,
			e.Error()+" from repo/dir:",
			"    - '"+e.Source+"'",
			"command:",
			"    - '"+e.CmdPath+"'",
			"file:",
			"    - '"+e.MetaFilePath+"'",
			"missed-key:",
			"    - "+e.MissedKey)
	case *core.CmdError:
		e := err.(*core.CmdError)
		sep := cc.Cmds.Strs.PathSep
		cmdName := strings.Join(e.Cmd.MatchedPath(), sep)
		printer := NewTipBoxPrinter(cc.Screen, env, true)
		printer.PrintWrap("[" + cmdName + "] failed: " + e.Error() + ".")
		printer.Prints("", "command detail:", "")
		dumpArgs := NewDumpFlowArgs().SetSimple()
		metFlows := map[string]bool{}
		// TODO: use DumpCmds here
		dumpFlowCmd(cc, printer, env.Clone(), e.Cmd, dumpArgs, 0, 0, metFlows)
		printer.Finish()
	default:
		PrintErrTitle(cc.Screen, env, err.Error())
	}
}

func PrintSepTitle(screen core.Screen, env *core.Env, msg string) {
	width := env.GetInt("display.width") - 3
	screen.Print(rpt("-", width-len(msg)) + "<[" + msg + "]\n")
}

func PrintTolerableErrs(screen core.Screen, env *core.Env, errs *core.TolerableErrs) {
	for _, err := range errs.Uncatalogeds {
		PrintErrTitle(screen, env,
			err.Reason+", from repo/dir:",
			"    - '"+err.Source+"'",
			"file:",
			"    - '"+err.File+"'",
			"detail:",
			"    - "+err.Err.(error).Error())
	}

	// Conflicted error list:
	//
	// CmdTreeErrSubCmdConflicted
	// CmdTreeErrSubAbbrConflicted
	// CmdTreeErrExecutableConflicted

	sep := env.GetRaw("strs.cmd-path-sep")

	for newSource, list := range errs.ConflictedWithBuiltin {
		if len(list) > 1 {
			PrintErrTitle(screen, env,
				fmt.Sprintf("this repo/dir has too many (%v) conflicts with builtin commands:", len(list)),
				"",
				"    - '"+newSource+"'",
				"",
				"conflicted commands are not loaded.",
				"use command 'hub.disable' to disable the repo/dir, or edit these commands.",
			)
		} else {
			for _, err := range list {
				cmdPath := err.Err.(core.ErrConflicted).GetConflictedCmdPath()
				PrintErrTitle(screen, env,
					err.Reason+", command conflicted with builtin's, from repo/dir:",
					"    - '"+err.Source+"'",
					"command:",
					"    - "+strings.Join(cmdPath, sep),
					"detail:",
					"    - "+err.Err.(error).Error(),
					"",
					"use command 'hub.disable' to disable the repo/dir, or edit this command.")
			}
		}
	}

	for oldSource, conflicteds := range errs.Conflicteds {
		for newSource, list := range conflicteds {
			if len(list) > 1 {
				PrintErrTitle(screen, env,
					fmt.Sprintf("too many (%v) conflicts between these two repos/dirs:", len(list)),
					"",
					"    - '"+oldSource+"'",
					"    - '"+newSource+"' (conflicteds are not loaded)",
					"",
					"use command 'h.disable' to disable one of them.",
				)
			} else {
				for _, err := range list {
					cmdPath := err.Err.(core.ErrConflicted).GetConflictedCmdPath()
					PrintErrTitle(screen, env,
						err.Reason+", command conflicted from repos/dirs:",
						"    - '"+oldSource+"'",
						"    - '"+newSource+"' (not loaded)",
						"command:",
						"    - "+strings.Join(cmdPath, sep),
						"detail:",
						"    - "+err.Err.(error).Error(),
						"",
						"use command 'h.disable' to disable one of the repo/dir, or edit the command.",
					)
				}
			}
		}
	}
}
