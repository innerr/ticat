# Manipulate env key-values

## Display env key-values
List all env key-values
```bash
$> ticat env.ls
```

Find env key-values
```bash
$> ticat env.ls <find-str>
```

Find env key-values in finding results
```bash
$> ticat env.ls <find-str> <find-str>
$> ticat env.ls <find-str> <find-str> <find-str>
```

Command `find` or `help` also could use to find env keys:
```bash
$> ticat help <find-str> [<find-str> <find-str>]
$> ticat find <find-str> [<find-str> <find-str>]
```

## Set value
The brackets `{``}` are used to set values during running,
it's OK if the key does't exist in **env**:
```bash
$> tiat {display.width=40}
```

Use another command `env.ls` to show the modified value:
```bash
$> tiat {display.width=40} : env.ls display.width
```

Change multi key-values in one pair of brackets:
```bash
$> tiat {display.width=40 display.style=utf8} : env.ls display
```

## Save value
```bash
$> tiat {display.width=40} env.save
$> tiat env.ls width
display.width = 40
$> tiat {display.width=60} env.save
display.width = 60
```
