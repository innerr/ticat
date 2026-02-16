# Commands in ticat

This guide explains how to work with **ticat**'s command system, including navigating the command tree, searching for commands, and understanding command properties.

## Dealing with large numbers of commands

### The command tree

Every command in **ticat** is identified by a path. Here are some examples:
- `cmds`
- `hub.add`
- `dummy`
- `dummy.power`
- `dummy.quiet`

**Important**: `dummy`, `dummy.power`, and `dummy.quiet` are unrelated commands. They share the same branch only because the author organized them that way.

These two commands do completely different things:
```bash
$> ticat dummy
$> ticat dummy.power
```

### Why are commands in tree form?

**ticat** is a platform with a potentially huge number of commands from different authors. The tree form provides a namespace mechanism to avoid command name conflicts.

### User tip: memorize nothing, just search

Looking through the command tree to find what you need is not recommended. Searching by keywords is much better.

Built-in commands can search across all command properties: command name, help string, argument names, environment operations, etc.

Sometimes users might not know which keywords to use. To solve this, **ticat** introduces **tags**: authors add tags to commands, and users search by tags.

### Tags and global searching

Most commands come from git repositories added via `hub.add`. We recommend searching with `/` when you add a new repo:

```bash
# Find ready-to-go commands from a specific repo
$> ticat / <repo-name> @ready
```

If there are too many results, add more keywords to narrow it down:
```bash
$> ticat / <repo-name> @ready <keyword> <keyword> ...
```

To search across all repos (omit the repo name):
```bash
$> ticat / <keyword> <keyword> <keyword>...
```

You can use up to 4 keywords per search.

**Conventional tags**:
- `@selftest`: Commands for self-testing the repository
- `@ready`: Ready-to-go (out-of-the-box) commands

For other tags, authors will explain their meanings in the repository's README.

### The `-` and `+` commands

`+` and `-` are important commands for finding and displaying information. They have the same usage, but different verbosity levels:
- `-` shows brief information (similar to `less`)
- `+` shows detailed information (similar to `more`)

View their help:
```bash
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

```bash
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

`+` and `-` are convenient shortcuts that we'll explore throughout this guide.

## Use command branch `cmds` to get command information

### Overview of branch `cmds`

`cmds` is a command toolset registered on this branch, and it's also a single command itself:

```bash
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

`c` is short for `cmds`. Use it to display a command's information:

```bash
$> ticat cmds dbg.echo
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

You can use `+` or `-` with `cmds`:
```bash
$> ticat c:+
```

## Show the command tree or branch

The command `cmds.tree` displays a command's info and its sub-tree.

Short name: `c.t`:
```bash
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

`cmds.tree.simple` displays only command names, convenient for observing the tree structure.

Short name: `c.t.s`:
```bash
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

**Recommendation**: Use `-` instead of `cmds.tree.simple`:
```bash
$> ticat c:-
```

Tree-style commands are helpful when you know which branch you're looking for.

## List and filter commands in flat mode

`cmds.flat` displays all commands in flat mode. Aliases: `flat`, `ls`, `f`:
```bash
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

Using `cmds.flat` without arguments is rarely useful because the output is very long.

### Filter by strings

You can filter using up to 4 search strings:

```bash
$> ticat c.f <str>
$> ticat c.f <str> <str> ...
```

**Example 1**: Find commands with tag `@selftest`:
```bash
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

**Example 2**: Find commands from a specific repository:
```bash
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

**Example 3**: Find commands that write to a specific environment key:
```bash
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

You can use `+` instead of `cmds.flat` for global searching.

### Quick start for new repositories

When you add a repository to **ticat**, the fastest way to get started:
1. Find ready-to-go commands: `ticat + <repo-address> @ready`
2. Use `desc` to check what environment keys these commands need
3. Find commands that provide those keys (or manually set them with `{k=v}`)

### Use lite flat mode for better searching

`cmds.flat.simple` shows less information. Short name: `c.f.s`:
```bash
$> ticat c.f.s mysql
[exec|exe|e|E]
    - full-cmd:
        mysql.exec
    - full-abbrs:
        mysql|my.exec|exe|e|E
...
```

`-` can replace `cmds.flat.simple`, just as `+` can replace `cmds.flat`.

### Search under a specific branch

`+` and `-` are the only commands for searching under a specific branch:

```bash
# Search "tree" in branch `cmds`
$> ticat cmds:- tree
[cmds]
     'display cmd info, sub tree cmds will not show'
[cmds.tree]
     'list builtin and loaded cmds'
```

Use `+` instead of `-` for more detail:
```bash
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

### Advanced search operators

Use `=` and `$` if `+` or `-` don't give what you want:

```bash
$> ticat cmds $
[tail-sub|$]
     'display commands on the branch of the last command'

$> ticat cmds =
[tail-info|=]
     'display the last command info, sub tree commands will not show'
```
