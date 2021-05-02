package main

import (
	"os"

	"github.com/pingcap/ticat/pkg/cli"
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/builtin"
	"github.com/pingcap/ticat/pkg/cli/parser"
)

func main() {
	// For more detail, in termial execute:
	// $> ticat desc: the-bootstrap-string
	bootstrap := "B.E.L.R:B.E.L.L:B.M.L.L"

	metaEnv := cli.GenEnvFromStdin(ProtoEnvMark, ProtoSep)

	globalEnv := core.NewEnv().NewLayers(
		core.EnvLayerDefault,
		core.EnvLayerPersisted,
		core.EnvLayerSession,
	)
	builtin.LoadDefaultEnv(globalEnv)

	// For the mods witch need to use the defined string set
	defEnv := globalEnv.GetLayer(core.EnvLayerDefault)
	defEnv.Set("strs.meta-ext", SelfName)
	defEnv.Set("strs.abbrs-sep", ProtoAbbrsSep)
	defEnv.Set("strs.env-sys-path", EnvRuntimeSysPrefix)
	defEnv.Set("strs.proto-env-mark", ProtoEnvMark)
	defEnv.Set("strs.proto-sep", ProtoSep)
	defEnv.Set("strs.proto-bash-ext", ProtoBashExt)
	defEnv.Set("strs.env-file-name", EnvFileName)

	tree := core.NewCmdTree(&core.CmdTreeStrs{
		CmdRootDisplayName,
		CmdPathSep,
		EnvValDelMark,
		EnvValDelAllMark,
		ProtoEnvMark,
		ProtoSep,
	})
	builtin.RegisterBuiltinMods(tree)

	seqParser := parser.NewSequenceParser(
		SequenceSep,
		[]string{"http", "HTTP"},
		[]string{"/"})
	envParser := parser.NewEnvParser(
		parser.Brackets{EnvBracketLeft, EnvBracketRight},
		Spaces,
		EnvKeyValSep,
		EnvPathSep)
	cmdParser := parser.NewCmdParser(
		envParser,
		CmdPathSep,
		CmdPathAlterSeps,
		Spaces,
		CmdRootDisplayName)
	cliParser := parser.NewParser(seqParser, cmdParser)

	screen := cli.NewScreen()
	envAbbrs := core.NewEnvAbbrs(CmdRootDisplayName)

	cc := &core.Cli{
		globalEnv,
		screen,
		tree,
		cliParser,
		envAbbrs,
	}

	executor := cli.Executor{}

	succeeded := executor.Execute(cc, bootstrap, metaEnv, os.Args[1:]...)
	if !succeeded {
		os.Exit(1)
	}
}

const (
	SelfName            string = "ticat"
	CmdRootDisplayName  string = "<root>"
	Spaces              string = "\t\n\r "
	SequenceSep         string = ":"
	CmdPathSep          string = "."
	CmdPathAlterSeps    string = Spaces + "./"
	EnvBracketLeft      string = "{"
	EnvBracketRight     string = "}"
	EnvKeyValSep        string = "="
	EnvPathSep          string = "."
	EnvValDelMark       string = "-"
	EnvValDelAllMark    string = "--"
	EnvRuntimeSysPrefix string = "sys."
	EnvFileName         string = "bootstrap.env"
	ProtoAbbrsSep       string = "|"
	ProtoMark           string = "proto." + SelfName
	ProtoEnvMark        string = ProtoMark + ".env"
	ProtoSep            string = "\t"
	ProtoBashExt        string = "bash"
)
