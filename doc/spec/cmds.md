# [Spec] Command tree

## Overview of command branch `cmds`
```
$> ticat cmds.tree.simple cmds
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

## Execute a command
```
## Execute a command under <root>, with different ways to pass args
$> ticat <command>

## Examples:
$> ticat dummy
$> ticat sleep

## Execute a command with path (not under <root>)
$> ticat <command-path>

## Path is composed by <segment> and "."
$> ticat <command-segment>.<command-segment>.<command-segment>

## Examples:
$> ticat dummy.power
$> ticat dummy.quiet

## Execute a command with path, in different ways to pass args
$> ticat <command> <arg-val-1> <arg-val-2>
$> ticat <command> <arg-name-1>=<arg-val-1> <arg-name-2>=<arg-val-2>
$> ticat <command> {<arg-name-1>=<arg-val-1> <arg-name-2>=<arg-val-2>}

## Examples:
$> ticat dbg.echo hello
$> ticat dbg.echo msg=hello
$> ticat dbg.echo {M=hello}

## Quoting is useful
$> ticat dbg.echo "hello world"

## Spaces(\s\t) are allowed
$> ticat dbg.echo m = hello
$> ticat dbg.echo {M = hello}
```

## List commands
```
## Display all in tree format
$> ticat cmds.tree

## Display all in tree format, only names
$> ticat cmds.tree.simple

## Display all in list format
$> ticat cmds.list
```

## Find commands
```
## Display a specific path of the command tree
$> ticat cmds.tree <path>

## Examples:
$> ticat cmds.tree dbg
$> ticat cmds.tree dbg.echo

# Find commands with any info about "echo"
$> ticat cmds.ls <find-str>
$> ticat help <find-str>
$> ticat find <find-str>
``
