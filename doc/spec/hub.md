# [Spec] Hub: list/add/add-local/disable/enable/purge

## List dirs in hub
Hub is a set of local local dirs witch **ticat** knows
```bash
$> ticat hub.list
$> ticat hub.list <find-str>
## Example:
$> ticat hub.list examples
```

## Add/update git addresses
```bash
## Add(link) git address
$> ticat hub.add <github-id/repo-name>
$> ticat hub.add <git-full-address>
## Example:
$> ticat hub.add innerr/tidb.ticat

## Update all linked git repos
$> ticat hub.update

## Add builtin (default) address
$> ticat hub.init
```

## Add local dirs
```bash
$> ticat hub.add.local path=<dir>
## Example:
$> ticat hub.add.local path=./mymods
```

## Disable repos or dirs, the modules in disabled repos or dirs can't be loaded
```bash
$> ticat hub.disable <find-str>
$> ticat hub.enable <find-str>
```

## Unlink repos/dirs from ticat
Purge will delete all content of linked repos,
but only remove meta info from **ticat** for local dirs.
Only disabled ones can be purged
```bash
$> ticat hub.purge <find-str>
$> ticat hub.purge.all
```

## The store repo/file
The saved dir is defined by env key "sys.paths.hub",
All git cloned repos will be here.

There is a repo list file, its name is defined by env key "strs.hub-file-name".
The format is multi lines, each line has fields `git-address` `add-reason` `dir-path` `help-str` `on-or-off` seperated by "\t".
