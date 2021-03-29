# Examples

## Compare performance of two version of tikv by auto benchmark
* create a cluster with 3 TiKV nodes from the very beginning
* then Auto tune the configs.
* then run tpcc benchmarks.
* switch to another version of tikv, do tune and bench again.
* output the bench report: QPS, jitter, resource comsumed, etc.
* remove everything at the end.
```
ticat new name=my-cluster host=172.16.5.4,5,6 port=+300
(input password of these hosts)

ticat tpcc/load thread=16 wh=1000 : backup/to t1k
ticat tune/tpcc output=master-best : bench/tpcc cfg=master-best thread=100,200,400,800 name=my-bench tag=master
ticat bin tikv=/data/tikv/target/release/tikv-server
ticat tune/tpcc output=my-dev-best : bench/tpcc cfg=my-dev-best thread=100,200,400,800 name=my-bench tag=my-dev
ticat tpcc/report name=my-bench : burn : remove
```
