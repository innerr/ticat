# [Spec] Module: meta file

## Ticat will load modules from hub, aka, local dirs
We could add local dir to **ticat** by:
```bash
$> ticat hub.add.local path=<dir-path>
```
Or add a git repo, **ticat** will do "git clone" to local dir:
```bash
$> ticat hub.add.local path=<dir-path>
```

## Dir struct and command tree
**Ticat** will search any file with ext name ".ticat" in a local dir in hub.
For each "file-or-dir-path.ticat" file, will try to register it to the command tree.

The position in the tree will depends on the relative path to the repo-root,
the ext(s) will be ignored in registering.

For a dir "dir-path" with "dir-path.ticat", the ".ticat" file format is:
```
help = <help string>
abbrs = <abbr-1>|<abbr-2>|<abbr-3>...
cmd = <a relative path based on this dir>
```
If the values are quoted, the "'\"" will be removed.
Key "abbrs" and "abbr" are equal.
If "cmd" is not provided, a empty command will be registered.

Multi-line value is supported with "\\"(single \) as line breaker.

A dir without meta file "dir-path.ticat" will not be searched,
so modules inside it will not be registered.

For a file "file-path" with "file-path.ticat", the ".ticat" file format is:
```
help = <help string>
abbrs = <abbr-1>|<abbr-2>|<abbr-3>...

[args]
arg-1|<abbr-x>|<abbr-y> = <arv-1 default value>
arg-2 = <arv-2 default value>
...

[env]
env-key-1 = <env-op>
env-key-2 = <env-op> : <env-op> : ...
...

[dep]
os-cmd-1 = <why this command depends on this os-cmd>
os-cmd-2 = <why this command depends on this os-cmd>
...
```
The "help" and "abbrs" are the same with dir type of registering.
The `[dep]` section defines what os-command will be called in the command's code.

The `[args]` section defines the command's args with order.
Abbrs definition are allowed, seperate them with "|".

The `[env]` section defines witch keys will read or write in the command's code.
"env-op" value could be: "read", "write", "may-read", "may-write".
The sequence of "env-op" could be one or more value with orders, seperated by ":".
Abbrs definition are also allowed in every path segment of the keys.

## Example
Dir struct:
```
<repo-root>
├── README.md
├── tidb (dir)
│   ├── stop.bash
│   └── stop.bash.ticat
├── tidb.ticat
└── misc (dir)
    ├── run.bash
    └── run.bash.ticat
```

File "tidb.ticat":
```
help = simple test toolbox for tidb
abbrs = ti|db
```

File "stop.bash.ticat":
```
help = stop a tidb cluster
abbr = down|dn

[args]
force|f = true

[env]
cluster|c.name|n = read:write
```

What will happend:
* "run.bash" won't be registered, because "misc" with not "misc.ticat"
* "tidb" will be registered, because of "tidb.ticat"

Usage:
```bash
## This will do nothing
$> ticat tidb
## This will do nothing either, we use abbr to call "tidb"
$> ticat db

## Display command info of "tidb.stop", the ext name ".bash" is ignored
$> ticat cmds.tree tidb.stop
$> ticat cmds.tree db.down
$> ticat m.t db.dn

## Error: "tidb.stop" will read a key from env without any provider:
$> ticat tidb.stop
$> ticat db.stop

## Proper ways to call "tidb.stop", some use abbrs in the key path:
$> ticat {cluster.name = test} db.stop
$> ticat {c.n = test} db.stop
$> ticat <a-command-provide-env-key-cluster-name> : db.stop

## Use saved env to call "tidb.stop":
$> ticat {cluster.name = test} env.save
...
$> ticat db.stop
```

## Register conflictions
When more than one command register to a same command path,
or more than one abbrs for a command path segment,
**ticat** will throws errors by confllictions.

In this case, manually edit the hub file is recommended for now.
(TODO: better way to detect and solve conflictions)
