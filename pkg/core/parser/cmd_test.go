package parser

import (
	"testing"

	"github.com/innerr/ticat/pkg/core/model"
)

func TestCmdParserParseSeg(t *testing.T) {
	assertEq := func(input []string, a []parsedSeg, b []parsedSeg) {
		fatal := func() {
			t.Fatalf("%#v: %#v != %#v\n", input, a, b)
		}
		if len(a) != len(b) {
			t.Fatalf("%#v len not match: %v != %v\n", input, len(a), len(b))
		}
		for i := range a {
			if a[i].Type != b[i].Type {
				fatal()
			}
			if a[i].Type == parsedSegTypeCmd {
				ac := a[i].Val.(model.MatchedCmd)
				bc := b[i].Val.(model.MatchedCmd)
				if ac.Name != bc.Name {
					fatal()
				}
			} else if a[i].Type == parsedSegTypeEnv {
				ae := a[i].Val.(model.ParsedEnv)
				be := b[i].Val.(model.ParsedEnv)
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

	root := newCmdTree()
	l2 := root.AddSub("X")
	l2.AddSub("21", "twenty-one")

	parser := &CmdParser{
		&EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"},
		".", "./", "\t ", "<root>", "^", map[byte]bool{'/': true, '\\': true},
	}

	sep := parsedSeg{Type: parsedSegTypeSep, Val: nil}

	cmd := func(name string) parsedSeg {
		return parsedSeg{Type: parsedSegTypeCmd, Val: model.MatchedCmd{Name: name, Cmd: nil}}
	}
	env := func(names ...string) parsedSeg {
		env := model.ParsedEnv{}
		for _, name := range names {
			env[name] = model.NewParsedEnvVal(name, "V")
		}
		return parsedSeg{Type: parsedSegTypeEnv, Val: env}
	}

	test := func(a []string, b []parsedSeg) {
		parsed, _, _, _ := parser.parse(root, nil, a)
		assertEq(a, parsed, b)
	}

	test(nil, []parsedSeg{})
	test([]string{}, []parsedSeg{})
	test([]string{}, []parsedSeg{})

	test([]string{"X"}, []parsedSeg{cmd("X")})
	test([]string{"X/", ".21"}, []parsedSeg{cmd("X"), sep, cmd("21")})
	test([]string{"X.21"}, []parsedSeg{cmd("X"), sep, cmd("21")})
	test([]string{"X/21"}, []parsedSeg{cmd("X"), sep, cmd("21")})

	test([]string{"X{}21"}, []parsedSeg{cmd("X"), cmd("21")})
	test([]string{"X / . / 21"}, []parsedSeg{cmd("X"), sep, cmd("21")})
	test([]string{"X.twenty-one"}, []parsedSeg{cmd("X"), sep, cmd("twenty-one")})

	test([]string{"{a=V}", "X", "/", "21"}, []parsedSeg{env("a"), cmd("X"), sep, cmd("21")})
	test([]string{"X", ".{a=V}.", "21"}, []parsedSeg{cmd("X"), sep, env("a"), sep, cmd("21")})
	test([]string{"X.", "21.", "{a=V}"}, []parsedSeg{cmd("X"), sep, cmd("21"), sep, env("a")})
	test([]string{"X.", ".21.", "{a=V}"}, []parsedSeg{cmd("X"), sep, cmd("21"), sep, env("a")})

	test([]string{"X.{a=V}21"}, []parsedSeg{cmd("X"), sep, env("a"), cmd("21")})
	test([]string{"X{a=V}.21"}, []parsedSeg{cmd("X"), env("a"), sep, cmd("21")})
	test([]string{"X{a=V}/21"}, []parsedSeg{cmd("X"), env("a"), sep, cmd("21")})
	test([]string{"X{a=V}/{b=V}21"}, []parsedSeg{cmd("X"), env("a"), sep, env("b"), cmd("21")})
	test([]string{"X{a=V} / / {b=V}21"}, []parsedSeg{cmd("X"), env("a"), sep, env("b"), cmd("21")})

	test([]string{"X/{a=V}21"}, []parsedSeg{cmd("X"), sep, env("a"), cmd("21")})
	test([]string{"X{a=V}.21"}, []parsedSeg{cmd("X"), env("a"), sep, cmd("21")})
	test([]string{"X/ {a=V}21"}, []parsedSeg{cmd("X"), sep, env("a"), cmd("21")})
	test([]string{"X /{a=V}21"}, []parsedSeg{cmd("X"), sep, env("a"), cmd("21")})
	test([]string{"X / {a=V} 21"}, []parsedSeg{cmd("X"), sep, env("a"), cmd("21")})
	test([]string{"{a=V}{b=V}X{c=V}21{d=V}{e=V}"},
		[]parsedSeg{env("a"), env("b"), cmd("X"), env("c"), cmd("21"), env("d"), env("e")})

	test([]string{"{}{}X{}"}, []parsedSeg{cmd("X")})
}

func TestCmdParserParse(t *testing.T) {
	assertEq := func(a model.ParsedCmd, b model.ParsedCmd) {
		for i, av := range a.Segments {
			if i >= len(b.Segments) || len(a.Segments) != len(b.Segments) {
				t.Fatalf("size not eq: %#v != %#v\n", len(a.Segments), len(b.Segments))
			}
			bv := b.Segments[i]
			if len(bv.Matched.Name) != 0 && av.Matched.Cmd == nil {
				t.Fatalf("#%d cmd not eq: '%#v' != '%#v'\n", i, av.Matched, bv.Matched)
			}
			if av.Matched.Cmd != nil && av.Matched.Cmd.Name() != bv.Matched.Name {
				t.Fatalf("#%d cmd name not eq: '%#v' != '%#v'\n", i, av.Matched.Cmd.Name(), bv.Matched.Name)
			}
			if (av.Env != nil) != (bv.Env != nil) {
				t.Fatalf("#%d env nil check not eq: %#v != %#v\n", i, av.Env, bv.Env)
			}
			if av.Env != nil && !av.Env.Equal(bv.Env) {
				t.Fatalf("#%d env not eq: %#v != %#v\n", i, av.Env, bv.Env)
			}
		}
	}

	root := newCmdTree()
	l2 := root.AddSub("X")
	l2.AddSub("21", "twenty-one")

	parser := &CmdParser{
		&EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"},
		".", "./", "\t ", "<root>", "^", map[byte]bool{'/': true, '\\': true},
	}

	seg := func(cmdName string, envKeyNames ...string) model.ParsedCmdSeg {
		var env model.ParsedEnv
		if len(envKeyNames) != 0 {
			env = model.ParsedEnv{}
			for _, name := range envKeyNames {
				env[name] = model.NewParsedEnvVal(name, "V")
			}
		}
		return model.ParsedCmdSeg{Env: env, Matched: model.MatchedCmd{Name: cmdName, Cmd: nil}}
	}

	cmd := func(segs ...model.ParsedCmdSeg) model.ParsedCmd {
		return model.ParsedCmd{Segments: segs, ParseResult: model.ParseResult{Input: nil, Error: nil, IsMinorErr: false}, TrivialLvl: 0}
	}

	test := func(a []string, b model.ParsedCmd) {
		parsed := parser.Parse(root, nil, a)
		assertEq(parsed, b)
	}

	test(nil, model.ParsedCmd{})
	test([]string{}, model.ParsedCmd{})
	test([]string{}, model.ParsedCmd{})

	test([]string{"X"}, cmd(seg("X")))
	test([]string{"X.", "/21"}, cmd(seg("X"), seg("21")))
	test([]string{"X/21"}, cmd(seg("X"), seg("21")))
	test([]string{"X.21"}, cmd(seg("X"), seg("21")))
	test([]string{"X{}21"}, cmd(seg("X"), seg("21")))
	test([]string{"X/ \t .21"}, cmd(seg("X"), seg("21")))
	test([]string{"X", "{}twenty-one"}, cmd(seg("X"), seg("21")))

	test([]string{"{a=V}", "X.", "/21"}, cmd(seg("", "a"), seg("X"), seg("21")))
	test([]string{"X", "{a=V}", "21"}, cmd(seg("X", "X.a"), seg("21")))
	test([]string{"X", ".", "21", "{a=V}"}, cmd(seg("X"), seg("21", "X.21.a")))
	test([]string{"X .", " / ", "/ twenty-one", "{a=V}"}, cmd(seg("X"), seg("21", "X.21.a")))

	test([]string{"X{a=V}/21"}, cmd(seg("X", "X.a"), seg("21")))
	test([]string{"X{a=V}/{b=V}21"}, cmd(seg("X", "X.a", "X.b"), seg("21")))
	test([]string{"X/{a=V}21"}, cmd(seg("X", "X.a"), seg("21")))
	test([]string{"X{a=V}{b=V} / / {c=V}{d=V}21{e=V}{f=V}"},
		cmd(seg("X", "X.a", "X.b", "X.c", "X.d"), seg("21", "X.21.e", "X.21.f")))

	test([]string{"X{a=V}21"}, cmd(seg("X", "X.a"), seg("21")))
	test([]string{"X/ {a=V}21"}, cmd(seg("X", "X.a"), seg("21")))
	test([]string{"X /{a=V}21"}, cmd(seg("X", "X.a"), seg("21")))
	test([]string{"X / {a=V} 21"}, cmd(seg("X", "X.a"), seg("21")))
	test([]string{"{a=V}{b=V}X{c=V}21{d=V}{e=V}"},
		cmd(seg("", "a", "b"), seg("X", "X.c"), seg("21", "X.21.d", "X.21.e")))
}

func newCmdTree() *model.CmdTree {
	return model.NewCmdTree(model.CmdTreeStrsForTest())
}

func TestCmdParserParseWithArgs(t *testing.T) {
	root := newCmdTree()
	echo := root.AddSub("echo")
	echo.RegEmptyCmd("print message").
		AddArg("message", "", "msg", "m").
		AddArg("color", "", "colour", "clr", "c")

	parser := &CmdParser{
		&EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"},
		".", "./", "\t ", "<root>", "^", map[byte]bool{'/': true, '\\': true},
	}

	t.Run("echo with positional arg", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"echo", "hello"})
		if len(parsed.Segments) != 1 {
			t.Fatalf("expected 1 segment, got %d", len(parsed.Segments))
		}
		if parsed.Segments[0].Matched.Name != "echo" {
			t.Errorf("expected cmd name 'echo', got '%s'", parsed.Segments[0].Matched.Name)
		}
		if parsed.Segments[0].Env == nil {
			t.Fatal("expected env to be set with argument")
		}
		if parsed.Segments[0].Env["echo.message"].Val != "hello" {
			t.Errorf("expected echo.message='hello', got '%s'", parsed.Segments[0].Env["echo.message"].Val)
		}
	})

	t.Run("echo with named arg", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"echo", "{message=world}"})
		if len(parsed.Segments) != 1 {
			t.Fatalf("expected 1 segment, got %d", len(parsed.Segments))
		}
		if parsed.Segments[0].Matched.Name != "echo" {
			t.Errorf("expected cmd name 'echo', got '%s'", parsed.Segments[0].Matched.Name)
		}
		if parsed.Segments[0].Env == nil {
			t.Fatal("expected env to be set with argument")
		}
		if parsed.Segments[0].Env["echo.message"].Val != "world" {
			t.Errorf("expected echo.message='world', got '%s'", parsed.Segments[0].Env["echo.message"].Val)
		}
	})

	t.Run("echo with two positional args", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"echo", "hi", "red"})
		if len(parsed.Segments) != 1 {
			t.Fatalf("expected 1 segment, got %d", len(parsed.Segments))
		}
		if parsed.Segments[0].Env == nil {
			t.Fatal("expected env to be set with arguments")
		}
		if parsed.Segments[0].Env["echo.message"].Val != "hi" {
			t.Errorf("expected echo.message='hi', got '%s'", parsed.Segments[0].Env["echo.message"].Val)
		}
		if parsed.Segments[0].Env["echo.color"].Val != "red" {
			t.Errorf("expected echo.color='red', got '%s'", parsed.Segments[0].Env["echo.color"].Val)
		}
	})

	t.Run("echo with abbreviated arg name", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{"echo", "m", "hello"})
		if len(parsed.Segments) != 1 {
			t.Fatalf("expected 1 segment, got %d", len(parsed.Segments))
		}
		if parsed.Segments[0].Env == nil {
			t.Fatal("expected env to be set with argument")
		}
		if parsed.Segments[0].Env["echo.message"].Val != "hello" {
			t.Errorf("expected echo.message='hello', got '%s'", parsed.Segments[0].Env["echo.message"].Val)
		}
	})
}
