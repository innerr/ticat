## Compare the performance of two version of tikv
* create a cluster with 3 TiKV instances
* Auto tune the configs.
* Run TPCC bench.
* Switch to another version of TiKV binary, do the same things.
* Output the bench report: QPS, jitter, resource comsumed, etc.
* Remove everything at the end.

```bash
ticat new name=my-cluster host=172.16.5.4,5,6
(input password of these hosts)

ticat tpcc/load thread=16 wh=1000
ticat backup/to name=t1k
ticat tune/tpcc output=master-best
ticat bench/tpcc cfg=master-best thread=100,200,400,800 name=my-bench tag=master

ticat env bin.tikv=./target/release/tikv-server
ticat tune/tpcc output=my-dev-best
ticat bench/tpcc cfg=my-dev-best thread=100,200,400,800 name=my-bench tag=my-dev

ticat tpcc/report name=my-bench
(report displaying)

ticat burn : remove
```
