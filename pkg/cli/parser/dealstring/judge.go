package dealstring

import (
	"regexp"
	"strconv"
)

func IsUpper(ch uint8) bool {
	return ch >= 'A' && ch <= 'Z'
}

func IsLower(ch uint8) bool {
	return ch >= 'a' && ch <= 'z'
}

func IsDigit(ch uint8) bool {
	return ch >= '0' && ch <= '9'
}

func IsXdigit(ch uint8) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func IsAlpha(ch uint8) bool {
	return IsLower(ch) || IsUpper(ch)
}

func IsAlnum(ch uint8) bool {
	return IsAlpha(ch) || IsDigit(ch)
}

func All(str string, op func(uint8) bool) bool {
	for i := range str {
		if !op(str[i]) {
			return false
		}
	}

	return true
}

func Any(str string, op func(uint8) bool) bool {
	for i := range str {
		if op(str[i]) {
			return true
		}
	}

	return false
}

var floatRegex = regexp.MustCompile(`^[+-]?\d+(.\d+)?([eE]\d+)?$`)

func IsFloat(str string) bool {
	return IsFloatV1(str)
}

func IsFloatV2(str string) bool {
	_, err := strconv.ParseFloat(str, 64)
	return err == nil
}

func IsFloatV1(str string) bool {
	return floatRegex.Match([]byte(str))
}

var identifierRegex = regexp.MustCompile(`^[a-zA-Z]\w+$`)

func IsIdentifier(str string) bool {
	return identifierRegex.Match([]byte(str))
}

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,4}$`)

func IsEmail(str string) bool {
	return emailRegex.Match([]byte(str))
}

var phoneRegex = regexp.MustCompile(`^1[345789][0-9]{9}$`)

func IsPhone(str string) bool {
	return phoneRegex.Match([]byte(str))
}
