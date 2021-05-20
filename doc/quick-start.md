# Quick start

## Build
Build ticat: (`golang` is needed)
```
$> git clone https://github.com/innerr/ticat
$> cd ticat
$> make
```
Recommand to set `ticat/bin` to system `$PATH`, it's handy.

## Run
Run a command sequence `dummy : dummy : dummy`:
```
$> ticat dummy : dummy : dummy
+-------------------+
| stack-level: [1]  |             05-19 01:03:38
+-------------------+----------------------------+
| >> dummy                                       |
|    dummy                                       |
|    dummy                                       |
+------------------------------------------------+
dummy cmd here
+-------------------+
| stack-level: [1]  |             05-19 01:03:38
+-------------------+----------------------------+
|    dummy                                       |
| >> dummy                                       |
|    dummy                                       |
+------------------------------------------------+
dummy cmd here
+-------------------+
| stack-level: [1]  |             05-19 01:03:38
+-------------------+----------------------------+
|    dummy                                       |
|    dummy                                       |
| >> dummy                                       |
+------------------------------------------------+
dummy cmd here
```
The command `dummy` sequencely executed 3 times,
It's a little bit like unix-pipe `|`, but different:
* use `:` to concatenate commands, not `|`
* execute commands one by one, the second one won't start untill the previous one finishes
* the `>>` in the box indicate the current command about to run

## Create a module
In a empty dir (say the path is `mymods`), create a bash script `mod-1.bash`:
```bash
# skip the first special arg
shift

# get args
name=$1
age=$2

# print args
echo "Got: name=$name, age=$age"
```

Then create a meta file `mod-1.bash.ticat` for this script in the same dir:
```
help = a simple ticat module

[args]
name = John
age = unknown
```
The `args` section defines the arg-names and default values.

Done! you just wrote a ticat module.

Add dir `mymods` to ticat so it could find the module:
```bash
$> ticat hub.add.local path=mymods
```

Let's run it:
```bash
$> ticat mod-1
Got: name=John, age=unkown
$> ticat mod-1 age=35
Got: name=John, age=35
$> ticat mod-1 Alex 6
Got: name=Alex, age=6
```

## Create more modules and put them working together
Create a new module `mod-w.bash` in the same dir:
```bash
# get env file path from the first special arg
env=$1/env
shift

# get message from arg
msg=$1

# pass message to env, format: "key \t value"
echo -e "mymsg\t$msg" >> $env
echo "Sent: $msg"
```

The related meta file `mod-w.bash.ticat`:
```
help = get message from arg then pass it to env

[args]
msg = ''

[env]
mymsg = write
```
The `env` section declares this module will write the env key `mymsg`.

Create another module `mod-r.go` in the same dir,
we use `golang` to show the cross-language ability:
```go
package main

import (
        "bufio"
        "os"
        "strings"
)

func main() {
        env := os.Args[1] + "/env"
        file, _ := os.Open(env)
        defer file.Close()
        scanner := bufio.NewScanner(file)
        kvs := map[string]string{}
        for scanner.Scan() {
                text := scanner.Text()
                i := strings.Index(text, "\t")
                if i > 0 {
                        kvs[text[0:i]] = text[i+1:]
                }
        }
        println("Recv:", kvs["mymsg"])
}
```

This module will read the env key `mymsg`,
the meta file `mod-r.go.ticat` is:
```
help = receive message from env

[env]
mymsg = read
```

Since dir `mymods` already added to ticat before,
so we could directly run the modules:
```
$> ticat mod-w hello : mod-r
+-------------------+
| stack-level: [1]  |             05-19 02:34:24
+-------------------+----------------------------+
| >> mod-w                                       |
|        msg = hello                             |
|    mod-r                                       |
+------------------------------------------------+
Sent: hello
+-------------------+
| stack-level: [1]  |             05-19 02:34:24
+-------------------+----------------------------+
|    mymsg = hello                               |
+------------------------------------------------+
|    mod-w                                       |
|        msg = hello                             |
| >> mod-r                                       |
+------------------------------------------------+
Recv: hello
```

Test the dependency checking:
```
$> ticat mod-r
[checkEnvOps] cmd 'mod-r' reads 'mymsg' but no provider
```
When we run `mod-r` without `mod-w`, there will be an error.

## Share your code
Push the dir `mymods` to github as a repo,
let's say your github id is `aCoolName`,
tell your mates to add your repo:
```bash
$> ticat hub.add aCoolName/mymods
```
Then they could use the modules you just wrote:
```bash
$> ticat mod-w : mod-r
...
```

We already prepared the code above in a [repo](https://github.com/innerr/quick-start.ticat),
You may need to disable yours to avoid name-conflicting:
```
$> ticat hub.disable mymods
```

Then fetch and run the demo:
```
$> ticat hub.add innerr/quict-start.ticat
$> ticat mod-w : mod-r
...
```

## How this could help us?
With this `env-read-write` loose-connecting,
we could break complicated projects into small parts,
or in another point-of-view,
we could glue existed tools to achieve powerful features.

Take an example,
a distributed cluster is deployed,
then the basic info will keep in env:
hosts, ports, names, configs, etc.

Then when a benchmark tool is run,
it could fetch all the needed info from env.
-- yes, some args like `data-scale` or `threads` still need to pass to the tool,
but those args could also put into env.
In a word, ticat totally seperated `logic` and `config`.

With this design,
We are able to create lots of self-sufficient modules,
present them to end-users in a `low code` style: command sequences.

Systems become flexable with ticat:
ad-hot feature assembling,
cut out unnecessary modules,
use different provider in different hardware, etc.

Furthermore, sequences could be save into `flow`,
```
$> ticat mod-w : mod-r : flow.save my-flow
(saved to my-flow)
$> ticat my-flow
(execute "mod-w : mod-r")
```
end-users could share and get it along a repo,
`no code` yet still have full control on everything.

With power comes burden,
end-users need to know lots of things.
So ticat put on a lot works on this,
to reduce memorizing pressure:
* use tree to organize info: commands, env, repos, etc
* full search supporting: commands, env, connecting-points, etc
* abbrs/alias supporting, eg, a command `mods` will have an alias `mod`

The best thing is,
we don't need to change anything to adapt ticat,
we only need a few minutes to wrap an existed tool into a ticat module.

:)
