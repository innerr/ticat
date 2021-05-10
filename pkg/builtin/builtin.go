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
	mod := cmds.AddSub("cmds", "cmd", "mod", "mods", "m", "M", "c", "C")
	mod.AddSub("tree", "t", "T").
		RegCmd(DumpCmdTree,
			"list builtin and loaded cmds").
		AddArg("path", "", "p", "P")
	mod.AddSub("list", "ls", "l", "flatten", "flat", "f", "F").
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
	env.AddSub("reset-and-save", "reset").
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
	add := hub.AddSub("add-and-update", "add", "a", "A")
	add.RegCmd(AddToLocalHub,
		"add and pull a git address to local hub").
		AddArg("git-address", "", "git", "address", "addr")
	add.AddSub("basic", "base", "default", "b", "B", "d", "D").
		RegCmd(AddDefaultToLocalHub,
			"add and pull basic hub-repo to local").
		AddArg("git-address", "", "git", "address", "addr")
	add.AddSub("local-dir", "local", "l", "L").
		RegCmd(AddLocalDirToLocalHub,
			"add a local dir (could be a git repo) to local hub").
		AddArg("path", "", "p", "P")
	hub.AddSub("list", "ls", "l", "L").
		RegCmd(ListLocalHub,
			"list local hub")
	hub.AddSub("update", "u", "U").
		RegCmd(UpdateLocalHub,
			"update mods defined in local hub")
	hub.AddSub("enable-git-address", "enable", "e", "E").
		RegCmd(EnableAddrInLocalHub,
			"enable a git repo address in local hub")
	hub.AddSub("disable-git-address", "disable", "d", "D").
		RegCmd(DisableAddrInLocalHub,
			"disable a git repo address in local hub")
	hub.AddSub("move-to-dir", "move", "m", "M").
		RegCmd(MoveSavedFlowsToLocalDir,
			"move all saved flows to a local dir (could be a git repo)")
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
	modLoad.AddSub("local", "l", "L").
		RegCmd(LoadLocalMods,
			"load mods from local").
		SetQuiet()
	modLoad.AddSub("flows", "flows", "f", "F").
		RegCmd(LoadLocalFlows,
			"load flows from local")
	modLoad.AddSub("ext-exec", "ext", "e", "E").
		RegCmd(SetExtExec,
			"load default setting of how to run a executable file by ext name")
	/*
		modLoad.AddSub("hub", "h", "H").
			RegCmd(LoadFromLocalHub,
				"load flows and mods from local hub")
	*/
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
	cmds.AddSub("tty-read", "tty").
		RegCmd(DbgReadFromTty,
			"verify stdin and tty could work together")
}
