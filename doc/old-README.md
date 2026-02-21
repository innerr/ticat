# ticat

A lightweight command-line component platform for workflow automation in unix-pipe style

**ticat** (Tiny Component Assembly Tool) is a modular CLI framework that enables you to:
- Share and reuse command-line tools across teams and projects
- Assemble complex workflows from simple, composable modules
- Manage configurations through a shared environment system
- Distribute components via git repositories

[![Go Version](https://img.shields.io/badge/Go-1.16%2B-blue)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)
[![CI](https://github.com/innerr/ticat/actions/workflows/ci.yml/badge.svg)](https://github.com/innerr/ticat/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/innerr/ticat)](https://goreportcard.com/report/github.com/innerr/ticat)

## Quick start

Suppose we are working on a distributed system. Let's run a demo to see how **ticat** works.

**Recommendation**: Type and execute the commands below as you read to get hands-on experience.

### Download and install

**Option 1: Install via curl**
```bash
$> curl --proto '=https' --tlsv1.2 -sSf https://raw.githubusercontent.com/innerr/ticat/main/install.sh | sh
```

**Option 2: Build from source**

Golang 1.16+ is required:
```bash
$> git clone https://github.com/innerr/ticat
$> cd ticat
$> make
```

**Recommendation**: Add `ticat/bin` to your system `$PATH` for convenient access.

## Run jobs shared by others

### Add a repository to **ticat**

We want to run a benchmark for our demo distributed system.

Someone has already written a benchmark tool and pushed it to a git server. We can fetch it easily using the `hub.add` command:
```bash
$> ticat hub.add innerr/quick-start-usage.ticat
```

### Discover available commands

The `/` and `//` commands are essential for finding commands. They work similarly:
- `/` displays brief information
- `//` displays detailed information

Let's use `/` to search for commands from the repo we just added:
```bash
$> ticat / quick-start-usage @ready
[bench]
     @ready
     'pretend to do bench.'
...
```
From the search results, we found the `bench` command. (Tags like `@ready` are defined by the module author. Use the `@` command to list all available tags.)

### Understand what a command does

The usage of **ticat** follows a unix-pipe style, but uses `:` instead of `|`.

Appending `-` to a command with `:` shows its information:
```bash
$> ticat bench:-
--->>>
[bench]
     'pretend to do bench. @ready'
        --->>>
        [bench.load]
             'pretend to load data'
        [bench.run]
             'pretend to run bench'
        <<<---
<<<---
```
You can also use `+` instead of `-` to see more detailed information. We use `-` for a cleaner view.

### Run the shared benchmark tool

The `bench` command looks like what we need to run benchmarks.

Suppose we have a single-node cluster running on port 4000.

Let's try to run it:
```bash
$> ticat bench
```

We'll get an error:
```
-------=<unsatisfied env read>=-------

<FATAL> 'cluster.port'
       - read by:
            [bench.load]
            [bench.run]
       - but not provided
```

We need to provide the cluster port. Let's try again:
```bash
$> ticat {cluster.port=4000} bench
```

Success! We ran a benchmark with a small dataset (scale=1):
```
┌───────────────────┐
│ stack-level: [2]  │             05-27 21:20:29
├───────────────────┴────────────────────────────┐
│    cluster.port = 4000                         │
├────────────────────────────────────────────────┤
│ >> bench.load                                  │
│    bench.run                                   │
└────────────────────────────────────────────────┘
data loading to 127.0.0.1:4000 begin, scale=1
...
data loading to 127.0.0.1:4000 finish
┌───────────────────┐
│ stack-level: [2]  │             05-27 21:20:30
├───────────────────┴────────────────────────────┐
│    bench.scale = 1                             │
│    cluster.host = 127.0.0.1                    │
│    cluster.port = 4000                         │
├────────────────────────────────────────────────┤
│    bench.load                                  │
│ >> bench.run                                   │
└────────────────────────────────────────────────┘
benchmark on 127.0.0.1:4000 begin, scale=1
...
benchmark on 127.0.0.1:4000 finish
```

# Manipulate environment key-values

We can save the port value to the environment:
```bash
$> ticat {cluster.port=4000} env.save
```

Now we don't need to type it every time:
```bash
$> ticat bench
```

To run with a larger dataset, let's enable "step-by-step" mode. It will ask for confirmation on each step:
```bash
$> ticat {bench.scale=10} dbg.step.on : bench
```

All these changes can be persisted using `env.save`.

## Assemble pieces into flows

### Call another command

There's another command `dev.bench` in the previous search results:
```
...
[dev.bench]
      @ready
     'build binary in pwd, then restart cluster and do benchmark.'
[bench.jitter-scan]
      @ready @scanner
     'pretend to scan jitter'
```
According to the help string, it does "build" and "restart" before running the benchmark - useful for development.

The default data scale is "1", let's use "4" for testing. Also, let's add a jitter detection step after the benchmark:
```bash
$> ticat {bench.scale=4} dev.bench : bench.jitter-scan
```

### Save commands to a flow for convenience

This command sequence runs perfectly, but typing it every time is tedious.

Let's save it as a `flow` with the name `xx`:
```bash
$> ticat {bench.scale=4} dev.bench : cluster.jitter-scan : flow.save xx
```

Now using it during development is convenient:
```bash
...
  (code editing)
$> ticat xx
  (code editing)
$> ticat xx
...
```

### Examine environment key-values

We can use "step-by-step" mode to confirm each step:
```bash
$> ticat dbg.step.on : xx
```

We can observe what will happen in the info box:
```
...
┌───────────────────┐
│ stack-level: [2]  │             05-27 23:32:07
├───────────────────┴────────────────────────────┐
│    bench.scale = 3                             │
│    cluster.host = 127.0.0.1                    │
│    cluster.port = 4000                         │
├────────────────────────────────────────────────┤
│    local.build                                 │
│    cluster.local                               │
│        port = ''                               │
│        host = 127.0.0.1                        │
│ >> cluster.restart                             │
│    ben(=bench).load                            │
│    ben(=bench).run                             │
└────────────────────────────────────────────────┘
...
```
The box shows that we're about to restart the cluster. The upper part displays current environment key-values.

## Dig deeper into commands

The commands we received are flows provided by the repo author, just like the `xx` flow we saved.

### Understanding flows: executing modules one by one

Sometimes it's helpful to do a preflight check before executing.

Appending `+` or `-` to a command sequence shows what will happen:

Let's examine the `xx` flow we just saved:
```bash
$> ticat xx:-
--->>>
[xx]
        --->>>
        [dev.bench]
             'build binary in pwd, then restart cluster and do benchmark. @ready'
                --->>>
                [local.build]
                     'pretend to build demo cluster's server binary'
                [cluster.local]
                     'set local cluster as active cluster to env'
                [cluster.restart]
                     'pretend to run bench'
                [bench.load]
                     'pretend to load data'
                [bench.run]
                     'pretend to run bench'
                <<<---
        [bench.jitter-scan]
             'pretend to scan jitter @ready'
        <<<---
<<<---
```

### Environment: a shared key-value set

Let's examine `bench` with `+` (the result for `xx` is a bit long, so we use `bench`):
```bash
$> ticat bench:+
```

The output will be a full description of the execution:
```
--->>>
[bench]
     'pretend to do bench'
    - flow:
        bench.load : bench.run
        --->>>
        [bench.load]
             'pretend to load data'
            - env-ops:
                cluster.host = may-read : write
                cluster.port = read
                bench.scale = may-read : write
        [bench.run]
             'pretend to run bench'
            - env-ops:
                cluster.host = may-read
                cluster.port = read
                bench.scale = may-read
                bench.begin = write
                bench.end = write
        <<<---
<<<---
```

From this description, we can see how modules execute one by one, and how each one reads from or writes to the environment.

### Environment read/write report from `+`

Besides the flow description, there's also a check result about environment read/write operations.

An environment key-value being read before being written will cause a `FATAL` error. A `risk` warning is normally acceptable:
```
-------=<unsatisfied env read>=-------

<risk>  'bench.scale'
       - may-read by:
            [bench.load]
       - but not provided
```

### Customize features by reassembling pieces

Now that we understand what's in the "ready-to-go" commands, we can customize them.

Let's remove the `bench.load` step from `dev.bench` to make development iterations faster:
```bash
$> ticat local.build : cluster.local : cluster.restart : bench.run : flow.save dev.bench.no-reload
```

We just saved a flow without data scale configuration - it's good practice to separate "process logic" from "configuration".

Then we can save a new flow with the data scale setting for convenience:
```bash
$> ticat {bench.scale=4} dev.bench.no-reload : flow.save z
```

Use it:
```bash
...
  (code editing)
$> ticat z
  (code editing)
$> ticat z
...
```

### Share your flows

Each flow is a small file. Move it to a local directory, then push it to a git server.

Share the repository address with friends, and they can use it in **ticat**.

It's helpful to write a clear help string and add relevant tags so other users can understand what the flow does.

For more details, check out the "Module developing zone" section below.

Writing new modules is also easy and quick - it only takes a few minutes to wrap an existing tool into a **ticat** module. Check out the [quick-start-for-module-developing](./doc/quick-start-mod.md).

## Important command branches

### Builtin commands

A branch is a set of commands like `env`, `env.tree`, `env.flat` - they share the same path prefix.

These builtin branches are important:
- `hub`: manage the git repository list. Abbreviation: `h`
- `env`: manage environment variables. Abbreviation: `e`
- `flow`: manage saved flows. Abbreviation: `f`
- `cmds`: manage all callable commands (modules and flows). Abbreviation: `c`

Use `~` and `~~` to navigate them. Here are some examples:

Overview of branch `cmds`:
```bash
$> ticat cmds:~
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

Overview of branch `env`:
```bash
$> ticat env:~
[env]
     'list env values in flatten format'
    [tree]
         'list all env layers and KVs in tree format'
    [abbrs]
         'list env tree and abbrs'
    [list]
         'list env values in flatten format'
    [save]
         'save session env changes to local'
    [remove-and-save]
         'remove specific env KV and save changes to local'
    [reset-and-save]
         'reset all local saved env KVs'
```

Search for "tree" (or any string) in the branch `cmds`:
```bash
$> ticat cmds:~ tree
[cmds]
     'display cmd info, sub tree cmds will not show'
[cmds.tree]
     'list builtin and loaded cmds'
```

Use `~~` instead of `~` to get more details:
```bash
$> ticat cmds:~~ tree
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

## Cheat sheet

**Essential syntax:**
- Use `:` to concatenate commands - they will be executed sequentially
- Use `{key=value}` to modify environment key-values

**Command inspection:**
- Append `=` or `==` to any command(s) to investigate them (used with `:`)

**Search commands:**
- `ticat / <str> <str> ..` - global search
- `ticat <command> :/ <str> <str> ..` - search within a branch

**Frequently used commands:**
- `hub.add <repo-addr>` - add a repository. Abbreviation: `h.+`
- `flow.save` - save a flow. Abbreviation: `f.+`
- `env.save` - save environment. Abbreviation: `e.+`
- `+` - show detailed information
- `-` - show brief information

**Tips:**
- Lots of abbreviations/aliases are available (e.g., `[bench|ben]` in search results) - use them to save typing time

## User manual
- [Usage examples](./doc/usage/user-manual.md)
  - [Basic: build, run commands](./doc/usage/basic.md)
  - [Hub: get modules and flows from others](./doc/usage/hub.md)
  - [Use commands](./doc/usage/cmds.md)
  - [Manipulate environment key-values](./doc/usage/env.md)
  - [Use flows](./doc/usage/flow.md)

## Module developing zone
- [Quick-start](./doc/quick-start-mod.md)
- [Examples: write modules in different languages](https://github.com/innerr/examples.ticat)
- [How modules work together (with graphics)](./doc/concept-graphics.md)
- [Specifications](./doc/spec/spec.md)
  - (this is only **ticat**'s spec; a repository providing modules and flows will have its own spec)
  - [Hub: list/add/disable/enable/purge](./doc/spec/hub.md)
  - [Command sequence](./doc/spec/seq.md)
  - [Command tree](./doc/spec/cmds.md)
  - [Environment: list/get/set/save](./doc/spec/env.md)
  - [Abbreviations of commands, env-keys and flows](./doc/spec/abbr.md)
  - [Flow: list/save/edit](./doc/spec/flow.md)
  - [Display control in executing](./doc/spec/display.md)
  - [Help info commands](./doc/spec/help.md)
  - [Local store directory](./doc/spec/local-store.md)
  - [Repository tree](./doc/spec/repo-tree.md)
  - [Module: environment and args](./doc/spec/mod-interact.md)
  - [Module: meta file](./doc/spec/mod-meta.md)

## Inside **ticat**
- [Roadmap and progress](./doc/progress.md)
- [Zen: how the choices are made](./doc/zen/zen.md)
  - [Why ticat](./doc/zen/why-ticat.md)
  - [Why use CLI as component platform](./doc/zen/why-cli.md)
  - [Why not use unix pipe](./doc/zen/why-not-pipe.md)
  - [Why the usage seems weird, especially `+` and `-`](./doc/zen/why-weird.md)
  - [Why use tags](./doc/zen/why-tags.md)
  - [Why so many abbreviations and aliases](./doc/zen/why-abbrs.md)
  - [Why commands and environment key-values are in tree form](./doc/zen/why-tree.md)
  - [Why use git repositories to distribute components](./doc/zen/why-hub.md)
  - [Why not support async/concurrent executing](./doc/zen/why-not-async.md)

## User stories
- [Try to be a happy TiDB developer](https://github.com/innerr/tidb.ticat) (ongoing)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Community

- GitHub Issues: [https://github.com/innerr/ticat/issues](https://github.com/innerr/ticat/issues)
- Pull Requests: [https://github.com/innerr/ticat/pulls](https://github.com/innerr/ticat/pulls)
