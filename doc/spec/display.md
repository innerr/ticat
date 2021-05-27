# [Spec] Display: styles and verb level

## The display flags
There is a set of env key defining how much info will be showed to users:
```
$> ticat env.ls display
display.bootstrap = false
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
display.style = ascii
display.utf8 = true
```

We could change this env manually, or use some builtin commands for convenient.

## The display style of the sequence box
The value of "display.style" could be "ascii" or "utf" or "slash" or "no-corner",
We could choose one as we like.

If "display.style" is "utf",
"display.utf8" need to be "true" to display frames in utf8 charset.

```
$> ticat dummy:dummy
+-------------------+
| stack-level: [1]  |                                           05-24 04:05:33
+-------------------+----------------------------------------------------------+
| >> dummy                                                                     |
|    dummy                                                                     |
+------------------------------------------------------------------------------+
dummy cmd here
+-------------------+
| stack-level: [1]  |                                           05-24 04:05:33
+-------------------+----------------------------------------------------------+
|    dummy                                                                     |
| >> dummy                                                                     |
+------------------------------------------------------------------------------+
dummy cmd here

$> ticat {dis.style=utf}:dummy:dummy
┌───────────────────┐
│ stack-level: [1]  │                                           05-24 04:05:03
├───────────────────┴──────────────────────────────────────────────────────────┐
│ >> dummy                                                                     │
│    dummy                                                                     │
└──────────────────────────────────────────────────────────────────────────────┘
dummy cmd here
┌───────────────────┐
│ stack-level: [1]  │                                           05-24 04:05:03
├───────────────────┴──────────────────────────────────────────────────────────┐
│    dummy                                                                     │
│ >> dummy                                                                     │
└──────────────────────────────────────────────────────────────────────────────┘
dummy cmd here
```

## The verb commands
The overview of verb commands:
```
$> ticat m.t.s verb
[verbose|verb|v|V]
    [default|def|d|D]
    [increase|inc|v+|+]
    [decrease|dec|v-|-]

$> ticat m.t.s q
[quiet|q|Q]
```

The commands "verb" and "quiet" is the fast way to display minimum/maximum info:
```
# Display all info when executing
$> ticat verb : dummy : dummy
$> ticat v : dummy : dummy

# Not display any info when executing
$> ticat quiet : dummy : dummy
$> ticat q : dummy : dummy
```

The commands under "verb"(alias: v) is convenient to ajust the amount:
```
$> ticat verb.increase : dummy : dummy
$> ticat v.+ : dummy : dummy
$> ticat v.+ 2 : dummy : dummy
$> ticat quiet : verb.default : dummy : dummy
```

The detail of verb commands:
```
$> ticat m.t verb
[verbose|verb|v|V]
     'change into verbose mode'
    - cmd-type:
        normal (quiet)
    [default|def|d|D]
         'set to default verbose mode'
        - full-cmd:
            verbose.default
        - full-abbrs:
            verbose|verb|v|V.default|def|d|D
        - cmd-type:
            normal (quiet)
    [increase|inc|v+|+]
         'increase verbose'
        - full-cmd:
            verbose.increase
        - full-abbrs:
            verbose|verb|v|V.increase|inc|v+|+
        - cmd-type:
            normal (quiet)
        - args:
            volume|vol|v|V = 1
    [decrease|dec|v-|-]
         'decrease verbose'
        - full-cmd:
            verbose.decrease
        - full-abbrs:
            verbose|verb|v|V.decrease|dec|v-|-
        - cmd-type:
            normal (quiet)
        - args:
            volume|vol|v|V = 1

$> ticat m.t q
[quiet|q|Q]
     'change into quiet mode'
    - cmd-type:
        normal (quiet)
```
