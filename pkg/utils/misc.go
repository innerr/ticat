package utils

import (
	"bufio"
	"fmt"
	"os"
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
