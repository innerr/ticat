package display

import (
	"fmt"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func PrintSepTitle(screen core.Screen, env *core.Env, msg string) {
	width := env.GetInt("display.width") - 3
	screen.Print(rpt("-", width-len(msg)) + "<[" + msg + "]\n")
}

// TODO: pad/cut title, make it fixed length
func PrintDisplayBlockSep(screen core.Screen, title string) {
	screen.Print(fmt.Sprintf("-------=<%s>=-------\n", title))
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
