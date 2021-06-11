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
		dumpFlowCmd(cc, printer, env, e.Cmd, 0, sep, 4,
			true, false, 0)
		printer.Finish()
	default:
		PrintErrTitle(cc.Screen, env, err.Error())
	}
}

func PrintErrTitle(screen core.Screen, env *core.Env, msgs ...string) {
	printTipTitle(screen, env, true, msgs...)
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
