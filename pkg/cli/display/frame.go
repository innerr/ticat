package display

import (
	"fmt"
)

type FrameChars struct {
	V string
	H string

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
}

func FrameCharsHeavy() *FrameChars {
	return &FrameChars{
		"|", "=",
		"=", "=", "=",
		"=", "=", "=",
		"=", "=", "=",
	}
}

func FrameCharsUtf8Heavy() *FrameChars {
	return &FrameChars{
		"┃", "━",
		"┏", "┳", "┓",
		"┣", "╋", "┫",
		"┗", "┻", "┛",
	}
}

func FrameCharsUtf8() *FrameChars {
	return &FrameChars{
		"│", "─",
		"┌", "┬", "┐",
		"├", "┼", "┤",
		"└", "┴", "┘",
	}
}

func FrameCharsUtf8Colored(colorCode int) *FrameChars {
	prefix := "\033[38;5;" + fmt.Sprintf("%d", colorCode) + "m"
	darken := func(s string) string {
		return prefix + s + "\033[0m"
	}
	return &FrameChars{
		darken("│"), darken("─"),
		darken("┌"), darken("┬"), darken("┐"),
		darken("├"), darken("┼"), darken("┤"),
		darken("└"), darken("┴"), darken("┘"),
	}
}

func FrameCharsAscii() *FrameChars {
	return &FrameChars{
		"|", "-",
		"+", "+", "+",
		"+", "+", "+",
		"+", "+", "+",
	}
}

func FrameCharsNoSlash() *FrameChars {
	return &FrameChars{
		"-", "-",
		"+", "+", "+",
		"+", "+", "+",
		"+", "+", "+",
	}
}

func FrameCharsNoBorder() *FrameChars {
	return &FrameChars{
		"|", " ",
		".", " ", ".",
		" ", "+", " ",
		"+", " ", "+",
	}
}

func FrameCharsNoCorner() *FrameChars {
	return &FrameChars{
		"|", "-",
		" ", " ", " ",
		" ", " ", " ",
		" ", " ", " ",
	}
}
