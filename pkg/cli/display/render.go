package display

import (
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func RenderCmdStack(l CmdStackLines, env *core.Env, screen core.Screen) (renderWidth int) {
	if !l.Display {
		return
	}

	pln := func(text string) {
		screen.Print(text + "\n")
	}

	meow := "   (=`ω´=)   "
	const meowLen = 3 + 7 + 3
	if !env.GetBool("display.utf8") || !env.GetBool("display.meow") {
		meow = rpt(" ", meowLen)
	}

	c := getFrameChars(env)

	titleLine := c.V + " " + l.StackDepth + c.V
	titleInner := 1 + l.StackDepthLen
	titleLineLen := 1 + titleInner + 1

	width := env.GetInt("display.width") - 2
	if width < titleLineLen+meowLen+l.TimeLen {
		width = titleLineLen + meowLen + l.TimeLen
	}

	pln(c.P1 + strings.Repeat(c.H, titleInner) + c.P3)
	pln(titleLine + meow + rpt(" ", width-titleLineLen-meowLen-l.TimeLen) + l.Time)
	pln(c.P4 + rpt(c.H, titleInner) + c.P8 + rpt(c.H, width-1-titleInner) + c.P3)

	for i, line := range l.Env {
		padWid := width - 1 - l.EnvLen[i]
		if padWid >= 0 {
			pln(c.V + " " + line + rpt(" ", padWid) + c.V)
		} else {
			pln(c.V + " " + line)
		}
	}
	if len(l.Env) != 0 {
		pln(c.P4 + rpt(c.H, width) + c.P6)
	}
	for i, line := range l.Flow {
		padWid := width - 1 - l.FlowLen[i]
		if padWid >= 0 {
			pln(c.V + " " + line + rpt(" ", padWid) + c.V)
		} else {
			pln(c.V + " " + line)
		}
	}
	pln(c.P7 + rpt(c.H, width) + c.P9)

	return width + 2
}

func RenderCmdResult(l CmdResultLines, env *core.Env, screen core.Screen, width int) {
	if !l.Display {
		return
	}

	pln := func(text string) {
		screen.Print(text + "\n")
	}

	width -= 2
	pad := width - 1 - l.ResLen - 1 - l.CmdLen - l.DurLen - 1

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

func getFrameChars(env *core.Env) *FrameChars {
	name := strings.ToLower(env.Get("display.style").Raw)
	if env.GetBool("display.utf8") {
		if strings.Index(name, "utf") >= 0 {
			return FrameCharsUtf8()
		}
		if strings.Index(name, "heavy") >= 0 || strings.Index(name, "bold") >= 0 {
			return FrameCharsUtf8Heavy()
		}
	}
	if strings.Index(name, "slash") >= 0 {
		return FrameCharsNoSlash()
	}
	if strings.Index(name, "corner") >= 0 {
		return FrameCharsNoCorner()
	}
	return FrameCharsAscii()
}
