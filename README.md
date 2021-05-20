# ticat
A casual command line components platform

## The things we want to solve
Let's take a distributed system as example:
```
 ┌───────────────────────┐           ┌────────────┐
 │ A Distributed Cluster ◄───────────┤ Users      │
 └───────────────────────┘           └────────────┘
```

During time, it grows to a eco-system, huge and complicated:
```
 ┌───────────────────────┐           ┌────────────┐
 │ Cluster Core System   ◄───────┬───┤ Users      │
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
           │  Intergration Tools          │
           │                              │
           └──────────────────────────────┘
```
User pain

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
 │                                       ticat Modules │
 └─────────────────────────────────────────────────────┘
```

```
 ┌────────────────────────────────────────────────────────────┐
 │                                           Flows            │
 │  ┌───┐ ┌───┐ ┌───┐ ┌───┐             ┌──────────────────┐  │
 │  │ A │ │ B │ │ C │ │ D │        ┌────► C->B->A->D       │  │
 │  └───┘ └───┘ └───┘ └───┘        │    └──────────────────┘  │
 │  ┌───┐ ┌───┐ ┌───┐              │    ┌──────────────────┐  │
 │  │ E │ │ F │ │ G │  ...         ├────► C->B->A->D->X->G │  │
 │  └───┘ └───┘ └───┘              │    └──────────────────┘  │
 │ └─────────┬─────────────┘       │    ┌──────────────────┐  │
 │           │                     ├────► C->B->A->F       │  │
 │           │                     │    └──────────────────┘  │
 │           │                     │    ┌──────────────────┐  │
 │           └────────────────►    ├────► X->G->Y->G->..   │  │
 │                                 │    └──────────────────┘  │
 │                                 ...                        │
 │           ▲                                                │
 └───────────┼──────────────────────────────┼─────────────────┘
             │ Assemble                     │ Get and run
             │ Share                        │
             │                     ┌────────▼─────────┐
    ┌────────┴────────┐            │ Other Developers ├─┐
    │ Some Developers │            └─┬────────────────┘ │
    └─────────────────┘              └──────────────────┘
```


```
 ┌───────────────────────┐           ┌────────────┐
 │ Cluster Core System   ◄───────┬───┤ Users      │
 └─┬─────────────────────┘       │   └────────────┘
   │                   ...       │
   │        ┌────────────┐       │   ┌────────────┐
   ├────────┤ Operatings ◄───────┘   │ Developers │
   │        └────┬─┬─────┘           └──┬─────────┘
   │        ┌────┴─┴─────┐           ┌──┴─────────────────────┐
   ├────────┤ Testing    ◄───────┬───┤ ticat: full automative │
   │        └──┬─┬─┬─────┘       │   ├────────────────────────┤
   │        ┌──┴─┴─┴─────┐       │   │┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼│
   └────────┤ Benchmark  ◄───────┤   │┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼│
            └──┬─┬─┬─────┘       │   │┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼┼│
               │ │ │   ...       │   └────────────────────────┘
           ┌───┴─┴─┴─────────────▼──┐
           │  Intergration Tools    │
           └────────────────────────┘
```

## How
## quick-start
## Examples
## Spec
## zen

## Target
* Human friendly
    * Easy to understand: lots of features, but well-organized (commands, env, etc)
    * Zero memorizing presure: good searching and full abbrs support
* Rich features
    * Easy to get lots of modules
        * Components can be easily written in any language
        * ..or from any existing utility by wrapping it up (in no time)
    * Easy and powerful configuring
        * Modules are automatically work together, by running on a shared env
        * Anything can be configured by modifying the env
    * Combine modules to flow
* Easy to share context, or to run others'
    * Use github repo-tree to distribute code
    * Share modules and flows easily by adding a top-repo
