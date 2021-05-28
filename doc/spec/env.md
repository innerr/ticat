# [Spec] Env get/set/save

## Display/find env values
```
$> ticat env.list
$> ticat env.list <find-str>
```

## Define(set) env key-values
```
## Set value
$> ticat {key=value}
## Examples:
$> ticat {display.width=40}
$> ticat {display.width=40} : env.list

## Set value between segments of a command:
$> ticat {key=value}

## Set value between segments of a command:
$> ticat <command-segment>{key=value}.<command-segment>
## Example:
$> ticat env{mykey=666}.ls mykey
env.mykey = 666
```

Extra space chars (space and tab) will be ignore:
```
$> ticat {display.width = 40}
```

## Save env key-values
The command "env.save" will persist any changes in the env to disk,
then use the values in later executing:
```
## Display the default value of "display.width"
$> ticat env.list width
display.width = 80

## Change the value but not save
$> ticat {display.width=40} env.list width
display.width = 40

## Save the value
$> ticat {display.width=60} env.save
display.width = 60

## Show the changed and saved value
$> ticat env.list width
display.width = 60
```

## Sessions
Each time ticat is run, there is a session.
When the execution finishes, the session ends.
```
$> ticat (session start) <command> : <command> : <command> : <command> (session end)
```

In a session, there is an env instance,
all key-value changes between "start" and "end" will be in this env.
When the session ends, all changed will be lost if "env.save" is not called.

If a command call other ticat commands,
then they all in the same session (share the same key-value set):
```
## In this sequence, the value of "display.width" for <commannd-x> also is "66"
$> ticat {display.width=66} <command-a> : <command-b-which-will-call-command-x> : <command-c>
```

## Env layers
Env has multi-layers, when getting a value, will find in the first layer,
if the key is found then return the value.
If the key is not found then looking in the next layer.
```
  command layer    - the first layer
  session layer
persisted layer
  default layer    - the last layer
```

Meanings of each layer:
```
  command layer    - the key-values only for this current command in the sequence
  session layer    - the key-values for the whole sequence
persisted layer    - the key-values from env.saved, for the whole sequence
  default layer    - the default values, hard-coded
```

Display each layer:
```
## Display all:
$> ticat env.tree

## Display in sequence execution:
$> ticat {display.layer=true} dummy : dummy
$> ticat {display.layer=true} dummy : {example-key=its-command-layer} dummy

## Save the display flay:
$> ticat {display.layer=true} env.save
## Then every sequence executions will display layers:
$> ticat dummy: dummy
```

## Difference of the command layer and the session layer
If the key-values settings has ":" in front of them, they are in command layer.
```
## This is a command layer key-value, will only affect the 2nd dummy command:
$> ticat dummy : {example-key=its-command-layer} dummy

## This is a session layer key-value, will only affect the 2nd dummy command:
$> ticat {example-key=its-session-layer} dummy : dummy

## Demonstrate:
$> ticat {display.width=40} dummy : dummy
$> ticat dummy : {display.width=40} dummy
```

If a command changes env in it's code (not in cli), the changes will be in the session layer:
```
## The "display.width" value for <command-2> will be a changed value
$> ticat <command-1-which-changes-display-width> : <command-2>
```

## The saved env file
The saved file dir is defined by env key "sys.paths.data",
the file name is defined by env key "strs.env-file-name".

The format is multi lines,
each line is a key-value pair seperated by a string defined by env key "strs.proto-sep"(default: \t).
