package builtin

import (
	"fmt"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

// TODO: register the env kvs these commands will modify

func RegisterCmds(cmds *core.CmdTree) {
	RegisterExecutorCmds(cmds)
	RegisterEnvCmds(cmds)
	RegisterVerbCmds(cmds)
	RegisterTrivialCmds(cmds)
	RegisterFlowCmds(cmds)
	RegisterHubCmds(cmds)
	RegisterSessionCmds(cmds.AddSub("sessions", "s", "S").RegEmptyCmd("manage sessions").Owner())
	RegisterDbgCmds(cmds.AddSub("dbg").RegEmptyCmd("debug related commands").Owner())
	RegisterMiscCmds(cmds)
	RegisterBgCmds(cmds.AddSub("background", "bg").RegEmptyCmd("background tasks management").Owner())
	RegisterDisplayCmds(cmds.AddSub("display", "disp", "dis", "di").RegEmptyCmd("display related commands").Owner())
	RegisterBuiltinCmds(cmds.AddSub("builtin", "b", "B").RegEmptyCmd("internal commands, mostly for init loading").Owner().SetHidden())
}

func RegisterExecutorCmds(cmds *core.CmdTree) {
	help := cmds.AddSub("help", "-help", "-HELP", "-h", "-H", "?")
	help.RegCmd(GlobalHelp,
		"get help")
	help.AddSub(cmds.Strs.SelfName, "self").
		RegCmd(SelfHelp,
			"get "+cmds.Strs.SelfName+" usage help")

	cmds.AddSub("usage", "args", "arg", "=").
		RegPowerCmd(DumpTailCmdUsage,
			"show usage of a command").
		SetQuiet().
		SetPriority()

	cmds.AddSub("system", "sys").
		RegPowerCmd(ExecCmds,
			"execute os command in bash").
		AddArg("command", "", "cmd")

	for i := 1; i < 6; i++ {
		defTrivial := fmt.Sprintf("%d", i)
		cmdSuffix := ""
		if i > 1 {
			cmdSuffix += fmt.Sprintf("-%d", i)
		}
		cmds.AddSub("more"+cmdSuffix, strings.Repeat("+", i)).
			RegPowerCmd(GlobalHelpMoreInfo,
				MoreHelpStr).
			SetQuiet().
			SetPriority().
			AddArg("trivial", defTrivial, "triv", "tri", "t", "T").
			AddArg("depth", "32", "d", "D")

		cmds.AddSub("less"+cmdSuffix, strings.Repeat("-", i)).
			RegPowerCmd(GlobalHelpLessInfo,
				LessHelpStr).
			SetQuiet().
			SetPriority().
			AddArg("trivial", defTrivial, "triv", "tri", "t", "T").
			AddArg("depth", "32", "d", "D")
	}

	find := cmds.AddSub("find", "search", "/").
		RegPowerCmd(GlobalFindCmd,
			"find commands with strings").
		SetAllowTailModeCall().
		SetPriority()
	addFindStrArgs(find)

	findUsage := cmds.AddSub("find-detail", "//").
		RegPowerCmd(GlobalFindCmdWithUsage,
			"find commands with strings, with details").
		SetAllowTailModeCall().
		SetPriority()
	addFindStrArgs(findUsage)

	findDetail := cmds.AddSub("find-detail", "///").
		RegPowerCmd(GlobalFindCmdWithDetail,
			"find commands with strings, with details").
		SetAllowTailModeCall().
		SetPriority()
	addFindStrArgs(findDetail)

	findTag := cmds.AddSub("tags", "tag", cmds.Strs.TagMark).
		RegPowerCmd(FindByTags,
			"list commands having the specified tags").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority()
	addFindStrArgs(findTag)

	desc := cmds.AddSub("desc", "d", "D").
		RegPowerCmd(DumpFlowAll,
			"desc the flow about to execute").
		SetQuiet().
		SetPriority().
		AddArg("trivial", "1", "triv", "tri", "t", "T").
		AddArg("depth", "32", "d", "D")
	desc.AddSub("simple", "sim", "s", "S").
		RegPowerCmd(DumpFlowAllSimple,
			"desc the flow about to execute in lite style").
		SetQuiet().
		SetPriority().
		AddArg("trivial", "1", "triv", "tri", "t", "T").
		AddArg("depth", "32", "d", "D")
	desc.AddSub("skeleton", "sk", "sl", "st", "-").
		RegPowerCmd(DumpFlowSkeleton,
			"desc the flow about to execute, skeleton only").
		SetQuiet().
		SetPriority().
		AddArg("trivial", "1", "triv", "tri", "t", "T").
		AddArg("depth", "32", "d", "D")
	desc.AddSub("dependencies", "depends", "depend", "dep", "os-cmd", "os").
		RegPowerCmd(DumpFlowDepends,
			"list the depended os-commands of the flow").
		SetQuiet().
		SetPriority()
	desc.AddSub("env-ops-check", "env-ops", "env-op", "env", "ops", "op", "e", "E").
		RegPowerCmd(DumpFlowEnvOpsCheckResult,
			"desc the env-ops check result of the flow").
		SetQuiet().
		SetPriority()

	descFlow := desc.AddSub("flow", "f", "F").
		RegPowerCmd(DumpFlow,
			"desc the flow execution").
		SetQuiet().
		SetPriority().
		AddArg("trivial", "1", "triv", "tri", "t", "T").
		AddArg("depth", "32", "d", "D")
	descFlow.AddSub("simple", "sim", "s", "S", "-").
		RegPowerCmd(DumpFlowSimple,
			"desc the flow execution in lite style").
		SetQuiet().
		SetPriority().
		AddArg("trivial", "1", "triv", "tri", "t", "T").
		AddArg("depth", "32", "d", "D")

	cmds.AddSub("tail-info", "==").
		RegPowerCmd(DumpTailCmdInfo,
			"display the last command info, sub tree commands will not show").
		SetQuiet().
		SetPriority()

	cmds.AddSub("tail-sub", "~").
		RegPowerCmd(DumpTailCmdSub,
			"display commands on the branch of the last command").
		SetQuiet().
		SetPriority()

	cmds.AddSub("tail-sub-with-usage", "~~").
		RegPowerCmd(DumpTailCmdSubUsage,
			"display commands on the branch of the last command, with usage").
		SetQuiet().
		SetPriority()

	cmds.AddSub("tail-sub-with-details", "~~~").
		RegPowerCmd(DumpTailCmdSubDetails,
			"display commands on the branch of the last command, with details").
		SetQuiet().
		SetPriority()

	// TODO: asign 'cmds' to a freq-use command
	mods := cmds.AddSub("cmds", "cmd", "c", "C").
		RegPowerCmd(DumpCmdNoRecursive,
			"display command info").
		SetAllowTailModeCall().
		AddArg("cmd-path", "", "path", "p", "P")

	tree := mods.AddSub("tree", "t", "T")
	tree.RegPowerCmd(DumpCmdTree,
		"list builtin and loaded commands").
		SetAllowTailModeCall().
		AddArg("cmd-path", "", "path", "p", "P")
	tree.AddSub("simple", "sim", "skeleton", "sk", "sl", "st", "s", "S", "-").
		RegPowerCmd(DumpCmdTreeSkeleton,
			"list builtin and loaded commands, skeleton only").
		AddArg("cmd-path", "", "path", "p", "P").
		AddArg("recursive", "true", "r", "R")

	list := mods.AddSub("list", "ls", "flatten", "flat", "f", "F").
		RegPowerCmd(DumpCmdList,
			"list builtin and loaded commands").
		SetAllowTailModeCall()
	addFindStrArgs(list)

	listSimple := list.AddSub("simple", "sim", "s", "S", "-").
		RegPowerCmd(DumpCmdListSimple,
			"list builtin and loaded commands in lite style").
		SetAllowTailModeCall()
	addFindStrArgs(listSimple)
}

func RegisterFlowCmds(cmds *core.CmdTree) {
	listFlowsHelpStr := "list local saved but unlinked (to any repo) flows"
	flow := cmds.AddSub("flow", "fl", "f", "F").
		RegPowerCmd(ListFlows,
			listFlowsHelpStr).
		SetAllowTailModeCall()
	addFindStrArgs(flow)

	// TODO: check the new cmd is conflicted with other cmds
	flow.AddSub("save", "persist", "s", "S", "+").
		RegPowerCmd(SaveFlow,
			"save current commands as a flow").
		SetQuiet().
		SetPriority().
		AddArg("to-cmd-path", "", "path", "p", "P")

	flow.AddSub("set-help-str", "help", "h", "H").
		RegPowerCmd(SetFlowHelpStr,
			"set help str to a saved flow").
		SetQuiet().
		AddArg("cmd-path", "", "path", "p", "P").
		AddArg("help-str", "", "help", "h", "H")

	// TODO: check the new cmd is conflicted with other cmds
	flow.AddSub("rename", "rn").
		RegPowerCmd(RenameFlow,
			"rename a saved flow").
		AddArg("src", "", "src-cmd", "s", "S").
		AddArg("dest", "", "dest-cmd", "d", "D")

	flow.AddSub("remove", "rm", "delete", "del", "-").
		RegPowerCmd(RemoveFlow,
			"remove a saved flow").
		SetAllowTailModeCall().
		SetPriority().
		AddArg("cmd-path", "", "path", "p", "P")

	flowList := flow.AddSub("list-local", "list", "ls").
		RegPowerCmd(ListFlows,
			listFlowsHelpStr).
		SetAllowTailModeCall()
	addFindStrArgs(flowList)

	flow.AddSub("load", "l", "L").
		RegPowerCmd(LoadFlowsFromDir,
			"load flows from local dir").
		AddArg("path", "", "p", "P")

	flow.AddSub("clear", "reset", "--").
		RegPowerCmd(RemoveAllFlows,
			"remove all flows saved in local")

	flow.AddSub("move-flows-to-dir", "move", "mv", "m", "M").
		RegPowerCmd(MoveSavedFlowsToLocalDir,
			MoveFlowsToDirHelpStr).
		SetAllowTailModeCall().
		AddArg("path", "", "p", "P")
}

func RegisterEnvCmds(cmds *core.CmdTree) {
	env := cmds.AddSub("env", "e", "E").
		RegPowerCmd(DumpEssentialEnvFlattenVals,
			"list essential env values in flatten format").
		SetAllowTailModeCall()
	addFindStrArgs(env)

	assert := env.AddSub("assert")
	assert.AddSub("equal").
		RegPowerCmd(EnvAssertEqual,
			"assert the value of a key in env equal to specified value").
		AddArg("key", "", "k", "K").
		AddArg("val", "", "value", "v", "V")
	assert.AddSub("not-exists").
		RegPowerCmd(EnvAssertNotExists,
			"assert the key not in env").
		AddArg("key", "", "k", "K")

	env.AddSub("tree", "t", "T").
		RegPowerCmd(DumpEnvTree,
			"list all env layers and values in tree format")

	// TODO: add search supporting
	abbrs := env.AddSub("abbrs", "abbr", "a", "A")
	abbrs.RegPowerCmd(DumpEnvAbbrs,
		"list env tree and abbrs")

	envList := env.AddSub("list", "ls", "flatten", "flat", "f", "F").
		RegPowerCmd(DumpEnvFlattenVals,
			"list env values in flatten format").
		SetAllowTailModeCall()
	addFindStrArgs(envList)

	env.AddSub("save", "persist", "s", "S", "+").
		RegPowerCmd(SaveEnvToLocal,
			"save session env changes to local").
		SetQuiet()

	env.AddSub("mapping", "map").
		RegPowerCmd(MapEnvKeyValueToAnotherKey,
			"read src-key's value and write to dest-key").
		AddArg("src-key", "", "source-key", "source", "src", "from").
		AddArg("dest-key", "", "dest", "to").
		AddEnvOp("[[dest-key]]", core.EnvOpTypeWrite)

	env.AddSub("remove-and-save", "remove", "rm", "delete", "del", "-").
		RegPowerCmd(RemoveEnvValAndSaveToLocal,
			"remove specified env value and save changes to local").
		SetAllowTailModeCall().
		AddArg("key", "", "k", "K")

	env.AddSub("reset-session", "reset", "--").
		RegPowerCmd(ResetSessionEnv,
			"clear all env values in current session")

	env.AddSub("reset-and-save", "clear", "---").
		RegPowerCmd(ResetLocalEnv,
			"clear all local saved env values")

	env.AddSub("who-write", "ww").
		RegPowerCmd(DumpCmdsWhoWriteKey,
			"find which commands write the specified key").
		SetAllowTailModeCall().
		AddArg("key", "", "k", "K")

	registerSimpleSwitch(abbrs,
		"borrowing commands' abbrs when setting env key-values",
		"sys.env.use-cmd-abbrs",
		"cmd")
}

func RegisterVerbCmds(cmds *core.CmdTree) {
	cmds.AddSub("quiet", "q", "Q").
		RegPowerCmd(SetQuietMode,
			"change into quiet mode").
		SetQuiet()

	verbose := cmds.AddSub("verbose", "verb", "v", "V")

	verbose.RegPowerCmd(SetVerbMode,
		"change into verbose mode").
		SetQuiet()

	verbose.AddSub("default", "def", "d", "D").
		RegPowerCmd(SetToDefaultVerb,
			"set to default verbose mode").
		SetQuiet()

	verbose.AddSub("increase", "inc", "v+", "+").
		RegPowerCmd(IncreaseVerb,
			"increase verbose").
		SetQuiet().
		AddArg("volume", "1", "vol", "v", "V")

	verbose.AddSub("decrease", "dec", "v-", "-").
		RegPowerCmd(DecreaseVerb,
			"decrease verbose").
		SetQuiet().
		AddArg("volume", "1", "vol", "v", "V")
}

func RegisterSessionCmds(cmds *core.CmdTree) {
	list := cmds.AddSub("list", "ls").
		RegPowerCmd(ListSessions,
			"list executed/ing sessions").
		SetAllowTailModeCall()
	addFindStrArgs(list)

	listDesc := list.AddSub("desc", "desc-less", "-").
		RegPowerCmd(ListedSessionDescLess,
			"desc executed/ing session").
		SetAllowTailModeCall()
	addFindStrArgs(listDesc)
	listDesc.AddArg("trivial", "1", "triv", "tri", "t", "T")
	listDesc.AddArg("depth", "32", "d", "D")

	listDescMore := list.AddSub("desc-more", "+").
		RegPowerCmd(ListedSessionDescMore,
			"desc executed/ing session with more details").
		SetAllowTailModeCall()
	addFindStrArgs(listDescMore)
	listDescMore.AddArg("trivial", "1", "triv", "tri", "t", "T")
	listDescMore.AddArg("depth", "32", "d", "D")

	listDescFull := list.AddSub("desc-full", "++").
		RegPowerCmd(ListedSessionDescFull,
			"desc executed/ing session with full details").
		SetAllowTailModeCall()
	addFindStrArgs(listDescFull)
	listDescFull.AddArg("trivial", "1", "triv", "tri", "t", "T")
	listDescFull.AddArg("depth", "32", "d", "D")

	last := cmds.AddSub("last", "l", "L")
	last.RegPowerCmd(LastSession,
		"show last session")

	last.AddSub("desc", "desc-less", "-").
		RegPowerCmd(LastSessionDescLess,
			"desc the execution status of last session").
		AddArg("trivial", "1", "triv", "tri", "t", "T").
		AddArg("depth", "32", "d", "D")

	last.AddSub("desc-more", "+").
		RegPowerCmd(LastSessionDescMore,
			"desc the execution status of last session with more details").
		AddArg("trivial", "1", "triv", "tri", "t", "T").
		AddArg("depth", "32", "d", "D")

	last.AddSub("desc-full", "++").
		RegPowerCmd(LastSessionDescFull,
			"desc the execution status of last session with full details").
		AddArg("trivial", "1", "triv", "tri", "t", "T").
		AddArg("depth", "32", "d", "D")

	cmds.AddSub("remove-all", "clean", "clear", "--").
		RegPowerCmd(RemoveAllSessions,
			"clean executed sessions")

	remove := cmds.AddSub("remove", "delete", "rm").
		RegPowerCmd(FindAndRemoveSessions,
			"clean executed sessions by find-strs").
		SetAllowTailModeCall()
	addFindStrArgs(remove)

	cmds.AddSub("set-keep-duration", "set-keep-dur", "keep-duration", "keep-dur", "k-d", "kd").
		RegPowerCmd(SetSessionsKeepDur,
			"set the keeping duration of executed sessions").
		AddArg("duration", "72h", "dur")
}

func RegisterHubCmds(cmds *core.CmdTree) {
	listHubHelpStr := "list dir and repo info in hub"
	hub := cmds.AddSub("hub", "h", "H").
		RegPowerCmd(ListHub,
			listHubHelpStr).
		SetAllowTailModeCall()
	addFindStrArgs(hub)

	hub.AddSub("clear", "reset", "--").
		RegPowerCmd(RemoveAllFromHub,
			"remove all repos from hub")

	hub.AddSub("init", "++").
		RegPowerCmd(AddDefaultGitRepoToHub,
			"add and pull basic hub-repo to local")

	add := hub.AddSub("add-and-update", "add", "a", "A", "+")
	add.RegPowerCmd(AddGitRepoToHub,
		"add and pull a git address to hub, do update if it already exists").
		SetAllowTailModeCall().
		AddArg("git-address", "", "git", "address", "addr")

	hub.AddSub("git-status", "status").
		RegPowerCmd(CheckGitRepoStatus,
			"check git status for all repos")

	add.AddSub("local-dir", "local", "l", "L").
		RegPowerCmd(AddLocalDirToHub,
			"add a local dir (could be a git repo) to hub").
		SetAllowTailModeCall().
		AddArg("path", "", "p", "P")

	hubList := hub.AddSub("list", "ls").
		RegPowerCmd(ListHub,
			listHubHelpStr).
		SetAllowTailModeCall()
	addFindStrArgs(hubList)

	purge := hub.AddSub("purge", "p", "P", "-")
	purge.RegPowerCmd(PurgeInactiveRepoFromHub,
		"remove an inactive repo from hub").
		SetAllowTailModeCall().
		AddArg("find-str", "", "s", "S")
	purge.AddSub("purge-all-inactive", "all", "inactive", "a", "A", "-").
		RegPowerCmd(PurgeAllInactiveReposFromHub,
			"remove all inactive repos from hub")

	hub.AddSub("update-all", "update", "u", "U").
		RegPowerCmd(UpdateHub,
			"update all repos and mods defined in hub")

	hub.AddSub("enable-repo", "enable", "ena", "en", "e", "E").
		RegPowerCmd(EnableRepoInHub,
			"enable matched git repos in hub").
		SetAllowTailModeCall().
		AddArg("find-str", "", "s", "S")

	hub.AddSub("disable-repo", "disable", "dis", "d", "D").
		RegPowerCmd(DisableRepoInHub,
			"disable matched git repos in hub").
		SetAllowTailModeCall().
		AddArg("find-str", "", "s", "S")

	hub.AddSub("move-flows-to-dir", "move", "mv", "m", "M").
		RegPowerCmd(MoveSavedFlowsToLocalDir,
			MoveFlowsToDirHelpStr).
		SetAllowTailModeCall().
		AddArg("path", "", "p", "P")
}

func RegisterBuiltinCmds(cmds *core.CmdTree) {
	env := cmds.AddSub("env", "e", "E")

	envLoad := env.AddSub("load", "l", "L")

	envLoad.AddSub("local", "l", "L").
		RegPowerCmd(LoadLocalEnv,
			"load env values from local").
		SetQuiet()

	envLoad.AddSub("runtime", "rt", "r", "R").
		RegPowerCmd(LoadRuntimeEnv,
			"setup runtime env values").
		SetQuiet()

	mod := cmds.AddSub("mod", "mods", "m", "M")

	modLoad := mod.AddSub("load", "l", "L")

	modLoad.AddSub("flows", "flows", "f", "F").
		RegPowerCmd(LoadFlows,
			"load saved flows from local")

	modLoad.AddSub("ext-exec", "ext", "e", "E").
		RegPowerCmd(SetExtExec,
			"load default setting of how to run a executable file by ext name")

	modLoad.AddSub("hub", "h", "H").
		RegPowerCmd(LoadModsFromHub,
			"load flows and mods from local hub")

	cmds.AddSub("display", "disp", "dis", "di", "d", "D").
		AddSub("load", "l", "L").
		AddSub("platform", "p", "P").
		RegPowerCmd(LoadPlatformDisplay,
			"load platform(OS) specialized display settings").
		SetQuiet()
}

func RegisterTrivialCmds(cmds *core.CmdTree) {
	cmds.AddSub("noop").
		RegPowerCmd(Noop,
			"do exactly nothing")

	cmds.AddSub("dummy", "dmy", "dm").
		RegPowerCmd(Dummy,
			"dummy command for testing")

	cmds.AddSub("sleep", "slp").
		RegPowerCmd(Sleep,
			"sleep for specified duration").
		AddArg("duration", "1s", "dur", "d", "D")
}

func RegisterBgCmds(cmds *core.CmdTree) {
	cmds.AddSub("wait").
		RegPowerCmd(WaitForAllBgTasksFinish,
			"wait for all tasks/threads to finish in current(must be main) thread")
}

func RegisterMiscCmds(cmds *core.CmdTree) {
	cmds.AddSub("mark-time", "time").
		RegPowerCmd(MarkTime,
			"set current timestamp to the specified key").
		AddArg("write-to-key", "", "key", "k", "K").
		AddEnvOp("[[write-to-key]]", core.EnvOpTypeWrite)

	timer := cmds.AddSub("timer")
	timer.AddSub("begin").
		RegPowerCmd(TimerBegin,
			"start timer, set current timestamp to the specified key").
		AddArg("begin-key", "", "key", "k", "K").
		AddEnvOp("[[begin-key]]", core.EnvOpTypeWrite)
	timer.AddSub("elapsed", "elapse", "end").
		RegPowerCmd(TimerElapsed,
			"set elapsed seconds to the specified key").
		AddArg("begin-key", "", "begin").
		AddArg("write-to-key", "", "key", "k", "K").
		AddEnvOp("[[begin-key]]", core.EnvOpTypeRead).
		AddEnvOp("[[write-to-key]]", core.EnvOpTypeWrite)
}

// This cmds are for debug
func RegisterDbgCmds(cmds *core.CmdTree) {
	registerSimpleSwitch(cmds,
		"step by step on executing",
		"sys.step-by-step",
		"step-by-step", "step", "confirm", "cfm")

	registerSimpleSwitch(cmds,
		"recover from internal error and give a frendly message",
		"sys.panic.recover",
		"recover")

	cmds.AddSub("delay-execute", "delay", "dl", "d", "D").
		RegPowerCmd(DbgDelayExecute,
			"wait for a while before executing each commands").
		SetQuiet().
		AddArg("seconds", "3", "second", "sec", "s", "S")

	cmds.AddSub("echo").
		RegPowerCmd(DbgEcho,
			"print message from argv").
		AddArg("message", "", "msg", "m", "M")

	panicTest := cmds.AddSub("panic")
	panicTest.RegPowerCmd(DbgPanic,
		"for panic test").
		AddArg("random-arg-1", "arg-1").
		AddArg("random-arg-2", "arg-2")

	panicTest.AddSub("cmd").
		RegPowerCmd(DbgPanicCmdError,
			"for specified-panic test").
		AddArg("random-arg-1", "arg-1").
		AddArg("random-arg-2", "arg-2")

	cmds.AddSub("error").
		RegPowerCmd(DbgError,
			"for execute error test").
		AddArg("random-arg-1", "arg-1").
		AddArg("random-arg-2", "arg-2")

	cmds.AddSub("exec").SetHidden().
		RegPowerCmd(DbgExecBash,
			"verify bash in os/exec").
		SetQuiet()
}

func RegisterDisplayCmds(cmds *core.CmdTree) {
	cmds.AddSub("style").
		RegPowerCmd(SetDisplayStyle,
			"set executing display style: bold, slash, corner, ascii, utf8(default)").
		AddArg("style", "s", "S").
		SetQuiet()

	utf8 := registerSimpleSwitchEx(cmds,
		"utf8 display",
		[]string{"display.utf8", "display.utf8.symbols"},
		"utf8", "utf")

	registerSimpleSwitch(utf8,
		"utf8 symbols display",
		"display.utf8.symbols",
		"symbols", "symbol", "sym")

	registerSimpleSwitch(cmds,
		"color display",
		"display.color",
		"color", "colors", "clr")

	env := registerSimpleSwitch(cmds,
		"env display",
		"display.env",
		"env")

	sys := registerSimpleSwitch(env,
		"values of env path 'sys.*' display in executing",
		"display.env.sys",
		"sys")

	registerSimpleSwitch(sys,
		"values of env path 'sys.paths.*' display in executing",
		"display.env.sys.paths",
		"paths", "path")

	registerSimpleSwitch(env,
		"values of env path 'display.*' display in executing",
		"display.env.display",
		"display", "disp", "dis", "di")

	registerSimpleSwitch(cmds,
		"stack display",
		"display.stack",
		"stack")

	registerSimpleSwitch(env,
		"env layer display in executing",
		"display.env.layer",
		"layer")

	registerSimpleSwitch(env,
		"env default layer display in executing",
		"display.env.default",
		"default", "def")

	registerSimpleSwitch(cmds,
		"meow display",
		"display.meow",
		"meow")

	mod := cmds.AddSub("mod")

	registerSimpleSwitch(mod,
		"quiet module display in executing",
		"display.mod.quiet",
		"quiet")

	registerSimpleSwitch(mod,
		"display realname of module in executing",
		"display.mod.realname",
		"realname", "real")

	registerSimpleSwitch(cmds,
		"executor display",
		"display.executor",
		"executor", "executer", "exe")

	registerSimpleSwitch(cmds,
		"executor finish footer display",
		"display.executor.end",
		"end", "footer", "foot")

	registerSimpleSwitch(cmds,
		"bootstrap executing display",
		"display.bootstrap",
		"bootstrap", "boot")

	registerSimpleSwitch(cmds,
		"display executor when only executing one command",
		"display.one-cmd",
		"one-cmd", "one")

	cmds.AddSub("set-width", "width", "wid", "w", "W").
		RegEmptyCmd(
			"set display width").
		SetQuiet().
		AddArg("width", "120", "wid", "w", "W").
		AddArg2Env("display.width", "width")
}

func registerSimpleSwitch(
	parent *core.CmdTree,
	function string,
	key string,
	name string,
	abbrs ...string) *core.CmdTree {

	return registerSimpleSwitchEx(parent, function, []string{key}, name, abbrs...)
}

func registerSimpleSwitchEx(
	parent *core.CmdTree,
	function string,
	keys []string,
	name string,
	abbrs ...string) *core.CmdTree {

	cmd := parent.AddSub(name, abbrs...)
	self := cmd.RegEmptyCmd("enable " + function).SetQuiet()
	on := cmd.AddSub("on", "yes", "y", "Y", "1", "+").RegEmptyCmd("enable " + function).SetQuiet()
	off := cmd.AddSub("off", "no", "n", "N", "0", "-").RegEmptyCmd("disable " + function).SetQuiet()

	for _, key := range keys {
		self.AddVal2Env(key, "true")
		on.AddVal2Env(key, "true")
		off.AddVal2Env(key, "false")
	}

	return self.Owner()
}

const LessHelpStr = "desc the flow about to execute"
const MoreHelpStr = LessHelpStr + ", with details"

const MoveFlowsToDirHelpStr = `move all saved flows to a local dir (could be a git repo).
auto move:
    * if one(and only one) local(not linked to a repo) dir exists in hub
    * and the arg "path" is empty
    then flows will move to that dir`
