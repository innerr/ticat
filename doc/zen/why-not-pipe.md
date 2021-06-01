# Zen: the choices we made

## Why not just use unix-pipe?

### Executing order

Even **ticat** use unix-pipe style, but it's not the same.

The executing orders are different:
* In pipe, all commands launch at the same time
* In **ticat**, commands launch one by one

Apparently, **ticat** do this to meet the "workflow control" demand, so pipe is not fit here.

### Unix-pipe is weak, ticat env is strong

* Unix-pipe is anonymous, hard to define and force to apply a specific protocol in concept.
    - named-pipe(mkfifo) could solve this, but recycling will be a hard job.
* There is only one pipe between commands, which we can only passing one type of data.

Of cause we could define protocols on pipe to solve all those,
but it will make it inconvenient to write a component(eg, in bash).

In **ticat**, env key-values can be considered as another form of named-pipe,
with a name, a key-value could bind to a format definition.

There is no recycling issue about env key-values,
even it have, **ticat** implemented session model, could handle it easily.

The env-ops(read/write) in a **ticat** command sequence will be checked before executing,
commands are required to register what keys them will read or write,
read before write will cause fatal report.
So the **ticat** model is a managed named-pipe-like system.

A **ticat** flow can be called in another flow(command sequence),
which forms callstacks, this is important in complicated assemblings.
To support this, unix-pipe is far not enough.

The callstack depth display in executing (stack-level):
```
┌───────────────────┐
│ stack-level: [3]  │             06-02 04:51:45
├───────────────────┴────────────────────────────┐
│    bench.scale = 4                             │
│    cluster.port = 4000                         │
├────────────────────────────────────────────────┤
│    local.build                                 │
│ >> cluster.local                               │
│        port = ''                               │
│        host = 127.0.0.1                        │
│    cluster.restart                             │
│    ben(=bench).load                            │
│    ben(=bench).run                             │
└────────────────────────────────────────────────┘
```
