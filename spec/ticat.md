## Env vars
```
$> ticat env
builtin
    cluster = cluster-1
    interact = on
    burn-confirm = off
    tidb-proxy = [none] (host:port)
tpcc
    ...
tpcc.report
    ...

$> ticat env interact
interact = on
$> ticat env interact=off
interact = off
$> ticat env interact off
interact = off
$> ticat env interact = off
interact = off

# one-time only env vars
$> ticat with interact=off cluster=cluster-2 exe (other commands...)
$> ticat with interact [= ]off exe (other commands...)
```

## Sub command
```
# the two same ways to call a sub-command
$> ticat status job
$> ticat status/job
```

## Sequence commands
```
$> ticat up : fstop
$> ticat up: fstop
$> ticat up :fstop
$> ticat up:fstop

$> ticat up : fstop : sleep 30s : up
```

## With ... exe
```
# usage, target-name could be any of component-name or host-name or cluster-name
$> ticat with [target-name] [target-name] [var=value] exe {sub-command}

# only show status of tikv and pd
$> ticat with tikv pd exe status
...

# only show status of tikv on 172.16.5.4
$> ticat with tikv [172.16.]5.4 exe status

# show status of cluster except promethus on 172.16.5.4
$> ticat with -prom[methus] exe status

# dry run mode
$> ticat with debug.dry=on exe tpcc/bench

# verbose mode
$> ticat with debug.verb=on exe tpcc/bench

# restart all tikv
$> ticat with tikv exe fstop : up
```
