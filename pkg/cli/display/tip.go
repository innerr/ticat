package display

import (
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

type TipBoxPrinter struct {
	screen   core.Screen
	env      *core.Env
	isErr    bool
	inited   bool
	buf      *CacheScreen
	maxWidth int
}

func NewTipBoxPrinter(screen core.Screen, env *core.Env, isErr bool) *TipBoxPrinter {
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

func (self *TipBoxPrinter) Print(msg string) {
	msg = strings.TrimRight(msg, "\n")
	msgs := strings.Split(msg, "\n")
	if len(msgs) > 1 {
		self.Prints(msgs...)
		return
	}

	if !self.inited {
		var tip string
		var tipLen int
		utf8 := self.env.GetBool("display.utf8.symbols")
		if self.isErr {
			tip = "<ERR>"
			tipLen = len(tip)
			if utf8 {
				tip = "🛑"
				tipLen = 2
			}
		} else {
			tip = "<TIP>"
			tipLen = len(tip)
			if utf8 {
				tip = "💡"
				tipLen = 2
			}
		}
		self.buf.PrintEx(tip+" "+msg, len(msg)+1+tipLen)
		self.inited = true
	} else {
		self.buf.Print("   " + msg)
	}
}

func (self *TipBoxPrinter) Error(msg string) {
	self.buf.Error(msg)
}

func (self *TipBoxPrinter) OutputNum() int {
	return self.buf.OutputNum()
}

func (self *TipBoxPrinter) Finish() {
	PrintFramedLines(self.screen, self.env, self.buf)
}

func PrintTipTitle(screen core.Screen, env *core.Env, msgs ...string) {
	printTipTitle(screen, env, false, msgs...)
}

func printTipTitle(screen core.Screen, env *core.Env, isErr bool, msgs ...string) {
	printer := NewTipBoxPrinter(screen, env, isErr)
	printer.Prints(msgs...)
	printer.Finish()
}
