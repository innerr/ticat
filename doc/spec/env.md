# Spec: env get/set/save/bootstrap

## Gramma

Define or set env between segments of a command:
```bash
$> ticat env{mykey=666}.ls mykey
env.mykey = 666
```

Extra space chars (space and tab) will be ignore:
```bash
$> ticat {display.width = 40}
```

## Env layers
```bash
$> ticat env.tree
```

## Session

