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

# one-time only env vars
$> ticat {interact=off cluster=cluster-2} (other commands...)
$> ticat {interact = off} (other commands...)
```

## The '{...}' statement
```
# usage, target-name could be any of module-name or host-name or cluster-name
$> ticat {[target-name] [target-name] [var=value]} sub-command

# only show status of tikv and pd
$> ticat {tikv pd} status
...

# only show status of tikv on 172.16.5.4
$> ticat {tikv [172.16.]5.4} status

# show status of cluster except promethus on 172.16.5.4
$> ticat {-prom[methus]} status

# dry run mode
$> ticat {debug.dry=on} tpcc/bench

# verbose mode
$> ticat {debug.verb=on} tpcc/bench

# restart all tikv
$> ticat {tikv} fstop : up

# restart all tikv+pd+tidb, then close all tidb
$> ticat {tikv pd tidb} fstop : {tikv} up
```
