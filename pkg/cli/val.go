package cli

func StrToBool(str string) bool {
	return str == "true" || str == "True" || str == "t" || str == "T" || str == "TRUE" || str == "1" ||
		str == "on" || str == "On" || str == "ON" || str == "y" || str == "Y"
}
