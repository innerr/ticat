# [Spec] Abbrs: commands, env-keys, flows

## Abbrs types
There are 3 types of abbrs/alias:
* for command(command segments in path)
* for env(key segments in path)
* for args

## Display abbrs in a command
Almost all displaying commands will show the abbrs,
there are many abbrs info in the example bellow:
* the command name `tree` with abbrs `t` and `T`
* the command path:
    - `cmds` with abbrs/alias: `cmd` `mod` `mods` `m` `M`
    - `tree` with abbrs `t` `T`
* the args `path` with abbrs `p` `P`
```
$> ticat cmds.tree cmds.tree
...
[tree|t|T]
    ...
    - full-abbrs:
        cmds|cmd|mod|mods|m|M.tree|t|T
    - args:
        path|p|P = ''
...
```

## Abbrs borrowing
When two command have a shared segment of path:
```
$> ticat dummy
$> ticat dummy.power
```

If one defined abbrs for the shared segment and the other one didn't define:
* `dummy`: no abbrs
* `dummy.power`: full-abbrs is `dummy|dm.power`
Then the other one cound also use the abbrs:
```
## The command "dummy" borrowed the abbrs from "dm.power":
$> ticat dm
```

## Use abbrs in command
```
$> ticat env.tree
$> ticat e.tree
$> ticat e.t
```

## Use abbrs in setting env key
```
$> ticat {display.width = 40} e.ls width
display.width = 40
$> ticat {disp.w = 60} e.ls width
display.width = 60
```

## Use abbrs in setting env key: with abbr-form command-prefix
Env key could be a path composed by segments just as commands,
so the env segments can form a tree.
Although env tree and command tree are totally not related,
we still let env tree borrow abbrs from commands for better usage:
```
$> ticat env{mykey=88}.ls mykey
env.mykey = 88
## Command env has abbr "e"
$> ticat e.ls
...
## Env key "env.mykey" borrowed abbrs from command "env"
$> ticat e{mykey=66}.ls mykey
env.mykey = 66
```

To show all env abbrs, including borrowed ones:
```
$> ticat env.abbrs
```

## Use abbrs in the argv (not in arg names)
(TODO: implement)
```
$> ticat flow.remove <command-path>
```
