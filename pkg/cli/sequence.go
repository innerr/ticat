package cli

import (
	"strings"
)

type SequenceBreaker struct {
	Sep            string
	UnbreakPrefixs []string
	UnbreakSuffixs []string
}

func (self *SequenceBreaker) Normalize(argv []string) []string {
	sep_n := len(self.Sep)
	res := []string{}
	for _, arg := range argv {
		handled_idx := 0
		search_idx := 0
		for ; handled_idx < len(arg); search_idx += sep_n {
			i := strings.Index(arg[search_idx:], self.Sep)
			if i < 0 {
				res = append(res, arg[handled_idx:])
				break
			}
			search_idx += i
			unbreakable := false
			for _, prefix := range self.UnbreakPrefixs {
				if search_idx >= len(prefix) && prefix == arg[search_idx-len(prefix):search_idx] {
					unbreakable = true
					break
				}
			}
			if unbreakable {
				continue
			}
			for _, suffix := range self.UnbreakSuffixs {
				if search_idx+sep_n+len(suffix) <= len(arg) && suffix == arg[search_idx+sep_n:search_idx+sep_n+len(suffix)] {
					unbreakable = true
					break
				}
			}
			if unbreakable {
				continue
			}
			if handled_idx != search_idx {
				res = append(res, arg[handled_idx:search_idx])
			}
			res = append(res, self.Sep)
			handled_idx = search_idx + sep_n
		}
	}
	return res
}

func (self *SequenceBreaker) Break(argv []string) [][]string {
	argv = self.Normalize(argv)
	res := [][]string{}
	curr := []string{}
	for _, arg := range argv {
		if arg != self.Sep {
			curr = append(curr, arg)
		} else if len(curr) != 0 {
			res = append(res, curr)
			curr = []string{}
		}
	}
	if len(curr) != 0 {
		res = append(res, curr)
	}
	return res
}
