package cli

import (
	"testing"
)

func TestSequenceBreaker(t *testing.T) {
	assert_eq := func(a []string, b []string) {
		if len(a) != len(b) {
			t.Fatalf("%#v vs %#v\n", a, b)
		}
		for i, _ := range a {
			if a[i] != b[i] {
				t.Fatalf("%#v vs %#v\n", a, b)
			}
		}
	}

	test_normalize := func(a []string, b []string) {
		breaker := SequenceBreaker{":", []string{"http", "HTTP"}, []string{"/"}}
		assert_eq(breaker.Normalize(a), b)
	}

	test_normalize([]string{"aa"}, []string{"aa"})
	test_normalize([]string{"aa", "bb"}, []string{"aa", "bb"})
	test_normalize([]string{"aa", "bb", "cc"}, []string{"aa", "bb", "cc"})

	test_normalize([]string{":aa"}, []string{":", "aa"})
	test_normalize([]string{":aa", "bb", "cc"}, []string{":", "aa", "bb", "cc"})
	test_normalize([]string{"aa:", "bb", "cc"}, []string{"aa", ":", "bb", "cc"})
	test_normalize([]string{"aa", ":bb", "cc"}, []string{"aa", ":", "bb", "cc"})
	test_normalize([]string{"aa", "bb:", "cc"}, []string{"aa", "bb", ":", "cc"})
	test_normalize([]string{"aa", "bb", ":cc"}, []string{"aa", "bb", ":", "cc"})
	test_normalize([]string{"aa", "bb", "cc:"}, []string{"aa", "bb", "cc", ":"})

	test_normalize([]string{"a:x"}, []string{"a", ":", "x"})
	test_normalize([]string{"a:x", "bb", "cc"}, []string{"a", ":", "x", "bb", "cc"})
	test_normalize([]string{"aa", "b:x", "cc"}, []string{"aa", "b", ":", "x", "cc"})
	test_normalize([]string{"aa", "bb", "c:x"}, []string{"aa", "bb", "c", ":", "x"})

	test_normalize([]string{"aa", ":", "bb"}, []string{"aa", ":", "bb"})
	test_normalize([]string{"aa", ":", ":", "bb"}, []string{"aa", ":", ":", "bb"})
	test_normalize([]string{"aa", "::", "bb"}, []string{"aa", ":", ":", "bb"})
	test_normalize([]string{"aa:", "::", ":bb"}, []string{"aa", ":", ":", ":", ":", "bb"})

	test_normalize([]string{"::aa"}, []string{":", ":", "aa"})
	test_normalize([]string{"::aa", "bb", "cc"}, []string{":", ":", "aa", "bb", "cc"})
	test_normalize([]string{"aa::", "bb", "cc"}, []string{"aa", ":", ":", "bb", "cc"})
	test_normalize([]string{"aa", "::bb", "cc"}, []string{"aa", ":", ":", "bb", "cc"})
	test_normalize([]string{"aa", "bb::", "cc"}, []string{"aa", "bb", ":", ":", "cc"})
	test_normalize([]string{"aa", "bb", "::cc"}, []string{"aa", "bb", ":", ":", "cc"})
	test_normalize([]string{"aa", "bb", "cc::"}, []string{"aa", "bb", "cc", ":", ":"})

	test_normalize([]string{"aa:", ":bb", "cc"}, []string{"aa", ":", ":", "bb", "cc"})
	test_normalize([]string{"aa::", ":bb", "cc"}, []string{"aa", ":", ":", ":", "bb", "cc"})
	test_normalize([]string{"aa:", "::bb", "cc"}, []string{"aa", ":", ":", ":", "bb", "cc"})

	test_normalize([]string{"aa:", ":", ":bb", "cc"}, []string{"aa", ":", ":", ":", "bb", "cc"})
	test_normalize([]string{"aa::", ":", ":bb", "cc"}, []string{"aa", ":", ":", ":", ":", "bb", "cc"})
	test_normalize([]string{"aa:", ":", "::bb", "cc"}, []string{"aa", ":", ":", ":", ":", "bb", "cc"})

	test_normalize([]string{"http:?"}, []string{"http:?"})
	test_normalize([]string{"HTTP:?"}, []string{"HTTP:?"})
	test_normalize([]string{"HTTP://"}, []string{"HTTP://"})
	test_normalize([]string{"Http:?"}, []string{"Http", ":", "?"})

	assert_eq_v := func(a [][]string, b [][]string) {
		if len(a) != len(b) {
			t.Fatalf("%#v vs %#v\n", a, b)
		}
		for i, _ := range a {
			if len(a[i]) != len(b[i]) {
				t.Fatalf("%#v vs %#v\n", a, b)
				for j, _ := range a[i] {
					if len(a[i][j]) != len(b[i][j]) {
						t.Fatalf("%#v vs %#v\n", a, b)
					}
				}
			}
		}
	}

	test_break := func(a []string, b [][]string) {
		breaker := SequenceBreaker{":", []string{"http", "HTTP"}, []string{"/"}}
		assert_eq_v(breaker.Break(a), b)
	}

	test_break([]string{"aa"}, [][]string{[]string{"aa"}})
	test_break([]string{"aa", "bb"}, [][]string{[]string{"aa", "bb"}})
	test_break([]string{"aa", "bb", "cc"}, [][]string{[]string{"aa", "bb", "cc"}})

	test_break([]string{":aa"}, [][]string{[]string{"aa"}})
	test_break([]string{":aa", "bb", "cc"}, [][]string{[]string{"aa", "bb", "cc"}})
	test_break([]string{"aa:", "bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})
	test_break([]string{"aa", ":bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})
	test_break([]string{"aa", "bb:", "cc"}, [][]string{[]string{"aa", "bb"}, []string{"cc"}})
	test_break([]string{"aa", "bb", ":cc"}, [][]string{[]string{"aa", "bb"}, []string{"cc"}})
	test_break([]string{"aa", "bb", "cc:"}, [][]string{[]string{"aa", "bb", "cc"}})

	test_break([]string{"a:x"}, [][]string{[]string{"a"}, []string{"x"}})
	test_break([]string{"a:x", "bb", "cc"}, [][]string{[]string{"a"}, []string{"x", "bb", "cc"}})
	test_break([]string{"aa", "b:x", "cc"}, [][]string{[]string{"aa", "b"}, []string{"x", "cc"}})
	test_break([]string{"aa", "bb", "c:x"}, [][]string{[]string{"aa", "bb", "c"}, []string{"x"}})

	test_break([]string{"aa", ":", "bb"}, [][]string{[]string{"aa"}, []string{"bb"}})
	test_break([]string{"aa", ":", ":", "bb"}, [][]string{[]string{"aa"}, []string{"bb"}})
	test_break([]string{"aa", "::", "bb"}, [][]string{[]string{"aa"}, []string{"bb"}})
	test_break([]string{"aa:", "::", ":bb"}, [][]string{[]string{"aa"}, []string{"bb"}})

	test_break([]string{"::aa"}, [][]string{[]string{"aa"}})
	test_break([]string{"::aa", "bb", "cc"}, [][]string{[]string{"aa", "bb", "cc"}})
	test_break([]string{"aa::", "bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})
	test_break([]string{"aa", "::bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})
	test_break([]string{"aa", "bb::", "cc"}, [][]string{[]string{"aa", "bb"}, []string{"cc"}})
	test_break([]string{"aa", "bb", "::cc"}, [][]string{[]string{"aa", "bb"}, []string{"cc"}})
	test_break([]string{"aa", "bb", "cc::"}, [][]string{[]string{"aa", "bb", "cc"}})

	test_break([]string{"aa:", ":bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})
	test_break([]string{"aa::", ":bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})
	test_break([]string{"aa:", "::bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})

	test_break([]string{"aa:", ":", ":bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})
	test_break([]string{"aa::", ":", ":bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})
	test_break([]string{"aa:", ":", "::bb", "cc"}, [][]string{[]string{"aa"}, []string{"bb", "cc"}})

	test_break([]string{"http:?"}, [][]string{[]string{"http:?"}})
	test_break([]string{"HTTP:?"}, [][]string{[]string{"HTTP:?"}})
	test_break([]string{"HTTP://"}, [][]string{[]string{"HTTP://"}})
	test_break([]string{"Http:?"}, [][]string{[]string{"Http"}, []string{"?"}})
}
