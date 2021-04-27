package parser

import (
	"strings"
)

type sequenceParser struct {
	Sep            string
	UnbreakPrefixs []string
	UnbreakSuffixs []string
}

func (self *sequenceParser) Normalize(argv []string) []string {
	sepN := len(self.Sep)
	res := []string{}
	for _, arg := range argv {
		handledIdx := 0
		searchIdx := 0
		for ; handledIdx < len(arg); searchIdx += sepN {
			i := strings.Index(arg[searchIdx:], self.Sep)
			if i < 0 {
				res = append(res, arg[handledIdx:])
				break
			}
			searchIdx += i
			unbreakable := false
			for _, prefix := range self.UnbreakPrefixs {
				if searchIdx >= len(prefix) && prefix == arg[searchIdx-len(prefix):searchIdx] {
					unbreakable = true
					break
				}
			}
			if unbreakable {
				continue
			}
			for _, suffix := range self.UnbreakSuffixs {
				if searchIdx+sepN+len(suffix) <= len(arg) && suffix == arg[searchIdx+sepN:searchIdx+sepN+len(suffix)] {
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
			res = append(res, self.Sep)
			handledIdx = searchIdx + sepN
		}
	}
	return res
}

func (self *sequenceParser) Parse(argv []string) (parsed [][]string, firstIsGlobal bool) {
	argv = self.Normalize(argv)

	firstIsGlobal = true
	if len(argv) != 0 && argv[0] == self.Sep {
		firstIsGlobal = false
	}

	parsed = [][]string{}
	curr := []string{}
	for _, arg := range argv {
		if arg != self.Sep {
			arg = strings.TrimSpace(arg)
			if len(arg) != 0 {
				curr = append(curr, arg)
			}
		} else if len(curr) != 0 {
			parsed = append(parsed, curr)
			curr = []string{}
		}
	}
	if len(curr) != 0 {
		parsed = append(parsed, curr)
	}
	return
}
