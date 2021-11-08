package parser

import (
	"strings"
)

type SequenceParser struct {
	sep            string
	unbreakPrefixs []string
	unbreakSuffixs []string
}

func NewSequenceParser(
	sep string,
	unbreakPrefixs []string,
	unbreakSuffixs []string) *SequenceParser {

	return &SequenceParser{
		sep,
		unbreakPrefixs,
		unbreakSuffixs,
	}
}

func (self *SequenceParser) Normalize(argv []string) []string {
	sepN := len(self.sep)
	res := []string{}
	for _, arg := range argv {
		handledIdx := 0
		searchIdx := 0
		for ; handledIdx < len(arg); searchIdx += sepN {
			i := strings.Index(arg[searchIdx:], self.sep)
			if i < 0 {
				res = append(res, arg[handledIdx:])
				break
			}
			searchIdx += i
			// TODO:
			//  1. put '\' to class init args
			//  2. this is a workaround, we also need to handle '\\', '\\\', ... things like that
			if searchIdx >= 1 && "\\" == arg[searchIdx-1:searchIdx] {
				arg = arg[0:searchIdx-1] + arg[searchIdx:]
				continue
			}
			unbreakable := false
			for _, prefix := range self.unbreakPrefixs {
				if searchIdx >= len(prefix) &&
					prefix == arg[searchIdx-len(prefix):searchIdx] {
					unbreakable = true
					break
				}
			}
			if unbreakable {
				continue
			}
			for _, suffix := range self.unbreakSuffixs {
				if searchIdx+sepN+len(suffix) <= len(arg) &&
					suffix == arg[searchIdx+sepN:searchIdx+sepN+len(suffix)] {
					unbreakable = true
					break
				}
			}
			if unbreakable {
				continue
			}
			if handledIdx != searchIdx {
				res = append(res, arg[handledIdx:searchIdx])
			}
			res = append(res, self.sep)
			handledIdx = searchIdx + sepN
		}
	}
	return res
}

func (self *SequenceParser) Parse(argv []string) (parsed [][]string, firstIsGlobal bool) {
	argv = self.Normalize(argv)

	firstIsGlobal = true
	if len(argv) != 0 && argv[0] == self.sep {
		firstIsGlobal = false
	}

	parsed = [][]string{}
	curr := []string{}
	for _, arg := range argv {
		if arg != self.sep {
			arg = strings.TrimSpace(arg)
			//if len(arg) != 0 {
			curr = append(curr, arg)
			//}
			//} else if len(curr) != 0 {
		} else {
			parsed = append(parsed, curr)
			curr = []string{}
		}
	}
	//if len(curr) != 0 {
	parsed = append(parsed, curr)
	//}
	return
}
