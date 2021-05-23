# [Spec] Ticat command sequences

## Execute a sequence of command
A command sequence will execute commands one by one,
the latter one won't start untill the previous one finishes.
Commands in a sequence are seperated by ":".
```bash
$> ticat <command> : <command> : <command>
## Example:
$> ticat dummy : sleep 1s : echo hello

## Spaces(\s\t) are allowed but not necessary:
$> ticat dummy:sleep 1s:echo hello
```

## Display what will happen without execute a sequence
```bash
$> ticat <command> : <command> : <command> : desc
## Exmaples:
$> ticat dummy : desc
$> ticat dummy : sleep 1s : echo hello : desc
```

## Execute a sequence step by step
The env key "sys.step-by-step" enable or disable the step-by-step feature:
```bash
$> ticat {sys.step-by-step = true} <command> : <command> : <command>
$> ticat {sys.step-by-step = on} <command> : <command> : <command>
$> ticat {sys.step = on} <command> : <command> : <command>

## Enable it only for <command-2>, to ask for confirmation from user
$> ticat <command-1> : {sys.step = on} <command-2> : <command-3>
```

A set of builtin commands could changes this env key for better usage:
```bash
## Find these two commands:
$> ticat m.ls step
[on|yes|y|Y|1]
     'enable step by step'
    - full-cmd:
        dbg.step-by-step.on
    - full-abbrs:
        dbg.step-by-step|step|s|S.on|yes|y|Y|1
    - cmd-type:
        normal (quiet)
[off|no|n|N|0]
     'disable step by step'
    - full-cmd:
        dbg.step-by-step.off
    - full-abbrs:
        dbg.step-by-step|step|s|S.off|no|n|N|0
    - cmd-type:
        normal (quiet)

## Use these commands:
$> ticat dbg.step.on : <command> : <command> : <command>
## Enable step-by-step in the middle
$> ticat <command> : <command> : dbg.step.on : <command>

## Enable and save, after this all executions will need confirming
$> ticat dbg.step.on : env.save
```

## The "desc" command
```bash
$> ticat <command> : <command> : <command> : desc

## Exmaples:
$> ticat dummy : desc
$> ticat dummy : sleep 1s : echo hello : desc
```

## Power/priority commands
Some commands have "power" flag, these type of command can changes the sequence.
Use "cmds.list <path>" or "cmds.tree <path>" can check a command's type.
```bash
## Example:
$> ticat cmds.tree dummy.power
[power|p|P]
     'power dummy cmd for testing'
    - full-cmd:
        dummy.power
    - full-abbrs:
        dummy|dmy|dm.power|p|P
    - cmd-type:
        power
```

The "desc" command have 3 flags:
* quiet: it would display in the executing sequence(the boxes)
* priority: it got to run first, then others could be executed.
* power: it can change the sequence about to execute.
```
## The command type of "desc"
$> ticat cmds.tree desc
[desc|d|D]
     'desc the flow about to execute'
    - cmd-type:
        power (quiet) (priority)

## The usage of "desc"
$> ticat <command> : <command> : <command> : desc
## The actual execute order
$> ticat desc : <command> : <command> : <command>
## The actual execution: "desc" remove all the commands after display the sequence's info
$> ticat desc [: <command> : <command> : <command>]
```

Other power commmands:
```bash
$> ticat cmds.tree help
[help|?]
     'get help'
    - cmd-type:
        power (priority)
    - args:
        1st-str|1|find|str|s|S = ''
        2rd-str|2 = ''
(and more)
```

When have more than one priority commands in a sequence:
```bash
## User input
$> ticat <command-1> : <command-2> : <priority-command-a> : <priority-command-b>
## Actual execute order:
$> ticat <priority-command-a> : <priority-command-b> : <command-1> : <command-2>
```
