package display

import (
	"fmt"
	"strings"
	"time"
)

func MayQuoteStr(origin string) string {
	trimed := strings.TrimSpace(origin)
	if len(trimed) == 0 || len(trimed) != len(origin) {
		return "'" + origin + "'"
	}
	fields := strings.Fields(origin)
	if len(fields) != 1 {
		return "'" + origin + "'"
	}
	return origin
}

func autoPadNewLine(padding string, msg string) string {
	msgNoPad := strings.TrimLeft(msg, "\t '\"")
	hiddenPad := rpt(" ", len(msg)-len(msgNoPad))
	msg = strings.ReplaceAll(msg, "\n", "\n"+padding+hiddenPad)
	return msg
}

func padRight(str string, pad string, width int) string {
	if len(str) >= width {
		return str
	}
	return str + strings.Repeat(pad, width-len(str))
}

func formatDuration(dur time.Duration) string {
	return strings.ReplaceAll(fmt.Sprintf("%s", dur), "µ", "u")
}

func rpt(char string, count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat(char, count)
}
