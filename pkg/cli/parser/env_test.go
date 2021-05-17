package parser

import (
	"fmt"
	"testing"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func TestEnvParserTryParseRaw(t *testing.T) {
	root := core.NewCmdTree(&core.CmdTreeStrs{"<root>", ".", ".", "|", "-", "--", "=", ".", "\t"})
	parser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", "."}

	test := func(a []string, bEnv core.ParsedEnv, bRest []string) {
		aEnv, aRest := parser.TryParseRaw(root, nil, a)
		aRestStr := fmt.Sprintf("%#v", aRest)
		bRestStr := fmt.Sprintf("%#v", bRest)
		if !bEnv.Equal(aEnv) || aRestStr != bRestStr {
			t.Fatalf("%#v: (%#v, %#v) != (%#v, %#v)\n", a, aEnv, aRestStr, bEnv, bRestStr)
		}
	}

	v := func(v string) core.ParsedEnvVal {
		return core.ParsedEnvVal{v, false, nil}
	}
	a := func(v string) core.ParsedEnvVal {
		return core.ParsedEnvVal{v, true, nil}
	}

	test(nil, nil, nil)
	test([]string{}, nil, nil)
	test([]string{"a=A"}, core.ParsedEnv{"a": v("A")}, nil)
	test([]string{"a=A", "b=B"}, core.ParsedEnv{"a": v("A"), "b": v("B")}, nil)
	test([]string{"a=A", "bB"}, core.ParsedEnv{"a": v("A")}, []string{"bB"})
	test([]string{"a=A", "bB", "c=C"}, core.ParsedEnv{"a": v("A")}, []string{"bB", "c=C"})
	test([]string{" a = A "}, core.ParsedEnv{"a": v("A")}, nil)
	test([]string{" a = A ", " b = B "}, core.ParsedEnv{"a": v("A"), "b": v("B")}, nil)
	test([]string{" a = A ", " bB "}, core.ParsedEnv{"a": v("A")}, []string{" bB "})
	test([]string{" a = A ", " bB ", " c = C "}, core.ParsedEnv{"a": v("A")}, []string{" bB ", " c = C "})
	test([]string{"a", "=", "A"}, core.ParsedEnv{"a": v("A")}, nil)
	test([]string{"a=", "A"}, core.ParsedEnv{"a": v("A")}, nil)
	test([]string{"a", "=A"}, core.ParsedEnv{"a": v("A")}, nil)
	test([]string{" a ", " = ", " A "}, core.ParsedEnv{"a": v("A")}, nil)
	test([]string{" a = ", " A "}, core.ParsedEnv{"a": v("A")}, nil)
	test([]string{" a ", " = A "}, core.ParsedEnv{"a": v("A")}, nil)

	dummy := func(core.ArgVals, *core.Cli, *core.Env) bool {
		return true
	}
	cmd := root.RegCmd(dummy, "")

	test(nil, nil, nil)
	test([]string{}, nil, nil)
	test([]string{"a=A"}, nil, []string{"a=A"})

	cmd.AddArg("aa", "da")

	test(nil, nil, nil)
	test([]string{}, nil, nil)
	test([]string{"a=A"}, nil, []string{"a=A"})
	test([]string{"aa=A"}, core.ParsedEnv{"aa": a("A")}, nil)
	test([]string{"aa=A", "aa=A"}, core.ParsedEnv{"aa": a("A")}, nil)
	test([]string{"aa=A", "aa=B"}, core.ParsedEnv{"aa": a("B")}, nil)
	test([]string{"bb=A", "aa=B"}, nil, []string{"bb=A", "aa=B"})
	test([]string{"aa A"}, core.ParsedEnv{"aa": a("aa A")}, nil)
	test([]string{"aa", "A"}, core.ParsedEnv{"aa": a("A")}, nil)
	test([]string{"", "aa=A", ""}, core.ParsedEnv{"aa": a("A")}, nil)
	test([]string{" aa = A ", " aa = A "}, core.ParsedEnv{"aa": a("A")}, nil)
	test([]string{" aa = A", " aa = B "}, core.ParsedEnv{"aa": a("B")}, nil)
	test([]string{"aa", "=", "A"}, core.ParsedEnv{"aa": a("A")}, nil)
	test([]string{"aa=", "A"}, core.ParsedEnv{"aa": a("A")}, nil)
	test([]string{"aa", "=A"}, core.ParsedEnv{"aa": a("A")}, nil)
	test([]string{" aa ", " = ", " A "}, core.ParsedEnv{"aa": a("A")}, nil)
	test([]string{" aa = ", " A "}, core.ParsedEnv{"aa": a("A")}, nil)
	test([]string{" aa ", " = A "}, core.ParsedEnv{"aa": a("A")}, nil)

	cmd.AddArg("bb", "db", "BB")

	test(nil, nil, nil)
	test([]string{}, nil, nil)
	test([]string{"a=A"}, nil, []string{"a=A"})
	test([]string{"aa=A"}, core.ParsedEnv{"aa": a("A")}, nil)
	test([]string{"aa=A", "aa=A", "BB=B"}, core.ParsedEnv{"aa": a("A"), "bb": a("B")}, nil)
	test([]string{"aa=A", "aa=B"}, core.ParsedEnv{"aa": a("B")}, nil)
	test([]string{"aa=A", "bb=B"}, core.ParsedEnv{"aa": a("A"), "bb": a("B")}, nil)
	test([]string{"bb=A", "aa=B"}, core.ParsedEnv{"bb": a("A"), "aa": a("B")}, nil)
	test([]string{"BB=A", "aa=B"}, core.ParsedEnv{"bb": a("A"), "aa": a("B")}, nil)
	test([]string{"aa=A", "BB=B"}, core.ParsedEnv{"aa": a("A"), "bb": a("B")}, nil)
	test([]string{"aa=A", "x", "BB=B"}, core.ParsedEnv{"aa": a("A")}, []string{"x", "BB=B"})
	test([]string{"aa=A, BB=B"}, core.ParsedEnv{"aa": a("A, BB=B")}, nil)
	test([]string{"aa=A, x, BB=B"}, core.ParsedEnv{"aa": a("A, x, BB=B")}, nil)
	test([]string{"aa=A, x, BB=B", "bb=C"}, core.ParsedEnv{"aa": a("A, x, BB=B"), "bb": a("C")}, nil)
	test([]string{"aa", "A"}, core.ParsedEnv{"aa": a("A")}, nil)
	test([]string{"A", "B"}, core.ParsedEnv{"aa": a("A"), "bb": a("B")}, nil)
	test([]string{"aa", "A", "bb", "B"}, core.ParsedEnv{"aa": a("A"), "bb": a("B")}, nil)
	test([]string{"aa", "A", "BB", "B"}, core.ParsedEnv{"aa": a("A"), "bb": a("B")}, nil)
	test([]string{"bb", "B", "aa", "A"}, core.ParsedEnv{"aa": a("A"), "bb": a("B")}, nil)
	test([]string{"BB", "B", "aa", "A"}, core.ParsedEnv{"aa": a("A"), "bb": a("B")}, nil)

	test([]string{" aa = A ", " aa = A ", " BB = B "}, core.ParsedEnv{"aa": a("A"), "bb": a("B")}, nil)
	test([]string{" aa = A ", " x ", " BB = B "}, core.ParsedEnv{"aa": a("A")}, []string{" x ", " BB = B "})
	test([]string{"aa = A, x, BB=B"}, core.ParsedEnv{"aa": a("A, x, BB=B")}, nil)
	test([]string{"aa = A, x, BB=B", "bb=C"}, core.ParsedEnv{"aa": a("A, x, BB=B"), "bb": a("C")}, nil)
	test([]string{" aa ", " A "}, core.ParsedEnv{"aa": a("A")}, nil)
	test([]string{" A ", "", " B "}, core.ParsedEnv{"aa": a("A"), "bb": a("B")}, nil)
	test([]string{" aa ", "", " A ", "", " bb ", "", " B "}, core.ParsedEnv{"aa": a("A"), "bb": a("B")}, nil)
}

func TestEnvParserFindLeft(t *testing.T) {
	parser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", "."}

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
	parser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", "."}

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
	root := core.NewCmdTree(&core.CmdTreeStrs{"<root>", ".", ".", "|", "-", "--", "=", ".", "\t"})
	parser := &EnvParser{Brackets{"{", "}"}, "\t ", "=", "."}

	test := func(a []string, bEnv core.ParsedEnv, bRest []string, bFound bool, bErr error) {
		aEnv, aRest, aFound, aErr := parser.TryParse(root, nil, a)
		aRestStr := fmt.Sprintf("%#v", aRest)
		bRestStr := fmt.Sprintf("%#v", bRest)
		if !bEnv.Equal(aEnv) || aRestStr != bRestStr || aFound != bFound || (bErr == nil) != (aErr == nil) {
			t.Fatalf("%#v: (%#v, %#v, %#v, %#v) != (%#v, %#v, %#v, %#v)\n", a,
				aEnv, aRestStr, aFound, aErr,
				bEnv, bRestStr, bFound, bErr,
			)
		}
	}

	v := func(v string) core.ParsedEnvVal {
		return core.ParsedEnvVal{v, false, nil}
	}

	test(nil, nil, nil, false, nil)
	test([]string{}, nil, nil, false, nil)
	test([]string{"{}"}, nil, nil, true, nil)
	test([]string{"{", "}"}, nil, nil, true, nil)
	test([]string{"{", "", "}"}, nil, nil, true, nil)
	test([]string{"{a=A}"}, core.ParsedEnv{"a": v("A")}, nil, true, nil)
	test([]string{"{a=A}", "bb"}, core.ParsedEnv{"a": v("A")}, []string{"bb"}, true, nil)
	test([]string{"{", "a=A", "}", "bb"}, core.ParsedEnv{"a": v("A")}, []string{"bb"}, true, nil)
	test([]string{"11", "{a=A}", "bb"}, nil, []string{"11", "{a=A}", "bb"}, false, nil)

	test([]string{"{", "a=A", "}"}, core.ParsedEnv{"a": v("A")}, nil, true, nil)
	test([]string{"{a=A", "}"}, core.ParsedEnv{"a": v("A")}, nil, true, nil)
	test([]string{"{ a=A", "}"}, core.ParsedEnv{"a": v("A")}, nil, true, nil)
	test([]string{"{ a =A", "}"}, core.ParsedEnv{"a": v("A")}, nil, true, nil)
	test([]string{"{ a = A", "}"}, core.ParsedEnv{"a": v("A")}, nil, true, nil)
	test([]string{"{", "a=A}"}, core.ParsedEnv{"a": v("A")}, nil, true, nil)
	test([]string{"{", "a =A}"}, core.ParsedEnv{"a": v("A")}, nil, true, nil)
	test([]string{"{", "a= A}"}, core.ParsedEnv{"a": v("A")}, nil, true, nil)
	test([]string{"{", "a = A}"}, core.ParsedEnv{"a": v("A")}, nil, true, nil)
	test([]string{"{", "a= A}"}, core.ParsedEnv{"a": v("A")}, nil, true, nil)
	test([]string{"{", "a = A }"}, core.ParsedEnv{"a": v("A")}, nil, true, nil)

	test([]string{"{", "a=A", "b=B", "}"}, core.ParsedEnv{"a": v("A"), "b": v("B")}, nil, true, nil)
	test([]string{"{", "a=A", "b=B}"}, core.ParsedEnv{"a": v("A"), "b": v("B")}, nil, true, nil)
	test([]string{"{a=A", "b=B}"}, core.ParsedEnv{"a": v("A"), "b": v("B")}, nil, true, nil)
	test([]string{"{a=A", "b=B}", "cc", "dd"}, core.ParsedEnv{"a": v("A"), "b": v("B")}, []string{"cc", "dd"}, true, nil)

	test([]string{"{a=A", "bB}"}, nil, []string{"{a=A", "bB}"}, true, fmt.Errorf("dumb"))
	test([]string{"{a=A", "bB", "c=C}"}, nil, []string{"{a=A", "bB", "c=C}"}, true, fmt.Errorf("dumb"))

	test([]string{"{}A"}, nil, []string{"A"}, true, nil)
	test([]string{"{}:"}, nil, []string{":"}, true, nil)
	test([]string{"{}:{}X"}, nil, []string{":{}X"}, true, nil)
}
