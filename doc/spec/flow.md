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

## Command `flow` == `flow.list`
$> ticat f
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
## Overview
$> ticat cmds.tree.simple flow
[flow]
     'list local saved but unlinked (to any repo) flows'
    [save]
         'save current cmds as a flow'
    [set-help-str]
         'set help str to a saved flow'
    [remove]
         'remove a saved flow'
    [list-local]
         'list local saved but unlinked (to any repo) flows'
    [load]
         'load flows from local dir'
    [clear]
         'remove all flows saved in local'
    [move-flows-to-dir]
         'move all saved flows to a local dir (could be a git repo).
          auto move:
              * if one (and only one) local dir exists in hub
              * and the arg "path" is empty
              then flows will move to that dir'
```

Usage examples:
```
## Move saved flows to a local dir
$> ticat hub.move-flows-to-dir <local-dir-path>
$> ticat hub.move <local-dir-path>

## Remove a saved flow
$> ticat flow.rm <command-saved-path>

## Remove all saved flows
$> ticat flow.clear
```
