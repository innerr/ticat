package main

import (
	"fmt"
	"os"

	"github.com/pingcap/ticat/pkg/builtin"
	"github.com/pingcap/ticat/pkg/cli"
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/parser"
)

func main() {
	globalEnv := core.NewEnv().NewLayers(
		core.EnvLayerDefault,
		core.EnvLayerPersisted,
		core.EnvLayerSession,
	)
	builtin.LoadDefaultEnv(globalEnv)

	// Any mod could get the specific string val from env when it's called
	defEnv := globalEnv.GetLayer(core.EnvLayerDefault)
	defEnv.Set("strs.meta-ext", SelfName)
	defEnv.Set("strs.abbrs-sep", AbbrsSep)
	defEnv.Set("strs.env-path-sep", EnvPathSep)
	defEnv.Set("strs.env-sys-path", EnvRuntimeSysPrefix)
	defEnv.Set("strs.env-strs-path", EnvStrsPrefix)
	defEnv.Set("strs.proto-env-mark", ProtoEnvMark)
	defEnv.Set("strs.proto-sep", ProtoSep)
	defEnv.Set("strs.env-file-name", EnvFileName)

	// The available cmds are organized in a tree, will grow bigger after running bootstrap
	tree := core.NewCmdTree(&core.CmdTreeStrs{
		CmdRootDisplayName,
		CmdPathSep,
		CmdPathAlterSeps,
		AbbrsSep,
		EnvValDelMark,
		EnvValDelAllMark,
		ProtoEnvMark,
		ProtoSep,
	})
	builtin.RegisterCmds(tree)

	// A simple parser, should be insteaded in the future
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

	// Virtual tty, for re-directing (in the future)
	screen := cli.NewScreen()

	// The Cli is a service set, the builtin mods will receive it as a arg when being called
	cc := core.NewCli(globalEnv, screen, tree, cliParser)

	bootstrap := `
		B.E.L.A:
		B.E.L.R:
		B.E.L.L:
		B.E.L.S:
		B.M.L.L:
		B.M.L.E:
		B.M.L.F:
	`

	// TODO: handle error by types
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			os.Exit(-1)
		}
	}()
	executor := cli.NewExecutor()
	cc.Executor = executor
	succeeded := executor.Run(cc, bootstrap, os.Args[1:]...)
	if !succeeded {
		os.Exit(1)
	}
}

const (
	SelfName            string = "ticat"
	CmdRootDisplayName  string = "<root>"
	Spaces              string = "\t\n\r "
	AbbrsSep            string = "|"
	SequenceSep         string = ":"
	CmdPathSep          string = "."
	CmdPathAlterSeps    string = "./"
	EnvBracketLeft      string = "{"
	EnvBracketRight     string = "}"
	EnvKeyValSep        string = "="
	EnvPathSep          string = "."
	EnvValDelMark       string = "-"
	EnvValDelAllMark    string = "--"
	EnvRuntimeSysPrefix string = "sys"
	EnvStrsPrefix       string = "strs"
	EnvFileName         string = "bootstrap.env"
	ProtoMark           string = "proto." + SelfName
	ProtoEnvMark        string = ProtoMark + ".env"
	ProtoSep            string = "\t"
)
