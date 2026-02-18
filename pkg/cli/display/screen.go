package display

import (
	"github.com/innerr/ticat/pkg/core/model"
)

func PrintFramedLines(screen model.Screen, env *model.Env, buf *CacheScreen, c *FrameChars) {
	if buf.IsEmpty() {
		return
	}
	width := env.GetInt("display.width") - 2
	if c == nil {
		c = getFrameChars(env)
	}
	screen.Print(c.P1 + rpt(c.H, width) + c.P3 + "\n")
	buf.WriteToEx(screen, func(line string, isError bool, lineLen int) (string, bool) {
		rightV := c.V
		if lineLen > width {
			rightV = ""
		}
		line = c.V + line + rpt(" ", width-lineLen) + rightV + "\n"
		return line, isError
	})
	screen.Print(c.P7 + rpt(c.H, width) + c.P9 + "\n")
}

type CacheScreen struct {
	data []CachedOutput
	outN int
}

type CachedOutput struct {
	Text    string
	IsError bool
	Len     int
}

func NewCacheScreen() *CacheScreen {
	return &CacheScreen{nil, 0}
}

func (self *CacheScreen) IsEmpty() bool {
	return len(self.data) == 0
}

func (self *CacheScreen) Print(text string) error {
	self.data = append(self.data, CachedOutput{text, false, len(text)})
	self.outN += 1
	return nil
}

func (self *CacheScreen) PrintEx(text string, textLen int) {
	self.data = append(self.data, CachedOutput{text, false, textLen})
	self.outN += 1
}

func (self *CacheScreen) Error(text string) error {
	self.data = append(self.data, CachedOutput{text, true, len(text)})
	return nil
}

func (self *CacheScreen) OutputtedLines() int {
	return self.outN
}

func (self *CacheScreen) WriteToEx(
	screen model.Screen,
	transformer func(text string, isError bool, textLen int) (string, bool)) {

	for _, it := range self.data {
		text, isError := transformer(it.Text, it.IsError, it.Len)
		if isError {
			screen.Error(text)
		} else {
			screen.Print(text)
		}
	}
}

func (self *CacheScreen) WriteTo(screen model.Screen) {
	self.WriteToEx(screen, func(text string, isError bool, textLen int) (string, bool) {
		return text, isError
	})
}
