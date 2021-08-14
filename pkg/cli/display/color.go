package display

import (
	"fmt"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

// TODO: this is slow, fetch flag from env too many times, and other issues, handle it later

func ColorExtraLen(env *core.Env, types ...string) (res int) {
	enabled := env.GetBool("display.color")
	if !enabled {
		return 0
	}
	lens := map[string]int{
		"hub":        3,
		"arg":        3,
		"key":        3,
		"warn":       3,
		"error":      3,
		"tip-symbol": 3,
		"symbol":     3,
		"prop":       3,
		"tag":        2,
		"help":       2,
		"flow":       2,
		"cmd":        2,
		"cmd-done":   2,
		"cmd-curr":   2,
		"flowing":    2,
		"enabled":    2,
		"disabled":   3,
	}
	for _, it := range types {
		extra, ok := lens[it]
		if !ok {
			panic(fmt.Errorf("unknown color class: %s", it))
		}
		res += extra + colorExtraLenWithoutCode
	}
	return
}

func ColorHub(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(220), env)
}

func ColorEnabled(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(34), env)
}

func ColorDisabled(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(202), env)
}

func ColorArg(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(215), env)
}

func ColorKey(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(135), env)
}

func ColorWarn(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(202), env)
}

func ColorError(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(124), env)
}

func ColorTipSymbol(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(214), env)
}

func ColorSymbol(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(130), env)
}

func ColorProp(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(130), env)
}

func ColorTag(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(91), env)
}

func ColorHelp(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(27), env)
}

func ColorFlow(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(86), env)
}

func ColorCmd(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(76), env)
}

func ColorCmdCurr(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(46), env)
}

func ColorCmdDone(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(34), env)
}

func ColorFlowing(origin string, env *core.Env) string {
	return colorize(origin, fromColor256(81), env)
}

func DecodeColor(text string, env *core.Env) string {
	for {
		prefix := strings.Index(text, colorEncodePrefix)
		if prefix < 0 {
			return text
		}
		suffix := strings.Index(text[prefix+len(colorEncodePrefix):], colorEncodeSuffix)
		if suffix < 0 {
			return text
		}
		suffix += prefix + len(colorEncodePrefix)
		finish := strings.Index(text[suffix+len(colorEncodeSuffix):], colorEncodeFinish)
		if finish < 0 {
			return text
		}
		finish += suffix + len(colorEncodeSuffix)
		rendering := text[suffix+len(colorEncodeSuffix) : finish]
		color := fromColorName(text[prefix+len(colorEncodePrefix) : suffix])
		text = text[:prefix] + colorize(rendering, color, env) + text[finish+len(colorEncodeFinish):]
	}
}

/*
func Red(origin string, env *core.Env) string {
	return colorize(origin, colorRed, env)
}

func Green(origin string, env *core.Env) string {
	return colorize(origin, colorGreen, env)
}

func Yellow(origin string, env *core.Env) string {
	return colorize(origin, colorYellow, env)
}

func Blue(origin string, env *core.Env) string {
	return colorize(origin, colorBlue, env)
}

func Purple(origin string, env *core.Env) string {
	return colorize(origin, colorPurple, env)
}

func Cyan(origin string, env *core.Env) string {
	return colorize(origin, colorCyan, env)
}

func BrightRed(origin string, env *core.Env) string {
	return colorize(origin, colorBrightRed, env)
}

func BrightGreen(origin string, env *core.Env) string {
	return colorize(origin, colorBrightGreen, env)
}

func BrightYellow(origin string, env *core.Env) string {
	return colorize(origin, colorBrightYellow, env)
}

func BrightBlue(origin string, env *core.Env) string {
	return colorize(origin, colorBrightBlue, env)
}

func BrightPurple(origin string, env *core.Env) string {
	return colorize(origin, colorBrightPurple, env)
}

func BrightCyan(origin string, env *core.Env) string {
	return colorize(origin, colorBrightCyan, env)
}
*/

func colorize(origin string, color string, env *core.Env) string {
	enabled := env.GetBool("display.color")
	if !enabled {
		return origin
	}
	return color + origin + colorReset
}

func fromColor256(code uint8) string {
	return "\033[38;5;" + fmt.Sprintf("%d", code) + "m"
}

func fromColorName(name string) string {
	if strings.HasPrefix(name, color256Prefix) {
		name = name[len(color256Prefix):]
		var code uint8
		scanned, err := fmt.Sscanf(name, "%d", &code)
		if err != nil || scanned != 1 {
			panic(fmt.Errorf("bad 256-color value '%v', err: %v", name, err))
		}
		return fromColor256(code)
	}

	color, ok := map[string]string{
		"red":      colorRed,
		"green":    colorGreen,
		"yellow":   colorYellow,
		"blue":     colorBlue,
		"purple":   colorPurple,
		"cyan":     colorCyan,
		"b-red":    colorBrightRed,
		"b-green":  colorBrightGreen,
		"b-yellow": colorBrightYellow,
		"b-blue":   colorBrightBlue,
		"b-purple": colorBrightPurple,
		"b-cyan":   colorBrightCyan,
	}[name]
	if !ok {
		panic(fmt.Errorf("unknown color name '%s'", name))
	}
	return color
}

const (
	colorReset = "\033[0m"

	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"

	colorBrightRed    = "\033[31;1m"
	colorBrightGreen  = "\033[32;1m"
	colorBrightYellow = "\033[33;1m"
	colorBrightBlue   = "\033[34;1m"
	colorBrightPurple = "\033[35;1m"
	colorBrightCyan   = "\033[36;1m"

	colorExtraLenWithoutCode = len(colorReset) + len("\033[38;5;m")
)

const (
	colorEncodePrefix = "[color="
	colorEncodeSuffix = "]"
	colorEncodeFinish = "[/color]"
	color256Prefix    = "term-"
)
