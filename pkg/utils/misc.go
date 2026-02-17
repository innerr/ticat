package utils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

func ReadLogFileLastLines(path string, bufSize int, maxLines int) (lines []string) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		panic(fmt.Errorf("[ReadLastLines] %v", err))
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(fmt.Errorf("[ReadLogFileLastLines] close file '%s' failed: %v", path, err))
		}
	}()

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

func GetTerminalWidth(defRow int, defCol int) (row int, col int) {
	size := &TerminalSize{}
	retCode, _, _ := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(size)))
	if int(retCode) == -1 {
		return defRow, defCol
	}
	return int(size.Row), int(size.Col)
}

func StdoutIsPipe() bool {
	fo, _ := os.Stdout.Stat()
	return (fo.Mode() & os.ModeCharDevice) == 0
}

func MoveFile(src string, dest string) error {
	err := os.Rename(src, dest)
	if err == nil {
		return nil
	}
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
	rand.Seed(time.Now().UnixNano())
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

// TODO: may not right, use PidExists to do that
func IsPidRunning(pid int) bool {
	// err := syscall.Kill(pid, syscall.Signal(0))
	// return err == nil || err != syscall.ESRCH
	exists, err := IsPidExists(pid)
	if err == nil && !exists {
		return false
	}
	return true
}

func IsPidExists(pid int) (bool, error) {
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return false, err
	}
	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true, nil
	}
	if err.Error() == "os: process already finished" {
		return false, nil
	}
	errno, ok := err.(syscall.Errno)
	if !ok {
		return false, err
	}
	switch errno {
	case syscall.ESRCH:
		return false, nil
	case syscall.EPERM:
		return true, nil
	}
	return false, err
}

func IsOsCmdExists(cmd string) bool {
	path, err := exec.LookPath(cmd)
	return err == nil && len(path) > 0
}

func FindPython() (path string) {
	cmds := []string{"python3", "python", "python2"}
	for _, cmd := range cmds {
		path, err := exec.LookPath(cmd)
		if err == nil && len(path) > 0 {
			return path
		}
	}
	return
}

func IpId() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return IpIdOnError
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			// interface down
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			// loopback interface
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return IpIdOnError
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				// not an ipv4 address
				continue
			}
			return ip.String()
		}
	}
	return IpIdOnNoNetwork
}

const (
	GoRoutineIdStrMain = "main"
	Chars              = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	IpIdOnError        = "unknown-ip-fail-to-get"
	IpIdOnNoNetwork    = "unknown-ip-no-network"
)
