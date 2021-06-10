package utils

import (
	"bufio"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func UserConfirm() (line string) {
	buf := bufio.NewReader(os.Stdin)
	text, err := buf.ReadBytes('\n')
	if err != nil {
		panic(fmt.Errorf("[readFromStdin] read from stdin failed: %v", err))
	}
	line = string(text)
	return
}

type TerminalSize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func GetTerminalWidth() (row int, col int) {
	size := &TerminalSize{}
	retCode, _, _ := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(size)))
	if int(retCode) == -1 {
		return -1, -1
	}
	return int(size.Row), int(size.Col)
}
