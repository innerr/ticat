package ticat

import (
	"fmt"
	"os"
	"strings"

	"github.com/innerr/ticat/pkg/cli/display"
	"github.com/innerr/ticat/pkg/core/execute"
	"github.com/innerr/ticat/pkg/core/model"
	"github.com/innerr/ticat/pkg/core/parser"
	"github.com/innerr/ticat/pkg/mods/builtin"
)

type TiCat struct {
	Env       *model.Env
	Cmds      *model.CmdTree
	Bootstrap []string
	cc        *model.Cli
	executor  *execute.Executor
}

func NewTiCat() *TiCat {
	env := model.NewEnvEx(model.EnvLayerDefault).NewLayers(
		model.EnvLayerPersisted,
		model.EnvLayerSession,
	)
	envKeysInfo := model.NewEnvKeysInfo()
	builtin.LoadDefaultEnv(env, envKeysInfo)

	// Any mod could get the specific string val from env when it's called
	defEnv := env.GetLayer(model.EnvLayerDefault)
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
	tree := model.NewCmdTree(&model.CmdTreeStrs{
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
	abbrs := model.NewEnvAbbrs(CmdRootDisplayName)
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
	screen := model.NewStdScreen(os.Stdout, os.Stderr)

	// Commands' input and output
	cmdIO := model.NewCmdIO(os.Stdin, os.Stdout, os.Stderr)

	// The Cli is a service set, the builtin mods will receive it as a arg when being called
	cc := model.NewCli(screen, tree, cliParser, abbrs, cmdIO, envKeysInfo)

	// Modules and env loaders
	bootstrap := []string{
		"builtin.env.load.runtime",
		"builtin.mod.load.ext-executor",
		"builtin.env.load.local",
		"builtin.mod.load.flows",
		"builtin.mod.load.hub",
		"builtin.display.load.platform",
		"builtin.hub.init",
	}

	// Main process
	executor := execute.NewExecutor(SessionEnvFileName, SessionStatusFileName, "<bootstrap>", "<entry>")
	cc.Executor = executor

	return &TiCat{
		Env:       env,
		Bootstrap: bootstrap,
		Cmds:      cc.Cmds,
		cc:        cc,
		executor:  executor,
	}
}

func (ticat *TiCat) SetScreen(screen model.Screen) {
	ticat.cc.Screen = screen
}

func (ticat *TiCat) AddIntegratedModVersion(ver string) {
	env := ticat.Env.GetLayer(model.EnvLayerDefault)
	strs := strings.Split(env.GetRaw("sys.mods.integrated"), ListSep)
	strs = append(strs, ver)
	env.Set("sys.mods.integrated", strings.Join(strs, ListSep))
}

func (ticat *TiCat) InjectBootstrap(cmds []string) {
	bootstrap := ticat.Bootstrap
	for i, cmd := range bootstrap {
		if cmd == "builtin.hub.init" {
			newBootstrap := append(bootstrap[0:i], cmds...)
			newBootstrap = append(newBootstrap, bootstrap[i:len(bootstrap)]...)
			ticat.Bootstrap = newBootstrap
			return
		}
	}
	ticat.Bootstrap = append(bootstrap, cmds...)
}

func (ticat *TiCat) RunCli(args ...string) bool {
	return ticat.run(args...)
}

func (ticat *TiCat) run(args ...string) bool {
	env := ticat.Env

	defer func() {
		// TODO: handle error by types
		recovered := recover()
		if recovered == nil {
			// Exit point: succeed
			ticat.exitEventHook("sys.event.hook.done")
			ticat.exitEventHook("sys.event.hook.exit")
			return
		}
		if env.GetBool("sys.panic.recover") {
			display.PrintError(ticat.cc, env, recovered.(error))
		}
		// Exit point: panic
		ticat.exitEventHook("sys.event.hook.error")
		ticat.exitEventHook("sys.event.hook.exit")
		// TODO: need to check this flag doesn't from user-level persisted env
		if env.GetBool("sys.panic.recover") {
			os.Exit(-1)
		} else {
			// Re-throw
			panic(recovered)
		}
	}()

	bootstrap := strings.Join(ticat.Bootstrap, ":")
	succeeded := ticat.executor.Run(ticat.cc, env, bootstrap, args...)
	// TODO: more exit codes
	if !succeeded {
		// Exit point: error
		ticat.exitEventHook("sys.event.hook.error")
		ticat.exitEventHook("sys.event.hook.exit")
		os.Exit(1)
	}
	return succeeded
}

func (ticat *TiCat) exitEventHook(hookKeyName string) (ok bool) {
	env := ticat.Env
	if len(env.GetRaw("session")) == 0 {
		return true
	}
	hookStr := env.GetRaw(hookKeyName)
	if !(len(hookStr) != 0 && ticat.executor != nil) {
		return true
	}

	hookFlow := model.FlowStrToStrs(hookStr)

	// TODO: not working, has bugs
	//env = env.NewLayer(model.EnvLayerTmp)

	env.SetBool("display.one-cmd", true)
	env.SetBool("sys.unlog-status", true)

	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}
		// Exit point: panic when executing exit-event-hook
		// TODO: need to check this flag doesn't from user-level persisted env
		if env.GetBool("sys.panic.recover") {
			display.PrintError(ticat.cc, env, recovered.(error))
			os.Exit(-2)
		} else {
			// Re-throw
			panic(recovered)
		}
	}()

	succeeded := ticat.executor.Execute(hookKeyName, true, ticat.cc, env, nil, hookFlow...)
	if !succeeded {
		// TODO: better error handle
		panic(fmt.Errorf("failed to execute exit-hook ('%s'): %s", hookKeyName, hookFlow))
	}

	return true
}
