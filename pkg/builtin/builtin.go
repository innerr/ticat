package builtin

import (
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
	RegisterDbgCmds(cmds.AddSub("dbg"))
	RegisterMiscCmds(cmds)
	RegisterDisplayCmds(cmds.AddSub("display", "disp", "dis", "di"))
	RegisterBuiltinCmds(cmds.AddSub("builtin", "b", "B").SetHidden())
}

func RegisterExecutorCmds(cmds *core.CmdTree) {
	help := cmds.AddSub("-help", "-HELP", "help", "-h", "-H")
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

	more := cmds.AddSub("more", "+").
		RegPowerCmd(GlobalHelpMoreInfo,
			MoreHelpStr).
		SetQuiet().
		SetPriority()
	addFindStrArgs(more)

	less := cmds.AddSub("less", "-").
		RegPowerCmd(GlobalHelpLessInfo,
			LessHelpStr).
		SetQuiet().
		SetPriority()
	addFindStrArgs(less)

	find := cmds.AddSub("search", "find", "fnd", "s", "S").
		RegCmd(FindAny,
			"find anything with given string")
	addFindStrArgs(find)

	findTag := cmds.AddSub("tags", "tag", "@").
		RegPowerCmd(FindByTags,
			"list commands having the specified tags").
		SetQuiet().
		SetPriority()
	addFindStrArgs(findTag)

	desc := cmds.AddSub("desc", "d", "D").
		RegPowerCmd(DumpFlowAll,
			"desc the flow about to execute").
		SetQuiet().
		SetPriority()
	desc.AddSub("simple", "sim", "s", "S").
		RegPowerCmd(DumpFlowAllSimple,
			"desc the flow about to execute in lite style").
		SetQuiet().
		SetPriority()
	desc.AddSub("skeleton", "sk", "sl", "st", "-").
		RegPowerCmd(DumpFlowSkeleton,
			"desc the flow about to execute, skeleton only").
		SetQuiet().
		SetPriority()
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

	desc.AddSub("depth").
		RegCmd(SetDumpFlowDepth,
			"setup display stack depth of flow desc").
		SetQuiet().
		SetPriority().
		AddArg("depth", "8", "d", "D")

	descFlow := desc.AddSub("flow", "f", "F").
		RegPowerCmd(DumpFlow,
			"desc the flow execution").
		SetQuiet().
		SetPriority()
	descFlow.AddSub("simple", "sim", "s", "S", "-").
		RegPowerCmd(DumpFlowSimple,
			"desc the flow execution in lite style").
		SetQuiet().
		SetPriority()

	cmds.AddSub("tail-info", "==").
		RegPowerCmd(DumpTailCmdInfo,
			"display the last command info, sub tree commands will not show").
		SetQuiet().
		SetPriority()

	cmds.AddSub("tail-sub-less", "--").
		RegPowerCmd(DumpTailCmdSubLess,
			"display commands on the branch of the last command").
		SetQuiet().
		SetPriority()

	cmds.AddSub("tail-sub-more", "++").
		RegPowerCmd(DumpTailCmdSubMore,
			"display commands on the branch of the last command, with details").
		SetQuiet().
		SetPriority()

	mods := cmds.AddSub("cmds", "cmd", "c", "C")
	mods.RegCmd(DumpCmdNoRecursive,
		"display command info, sub tree commands will not show").
		AddArg("cmd-path", "", "path", "p", "P")

	tree := mods.AddSub("tree", "t", "T")
	tree.RegCmd(DumpCmdTree,
		"list builtin and loaded commands").
		AddArg("cmd-path", "", "path", "p", "P")
	tree.AddSub("simple", "sim", "skeleton", "sk", "sl", "st", "s", "S", "-").
		RegCmd(DumpCmdTreeSkeleton,
			"list builtin and loaded commands, skeleton only").
		AddArg("cmd-path", "", "path", "p", "P")

	list := mods.AddSub("list", "ls", "flatten", "flat", "f", "F", "~").
		RegCmd(DumpCmdList,
			"list builtin and loaded commands")
	addFindStrArgs(list)

	listSimple := list.AddSub("simple", "sim", "s", "S", "-").
		RegCmd(DumpCmdListSimple,
			"list builtin and loaded commands in lite style")
	addFindStrArgs(listSimple)
}

func RegisterFlowCmds(cmds *core.CmdTree) {
	listFlowsHelpStr := "list local saved but unlinked (to any repo) flows"
	flow := cmds.AddSub("flow", "fl", "f", "F").
		RegCmd(ListFlows,
			listFlowsHelpStr)
	addFindStrArgs(flow)

	flow.AddSub("save", "persist", "s", "S", "+").
		RegPowerCmd(SaveFlow,
			"save current commands as a flow").
		SetQuiet().
		SetPriority().
		AddArg("to-cmd-path", "", "path", "p", "P")

	flow.AddSub("set-help-str", "help", "h", "H").
		RegCmd(SetFlowHelpStr,
			"set help str to a saved flow").
		SetQuiet().
		AddArg("cmd-path", "", "path", "p", "P").
		AddArg("help-str", "", "help", "h", "H")

	flow.AddSub("remove", "rm", "delete", "del", "-").
		RegCmd(RemoveFlow,
			"remove a saved flow").
		AddArg("cmd-path", "", "path", "p", "P")

	flowList := flow.AddSub("list-local", "list", "ls", "~").
		RegCmd(ListFlows,
			listFlowsHelpStr)
	addFindStrArgs(flowList)

	flow.AddSub("load", "l", "L").
		RegCmd(LoadFlowsFromDir,
			"load flows from local dir").
		AddArg("path", "", "p", "P")

	flow.AddSub("clear", "reset", "--").
		RegCmd(RemoveAllFlows,
			"remove all flows saved in local")

	flow.AddSub("move-flows-to-dir", "move", "mv", "m", "M").
		RegCmd(MoveSavedFlowsToLocalDir,
			MoveFlowsToDirHelpStr).
		AddArg("path", "", "p", "P")
}

func RegisterEnvCmds(cmds *core.CmdTree) {
	env := cmds.AddSub("env", "e", "E").
		RegCmd(DumpEssentialEnvFlattenVals,
			"list essential env values in flatten format")
	addFindStrArgs(env)

	env.AddSub("tree", "t", "T").
		RegCmd(DumpEnvTree,
			"list all env layers and KVs in tree format")

	// TODO: add search supporting
	abbrs := env.AddSub("abbrs", "abbr", "a", "A")
	abbrs.RegCmd(DumpEnvAbbrs,
		"list env tree and abbrs")

	envList := env.AddSub("list", "ls", "flatten", "flat", "f", "F", "~").
		RegCmd(DumpEnvFlattenVals,
			"list env values in flatten format")
	addFindStrArgs(envList)

	env.AddSub("save", "persist", "s", "S", "+").
		RegCmd(SaveEnvToLocal,
			"save session env changes to local").
		SetQuiet()

	env.AddSub("remove-and-save", "remove", "rm", "delete", "del", "-").
		RegCmd(RemoveEnvValAndSaveToLocal,
			"remove specified env KV and save changes to local").
		AddArg("key", "", "k", "K")

	env.AddSub("reset-and-save", "reset", "clear", "--").
		RegCmd(ResetLocalEnv,
			"reset all local saved env KVs")

	registerSimpleSwitch(abbrs,
		"borrowing commands' abbrs when setting KVs",
		"sys.env.use-cmd-abbrs",
		"cmd")
}

func RegisterVerbCmds(cmds *core.CmdTree) {
	cmds.AddSub("quiet", "q", "Q").
		RegCmd(SetQuietMode,
			"change into quiet mode").
		SetQuiet()

	verbose := cmds.AddSub("verbose", "verb", "v", "V")

	verbose.RegCmd(SetVerbMode,
		"change into verbose mode").
		SetQuiet()

	verbose.AddSub("default", "def", "d", "D").
		RegCmd(SetToDefaultVerb,
			"set to default verbose mode").
		SetQuiet()

	verbose.AddSub("increase", "inc", "v+", "+").
		RegCmd(IncreaseVerb,
			"increase verbose").
		SetQuiet().
		AddArg("volume", "1", "vol", "v", "V")

	verbose.AddSub("decrease", "dec", "v-", "-").
		RegCmd(DecreaseVerb,
			"decrease verbose").
		SetQuiet().
		AddArg("volume", "1", "vol", "v", "V")
}

func RegisterHubCmds(cmds *core.CmdTree) {
	listHubHelpStr := "list dir and repo info in hub"
	hub := cmds.AddSub("hub", "h", "H").
		RegCmd(ListHub,
			listHubHelpStr)
	addFindStrArgs(hub)

	hub.AddSub("clear", "reset", "--").
		RegCmd(RemoveAllFromHub,
			"remove all repos from hub")

	hub.AddSub("init", "++").
		RegCmd(AddGitDefaultToHub,
			"add and pull basic hub-repo to local")

	add := hub.AddSub("add-and-update", "add", "a", "A", "+")
	add.RegCmd(AddGitRepoToHub,
		"add and pull a git address to hub, do update if it already exists").
		AddArg("git-address", "", "git", "address", "addr")

	add.AddSub("local-dir", "local", "l", "L").
		RegCmd(AddLocalDirToHub,
			"add a local dir (could be a git repo) to hub").
		AddArg("path", "", "p", "P")

	hubList := hub.AddSub("list", "ls", "~").
		RegCmd(ListHub,
			listHubHelpStr)
	addFindStrArgs(hubList)

	purge := hub.AddSub("purge", "p", "P", "-")
	purge.RegCmd(PurgeInactiveRepoFromHub,
		"remove an inactive repo from hub").
		AddArg("find-str", "", "s", "S")
	purge.AddSub("purge-all-inactive", "all", "inactive", "a", "A", "-").
		RegCmd(PurgeAllInactiveReposFromHub,
			"remove all inactive repos from hub")

	hub.AddSub("update-all", "update", "u", "U").
		RegCmd(UpdateHub,
			"update all repos and mods defined in hub")

	hub.AddSub("enable-repo", "enable", "ena", "en", "e", "E").
		RegCmd(EnableRepoInHub,
			"enable matched git repos in hub").
		AddArg("find-str", "", "s", "S")

	hub.AddSub("disable-repo", "disable", "dis", "d", "D").
		RegCmd(DisableRepoInHub,
			"disable matched git repos in hub").
		AddArg("find-str", "", "s", "S")

	hub.AddSub("move-flows-to-dir", "move", "mv", "m", "M").
		RegCmd(MoveSavedFlowsToLocalDir,
			MoveFlowsToDirHelpStr).
		AddArg("path", "", "p", "P")
}

func RegisterBuiltinCmds(cmds *core.CmdTree) {
	env := cmds.AddSub("env", "e", "E")

	envLoad := env.AddSub("load", "l", "L")

	envLoad.AddSub("local", "l", "L").
		RegCmd(LoadLocalEnv,
			"load env KVs from local").
		SetQuiet()

	envLoad.AddSub("runtime", "rt", "r", "R").
		RegCmd(LoadRuntimeEnv,
			"setup runtime env KVs").
		SetQuiet()

	mod := cmds.AddSub("mod", "mods", "m", "M")

	modLoad := mod.AddSub("load", "l", "L")

	modLoad.AddSub("flows", "flows", "f", "F").
		RegCmd(LoadFlows,
			"load saved flows from local")

	modLoad.AddSub("ext-exec", "ext", "e", "E").
		RegCmd(SetExtExec,
			"load default setting of how to run a executable file by ext name")

	modLoad.AddSub("hub", "h", "H").
		RegCmd(LoadModsFromHub,
			"load flows and mods from local hub")

	cmds.AddSub("display", "disp", "dis", "di", "d", "D").
		AddSub("load", "l", "L").
		AddSub("platform", "p", "P").
		RegCmd(LoadPlatformDisplay,
			"load platform(OS) specialized display settings").
		SetQuiet()
}

func RegisterTrivialCmds(cmds *core.CmdTree) {
	dummy := cmds.AddSub("dummy", "dmy", "dm")

	dummy.RegCmd(Dummy,
		"dummy command for testing")

	dummy.AddSub("quiet", "q", "Q").
		RegCmd(QuietDummy,
			"quiet dummy command for testing").
		SetQuiet()

	dummy.AddSub("power", "p", "P").
		RegPowerCmd(PowerDummy,
			"power dummy command for testing")

	dummy.AddSub("priority", "prior", "prio", "pri").
		RegPowerCmd(PriorityPowerDummy,
			"power dummy command for testing").
		SetPriority()

	cmds.AddSub("sleep", "slp").
		RegCmd(Sleep,
			"sleep for specified duration").
		AddArg("duration", "1s", "dur", "d", "D")
}

func RegisterMiscCmds(cmds *core.CmdTree) {
	cmds.AddSub("mark-time", "time").
		RegCmd(Time,
			"set current timestamp to the specified key").
		AddArg("write-to-key", "key", "k", "K")
}

// This cmds are for debug
func RegisterDbgCmds(cmds *core.CmdTree) {
	registerSimpleSwitch(cmds,
		"step by step on executing",
		"sys.step-by-step",
		"step-by-step", "step", "confirm", "cfm")

	cmds.AddSub("delay-execute", "delay", "dl", "d", "D").
		RegCmd(DbgDelayExecute,
			"wait for a while before executing a command").
		SetQuiet().
		AddArg("seconds", "3", "second", "sec", "s", "S")

	cmds.AddSub("echo").
		RegCmd(DbgEcho,
			"print message from argv").
		AddArg("message", "", "msg", "m", "M")

	panicTest := cmds.AddSub("panic")
	panicTest.RegCmd(DbgPanic,
		"for panic test").
		AddArg("random-arg-1", "arg-1").
		AddArg("random-arg-2", "arg-2")

	panicTest.AddSub("cmd").
		RegCmd(DbgPanicCmdError,
			"for specified-panic test").
		AddArg("random-arg-1", "arg-1").
		AddArg("random-arg-2", "arg-2")

	cmds.AddSub("error").
		RegCmd(DbgError,
			"for execute error test").
		AddArg("random-arg-1", "arg-1").
		AddArg("random-arg-2", "arg-2")

	cmds.AddSub("exec").SetHidden().
		RegCmd(DbgExecBash,
			"verify bash in os/exec")
}

func RegisterDisplayCmds(cmds *core.CmdTree) {
	cmds.AddSub("style").
		RegCmd(SetDisplayStyle,
			"set executing display style: bold, slash, corner, ascii, utf8(default)").
		AddArg("style", "s", "S").
		SetQuiet()

	registerSimpleSwitchEx(cmds,
		"utf8 display",
		[]string{"display.utf8", "display.utf8.symbols"},
		"utf8", "utf")

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
		"display.env.sys",
		"sys")

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

const LessHelpStr = "display/search info base on the current flow and args"
const MoreHelpStr = LessHelpStr + ", with details"

const MoveFlowsToDirHelpStr = `move all saved flows to a local dir (could be a git repo).
auto move:
    * if one(and only one) local(not linked to a repo) dir exists in hub
    * and the arg "path" is empty
    then flows will move to that dir`
