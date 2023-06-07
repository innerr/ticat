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

	env := RegisterEnvManageCmds(cmds)
	RegisterEnvSnapshotManageCmds(env)

	RegisterFlowManageCmds(cmds)
	RegisterBgManageCmds(cmds)

	RegisterSessionCmds(cmds)

	//RegisterBlenderCmds(cmds.AddSub(
	//	"blender", "blend").RegEmptyCmd(
	//	"a toolbox to modify flows during running").Owner())
	RegisterBlenderCmds(cmds.GetOrAddSub("flow"))

	RegisterDbgCmds(cmds.AddSub(
		"dbg").RegEmptyCmd(
		"debug related commands").Owner())

	RegisterDisplayCmds(cmds.AddSub(
		"display", "disp").RegEmptyCmd(
		"display related commands").Owner())

	RegisterBuiltinCmds(cmds.AddSub(
		"builtin").RegEmptyCmd(
		"internal commands, mostly for init loading").Owner().SetHidden())

	RegisterCtrlCmds(cmds)
	RegisterTimerCmds(cmds)
	RegisterOsCmds(cmds)
	RegisterJoinCmds(cmds)
	RegisterNoopCmds(cmds)
	RegisterHookCmds(cmds)

	cmds.AddSub("selftest").
		RegAdHotFlowCmd(Selftest,
			"run all commands having selftest tag, run in forest-mode to keep env clean for each test").
		AddArg("match-source", "", "match-src", "match", "src").
		AddArg("filter-source", "", "filter").
		AddArg("tag", "selftest").
		AddArg("parallel", "false", "parall", "paral", "para")

	cmds.AddSub("repeat", "rpt").
		RegAdHotFlowCmd(Repeat,
			"run a command many times").
		AddArg("cmd", "").
		AddArg("times", "1", "t")

	api := cmds.AddSub("api")
	api.RegEmptyCmd("api toolbox")
	RegisterApiCmds(api)
}

func RegisterCmdsFindingCmds(cmds *core.CmdTree) {
	find := cmds.AddSub("find").
		RegPowerCmd(GlobalFindCmds,
			"search commands by keywords").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority()
	addFindStrArgs(find)

	findShort := cmds.AddSub("/").
		RegPowerCmd(GlobalFindCmds,
			"shortcut of [find]").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority()
	addFindStrArgs(findShort)

	findUsage := find.AddSub("with-usage", "usage", "usg", "more", "m").
		RegPowerCmd(GlobalFindCmdsWithUsage,
			"search commands by find-strings, with usage info").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority()
	addFindStrArgs(findUsage)

	findUsageShort := cmds.AddSub("//").
		RegPowerCmd(GlobalFindCmdsWithUsage,
			"shortcut of [find.with-usage]").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority()
	addFindStrArgs(findUsageShort)

	findDetail := find.AddSub("with-full-info", "full", "f").
		RegPowerCmd(GlobalFindCmdsWithDetails,
			"search commands with strings, with full details").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority()
	addFindStrArgs(findDetail)

	findDetailShort := cmds.AddSub("///").
		RegPowerCmd(GlobalFindCmdsWithDetails,
			"shortcut of [find.with-full-info]").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority()
	addFindStrArgs(findDetailShort)
}

func RegisterTailSubCmds(cmds *core.CmdTree) {
	sub := cmds.AddSub("tail-sub").
		RegPowerCmd(DumpTailCmdSub,
			"display commands on the branch of the last command").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority()
	addFindStrArgs(sub)
	sub.AddArg("tag", "", "t")
	sub.AddArg("depth", "0", "d")

	subShort := cmds.AddSub("~").
		RegPowerCmd(DumpTailCmdSub,
			"shortcut of [tail-sub]").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority()
	addFindStrArgs(subShort)
	subShort.AddArg("tag", "", "t")
	subShort.AddArg("depth", "0", "d")

	subMore := sub.AddSub("with-usage", "usage", "usg", "more", "m").
		RegPowerCmd(DumpTailCmdSubWithUsage,
			"display commands on the branch of the last command, with usage info").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority()
	addFindStrArgs(subMore)
	subMore.AddArg("tag", "", "t")
	subMore.AddArg("depth", "0", "d")

	subMoreShort := cmds.AddSub("~~").
		RegPowerCmd(DumpTailCmdSubWithUsage,
			"shortcut of [tail-sub.with-usage]").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority()
	addFindStrArgs(subMoreShort)
	subMoreShort.AddArg("tag", "", "t")
	subMoreShort.AddArg("depth", "0", "d")

	subFull := sub.AddSub("with-full-info", "full", "f").
		RegPowerCmd(DumpTailCmdSubWithDetails,
			"display commands on the branch of the last command, with full details").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority()
	addFindStrArgs(subFull)
	subFull.AddArg("tag", "", "t")
	subFull.AddArg("depth", "0", "d")

	subFullShort := cmds.AddSub("~~~").
		RegPowerCmd(DumpTailCmdSubWithDetails,
			"shortcut of [tail-sub.with-full-info]").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority()
	addFindStrArgs(subFullShort)
	subFullShort.AddArg("tag", "", "t")
	subFullShort.AddArg("depth", "0", "d")
}

func RegisterCmdsListCmds(cmds *core.CmdTree) {
	addDumpCmdsArgs := func(cmd *core.Cmd) {
		cmd.SetPriority().
			SetQuiet().
			SetIgnoreFollowingDeps().
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
		SetIgnoreFollowingDeps().
		SetPriority().
		AddArg("cmd-path", "", "path", "p")

	cmd.AddSub("full", "f", "more", "m").
		RegPowerCmd(DumpCmdWithDetails,
			"display command full info").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority().
		AddArg("cmd-path", "", "path", "p")

	cmds.AddSub("=").
		RegPowerCmd(DumpTailCmdWithUsage,
			"shortcut of [cmd], if this is at the end of a flow:\n   * display usage of the last command\n   * flow will not execute").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority().
		AddArg("cmd-path", "", "path", "p")

	cmds.AddSub("==", "===").
		RegPowerCmd(DumpTailCmdWithDetails,
			"shortcut of [cmd], if this is at the end of a flow:\n   * display full info of the last command\n   * flow will not execute").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority().
		AddArg("cmd-path", "", "path", "p")
}

func RegisterFlowDescCmds(cmds *core.CmdTree) {
	desc := cmds.AddSub("desc").
		RegPowerCmd(DumpFlowSkeleton,
			"desc the flow execution in skeleton style").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority().
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
	cmds.AddSub("-").
		RegPowerCmd(DumpFlowSkeleton,
			"shortcut of [desc]").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority().
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
	cmds.AddSub("--").
		RegPowerCmd(DumpFlowSkeleton,
			"shortcut of [desc], unfold more(2L) trivial subflows").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority().
		AddArg("unfold-trivial", "2", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
	cmds.AddSub("---").
		RegPowerCmd(DumpFlowSkeleton,
			"shortcut of [desc], unfold more(3L) trivial subflows").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority().
		AddArg("unfold-trivial", "3", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")

	desc.AddSub("more", "m").
		RegPowerCmd(DumpFlowAllSimple,
			"desc the flow about to execute in lite style").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority().
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
	cmds.AddSub("+").
		RegPowerCmd(DumpFlowAllSimple,
			"shortcut of [desc.more]").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority().
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
	cmds.AddSub("++").
		RegPowerCmd(DumpFlowAllSimple,
			"shortcut of [desc.more], unfold more(2L) trivial subflows").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority().
		AddArg("unfold-trivial", "2", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
	cmds.AddSub("+++").
		RegPowerCmd(DumpFlowAllSimple,
			"shortcut of [desc.more], unfold more(3L) trivial subflows").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority().
		AddArg("unfold-trivial", "3", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")

	desc.AddSub("full", "f", "F").
		RegPowerCmd(DumpFlowAll,
			"desc the flow about to execute with details").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority().
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")

	desc.AddSub("dependencies", "depends", "depend", "dep", "os-cmd", "os").
		RegPowerCmd(DumpFlowDepends,
			"list the depended os-commands of the flow").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority()
	desc.AddSub("env-ops-check", "env-ops", "env-op", "env", "ops", "op", "e").
		RegPowerCmd(DumpFlowEnvOpsCheckResult,
			"desc the env-ops check result of the flow").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority()

	descFlow := desc.AddSub("flow").
		RegPowerCmd(DumpFlowSkeleton,
			"desc the flow execution in skeleton style").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority().
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
	descFlow.AddSub("more", "m").
		RegPowerCmd(DumpFlowSimple,
			"desc the flow execution in lite style").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority().
		AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d")
	descFlow.AddSub("full", "f").
		RegPowerCmd(DumpFlow,
			"desc the flow execution with details").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
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
		AddArg("git-address", "", "git", "address", "addr").
		AddArg("git-branch", "", "branch", "b")

	repoStatus := hub.AddSub("git-status", "status").
		RegPowerCmd(CheckGitRepoStatus,
			"check git status for all repos")
	addFindStrArgs(repoStatus)

	add.AddSub("local-dir", "local", "l").
		RegPowerCmd(AddLocalDirToHub,
			"add a local dir (could be a git repo) to hub").
		SetAllowTailModeCall().
		AddArg("path", "", "p")

	add.AddSub("pwd", "here", "p", "h").
		RegPowerCmd(AddPwdToHub,
			"add current working dir to hub")

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
}

func RegisterEnvManageCmds(cmds *core.CmdTree) *core.CmdTree {
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
			"save session env changes to local")

	envRm := env.AddSub("remove", "rm", "delete", "del").
		RegPowerCmd(RemoveEnvValNotSave,
			"remove specified env value in current session").
		SetAllowTailModeCall().
		AddArg("key", "", "k")

	envRm.AddSub("prefix", "pref", "pre").
		RegPowerCmd(RemoveEnvValHavePrefixNotSave,
			"remove env key-values with the specified key-prefix in current session").
		AddArg("prefix", "", "pre")

	env.AddSub("reset-session", "reset").
		RegPowerCmd(ResetSessionEnv,
			"clear all env values in current session, will not remove persisted values").
		AddArg("exclude-keys", "", "excludes", "exclude", "excl", "exc")

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
		AddArg("value", "", "val", "v").
		AddEnvOp("[[key]]", core.EnvOpTypeWrite)

	env.AddSub("add").
		RegPowerCmd(AddEnvKeyValue,
			"add key-value to env in current session, throws error if old value exists").
		AddArg("key", "", "k").
		AddArg("value", "", "val", "v").
		AddEnvOp("[[key]]", core.EnvOpTypeWrite)

	env.AddSub("update").
		RegPowerCmd(UpdateEnvKeyValue,
			"update key-value in env, throws error if old value not exists").
		AddArg("key", "", "k").
		AddArg("value", "", "val", "v")

	registerSimpleSwitch(abbrs,
		"borrowing commands' abbrs when setting env key-values",
		"sys.env.use-cmd-abbrs",
		"cmd")

	return env.Owner()
}

func RegisterEnvSnapshotManageCmds(cmds *core.CmdTree) {
	env := cmds.AddSub("snapshot", "snap", "ss").
		RegEmptyCmd(
			"env snapshot management")

	env.AddSub("list", "ls").
		RegPowerCmd(EnvListSnapshots,
			"list all saved snapshots")

	env.AddSub("remove", "rm", "delete", "del").
		RegPowerCmd(EnvRemoveSnapshot,
			"remove a saved snapshot").
		AddArg("snapshot-name", "", "snapshot", "name", "sn")

	env.AddSub("save", "s").
		RegPowerCmd(EnvSaveToSnapshot,
			"save session env to a named snapshot").
		AddArg("snapshot-name", "", "snapshot", "name", "sn").
		AddArg("overwrite", "true", "ow")

	load := env.AddSub("load", "l").
		RegPowerCmd(EnvLoadFromSnapshot,
			"load key-values from a saved snapshot").
		AddArg("snapshot-name", "", "snapshot", "name", "sn")

	load.AddSub("non-exist", "non-exists", "non", "n").
		RegPowerCmd(EnvLoadNonExistFromSnapshot,
			"load key-values from a saved snapshot if the keys are not in current session").
		AddArg("snapshot-name", "", "snapshot", "name", "sn")
}

func RegisterFlowManageCmds(cmds *core.CmdTree) {
	flow := cmds.AddSub("flow", "f")
	flow.RegPowerCmd(ListFlows,
		"same as [flow.list], list local saved but unlinked (to any repo) flows").
		SetAllowTailModeCall()
	addFindStrArgs(flow.Cmd())

	list := flow.AddSub("list", "ls").
		RegPowerCmd(ListFlows,
			"list local saved but unlinked (to any repo) flows").
		SetAllowTailModeCall()
	addFindStrArgs(list)

	flow.AddSub("save", "s").
		RegPowerCmd(SaveFlow,
			"save current commands as a flow").
		SetAllowTailModeCall().
		SetQuiet().
		SetIgnoreFollowingDeps().
		SetPriority().
		AddArg("new-cmd-path", "", "new-cmd", "cmd", "path", "new", "c").
		AddArg("help-str", "", "help", "h").
		AddArg("unfold-trivial", "0", "ut", "unfold", "unf", "uf", "u", "U", "trivial", "triv", "tri", "t").
		AddArg("auto-args", "", "args", "a").
		AddArg("to-dir", "", "dir", "d").
		AddArg("pack-subflow", "false", "pack-sub", "pack", "p").
		AddArg("quiet-overwrite", "false", "overwrite", "quiet", "q")

	flow.AddSub("set-help-str", "help", "h").
		RegPowerCmd(SetFlowHelpStr,
			"set help str to a saved flow").
		SetQuiet().
		AddArg("cmd-path", "", "path", "p").
		AddArg("help-str", "", "help", "h")

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

	mv := flow.AddSub("move-flows-to-dir", "move", "mv", "m")
	// TODO: impl flow.mv, flow.mv.here

	mvAll := mv.AddSub("all", "a")
	mvAll.RegPowerCmd(MoveSavedFlowsToLocalDir,
		moveFlowsToDirHelpStr).
		AddArg("path", "", "p")
	mvAll.AddSub("here", "h").
		RegPowerCmd(MoveSavedFlowsToPwd,
			"move all saved flows to current working dir")
}

func RegisterBgManageCmds(cmds *core.CmdTree) {
	bg := cmds.AddSub("background", "bg").
		RegEmptyCmd(
			"background tasks management")
	wait := bg.AddSub("wait").
		RegPowerCmd(WaitForAllBgTasksFinish,
			"wait for all tasks/threads to finish in current(must be main) thread")
	wait.AddSub("last", "latest").
		RegPowerCmd(WaitForLatestBgTaskFinish,
			"wait for the last tasks/threads to finish in current(must be main) thread")

	afterMain := wait.AddSub("after-main", "on-main", "at-main", "auto")
	afterMain.RegEmptyCmd(
		"auto wait for all tasks/threads to finish after main thread ends").
		AddVal2Env("sys.bg.wait", "true")
	afterMain.AddSub("off").
		RegEmptyCmd(
			"auto wait for all tasks/threads to finish after main thread ends").
		AddVal2Env("sys.bg.wait", "false")
}

func RegisterSessionCmds(cmds *core.CmdTree) {
	sessions := cmds.AddSub("sessions", "session", "s")

	sessions.AddSub("status").
		RegPowerCmd(SessionStatus,
			"show an executed or executing session by id").
		SetAllowTailModeCall().
		SetQuiet().
		AddArg("session-id", "", "session", "id")

	sessionsList := sessions.RegPowerCmd(
		ListSessions,
		"list or find executed and executing sessions").
		SetAllowTailModeCall().
		SetQuiet()
	addFindStrArgs(sessionsList)
	sessionsList.AddArg("max-count", "32", "limit", "max-cnt", "max")

	sessionsErr := sessions.AddSub("error", "failed", "fail", "err", "e", "f").
		RegPowerCmd(ListSessionsError,
			"list executed error sessions").
		SetAllowTailModeCall().
		SetQuiet()
	addFindStrArgs(sessionsErr)
	sessionsErr.AddArg("max-count", "32", "limit", "max-cnt", "max")

	sessionsDone := sessions.AddSub("done", "ok", "o").
		RegPowerCmd(ListSessionsDone,
			"list success executed sessions").
		SetAllowTailModeCall().
		SetQuiet()
	addFindStrArgs(sessionsDone)
	sessionsDone.AddArg("max-count", "32", "limit", "max-cnt", "max")

	sessionsRunning := sessions.AddSub("running", "run", "ing", "i").
		RegPowerCmd(ListSessionsRunning,
			"list executing sessions").
		SetAllowTailModeCall().
		SetQuiet()
	addFindStrArgs(sessionsRunning)
	sessionsRunning.AddArg("max-count", "32", "limit", "max-cnt", "max")

	errDesc := sessionsErr.AddSub("desc", "d", "-").
		RegPowerCmd(ErrorSessionDescLess,
			"desc the last failed session").
		SetAllowTailModeCall().
		SetQuiet().
		AddArg("unfold-trivial", "1", "ut", "unfold", "unf", "uf", "u", "U", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d", "D")

	errDesc.AddSub("more", "m").
		RegPowerCmd(ErrorSessionDescMore,
			"desc the last failed session with more details").
		SetAllowTailModeCall().
		SetQuiet().
		AddArg("unfold-trivial", "1", "ut", "unfold", "unf", "uf", "u", "U", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d", "D")

	errDesc.AddSub("full", "f").
		RegPowerCmd(ErrorSessionDescFull,
			"desc the last failed session with full details").
		SetAllowTailModeCall().
		SetQuiet().
		AddArg("unfold-trivial", "1", "ut", "unfold", "unf", "uf", "u", "U", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d", "D")

	doneDesc := sessionsDone.AddSub("desc", "d", "-").
		RegPowerCmd(DoneSessionDescLess,
			"desc the last succeeded session").
		SetAllowTailModeCall().
		SetQuiet().
		AddArg("unfold-trivial", "1", "ut", "unfold", "unf", "uf", "u", "U", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d", "D")

	doneDesc.AddSub("more", "m").
		RegPowerCmd(DoneSessionDescMore,
			"desc the last succeeded session with more details").
		SetAllowTailModeCall().
		SetQuiet().
		AddArg("unfold-trivial", "1", "ut", "unfold", "unf", "uf", "u", "U", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d", "D")

	doneDesc.AddSub("full", "f").
		RegPowerCmd(DoneSessionDescFull,
			"desc the last succeeded session with full details").
		SetAllowTailModeCall().
		SetQuiet().
		AddArg("unfold-trivial", "1", "ut", "unfold", "unf", "uf", "u", "U", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d", "D")

	runningDesc := sessionsRunning.AddSub("desc", "d", "-").
		RegPowerCmd(RunningSessionDescLess,
			"desc the last executing session").
		SetAllowTailModeCall().
		SetQuiet().
		SetHideInSessionsLast().
		AddArg("unfold-trivial", "1", "ut", "unfold", "unf", "uf", "u", "U", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d", "D")

	runningDesc.AddSub("more", "m").
		RegPowerCmd(RunningSessionDescMore,
			"desc the last executing session with more details").
		SetAllowTailModeCall().
		SetQuiet().
		SetHideInSessionsLast().
		AddArg("unfold-trivial", "1", "ut", "unfold", "unf", "uf", "u", "U", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d", "D")

	runningDesc.AddSub("full", "f").
		RegPowerCmd(RunningSessionDescFull,
			"desc the last executing session with full details").
		SetAllowTailModeCall().
		SetQuiet().
		SetHideInSessionsLast().
		AddArg("unfold-trivial", "1", "ut", "unfold", "unf", "uf", "u", "U", "trivial", "triv", "tri", "t").
		AddArg("depth", "32", "d", "D")

	runningDesc.AddSub("monitor", "mon").
		RegPowerCmd(RunningSessionDescMonitor,
			"desc the last executing session in monitor (compact) mode").
		SetAllowTailModeCall().
		SetQuiet().
		SetHideInSessionsLast()

	desc := sessions.AddSub("desc", "d", "-").
		RegPowerCmd(SessionDescLess,
			"desc executed/ing session").
		SetAllowTailModeCall()
	desc.AddArg("session-id", "", "session", "id")
	desc.AddArg("unfold-trivial", "1", "ut", "unfold", "unf", "uf", "u", "U", "trivial", "triv", "tri", "t")
	desc.AddArg("depth", "32", "d", "D")

	descMore := desc.AddSub("more", "m").
		RegPowerCmd(SessionDescMore,
			"desc executed/ing session with more details").
		SetAllowTailModeCall()
	descMore.AddArg("session-id", "", "session", "id")
	descMore.AddArg("unfold-trivial", "1", "ut", "unfold", "unf", "uf", "u", "U", "trivial", "triv", "tri", "t")
	descMore.AddArg("depth", "32", "d")

	descFull := desc.AddSub("full", "f").
		RegPowerCmd(SessionDescFull,
			"desc executed/ing session with full details").
		SetAllowTailModeCall()
	descFull.AddArg("session-id", "", "session", "id")
	descFull.AddArg("unfold-trivial", "1", "unfold", "unf", "uf", "u", "trivial", "triv", "tri", "t")
	descFull.AddArg("depth", "32", "d")

	desc.AddSub("monitor", "mon").
		RegPowerCmd(SessionDescMonitor,
			"desc executed/ing session in monitor (compact) mode").
		SetAllowTailModeCall().
		AddArg("session-id", "", "session", "id")

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

	sessions.AddSub("retry", "r").
		RegAdHotFlowCmd(SessionRetry,
			"find a session by id, retry running it, executed commands will be skipped").
		AddArg("session-id", "", "session", "id")

	last.AddSub("retry", "r", "R").
		RegAdHotFlowCmd(LastSessionRetry,
			"retry running the last session, executed commands will be skipped")

	sessionsErr.AddSub("retry", "r", "R").
		RegAdHotFlowCmd(LastErrorSessionRetry,
			"find the last failed session, retry running it, executed commands will be skipped")

	remove := sessions.AddSub("remove", "delete", "rm")
	remove.RegPowerCmd(RemoveSession,
		"remove an executed or (wrongly classify as)executing session by id").
		SetAllowTailModeCall().
		SetQuiet().
		AddArg("session-id", "", "session", "id").
		AddArg("remove-running", "false", "force")

	removeByFind := remove.AddSub("by-search", "by-find", "search", "find", "batch").
		RegPowerCmd(FindAndRemoveSessions,
			"clear executed sessions by keywords").
		SetAllowTailModeCall().
		SetQuiet()
	addFindStrArgs(removeByFind)
	removeByFind.AddArg("remove-running", "false", "force")

	remove.AddSub("all", "a").
		RegPowerCmd(RemoveAllSessions,
			"clear all executed sessions")

	sessions.AddSub("set-keep-duration", "keep-duration", "keep", "kd").
		RegPowerCmd(SetSessionsKeepDur,
			"set the keeping duration of executed sessions").
		AddArg("duration", "72h", "dur")
}

func RegisterBlenderCmds(cmds *core.CmdTree) {
	cmds.AddSub("forest-mode", "forest").
		RegPowerCmd(SetForestMode,
			"run following commands in forest-mode: reset env on each command, but not reset on their subflows").
		SetExeInExecuted().
		SetQuiet()

	replace := cmds.AddSub("replace", "repl").
		RegPowerCmd(BlenderReplaceOnce,
			"during executing, replace a command with another one (only replace once)").
		SetQuiet().
		SetIsBlenderCmd().
		AddArg("src", "", "source", "target").
		AddArg("dest", "")

	replace.AddSub("all").
		RegPowerCmd(BlenderReplaceForAll,
			"during executing, replace a command with another one (replace all)").
		SetQuiet().
		SetIsBlenderCmd().
		AddArg("src", "", "source", "target").
		AddArg("dest", "")

	// TODO: disable blenders now, too many bugs
	return

	remove := cmds.AddSub("remove", "rm", "delete", "del").
		RegPowerCmd(BlenderRemoveOnce,
			"during executing, remove a command in flow (only remove once)").
		SetQuiet().
		SetIsBlenderCmd().
		AddArg("target", "")

	remove.AddSub("all").
		RegPowerCmd(BlenderRemoveForAll,
			"during executing, remove a command in flow (remove all)").
		SetQuiet().
		SetIsBlenderCmd().
		AddArg("target", "")

	insert := cmds.AddSub("insert", "ins").
		RegPowerCmd(BlenderInsertOnce,
			"during executing, insert a command before another matched one (insert once)").
		SetQuiet().
		SetIsBlenderCmd().
		AddArg("target", "").
		AddArg("new", "")

	insert.AddSub("all").
		RegPowerCmd(BlenderInsertForAll,
			"during executing, insert a command before another matched one (insert for all matched)").
		SetQuiet().
		SetIsBlenderCmd().
		AddArg("target", "").
		AddArg("new", "")

	after := insert.AddSub("after").
		RegPowerCmd(BlenderInsertAfterOnce,
			"during executing, insert a command after another matched one (insert once)").
		SetQuiet().
		SetIsBlenderCmd().
		AddArg("target", "").
		AddArg("new", "")

	after.AddSub("all").
		RegPowerCmd(BlenderInsertAfterForAll,
			"during executing, insert a command after another matched one (insert for all matched)").
		SetQuiet().
		SetIsBlenderCmd().
		AddArg("target", "").
		AddArg("new", "")

	cmds.AddSub("clear", "clean", "reset").
		RegPowerCmd(BlenderClear,
			"clear all flow modification schedulings").
		SetQuiet().
		SetIsBlenderCmd()
}

func RegisterCtrlCmds(cmds *core.CmdTree) {
	breaks := cmds.AddSub("break", "breaks", "pause")
	breaks.RegEmptyCmd("set break point at the position of this command").
		SetQuiet().
		AddVal2Env("sys.breakpoint.here.now", "true")

	breaksAt := breaks.AddSub("at", "before").
		RegPowerCmd(DbgBreakBefore,
			// TODO: get 'sep' from env or other config
			"setup before-command break points, use ',' to seperate multipy commands").
		SetQuiet().
		AddArg("break-points", "", "break-point", "cmds", "cmd")

	breaks.AddSub("after", "post").
		RegPowerCmd(DbgBreakAfter,
			// TODO: get 'sep' from env or other config
			"setup after-command break points, use ',' to seperate multipy commands").
		SetQuiet().
		AddArg("break-points", "", "break-point", "cmds", "cmd")

	breaksAt.AddSub("end", "finish").
		RegPowerCmd(DbgBreakAtEnd,
			"set break point after the last command").
		SetQuiet().
		SetPriority()

	breaks.AddSub("list", "ls").
		RegPowerCmd(DbgBreakStatus,
			"list all break point status").
		SetQuiet()

	breaks.AddSub("clear", "clean").
		RegPowerCmd(DbgBreakClean,
			"clear all break points except [break.here] and [break]").
		SetQuiet()

	wait := cmds.AddSub("wait-execute", "wait-exec", "wait-exe").
		RegPowerCmd(DbgWaitSecExecute,
			"wait for specified duration before executing each commands").
		SetQuiet().
		AddArg("seconds", "3", "second", "sec", "s", "S")

	wait.AddSub("after", "at-end", "at-finish", "post-execute", "end", "finish").
		RegPowerCmd(DbgWaitSecExecuteAtEnd,
			"wait for specified duration after executing each commands").
		SetQuiet().
		AddArg("seconds", "3", "second", "sec", "s", "S")
}

func RegisterDbgCmds(cmds *core.CmdTree) {
	registerSimpleSwitch(cmds,
		"recover from internal error and give a frendly message",
		"sys.panic.recover",
		"recover")

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
}

func RegisterDisplayCmds(cmds *core.CmdTree) {
	cmds.AddSub("style").
		RegPowerCmd(SetDisplayStyle,
			"set executing display style: bold, slash, corner, ascii, utf8(default)").
		AddArg("style", "s").
		SetQuiet()

	completion := cmds.AddSub("completion")
	registerSimpleSwitch(completion,
		"hidden-style completion",
		"display.completion.hidden",
		"hidden")

	registerSimpleSwitch(completion,
		"command abbrs completion",
		"display.completion.abbr",
		"abbrs", "abbr")

	registerSimpleSwitch(completion,
		"shortcut command completion",
		"display.completion.shortcut",
		"shortcut")

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
		"values of env path sys.*' display in executing",
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

	registerSimpleSwitch(cmds,
		"display sensitive key-value of env and args",
		"display.sensitive",
		"sensitive", "pwd", "masked")

	mod := cmds.AddSub("command", "cmd")

	registerSimpleSwitch(mod,
		"quiet module display in executing",
		"display.mod.quiet",
		"quiet")

	registerSimpleSwitch(mod,
		"display user input command name in executing",
		"display.mod.input-name",
		"input-name", "input")

	registerSimpleSwitchEx(mod.GetOrAddSub("input-name"),
		"display user input command name with realname as comment in executing",
		[]string{
			"display.mod.input-name",
			"display.mod.input-name.with-realname",
		},
		"with-realname", "realname", "real", "r")

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

	registerSimpleSwitch(cmds,
		"show tips box and tip messages",
		"display.tip",
		"tips")

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

	hub := cmds.AddSub("hub", "h")
	hub.AddSub("init").
		RegPowerCmd(EnsureDefaultGitRepoInHub,
			"trigger [hub.init] at the first time running")
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

	echo := cmds.AddSub("echo")
	echo.RegPowerCmd(DbgEcho,
		"print message from argv").
		AddArg("message", "", "msg", "m").
		AddArg("color", "", "colour", "clr", "c")

	echo.AddSub("ln").
		RegPowerCmd(DbgEchoLn,
			"print an empty line").
		SetQuiet()

	sys := cmds.AddSub("sys")
	sys.AddSub("ext-executor", "ext-exec", "ext-exe").
		RegPowerCmd(SysSetExtExecutor,
			"set ext-executor for modules").
		AddArg("ext", "").
		AddArg("executor", "", "exec", "exe")
}

func RegisterJoinCmds(cmds *core.CmdTree) {
	join := cmds.AddSub("join")
	join.AddSub("new", "add", "arg", "new_arg").
		RegPowerCmd(JoinNew,
			"add new argument").
		AddArg("key", "").
		AddArg("value", "", "val")
	join.AddSub("run").
		RegPowerCmd(JoinRun,
			"run with arguments").
		AddArg("cmd", "")
}

func RegisterNoopCmds(cmds *core.CmdTree) {
	cmds.AddSub("noop").
		RegPowerCmd(Noop,
			"do exactly nothing")

	cmds.AddSub("dummy", "dm").
		RegPowerCmd(Dummy,
			"dummy command for testing")
}

func RegisterHookCmds(cmds *core.CmdTree) {
	hook := cmds.AddSub("hook").
		RegEmptyCmd("a toolbox for hooking flows to ticat system events")

	hook.AddSub("exit").
		RegEmptyCmd(
			"execute flow when session exit: done or error").
		AddArg("flow", "", "f").
		AddArg2Env("sys.event.hook.exit", "flow")

	hook.AddSub("error").
		RegEmptyCmd(
			"execute flow when session exit with error").
		AddArg("flow", "", "f").
		AddArg2Env("sys.event.hook.error", "flow")

	hook.AddSub("done").
		RegEmptyCmd(
			"execute flow when session done without error").
		AddArg("flow", "", "f").
		AddArg2Env("sys.event.hook.done", "flow")
}

func RegisterApiCmds(cmds *core.CmdTree) {
	cmd := cmds.AddSub("cmd")

	cmd.AddSub("type").
		RegPowerCmd(ApiCmdType,
			"get command type").
		SetIsApi().
		AddArg("cmd", "")

	cmd.AddSub("meta").
		RegPowerCmd(ApiCmdMeta,
			"get command meta file path").
		SetIsApi().
		AddArg("cmd", "")

	cmd.AddSub("dir").
		RegPowerCmd(ApiCmdDir,
			"get command dir path").
		SetIsApi().
		AddArg("cmd", "")

	cmd.AddSub("path").
		RegPowerCmd(ApiCmdPath,
			"get command executable file path").
		SetIsApi().
		AddArg("cmd", "")

	cmdList := cmd.AddSub("list")
	cmdList.AddSub("all").
		RegPowerCmd(ApiCmdListAll,
			"list all commands").
		SetIsApi()

	env := cmds.AddSub("env")

	env.AddSub("value", "val", "v").
		RegPowerCmd(DisplayEnvVal,
			"show specified env value in current session").
		SetIsApi().
		AddArg("key", "", "k")
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
will do auto moving if the arg "path" is empty and:
    * one(and only one) local(not linked to a repo) dir exists in hub
    then flows will move to that dir`
