# Manipulate env key-values

Env is a set of key-values, it's shared by all modules in an execution.

For users, it's important to find out what keys a command(module or flow) need.

## Display a commands env-ops(read/write)

For a module, `cmds` shows its detail included env-ops:
```
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

For a flow, `desc` shows its unsatisfied env-ops:
```
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
An env key-value being read before write will cause a `FATAL` error, `risk` is normally fine.

`+` can get instead `cmds` or `desc` for either command or flow:
```
$> ticat bench.run:+
$> ticat bench:+
```

## Set env key-values

Modules are recommended to read env instead of reading from args as much as possible,
it's the way to accomplish automatic assembly.

So how to pass values to modules by env is important.
The brackets `{``}` are used to set values during running.

it's OK if the key does't exist in **env**:
```
$> tiat {display.width=40}
```

Use another command `env.ls` to show the modified value:
```
$> tiat {display.width=40} : env.ls display.width
```

Change multi key-values in one pair of brackets:
```
$> tiat {display.width=40 display.style=utf8} : env.ls display
```

## Save env key-values

By saving key-values to env, we don't need to type them down every time.

Save changes of env by `env.save`, short name `e.+`
```
$> tiat {display.width=40} env.save
$> tiat env.ls width
display.width = 40
$> tiat {display.width=60} env.save
display.width = 60
```

## Observe env key-values during running

In the executing info box,  the upper part has the current env key-values.
```
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

There are a large amount of key-values hidden,
to show then we could use `verb` command.
```
$> ticat verb : dummy : dummy
```

`verb` shows maximum infos, shor name `v`.
To show a little more info but not too much, we could use `v.+`:
```
$> ticat v.+ : dummy : dummy
$> ticat v.+ 1 : dummy : dummy
$> ticat v.+ 2 : dummy : dummy
```

`quiet` `q` chould totally hide the executing info:

```
$> ticat q : dummy : dummy
```

## Display env key-values

We know there are lots key-values besides we manually set into env,
`env.flat` is the command to list them all, short name `e.f`:
```
$> ticat e.f
```

Find env key-values:
```
$> ticat e.f <find-str>
```

Find env key-values in finding results
```
$> ticat e.f <find-str> <find-str>
$> ticat e.f <find-str> <find-str> <find-str>
```

The `v` `q` `v.+` commands alter some values to change the display behavior,
we could find those key-values by:
```
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
