# [Spec] Module: env interaction with ticat

## Receive env
A module is en executable file,
the first arg "arg-0" will be it's own path or name as normal cases.

The second arg "arg-1" will be an dir stored the ticat session info,
The file "arg-1"/env is a file contains all the env values.
The format is multi lines, each line is a key-value pair, seperated by "\t".

The rest of args will be the normal args defined by ".ticat", in order.

## Change env
When a module want to change env, it could simply append key-value to the env file,
in the same "key \t value" format.

## Call other ticat modules
A env key "sys.paths.ticat" defined the binary path of **ticat**,
so it's easy to call other modules.

We need to notify ticat it's in a same session,
so all the modules could use the same env.
That's how we should do to call other modules inside a module:
```
$> ticat {session=<arg-1>} <any-ticat-command>
$> ticat {session=<arg-1>} : <command-1> : <command-2>
```
