# [Spec] Flow save/edit

## Save sequecen to a flow / use a saved flow
```
## Save
$> ticat <command-1> : <command-2> : flow.save <command-save-path>
## Call
$> ticat <command-save-path>

## Examples:
$> ticat dummy : dummy : flow.save dummy.2
$> ticat dummy.2
$> ticat sleep 1s : dummy.2 : echo hello
```

## List all manually saved flows
The flows from repos added by "hub.add" or "hub.add.local" will be not listed:
```
$> ticat flow.list
$> ticat f.ls
```

## Add more info to the saved flow
```
## Add help
$> ticat flow.help <command-saved-path>
## Examples:
$> ticat flow.help dummy.2 "just a simple test of flow"

## Add abbrs
(TODO: implement, for now we could manually edit the flow file)
```

## The saved flow files
The saved file dir is defined by env key "sys.paths.flows",
the file name is `<command-path>` plus suffix `.flow.ticat`.

The format is:
* the comment lines start with "#"
* the special comment lines start with "# help = " is for help string definition.
* the special comment lines start with "# abbrs = " is for abbrs definition.
    - abbrs format: `<path-segment>.<path-segment>...`, for each segment: `<abbr-1>|<abbr-2>|...`
* the lines without "#" is the sequence, if there are more than one line, they will concate into one when being executed.

## Flow commands overview
```
## "m.t.s" is "cmds.tree.simple"
$> ticat m.t.s flow
[flow|fl|f|F]
    [save|persist|s|S]
    [set-help-str|help|h|H]
    [remove|rm|delete|del]
    [list-local|list|ls]
    [load|l|L]
    [clear|reset]

## Move saved flows to a local dir
$> ticat hub.move-flows-to-dir <local-dir-path>
$> ticat hub.move <local-dir-path>

## Remove a saved flow
$> ticat flow.rm <command-saved-path>
## Remove all saved flows
$> ticat flow.clear
```
