# Basic usage of ticat: build, run commands

This guide covers the fundamental operations of **ticat**, including building from source and running commands.

## Build

### Prerequisites

- Go 1.16 or later
- Git

### Installation steps

Golang is needed to build ticat:

```bash
# Clone the repository
$> git clone https://github.com/innerr/ticat
$> cd ticat

# Build the project
$> make
```

**Recommendation**: Add `ticat/bin` to your system `$PATH` for convenient access:

```bash
# Add to your shell profile (~/.bashrc, ~/.zshrc, etc.)
export PATH="/path/to/ticat/bin:$PATH"
```

## Run a command

### Simple commands

Run a simple command that prints a message:

```bash
$> ticat dummy
dummy cmd here
```

### Commands with arguments

Pass arguments to a command. For example, `sleep` will pause for the specified duration:

```bash
$> ticat sleep 20s
.zzZZ .................... *\O/*
```

### Using abbreviations/aliases

**ticat** supports abbreviations and aliases to save typing:

```bash
# All of these are equivalent
$> ticat slp 3s
$> ticat slp dur=3s
$> ticat slp d=3s
```

## Run a command in the command-tree

### Understanding command organization

All commands are organized into a tree structure. The `sleep` and `dummy` commands are under the `root`, so we can call them directly.

Some commands share the same branch for better organization:

```bash
$> ticat dummy.power
power dummy cmd here

$> ticat dummy.quiet
quiet dummy cmd here
```

**Important**: `dummy`, `dummy.power`, and `dummy.quiet` are completely different commands. They're in the same command-branch only because it makes it easier for users to find related commands.

### Display command information

Use `==` to display detailed information about a command:

```bash
$> ticat dbg.echo :==
[echo]
     'print message from argv'
    - full-cmd:
        dbg.echo
    - args:
        message|msg|m|M = ''
```

This shows that `dbg.echo` has an argument called `message`, which has abbreviations: `msg`, `m`, `M`.

### Different ways to call commands

You can pass arguments in multiple ways:

```bash
# Positional argument
$> ticat dbg.echo hello

# Quoted argument for strings with spaces
$> ticat dbg.echo "hello world"

# Named argument
$> ticat dbg.echo msg=hello

# Named argument with spaces
$> ticat dbg.echo m = hello

# Using braces
$> ticat dbg.echo {M=hello}

# Using braces with spaces
$> ticat dbg.echo {M = hello}
```

## Use abbreviations/aliases

### Understanding the abbreviation syntax

When searching commands, the output shows available abbreviations:

```
...
    [hub|h|H]
        [clear|reset]
             'remove all repos from hub'
            - full-cmd:
                hub.clear
            - full-abbrs:
                hub|h|H.clear|reset
...
```

This tells us:
- The command name `hub` has abbreviations: `h`, `H`
- The subcommand `clear` has an alias: `reset`

### Using abbreviations

All of these call the same command:

```bash
$> ticat hub.clear
$> ticat h.reset
$> ticat H.clear
```

## Next steps

- Learn about [Hub: getting modules and flows from others](./hub.md)
- Understand [environment key-values](./env.md)
- Explore [flows and command sequences](./flow.md)
- Discover [searching and filtering commands](./cmds.md)
