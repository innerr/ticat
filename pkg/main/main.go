package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/pingcap/ticat/pkg/builtin"
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
	"github.com/pingcap/ticat/pkg/cli/execute"
	"github.com/pingcap/ticat/pkg/cli/parser"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	env := core.NewEnvEx(core.EnvLayerDefault).NewLayers(
		core.EnvLayerPersisted,
		core.EnvLayerSession,
	)
	envKeysInfo := core.NewEnvKeysInfo()
	builtin.LoadDefaultEnv(env, envKeysInfo)

	// Any mod could get the specific string val from env when it's called
	defEnv := env.GetLayer(core.EnvLayerDefault)
	defEnv.Set("strs.self-name", SelfName)
	defEnv.Set("strs.list-sep", ListSep)
	defEnv.Set("strs.cmd-builtin-name", CmdBuiltinName)
	defEnv.Set("strs.cmd-builtin-display-name", CmdBuiltinDisplayName)
	defEnv.Set("strs.meta-ext", MetaExt)
	defEnv.Set("strs.flow-ext", FlowExt)
	defEnv.Set("strs.help-ext", HelpExt)
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
	defEnv.Set("strs.session-status-file", SessionStatusFileName)
	defEnv.Set("strs.hub-file-name", HubFileName)
	defEnv.Set("strs.repos-file-name", ReposFileName)
	defEnv.Set("strs.mods-repo-ext", ModsRepoExt)
	defEnv.Set("strs.proto-sep", ProtoSep)
	defEnv.Set("strs.tag-out-of-the-box", TagOutOfTheBox)
	defEnv.Set("strs.tag-provider", TagProvider)
	defEnv.Set("strs.tag-self-test", TagSelfTest)
	defEnv.Set("strs.flow-template-bracket-left", FlowTemplateBracketLeft)
	defEnv.Set("strs.flow-template-bracket-right", FlowTemplateBracketRight)
	defEnv.Set("strs.flow-template-multiply-mark", FlowTemplateMultiplyMark)
	defEnv.Set("strs.tag-mark", TagMark)
	defEnv.Set("strs.trivial-mark", TrivialMark)
	defEnv.Set("strs.sys-arg-prefix", SysArgPrefix)
	defEnv.Set("strs.env-snapshot-ext", EnvSnapshotExt)
	defEnv.Set("strs.env-del-all-mark", EnvValDelAllMark)
	defEnv.Set("strs.cmd-path-str-session", CmdPathSession)
	defEnv.Set("strs.arg-enum-sep", ArgEnumSep)

	envKeysInfo.GetOrAdd("strs.proto-sep").InvisibleDisplay = "<tab>"

	// The available cmds are organized in a tree, will grow bigger after running bootstrap
	tree := core.NewCmdTree(&core.CmdTreeStrs{
		SelfName,
		CmdRootDisplayName,
		CmdBuiltinName,
		CmdBuiltinDisplayName,
		CmdPathSep,
		CmdPathAlterSeps,
		AbbrsSep,
		EnvOpSep,
		EnvValDelAllMark,
		EnvKeyValSep,
		EnvPathSep,
		ProtoSep,
		ListSep,
		FlowTemplateBracketLeft,
		FlowTemplateBracketRight,
		FlowTemplateMultiplyMark,
		TagMark,
		ArgEnumSep,
	})
	builtin.RegisterCmds(tree)

	// Extra abbrs definition
	abbrs := core.NewEnvAbbrs(CmdRootDisplayName)
	builtin.LoadEnvAbbrs(abbrs)

	// A simple parser, should be insteaded in the future
	seqParser := parser.NewSequenceParser(
		SequenceSep,
		[]string{"http", "HTTP", "https", "HTTPS"},
		nil,
	)
	envParser := parser.NewEnvParser(
		parser.Brackets{EnvBracketLeft, EnvBracketRight},
		Spaces,
		EnvKeyValSep,
		EnvPathSep,
		SysArgPrefix)
	cmdParser := parser.NewCmdParser(
		envParser,
		CmdPathSep,
		CmdPathAlterSeps,
		Spaces,
		CmdRootDisplayName,
		TrivialMark,
		FakeCmdPathSepSuffixs)
	cliParser := parser.NewParser(seqParser, cmdParser)

	// Executing info, commands' output are not included
	screen := core.NewStdScreen(os.Stdout, os.Stderr)

	// Commands' input and output
	cmdIO := core.NewCmdIO(os.Stdin, os.Stdout, os.Stderr)

	// The Cli is a service set, the builtin mods will receive it as a arg when being called
	cc := core.NewCli(screen, tree, cliParser, abbrs, cmdIO, envKeysInfo)

	// Modules and env loaders
	bootstrap := `
		builtin.env.load.runtime:
		builtin.mod.load.ext-executor:
		builtin.env.load.local:
		builtin.mod.load.flows:
		builtin.mod.load.hub:
		builtin.display.load.platform:
		builtin.hub.init:
	`

	exitEventHook := func(hookKeyName string) {
		if len(env.GetRaw("session")) == 0 {
			return
		}
		hookStr := env.GetRaw(hookKeyName)
		if len(hookStr) != 0 && cc.Executor != nil {
			hookFlow := core.FlowStrToStrs(hookStr)

			// TODO: not working, has bugs
			//env = env.NewLayer(core.EnvLayerTmp)
			env.SetBool("display.one-cmd", true)
			env.SetBool("sys.unlog-status", true)

			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}
				if env.GetBool("sys.panic.recover") {
					display.PrintError(cc, env, recovered.(error))
					os.Exit(-2)
				} else {
					panic(recovered)
				}
			}()

			succeeded := cc.Executor.Execute(hookKeyName, true, cc, env, nil, hookFlow...)
			if !succeeded {
				os.Exit(2)
			}
		}
	}

	// TODO: handle error by types
	defer func() {
		recovered := recover()
		if recovered == nil {
			// Exit point: succeed
			exitEventHook("sys.event.hook.done")
			exitEventHook("sys.event.hook.exit")
			return
		}
		if env.GetBool("sys.panic.recover") {
			display.PrintError(cc, env, recovered.(error))
		}
		// Exit point: panic
		exitEventHook("sys.event.hook.error")
		exitEventHook("sys.event.hook.exit")
		if env.GetBool("sys.panic.recover") {
			os.Exit(-1)
		} else {
			panic(recovered)
		}
	}()

	// Main process
	executor := execute.NewExecutor(SessionEnvFileName,
		SessionStatusFileName, "<bootstrap>", "<entry>")
	cc.Executor = executor
	succeeded := executor.Run(cc, env, bootstrap, os.Args[1:]...)

	// TODO: more exit codes
	if !succeeded {
		// Exit point: error
		exitEventHook("sys.event.hook.error")
		exitEventHook("sys.event.hook.exit")
		os.Exit(1)
	}
}

const (
	SelfName                 string = "ticat"
	ListSep                  string = ","
	CmdRootDisplayName       string = "<root>"
	CmdBuiltinName           string = "builtin"
	CmdBuiltinDisplayName    string = "<builtin>"
	Spaces                   string = "\t\n\r "
	AbbrsSep                 string = "|"
	EnvOpSep                 string = ":"
	SequenceSep              string = ":"
	CmdPathSep               string = "."
	FakeCmdPathSepSuffixs    string = "/\\"
	CmdPathAlterSeps         string = "."
	EnvBracketLeft           string = "{"
	EnvBracketRight          string = "}"
	EnvKeyValSep             string = "="
	EnvPathSep               string = "."
	SysArgPrefix             string = "%"
	EnvValDelAllMark         string = "--"
	EnvRuntimeSysPrefix      string = "sys"
	EnvStrsPrefix            string = "strs"
	EnvFileName              string = "bootstrap.env"
	ProtoSep                 string = "\t"
	ModsRepoExt              string = "." + SelfName
	MetaExt                  string = "." + SelfName
	FlowExt                  string = ".tiflow"
	HelpExt                  string = ".tihelp"
	HubFileName              string = "repos.hub"
	ReposFileName            string = "hub.ticat"
	SessionEnvFileName       string = "env"
	SessionStatusFileName    string = "status"
	FlowTemplateBracketLeft  string = "[["
	FlowTemplateBracketRight string = "]]"
	FlowTemplateMultiplyMark string = "*"
	TagMark                  string = "@"
	TrivialMark              string = "^"
	TagOutOfTheBox           string = TagMark + "ready"
	TagProvider              string = TagMark + "config"
	TagSelfTest              string = TagMark + "selftest"
	EnvSnapshotExt           string = ".env"
	CmdPathSession           string = "sessions"
	ArgEnumSep               string = "|"
)
