package model

import (
	"strings"
)

func StrToBool(s string) bool {
	return StrToTrue(s)
}

func StrToTrue(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "t" || s == "1" || s == "on" || s == "y" || s == "yes"
}

func StrToFalse(s string) bool {
	s = strings.ToLower(s)
	return s == "false" || s == "f" || s == "0" || s == "off" || s == "n" || s == "no"
}
