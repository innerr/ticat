# Flow: assemble pieces into powerful workflows

Flows are one of **ticat**'s most powerful features. They allow you to combine simple commands into complex, reusable workflows.

## Command sequences and flows

### Command sequences

We're familiar with using commands in **ticat**. It's similar to unix-pipe style, but with differences:
- Use `:` to concatenate commands (not `|`)
- Commands execute sequentially - the second one won't start until the previous finishes
- Execution info displays in a box, with `>>` indicating the current command

Example of a command sequence:
```
┌───────────────────┐
│ stack-level: [1]  │             06-01 17:07:41
├───────────────────┴────────────────────────────┐
│ >> dummy                                       │
│    sleep                                       │
│        duration = 3s                           │
│    dummy                                       │
└────────────────────────────────────────────────┘
dummy cmd here
┌───────────────────┐
│ stack-level: [1]  │             06-01 17:07:41
├───────────────────┴────────────────────────────┐
│    dummy                                       │
│ >> sleep                                       │
│        duration = 3s                           │
│    dummy                                       │
└────────────────────────────────────────────────┘
.zzZZ ... *\O/*
┌───────────────────┐
│ stack-level: [1]  │             06-01 17:07:44
├───────────────────┴────────────────────────────┐
│    dummy                                       │
│    sleep                                       │
│        duration = 3s                           │
│ >> dummy                                       │
└────────────────────────────────────────────────┘
dummy cmd here
```
The boxes indicate the running command with `>>`.

### Command sequence == `flow`

We use the name `flow` to refer to these sequences. Flows can be:
- **Persisted** to local disk for reuse
- Called like regular commands
- **Nested** - flows can call other flows

The `flow` command branch manages saved flows, and the `desc` branch displays flow execution plans. `+` and `-` are shortcuts for the most common operations.

## Save, call, edit, or remove flows

### Command branch `flow` overview

```bash
$> ticat flow:-
[flow]
     'list local saved but unlinked (to any repo) flows'
    [save]
         'save current cmds as a flow'
    [set-help-str]
         'set help str to a saved flow'
    [remove]
         'remove a saved flow'
    [clear]
         'remove all flows saved in local'
    [move-flows-to-dir]
         'move all saved flows to a local dir (could be a git repo).
          auto move:
              * if one(and only one) local dir exists in hub
              * and the arg "path" is empty
              then flows will move to that dir'
```

### Save a command sequence as a flow

Use `flow.save` to persist a sequence. Alias: `f.+`

```bash
$> ticat dummy : dummy : dummy : f.+ x
```

If flow "x" already exists, you'll be asked to confirm overwriting.

Save with a nested path:
```bash
$> ticat x : x : flow.save aa.bb.cc
```

### Run a saved flow

A flow is a regular command - all command rules apply:

```bash
# Run a simple flow
$> ticat x
(execute dummy * 3)

# Run a nested flow
$> ticat aa.bb.cc
(execute dummy * 6)
```

### List all saved flows

The `flow` command (also a branch) shows all saved flows. Alias: `f`

```bash
$> ticat f
[aa.bb.cc]
    - flow:
        dummy : dummy : dummy
    - executable:
        ...
[x]
    - flow:
        dummy : dummy : dummy
    - executable:
        ...
```

The output shows flow file paths - you can manually edit them if needed.

### Remove saved flows

Remove a single flow with `flow.remove`. Alias: `f.-`

```bash
$> ticat f.- aa.bb.cc
```

Remove all flows with `flow.clear`. Alias: `f.--`

```bash
$> ticat f.--
```

## Share saved flows

### Add a help string to a saved flow

When you have many saved flows (or before sharing them), it's helpful to add descriptive help strings.

Use `flow.set-help-str` to set a help string. Aliases: `f.help`, `f.h`

```bash
# Save a flow
$> ticat dummy : dummy : dummy : f.+ x

# Set help string
$> ticat f.h x 'power test'

# Show the help string
$> ticat c x
[x]
     'power test'
    - from:
        ...
    - flow:
        dummy : dummy : dummy
```

### Share saved flows with others

To share saved flows:
1. Move the saved files from **ticat**'s storage directory to a specific directory
2. Push that directory to a git repository

**Recommended locations** in your repository:
- Root directory
- `flows/` subdirectory

(Directory scanning can be slow, so **ticat** may only scan specific directories in the future.)

Use `flow.move-flows-to-dir` to relocate saved flow files. Alias: `f.mv`

```bash
$> ticat f.mv path=./tmp
```

### Advanced flow file moving

If one (and only one) local directory exists in the hub, and the `path` argument is empty, flows will automatically move to that directory.

**Note**: If the destination directory isn't in the hub, you won't be able to call those moved flows after moving.

You can also manually move the files. They're stored in a directory defined by the environment key `sys.paths.flows`:

```bash
$> ticat env.flat flow path
sys.paths.flows = (a local dir)
```

## Dig into a flow

### Display properties of a flow

Use `cmds` to show a flow's detailed properties (same as other commands). Alias: `c`

```bash
# Save a flow
$> ticat dummy : dummy : dummy : f.+ x

# Show flow info
$> ticat c x
[x]
    - flow:
        dummy : dummy : dummy
...
```

### Display how a flow will execute

Use `desc` and `desc.simple` to preview a flow's execution without running it. Aliases: `d`, `d.s`

```bash
# Full description
$> ticat dummy : dummy : dummy : desc
$> ticat x : desc
$> ticat x : d

# Full description, with less module info
$> ticat x : desc.simple
$> ticat x : d.s
```

The `desc` commands display:
1. Full execution description
2. OS command dependency check
3. Environment operation check

**Example OS command dependency report**:
```
-------=<depended os commands>=-------

[tiup]
        'to verify cluster name'
            [tidb.link]
        'to destroy cluster'
            [tidb.destroy]
        'to deploy cluster'
            [tidb.deploy]
        'to start cluster'
            [tidb.start]
        'to display tidb cluster info'
            [mysql.link.tidb]
        'to stop cluster'
            [tidb.stop]

[mysql]
        'to verify the address'
            [mysql.link.tidb]
        'as client'
            [mysql.exec]
```

**Environment operation check examples**:

Fatal error (read before write):
```
-------=<unsatisfied env read>=-------

<FATAL> 'cluster.port'
       - read by:
            [bench.load]
            [bench.run]
       - but not provided
```

Risk warning (may-read or may-write):
```
-------=<unsatisfied env read>=-------

<risk>  'bench.scale'
       - may-read by:
            [bench.load]
       - but not provided
```

`risk` warnings come from `may-read` or `may-write` statements, used for cases like "if provided in args, use it; otherwise look in environment".

### Cleaner flow display

Use `desc.flow` and `desc.flow.simple` for cleaner views. Aliases: `d.f`, `d.f.s`

```bash
# Description without OS-command report and env-ops checking
$> ticat dummy : dummy : dummy : desc.flow
$> ticat x : desc.flow
$> ticat x : d.f

# Less info, cleaner view
$> ticat x : desc.flow.simple
$> ticat x : d.f.s
```

### Using shortcuts

Use `+` as a shortcut for `desc`:
```bash
$> ticat dummy : dummy : dummy : +
$> ticat x:+
```

Use `-` as a shortcut for `desc.flow.simple`:
```bash
$> ticat dummy : dummy : dummy : -
--->>>
[dummy]
     'dummy cmd for testing'
[dummy]
     'dummy cmd for testing'
[dummy]
     'dummy cmd for testing'
<<<---
```

```bash
$> ticat x:-
--->>>
[x]
        --->>>
        [dummy]
             'dummy cmd for testing'
        [dummy]
             'dummy cmd for testing'
        [dummy]
             'dummy cmd for testing'
        <<<---
<<<---
```

## Best practices

Here are recommended practices for working with flows:

1. **Use `-` for general checking** - It's faster and shows the essential information
2. **Always use `+` to check a flow before executing** - Verify environment requirements and dependencies
3. **Set descriptive help strings** - Help yourself and others understand what the flow does
4. **Avoid environment definitions in flows** - Separate process logic from configuration for better reusability
5. **Organize flows with nested paths** - Use paths like `team.project.feature` for better organization
6. **Share useful flows** - Move flows to git repositories for team collaboration

## Flow workflow examples

### Development workflow

```bash
# Create a development flow
$> ticat local.build : cluster.local : cluster.restart : bench.run : flow.save dev.test

# Run it during development
$> ticat dev.test

# Add step-by-step debugging when needed
$> ticat dbg.step.on : dev.test
```

### CI/CD workflow

```bash
# Create a comprehensive test flow
$> ticat build : test.unit : test.integration : test.e2e : flow.save ci.full

# Run preflight check
$> ticat ci.full :+

# Execute
$> ticat ci.full
```

### Customization workflow

```bash
# Start from a shared flow
$> ticat shared.bench :+

# Customize it
$> ticat shared.bench : custom.report : flow.save my.bench

# Save configuration separately
$> ticat {bench.scale=10 bench.threads=8} my.bench : flow.save my.bench.large
```
