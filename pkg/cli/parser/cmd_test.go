package parser

import (
	"testing"

	"github.com/pingcap/ticat/pkg/cli"
)

func TestCmdParserParseSeg(t *testing.T) {
	assertEq := func(input []string, a []parsedSeg, b []parsedSeg) {
		fatal := func() {
			t.Fatalf("%#v: %#v != %#v\n", input, a, b)
		}
		if len(a) != len(b) {
			fatal()
		}
		for i, _ := range a {
			if a[i].Type != b[i].Type {
				fatal()
			}
			if a[i].Type == parsedSegTypeCmd {
				ac := a[i].Val.(cli.MatchedCmd)
				bc := b[i].Val.(cli.MatchedCmd)
				if ac.Name != bc.Name {
					fatal()
				}
			} else if a[i].Type == parsedSegTypeEnv {
				ae := a[i].Val.(cli.ParsedEnv)
				be := b[i].Val.(cli.ParsedEnv)
				if !ae.Equal(be) {
					fatal()
				}
			} else if a[i].Type == parsedSegTypeSep {
				if b[i].Type != parsedSegTypeSep {
					fatal()
				}
			}
		}
	}

	root := cli.NewCmdTree("<root>", ".")
	l2 := root.AddSub("X")
	l2.AddSub("21", "twenty-one")

	parser := &cmdParser{
		&envParser{&brackets{"{", "}"}, "\t ", "=", "."},
		".", "\t ./", "\t ", "<root>",
	}

	sep := parsedSeg{parsedSegTypeSep, nil}

	cmd := func(name string) parsedSeg {
		return parsedSeg{parsedSegTypeCmd, cli.MatchedCmd{name, nil}}
	}
	env := func(names ...string) parsedSeg {
		env := cli.ParsedEnv{}
		for _, name := range names {
			env[name] = cli.ParsedEnvVal{"V", false}
		}
		return parsedSeg{parsedSegTypeEnv, env}
	}

	test := func(a []string, b []parsedSeg) {
		parsed := parser.parse(root, nil, a)
		assertEq(a, parsed, b)
	}

	test(nil, []parsedSeg{})
	test([]string{}, []parsedSeg{})
	test([]string{}, []parsedSeg{})

	test([]string{"X"}, []parsedSeg{cmd("X")})
	test([]string{"X", "21"}, []parsedSeg{cmd("X"), cmd("21")})
	test([]string{"X.21"}, []parsedSeg{cmd("X"), sep, cmd("21")})
	test([]string{"X/21"}, []parsedSeg{cmd("X"), sep, cmd("21")})
	test([]string{"X 21"}, []parsedSeg{cmd("X"), sep, cmd("21")})
	test([]string{"X \t 21"}, []parsedSeg{cmd("X"), sep, cmd("21")})
	test([]string{"X", "twenty-one"}, []parsedSeg{cmd("X"), cmd("twenty-one")})

	test([]string{"{a=V}", "X", "21"}, []parsedSeg{env("a"), cmd("X"), cmd("21")})
	test([]string{"X", "{a=V}", "21"}, []parsedSeg{cmd("X"), env("a"), cmd("21")})
	test([]string{"X", "21", "{a=V}"}, []parsedSeg{cmd("X"), cmd("21"), env("a")})

	test([]string{"X{a=V}/21"}, []parsedSeg{cmd("X"), env("a"), sep, cmd("21")})
	test([]string{"X{a=V}/{b=V}21"}, []parsedSeg{cmd("X"), env("a"), sep, env("b"), cmd("21")})
	test([]string{"X{a=V}./{b=V}21"}, []parsedSeg{cmd("X"), env("a"), sep, env("b"), cmd("21")})
	test([]string{"X{a=V} / / {b=V}21"}, []parsedSeg{cmd("X"), env("a"), sep, env("b"), cmd("21")})

	test([]string{"X/{a=V}21"}, []parsedSeg{cmd("X"), sep, env("a"), cmd("21")})
	test([]string{"X{a=V}21"}, []parsedSeg{cmd("X"), env("a"), cmd("21")})
	test([]string{"X/ {a=V}21"}, []parsedSeg{cmd("X"), sep, env("a"), cmd("21")})
	test([]string{"X /{a=V}21"}, []parsedSeg{cmd("X"), sep, env("a"), cmd("21")})
	test([]string{"X / {a=V} 21"}, []parsedSeg{cmd("X"), sep, env("a"), cmd("21")})
	test([]string{"{a=V}{b=V}X{c=V}21{d=V}{e=V}"},
		[]parsedSeg{env("a"), env("b"), cmd("X"), env("c"), cmd("21"), env("d"), env("e")})

	test([]string{"{}{}X{}"}, []parsedSeg{cmd("X")})
}

func TestCmdParserParse(t *testing.T) {
	assertEq := func(a cli.ParsedCmd, b cli.ParsedCmd) {
		for i, av := range a {
			if i >= len(b) {
				t.Fatalf("size not eq: %#v != %#v\n", len(a), len(b))
			}
			bv := b[i]
			if len(bv.Cmd.Name) != 0 && av.Cmd.Cmd == nil {
				t.Fatalf("#%d cmd not eq: '%#v' != '%#v'\n", i, av.Cmd, bv.Cmd)
			}
			if av.Cmd.Cmd != nil && av.Cmd.Cmd.Name() != bv.Cmd.Name {
				t.Fatalf("#%d cmd name not eq: '%#v' != '%#v'\n", i, av.Cmd.Cmd.Name(), bv.Cmd.Name)
			}
			if (av.Env != nil) != (bv.Env != nil) {
				t.Fatalf("#%d env nil check not eq: %#v != %#v\n", i, av.Env, bv.Env)
			}
			if av.Env != nil && !av.Env.Equal(bv.Env) {
				t.Fatalf("#%d env not eq: %#v != %#v\n", i, av.Env, bv.Env)
			}
		}
	}

	root := cli.NewCmdTree("<root>", ".")
	l2 := root.AddSub("X")
	l2.AddSub("21", "twenty-one")

	parser := &cmdParser{
		&envParser{&brackets{"{", "}"}, "\t ", "=", "."},
		".", "\t ./", "\t ", "<root>",
	}

	seg := func(cmdName string, envKeyNames ...string) cli.ParsedCmdSeg {
		var env cli.ParsedEnv
		if len(envKeyNames) != 0 {
			env = cli.ParsedEnv{}
			for _, name := range envKeyNames {
				env[name] = cli.ParsedEnvVal{"V", false}
			}
		}
		return cli.ParsedCmdSeg{env, cli.MatchedCmd{cmdName, nil}}
	}

	test := func(a []string, b cli.ParsedCmd) {
		parsed := parser.Parse(root, nil, a)
		assertEq(parsed, b)
	}

	test(nil, cli.ParsedCmd{})
	test([]string{}, cli.ParsedCmd{})
	test([]string{}, cli.ParsedCmd{})

	test([]string{"X"}, cli.ParsedCmd{seg("X")})
	test([]string{"X", "21"}, cli.ParsedCmd{seg("X"), seg("21")})
	test([]string{"X/21"}, cli.ParsedCmd{seg("X"), seg("21")})
	test([]string{"X.21"}, cli.ParsedCmd{seg("X"), seg("21")})
	test([]string{"X 21"}, cli.ParsedCmd{seg("X"), seg("21")})
	test([]string{"X \t 21"}, cli.ParsedCmd{seg("X"), seg("21")})
	test([]string{"X", "twenty-one"}, cli.ParsedCmd{seg("X"), seg("21")})

	test([]string{"{a=V}", "X", "21"}, cli.ParsedCmd{seg("", "a"), seg("X"), seg("21")})
	test([]string{"X", "{a=V}", "21"}, cli.ParsedCmd{seg("X", "X.a"), seg("21")})
	test([]string{"X", "21", "{a=V}"}, cli.ParsedCmd{seg("X"), seg("21", "X.21.a")})
	test([]string{"X", "twenty-one", "{a=V}"}, cli.ParsedCmd{seg("X"), seg("21", "X.21.a")})

	test([]string{"X{a=V}/21"}, cli.ParsedCmd{seg("X", "X.a"), seg("21")})
	test([]string{"X{a=V}/{b=V}21"}, cli.ParsedCmd{seg("X", "X.a", "X.b"), seg("21")})
	test([]string{"X{a=V}./{b=V}21"}, cli.ParsedCmd{seg("X", "X.a", "X.b"), seg("21")})
	test([]string{"X/{a=V}21"}, cli.ParsedCmd{seg("X", "X.a"), seg("21")})
	test([]string{"X{a=V}{b=V} / / {c=V}{d=V}21{e=V}{f=V}"},
		cli.ParsedCmd{seg("X", "X.a", "X.b", "X.c", "X.d"), seg("21", "X.21.e", "X.21.f")})

	test([]string{"X{a=V}21"}, cli.ParsedCmd{seg("X", "X.a"), seg("21")})
	test([]string{"X/ {a=V}21"}, cli.ParsedCmd{seg("X", "X.a"), seg("21")})
	test([]string{"X /{a=V}21"}, cli.ParsedCmd{seg("X", "X.a"), seg("21")})
	test([]string{"X / {a=V} 21"}, cli.ParsedCmd{seg("X", "X.a"), seg("21")})
	test([]string{"{a=V}{b=V}X{c=V}21{d=V}{e=V}"},
		cli.ParsedCmd{seg("", "a", "b"), seg("X", "X.c"), seg("21", "X.21.d", "X.21.e")})
}
