package parser

import (
	"testing"

	"github.com/pingcap/ticat/pkg/cli"
)

func TestCmdParserParseSeg(t *testing.T) {
	assert_eq := func(a []parsedSeg, b []parsedSeg) {
		fatal := func() {
			t.Fatalf("%#v != %#v\n", a, b)
		}
		if len(a) != len(b) {
			fatal()
		}
		for i, _ := range a {
			if a[i].Type != b[i].Type {
				fatal()
			}
			if a[i].Type == parsedSegTypeCmd {
				ac := a[i].Val.(MatchedCmd)
				bc := b[i].Val.(MatchedCmd)
				if ac.Name != bc.Name {
					fatal()
				}
			} else if a[i].Type == parsedSegTypeEnv {
				ae := a[i].Val.(ParsedEnv)
				be := b[i].Val.(ParsedEnv)
				if !ae.Equal(be) {
					fatal()
				}
			}
		}
	}

	root := cli.NewCmdTree()
	l2:= root.AddSub("X")
	l2.AddSub("21", "twenty-one")

	parser := &cmdParser{
		&envParser{&brackets{"{", "}"}},
		".",
		"\t\n\r./ ",
		"<root>",
	}

	cmd := func(name string) parsedSeg {
		return parsedSeg{parsedSegTypeCmd, MatchedCmd{name, nil}}
	}

	env := func(names ...string) parsedSeg {
		env := ParsedEnv{}
		for _, name := range names {
			env[name] = "V"
		}
		return parsedSeg{parsedSegTypeEnv, env}
	}

	test := func(a []string, b []parsedSeg) {
		parsed := parser.parse(root, a)
		assert_eq(parsed, b)
	}

	test(nil, []parsedSeg{})
	test([]string{}, []parsedSeg{})
	test([]string{}, []parsedSeg{})
	test([]string{"X"}, []parsedSeg{cmd("X")})
	test([]string{"X", "21"}, []parsedSeg{cmd("X"), cmd("21")})
	test([]string{"X", "twenty-one"}, []parsedSeg{cmd("X"), cmd("twenty-one")})
	test([]string{"{a=V}", "X", "21"}, []parsedSeg{env("a"), cmd("X"), cmd("21")})
	test([]string{"X", "{a=V}", "21"}, []parsedSeg{cmd("X"), env("a"), cmd("21")})
	test([]string{"X", "21", "{a=V}"}, []parsedSeg{cmd("X"), cmd("21"), env("a")})
	test([]string{"{a=V}{b=V}X{c=V}21{d=V}{e=V}"},
		[]parsedSeg{env("a"), env("b"), cmd("X"), env("c"), cmd("21"), env("d"), env("e")})
}

func TestCmdParserParse(t *testing.T) {
	assert_eq := func(a ParsedCmd, b ParsedCmd) {
		if len(a) != len(b) {
			t.Fatalf("size not eq: %#v != %#v\n", len(a), len(b))
		}
		for i, av := range a {
			bv := b[i]
			if (av.Env != nil) != (bv.Env != nil) {
				t.Fatalf("#%d env nil check not eq: %#v != %#v\n", i, av.Env != nil, bv.Env != nil)
			}
			if av.Env != nil && !av.Env.Equal(bv.Env) {
				t.Fatalf("#%d env not eq: %#v != %#v\n", i, av.Env, bv.Env)
			}
			if av.Cmd.Cmd != nil && av.Cmd.Cmd.Name() != bv.Cmd.Name {
				t.Fatalf("#%d cmd name not eq: '%#v' != '%#v'\n", i, av.Cmd.Cmd.Name(), bv.Cmd.Name)
			}
		}
	}

	root := cli.NewCmdTree()
	l2:= root.AddSub("X")
	l2.AddSub("21", "twenty-one")

	parser := &cmdParser{
		&envParser{&brackets{"{", "}"}},
		".",
		"\t\n\r./ ",
		"<root>",
	}

	seg := func(cmdName string, envKeyNames ...string) ParsedCmdSeg {
		var env ParsedEnv
		if len(envKeyNames) != 0 {
			env = ParsedEnv{}
			for _, name := range envKeyNames {
				env[name] = "V"
			}
		}
		return ParsedCmdSeg{env, MatchedCmd{cmdName, nil}}
	}

	test := func(a []string, b ParsedCmd) {
		parsed := parser.Parse(root, a)
		assert_eq(parsed, b)
	}

	test(nil, ParsedCmd{})
	test([]string{}, ParsedCmd{})
	test([]string{}, ParsedCmd{})

	test([]string{"X"}, ParsedCmd{seg("X")})
	test([]string{"X", "21"}, ParsedCmd{seg("X"), seg("21")})
	test([]string{"X", "twenty-one"}, ParsedCmd{seg("X"), seg("21")})
	test([]string{"{a=V}", "X", "21"}, ParsedCmd{seg("", "a"), seg("X"), seg("21")})
	test([]string{"X", "{a=V}", "21"}, ParsedCmd{seg("X", "X.a"), seg("21")})
	test([]string{"X", "21", "{a=V}"}, ParsedCmd{seg("X"), seg("21", "X.21.a")})
	test([]string{"X", "twenty-one", "{a=V}"}, ParsedCmd{seg("X"), seg("21", "X.21.a")})

	test([]string{"{a=V}{b=V}X{c=V}21{d=V}{e=V}"},
		ParsedCmd{seg("", "a", "b"), seg("X", "X.c"), seg("21", "X.21.d", "X.21.e")})
}
