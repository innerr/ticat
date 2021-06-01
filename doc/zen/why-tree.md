# Zen: the choices we made

## Why commands and env key-values are in tree form

In fact, most command line tools choose 2~3 depth sub-command organizating,
for example:
```
$> tiup cluster start
$> tiup cluster deploy
```

The differences are little:
* Ticat join the command path with `.` into an id like `cluster.start`
* Ticat has unlimit depth

The reason is to deal with huge number of commands,
**ticat** is a platform with huge amount of commands from different authors,
the tree form provides a namespace mechanism to avoid command name conflictings.

The branch path could be used for a specific concept,
the authors provide tools in this sub-tree,
and the uses can explore this branch for specific tools.

On the end-user side,
looking through the command tree to find what we want is not sugguested,
searching by keywords is a better way.

The reasons are the same in env key-values.
