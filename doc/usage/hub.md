# Hub: get modules and flows from others

The **hub** is **ticat**'s package management system. It allows you to discover, install, and manage modules and flows from various sources.

## What is hub?

The **hub** is a collection of local directories linked with **ticat**. Modules are loaded on startup by scanning these directories:

```
 ┌────────────────────────────────┐
 │ Ticat Hub                      │
 │  ┌──────────────────────────┐  │
 │  │ Normal Local Dir         │  │
 │  │ ┌─────┐  ┌─────┐         │  │
 │  │ │ Mod │  │ Mod │  ...    │  │
 │  │ └─────┘  └─────┘         │  │
 │  └──────────────────────────┘  │
 │  ┌──────────────────────────┐  │
 │  │ Repo cloned to Local Dir │  │
 │  │ ┌─────┐                  │  │
 │  │ │ Mod │  ...             │  │
 │  │ └─────┘                  │  │
 │  └──────────────────────────┘  │
 │  ┌──────────────────────────┐  │
 │  │ Repo cloned to Local Dir │  │
 │  │ ┌─────┐  ┌─────┐         │  │
 │  │ │ Mod │  │ Mod │  ...    │  │
 │  │ └─────┘  └─────┘         │  │
 │  └──────────────────────────┘  │
 └────────────────────────────────┘
```

### List directories in hub

```bash
# List all directories
$> ticat hub

# Find/filter directories
$> ticat hub <find-str>
```

## Add a git repository to hub

### Add from GitHub

You can add repositories from GitHub using the short format:

```bash
$> ticat hub.add <github-id/repo-name>

# Example:
$> ticat hub.add innerr/tidb.ticat
```

### Add from any git server

Or use the full git address:

```bash
$> ticat hub.add <git-full-address>

# Example:
$> ticat hub.add git@github.com:innerr/tidb.ticat.git
```

### Recursive cloning

If a repository has sub-repositories ([what is sub-repo](../spec/repo-tree.md)), they will be recursively cloned to local as well.

All cloned repositories are stored under a specific folder defined by the environment key `sys.paths.hub`. These directories are under **ticat**'s management and are called `managed` directories.

**Note**: If you add a repository that already exists, **ticat** will update it using `git pull`.

## Initialize hub with default repository

**ticat** provides a command to add a default repository defined by the environment key `sys.hub.init-repo`:

```bash
$> ticat hub.init
```

This repository contains the most common modules and is useful for new users to get started quickly.

## Update all managed repositories

Keep all your repositories up to date:

```bash
$> ticat hub.update
```

**Note**: Disabled repositories won't be updated.

## Add a local directory to hub

You can add any local directory to the hub:

```bash
$> ticat hub.add.local path=<dir>

# Example:
$> ticat hub.add.local path=./mymods
```

The directory could be:
- A regular directory with your modules
- A repository you manually cloned

**ticat** treats local directories as `unmanaged`, meaning it loads modules from them but won't modify their contents.

## Disable/enable directories

You can temporarily disable directories without removing them:

```bash
# Disable directories matching a pattern
$> ticat hub.disable <find-str>

# Enable directories matching a pattern
$> ticat hub.enable <find-str>

# Examples:
$> ticat hub.disable mymods
$> ticat hub.enable mymods
```

Disabled directories remain in the hub but their modules won't be loaded.

## Permanently remove directories from hub

### Remove specific directories

A directory must be disabled first, then you can purge it:

```bash
# Disable first
$> ticat hub.disable <find-str>

# Then purge
$> ticat hub.purge <find-str>
```

### Remove all inactive directories

```bash
$> ticat hub.purge.all
```

**Important behavior**:
- **Managed directories** (cloned repositories): Will be completely deleted from the file system
- **Unmanaged directories** (local dirs): Will be removed from hub but kept on the file system

## All hub commands overview

```bash
$> ticat h:-
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
         'move all saved flows to a local dir (could be a git repo)'
```

## Best practices

1. **Use tags for organization**: Authors can add tags to modules to help users find them
2. **Keep repositories updated**: Run `hub.update` regularly to get the latest modules
3. **Disable instead of purge**: If you're unsure, disable repositories first before purging
4. **Share your modules**: Create your own repository and share it with the community

## Next steps

- Learn how to [search and use commands](./cmds.md)
- Understand [environment key-values](./env.md)
- Create [flows](./flow.md) to automate your workflows
