package display

import (
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/utils"
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

	titleLine := c.V + " " + l.Title + " " + c.V
	titleInner := 1 + l.TitleLen + 1
	titleLineLen := 1 + titleInner + 1

	width := env.GetInt("display.width") - 2
	if width < titleLineLen+meowLen+l.TimeLen {
		width = titleLineLen + meowLen + l.TimeLen
	}

	pln(c.P1 + strings.Repeat(c.H, titleInner) + c.P3)
	pln(titleLine + meow + rpt(" ", width-titleLineLen-meowLen-l.TimeLen) + l.Time)
	pln(c.P4 + rpt(c.H, titleInner) + c.P8 + rpt(c.H, width-1-titleInner) + c.P3)

	plns := func(lines []string, lens []int, name string) {
		name = "  " + name + " "
		nameLen := len(name)
		name = ColorProp(name, env)
		for i, line := range lines {
			padWid := width - 1 - lens[i]
			tail := ""
			line = c.V + " " + line
			if i == 0 && len(lines) >= 1 && padWid >= nameLen {
				tail = name
				padWid -= nameLen
			}
			if padWid >= 0 {
				line += rpt(" ", padWid) + tail + c.V
			}
			pln(line)
		}
	}

	plns(l.Bg, l.BgLen, "bg")
	if len(l.Bg) != 0 {
		pln(c.P4 + rpt(c.H, width) + c.P6)
	}

	plns(l.Env, l.EnvLen, "env")
	if len(l.Env) != 0 {
		pln(c.P4 + rpt(c.H, width) + c.P6)
	}

	plns(l.Stack, l.StackLen, "stack")
	if len(l.Stack) != 0 {
		pln(c.P4 + rpt(c.H, width) + c.P6)
	}

	plns(l.Flow, l.FlowLen, "flow")
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

func PrintSwitchingThreadDisplay(preTid string, info core.BgTaskInfo, env *core.Env, screen core.Screen) {

	var title string
	var extraLen int

	if preTid == utils.GoRoutineIdStrMain {
		title = ColorThread("thread ", env) + preTid +
			ColorThread(" ended, switch display to thread ", env) + info.Tid +
			ColorThread(", command ", env) + ColorCmdDelay("["+info.Cmd+"]", env)
		extraLen = ColorExtraLen(env, "thread", "thread", "thread", "cmd-delay")
	} else {
		title = ColorThread("switch display to thread ", env) + info.Tid +
			ColorThread(", command ", env) + ColorCmdDelay("["+info.Cmd+"]", env)
		extraLen = ColorExtraLen(env, "thread", "thread", "cmd-delay")
	}
	if !info.Started {
		title += ColorThread(", not started now", env)
		extraLen += ColorExtraLen(env, "thread")
	} else if info.Started && !info.Finished {
		title += ColorThread(", still running", env)
		extraLen += ColorExtraLen(env, "thread")
	} else if info.Finished {
		title += ColorThread(", already ended", env)
		extraLen += ColorExtraLen(env, "thread")
	}
	titleLen := len(title) - extraLen

	width := env.GetInt("display.width") - 2
	pad := width - titleLen - 1
	c := getFrameCharsByName(env, "heavy")

	pln := func(text string) {
		screen.Print(text + "\n")
	}

	pln(ColorTip(c.P1+rpt(c.H, width)+c.P3, env))
	if pad >= 0 {
		pln(ColorTip(c.V, env) + " " + title + rpt(" ", pad) + ColorTip(c.V, env))
	} else {
		pln(ColorTip(c.V, env) + " " + title)
	}
	pln(ColorTip(c.P7+rpt(c.H, width)+c.P9, env))
}

func getFrameChars(env *core.Env) *FrameChars {
	name := strings.ToLower(env.Get("display.style").Raw)
	return getFrameCharsByName(env, name)
}

func getFrameCharsByName(env *core.Env, name string) *FrameChars {
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
	if strings.Index(name, "heavy") >= 0 {
		return FrameCharsHeavy()
	}
	return FrameCharsAscii()
}
