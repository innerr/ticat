package core

import (
	"bufio"
	"fmt"
	"os"
)

type Help struct {
	Title string
	Text  []string
}

type Helps struct {
	Sections []Help
}

func NewHelps() *Helps {
	return &Helps{nil}
}

func (self *Helps) RegHelpFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("read help file failed: %v", err))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	self.RegHelp(lines[0], lines[1:])
}

func (self *Helps) RegHelp(title string, text []string) {
	self.Sections = append(self.Sections, Help{title, text})
}
