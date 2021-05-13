package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

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
	cmds.AddSub("help", "?").
		RegPowerCmd(GlobalHelp,
			"get help").
		SetPriority().
		AddArg("1st-str", "", "1", "find", "str", "s", "S").
		AddArg("2rd-str", "", "2").
		AddArg("3th-str", "", "3")
	cmds.AddSub("search", "find", "fnd", "s", "S").
		RegCmd(FindAny,
			"find anything with given string").
		AddArg("1st-str", "", "1", "find", "str", "s", "S").
		AddArg("2rd-str", "", "2").
		AddArg("3th-str", "", "3")
	cmds.AddSub("desc", "d", "D").
		RegPowerCmd(DumpFlow,
			"desc the flow about to execute").
		SetQuiet().
		SetPriority()
	mod := cmds.AddSub("cmds", "cmd", "mod", "mods", "m", "M")
	mod.AddSub("tree", "t", "T").
		RegCmd(DumpCmdTree,
			"list builtin and loaded cmds").
		AddArg("path", "", "p", "P")
	mod.AddSub("list", "ls", "flatten", "flat", "f", "F").
		RegCmd(DumpCmds,
			"list builtin and loaded cmds").
		AddArg("1st-str", "", "1", "find", "str", "s", "S").
		AddArg("2rd-str", "", "2").
		AddArg("3th-str", "", "3")
}

func RegisterFlowCmds(cmds *core.CmdTree) {
	flow := cmds.AddSub("flow", "fl", "f", "F")
	flow.AddSub("save", "persist", "s", "S").
		RegPowerCmd(SaveFlow,
			"save current cmds as a flow").
		SetQuiet().
		SetPriority().
		AddArg("to-cmd-path", "", "path", "p", "P")
	flow.AddSub("remove", "rm", "delete", "del").
		RegCmd(RemoveFlow,
			"remove a saved flow").
		AddArg("cmd-path", "", "path", "p", "P")
	flow.AddSub("list-local", "list", "ls").
		RegCmd(ListFlows,
			"list local saved but unsynced (to any repo) flows")
	flow.AddSub("load", "l", "L").
		RegCmd(LoadFlowsFromDir,
			"load flows from local dir").
		AddArg("path", "", "p", "P")
	flow.AddSub("clear").
		RegCmd(RemoveAllFlows,
			"remove all flows saved in local")
}

func RegisterEnvCmds(cmds *core.CmdTree) {
	env := cmds.AddSub("env", "e", "E")
	env.AddSub("tree", "t", "T").
		RegCmd(DumpEnv,
			"list all env layers and KVs in tree format")
	// TODO: add search supporting
	env.AddSub("abbrs", "abbr", "a", "A").
		RegCmd(DumpEnvAbbrs,
			"list env tree and abbrs")
	env.AddSub("list", "ls", "flatten", "flat", "f", "F").
		RegCmd(DumpEnvFlattenVals,
			"list env values in flatten format").
		AddArg("1st-str", "", "1", "find", "str", "s", "S").
		AddArg("2rd-str", "", "2").
		AddArg("3th-str", "", "3")

	env.AddSub("save", "persist", "s", "S").
		RegCmd(SaveEnvToLocal,
			"save session env changes to local").
		SetQuiet()
	env.AddSub("remove-and-save", "remove", "rm", "delete", "del").
		RegCmd(RemoveEnvValAndSaveToLocal,
			"remove specific env KV and save changes to local").
		AddArg("key", "", "k", "K")
	env.AddSub("reset-and-save", "reset", "clear").
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
	hub := cmds.AddSub("hub", "h", "H")
	hub.AddSub("clear").
		RegCmd(RemoveAllFromHub,
			"remove all repos from hub")
	hub.AddSub("init").
		RegCmd(AddGitDefaultToHub,
			"add and pull basic hub-repo to local")
	add := hub.AddSub("add-and-update", "add", "a", "A")
	add.RegCmd(AddGitRepoToHub,
		"add and pull a git address to hub").
		AddArg("git-address", "", "git", "address", "addr")
	add.AddSub("local-dir", "local", "l", "L").
		RegCmd(AddLocalDirToHub,
			"add a local dir (could be a git repo) to hub").
		AddArg("path", "", "p", "P")
	hub.AddSub("list", "ls").
		RegCmd(ListHub,
			"list dir and repo info in hub")
	purge := hub.AddSub("purge", "p", "P")
	purge.RegCmd(PurgeInactiveRepoFromHub,
		"remove an inactive repo from hub").
		AddArg("find-str", "", "s", "S")
	purge.AddSub("purge-all-inactive", "all", "inactive", "a", "A").
		RegCmd(PurgeAllInactiveReposFromHub,
			"remove all inactive repos from hub")
	hub.AddSub("update", "u", "U").
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
			"move all saved flows to a local dir (could be a git repo)").
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
	envLoad.AddSub("stdin", "s", "S").
		RegCmd(LoadStdinEnv,
			"load env KVs from stdin").
		SetQuiet()
	envLoad.AddSub("abbrs", "abbr", "a", "A").
		RegCmd(LoadEnvAbbrs,
			"setup runtime env abbrs").
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

func RegisterDbgCmds(cmds *core.CmdTree) {
	// This cmds are just for debug
	cmds.AddSub("echo").
		RegCmd(DbgEcho,
			"print message from argv").
		AddArg("messsage", "", "msg", "m", "M")
	return
	cmds.AddSub("tty-read", "tty").
		RegCmd(DbgReadFromTty,
			"verify stdin and tty could work together")
}
