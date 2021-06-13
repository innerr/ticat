package main

import (
	"os"

	"github.com/pingcap/ticat/pkg/builtin"
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
	"github.com/pingcap/ticat/pkg/cli/execute"
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
	defEnv.Set("strs.self-name", SelfName)
	defEnv.Set("strs.cmd-builtin-display-name", CmdBuiltinDisplayName)
	defEnv.Set("strs.meta-ext", MetaExt)
	defEnv.Set("strs.flow-ext", FlowExt)
	defEnv.Set("strs.abbrs-sep", AbbrsSep)
	defEnv.Set("strs.seq-sep", SequenceSep)
	defEnv.Set("strs.cmd-path-sep", CmdPathSep)
	defEnv.Set("strs.env-path-sep", EnvPathSep)
	defEnv.Set("strs.env-op-sep", EnvOpSep)
	defEnv.Set("strs.env-sys-path", EnvRuntimeSysPrefix)
	defEnv.Set("strs.env-strs-path", EnvStrsPrefix)
	defEnv.Set("strs.env-kv-sep", EnvKeyValSep)
	defEnv.Set("strs.env-bracket-left", EnvBracketLeft)
	defEnv.Set("strs.env-bracket-right", EnvBracketRight)
	defEnv.Set("strs.env-file-name", EnvFileName)
	defEnv.Set("strs.session-env-file", SessionEnvFileName)
	defEnv.Set("strs.hub-file-name", HubFileName)
	defEnv.Set("strs.repos-file-name", ReposFileName)
	defEnv.Set("strs.mods-repo-ext", ModsRepoExt)
	defEnv.Set("strs.proto-sep", ProtoSep)
	defEnv.Set("strs.tag-out-of-the-box", TagOutOfTheBox)
	defEnv.Set("strs.tag-provider", TagProvider)
	defEnv.Set("strs.tag-self-test", TagSelfTest)

	// The available cmds are organized in a tree, will grow bigger after running bootstrap
	tree := core.NewCmdTree(&core.CmdTreeStrs{
		CmdRootDisplayName,
		CmdBuiltinDisplayName,
		CmdPathSep,
		CmdPathAlterSeps,
		AbbrsSep,
		EnvOpSep,
		EnvValDelAllMark,
		EnvKeyValSep,
		EnvPathSep,
		ProtoSep,
	})
	builtin.RegisterCmds(tree)

	// Extra abbrs definition
	abbrs := core.NewEnvAbbrs(CmdRootDisplayName)
	builtin.LoadEnvAbbrs(abbrs)

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

	// Virtual tty, for re-directing
	screen := execute.NewScreen()

	// The Cli is a service set, the builtin mods will receive it as a arg when being called
	cc := core.NewCli(globalEnv, screen, tree, cliParser, abbrs)

	// Modules and env loaders
	bootstrap := `
		B.E.L.R:
		B.M.L.E:
		B.E.L.L:
		B.M.L.F:
		B.M.L.H:
		B.D.L.P:
	`

	// TODO: handle error by types
	defer func() {
		if r := recover(); r != nil {
			display.PrintError(cc, cc.GlobalEnv, r.(error))
			os.Exit(-1)
		}
	}()

	// Main process
	executor := execute.NewExecutor(SessionEnvFileName)
	cc.Executor = executor
	succeeded := executor.Run(cc, bootstrap, os.Args[1:]...)

	// TODO: more exit codes
	if !succeeded {
		os.Exit(1)
	}
}

const (
	SelfName              string = "ticat"
	CmdRootDisplayName    string = "<root>"
	CmdBuiltinDisplayName string = "<builtin>"
	Spaces                string = "\t\n\r "
	AbbrsSep              string = "|"
	EnvOpSep              string = ":"
	SequenceSep           string = ":"
	CmdPathSep            string = "."
	CmdPathAlterSeps      string = "./"
	EnvBracketLeft        string = "{"
	EnvBracketRight       string = "}"
	EnvKeyValSep          string = "="
	EnvPathSep            string = "."
	EnvValDelAllMark      string = "--"
	EnvRuntimeSysPrefix   string = "sys"
	EnvStrsPrefix         string = "strs"
	EnvFileName           string = "bootstrap.env"
	ProtoSep              string = "\t"
	ModsRepoExt           string = "." + SelfName
	MetaExt               string = "." + SelfName
	FlowExt               string = ".flow." + SelfName
	HubFileName           string = "repos.hub"
	ReposFileName         string = "README.md"
	SessionEnvFileName    string = "env"
	TagOutOfTheBox        string = "@ready"
	TagProvider           string = "@provider"
	TagSelfTest           string = "@selftest"
)
