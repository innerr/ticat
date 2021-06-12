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
	} else {
		PrintTipTitle(screen, env,
			fmt.Sprintf("'%v' is not executable, but has commands on this branch:", name))
		dumpArgs := NewDumpCmdArgs().SetSkeleton()
		DumpCmds(last, screen, dumpArgs)
	}
}

func PrintError(cc *core.Cli, env *core.Env, err error) {
	switch err.(type) {
	case *core.CmdError:
		e := err.(*core.CmdError)
		sep := cc.Cmds.Strs.PathSep
		cmdName := strings.Join(e.Cmd.MatchedPath(), sep)
		printer := NewTipBoxPrinter(cc.Screen, env, true)
		printer.PrintWrap("[" + cmdName + "] failed: " + e.Error() + ".")
		printer.Prints("", "command detail:", "")
		dumpArgs := NewDumpFlowArgs().SetSimple()
		dumpFlowCmd(cc, printer, env, e.Cmd, dumpArgs, 0, 0)
		printer.Finish()
	default:
		PrintErrTitle(cc.Screen, env, err.Error())
	}
}

func PrintSepTitle(screen core.Screen, env *core.Env, msg string) {
	width := env.GetInt("display.width") - 3
	screen.Print(rpt("-", width-len(msg)) + "<[" + msg + "]\n")
}

// TODO: unused
func PrintDisplayBlockSep(screen core.Screen, env *core.Env, title string) {
	if env.GetBool("display.utf8") {
		PrintTipTitle(screen, env, title)
	} else {
		screen.Print(fmt.Sprintf("-------=<%s>=-------\n", title))
	}
}

func PrintPanicHeader(screen core.Screen, title string) {
	screen.Error("======================================\n\n")
	screen.Error(fmt.Sprintf("[ERR] %s:\n", title))
}

func PrintPanicFooter(screen core.Screen) {
	screen.Error("\n======================================\n\n")
}

func PrintPanic(screen core.Screen, title string, kvs []string) {
	PrintPanicHeader(screen, title)
	for i := 0; i+1 < len(kvs); i += 2 {
		screen.Error(fmt.Sprintf("    - %s:\n", kvs[i]))
		screen.Error(fmt.Sprintf("        %s\n", kvs[i+1]))
	}
	PrintPanicFooter(screen)
}

func PrintTolerableErrs(screen core.Screen, env *core.Env, errs *core.TolerableErrs) {
	for _, err := range errs.Uncatalogeds {
		PrintPanic(screen, err.Reason, []string{
			"source", err.Source,
			"error", err.Err.(error).Error(),
		})
	}

	// Conflicted error list:
	// CmdTreeErrSubCmdConflicted
	// CmdTreeErrSubAbbrConflicted
	// CmdTreeErrExecutableConflicted

	// TODO: add '<builtin>' to env strs
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
				PrintErrTitle(screen, env,
					err.Reason+", command conflicted with builtin's, from repo/dir:",
					"    - '"+err.Source+"'",
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
					"use command 'hub.disable' to disable one of them.",
				)
			} else {
				for _, err := range list {
					PrintErrTitle(screen, env,
						err.Reason+", command conflicted from repos/dirs:",
						"    - '"+oldSource+"'",
						"    - '"+newSource+"' (not loaded)",
						"detail:",
						"    - "+err.Err.(error).Error(),
						"",
						"use command 'hub.disable' to disable one of the repo/dir, or edit the command.",
					)
				}
			}
		}
	}
}
