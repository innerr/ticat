package main

import (
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
	defEnv.Set("strs.env-sys-path", EnvRuntimeSysPrefix)
	defEnv.Set("strs.env-strs-path", EnvStrsPrefix)
	defEnv.Set("strs.proto-env-mark", ProtoEnvMark)
	defEnv.Set("strs.proto-sep", ProtoSep)
	defEnv.Set("strs.proto-bash-ext", ProtoBashExt)
	defEnv.Set("strs.env-file-name", EnvFileName)

	// The available cmds are organized in a tree, will grow bigger after running bootstrap
	tree := core.NewCmdTree(&core.CmdTreeStrs{
		CmdRootDisplayName,
		CmdPathSep,
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

	// Load runtime env, load local-stored env, load local mods
	bootstrap := "B.E.L.R:B.E.L.L:B.M.L.L"

	// Interacting methods between ticat and mods:
	//   1. mod.stdin(as mod's input args) -> mod.stderr(as mods's return)
	//   2. (recursively) calling ticat inside a mod -> ticat.stdin(pass the env from mod to ticat)
	//
	// The stdin-env could be very useful for customized mods-loader or env-loader
	//   1. those loaders will be loaded from 'bootstrap' string above
	//   2. put a string val with key 'bootstrap' to env could launch it as an extra bootstrap
	stdinEnv := cli.GenEnvFromStdin(ProtoEnvMark, ProtoSep)
	if stdinEnv != nil {
		globalEnv.GetLayer(core.EnvLayerSession).Merge(stdinEnv)
	}

	executor := cli.Executor{}
	succeeded := executor.Execute(cc, bootstrap, os.Args[1:]...)
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
	CmdPathAlterSeps    string = Spaces + "./"
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
	ProtoBashExt        string = "bash"
)
