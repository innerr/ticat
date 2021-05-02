package display

import (
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func RenderCmdStack(l CmdStackLines, env *core.Env, screen core.Screen) {
	if !l.Display {
		return
	}

	pln := func(text string) {
		screen.Print(text + "\n")
	}

	meow := "   (=`ω´=)   "
	const meowLen = 3 + 7 + 3
	if !env.GetBool("display.utf8") {
		meow = rpt(" ", meowLen)
	}

	c := getFrameChars(env)

	titleLine := c.V + " " + l.StackDepth + c.V
	titleInner := 1 + l.StackDepthLen
	titleLineLen := 1 + titleInner + 1

	width := env.GetInt("display.width") - 2
	if width < titleLineLen+meowLen+l.TimeLen {
		width = titleLineLen+meowLen+l.TimeLen
	}

	pln(c.P1 + strings.Repeat(c.H, titleInner) + c.P3)
	pln(titleLine + meow + rpt(" ", width-titleLineLen-meowLen-l.TimeLen) + l.Time)
	pln(c.P4 + rpt(c.H, titleInner) + c.P8 + rpt(c.H, width-1-titleInner) + c.P3)

	for i, line := range l.Env {
		pln(c.V + " " + line + rpt(" ", width-1-l.EnvLen[i]) + c.V)
	}
	if len(l.Env) != 0 {
		pln(c.P4 + rpt(c.H, width) + c.P6)
	}
	for i, line := range l.Flow {
		pln(c.V + " " + line + rpt(" ", width-1-l.FlowLen[i]) + c.V)
	}
	pln(c.P7 + rpt(c.H, width) + c.P9)
}

func RenderCmdResult(l CmdResultLines, env *core.Env, screen core.Screen) {
	if !l.Display {
		return
	}

	pln := func(text string) {
		screen.Print(text + "\n")
	}

	width := env.GetInt("display.width") - 2
	pad := width-1-l.ResLen-1-l.CmdLen-l.DurLen-1

	if pad < 0 {
		width += -pad
		pad = 0
	}

	c := getFrameChars(env)

	pln(c.P1 + rpt(c.H, width) + c.P3)
	pln(c.V + " " + l.Res + " " + l.Cmd + rpt(" ", pad) + l.Dur + " " + c.V)
	pln(c.P7 + rpt(c.H, width) + c.P9)
	if l.FooterLen != 0 {
		pln(rpt(" ", width-l.FooterLen) + l.Footer)
	}
}

type FrameChars struct {
	// Sudoku positions
	P1 string
	P2 string
	P3 string
	P4 string
	P5 string
	P6 string
	P7 string
	P8 string
	P9 string

	V string
	H string
}

func FrameCharsUtf8() *FrameChars {
	return &FrameChars {
		"┌", "┬", "┐",
		"├", "┼", "┤",
		"└", "┴", "┘",
		"│", "─",
	}
}

func FrameCharsAscii() *FrameChars {
	return &FrameChars {
		"+", "+", "+",
		"+", "+", "+",
		"+", "+", "+",
		"|", "-",
	}
}

func FrameCharsNoCorner() *FrameChars {
	return &FrameChars {
		" ", " ", " ",
		" ", " ", " ",
		" ", " ", " ",
		"|", "-",
	}
}

// TODO: respect env val "display.utf8"
func getFrameChars(env *core.Env) *FrameChars {
	if env.GetBool("display.utf8") {
		return FrameCharsUtf8()
	} else {
		return FrameCharsAscii()
	}
}

func rpt(char string, count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat(char, count)
}
