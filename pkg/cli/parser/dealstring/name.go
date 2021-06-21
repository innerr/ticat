package dealstring

import (
	"bytes"
	"strings"
)

func ToLower(ch uint8) uint8 {
	if IsUpper(ch) {
		return ch + 32
	}
	return ch
}

func ToUpper(ch uint8) uint8 {
	if IsLower(ch) {
		return ch - 32
	}
	return ch
}

func CamelName(str string) string {
	var buf bytes.Buffer
	firstWord := true
	for i := 0; i < len(str); {
		for str[i] == '_' || str[i] == '-' {
			i++
		}
		j := i
		for j < len(str) && IsUpper(str[j]) {
			j++
		}
		if j-i < 2 { // 只有一个大写字母
			for ; j < len(str); j++ {
				if str[j] == '_' || str[j] == '-' || IsUpper(str[j]) {
					break
				}
			}
		} else {
			if j < len(str) && str[j] != '-' && str[j] != '_' {
				j--
			}
		}
		if j-i >= 2 && IsUpper(str[i]) && IsUpper(str[i+1]) { // HELLO
			if firstWord {
				firstWord = false
				for k := i; k < j; k++ {
					buf.WriteByte(ToLower(str[k]))
				}
			} else {
				for k := i; k < j; k++ {
					buf.WriteByte(str[k])
				}
			}
		} else if IsUpper(str[i]) { // Hello
			if firstWord {
				buf.WriteByte(ToLower(str[i]))
				firstWord = false
			} else {
				buf.WriteByte(str[i])
			}
			for k := i + 1; k < j; k++ {
				buf.WriteByte(str[k])
			}
		} else { // hello
			if firstWord {
				buf.WriteByte(str[i])
				firstWord = false
			} else {
				buf.WriteByte(ToUpper(str[i]))
			}
			for k := i + 1; k < j; k++ {
				buf.WriteByte(str[k])
			}
		}
		i = j
	}

	return buf.String()
}

func PascalName(str string) string {
	var buf bytes.Buffer
	firstWord := true
	for i := 0; i < len(str); {
		for str[i] == '_' || str[i] == '-' {
			i++
		}
		j := i
		for j < len(str) && IsUpper(str[j]) {
			j++
		}
		if j-i < 2 { // 只有一个大写字母
			for ; j < len(str); j++ {
				if str[j] == '_' || str[j] == '-' || IsUpper(str[j]) {
					break
				}
			}
		} else {
			if j < len(str) && str[j] != '-' && str[j] != '_' {
				j--
			}
		}
		if j-i >= 2 && IsUpper(str[i]) && IsUpper(str[i+1]) { // HELLO
			if firstWord {
				firstWord = false
			}
			for k := i; k < j; k++ {
				buf.WriteByte(str[k])
			}
		} else if IsUpper(str[i]) { // Hello
			if firstWord {
				firstWord = false
			}
			for k := i; k < j; k++ {
				buf.WriteByte(str[k])
			}
		} else { // hello
			if firstWord {
				firstWord = false
			}
			buf.WriteByte(ToUpper(str[i]))
			for k := i + 1; k < j; k++ {
				buf.WriteByte(str[k])
			}
		}
		i = j
	}

	return buf.String()
}

func SnakeName(str string) string {
	return snakeName(str, '_')
}

func KebabName(str string) string {
	return snakeName(str, '-')
}

func SnakeNameAllCaps(str string) string {
	return strings.ToUpper(SnakeName(str))
}

func KebabNameAllCaps(str string) string {
	return strings.ToUpper(KebabName(str))
}

func snakeName(str string, separator uint8) string {
	var buf bytes.Buffer
	firstWord := true
	for i := 0; i < len(str); {
		for str[i] == '_' || str[i] == '-' {
			i++
		}
		j := i
		for j < len(str) && IsUpper(str[j]) {
			j++
		}
		if j-i < 2 { // 只有一个大写字母
			for ; j < len(str); j++ {
				if str[j] == '_' || str[j] == '-' || IsUpper(str[j]) {
					break
				}
			}
		} else {
			if j < len(str) && str[j] != '-' && str[j] != '_' {
				j--
			}
		}
		if j-i >= 2 && IsUpper(str[i]) && IsUpper(str[i+1]) { // HELLO
			if firstWord {
				firstWord = false
			} else {
				buf.WriteByte(separator)
			}
			for k := i; k < j; k++ {
				buf.WriteByte(ToLower(str[k]))
			}
		} else if IsUpper(str[i]) { // Hello
			if firstWord {
				firstWord = false
			} else {
				buf.WriteByte(separator)
			}
			buf.WriteByte(ToLower(str[i]))
			for k := i + 1; k < j; k++ {
				buf.WriteByte(str[k])
			}
		} else { // hello
			if firstWord {
				firstWord = false
			} else {
				buf.WriteByte(separator)
			}
			for k := i; k < j; k++ {
				buf.WriteByte(str[k])
			}
		}
		i = j
	}

	return buf.String()
}
