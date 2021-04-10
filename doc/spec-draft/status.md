## Status
```
$> ticat [status]
#appear when have more than on cluster#
clusters:
    cluster-1[current, running] cluster-2[running] culster-3[stopped]

cluster-1:
    PD * 1: good
    TiKV * 3: good
    TiDB * 3: good
job:
    => start at 2021-04-01_00:00
    => config auto tuning on TPCC
    tikv.scheduler-pool-size done, best = 15
    tikv.raft-store.store-pool-size,apply-pool-size: done, best = 4,4
workload:
    => started at 2021-04-01_00:00
    => TPCC
    QPS trend: decrease -> stable [42K, 45K, 43K, 40K, 40K, 40K]
    latency 99th and jitter (latest 1H):
        TiDB         200ms     ±05%
        TiKV-write    50ms     ±18%
        TiKV-read     35ms     ±48%
evenness:
    avg host CPU:       60%  min 20% at [172.5.5.4], max 98% at [172.5.5.7]
    avg TiDB CPU:       40%  ±20%
    avg TiKV CPU:       80%  ±56%
    avg TiDB QPS:       40K  ±20%
    avg TiKV QPS:       80K  ±56%
    avg storge used:    80%  ±78%
    avg disk IOBW:      55%  ±60%
    avg regions:        12K  ±04%
    avg region leaders:  4K  ±02%

$> ticat status
...
$> ticat status even
eventness:
    disk space:
        80% avg used
        30% min used at [172.16.5.4]
        99% max used at [172.16.5.8]
    ...
$> ticat status job
$> ticat status workload
```
