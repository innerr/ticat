package cli

import (
	"testing"
)

func TestSequenceParserNormalize(t *testing.T) {
	assert_eq := func(a []string, b []string) {
		fatal := func() {
			t.Fatalf("%#v != %#v\n", a, b)
		}
		if len(a) != len(b) {
			fatal()
		}
		for i, _ := range a {
			if a[i] != b[i] {
				fatal()
			}
		}
	}

	parser := sequenceParser{":", []string{"http", "HTTP"}, []string{"/"}}
	test := func(a []string, b []string) {
		assert_eq(parser.Normalize(a), b)
	}

	test([]string{"aa"}, []string{"aa"})
	test([]string{"aa", "bb"}, []string{"aa", "bb"})
	test([]string{"aa", "bb", "cc"}, []string{"aa", "bb", "cc"})

	test([]string{":aa"}, []string{":", "aa"})
	test([]string{":aa", "bb", "cc"}, []string{":", "aa", "bb", "cc"})
	test([]string{"aa:", "bb", "cc"}, []string{"aa", ":", "bb", "cc"})
	test([]string{"aa", ":bb", "cc"}, []string{"aa", ":", "bb", "cc"})
	test([]string{"aa", "bb:", "cc"}, []string{"aa", "bb", ":", "cc"})
	test([]string{"aa", "bb", ":cc"}, []string{"aa", "bb", ":", "cc"})
	test([]string{"aa", "bb", "cc:"}, []string{"aa", "bb", "cc", ":"})

	test([]string{"a:x"}, []string{"a", ":", "x"})
	test([]string{"a:x", "bb", "cc"}, []string{"a", ":", "x", "bb", "cc"})
	test([]string{"aa", "b:x", "cc"}, []string{"aa", "b", ":", "x", "cc"})
	test([]string{"aa", "bb", "c:x"}, []string{"aa", "bb", "c", ":", "x"})

	test([]string{"aa", ":", "bb"}, []string{"aa", ":", "bb"})
	test([]string{"aa", ":", ":", "bb"}, []string{"aa", ":", ":", "bb"})
	test([]string{"aa", "::", "bb"}, []string{"aa", ":", ":", "bb"})
	test([]string{"aa:", "::", ":bb"}, []string{"aa", ":", ":", ":", ":", "bb"})

	test([]string{"::aa"}, []string{":", ":", "aa"})
	test([]string{"::aa", "bb", "cc"}, []string{":", ":", "aa", "bb", "cc"})
	test([]string{"aa::", "bb", "cc"}, []string{"aa", ":", ":", "bb", "cc"})
	test([]string{"aa", "::bb", "cc"}, []string{"aa", ":", ":", "bb", "cc"})
	test([]string{"aa", "bb::", "cc"}, []string{"aa", "bb", ":", ":", "cc"})
	test([]string{"aa", "bb", "::cc"}, []string{"aa", "bb", ":", ":", "cc"})
	test([]string{"aa", "bb", "cc::"}, []string{"aa", "bb", "cc", ":", ":"})

	test([]string{"aa:", ":bb", "cc"}, []string{"aa", ":", ":", "bb", "cc"})
	test([]string{"aa::", ":bb", "cc"}, []string{"aa", ":", ":", ":", "bb", "cc"})
	test([]string{"aa:", "::bb", "cc"}, []string{"aa", ":", ":", ":", "bb", "cc"})

	test([]string{"aa:", ":", ":bb", "cc"}, []string{"aa", ":", ":", ":", "bb", "cc"})
	test([]string{"aa::", ":", ":bb", "cc"}, []string{"aa", ":", ":", ":", ":", "bb", "cc"})
	test([]string{"aa:", ":", "::bb", "cc"}, []string{"aa", ":", ":", ":", ":", "bb", "cc"})

	test([]string{"http:?"}, []string{"http:?"})
	test([]string{"HTTP:?"}, []string{"HTTP:?"})
	test([]string{"HTTP://"}, []string{"HTTP://"})
	test([]string{"Http:?"}, []string{"Http", ":", "?"})
}

func TestSequenceParserBreak(t *testing.T) {
	assert_eq := func(a [][]string, b [][]string) {
		fatal := func() {
			t.Fatalf("%#v != %#v\n", a, b)
		}
		if len(a) != len(b) {
			fatal()
		}
		for i, _ := range a {
			if len(a[i]) != len(b[i]) {
				fatal()
			}
			for j, _ := range a[i] {
				if len(a[i][j]) != len(b[i][j]) {
					fatal()
				}
			}
		}
	}

	parser := sequenceParser{":", []string{"http", "HTTP"}, []string{"/"}}
	test := func(a []string, b [][]string) {
		parsed, _ := parser.Parse(a)
		assert_eq(parsed, b)
	}

	test([]string{"aa"}, [][]string{[]string{"aa"}})
	test([]string{"aa", "bb"}, [][]string{[]string{"aa", "bb"}})
	test([]string{"aa", "bb", "cc"}, [][]string{[]string{"aa", "bb", "cc"}})

	test([]string{":aa"}, [][]string{[]string{"aa"}})
	test([]string{":aa", "bb", "cc"}, [][]string{[]string{"aa", "bb", "cc"}})
	test([]string{"aa:", "bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})
	test([]string{"aa", ":bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})
	test([]string{"aa", "bb:", "cc"}, [][]string{[]string{"aa", "bb"}, []string{"cc"}})
	test([]string{"aa", "bb", ":cc"}, [][]string{[]string{"aa", "bb"}, []string{"cc"}})
	test([]string{"aa", "bb", "cc:"}, [][]string{[]string{"aa", "bb", "cc"}})

	test([]string{"a:x"}, [][]string{[]string{"a"}, []string{"x"}})
	test([]string{"a:x", "bb", "cc"}, [][]string{[]string{"a"}, []string{"x", "bb", "cc"}})
	test([]string{"aa", "b:x", "cc"}, [][]string{[]string{"aa", "b"}, []string{"x", "cc"}})
	test([]string{"aa", "bb", "c:x"}, [][]string{[]string{"aa", "bb", "c"}, []string{"x"}})

	test([]string{"aa", ":", "bb"}, [][]string{[]string{"aa"}, []string{"bb"}})
	test([]string{"aa", ":", ":", "bb"}, [][]string{[]string{"aa"}, []string{"bb"}})
	test([]string{"aa", "::", "bb"}, [][]string{[]string{"aa"}, []string{"bb"}})
	test([]string{"aa:", "::", ":bb"}, [][]string{[]string{"aa"}, []string{"bb"}})

	test([]string{"::aa"}, [][]string{[]string{"aa"}})
	test([]string{"::aa", "bb", "cc"}, [][]string{[]string{"aa", "bb", "cc"}})
	test([]string{"aa::", "bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})
	test([]string{"aa", "::bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})
	test([]string{"aa", "bb::", "cc"}, [][]string{[]string{"aa", "bb"}, []string{"cc"}})
	test([]string{"aa", "bb", "::cc"}, [][]string{[]string{"aa", "bb"}, []string{"cc"}})
	test([]string{"aa", "bb", "cc::"}, [][]string{[]string{"aa", "bb", "cc"}})

	test([]string{"aa:", ":bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})
	test([]string{"aa::", ":bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})
	test([]string{"aa:", "::bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})

	test([]string{"aa:", ":", ":bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})
	test([]string{"aa::", ":", ":bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})
	test([]string{"aa:", ":", "::bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})

	test([]string{"http:?"}, [][]string{[]string{"http:?"}})
	test([]string{"HTTP:?"}, [][]string{[]string{"HTTP:?"}})
	test([]string{"HTTP://"}, [][]string{[]string{"HTTP://"}})
	test([]string{"Http:?"}, [][]string{[]string{"Http"}, []string{"?"}})
}
