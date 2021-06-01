# [Spec] Hub: list/add/add-local/disable/enable/purge

## Overview
```
$> ticat cmds.tree.simple hub
[hub]
     'list dir and repo info in hub'
    [clear]
         'remove all repos from hub'
    [init]
         'add and pull basic hub-repo to local'
    [add-and-update]
         'add and pull a git address to hub'
        [local-dir]
             'add a local dir (could be a git repo) to hub'
    [list]
         'list dir and repo info in hub'
    [purge]
         'remove an inactive repo from hub'
        [purge-all-inactive]
             'remove all inactive repos from hub'
    [update-all]
         'update all repos and mods defined in hub'
    [enable-repo]
         'enable matched git repos in hub'
    [disable-repo]
         'disable matched git repos in hub'
    [move-flows-to-dir]
         'move all saved flows to a local dir (could be a git repo).
          auto move:
              * if one (and only one) local dir exists in hub
              * and the arg "path" is empty
              then flows will move to that dir'
```

## List dirs in hub
Hub is a set of local local dirs which **ticat** knows
```
$> ticat hub.list
$> ticat hub.list <find-str>

## Command `hub` == `hub.list`
$> ticat hub

## Example:
$> ticat hub.list examples
```

## Add/update git addresses
```
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
```
$> ticat hub.add.local path=<dir>

## Example:
$> ticat hub.add.local path=./mymods
```

## Disable repos or dirs, the modules in disabled repos or dirs can't be loaded
```
$> ticat hub.disable <find-str>
$> ticat hub.enable <find-str>
```

## Unlink repos/dirs from ticat

Purge will delete all content of linked repos,
but only remove meta info from **ticat** for local dirs.
Only disabled ones can be purged
```
$> ticat hub.purge <find-str>
$> ticat hub.purge.all
```

## The stored repo/file
The saved dir is defined by env key "sys.paths.hub",
All git cloned repos will be here.

There is a repo list file, its name is defined by env key "strs.hub-file-name".
The format is multi lines, each line has fields `git-address` `add-reason` `dir-path` `help-str` `on-or-off` seperated by "\t".
