# Manipulate environment key-values

The **environment** (env) is a set of key-values shared by all modules during an execution session. Understanding how to manage environment variables is crucial for using **ticat** effectively.

For users, it's important to find out what environment keys a command (module or flow) requires.

## Display a command's environment operations (read/write)

### For modules

Use `cmds` to show a module's details, including environment operations:

```bash
$> ticat cmds bench.run
[run]
     'pretend to run bench'
    - full-cmd:
        bench.run
    - env-ops:
        cluster.host = may-read
        cluster.port = read
        bench.scale = may-read
        bench.begin = write
        bench.end = write
...
```

### For flows

Use `desc` to show a flow's unsatisfied environment operations:

```bash
$> ticat bench : desc
--->>>
[bench]
...
<<<---

-------=<unsatisfied env read>=-------

<FATAL> 'cluster.port'
       - read by:
            [bench.load]
            [bench.run]
       - but not provided
```

An environment key-value being read before write will cause a `FATAL` error. A `risk` warning is normally acceptable.

**Tip**: Use `+` instead of `cmds` or `desc` for either commands or flows:

```bash
$> ticat bench.run:+
$> ticat bench:+
```

## Set environment key-values

Modules are recommended to read from the environment instead of reading from arguments whenever possible. This enables automatic assembly.

### Basic syntax

Use curly braces `{` `}` to set values during execution:

```bash
# Set a single value
$> ticat {display.width=40}
```

**Note**: It's OK if the key doesn't already exist in the environment.

### Verify values

Use `env.ls` (or `e.ls`) to display the modified value:

```bash
$> ticat {display.width=40} : env.ls display.width
```

### Set multiple values

Set multiple key-values in one pair of brackets:

```bash
$> ticat {display.width=40 display.style=utf8} : env.ls display
```

### Set values inline

You can also set values inline within command paths:

```bash
$> ticat env{mykey=666}.ls mykey
env.mykey = 666
```

### Whitespace handling

Extra space characters (space and tab) are ignored:

```bash
$> ticat {display.width = 40}
```

## Save environment key-values

By saving key-values to the environment, you don't need to type them every time.

Use `env.save` (short name: `e.+`) to persist changes:

```bash
# Set and save
$> ticat {display.width=40} env.save

# Verify
$> ticat env.ls width
display.width = 40

# Update and save again
$> ticat {display.width=60} env.save
display.width = 60
```

## Observe environment key-values during execution

### Display box

During execution, the info box shows current environment key-values in the upper part:

```bash
$> ticat {foo=bar} sleep 3m : dummy : dummy
┌───────────────────┐
│ stack-level: [1]  │             05-31 20:07:39
├───────────────────┴────────────────────────────┐
│    foo = bar                                   │
├────────────────────────────────────────────────┤
│ >> sleep                                       │
│        duration = 3m                           │
│    dummy                                       │
│    dummy                                       │
└────────────────────────────────────────────────┘
...
```

### Verbosity control

There are many hidden key-values. Use the `verb` command to show them:

```bash
$> ticat verb : dummy : dummy
```

`verb` shows maximum information. Short name: `v`.

For moderate verbosity, use `v.+`:

```bash
$> ticat v.+ : dummy : dummy
$> ticat v.+ 1 : dummy : dummy
$> ticat v.+ 2 : dummy : dummy
```

### Quiet mode

Use `quiet` (short name: `q`) to completely hide execution info:

```bash
$> ticat q : dummy : dummy
```

## Display environment key-values

### List all

Use `env.flat` to list all environment key-values. Short name: `e.f`:

```bash
$> ticat e.f
```

### Find/filter

```bash
# Find by single string
$> ticat e.f <find-str>

# Find by multiple strings (up to 3)
$> ticat e.f <find-str> <find-str>
$> ticat e.f <find-str> <find-str> <find-str>
```

### Example: Find display settings

The `v`, `q`, and `v.+` commands alter display behavior values. Find them with:

```bash
$> ticat e.f display
display.env = true
display.env.default = false
display.env.display = false
display.env.layer = false
display.env.sys = false
display.env.sys.paths = false
display.executor = true
display.executor.end = false
display.max-cmd-cnt = 7
display.meow = false
display.mod.quiet = false
display.mod.realname = true
display.one-cmd = false
display.style = utf8
display.utf8 = true
display.width = 40
...
```

## Environment layers

The environment has multiple layers. When getting a value, **ticat** searches from the first layer down:

```
  command layer    - the first layer
  session layer
  persisted layer
  default layer    - the last layer
```

### Layer meanings

- **command layer**: Key-values for the current command only
- **session layer**: Key-values for the entire sequence
- **persisted layer**: Key-values saved via `env.save`
- **default layer**: Hard-coded default values

### Display layers

```bash
# Display all layers
$> ticat env.tree

# Display layers during sequence execution
$> ticat {display.layer=true} dummy : dummy
$> ticat {display.layer=true} dummy : {example-key=its-command-layer} dummy

# Save the display flag
$> ticat {display.layer=true} env.save
# Now all sequence executions will display layers
$> ticat dummy: dummy
```

## Command layer vs. session layer

### Command layer (with `:` prefix)

If key-value settings have `:` in front, they're in the command layer:

```bash
# This only affects the 2nd dummy command
$> ticat dummy : {example-key=its-command-layer} dummy
```

### Session layer (without `:` prefix)

Without the `:` prefix, values go to the session layer:

```bash
# This affects all commands
$> ticat {example-key=its-session-layer} dummy : dummy
```

### Demonstration

```bash
# Session layer - affects both
$> ticat {display.width=40} dummy : dummy

# Command layer - only affects second
$> ticat dummy : {display.width=40} dummy
```

### Module code changes

If a command changes environment in its code (not via CLI), the changes go to the session layer:

```bash
# The "display.width" value for <command-2> will be the changed value
$> ticat <command-1-which-changes-display-width> : <command-2>
```

## Best practices

1. **Save frequently used values**: Use `env.save` for values you use often
2. **Use command layer for one-time changes**: Use `:` prefix for temporary overrides
3. **Document required keys**: If writing modules, clearly document what environment keys are needed
4. **Check before running**: Use `+` to verify environment requirements before executing flows
