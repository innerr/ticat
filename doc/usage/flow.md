# Flow: assemble pieces into greate power

## Command sequences and flows

### Command sequences

We are fimiliar with using commands in **ticat**,
It's a little bit like unix-pipe `|`, but different:
* use `:` to concatenate commands, not `|`
* execute commands one by one, the second one won't start untill the previous one finishes
* show executing info in a box, the `>>` in the box indicate the current command about to run

Sequences are like unix-pipe, but use `:` instead of `|`:
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
The boxes indicate the running command by `>>`

### Command sequence == `flow`

We use the name `flow` to call those sequences,
`flow` could easily persisted to local disk, then it could be called like a regular command.
Commands in a flow also could be other flows, in that we are able to assemble complicated features.

The command branch `flow` is for managing saved flows,
the command branch `desc` is for displaying how the flow will do (without executing it).
`+` `-` can do the most jobs of `desc` as a shortcut.

## Save, call, edit or remove flows

Command branch `flow` overview:
```
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

### Save command sequence to a flow

Use `flow.save` to save a sequence to local, alias `f.+`
```
$> ticat dummy : dummy : dummy : f.+ x
```
If the flow "x" already exists, there will be an overwriting confirming.

Save command sequence to a flow with longer path:
```
$> ticat x : x : flow.save aa.bb.cc
```

### Run a saved flow

A flow is a regular command, any rule suits a command also suits a flow:
```
$> ticat x
(execute dummy * 3)
```

Call a nested flow:
```
$> ticat aa.bb.cc
(execute dummy * 6)
```

### List all saved flows

The command `flow` (also a branch) will show all saves flows, alias `f`:
```
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
The flow saved file paths are showed, manually edit them if we like.

### Remove saved flows

`flow.remove` will delete a single flow, alias `f.-`:
```
$> ticat f.- aa.bb.cc
```

`flow.clear` will delete all flows, alias `f.--`:
```
$> ticat f.--
```

## Share saved flows

### Add a help string to a saved flow

When we have lots of saved flows, or before sharing them,
it's helpful to add help strings on flows.

Use `flow.set-help-str` to set help string, alias `f.help` or `f.h`:
```
## save a flow
$> ticat dummy : dummy : dummy : f.+ x
## set help string
$> ticat f.h x 'power test'
## show the help string
$> ticat c x
[x]
     'power test'
    - from:
        ...
    - flow:
        dummy : dummy : dummy
```

### Share saved flows

To share saved flows, we need to move the saved files from **ticat** storing dir to specific dir,
then if we push the dir to(as) a git repo, those files are being shared.

The flow files could be any place in the repo,
but we recommend to put these files on root dir of the repo, or sub dir name `flows`.
Because dir scanning are slow, one day **ticat** may only scan some specific dirs.

Use `flow.move-flows-to-dir` to relocate saved flow files, alias `f.mv`:
```
$> ticat f.mv path=./tmp
```

### Advanced flow file moving

If one(and only one) local dir exists in hub
(local dir means the dir is not managed by hub as a repo),
and the arg "path" is empty, then flows will move to that dir.

Notice that if the destiny dir is not in hub, we can't call those moved flows after moving.

We could also manually move the files, they are at a dir defined by env `sys.paths.flows`:
```
$> ticat env.flat flow path
sys.paths.flows = (a local dir)
```

## Dig into a flow

### Display properties of a flow

Use `cmds` to show the detail properties of a flow (the same way as other commands), alias `c`:
```
## save a flow
$> ticat dummy : dummy : dummy : f.+ x

## show info of a flow
$> ticat x
[x]
    - flow:
        dummy : dummy : dummy
...
```

### Display how a flow will execute

`desc` and `desc.simple` show how a flow will do without executing it, abbrs `d` `d.s`:
```
## full description
$> ticat dummy : dummy : dummy : desc
$> ticat x : desc
$> ticat x : d

## full description, but less info about modules
$> ticat x : desc.simple
$> ticat x : d.s
```

The commands `desc` and `desc.simple` display full description of the execution,
they also check and give reports about module dependencies of os-commands.

An example of os-commands report:
* the os-command name
* which modules are using this os-command
* why a module depends on this os-command
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

The commands `desc` and `desc.simple` also give an report about env-ops.
An env key-value being read before write will cause a `FATAL` error.

`risk` is caused by `may-read` or `may-write` statements,
these statements are use for cases like "if it's provided in args then use it, or else looking in env".

Examples of env-ops check results:
```
-------=<unsatisfied env read>=-------

<FATAL> 'cluster.port'
       - read by:
            [bench.load]
            [bench.run]
       - but not provided
```
```
-------=<unsatisfied env read>=-------

<risk>  'bench.scale'
       - may-read by:
            [bench.load]
       - but not provided
```

### Better way to display a flow execution

Use `desc.flow` `desc.flow.simple` to get a cleaner view, abbrs `d.f` `d.f.s`:
```
## description without os-command report and env-ops checking
$> ticat dummy : dummy : dummy : desc.flow
$> ticat x : desc.flow
$> ticat x : d.f

## less info, cleaner view
$> ticat x : desc.flow.simple
$> ticat x : d.f.s
```

Use `+` as `desc`:
```
$> ticat dummy : dummy : dummy : +
$> ticat x:+
```

Use `-` as `desc.flow.simple`:
```
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
```
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

## Best practice

Here are some recommended practices
* Use `-` (not `desc`) to do general checking
* Always `+` to check a flow before executing it
* Set help strings to flows
* Better no env definitions in a flow
