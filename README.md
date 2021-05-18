# ticat
A casual command line components platform

## Examples
Build ticat:
```bash
$> git clone https://github.com/innerr/ticat
$> cd ticat
$> make
```
Recommand to set `ticat/bin` to system `$PATH`, it's handy.

### Run a command
Run a simple command, it does not thing but print a message:
```bash
$> ticat dummy
dummy cmd here
```
Pass arg(s) to a command, `sleep` will pause for `3s` then wake up:
```bash
$> ticat sleep 3s
.zzZZ ... *\O/*
```
Defferent ways to pass arg(s) to a command:
```bash
$> ticat sleep duration=3s
$> ticat sleep {duration=3s}
```
Use abbrs(or alias) to call a command:
```bash
$> ticat slp 3s
$> ticat slp dur=3s
$> ticat slp d=3s
```

### Run a command in the command-tree
All commands are organized to a tree,
the `sleep` and `dummy` commands are under the `root`,
so we could call them directly.

Another two commands does nothing:
```bash
$> ticat dummy.power
power dummy cmd here
$> ticat dummy.quiet
quiet dummy cmd here
```
Do notice that `dummy` `dummy.power` `dummy.quiet` are totally different commands,
they are in the same command-branch just because users can find related commands easily in this way.

Display a command's info:
```bash
$> ticat cmds.tree dbg.echo
[echo]
     'print message from argv'
    - full-cmd:
        dbg.echo
    - args:
        message|msg|m|M = ''
```
From this we know that `dbg.echo` has an arg `message`, this arg has some abbrs: `msg` `m` `M`.

Different ways to call the command:
```bash
$> ticat dbg.echo hello
$> ticat dbg.echo "hello world"
$> ticat dbg.echo msg=hello
$> ticat dbg.echo m = hello
$> ticat dbg.echo {M=hello}
$> ticat dbg.echo {M = hello}
```

Browse the command tree:
```bash
$> ticat cmds.tree
```
The output result will be a lot, below is one of the command:
```bash
...
    [hub|h|H]
        [clear|reset]
             'remove all repos from hub'
            - full-cmd:
                hub.clear
            - full-abbrs:
                hub|h|H.clear|reset
...
```
From this we know a command `hub.clear`,
the name `hub` has abbrs `h` `H`,
and the `clear` has an alias `reset`,
so `hub.clear` and `h.reset` are the same thing:

We could also browse a specific branch:
```bash
$> ticat cmds.tree dbg
...
```

Or view the tree in flatten mode (both are the same):
```bash
$> ticat cmds.flat
$> ticat cmds.ls
```

The most useful is searching support:
```bash
$> ticat cmds.ls echo
[echo]
     'print message from argv'
    - full-cmd:
        dbg.echo
    - args:
        messsage|msg|m|M = ''
```

`help` or `find` can search anything, so apparently can use for finding commands:
```bash
$> ticat help echo
$> ticat find echo
```

### Run command sequences
Sequences are like unix-pipe, but use `:` instead of `|`:
```bash
$> ticat dummy : sleep 3s : dummy
+-------------------+
| stack-level: [1]  |             05-18 18:51:47
+-------------------+----------------------------+
| >> dummy                                       |
|    sleep                                       |
|        duration = 3s                           |
|    dummy                                       |
+------------------------------------------------+
dummy cmd here
+-------------------+
| stack-level: [1]  |             05-18 18:51:47
+-------------------+----------------------------+
|    dummy                                       |
| >> sleep                                       |
|        duration = 3s                           |
|    dummy                                       |
+------------------------------------------------+
.zzZZ ... *\O/*
+-------------------+
| stack-level: [1]  |             05-18 18:51:50
+-------------------+----------------------------+
|    dummy                                       |
|    sleep                                       |
|        duration = 3s                           |
| >> dummy                                       |
+------------------------------------------------+
dummy cmd here
```
The boxes indicate the running command by `>>`

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
