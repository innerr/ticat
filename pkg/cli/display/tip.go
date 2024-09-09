package display

import (
	"fmt"
	"strings"

	"github.com/innerr/ticat/pkg/core/model"
)

func PrintErrTitle(screen model.Screen, env *model.Env, msgs ...interface{}) {
	printTipTitle(screen, env, true, msgs...)
}

func PrintTipTitle(screen model.Screen, env *model.Env, msgs ...interface{}) {
	printTipTitle(screen, env, false, msgs...)
}

func printTipTitle(screen model.Screen, env *model.Env, isErr bool, msgs ...interface{}) {
	var strs []string
	for _, it := range msgs {
		switch it.(type) {
		case string:
			strs = append(strs, it.(string))
		case []string:
			strs = append(strs, it.([]string)...)
		default:
			panic(fmt.Errorf("[PrintTipTitle] invalid msg type, should never happen"))
		}
	}

	printer := NewTipBoxPrinter(screen, env, isErr)
	printer.Prints(strs...)
	printer.Finish()
}

type TipBoxPrinter struct {
	screen   model.Screen
	env      *model.Env
	isErr    bool
	inited   bool
	buf      *CacheScreen
	maxWidth int
}

func NewTipBoxPrinter(screen model.Screen, env *model.Env, isErr bool) *TipBoxPrinter {
	return &TipBoxPrinter{
		screen,
		env,
		isErr,
		false,
		NewCacheScreen(),
		env.GetInt("display.width") - 4 - 2,
	}
}

func (self *TipBoxPrinter) PrintWrap(msgs ...string) {
	for _, msg := range msgs {
		for len(msg) > self.maxWidth {
			self.Print(msg[0:self.maxWidth])
			msg = msg[self.maxWidth:]
		}
		self.Print(msg)
	}
}

func (self *TipBoxPrinter) Prints(msgs ...string) {
	for _, msg := range msgs {
		self.Print(msg)
	}
}

func (self *TipBoxPrinter) colorize(msg string) (string, int) {
	if !self.env.GetBool("display.color") {
		return msg, 0
	}
	return "\033[38;5;242m" + msg + "\033[0m", ColorExtraLen(self.env, "tip-dark")
}

func (self *TipBoxPrinter) Print(msg string) {
	msg = strings.TrimRight(msg, "\n")
	msgs := strings.Split(msg, "\n")
	if len(msgs) > 1 {
		self.Prints(msgs...)
		return
	}

	msg, colorLen := self.colorize(msg)

	// TODO: put ERR TIP to env strs

	if !self.inited {
		var tip string
		var tipLen int
		utf8 := self.env.GetBool("display.utf8.symbols")
		if self.isErr {
			tip = " <ERR> "
			tipLen = len(tip)
			if utf8 {
				tip = self.env.GetRaw("display.utf8.symbols.err")
				tipLen = self.env.GetInt("display.utf8.symbols.err.len")
			}
			tip = ColorError(tip, self.env)
		} else {
			tip = " <TIP> "
			tipLen = len(tip)
			if utf8 {
				tip = self.env.GetRaw("display.utf8.symbols.tip")
				tipLen = self.env.GetInt("display.utf8.symbols.tip.len")
			}
			tip = ColorTip(tip, self.env)
		}
		self.buf.PrintEx(tip+msg, len(msg)+tipLen-colorLen)
		self.inited = true
	} else {
		msg = "   " + msg
		self.buf.PrintEx(msg, len(msg)-colorLen)
	}
}

func (self *TipBoxPrinter) Error(msg string) {
	self.buf.Error(msg)
}

func (self *TipBoxPrinter) OutputtedLines() int {
	return self.buf.OutputtedLines()
}

func (self *TipBoxPrinter) Finish() {
	if !self.env.GetBool("display.tip") {
		return
	}
	colorCode := colorCodeTipDark
	if self.isErr {
		colorCode = colorCodeError
	}
	var frameChars *FrameChars
	if self.env.GetBool("display.utf8") {
		if self.env.GetBool("display.color") {
			frameChars = FrameCharsUtf8Colored(colorCode)
		} else {
			frameChars = FrameCharsUtf8()
		}
	} else {
		if self.env.GetBool("display.color") {
			frameChars = FrameCharsAsciiColored(colorCode)
		} else {
			frameChars = FrameCharsAscii()
		}
	}
	PrintFramedLines(self.screen, self.env, self.buf, frameChars)
}
