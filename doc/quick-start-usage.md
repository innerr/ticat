# Quick start

## Take tour by running a demo
Suppose we are working for a distributed system,
Let's run a demo to see how **ticat** works.

Follow the procedure bellow step by step, (copy + paste to the terminal and run)
notice that this is a demo, things like "create a cluster" are emulations.

## Get or build ticat
We could download a **ticat** binary file or build it by ourselves.

Build ticat: (`golang` is needed)
```bash
$> git clone https://github.com/innerr/ticat
$> cd ticat
$> make
```

Recommand to set `ticat/bin` to system `$PATH`, it's handy.

## Run a job(flow) shared by others(workmates)
We want to do a benchmark for the system,
others already wrote some tools, so we just donwload them by:
```bash
$> ticat hub.add innerr/quick-start-usage.ticat
```

Say we have a debuging single node cluster running, the access port is 4000.
Run benchmark by:
```bash
$> ticat {cluster.port=4000} : bench.load : bench.run
```

Save the port config, then run again:
**ticat** support lots of abbrs and alias, here we use abbrs "ben" for "bench"
```bash
$> ticat {cluster.port=4000} env.save
$> ticat ben.load : ben.run
```

Enable step-by-step for better view, run again:
```bash
## Enable step-by-step
$> ticat dbg.step.on : env.save

## Run again:
$> ticat ben.load : ben.run
```

We could see the default data-scale is small(=1),
Increase the scale and run again:
```bash
## Run again:
$> ticat ben.scale 10 : ben.load : ben.run
```

## Daily coding
We are now trying to improve performance,
so re-build and run benchmark after code modifications by:
```bash
$> ticat local.build : cluster.local port=4000 : cluster.restart : ben.scale 10 : ben.load : ben.run
```

We need to do this many times, so save it as a flow:
```bash
## Save the commands to "bb"
$> ticat local.build : cluster.local port=4000 : cluster.restart : ben.load : ben.run

## Run flow "bb" every time we edited the code, to see the performance improvement
(edit code)
$> ticat ben.scale 10 : bb

## Save a new flow "b10"
$> ticat ben.scale 10 : bb : flow.save b10

(edit code)
$> ticat b10

## Close step-by-step:
$> ticat dbg.step.off : env.save

## Show what "b10" will execute:
$> ticat b10 : help
```

## Run a formal benchmark
Run a formal benchmark on a 3 nodes cluster and create report when the optimization is done:
```bash
## Acquire a server to run benchmark client, login to it
$> ticat host.acquire core=16 : host.clone-ticat-to : host.login

## Acquire cluster hardware resource, deploy cluster, then save cluster info
$> ticat cluster.deploy : env.save
$> ticat hosts.acquire core=8 min-mem=16G cnt=3 : cluster.deploy : env.save

## Run the optimized new code on the new cluster, then release hardware resource
$> ticat bb : ben.report : hosts.release
```
(actually there is a saved flow for formal benchmark, so we just run the command bellow)
```bash
$> ticat ben.scale 10 : ben.base
```

## Investigate a performance issue
Someone talk to us that our system has preformance jitter under a type of specific workload,
So we fetch this workload and test it:
```bash
## Fetch this workload
$> ticat hub.add innerr/workload-x.demo.ticat

## Start local cluster and save info to env
$> ticat cluster.local port=4000 : cluster.start : env.save

## Run benchmark and detect jitter
$> ticat ben.x.scale 10 : ben.x.load : ben.x.run dur=10s : ben.scanner.jitter
```

## Use abbrs
Abbrs or alias could make commands shorter, some common ones could be very useful:
```bash
## Some abbrs:
## e.s == env.save
## c.t == cmds.tree

## Use "cmds.tree" to check the abbrs of a command:
$> ticat cmds.tree env.save
$> ticat m.t e.s
```

## What is a module, how to write it
