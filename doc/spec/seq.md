# [Spec] ticat command sequences

This specification describes how command sequences work in **ticat**.

## Execute a sequence of commands

A command sequence executes commands one by one. The next command won't start until the previous one finishes. Commands in a sequence are separated by `:`.

```bash
$> ticat <command> : <command> : <command>

# Example:
$> ticat dummy : sleep 1s : echo hello

# Whitespace (spaces/tabs) is allowed but not necessary:
$> ticat dummy:sleep 1s:echo hello
```

## Display what will happen without executing

Use `desc` to preview a sequence:

```bash
$> ticat <command> : <command> : <command> : desc

# Examples:
$> ticat dummy : desc
$> ticat dummy : sleep 1s : echo hello : desc
```

## Execute a sequence step by step

The environment key `sys.step-by-step` enables step-by-step mode:

```bash
# Enable step-by-step
$> ticat {sys.step-by-step = true} <command> : <command> : <command>
$> ticat {sys.step-by-step = on} <command> : <command> : <command>
$> ticat {sys.step = on} <command> : <command> : <command>

# Enable only for <command-2>, to ask for confirmation
$> ticat <command-1> : {sys.step = on} <command-2> : <command-3>
```

### Built-in step-by-step commands

A set of built-in commands provides easier access to this feature:

```bash
# Find step-by-step commands
$> ticat cmds.tree dbg.step
[step-by-step|step|s|S]
    - full-cmd:
        dbg.step-by-step
    - full-abbrs:
        dbg.step-by-step|step|s|S
    [on|yes|y|Y|1|+]
         'enable step by step'
        - full-cmd:
            dbg.step-by-step.on
        - full-abbrs:
            dbg.step-by-step|step|s|S.on|yes|y|Y|1|+
        - cmd-type:
            normal (quiet)
        - from:
            builtin
    [off|no|n|N|0|-]
         'disable step by step'
        - full-cmd:
            dbg.step-by-step.off
        - full-abbrs:
            dbg.step-by-step|step|s|S.off|no|n|N|0|-
        - cmd-type:
            normal (quiet)
        - from:
            builtin

# Enable step-by-step for a sequence
$> ticat dbg.step.on : <command> : <command> : <command>

# Enable in the middle of a sequence
$> ticat <command> : <command> : dbg.step.on : <command>

# Enable and save - all future executions will need confirmation
$> ticat dbg.step.on : env.save
```

## The "desc" command branch

### Overview

```bash
$> ticat cmds.tree.simple desc
[desc]
     'desc the flow about to execute'
    [simple]
         'desc the flow about to execute in lite style'
    [skeleton]
         'desc the flow about to execute, skeleton only'
    [dependencies]
         'list the depended os-commands of the flow'
    [env-ops-check]
         'desc the env-ops check result of the flow'
    [flow]
         'desc the flow execution'
        [simple]
             'desc the flow execution in lite style'
```

### Examples

```bash
$> ticat <command> : <command> : <command> : desc

# Examples:
$> ticat dummy : desc
$> ticat dummy : sleep 1s : echo hello : desc
```

## Power and priority commands

Some commands have special flags that affect sequence execution:

### Power commands

Power commands can modify the sequence. Check a command's type with `cmds.list` or `cmds.tree`:

```bash
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

### The "desc" command flags

The `desc` command has three flags:
- **quiet**: Doesn't display in the executing sequence (the boxes)
- **priority**: Runs first, before other commands
- **power**: Can modify the sequence about to execute

```bash
# The command type of "desc"
$> ticat cmds.tree desc
[desc|d|D]
     'desc the flow about to execute'
    - cmd-type:
        power (quiet) (priority)

# Usage
$> ticat <command> : <command> : <command> : desc
# Actual execute order
$> ticat desc : <command> : <command> : <command>
# Actual execution: "desc" removes all commands after displaying sequence info
$> ticat desc [: <command> : <command> : <command>]
```

### Other power commands

```bash
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
    - cmd-type:
        power (quiet) (priority)
    - from:
        builtin
...
```

### Multiple priority commands

When there's more than one priority command in a sequence:

```bash
# User input
$> ticat <command-1> : <command-2> : <priority-command-a> : <priority-command-b>

# Actual execute order:
$> ticat <priority-command-a> : <priority-command-b> : <command-1> : <command-2>
```

Priority commands maintain their relative order from the original input.
