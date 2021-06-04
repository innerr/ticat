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
	RegisterBuiltinCmds(cmds.AddSub("builtin", "b", "B"))
}

func RegisterExecutorCmds(cmds *core.CmdTree) {
	more := cmds.AddSub("more", "+").
		RegPowerCmd(GlobalHelp,
			GlobalHelpHelpStr).
		SetQuiet().
		SetPriority()
	addFindStrArgs(more)

	less := cmds.AddSub("less", "-").
		RegPowerCmd(GlobalSkeleton,
			SkeletonHelpStr).
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

	desc.AddSub("depth", "d", "D").
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

	cmds.AddSub("tail", "$").
		RegPowerCmd(DumpTailCmd,
			"display the last cmd info, sub tree cmds will not show").
		SetQuiet().
		SetPriority()

	mods := cmds.AddSub("cmds", "cmd", "c", "C")
	mods.RegCmd(DumpCmdNoRecursive,
		"display cmd info, sub tree cmds will not show").
		AddArg("cmd-path", "", "path", "p", "P")

	tree := mods.AddSub("tree", "t", "T")
	tree.RegCmd(DumpCmdTree,
		"list builtin and loaded cmds").
		AddArg("cmd-path", "", "path", "p", "P")
	tree.AddSub("simple", "sim", "skeleton", "sk", "sl", "st", "s", "S", "-").
		RegCmd(DumpCmdTreeSkeleton,
			"list builtin and loaded cmds, skeleton only").
		AddArg("cmd-path", "", "path", "p", "P")

	list := mods.AddSub("list", "ls", "flatten", "flat", "f", "F", "~").
		RegCmd(DumpCmds,
			"list builtin and loaded cmds")
	addFindStrArgs(list)

	listSimple := list.AddSub("simple", "sim", "s", "S", "-").
		RegCmd(DumpCmdListSimple,
			"list builtin and loaded cmds in lite style")
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
			"save current cmds as a flow").
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
		RegCmd(DumpEnv,
			"list all env layers and KVs in tree format")

	// TODO: add search supporting
	env.AddSub("abbrs", "abbr", "a", "A").
		RegCmd(DumpEnvAbbrs,
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
		"add and pull a git address to hub").
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
}

func RegisterTrivialCmds(cmds *core.CmdTree) {
	dummy := cmds.AddSub("dummy", "dmy", "dm")

	dummy.RegCmd(Dummy,
		"dummy cmd for testing")

	dummy.AddSub("quiet", "q", "Q").
		RegCmd(QuietDummy,
			"quiet dummy cmd for testing").
		SetQuiet()

	dummy.AddSub("power", "p", "P").
		RegPowerCmd(PowerDummy,
			"power dummy cmd for testing")

	dummy.AddSub("priority", "prior", "prio", "pri").
		RegPowerCmd(PriorityPowerDummy,
			"power dummy cmd for testing").
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
	step.AddSub("on", "yes", "y", "Y", "1", "+").
		RegCmd(DbgStepOn,
			"enable step by step").
		SetQuiet()
	step.AddSub("off", "no", "n", "N", "0", "-").
		RegCmd(DbgStepOff,
			"disable step by step").
		SetQuiet()

	cmds.AddSub("delay-execute", "delay", "dl", "d", "D").
		RegCmd(DbgDelayExecute,
			"wait for a while before executing a cmd").
		SetQuiet().
		AddArg("seconds", "5", "second", "sec", "s", "S")

	cmds.AddSub("exec").
		RegCmd(DbgExecBash,
			"verify bash in os/exec")
}

func addFindStrArgs(cmd *core.Cmd) {
	cmd.AddArg("1st-str", "", "find-str").
		AddArg("2nd-str", "").
		AddArg("3rh-str", "").
		AddArg("4th-str", "").
		AddArg("5th-str", "").
		AddArg("6th-str", "")
}

const LessMoreHelpStr = `
* if in a sequence having
    * more than 1 other commands: show the sequence execution.
    * only 1 other command and
        * has no args and the other command is
            * a flow: show the flow execution.
            * not a flow: show the command or the branch info.
        * has args: find commands under the branch of the other command.
* if not in a sequence and
    * has args: do global search.
    * has no args: show global help.`

const GlobalHelpHelpStr = "display rich info base on:" + LessMoreHelpStr
const SkeletonHelpStr = "display brief info base on:" + LessMoreHelpStr

const MoveFlowsToDirHelpStr = `move all saved flows to a local dir (could be a git repo).
auto move:
    * if one(and only one) local(not linked to a repo) dir exists in hub
    * and the arg "path" is empty
    then flows will move to that dir`
