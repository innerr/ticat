# Commands in ticat

## Deal with huge number of commands

### The command tree

Any command in **ticat** are identitied by a path, here are some examples:
* `cmds`
* `hub.add`
* `dummy`
* `dummy.power`
* `dummy.quiet`
Notice that `dummy` `dummy.power` and `dummy.quiet` are nothing related,
they share the same branch only because the author put them that way.

These two callings do totally different things:
```
$> ticat dummy
$> ticat dummy.power
```

### Why commands are in tree form?

**Ticat** is a platform with huge amount of commands from different authors,
the tree form provides a namespace mechanism to avoid command name conflictings.

### User tip: memorize nothing, just search

Looking through the command tree to find what we want is not sugguested,
searching by keywords is a better way.

Some builtin commands can do a excellent job in searching,
any properties in a command could match by keywords: command name, help string, arg names, env ops, anyting.

However, sometimes users could still not sure what keywords they should type down.
to solve this issue we introduce tags: authors put tags on commands, users search commands by tags.

### Tags and global searching

There are a few builtin commands, but most of them are provided by git repos,
we add repos to **ticat** by `hub.add`, then we have new commands.

We have some conventional tags have specific meanings:
* `@selftest`: indicate that this command are for self-testing of this repo.
* `@ready`: indicate that this is a "ready-to-go"(out-of-the-box) command.

For other tags, authors will explain them in the repo `README`.

So when we added a repo, the first thing will be find out what we got by searching with `-`:
```
$> ticat - <repo-name> @ready
```

The result might be a lot, adding more keywords could screen out what we need.
```
$> ticat - <repo-name> @ready <keyword> <keyword> ...
```

If the command source is irrelevant, remove it from the keyword list(so does for tags):
```
$> ticat - <keyword> <keyworkd> <keyword>...
```
The keyword number is up to 4, sould be enough

### The `-` and `+`

We show how to use `-` to do search above, it only shows names and helps of commands.
Use `+` instead of `-` will shows all details.

`+` and `-` are important commands to find and display infos, they have the same usage.
The difference is `-` shows brief messages, and `+` shows rich infos.

`+` is a short name for command `more`, `-` is short for `less`. Command `cmds` will show a command's detail:
```
$> ticat cmds +
[more|+]
     'display rich info base on:
      * if in a sequence having
          * more than 1 other commands: show the sequence execution.
          * only 1 other command and
              * has no args and the other command is
                  * a flow: show the flow execution.
                  * not a flow: show the command or the branch info.
              * has args: find commands under the branch of the other command.
      * if not in a sequence and
          * has args: do global search.
          * has no args: show global help.'
    - args:
        1st-str|find-str = ''
        2nd-str = ''
        3rh-str = ''
        4th-str = ''
    - cmd-type:
        power (quiet) (priority)
```
```
$> ticat cmds -
[less|-]
     'display brief info base on:
      * if in a sequence having
          * more than 1 other commands: show the sequence execution.
          * only 1 other command and
              * has no args and the other command is
                  * a flow: show the flow execution.
                  * not a flow: show the command or the branch info.
              * has args: find commands under the branch of the other command.
      * if not in a sequence and
          * has args: do global search.
          * has no args: show global help.'
    - args:
        1st-str|find-str = ''
        2nd-str = ''
        3rh-str = ''
        4th-str = ''
    - cmd-type:
        power (quiet) (priority)
```

`+` `-` are convenient shortcuts, we will learn their usage during introducing the formal commands.

## Use command branch `cmds` to get command infos

### Overview of branch `cmds`

`cmds` is a command toolset registered on this branch,
it is also a single command itself.

The content of `cmds` branch:
```
$> ticat cmds:-
[cmds]
     'display cmd info, sub tree cmds will not show'
    [tree]
         'list builtin and loaded cmds'
        [simple]
             'list builtin and loaded cmds, skeleton only'
    [list]
         'list builtin and loaded cmds'
        [simple]
             'list builtin and loaded cmds in lite style'
```

### Get single command's info

`c` is short for `cmds`, will display the info for a command
```
$> ticat cmds dbg.ccho
[echo]
     'print message from argv'
    - full-cmd:
        dbg.echo
    - args:
        message|msg|m|M = ''

$> ticat c desc
[desc|d|D]
     'desc the flow about to execute'
    - cmd-type:
        power (quiet) (priority)

$> ticat c cmds
[cmds|cmd|c|C]
     'display cmd info, no sub tree'
    - args:
        cmd-path|path|p|P = ''

$> ticat c c
[cmds|cmd|c|C]
     'display cmd info, no sub tree'
    - args:
        cmd-path|path|p|P = ''
```

`+` `-` could use for `cmds` by concate them to the end of the command sequence:
```
$> ticat c:+
```

## Show the command tree or branch

The command `cmds.tree` will display the command's info and it's sub-tree's.

It has a short name `c.t`:
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

`cmds.tree.simple` display only the command names, convenient to observe the tree struct.

It has short name `c.t.s`:
```
$> ticat c.t.s c
[cmds]
     'display cmd info, sub tree cmds will not show'
    [tree]
         'list builtin and loaded cmds'
        [simple]
             'list builtin and loaded cmds, skeleton only'
    [list]
         'list builtin and loaded cmds'
        [simple]
             'list builtin and loaded cmds in lite style'
```

It's highly recommended to use `-` instead of `cmds.tree.simple`,
below command could get the same result:
```
$> ticat c:-
```

When we know where the branch we are looking for,
the tree-style toolset will be helpful.

## List and filter commands in flat mode

`cmds.flat` is for display all commands in flat mode, `flat` `ls` and `f` are equal:
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

Using `cmds.flat` without args are barely useless because the output is way too long.
`grep` might helps but we have better way to looking for commands.

Filter(find) by strings, up to 4 strings, the finding strings could be anything:
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

As in the quick-start doc, `+` could instead of `cmds.flat` for global searching.

When we just added a repo to **ticat**,
The fast way to put it into work is:
* Find what's in it with tag `@ready` by searsh: `ticat + <git-address> @ready`.
* Use `desc` to check out what env keys these ready-to-go commands need.
* Find what commands provide those env keys (or manually set to env by `{k=v}`).

### Use lite flat mode for better searching

`cmds.flat.simple` is similar to `cmds.flat`(c.f), just show less info.

It has short name `c.f.s`:
```
$> ticat c.f.s mysql
[exec|exe|e|E]
    - full-cmd:
        mysql.exec
    - full-abbrs:
        mysql|my.exec|exe|e|E
...
```

`-` could instead of `cmds.flat.simple`, just as `+` could instead of `cmds.flat`.

### Search under a branch

`+` `-` are the only commands for searching commands under a specific branch.

Search "tree"(could be any string) in the branch `cmds`:
```
$> ticat cmds:- tree
[cmds]
     'display cmd info, sub tree cmds will not show'
[cmds.tree]
     'list builtin and loaded cmds'
```

Use `+` instead of `-` to get more detail:
```
$> ticat cmds:+ tree
[cmds|cmd|c|C]
     'display cmd info, no sub tree'
    - args:
        cmd-path|path|p|P = ''
[tree|t|T]
     'list builtin and loaded cmds'
    - full-cmd:
        cmds.tree
    - full-abbrs:
        cmds|cmd|c|C.tree|t|T
    - args:
        cmd-path|path|p|P = ''
```

Use `=` and `$` if the result of `+` or `-` is not what we want.
```
$> ticat cmds $
[tail-sub|$]
     'display commands on the branch of the last command'

$> ticat cmds =
[tail-info|=]
     'display the last command info, sub tree commands will not show'
```
