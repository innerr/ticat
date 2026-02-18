package model

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

func (self *Helps) RegHelpFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("read help file failed: %v", err)
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	self.RegHelp(lines[0], lines[1:])
	return nil
}

func (self *Helps) RegHelp(title string, text []string) {
	self.Sections = append(self.Sections, Help{title, text})
}
