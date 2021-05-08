# ticat
TiDB wizard

## Target
* Improve the experience and efficiency in non-production scenarios
    * TiDB developing, integration testing, benchmark, POC, etc
* More details: [why-ticat](./doc/why-ticat.md)

## How ticat can achieve that?
* Human friendly
    * Organize job flow with (shell) commands
    * All commands are highly compacted, support fuzzy input, hands on in no time
* Scenario-centered
    * Focus on get things done smoothly in a scenario
* Feature-rich
    * Large amount of modules
        * Components can be easily written in any language
        * ..or from any existing utility by wrapping it up
    * Components' interacting form high-level features
* Write once, run anywhere
    * Save or edit flow easily
    * Share modules and flows easily
* An example: [autotune + benchmark](./doc/usage-draft/benchmark.md)
* More details: [how-ticat-works](./doc/how-ticat-works.md)

## Progess
```
****-  Cli framework
***--      Command line parsing
****-      Full abbrs supporting. TODO: env abbrs from cmds
****-      Env save and load. TODO: save or load from a tag
****-      Self status dumping
***--      Command line help/search
****-  Mod framework
****-      Builtin mod supporting
****-      Bash mod supporting
-----      Mocking execute
-----      Background running
*****      Support more typs: python, golang, dir(repo), etc
-----  Flow framework
***--      Save and edit flow
****-      Connector-env-val checking
-----      Flatten in executing and desc
-----  Hub framework
-----      Mod and flow sharing
-----      Authority control
-----  Scenarios
-----      Benchmark
-----      Integration testing
-----      (TBD)
-----  Components
-----      Tiup cluster operating
-----      Ti.sh cluster operating
-----      Cluster raw backup
-----      Jitter detecting
-----      Simple auto config tuning
-----      Workloads: TPCC
-----      Workloads: sysbench
-----      Workloads: ycsb
-----      (TBD)
```
