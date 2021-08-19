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

# restart all tikv
$> ticat {tikv} fstop : up

# restart all tikv+pd+tidb, then close all tidb
$> ticat {tikv pd tidb} fstop : {tikv} up
```

## Run a formal benchmark
Run a formal benchmark on a 3 nodes cluster and create report when the optimization is done:
```
## Acquire a server to run benchmark client, login to it
$> ticat host.acquire core=16 : host.clone-ticat-to : host.login

## Acquire cluster hardware resource, deploy cluster, then save cluster info
$> ticat cluster.deploy : env.save
$> ticat hosts.acquire core=8 min-mem=16G cnt=3 : cluster.deploy : env.save

## Run the optimized new code on the new cluster, then release hardware resource
$> ticat bb : ben.report : hosts.release
```
(actually there is a saved flow for formal benchmark, so we just run the command bellow)
```
$> ticat ben.scale 10 : ben.base
```

## Investigate a performance issue
(TODO: better example)
Someone talk to us that our system has preformance jitter under a type of specific workload,
So we fetch this workload and test it:
```
## Fetch this workload
$> ticat hub.add innerr/workload-x.demo.ticat

## Start local cluster and save info to env
$> ticat cluster.local port=4000 : cluster.start : env.save

## Run benchmark and detect jitter
$> ticat ben.x.scale 10 : ben.x.load : ben.x.run dur=10s : ben.scanner.jitter
```
