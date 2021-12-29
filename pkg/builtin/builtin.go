package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
)

// TODO: register the env kvs these commands will modify

func RegisterCmds(cmds *core.CmdTree) {
	RegisterCmdAndHelpCmds(cmds)
	RegisterTagsCmds(cmds)
	RegisterTailSubCmds(cmds)
	RegisterCmdsFindingCmds(cmds)
	RegisterCmdsListCmds(cmds)

	RegisterFlowDescCmds(cmds)

	RegisterHubManageCmds(cmds)
	RegisterEnvManageCmds(cmds)
	RegisterFlowManageCmds(cmds)
	RegisterBgManageCmds(cmds)

	RegisterSessionCmds(cmds)
	RegisterDbgCmds(cmds.AddSub(
		"dbg").RegEmptyCmd(
		"debug related commands").Owner())
	RegisterDisplayCmds(cmds.AddSub(
		"display", "disp").RegEmptyCmd(
		"display related commands").Owner())

	RegisterBuiltinCmds(cmds.AddSub(
		"builtin").RegEmptyCmd(
		"internal commands, mostly for init loading").Owner().SetHidden())

	RegisterTimerCmds(cmds)
	RegisterOsCmds(cmds)
	RegisterNoopCmds(cmds)
}

func RegisterCmdsFindingCmds(cmds *core.CmdTree) {
	find := cmds.AddSub("find").
		RegPowerCmd(GlobalFindCmds,
			"search commands by keywords").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority()
	addFindStrArgs(find)

	findShort := cmds.AddSub("/").
		RegPowerCmd(GlobalFindCmds,
			"shortcut of [find]").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority()
	addFindStrArgs(findShort)

	findUsage := find.AddSub("with-usage", "usage", "usg", "more", "m").
		RegPowerCmd(GlobalFindCmdsWithUsage,
			"search commands by find-strings, with usage info").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority()
	addFindStrArgs(findUsage)

	findUsageShort := cmds.AddSub("//").
		RegPowerCmd(GlobalFindCmdsWithUsage,
			"shortcut of [find.with-usage]").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority()
	addFindStrArgs(findUsageShort)

	findDetail := find.AddSub("with-full-info", "full", "f").
		RegPowerCmd(GlobalFindCmdsWithDetails,
			"search commands with strings, with full details").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority()
	addFindStrArgs(findDetail)

	findDetailShort := cmds.AddSub("///").
		RegPowerCmd(GlobalFindCmdsWithDetails,
			"shortcut of [find.with-full-info]").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority()
	addFindStrArgs(findDetailShort)
}

func RegisterTailSubCmds(cmds *core.CmdTree) {
	sub := cmds.AddSub("tail-sub").
		RegPowerCmd(DumpTailCmdSub,
			"display commands on the branch of the last command").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority()
	addFindStrArgs(sub)
	sub.AddArg("tag", "", "t")
	sub.AddArg("depth", "0", "d")

	subShort := cmds.AddSub("~").
		RegPowerCmd(DumpTailCmdSub,
			"shortcut of [tail-sub]").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority()
	addFindStrArgs(subShort)
	subShort.AddArg("tag", "", "t")
	subShort.AddArg("depth", "0", "d")

	subMore := sub.AddSub("with-usage", "usage", "usg", "more", "m").
		RegPowerCmd(DumpTailCmdSubWithUsage,
			"display commands on the branch of the last command, with usage info").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority()
	addFindStrArgs(subMore)
	subMore.AddArg("tag", "", "t")
	subMore.AddArg("depth", "0", "d")

	subMoreShort := cmds.AddSub("~~").
		RegPowerCmd(DumpTailCmdSubWithUsage,
			"shortcut of [tail-sub.with-usage]").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority()
	addFindStrArgs(subMoreShort)
	subMoreShort.AddArg("tag", "", "t")
	subMoreShort.AddArg("depth", "0", "d")

	subFull := sub.AddSub("with-full-info", "full", "f").
		RegPowerCmd(DumpTailCmdSubWithDetails,
			"display commands on the branch of the last command, with full details").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority()
	addFindStrArgs(subFull)
	subFull.AddArg("tag", "", "t")
	subFull.AddArg("depth", "0", "d")

	subFullShort := cmds.AddSub("~~~").
		RegPowerCmd(DumpTailCmdSubWithDetails,
			"shortcut of [tail-sub.with-full-info]").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority()
	addFindStrArgs(subFullShort)
	subFullShort.AddArg("tag", "", "t")
	subFullShort.AddArg("depth", "0", "d")
}

func RegisterCmdsListCmds(cmds *core.CmdTree) {
	addDumpCmdsArgs := func(cmd *core.Cmd) {
		cmd.SetPriority().
			SetQuiet().
			SetAllowTailModeCall().
			AddArg("cmd-path", "", "path", "p").
			AddArg("source", "", "repo", "src", "s").
			AddArg("tag", "", "t").
			AddArg("depth", "0", "d")
		addFindStrArgs(cmd)
	}

	list := cmds.AddSub("cmds").
		RegPowerCmd(DumpCmds,
			"list commands by command-path-branch, keywords, source and tag")
	addDumpCmdsArgs(list)

	listMore := list.AddSub("with-usage", "usage", "usg", "more", "m").
		RegPowerCmd(DumpCmdsWithUsage,
			"list commands by command-path-branch, keywords, source and tag, with usage info")
	addDumpCmdsArgs(listMore)

	listFull := list.AddSub("with-full-info", "full", "f").
		RegPowerCmd(DumpCmdsWithDetails,
			"list commands by command-path-branch, keywords, source and tag, with full info")
	addDumpCmdsArgs(listFull)

	list.AddSub("tree", "t").
		RegPowerCmd(DumpCmdsTree,
			"list commands in tree form by command-path-branch").
		SetAllowTailModeCall().
		AddArg("cmd-path", "", "path", "p").
		AddArg("depth", "1", "d")

	/*
		/*
		list := mods.AddSub("list", "ls").
			RegPowerCmd(DumpCmdList,
				"list builtin and loaded commands").
			SetAllowTailModeCall()
		addFindStrArgs(list)

		listSimple := list.AddSub("simple", "sim", "s").
			RegPowerCmd(DumpCmdListSimple,
				"list builtin and loaded commands in lite style").
			SetAllowTailModeCall()
		addFindStrArgs(listSimple)

		tree := mods.AddSub("tree", "t")
		tree.RegPowerCmd(DumpCmdTree,
			"list builtin and loaded commands").
			SetAllowTailModeCall().
			AddArg("cmd-path", "", "path", "p")
		tree.AddSub("simple", "sim", "skeleton", "s").
			RegPowerCmd(DumpCmdTreeSkeleton,
				"list builtin and loaded commands, skeleton only").
			SetAllowTailModeCall().
			AddArg("cmd-path", "", "path", "p").
			AddArg("recursive", "true", "r")
	*/
}

func RegisterTagsCmds(cmds *core.CmdTree) {
	cmds.AddSub("tags").
		RegPowerCmd(ListTags,
			"list all tags")

	findByTag := cmds.AddSub("tag", cmds.Strs.TagMark).
		RegPowerCmd(FindByTags,
			"list commands having the specified tags").
		SetAllowTailModeCall()
	addFindStrArgs(findByTag)

	byTagUsage := findByTag.AddSub("with-usage", "usage", "more", "m").
		RegPowerCmd(FindByTagsWithUsage,
			"list commands having the specified tags, with usage info").
		SetAllowTailModeCall()
	addFindStrArgs(byTagUsage)

	byTagDetails := findByTag.AddSub("with-full-info", "full", "f").
		RegPowerCmd(FindByTagsWithDetails,
			"list commands having the specified tags, with full details").
		SetAllowTailModeCall()
	addFindStrArgs(byTagDetails)
}

func RegisterCmdAndHelpCmds(cmds *core.CmdTree) {
	help := cmds.AddSub("help")
	help.RegPowerCmd(GlobalHelp,
		"get help").
		AddArg("target", "")

	help.AddSub(cmds.Strs.SelfName, "self").
		RegPowerCmd(SelfHelp,
			"get "+cmds.Strs.SelfName+" usage help")

	cmd := cmds.AddSub("cmd").
		RegPowerCmd(DumpCmdUsage,
			"display command usage").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority().
		AddArg("cmd-path", "", "path", "p")

	cmd.AddSub("full", "f", "more", "m").
		RegPowerCmd(DumpCmdWithDetails,
			"display command full info").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority().
		AddArg("cmd-path", "", "path", "p")

	cmds.AddSub("=").
		RegPowerCmd(DumpTailCmdWithUsage,
			"shortcut of [cmd], if at the end of a flow:\n   * display usage of the last command\n   * flow will not execute").
		SetQuiet().
		SetPriority().
		AddArg("cmd-path", "", "path", "p")

	cmds.AddSub("==", "===").
		RegPowerCmd(DumpTailCmdWithDetails,
			"shortcut of [cmd], if at the end of a flow:\n   * display full info of the last command\n   * flow will not execute").
		SetQuiet().
		SetPriority().
		AddArg("cmd-path", "", "path", "p")
}

func RegisterFlowDescCmds(cmds *core.CmdTree) {
	desc := cmds.AddSub("desc").
		RegPowerCmd(DumpFlowSkeleton,
			"desc the flow execution in skeleton style").
		SetQuiet().
		SetPriority().
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
	cmds.AddSub("-").
		RegPowerCmd(DumpFlowSkeleton,
			"shortcut of [desc]").
		SetQuiet().
		SetPriority().
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
	cmds.AddSub("--").
		RegPowerCmd(DumpFlowSkeleton,
			"shortcut of [desc], unfold more(2L) trivial subflows").
		SetQuiet().
		SetPriority().
		AddArg("unfold-trivial", "2", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
	cmds.AddSub("---").
		RegPowerCmd(DumpFlowSkeleton,
			"shortcut of [desc], unfold more(3L) trivial subflows").
		SetQuiet().
		SetPriority().
		AddArg("unfold-trivial", "3", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")

	desc.AddSub("more", "m").
		RegPowerCmd(DumpFlowAllSimple,
			"desc the flow about to execute in lite style").
		SetQuiet().
		SetPriority().
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
	cmds.AddSub("+").
		RegPowerCmd(DumpFlowAllSimple,
			"shortcut of [desc.more]").
		SetQuiet().
		SetPriority().
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
	cmds.AddSub("++").
		RegPowerCmd(DumpFlowAllSimple,
			"shortcut of [desc.more], unfold more(2L) trivial subflows").
		SetQuiet().
		SetPriority().
		AddArg("unfold-trivial", "2", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
	cmds.AddSub("+++").
		RegPowerCmd(DumpFlowAllSimple,
			"shortcut of [desc.more], unfold more(3L) trivial subflows").
		SetQuiet().
		SetPriority().
		AddArg("unfold-trivial", "3", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")

	desc.AddSub("full", "f", "F").
		RegPowerCmd(DumpFlowAll,
			"desc the flow about to execute with details").
		SetQuiet().
		SetPriority().
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")

	desc.AddSub("dependencies", "depends", "depend", "dep", "os-cmd", "os").
		RegPowerCmd(DumpFlowDepends,
			"list the depended os-commands of the flow").
		SetQuiet().
		SetPriority()
	desc.AddSub("env-ops-check", "env-ops", "env-op", "env", "ops", "op", "e").
		RegPowerCmd(DumpFlowEnvOpsCheckResult,
			"desc the env-ops check result of the flow").
		SetQuiet().
		SetPriority()

	descFlow := desc.AddSub("flow").
		RegPowerCmd(DumpFlowSkeleton,
			"desc the flow execution in skeleton style").
		SetQuiet().
		SetPriority().
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
	descFlow.AddSub("more", "m").
		RegPowerCmd(DumpFlowSimple,
			"desc the flow execution in lite style").
		SetQuiet().
		SetPriority().
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
	descFlow.AddSub("full", "f").
		RegPowerCmd(DumpFlow,
			"desc the flow execution with details").
		SetQuiet().
		SetPriority().
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
}

func RegisterHubManageCmds(cmds *core.CmdTree) {
	hub := cmds.AddSub("hub", "h").
		RegPowerCmd(ListHub,
			"list dir and repo info in hub").
		SetAllowTailModeCall()
	addFindStrArgs(hub)

	hub.AddSub("clear", "clean").
		RegPowerCmd(RemoveAllFromHub,
			"remove all repos from hub")

	hub.AddSub("init").
		RegPowerCmd(AddDefaultGitRepoToHub,
			"add and pull basic hub-repo to local").
		AddArg("show-tip", "true", "tip")

	add := hub.AddSub("add-and-update", "add", "a")
	add.RegPowerCmd(AddGitRepoToHub,
		"add and pull a git address to hub, do update if it already exists").
		SetAllowTailModeCall().
		AddArg("git-address", "", "git", "address", "addr")

	repoStatus := hub.AddSub("git-status", "status").
		RegPowerCmd(CheckGitRepoStatus,
			"check git status for all repos")
	addFindStrArgs(repoStatus)

	add.AddSub("local-dir", "local", "l").
		RegPowerCmd(AddLocalDirToHub,
			"add a local dir (could be a git repo) to hub").
		SetAllowTailModeCall().
		AddArg("path", "", "p")

	purge := hub.AddSub("purge", "p")
	purge.RegPowerCmd(PurgeInactiveRepoFromHub,
		"remove an inactive repo from hub").
		SetAllowTailModeCall().
		AddArg("find-str", "", "s", "S")
	purge.AddSub("purge-all-inactive", "all", "inactive", "a").
		RegPowerCmd(PurgeAllInactiveReposFromHub,
			"remove all inactive repos from hub")

	hub.AddSub("update-all", "update", "u").
		RegPowerCmd(UpdateHub,
			"update all repos and mods defined in hub").
		AddArg("show-tip", "true", "tip")

	hub.AddSub("enable-repo", "enable", "en", "e").
		RegPowerCmd(EnableRepoInHub,
			"enable matched git repos in hub").
		SetAllowTailModeCall().
		AddArg("find-str", "", "s")

	hub.AddSub("disable-repo", "disable", "dis", "d").
		RegPowerCmd(DisableRepoInHub,
			"disable matched git repos in hub").
		SetAllowTailModeCall().
		AddArg("find-str", "", "s")

	hub.AddSub("move-flows-to-dir", "mv", "m").
		RegPowerCmd(MoveSavedFlowsToLocalDir,
			moveFlowsToDirHelpStr).
		SetAllowTailModeCall().
		AddArg("path", "", "p")
}

func RegisterEnvManageCmds(cmds *core.CmdTree) {
	env := cmds.AddSub("env", "e").
		RegPowerCmd(DumpEssentialEnvFlattenVals,
			"list essential env values in flatten format").
		SetAllowTailModeCall()
	addFindStrArgs(env)

	env.AddSub("tree", "t").
		RegPowerCmd(DumpEnvTree,
			"list all env layers and values in tree format")

	// TODO: add search supporting
	abbrs := env.AddSub("abbrs", "abbr", "a")
	abbrs.RegPowerCmd(DumpEnvAbbrs,
		"list env tree and abbrs")

	envList := env.AddSub("list", "ls").
		RegPowerCmd(DumpEnvFlattenVals,
			"list all env values in flatten format").
		SetAllowTailModeCall()
	addFindStrArgs(envList)

	env.AddSub("save", "s", "S").
		RegPowerCmd(SaveEnvToLocal,
			"save session env changes to local").
		SetQuiet()

	env.AddSub("remove-and-save", "remove", "rm", "delete", "del").
		RegPowerCmd(RemoveEnvValAndSaveToLocal,
			"remove specified env value and save changes to local").
		SetAllowTailModeCall().
		AddArg("key", "", "k", "K")

	// '--' is for compatible only, will remove late
	env.AddSub("reset-session", "reset", "--").
		RegPowerCmd(ResetSessionEnv,
			"clear all env values in current session")

	env.AddSub("clear-and-save", "clean-and-save", "clear", "clean").
		RegPowerCmd(ResetLocalEnv,
			"clear all local saved env values")

	env.AddSub("who-write", "ww").
		RegPowerCmd(DumpCmdsWhoWriteKey,
			"find which commands provide the value of the specified key").
		SetAllowTailModeCall().
		AddArg("key", "", "k")

	assert := env.AddSub("assert")
	assert.AddSub("equal").
		RegPowerCmd(EnvAssertEqual,
			"assert the value of a key in env equal to specified value").
		AddArg("key", "", "k").
		AddArg("value", "", "val", "v")
	assert.AddSub("not-exists").
		RegPowerCmd(EnvAssertNotExists,
			"assert the key not in env").
		AddArg("key", "", "k")

	env.AddSub("map").
		RegPowerCmd(MapEnvKeyValueToAnotherKey,
			"read src-key's value and write to dest-key").
		AddArg("src-key", "", "source-key", "source", "src", "from").
		AddArg("dest-key", "", "dest", "to").
		AddEnvOp("[[dest-key]]", core.EnvOpTypeWrite)

	env.AddSub("set").
		RegPowerCmd(SetEnvKeyValue,
			"set key-value to env in current session").
		AddArg("key", "", "k").
		AddArg("value", "", "val", "v")

	env.AddSub("add").
		RegPowerCmd(AddEnvKeyValue,
			"add key-value to env in current session, throws error if old value exists").
		AddArg("key", "", "k").
		AddArg("value", "", "val", "v")

	env.AddSub("update").
		RegPowerCmd(UpdateEnvKeyValue,
			"update key-value in env, throws error if old value not exists").
		AddArg("key", "", "k").
		AddArg("value", "", "val", "v")

	registerSimpleSwitch(abbrs,
		"borrowing commands' abbrs when setting env key-values",
		"sys.env.use-cmd-abbrs",
		"cmd")
}

func RegisterFlowManageCmds(cmds *core.CmdTree) {
	flow := cmds.AddSub("flow", "f").
		RegPowerCmd(ListFlows,
			"list local saved but unlinked (to any repo) flows").
		SetAllowTailModeCall()
	addFindStrArgs(flow)

	// TODO: check the new cmd is conflicted with other cmds
	flow.AddSub("save", "s").
		RegPowerCmd(SaveFlow,
			"save current commands as a flow").
		SetQuiet().
		SetPriority().
		AddArg("to-cmd-path", "", "path", "p")

	flow.AddSub("set-help-str", "help", "h").
		RegPowerCmd(SetFlowHelpStr,
			"set help str to a saved flow").
		SetQuiet().
		AddArg("cmd-path", "", "path", "p").
		AddArg("help-str", "", "help", "h")

	// TODO: check the new cmd is conflicted with other cmds
	flow.AddSub("rename").
		RegPowerCmd(RenameFlow,
			"rename a saved flow").
		AddArg("src", "", "src-cmd", "s").
		AddArg("dest", "", "dest-cmd", "d")

	flow.AddSub("remove", "rm", "delete", "del").
		RegPowerCmd(RemoveFlow,
			"remove a saved flow").
		SetAllowTailModeCall().
		AddArg("cmd-path", "", "path", "p")

	flow.AddSub("load", "l").
		RegPowerCmd(LoadFlowsFromDir,
			"load flows from local dir").
		SetAllowTailModeCall().
		AddArg("path", "", "p")

	flow.AddSub("clear", "clean").
		RegPowerCmd(RemoveAllFlows,
			"remove all flows saved in local")

	flow.AddSub("move-flows-to-dir", "move", "mv", "m").
		RegPowerCmd(MoveSavedFlowsToLocalDir,
			moveFlowsToDirHelpStr).
		SetAllowTailModeCall().
		AddArg("path", "", "p")
}

func RegisterBgManageCmds(cmds *core.CmdTree) {
	bg := cmds.AddSub("background", "bg").
		RegEmptyCmd(
			"background tasks management")
	wait := bg.AddSub("wait").
		RegPowerCmd(WaitForBgTaskFinish,
			"wait for a task/thread to finish, if no specify thread id, wait for all").
		AddArg("thread", "", "id")
	wait.AddSub("all").
		RegPowerCmd(WaitForAllBgTasksFinish,
			"wait for all tasks/threads to finish in current(must be main) thread")
}

func RegisterSessionCmds(cmds *core.CmdTree) {
	sessions := cmds.AddSub("sessions", "session", "s")
	sessionsCmd := sessions.RegPowerCmd(ListSessions,
		"list or find executed and executing sessions").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority()
	addFindStrArgs(sessionsCmd)
	sessionsCmd.AddArg("session-id", "", "session", "id")

	sessions.AddSub("retry", "r").
		RegAdHotFlowCmd(SessionRetry,
			//"find a session by keywords and id, if it's failed, retry running from the error point").
			"find a session by id, if it's failed, retry running from the error point").
		AddArg("session-id", "", "session", "id")

	desc := sessions.AddSub("desc", "d", "-").
		RegPowerCmd(ListedSessionDescLess,
			"desc executed/ing session").
		SetAllowTailModeCall()
	addFindStrArgs(desc)
	desc.AddArg("unfold-trivial", "1", "ut", "unfold", "unf", "uf", "u", "U", "trivial", "triv", "tri", "t", "T")
	desc.AddArg("depth", "32", "d", "D")
	desc.AddArg("session-id", "", "session", "id")

	descMore := desc.AddSub("more", "m").
		RegPowerCmd(ListedSessionDescMore,
			"desc executed/ing session with more details").
		SetAllowTailModeCall()
	addFindStrArgs(descMore)
	descMore.AddArg("unfold-trivial", "1", "ut", "unfold", "unf", "uf", "u", "U", "trivial", "triv", "tri", "t", "T")
	descMore.AddArg("depth", "32", "d")
	descMore.AddArg("session-id", "", "session", "id")

	descFull := desc.AddSub("full", "f").
		RegPowerCmd(ListedSessionDescFull,
			"desc executed/ing session with full details").
		SetAllowTailModeCall()
	addFindStrArgs(descFull)
	descFull.AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t")
	descFull.AddArg("depth", "32", "d")
	descFull.AddArg("session-id", "", "session", "id")

	last := sessions.AddSub("last", "l")
	last.RegPowerCmd(LastSession,
		"show last session")

	lastDesc := last.AddSub("desc", "d", "-")
	lastDesc.RegPowerCmd(LastSessionDescLess,
		"desc the execution status of last session").
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")

	lastDesc.AddSub("more", "m").
		RegPowerCmd(LastSessionDescMore,
			"desc the execution status of last session with more details").
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")

	lastDesc.AddSub("full", "f").
		RegPowerCmd(LastSessionDescFull,
			"desc the execution status of last session with full details").
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")

	/*
		last.AddSub("retry", "r", "R").
			RegAdHotFlowCmd(LastSessionRetry,
				"if the last session failed, retry running from the error point")
	*/

	remove := sessions.AddSub("remove", "delete", "rm").
		RegPowerCmd(FindAndRemoveSessions,
			"clear executed sessions by keywords").
		SetAllowTailModeCall().
		SetQuiet().
		SetPriority()
	addFindStrArgs(remove)
	remove.AddArg("session-id", "", "session", "id")
	remove.AddArg("remove-running", "false", "force")

	remove.AddSub("all", "a").
		RegPowerCmd(RemoveAllSessions,
			"clear all executed sessions")
	sessions.AddSub("clear", "clean").
		RegPowerCmd(RemoveAllSessions,
			"clear all executed sessions")

	sessions.AddSub("set-keep-duration", "keep-duration", "keep", "kd").
		RegPowerCmd(SetSessionsKeepDur,
			"set the keeping duration of executed sessions").
		AddArg("duration", "72h", "dur")
}

func RegisterDbgCmds(cmds *core.CmdTree) {
	registerSimpleSwitch(cmds,
		"step by step on executing",
		"sys.step-by-step",
		"step-by-step", "step", "confirm", "cfm")

	registerSimpleSwitch(cmds,
		"recover from internal error and give a frendly message",
		"sys.panic.recover",
		"recover")

	delay := cmds.AddSub("delay-execute", "delay").
		RegPowerCmd(DbgDelayExecute,
			"wait for specified duration before executing each commands").
		SetQuiet().
		AddArg("seconds", "3", "second", "sec", "s", "S")

	delay.AddSub("at-end", "at-finish", "post-execute", "end", "finish").
		RegPowerCmd(DbgDelayExecuteAtEnd,
			"wait for specified duration after executing each commands").
		SetQuiet().
		AddArg("seconds", "3", "second", "sec", "s", "S")

	breaks := cmds.AddSub("breaks", "break")
	breaks.RegPowerCmd(DbgBreakAtBegin,
		"set break point at the first command").
		SetQuiet()

	breaksAt := breaks.AddSub("at", "before").
		RegPowerCmd(DbgBreakBefore,
			// TODO: get 'sep' from env or other config
			"setup before-command break points, use ',' to seperate multipy commands").
		SetQuiet().
		AddArg("break-points", "", "break-point", "cmds", "cmd")

	breaksAt.AddSub("begin", "start", "first", "0").
		RegPowerCmd(DbgBreakAtBegin,
			"set break point at the first command").
		SetQuiet()

	breaksAt.AddSub("end", "finish").
		RegPowerCmd(DbgBreakAtEnd,
			"set break point after the last command").
		SetQuiet()

	breaks.AddSub("after", "post").
		RegPowerCmd(DbgBreakAfter,
			// TODO: get 'sep' from env or other config
			"setup after-command break points, use ',' to seperate multipy commands").
		SetQuiet().
		AddArg("break-points", "", "break-point", "cmds", "cmd")

	panicTest := cmds.AddSub("panic")
	panicTest.RegPowerCmd(DbgPanic,
		"for panic test").
		AddArg("random-arg-1", "arg-1").
		AddArg("random-arg-2", "arg-2")

	panicTest.AddSub("cmd").
		RegPowerCmd(DbgPanicCmdError,
			"for specified-error-type panic test").
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

	interact := cmds.AddSub("interact", "interactive", "i").
		RegPowerCmd(DbgInteract,
			"enter interact mode").
		SetQuiet()

	interact.AddSub("leave", "l").
		RegPowerCmd(DbgInteractLeave,
			"leave interact mode and continue to run")

	// For compatible only, will remove late
	cmds.AddSub("echo").
		RegPowerCmd(DbgEcho,
			"! moved to 'echo', this is just for compatible only, will be removed soon").
		AddArg("message", "", "msg", "m", "M")
}

func RegisterDisplayCmds(cmds *core.CmdTree) {
	cmds.AddSub("style").
		RegPowerCmd(SetDisplayStyle,
			"set executing display style: bold, slash, corner, ascii, utf8(default)").
		AddArg("style", "s").
		SetQuiet()

	completion := cmds.AddSub("completion")
	registerSimpleSwitch(completion,
		"hidden style completion",
		"display.completion.hidden",
		"hidden")

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
		"display", "disp")

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
		"default")

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
		"realname")

	registerSimpleSwitch(cmds,
		"executor display",
		"display.executor",
		"executor")

	registerSimpleSwitch(cmds,
		"executor finish footer display",
		"display.executor.end",
		"end", "footer")

	registerSimpleSwitch(cmds,
		"bootstrap executing display",
		"display.bootstrap",
		"bootstrap", "boot")

	registerSimpleSwitch(cmds,
		"display executor when only executing one command",
		"display.one-cmd",
		"one-cmd")

	cmds.AddSub("set-width", "width", "wid", "w").
		RegPowerCmd(SetDisplayWidth,
			"set display width, if arg 'width' is empty or 0, set width to screen size").
		SetQuiet().
		AddArg("width", "", "wid", "w")

	RegisterVerbCmds(cmds)
}

func RegisterVerbCmds(cmds *core.CmdTree) {
	cmds.AddSub("quiet", "q").
		RegPowerCmd(SetQuietMode,
			"change into quiet mode").
		SetQuiet()

	verbose := cmds.AddSub("verbose", "verb", "v")

	verbose.RegPowerCmd(SetVerbMode,
		"change into verbose mode").
		SetQuiet()

	verbose.AddSub("default", "def").
		RegPowerCmd(SetToDefaultVerb,
			"set to default verbose mode").
		SetQuiet()

	verbose.AddSub("increase", "inc").
		RegPowerCmd(IncreaseVerb,
			"increase verbose").
		SetQuiet().
		AddArg("volume", "1", "vol", "v")

	verbose.AddSub("decrease", "dec").
		RegPowerCmd(DecreaseVerb,
			"decrease verbose").
		SetQuiet().
		AddArg("volume", "1", "vol", "v")
}

func RegisterBuiltinCmds(cmds *core.CmdTree) {
	env := cmds.AddSub("env", "e")

	envLoad := env.AddSub("load", "l")

	envLoad.AddSub("local", "l").
		RegPowerCmd(LoadLocalEnv,
			"load env values from local")

	envLoad.AddSub("runtime", "rt", "r").
		RegPowerCmd(LoadRuntimeEnv,
			"setup runtime env values")

	mod := cmds.AddSub("mod", "mods", "m")

	modLoad := mod.AddSub("load", "l")

	modLoad.AddSub("flows", "flows", "f").
		RegPowerCmd(LoadFlows,
			"load saved flows from local")

	modLoad.AddSub("ext-executor", "ext-exec", "ext").
		RegPowerCmd(SetExtExec,
			"load default setting of how to run a executable file by ext name")

	modLoad.AddSub("hub", "h").
		RegPowerCmd(LoadModsFromHub,
			"load flows and mods from local hub")

	cmds.AddSub("display", "disp", "d").
		AddSub("load", "l").
		AddSub("platform", "p").
		RegPowerCmd(LoadPlatformDisplay,
			"load platform(OS) specialized display settings")
}

func RegisterTimerCmds(cmds *core.CmdTree) {
	cmds.AddSub("mark-time").
		RegPowerCmd(MarkTime,
			"set current timestamp to the specified key").
		AddArg("write-to-key", "", "key", "k").
		AddEnvOp("[[write-to-key]]", core.EnvOpTypeWrite)

	timer := cmds.AddSub("timer")
	timer.AddSub("begin").
		RegPowerCmd(TimerBegin,
			"start timer, set current timestamp to the specified key").
		AddArg("begin-key", "", "key", "k").
		AddEnvOp("[[begin-key]]", core.EnvOpTypeWrite)

	timer.AddSub("elapsed", "elapse", "end").
		RegPowerCmd(TimerElapsed,
			"set elapsed seconds to the specified key").
		AddArg("begin-key", "", "begin").
		AddArg("write-to-key", "", "key", "k").
		AddEnvOp("[[begin-key]]", core.EnvOpTypeRead).
		AddEnvOp("[[write-to-key]]", core.EnvOpTypeWrite)
}

func RegisterOsCmds(cmds *core.CmdTree) {
	cmds.AddSub("bash").
		RegPowerCmd(ExecCmds,
			"execute os command in bash").
		AddArg("command", "", "cmd")

	cmds.AddSub("sleep", "slp").
		RegPowerCmd(Sleep,
			"sleep for specified duration").
		AddArg("duration", "1s", "dur", "d")

	cmds.AddSub("echo").
		RegPowerCmd(DbgEcho,
			"print message from argv").
		AddArg("message", "", "msg", "m")
}

func RegisterNoopCmds(cmds *core.CmdTree) {
	cmds.AddSub("noop").
		RegPowerCmd(Noop,
			"do exactly nothing")

	cmds.AddSub("dummy", "dm").
		RegPowerCmd(Dummy,
			"dummy command for testing")
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
	on := cmd.AddSub("on").RegEmptyCmd("enable " + function).SetQuiet()
	off := cmd.AddSub("off").RegEmptyCmd("disable " + function).SetQuiet()

	for _, key := range keys {
		self.AddVal2Env(key, "true")
		on.AddVal2Env(key, "true")
		off.AddVal2Env(key, "false")
	}

	return self.Owner()
}

const moveFlowsToDirHelpStr = `move all saved flows to a local dir (could be a git repo).
will do auto move if:
    * if one(and only one) local(not linked to a repo) dir exists in hub
    * and the arg "path" is empty
    then flows will move to that dir`
