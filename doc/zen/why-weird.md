# Why does the usage seem weird?

## The short answer

The core syntax is actually quite simple:
- Run command sequences in unix-pipe style, just use `:` instead of `|`
- Use `{key=value}` to set values in the shared environment

That's all. The rest are optional features you can ignore if you prefer.

## The "weird" parts explained

Some aspects might feel unusual at first. Here's why they exist:

### Power commands

This command might seem weird:

```bash
$> ticat desc : dummy : dummy
--->>>
[dummy]
     'dummy cmd for testing'
[dummy]
     'dummy cmd for testing'
<<<---
```

From the usage guide, we know `desc` displays how a sequence will execute. But in unix-pipe style, shouldn't `dummy` run? Why doesn't it?

**The reason**: `desc` is a builtin "power" command. Power commands can access and modify the sequence. `desc` shows the info and then removes all commands, so nothing actually runs.

This mechanism allows **ticat** to add powerful features to the executor. For example, you could write a "mock" command that replaces specific commands with mock implementations.

### Priority commands

These two commands produce the same result:

```bash
$> ticat desc : dummy : dummy
$> ticat dummy : dummy : desc
```

**The reason**: `desc` has a "priority" flag. Priority commands move to the front of the sequence automatically.

If there are multiple priority commands, they maintain their relative order.

This might seem nonsensical at first, but it's actually quite handy:

```bash
# Suppose you're typing a sequence:
$> ticat dummy:dummy:dummy:dummy:dummy:dummy
```

You want to do a preflight check. Without priority:

1. Move cursor to the front
2. Type "desc:"
3. Execute
4. Roll up history, edit, re-execute

With priority:

1. Append ":desc"
2. Execute
3. Roll up history, delete ":desc", execute

In terminals where `$PS1` isn't properly set, or in WSL, history editing can cause display issues. Priority commands make this much easier.

### The `+` and `-` shortcuts

**ticat** has lots of information to display. Different commands like `cmds.list` and `desc` show different levels of detail.

We needed a single command to cover all frequent info queries. First, we created `help` (abbr `?`), but:
- One command wasn't enough - either too much or too little info
- `?` is intercepted by some shells (like zsh)

So we created `more` and `less`, with abbreviations `+` and `-`.

**How they work**:

```
$> ticat +:-
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
    - from:
        builtin
```

**Using them together gives an excellent browsing experience**:

Global searching:
```bash
# Find commands
$> ticat - <find-str>

# Filter results
$> ticat - <find-str> <find-str> <find-str>

# Switch to "+" for details
$> ticat + <find-str> <find-str> <find-str>
```

Investigate a branch:
```bash
# Brief view
$> ticat <command> :-

# Detailed view
$> ticat <command> :+
```

Preflight check:
```bash
# Focus on what the flow does
$> ticat <flow> :-

# Focus on dependency report
$> ticat <flow> :+
```

In-place searching while writing a sequence:
```bash
# Search for what command to use next
$> ticat <command>:<command> :- <find-str>

# Continue filtering
$> ticat <command>:<command> :- <find-str> <find-str>

# Show details
$> ticat <command>:<command> :+ <find-str> <find-str>

# Add the found command
$> ticat <command>:<command>:<new-command>
```

### Advanced tail editing

Sometimes `+` and `-` aren't enough. For example:
```bash
$> ticat <flow-command> :-
$> ticat <flow-command> :+
$> ticat <cmd>:<cmd>:<cmd>:<cmd> :+
```

In these cases, `+` and `-` are occupied by describing flow execution.

We tentatively use `=` for "show the last command's info". This command is less stable and may change in the future.

## The golden rule

To avoid confusion, remember this simple rule:

**All command-related operations support tail editing. Others do not.**

Command-related operations include:
- Searching commands
- Checking command details
- Concatenating commands into flows
- Checking flow details

## You can ignore the weirdness

Remember: **you can ignore all the "weird" features**. The core functionality works perfectly with just:
- `:` for command sequences
- `{key=value}` for environment settings

Everything else is optional convenience.
