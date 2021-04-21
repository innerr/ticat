# Why We Need ticat

## Non-production Cluster Matters
There are lots of production TiDB clusters, which will cause huge loss if some were down.
But there are also lots of other clusters not that serious.
We run these clusters all kind of activities, such as developing, testing, benchmark, POC, etc.

A fact we may ignored is: the quality of those non-prod clusters define the quality of the online ones.

Hence non-prod clusters matter: the quality of them, the amount of them, the diversity of them, and so on.

## The Improvable Situation: Expensive to Run
The tool tiup provide the essencial core functions to maintain a cluster,
It's nice to use it on online clusters.
However, tiup still too expensive to run on non-prod clusters.

### The Human Adapter
When running a cluster, we are not just launch it, normally there is a job flow running on it.
A job flow is a step-by-step routine, as we are programmers, it should be codes or scripts by nature.

But in most cases it's not, we execute the job flow by hand, take benchmark as an example:
* We create and launch the cluster.
* We observe the 'start' result, when it's finished, we check the cluster is good to go on Grafana.
* We load data into the cluster.
* We take a try-run, during the running we check write-stall, check QPS, check jitter, check resource comsuming.
* Then we try to tune some config entries to improve the performance.
* ...

We involed deeply in the job flow, it's like the script is written in our brain, and we are human interpreters executing the script.
There is a pattern here:
* We execute a step by running a module(a tool or a tiup command).
* Then we observe the output, then do some 'if-else' in our brain.
* Then we decide some arguements, execute a module again.

Here we are the human adapter connected the exist modules: do-run-with-input, parse-output, do-if-else

It's easy to see that the reasons why a job flow is not a script:
* Lack of some modules, eg, tools to check write-stall, check QPS trend.
* The modules we have now are independent bottom-level ones, one hard to call another one.

That makes it expensive to run job flows, both in manpower aspect and hardware resource aspect.
Thus compromise the total quality of non-prod clusters.

### The factured knowledge

### Too much routine works

