package flow_file

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func LoadFlowFile(path string) (flow string, help string, abbrs string) {
	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("[LoadFlowFile] open flow file '%s' failed: %v", path, err))
	}
	defer file.Close()
	return LoadFlow(file)
}

func LoadFlow(reader io.Reader) (flow string, help string, abbrs string) {
	const kvSep = " =\t"
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	var lines []string
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(text, "#") {
			lines = append(lines, text)
			continue
		}
		text = strings.TrimSpace(strings.TrimLeft(text, "#"))
		if strings.HasPrefix(text, "help") {
			help = strings.Trim(strings.TrimLeft(text, "help"), kvSep)
			continue
		}
		if strings.HasPrefix(text, "abbrs") {
			abbrs = strings.Trim(strings.TrimLeft(text, "abbrs"), kvSep)
			continue
		}
		if strings.HasPrefix(text, "abbr") {
			abbrs = strings.Trim(strings.TrimLeft(text, "abbr"), kvSep)
			continue
		}
	}
	flow = strings.Join(lines, " ")
	return
}

func SaveFlowFile(path string, flow string, help string, abbrs string) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(fmt.Errorf("[SaveFlowFile] open flow file '%s' failed: %v", path, err))
	}
	defer file.Close()
	SaveFlow(file, flow, help, abbrs)
}

func SaveFlow(w io.Writer, flow string, help string, abbrs string) {
	if len(help) != 0 || len(abbrs) != 0 {
		fmt.Fprintf(w, "#\n")
	}
	if len(help) != 0 {
		fmt.Fprintf(w, "# help = %s\n", help)
	}
	if len(abbrs) != 0 {
		fmt.Fprintf(w, "# abbrs = %s\n", abbrs)
	}
	if len(help) != 0 || len(abbrs) != 0 {
		fmt.Fprintf(w, "#\n")
	}
	fmt.Fprintf(w, "%s\n", flow)
}
