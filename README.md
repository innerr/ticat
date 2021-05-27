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
We want to do a benchmark for the demo distributed system,
our workmates already wrote a bench tool and push to git server,
it's easy to fetch it:
```
$> ticat hub.add innerr/quick-start-usage.ticat
```


Find out what we got by search the repo's name.
`@ready` is a conventional tag use for ready-to-go commands:
```
$> ticat find quick-start-usage @ready
[bench|ben]
...
```
From the search we know the command `bench`(has a alias `ben`).


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


We could save the port value to env:
```
$> ticat {cluster.port=4000} env.save
```
So we don't need to type it down every time:
```
$> ticat bench
```

Run benchmark with a larger dataset,
use step-by-step, it will ask for confirming on each step:
(both could save as default by `env.save`)
```
$> ticat {bench.scale=10} dbg.step.on : bench
```


## Assamble pieces to powerful flows
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
$> ticat {bench.scale=4} dev.bench : cluster.jitter-scan
```


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


We have the info during running.
But sometimes it's nice to have a preflight check,
command `desc` is what we need.
The desc result of `xx` will be long,
so let's checkout `bench`:
```
$> ticat bench : desc
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


## Cheat sheet
* Use `:` to concate commands into sequence, will be execute one by one
* Use `{key=value}` to modify env key-values
* Abbreviations display like `real-name|abbr-name|abbr-name`
* Use `ticat find <str> <str> ..` to locate the command we needed, the amount will be huge.
* `ticat cmds.tree <command>` shows a command sub-tree. the sub-trees below are important:
    - `ticat cmds.tree hub`: manage the git repo list we added.
        - `ticat hub.init`: add default repo
        - `ticat hub.add <repo-addr>`
    - `ticat cmds.tree cmds`: manage all commands we could call. abbr: `m.t`
        - `ticat cmds.tree.simple <command>`: short and clean display. abbr: `m.t.s`
    - `ticat cmds.tree flow`: manage saved flows.
        - `flow.save`: save flows, call flows like normal commands. abbr: `f.s`
    - `ticat cmds.tree env`: manage env key-values.
        - `env.save`: save any env modifications. abbr: `e.s`
        - `env.ls`: list all key-values.
    - `ticat cmds.tree desc`:
        - `desc`: full info.
        - `desc.simple`: give a lite description. abbr: `d.s`


## All we need to know
* [Usage examples](./doc/usage)
    - [Basic: build, run commands](./doc/usage/basic.md)
    - [Hub: get modules and flows from others](./doc/usage/hub.md)
    - [Manipulate env key-values](./doc/usage/env.md)
    - [Use flows](./doc/usage/flow.md)
    - [Use abbreviations / alias](./doc/usage/abbr.md)


## Module developing zone
* [Quick-start](./doc/quick-start-mod.md)
* [Examples: write modules in different languages](https://github.com/innerr/examples.ticat)
* [How modules work together (with graphics)](./doc/concept-graphics.md)
* [Specifications](./doc/spec)
    - [Hub: list/add/disable/enable/purge](./doc/spec/hub.md)
    - [Command sequence](./doc/spec/seq.md)
    - [Command tree](./doc/spec/cmd.md)
    - [Env: list/get/set/save](./doc/spec/env.md)
    - [Abbrs of commands, env-keys and flows](./doc/spec/abbr.md)
    - [Flow: list/save/edit](./doc/spec/flow.md)
    - [Display control in executing](./doc/spec/display.md)
    - [Help command](./doc/spec/help.md)
    - [Local store dir](./doc/spec/local-store.md)
    - [Repo tree](./doc/spec/repo-tree.md)
    - [Module: env and args](./doc/spec/mod-interact.md)
    - [Module: meta file](./doc/spec/mod-meta.md)
    - (this is only **ticat**'s spec, a repo provides modules and flows will have it's own spec)


## Inside **ticat**
* [Roadmap and progress](./doc/progress.md)
* [Zen: how the choices are made](./doc/zen.md)
* [An user story: try to be a happy TiDB developer](https://github.com/innerr/tidb.ticat) (on going)
