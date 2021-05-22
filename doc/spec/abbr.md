# [Spec] Abbrs: commands, env-keys, flows

## Abbrs in command
```bash
$> ticat env.tree
...
$> ticat e.tree
...
$> ticat e.t
...
```

## Abbrs in setting env key
```bash
$> ticat {display.width = 40} e.ls width
display.width = 40
$> ticat {disp.w = 60} e.ls width
display.width = 60
```

## Abbrs in setting env key: with abbr-form command-prefix
```bash
$> ticat env{mykey=88}.ls mykey
env.mykey = 88
$> ticat e{mykey=66}.ls mykey
env.mykey = 66
```
