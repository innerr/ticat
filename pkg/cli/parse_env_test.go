package cli

import (
	"fmt"
	"testing"
)

func TestEnvParserTryParseRaw(t *testing.T) {
	root := NewCmdTree()
	parser := &envParser{&brackets{"{", "}"}}

	test := func(a []string, bEnv ParsedEnv, bRest []string) {
		aEnv, aRest := parser.TryParseRaw(root, a)
		aRestStr := fmt.Sprintf("%#v", aRest)
		bRestStr := fmt.Sprintf("%#v", bRest)
		if !bEnv.Equal(aEnv) || aRestStr != bRestStr {
			t.Fatalf("%#v: (%#v, %#v) != (%#v, %#v)\n", a, aEnv, aRestStr, bEnv, bRestStr)
		}
	}

	test(nil, nil, nil)
	test([]string{}, nil, nil)
	test([]string{"a=A"}, ParsedEnv{"a": "A"}, nil)
	test([]string{"a=A", "b=B"}, ParsedEnv{"a": "A", "b": "B"}, nil)
	test([]string{"a=A", "bB"}, ParsedEnv{"a": "A"}, []string{"bB"})
	test([]string{"a=A", "bB", "c=C"}, ParsedEnv{"a": "A"}, []string{"bB", "c=C"})
}

func TestEnvParserFindLeft(t *testing.T) {
	parser := &envParser{&brackets{"{", "}"}}

	test := func(a []string, bRest []string, bFound bool, bAgain bool) {
		aRest, aFound, aAgain := parser.findLeft(a)
		aRestStr := fmt.Sprintf("%#v", aRest)
		bRestStr := fmt.Sprintf("%#v", bRest)

		if aFound != bFound || aAgain != bAgain || aRestStr != bRestStr {
			t.Fatalf(
				"%#v: %#v, %v, %v != %#v, %v, %v\n", a,
				aRestStr, aFound, aAgain,
				bRestStr, bFound, bAgain,
			)
		}
	}

	test(nil, nil, false, false)
	test([]string{}, nil, false, false)
	test([]string{"aaa"}, []string{"aaa"}, false, false)
	test([]string{"{"}, nil, true, false)
	test([]string{"{aaa"}, []string{"aaa"}, true, false)
	test([]string{"{aaa", "bbb"}, []string{"aaa", "bbb"}, true, false)
	test([]string{"aaa", "{", "bbb"}, []string{"aaa", "{", "bbb"}, false, false)
	test([]string{"aa{a", "bbb"}, []string{"aa", "{", "a", "bbb"}, true, true)

	test([]string{"{}A"}, []string{"}A"}, true, false)
	test([]string{"{}:"}, []string{"}:"}, true, false)
	test([]string{"{}:{}X"}, []string{"}:{}X"}, true, false)
	test([]string{"A{}:{}X"}, []string{"A", "{", "}:{}X"}, true, true)
}

func TestEnvParserFindRight(t *testing.T) {
	parser := &envParser{&brackets{"{", "}"}}

	test := func(a []string, bEnv []string, bRest []string, bFound bool) {
		aEnv, aRest, aFound := parser.findRight(a)
		aEnvStr := fmt.Sprintf("%#v", aEnv)
		bEnvStr := fmt.Sprintf("%#v", bEnv)
		aRestStr := fmt.Sprintf("%#v", aRest)
		bRestStr := fmt.Sprintf("%#v", bRest)

		if aFound != bFound || aEnvStr != bEnvStr || aRestStr != bRestStr {
			t.Fatalf(
				"%#v: %#v, %#v, %#v != %#v, %#v, %#v\n", a,
				aEnvStr, aRestStr, aFound,
				bEnvStr, bRestStr, bFound,
			)
		}
	}

	test([]string{"}A"}, nil, []string{"A"}, true)

	test([]string{"aaa"}, nil, nil, false)
	test([]string{"aaa", "{"}, nil, nil, false)
	test([]string{"}"}, nil, nil, true)
	test([]string{"}", "aaa"}, nil, []string{"aaa"}, true)
	test([]string{"aaa", "}"}, []string{"aaa"}, nil, true)
	test([]string{"a}"}, []string{"a"}, nil, true)
	test([]string{"a}bb"}, []string{"a"}, []string{"bb"}, true)
	test([]string{"a}bb", "}"}, []string{"a"}, []string{"bb", "}"}, true)
	test([]string{"a}bb", "cc"}, []string{"a"}, []string{"bb", "cc"}, true)

	test([]string{"a }"}, []string{"a"}, nil, true)
	test([]string{"a } bb"}, []string{"a"}, []string{"bb"}, true)
	test([]string{"a } bb", "}"}, []string{"a"}, []string{"bb", "}"}, true)
	test([]string{"a } bb", "cc"}, []string{"a"}, []string{"bb", "cc"}, true)

	test([]string{"a=A", "b=B}"}, []string{"a=A", "b=B"}, nil, true)

	test([]string{"}A"}, nil, []string{"A"}, true)
	test([]string{"}:"}, nil, []string{":"}, true)
	test([]string{"}:{}X"}, nil, []string{":{}X"}, true)
}

func TestEnvParserTryParse(t *testing.T) {
	root := NewCmdTree()
	parser := &envParser{&brackets{"{", "}"}}

	test := func(a []string, bEnv ParsedEnv, bRest []string, bFound bool, bErr error) {
		aEnv, aRest, aFound, aErr := parser.TryParse(root, a)
		aRestStr := fmt.Sprintf("%#v", aRest)
		bRestStr := fmt.Sprintf("%#v", bRest)
		if !bEnv.Equal(aEnv) || aRestStr != bRestStr || aFound != bFound || (bErr == nil) != (aErr == nil) {
			t.Fatalf("%#v: (%#v, %#v, %#v, %#v) != (%#v, %#v, %#v, %#v)\n", a,
				aEnv, aRestStr, aFound, aErr,
				bEnv, bRestStr, bFound, bErr,
			)
		}
	}

	test(nil, nil, nil, false, nil)
	test([]string{}, nil, nil, false, nil)
	test([]string{"{}"}, nil, nil, true, nil)
	test([]string{"{", "}"}, nil, nil, true, nil)
	test([]string{"{", "", "}"}, nil, nil, true, nil)
	test([]string{"{a=A}"}, ParsedEnv{"a": "A"}, nil, true, nil)
	test([]string{"{a=A}", "bb"}, ParsedEnv{"a": "A"}, []string{"bb"}, true, nil)
	test([]string{"{", "a=A", "}", "bb"}, ParsedEnv{"a": "A"}, []string{"bb"}, true, nil)
	test([]string{"11", "{a=A}", "bb"}, nil, []string{"11", "{a=A}", "bb"}, false, nil)

	test([]string{"{", "a=A", "}"}, ParsedEnv{"a": "A"}, nil, true, nil)
	test([]string{"{a=A", "}"}, ParsedEnv{"a": "A"}, nil, true, nil)
	test([]string{"{ a=A", "}"}, ParsedEnv{"a": "A"}, nil, true, nil)
	test([]string{"{ a =A", "}"}, ParsedEnv{"a": "A"}, nil, true, nil)
	test([]string{"{ a = A", "}"}, ParsedEnv{"a": "A"}, nil, true, nil)
	test([]string{"{", "a=A}"}, ParsedEnv{"a": "A"}, nil, true, nil)
	test([]string{"{", "a =A}"}, ParsedEnv{"a": "A"}, nil, true, nil)
	test([]string{"{", "a= A}"}, ParsedEnv{"a": "A"}, nil, true, nil)
	test([]string{"{", "a = A}"}, ParsedEnv{"a": "A"}, nil, true, nil)
	test([]string{"{", "a= A}"}, ParsedEnv{"a": "A"}, nil, true, nil)
	test([]string{"{", "a = A }"}, ParsedEnv{"a": "A"}, nil, true, nil)

	test([]string{"{", "a=A", "b=B", "}"}, ParsedEnv{"a": "A", "b": "B"}, nil, true, nil)
	test([]string{"{", "a=A", "b=B}"}, ParsedEnv{"a": "A", "b": "B"}, nil, true, nil)
	test([]string{"{a=A", "b=B}"}, ParsedEnv{"a": "A", "b": "B"}, nil, true, nil)
	test([]string{"{a=A", "b=B}", "cc", "dd"}, ParsedEnv{"a": "A", "b": "B"}, []string{"cc", "dd"}, true, nil)

	test([]string{"{a=A", "bB}"}, nil, []string{"{a=A", "bB}"}, true, fmt.Errorf("dumb"))
	test([]string{"{a=A", "bB", "c=C}"}, nil, []string{"{a=A", "bB", "c=C}"}, true, fmt.Errorf("dumb"))

	test([]string{"{}A"}, nil, []string{"A"}, true, nil)
	test([]string{"{}:"}, nil, []string{":"}, true, nil)
	test([]string{"{}:{}X"}, nil, []string{":{}X"}, true, nil)
}
