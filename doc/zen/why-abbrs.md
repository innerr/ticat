# Why so many abbreviations and aliases?

**ticat** has extensive abbreviation support. Every command, argument, and environment key can have multiple aliases. Why?

## Two main reasons

1. **Reduce memorization burden**
2. **Enable long, meaningful command names**

## Reason 1: Reduce memorization

### The information challenge

With flexible ad-hoc feature assembling comes overwhelming information:
- Hundreds of commands from different repositories
- Numerous arguments per command
- Many environment keys

**ticat** reduces this pressure through:

- **Full search support** - Don't memorize, just search
- **Various info displays** - Different verbosity levels
- **Abbreviations** - Guess and still be right

### Forgiving input

Module developers are encouraged to set up:
- Common abbreviations
- Alternative spellings
- Even common misspellings

**Example**: For a command `tpcc.run` with argument `terminal`:
```bash
# All of these work:
$> ticat tpcc.run terminal=10
$> ticat tpcc.run term=10
$> ticat tpcc.run thread=10
$> ticat tpcc.run threads=10
$> ticat tpcc.run t=10
```

Users can make a rough guess and still succeed.

### Search results show aliases

```bash
$> ticat / sleep
[sleep|slp|zzz]
     'pause for duration'
    - args:
        duration|dur|d|time|t = ''
...
```

Users immediately see all valid options.

## Reason 2: Long command names

### The naming dilemma

Some commands have complex meanings requiring descriptive names:
- `cluster.kubernetes.deploy`
- `benchmark.tpch.run`
- `database.mysql.backup`

Long names are:
- **Good**: Self-documenting, clear purpose
- **Bad**: Hard to type, hard to remember

### Abbreviations solve both

With abbreviation support:
- Commands can have **long, meaningful names**
- Users can type **short, easy aliases**

```bash
# These are all the same command:
$> ticat cluster.kubernetes.deploy
$> ticat c.k.deploy
$> ticat k8s.deploy
```

### Realname always displayed

When an abbreviation is used, the real name appears:

```bash
$> ticat slp 5s
┌───────────────────┐
│ stack-level: [1]  │
├───────────────────┴────────────────────────────┐
│ >> sleep                                       │  # Real name shown
│        duration = 5s                           │
└────────────────────────────────────────────────┘
```

This reinforces learning while maintaining convenience.

## Best practices

### For module authors

```bash
# Good: Multiple sensible aliases
[args]
terminal = 10    # term|thread|threads|t|T

# Good: Common misspellings
[abbrs]
benchmrk = benchmark
ben = bench
```

### For users

```bash
# Use what you remember
$> ticat h.add user/repo    # h = hub
$> ticat e.save             # e = env
$> ticat f.+ my-flow        # f = flow

# Mix abbreviations freely
$> ticat c.f.s @ready       # cmds.flat.simple
```

## How abbreviations work

### Command abbreviations

Commands can have multiple names at any path level:

```bash
# Full command: hub.add.local
$> ticat hub.add.local path=./dir
$> ticat h.add.local path=./dir
$> ticat h.+ path=./dir      # + is alias for add-and-update
$> ticat h.+.local path=./dir
```

### Argument abbreviations

Arguments support multiple aliases:

```bash
# All equivalent:
$> ticat cmd duration=10s
$> ticat cmd dur=10s
$> ticat cmd d=10s
$> ticat cmd time=10s
$> ticat cmd t=10s
```

### Environment key abbreviations

Environment keys also support aliases:

```bash
$> ticat {sys.step=on}
$> ticat {sys.step-by-step=on}
$> ticat {sys.s=on}
```

## Summary

Abbreviations in **ticat** serve both users and developers:
- **Users**: Type less, guess safely, learn gradually
- **Developers**: Use descriptive names without burdening users
- **Everyone**: Focus on what matters, not exact syntax
