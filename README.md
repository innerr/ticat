# ticat
A casual command line components platform

Workflow automating in unix-pipe style

## Quick start
Suppose we are working for a distributed system,
let's run a demo to see how **ticat** works.
Type and execute the commands we memtioned below.

## Build
(`golang` is needed)
```
$> git clone https://github.com/innerr/ticat
$> cd ticat
$> make
```
Recommend to set `ticat/bin` to system `$PATH`, it's handy.

## Run jobs shared by others

### Add repo to **ticat**

We want to do a benchmark for the demo distributed system,
our workmates already wrote a bench tool and push to git server,
it's easy to fetch it:
```
$> ticat hub.add innerr/quick-start-usage.ticat
```

### The basic usage about `:`, `+` and `-`

`+` and `-` are important commands to find and display infos, they have the same usage.
The difference is `-` only shows brief messages, and `+` shows rich infos.
Apparently `-` will be used more often.


Use `+` as search command to find out what we got by search the repo's name.
`@ready` is a conventional tag use for ready-to-go commands:
```
$> ticat + @ready quick-start
[bench|ben]
     'pretend to do bench. @ready'
...
```
From the search we know the command `bench`(has a alias `ben`).

The usage of **ticat** has similar style with unix pipe, but use `:` instead of `|`.

Concat `bench` and `-` with `:`, it shows the info of command `bench`:
(`+` could do the same job, we choose `-` for a clean view)
```
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

### Try to run benchmark

Looks like `bench` is what we need.
Say we have a single node cluster running, the access port is 4000.
Try to run benchmark by:
```
$> ticat bench
```

Got an error, we should provide the cluster port:
```
-------=<unsatisfied env read>=-------

<FATAL> 'cluster.port'
       - read by:
            [bench.load]
            [bench.run]
       - but not provided
```

Try again:
```
$> ticat {cluster.port=4000} bench
```
Succeeded, we ran a benchmark with a small dataset (scale=1):
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

# Run it easier by manipulating env key-values

We could save the port value to env:
```
$> ticat {cluster.port=4000} env.save
```
So we don't need to type it down every time:
```
$> ticat bench
```

Run benchmark with a larger dataset,
here we use step-by-step, it will ask for confirming on each step:
(both could persist to env by `env.save`)
```
$> ticat {bench.scale=10} dbg.step.on : bench
```

## Assamble pieces to powerful flows

### Call another command

There is another command `dev.bench` in the previous search result:
```
...
[bench]
     'build binary in pwd, then restart cluster and do benchmark. @ready'
    - full-cmd:
        dev.bench
...
```
It do "build" and "restart" before bench based on the comment,
useful for develeping.

The default data scale is 1, we use 4 for a test.
We add a jitter detecting step after benchmark,
this command also have the `@ready` tag so we found it.
```
$> ticat {bench.scale=4} dev.bench : bench.jitter-scan
```

### Save commands to a flow for convenient

This command sequence runs perfectly.
But it's annoying to type all those every time.

So we save it to a `flow` with name `xx`:
```
$> ticat {bench.scale=4} dev.bench : cluster.jitter-scan : flow.save xx
```
Using it in coding is convenient:
```
  (code editing)
$> ticat xx
  (code editing)
$> ticat xx
...
```

### Take a good look at the env key-values

We could use step-by-step to confirm every step,
```
$> ticat dbg.step.on : xx
```
and we could observe what will happen in the info box:
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
From the box we could see it's about to restart cluster,
the upper part has the current env key-values.

### Dig inside the command we got, it's a flow provided by repo author

### Understand the flow: executing modules one by one

Sometimes it's nice to have a preflight check,
appending a command `+` or `-` to the sequence we got answers,
let's check out the flow `xx` we just saved:
```
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

### Understand the env: modules read/write a shared key-value set

The `+` result of `xx` is a bit long, here is `bench`'s:
```
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
From the description, we know how modules are executed one by one,
each one may read or write from the env.

### The env read/write report from `+`

Beside the flow description, there is a check result about env read/write.
An env key-value being read before write will cause a `FATAL` error,
`risk` is normally fine.
```
-------=<unsatisfied env read>=-------

<risk>  'bench.scale'
       - may-read by:
            [cluster.restart]
            [bench.load]
       - but not provided
```

### Customize features by re-assemble pieces

Now we know what's in the "ready-to-go" commands,
we are able to do customizations,
let's remove the `bench.load` step from `dev.bench`,
to make it faster when on coding:
```
$> ticat local.build : cluster.local : cluster.restart : ben.run : flow.save dev.bench.no-reload
```

We just saved a flow without data scale config,
it's a good practice seperating "process-logic" and "config".
We then save a new flow with data scale to get a handy command:
```
$> ticat {bench.scale=4} dev.bench.no-reload : flow.save z
```

More fun:
```
  (code editing)
$> ticat z
...
```

### Share our flows

Each flow is a small file, move it to a local dir, then push it to git server.
Then the repo address with friends then they can use it in **ticat**.

Don't forget to write the help string and add some tags to it,
so other users can tell what it's use for.

For more details, checkout the "Module developing zone" bellow.

Writing new modules also easy and quick,
it only take some minutes to wrap a existing tool into a **ticat** module.
Check out the [quick-start-for-module-developing](./doc/quick-start-mod.md).

## Important command branchs

### The builtin commands

A branch is a set of commands like `env` `env.tree` `env.flat`,
they share a same path branch.

These builtin branchs are important:
* `hub`: manage the git repo list we added. abbr `h`.
* `env`: manage env. abbr `e`.
* `flow`: manage saved flows. abbr `f`.
* `cmds`: manage all callable commands(modules and flows). abbr `c`.

Use `+` `-` to navigate them, here are some usage examples.

Overview branch `cmds`:
```
$> ticat cmds:-
[cmds|cmd|c|C]
    [tree|t|T]
        [simple|sim|skeleton|sk|sl|st|s|S|-]
    [list|ls|flatten|flat|f|F|~]
        [simple|sim|s|S|-]
```

Search "tree"(could be any string) in the branch:
```
$> ticat cmds:- tree
[cmds|cmd|c|C]
[tree|t|T]
    - full-cmd:
        cmds.tree
    - full-abbrs:
        cmds|cmd|c|C.tree|t|T
```

Use `+` instead of `-` to get more detail:
```
$> ticat cmds:+ tree
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
* Use `:` to concate commands into sequence, will be execute one by one
* Use `{key=value}` to modify env key-values
* (With `:`) append `+` or `-`  to any command(s) we want to investigate
* Search commands:
    - `ticat + <str> <str> ..`
    - `ticat <command> :+ <str> <str> ..`
* Frequently-used commands:
    - `hub.add <repo-addr>`, abbr `h.+`
    - `flow.save`, abbr `f.+`
    - `env.save`, abbr `e.+`
* Lots of abbrs like `[bench|ben]` in search result, use them to save typing time

## User manual
* [Usage examples](./doc/usage)
    - [Basic: build, run commands](./doc/usage/basic.md)
    - [Hub: get modules and flows from others](./doc/usage/hub.md)
    - [Use commands](./doc/usage/cmds.md)
    - [Manipulate env key-values](./doc/usage/env.md)
    - [Use flows](./doc/usage/flow.md)
    - [Use abbrs/alias](./doc/usage/abbr.md)

## Module developing zone
* [Quick-start](./doc/quick-start-mod.md)
* [Examples: write modules in different languages](https://github.com/innerr/examples.ticat)
* [How modules work together (with graphics)](./doc/concept-graphics.md)
* [Specifications](./doc/spec)
    - (this is only **ticat**'s spec, a repo provides modules and flows will have it's own spec)
    - [Hub: list/add/disable/enable/purge](./doc/spec/hub.md)
    - [Command sequence](./doc/spec/seq.md)
    - [Command tree](./doc/spec/cmds.md)
    - [Env: list/get/set/save](./doc/spec/env.md)
    - [Abbrs of commands, env-keys and flows](./doc/spec/abbr.md)
    - [Flow: list/save/edit](./doc/spec/flow.md)
    - [Display control in executing](./doc/spec/display.md)
    - [Help command](./doc/spec/help.md)
    - [Local store dir](./doc/spec/local-store.md)
    - [Repo tree](./doc/spec/repo-tree.md)
    - [Module: env and args](./doc/spec/mod-interact.md)
    - [Module: meta file](./doc/spec/mod-meta.md)

## Inside **ticat**
* [Roadmap and progress](./doc/progress.md)
* [Zen: how the choices are made](./doc/zen.md)
* [An user story: try to be a happy TiDB developer](https://github.com/innerr/tidb.ticat) (on going)
