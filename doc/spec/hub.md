# [Spec] Hub: list/add/add-local/disable/enable/purge

This specification describes the hub system in **ticat**, which manages repositories and directories containing modules and flows.

## Overview

```bash
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

## List directories in hub

The hub is a collection of local directories that **ticat** knows about:

```bash
# List all directories
$> ticat hub.list

# Filter directories
$> ticat hub.list <find-str>

# Short form (hub == hub.list)
$> ticat hub

# Example:
$> ticat hub.list examples
```

## Add and update git repositories

### Add a repository

```bash
# Add from GitHub (short format)
$> ticat hub.add <github-id/repo-name>

# Add from any git server (full address)
$> ticat hub.add <git-full-address>

# Example:
$> ticat hub.add innerr/tidb.ticat
```

### Update all linked repositories

```bash
# Update all active repositories
$> ticat hub.update
```

### Add the default/builtin repository

```bash
# Add the default repository defined by sys.hub.init-repo
$> ticat hub.init
```

## Add local directories

```bash
$> ticat hub.add.local path=<dir>

# Example:
$> ticat hub.add.local path=./mymods
```

Local directories are "unmanaged" - **ticat** loads modules from them but won't modify their contents.

## Disable and enable repositories

Disabled repositories remain in the hub but their modules won't be loaded:

```bash
# Disable repositories matching a pattern
$> ticat hub.disable <find-str>

# Enable repositories matching a pattern
$> ticat hub.enable <find-str>
```

## Remove repositories from hub

### Pruning rules

- **Managed directories** (cloned repos): Will be completely deleted from the file system
- **Unmanaged directories** (local dirs): Will be removed from hub but kept on file system
- **Prerequisite**: Directories must be disabled before purging

### Commands

```bash
# Purge a specific repository (must be disabled first)
$> ticat hub.purge <find-str>

# Purge all inactive repositories
$> ticat hub.purge.all
```

## Storage locations

### Repository storage

The directory for cloned repositories is defined by the environment key `sys.paths.hub`.

### Hub configuration file

The hub configuration file stores the list of repositories:
- **Location**: Defined by `strs.hub-file-name`
- **Format**: Multi-line file
- **Fields per line** (tab-separated):
  1. `git-address`
  2. `add-reason`
  3. `dir-path`
  4. `help-str`
  5. `on-or-off`

## Best practices

1. **Use descriptive repository names**: Helps when searching and filtering
2. **Disable before purging**: Prevents accidental deletion
3. **Update regularly**: Keep repositories current with `hub.update`
4. **Organize local development**: Use `hub.add.local` for your development directories
5. **Share useful modules**: Create repositories to share modules with your team
