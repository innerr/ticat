# [Spec] Help

## The helping commands (commands provide info)
```
## Sequence info
$> ticat <command> : <command> : desc

## Hub info
$> ticat hub.list

## Flow info
$> ticat flow.list

## Command info
$> ticat cmds.list
$> ticat cmds.tree
$> ticat cmds.tree.simple

## Env info
$> ticat env.list
$> ticat env.tree

## Abbrs info
$> ticat env.abbrs

## Find command, env
## ..and abbrs(TODO: implement)
$> ticat find <find-str> <find-str> ...
```

## The commands `+` and `-`
These two are shortcuts, could display info base on the situation:
```
$> ticat cmd -
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
        5th-str = ''
        6th-str = ''
    - cmd-type:
        power (quiet) (priority)
    - from:
        builtin

$> ticat cmd +
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
        5th-str = ''
        6th-str = ''
    - cmd-type:
        power (quiet) (priority)
    - from:
        builtin
```

Examples:
```
## Global help, tips for how to use ticat
$> ticat +
$> ticat -

## Use them to find commands
$> ticat - <find-str> <find-str> ...
$> ticat + <find-str> <find-str> ...

## Other usages:
(TODO: doc, most are already in usage doc)
```
