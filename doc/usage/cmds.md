# Locate and display commands' info

The amount of commands will be huge,
`cmds` is a toolset for us to find what we need.


## Flat(list) mode are mostly for searching
Display in flat mode, `flat` `ls` and `f` are equal:
```
$> ticat cmds.ls
$> ticat cmds.flat
$> ticat c.f
...
[desc|d|D]
     'desc the flow about to execute'
    - cmd-type:
        power (quiet) (priority)
[flow|f|F]
     'desc the flow execution'
    - full-cmd:
        desc.flow
    - full-abbrs:
        desc|d|D.flow|f|F
    - cmd-type:
        power (quiet) (priority)
...
```
Using `c.f` without args are barely useless because the output is way too long.
`grep` might helps but we have better way to looking for commands.


Filter(find) by strings, up to 6 strings,
the finding strings could be anything:
```
$> ticat c.f <str>
$> ticat c.f <str> <str> ...
```

Find commands which has tag `@selftest`:
```
$> ticat c.f @selftest
[test|t]
     '@selftest @ready'
    - full-cmd:
        examples.test
    - full-abbrs:
        examples|example|exam|ex|samples|sample|sam.test|t
    - from:
        git@github.com:innerr/examples.ticat
    - flow:
        dbg.step.on : ex.test.conn : ex.test.interact : ex.test.exts
...
```

Tags are only normal strings,
we recommend that module authors adding tags in the help string,
then explain the meaning of the tags in the repo README,
so users could find things they want.

There are a few tags has conventional meanings:
* `@ready` means "ready-to-go", normally are powerful/useful flows
* `@selftest` means this command is for self testing modules in the repo it belongs.

Find commands which comes from git address "quick-start-mod.ticat":
```
$> ticat c.f quick-start-mod
[mod-1]
     'a simple ticat module'
    - args:
        name = foo
        age = unknown
    - cmd-type:
        file
    - from:
        git@github.com:innerr/quick-start-mod.ticat
```

Find commands which write(provide) env key-value "examples.my-key",
```
$> ticat c.f examples.my-key write
[conn-may-write|maywrite|mayw|mw]
     'this module may-write the example key'
    - full-cmd:
        examples.conn-may-write
    - full-abbrs:
        examples|example|exam|ex|samples|sample|sam.conn-may-write|maywrite|mayw|mw
    - env-ops:
        examples.my-key = may-write
    - from:
        git@github.com:innerr/examples.ticat
...
```

When we just added a repo to **ticat**,
The fast way to put it into work is:
* Find what's in it with tag `@ready` by searsh: `ticat c.f <git-address> @ready`.
* Use `desc` to check out what env keys these ready-to-go commands need.
* Find what commands provide those env keys (or manually set to env by `{k=v}`).


### Lite mode
`cmds.flat.simple`(c.f.s) is similar to `cmds.flat`(c.f), just show less info:
```
$> ticat c.f.s mysql
[exec|exe|e|E]
    - full-cmd:
        mysql.exec
    - full-abbrs:
        mysql|my.exec|exe|e|E
...
```


## Non-flat mode
The command `cmds` itself could display all info about a command:
```
$> ticat cmds desc
[desc|d|D]
     'desc the flow about to execute'
    - cmd-type:
        power (quiet) (priority)
```
`cmds` is used very frequently, it has a short name `c`.
```
$> ticat c cmds
[cmds|cmd|c|C]
     'display cmd info, no sub tree'
    - args:
        path|p|P = ''
```

The command `cmds.tree`(c.t) will display the command's info and it's sub-tree's:
```
$> ticat c.t desc
[desc|d|D]
     'desc the flow about to execute'
    - cmd-type:
        power (quiet) (priority)
    [lite|simple|sim|s|S]
         'desc the flow about to execute in lite style'
        - full-cmd:
            desc.lite
        - full-abbrs:
            desc|d|D.lite|simple|sim|s|S
        - cmd-type:
            power (quiet) (priority)
    [flow|f|F]
         'desc the flow execution'
        - full-cmd:
            desc.flow
        - full-abbrs:
            desc|d|D.flow|f|F
        - cmd-type:
            power (quiet) (priority)
...
```

`cmds.tree.simple`(c.t.s) display only the command names,
convenient to observe the tree struct:
```
$> ticat c.t.s c
[cmds|cmd|c|C]
    [tree|t|T]
        [lite|simple|sim|s|S]
    [list|ls|flatten|flat|f|F]
        [lite|simple|sim|s|S]
```

When we know where the branch we are looking for,
the tree-style toolset will be helpful.


## Find and help
These are equal to `cmds.flat <str> ...`
```
$> ticat find <str> <str> <str>
$> ticat help <str> <str> <str>
```
