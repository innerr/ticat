package utils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func ReadLogFileLastLines(path string, bufSize int, maxLines int) (lines []string) {
	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("[ReadLastLines] %v", err))
	}
	defer file.Close()

	stat, err := os.Stat(path)
	if int64(bufSize) >= stat.Size() {
		bufSize = int(stat.Size())
	}

	buf := make([]byte, bufSize)
	start := stat.Size() - int64(bufSize)
	if start < 0 {
		start = 0
	}
	_, err = file.ReadAt(buf, start)
	if err != nil && !errors.Is(err, io.EOF) {
		panic(fmt.Errorf("[ReadLastLines] %v", err))
	}

	for _, line := range strings.Split(string(buf), "\n") {
		if len(strings.TrimSpace(line)) != 0 {
			lines = append(lines, line)
		}
	}

	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return
}

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
		return GoRoutineIdStrMain
	}
	return strconv.Itoa(id)
}

func RandomName(n uint) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = Chars[rand.Intn(len(Chars))]
	}
	return string(b)
}

func NormalizeDurStr(durStr string) string {
	// Default unit is 's'
	_, err := strconv.ParseFloat(durStr, 64)
	if err == nil {
		durStr += "s"
	}
	return durStr
}

func QuoteStrIfHasSpace(str string) string {
	if strings.IndexAny(str, " \t\r\n") < 0 {
		return str
	}
	i := strings.Index(str, "\"")
	if i < 0 {
		return "\"" + str + "\""
	}
	i = strings.Index(str, "'")
	if i < 0 {
		return "'" + str + "'"
	}
	return str
}

const (
	GoRoutineIdStrMain = "main"
	Chars              = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)
