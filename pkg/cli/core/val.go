package core

import (
	"strings"
)

func StrToBool(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "t" || s == "1" || s == "on" || s == "y"
}
