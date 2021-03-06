# Ticat: workflow automating in unix-pipe style


## The problem: workflow are complicated with multi-demension requirements
Let's take a distributed system as example:
```
 ┌───────────────────────┐           ┌────────────┐
 │ A Distributed System  ◄───────────┤ Users      │
 └───────────────▲───────┘           └────────────┘
                 │                   ┌────────────┐
                 └───────────────────┤ Developers │
                                     └────────────┘
```


During time, the system grows big, many peripheral tools are developed.
The intergration jobs become complicated to adapt the multi-dimension requirements.
```
 ┌───────────────────────┐           ┌────────────┐
 │ The Core System       ◄───────┬───┤ Users      │
 └─┬─────────────────────┘       │   └────────────┘
   │                   ...       │
   │        ┌────────────┐       │
   ├────────┤ Operatings ◄───────┘
   │        └────┬─┬─────┘
   │        ┌────┴─┴─────┐           ┌────────────┐
   ├────────┤ Testing    ◄───────┬───┤ Developers │
   │        └──┬─┬─┬─────┘       │   └────────────┘
   │        ┌──┴─┴─┴─────┐       │
   └────────┤ Benchmark  ◄───────┤
            └──┬─┬─┬─────┘       │
               │ │ │   ...       │
           ┌───┴─┴─┴─────────────▼────────┐
           │                              │
           │  Intergration System/Tools   │
           │                              │
           └──────────────────────────────┘
```


## The cure of complication: break down by orthogonal dimensions
The unix philosophy inspired us, `Simple parts that work together`, like the unix-pipe:
```
$> cat my.log | grep ERR | awk -F 'reason' '{print $2}'
```


## **Ticat**: makes sure the parts can be easily built
The ad-hot feature assembling give us the most flexable yet powerful controlling.


To apply this, in **ticat** we provide a easy way to wrap any existed tools into components(alias: modules):
```
 ┌───────────────────────┐        ┌──────────────────┐
 │ Alloc Server Resource │        │ Jitter Detecting │
 └────────────────┬──────┘        └───────┬──────────┘
                  │                       │
 ┌─────────────┐  │             ┌─────────┴────────────┐
 │ Auto Config │  │             │ Benchmark Workload X │
 └──────────┬──┘ ┌┴──────────┐  └───┬─────┬────────────┘
            │    │ Auto Tune │      │     │
 ┌──────────┴─┐  └┬─────┬────┘      │     │
 │ Deployment │   │     │           │     │
 └────┬─────┬─┘   │     │   ┌───────┴─────┴────────┐
      │     │     │     │   │ Benchmark Workload Y │
      │     │     │     │   └─┬─────┬──────────────┘
      │     │     │     │     │     │     │
 ┌────┼─────┼─────┼─────┼─────┼─────┼─────┼────────────┐
 │    │     │     │     │     │     │     │            │
 │  ┌─▼─┐ ┌─▼─┐ ┌─▼─┐ ┌─▼─┐ ┌─▼─┐ ┌─▼─┐ ┌─▼─┐          │
 │  │ A │ │ B │ │ C │ │ D │ │ E │ │ F │ │ G │          │
 │  └───┘ └───┘ └───┘ └───┘ └───┘ └───┘ └───┘          │
 │                                       Ticat Modules │
 └─────────────────────────────────────────────────────┘
```


## **Ticat**: easily share the parts, and share the assembled workflows
Sometimes the pipelined-command could be long,
in **ticat** we could save the pipeline(alias: flow) in a heartbeat,
then the saved one could be used in other flow as a new command.


And the most important, we could easily distribute the code so our workmate could get and run it in no time.
```
 ┌──────────────────────────────────────────────────────────────┐
 │         Modules                             Flows            │
 │  ┌───┐ ┌───┐ ┌───┐ ┌───┐        │┼┼    ┌──────────────────┐  │
 │  │ A │ │ B │ │ C │ │ D │        │┼┼────► C->B->A->D       │  │
 │  └───┘ └───┘ └───┘ └───┘        │┼┼    └──────────────────┘  │
 │  ┌───┐ ┌───┐ ┌───┐              │┼┼    ┌──────────────────┐  │
 │  │ E │ │ F │ │ G │  ...         │┼┼────► C->B->A->D->X->G │  │
 │  └───┘ └───┘ └───┘              │┼┼    └──────────────────┘  │
 │                                 │┼┼    ┌──────────────────┐  │
 │ │┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼│       │┼┼────► C->B->A->F       │  │
 │ └─────────┼─────────────┘       │┼┼    └──────────────────┘  │
 │           │                     │┼┼    ┌──────────────────┐  │
 │           └──────────────►      │┼┼────► X->G->Y->G->..   │  │
 │                                 │┼┼    └──────────────────┘  │
 │ Ticat     ▲                                                  │
 └───────────┼────────────────────────────────┼─────────────────┘
             │ Assemble,                      │ Get and run
             │ Share                          │
             │                       ┌────────▼─────────┐
    ┌────────┴────────┐              │ Other Developers ├─┐
    │ Some Developers │              └─┬────────────────┘ │
    └─────────────────┘                └──────────────────┘
```


## How these parts could work together in **ticat**?
In a unix-pipeline,
the up-stream command and the down-stream command shared an anonymous pipe,
the former is data provider and the latter is consumer.

In **ticat** things are alike, modules run on a shared **env**, a key-value set.

Modules could get any info from **env**(mostly provide by up-stream),
so they are self-sufficient thus can be flexably assembled.

A module need to register which keys it wants to read or write,
**ticat** will check the independency.
**Ticat** also provide a cli gramma for user to manipulate key-values in any time.
```
 ┌─────────────────────────────┐
 │                       Ticat │
 │                             │
 │   ┌─────────────────────┐   │
 │   │ key1 = val1     Env │   │ Manipulate
 │   │ key2 = val2         ◄────────────────┐
 │   │ ...                 │   │            │
 │   │                     │   │            │
 │   └─▲─────▲─────▲─────▲─┘   │            │
 │     │     │     │     │     │            │
 │   ┌─▼─┐ ┌─▼─┐ ┌─▼─┐ ┌─▼─┐   │            │
 │   │ A │ │ B │ │ C │ │ D │   │            │
 │   └─▲─┘ └─▲─┘ └─▲─┘ └─▲─┘   │            │
 │     │     │     │     │     │     ┌──────┴───────┐
 │   ┌─┴─────┴─────┴─────┴─┐   │     │ Ticat Users  │
 │   │ C->B->A->D          ◄─────────┤ (Developers) │
 │   └─────────────────────┘   │ Run └──────────────┘
 │                             │
 └─────────────────────────────┘
```


## Practice **ticat** with zero cost
Any existed tools can be wrapped into modules in a small cost,
so putting **ticat** into use is quick and easy.

**Ticat** does't intrude any framework we are using,
can be appled to a small area at the beginning,
then gradually use it to improve the whole system.
```
 ┌───────────────────────┐           ┌────────────┐
 │ Cluster Core System   ◄───────┬───┤ Users      │
 └─┬─────────────────────┘       │   └────────────┘
   │                   ...       │
   │        ┌────────────┐       │   ┌────────────┐
   ├────────┤ Operatings ◄───────┘   │ Developers │
   │        └────┬─┬─────┘           └──┬─────────┘
   │        ┌────┴─┴─────┐           ┌──┴─────────────────────┐
   ├────────┤ Testing    ◄───────┬───┤ Ticat: Full Automation │
   │        └──┬─┬─┬─────┘       │   ├────────────────────────┤
   │        ┌──┴─┴─┴─────┐       │   │┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼│
   └────────┤ Benchmark  ◄───────┤   │┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼│
            └──┬─┬─┬─────┘       │   │┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼│
               │ │ │   ...       │   └────────────────────────┘
           ┌───┴─┴─┴─────────────▼──┐
           │  Intergration Tools    │
           └────────────────────────┘
```
