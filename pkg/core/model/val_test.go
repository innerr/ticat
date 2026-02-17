package model

import (
	"testing"
)

func TestStrToTrue(t *testing.T) {
	trueValues := []string{"true", "TRUE", "True", "t", "T", "1", "on", "ON", "y", "Y", "yes", "YES", "Yes"}
	for _, v := range trueValues {
		if !StrToTrue(v) {
			t.Errorf("StrToTrue(%s) should return true", v)
		}
	}

	falseValues := []string{"false", "FALSE", "f", "F", "0", "off", "OFF", "n", "N", "no", "NO", "random", ""}
	for _, v := range falseValues {
		if StrToTrue(v) {
			t.Errorf("StrToTrue(%s) should return false", v)
		}
	}
}

func TestStrToFalse(t *testing.T) {
	falseValues := []string{"false", "FALSE", "False", "f", "F", "0", "off", "OFF", "n", "N", "no", "NO", "No"}
	for _, v := range falseValues {
		if !StrToFalse(v) {
			t.Errorf("StrToFalse(%s) should return true", v)
		}
	}

	trueValues := []string{"true", "TRUE", "t", "T", "1", "on", "y", "yes", "random", ""}
	for _, v := range trueValues {
		if StrToFalse(v) {
			t.Errorf("StrToFalse(%s) should return false", v)
		}
	}
}

func TestStrToBool(t *testing.T) {
	trueValues := []string{"true", "TRUE", "t", "T", "1", "on", "y", "yes"}
	for _, v := range trueValues {
		if !StrToBool(v) {
			t.Errorf("StrToBool(%s) should return true", v)
		}
	}

	falseValues := []string{"false", "f", "0", "off", "n", "no", "random", ""}
	for _, v := range falseValues {
		if StrToBool(v) {
			t.Errorf("StrToBool(%s) should return false", v)
		}
	}
}

func TestStrToTrueCaseInsensitive(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"TRUE", true},
		{"true", true},
		{"TrUe", true},
		{"T", true},
		{"t", true},
		{"1", true},
		{"ON", true},
		{"on", true},
		{"Y", true},
		{"y", true},
		{"YES", true},
		{"yes", true},
		{"FALSE", false},
		{"false", false},
		{"0", false},
		{"OFF", false},
		{"off", false},
		{"N", false},
		{"n", false},
		{"NO", false},
		{"no", false},
	}

	for _, test := range tests {
		result := StrToTrue(test.input)
		if result != test.expected {
			t.Errorf("StrToTrue(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestStrToFalseCaseInsensitive(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"FALSE", true},
		{"false", true},
		{"FaLsE", true},
		{"F", true},
		{"f", true},
		{"0", true},
		{"OFF", true},
		{"off", true},
		{"N", true},
		{"n", true},
		{"NO", true},
		{"no", true},
		{"TRUE", false},
		{"true", false},
		{"1", false},
		{"ON", false},
		{"on", false},
		{"Y", false},
		{"y", false},
		{"YES", false},
		{"yes", false},
	}

	for _, test := range tests {
		result := StrToFalse(test.input)
		if result != test.expected {
			t.Errorf("StrToFalse(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestStrToBoolEdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{" ", false},
		{"2", false},
		{"-1", false},
		{"null", false},
		{"none", false},
		{"enabled", false},
		{"disabled", false},
	}

	for _, test := range tests {
		result := StrToBool(test.input)
		if result != test.expected {
			t.Errorf("StrToBool(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestStrToTrueWhitespace(t *testing.T) {
	// Whitespace should not be trimmed - these should return false
	whitespaceValues := []string{" true ", "true ", " true", "\ttrue", "true\n"}
	for _, v := range whitespaceValues {
		if StrToTrue(v) {
			t.Errorf("StrToTrue(%q) should return false (whitespace not trimmed)", v)
		}
	}
}

func TestStrToFalseWhitespace(t *testing.T) {
	// Whitespace should not be trimmed - these should return false
	whitespaceValues := []string{" false ", "false ", " false", "\tfalse", "false\n"}
	for _, v := range whitespaceValues {
		if StrToFalse(v) {
			t.Errorf("StrToFalse(%q) should return false (whitespace not trimmed)", v)
		}
	}
}

func TestStrToBoolConsistency(t *testing.T) {
	// StrToBool should be equivalent to StrToTrue
	testValues := []string{"true", "false", "1", "0", "on", "off", "yes", "no", "y", "n", "t", "f", "", "random"}
	for _, v := range testValues {
		boolResult := StrToBool(v)
		trueResult := StrToTrue(v)
		if boolResult != trueResult {
			t.Errorf("StrToBool(%q) = %v, StrToTrue(%q) = %v - should be equal", v, boolResult, v, trueResult)
		}
	}
}
