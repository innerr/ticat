# ticat
TiDB wizard

## Target
* Improve the experience and efficiency in non-production scenarios
    * TiDB developing, integration testing, benchmark, POC, etc
* More details: [why-ticat](./doc/why-ticat.md)

## How ticat can achieve that?
* Human friendly
    * Organize job flow with (shell) commands
    * All commands are highly compacted, support fuzzy input, hands on in on time
* Scenario-centered
    * Focus on get things done smoothly in a scenario
* Feature-rich
    * Large amount of modules
        * Components can be easily written in any language
        * ..or from any existing utility by wrapping it up
    * Components' interacting form high-level features
* An example: [autotune + benchmark](./doc/usage-draft/benchmark.md)
* More details: [how-ticat-works](./doc/how-ticat-works.md)

## Progess
* Design
```
****-  The concept of ticat
***--  CLI framework
**---  Component framework
-----  Scenarios
*----    Benchmark
-----    Integration testing
-----    (TBD)
*----  Components
***--    Cluster
****-    Raw backup
-----    (TBD)
```
* Implement
```
*----  CLI framework
-----  Component framework
-----  Scenarios
-----    Benchmark
-----    Integration testing
-----    (TBD)
-----  Components
-----    Cluster
-----    Raw backup
-----    (TBD)
```
