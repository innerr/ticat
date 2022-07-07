package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func TooMuchOutput(env *core.Env, screen core.Screen) bool {
	height := env.GetInt("display.height")
	return height > 0 && screen.OutputNum() > int(float64(height)*1.1)
}

func mayQuoteStr(origin string) string {
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

func padCmd(str string, width int, env *core.Env) string {
	return ColorCmd(padR(str, width), env)
}

func padR(str string, width int) string {
	return padRight(str, " ", width)
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

func mayMaskSensitiveVal(env *core.Env, key string, val string) string {
	if env.GetBool("display.sensitive") {
		return val
	}
	if core.IsSensitiveKeyVal(key, val) && len(val) != 0 {
		val = "***"
	}
	return val
}
