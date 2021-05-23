# Hub: get modules and flows from others

## What is hub?
The **hub** is all local dirs witch **ticat** knows,
these dirs have all modules, **ticat** will scan them on bootstrap:
```
 ┌────────────────────────────────┐
 │ TiCat Hub                      │
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

This command list all dirs in hub:
```bash
$> ticat hub.ls
```
Or find dirs: (TODO: impl)
```bash
$> ticat hub.ls <find-str>
```

### Add a git repo to hub
These dirs could be a git repo, **ticat** will `git clone` it to local dir when:
```bash
$> ticat hub.add <github-id/repo-name>
```
or
```bash
$> ticat hub.add <git-full-address>
```

If a repo has sub-repos([what is sub-repo](../spec/repo-tree.md)),
they will be recursively clone to local too.

All cloned repos store under a specific folder defined by env key `sys.paths.hub`.
These dirs are under **ticat**'s management, are `managed` dirs.

If we add an existed repo to hub, the repo will be updated by `git pull`.

### Init hub by adding default git address
This command could add the default git address defined by env key `sys.hub.init-repo` to hub:
```bash
$> ticat hub.init
```
This repo has the most common modules, it's useful for new users.

### Update all managed repos
```bash
$> ticat hub.update
```
The disabled repos won't be updated.

### Add a local dir to hub
Dirs in hub could be a normal dir added to **ticat** by:
```bash
$> ticat hub.add.local path=<dir>
```

The dir could be a repo manually cloned to local,
**ticat** treat it as normal dir, they are `unmanaged`,
it means **ticat** load modules from it but won't change anything in it.

## Disable/enable dirs
We use find string as arg, to disable/enable multi dirs in one command:
```bash
$> ticat hub.disable <find-str>
$> ticat hub.enable <find-str>
```

## Permanently remove dirs from hub
A dir must be disabled first, then use **purge** command to remove it:
```bash
$> ticat hub.purge <find-str>
```
This command remove all disabled dirs:
```bash
$> ticat hub.purge.all
```

A managed dir will be totally deleted from file system.
A normal(unmanaged) dir will be removed from hub but keep on file system.
