# Usage

## Examples
Build ticat:
```
$> git clone https://github.com/innerr/ticat
$> cd ticat
$> make
```
Recommand to set `ticat/bin` to system `$PATH`, it's handy.

### Run a command
Run a simple command, it does not thing but print a message:
```
$> ticat dummy
dummy cmd here
```
Pass arg(s) to a command, `sleep` will pause for `3s` then wake up:
```
$> ticat sleep 3s
.zzZZ ... *\O/*
```
Defferent ways to pass arg(s) to a command:
```
$> ticat sleep duration=3s
$> ticat sleep {duration=3s}
```
Use abbrs(or alias) to call a command:
```
$> ticat slp 3s
$> ticat slp dur=3s
$> ticat slp d=3s
```

### Run a command in the command-tree
All commands are organized to a tree,
the `sleep` and `dummy` commands are under the `root`,
so we could call them directly.

Another two commands does nothing:
```
$> ticat dummy.power
power dummy cmd here
$> ticat dummy.quiet
quiet dummy cmd here
```
Do notice that `dummy` `dummy.power` `dummy.quiet` are totally different commands,
they are in the same command-branch just because users can find related commands easily in this way.

Display a command's info:
```
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
```
$> ticat dbg.echo hello
$> ticat dbg.echo "hello world"
$> ticat dbg.echo msg=hello
$> ticat dbg.echo m = hello
$> ticat dbg.echo {M=hello}
$> ticat dbg.echo {M = hello}
```

Browse the command tree:
```
$> ticat cmds.tree
```
The output result will be a lot, below is one of the command:
```
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
```
$> ticat cmds.tree dbg
...
```

Or view the tree in flatten mode (both are the same):
```
$> ticat cmds.flat
$> ticat cmds.ls
```

The most useful is searching support:
```
$> ticat cmds.ls echo
[echo]
     'print message from argv'
    - full-cmd:
        dbg.echo
    - args:
        messsage|msg|m|M = ''
```

`help` or `find` can search anything, so apparently can use for finding commands:
```
$> ticat help echo
$> ticat find echo
```

### Run command sequences
Sequences are like unix-pipe, but use `:` instead of `|`:
```
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


