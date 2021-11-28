package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func UserConfirm() (yes bool) {
	buf := bufio.NewReader(os.Stdin)
	for {
		line, err := buf.ReadBytes('\n')
		if err != nil {
			panic(fmt.Errorf("[readFromStdin] read from stdin failed: %v", err))
		}
		if len(line) > 0 && (line[0] == 'y' || line[0] == 'Y') {
			return true
		}
	}
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

func StdoutIsPipe() bool {
	fo, _ := os.Stdout.Stat()
	return (fo.Mode() & os.ModeCharDevice) == 0
}

func MoveFile(src string, dest string) error {
	err := os.Rename(src, dest)
	if strings.Index(err.Error(), "invalid cross-device link") < 0 {
		return err
	}
	cmd := exec.Command("mv", src, dest)
	_, err = cmd.Output()
	return err
}

func GoRoutineId() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func GoRoutineIdStr() string {
	id := GoRoutineId()
	if id == 1 {
		return "main"
	}
	return strconv.Itoa(id)
}
