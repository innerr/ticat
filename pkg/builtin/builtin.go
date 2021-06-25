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
	RegisterDisplayCmds(cmds.AddSub("display", "disp", "dis", "di"))
	RegisterBuiltinCmds(cmds.AddSub("builtin", "b", "B").SetHidden())
}

func RegisterExecutorCmds(cmds *core.CmdTree) {
	cmds.AddSub("-help", "-HELP", "-h", "-H").
		RegCmd(GlobalHelp,
			"get help")

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

	cmds.AddSub("tail-info", "=").
		RegPowerCmd(DumpTailCmdInfo,
			"display the last command info, sub tree commands will not show").
		SetQuiet().
		SetPriority()

	cmds.AddSub("tail-sub", "$").
		RegPowerCmd(DumpTailCmdSub,
			"display commands on the branch of the last command").
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
			"remove specific env KV and save changes to local").
		AddArg("key", "", "k", "K")

	env.AddSub("reset-and-save", "reset", "clear", "--").
		RegCmd(ResetLocalEnv,
			"reset all local saved env KVs")

	abbrsCmdHelpStr := "enable borrowing commands' abbrs when setting KVs"
	abbrsCmd := abbrs.AddSub("cmd")
	abbrsCmd.RegEmptyCmd(
		abbrsCmdHelpStr).
		AddVal2Env("sys.env.use-cmd-abbrs", "true")
	abbrsCmd.AddSub("on", "yes", "y", "Y", "1", "+").
		RegEmptyCmd(
			abbrsCmdHelpStr).
		AddVal2Env("sys.env.use-cmd-abbrs", "true")
	abbrsCmd.AddSub("off", "no", "n", "N", "0", "-").
		RegEmptyCmd(
			"disable borrowing commands' abbrs when setting KVs").
		AddVal2Env("sys.env.use-cmd-abbrs", "false")
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
			"sleep for specific duration").
		AddArg("duration", "1s", "dur", "d", "D")
}

// This cmds are for debug
func RegisterDbgCmds(cmds *core.CmdTree) {
	cmds.AddSub("echo").
		RegCmd(DbgEcho,
			"print message from argv").
		AddArg("message", "", "msg", "m", "M")

	step := cmds.AddSub("step-by-step", "step", "s", "S")
	step.RegEmptyCmd(
		"enable step by step").
		AddVal2Env("sys.step-by-step", "true").
		SetQuiet()
	step.AddSub("on", "yes", "y", "Y", "1", "+").
		RegEmptyCmd(
			"enable step by step").
		AddVal2Env("sys.step-by-step", "true").
		SetQuiet()
	step.AddSub("off", "no", "n", "N", "0", "-").
		RegEmptyCmd(
			"disable step by step").
		AddVal2Env("sys.step-by-step", "false").
		SetQuiet()

	cmds.AddSub("delay-execute", "delay", "dl", "d", "D").
		RegCmd(DbgDelayExecute,
			"wait for a while before executing a command").
		SetQuiet().
		AddArg("seconds", "3", "second", "sec", "s", "S")

	cmds.AddSub("exec").SetHidden().
		RegCmd(DbgExecBash,
			"verify bash in os/exec")
}

func RegisterDisplayCmds(cmds *core.CmdTree) {
	utf8 := cmds.AddSub("utf8", "utf")
	utf8.RegEmptyCmd(
		"enable utf8 display").
		AddVal2Env("display.utf8", "true").
		AddVal2Env("display.utf8.symbols", "true").
		SetQuiet()
	utf8.AddSub("on", "yes", "y", "Y", "1", "+").
		RegEmptyCmd(
			"enable utf8 display").
		AddVal2Env("display.utf8", "true").
		AddVal2Env("display.utf8.symbols", "true").
		SetQuiet()
	utf8.AddSub("off", "no", "n", "N", "0", "-").
		RegEmptyCmd(
			"disable utf8 display").
		AddVal2Env("display.utf8", "false").
		AddVal2Env("display.utf8.symbols", "false").
		SetQuiet()
}

const LessHelpStr = "display/search info base on the current flow and args"
const MoreHelpStr = LessHelpStr + ", with details"

const MoveFlowsToDirHelpStr = `move all saved flows to a local dir (could be a git repo).
auto move:
    * if one(and only one) local(not linked to a repo) dir exists in hub
    * and the arg "path" is empty
    then flows will move to that dir`
