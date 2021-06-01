# Basic usage of ticat: build, run commands

### Build

`golang` is needed to build ticat:
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
$> ticat sleep 20s
.zzZZ .................... *\O/*
```

Use abbrs(or aliases) to call a command:
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
$> ticat dbg.echo :+
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

### Use abbrs/aliases

When we are searching commands, the output is like:
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
and the `clear` has an alias `reset`.

So `hub.clear` and `h.reset` are the same command:
```
$> ticat hub.clear
$> ticat h.reset
```
