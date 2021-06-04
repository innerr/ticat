# Zen: the choices we made

## Why the usage so weird?

Q: the usage feels weird, why?

There are some necessary design, it's unfair to call them weird:
* Run command sequence in unix-pipe style, just use `:` instead of `|`.
* Use `{key=value}` to set values to the shared env.

That all, nothing special.
Maybe we could call the `foo.bar` form command name weird,
there is another page to explain that.

Other than that, all things "weird" are extra features,
**we could totally ignore them as we like**.

## The "power" commands may make us feel weird

This command maybe a little weird:
```
$> ticat desc : dummy : dummy
--->>>
[dummy]
     'dummy cmd for testing'
[dummy]
     'dummy cmd for testing'
<<<---
```

From the usage doc we know `desc` is for display how this sequence will execute, and it did.

But as in unix-pipe style, `dummy` should run, why not? that's weird.

The reason is, `desc` is a builtin "power" command,
"power" command have the mechanism to access and modify the sequence.
`desc` will show the info and then remove all commands, so nothing will run.

This mechanism is useful for **ticat** core to adding features to executor.
For example, we could write a "mockup" command, to replace specific commands with "mockup commands".

We admit that it's a little unusual, but it's worth it.

## The "priority" commands may make us feel weird

These two commands have the same result:
```
$> ticat desc : dummy : dummy
$> ticat dummy : dummy : desc
--->>>
[dummy]
     'dummy cmd for testing'
[dummy]
     'dummy cmd for testing'
<<<---
```

The reason is `desc` have a "priority" flag,
the "priority" commands will be move to the front of the sequence.

If there are more than one priority commands in a sequence,
their executing order will follow their appear order.
(not a good thing, we consider adding a priority-level to command properties)

So, below the first `desc` will reveal the real executing order:
```
$> ticat desc: dummy:dummy:desc
--->>>
[desc]
     'desc the flow about to execute'
    - cmd-type:
        power (quiet) (priority)
[dummy]
     'dummy cmd for testing'
[dummy]
     'dummy cmd for testing'
<<<---
```

This looks extremely nonsense, but trust us, it's handy.

Suppose we are typing a sequence:
```
$> ticat dummy:dummy:dummy:dummy:dummy:dummy
```

We want to take a preflight check before executing.

If we don't use the priority featue,
(like we just said, it's extra, we could pretend it's not exists)
We need to move the cursor to the front, type "desc:", then hit "enter":
```
$> ticat desc: dummy:dummy:dummy:dummy:dummy:dummy
```

After checking, we roll up command line history, move cursor, edit, then execute.

(roll up twice then execute the history? no it can't be done,
because this sequence never execute so not in the history)

In a terminal which the os variable "$PS1" not properly setup, or in WSL,
history editing will cause incorrect display.

Now do it again with the priority featue, things are a little easier.
Do preflight check by append ":desc":
```
$> ticat dummy:dummy:dummy:dummy:dummy:dummy :desc
```
Then roll up history, delete ":desc", hit "enter", done.

## The "weird" `+` and `-`

As a small platform running on the user-end, **ticat** has lots of infos.
To show different infos we have different commands such like `cmds.list` `desc`.

We need a single command to cover all frequent info queries, to make **ticat** easy to use.

At first, command `help` is created, abbr `?`.
But we realize that one command is not enough,
it either show too much or too little.
Bisides, `?` is intercepted by "zsh".

So commands `more` and `less` are here to do the "help" job, abbrs `+` `-`.

They works so far so good, show infos base on args and sequence they are in.
Here is their info:
```
$> ticat +:-
[less|-]
     'display brief info base on:
      * if in a sequence having
          * more than 1 other commands: show the sequence execution.
          * only 1 other command and
              * has no args and the other command is
                  * a flow: show the flow execution.
                  * not a flow: show the command or the branch info.
              * has args: find commands under the branch of the other command.
      * if not in a sequence and
          * has args: do global search.
          * has no args: show global help.'
    - args:
        1st-str|find-str = ''
        2nd-str = ''
        3rh-str = ''
        4th-str = ''
        5th-str = ''
        6th-str = ''
    - cmd-type:
        power (quiet) (priority)
    - from:
        builtin
```
```
$> ticat -:+
[more|+]
     'display rich info base on:
      * if in a sequence having
          * more than 1 other commands: show the sequence execution.
          * only 1 other command and
              * has no args and the other command is
                  * a flow: show the flow execution.
                  * not a flow: show the command or the branch info.
              * has args: find commands under the branch of the other command.
      * if not in a sequence and
          * has args: do global search.
          * has no args: show global help.'
    - args:
        1st-str|find-str = ''
        2nd-str = ''
        3rh-str = ''
        4th-str = ''
        5th-str = ''
        6th-str = ''
    - cmd-type:
        power (quiet) (priority)
    - from:
        builtin
```

Alternately using these pair of commands give us excellent experience browsing infos.
Let's check it out.

Global searching:
```
## Find commands:
$> ticat - <find-str>

## Find commands base on the previous result:
$> ticat - <find-str> <find-str> <find-str>

## Continue filtering the result:
$> ticat - <find-str> <find-str> <find-str>

## Switch to "+" to show details in a small result set:
$> ticat - <find-str> <find-str> <find-str>
```

Investigate a command branch:
```
## Investigate command branch:
$> ticat <command> :-

## Investigate command from the branch:
$> ticat <command> :+

## Investigate more command from the branch:
$> ticat <command>.<command> :+
```

Preflight check:
```
## Preflight check, focus on what this flow do:
$> ticat <flow> :-

## Preflight check again, focus on dependecies report:
$> ticat <flow> :+
```

Inplace searching when writing a sequence:
```
## Inplace search for what command we should use after a sequence of commands:
$> ticat <command>:<command> :- <find-str>

## Continue searching
$> ticat <command>:<command> :- <find-str> <find-str>

## Switch to "+" to show the result detail
$> ticat <command>:<command> :+ <find-str> <find-str>

## Add a new command we just found:
$> ticat <command>:<command>:<commad>
...
```

## Full tail editing style support

We already discussed how we could benefit from tail editing above.

Sometimes `+` and `-` are not enough for full tail editing.

For example, in these cases `+` and `-` are occupied by describing flow execution:
```
$> ticat <flow-command> :-
$> ticat <flow-command> :+
$> ticat <command>:<command>:<command>:<command> :+
$> ticat <command>:<command>:<command>:<command> :-
```

Obviously, more "priority" and "power" commands are needed to avoid long-cusor-move editing.

We tentatively use `$` for "show the last command's info",
there might be more function giving to `$` in the future,
but this command is not stable, maybe we would cancel it someday,
because there are too much weird things for users already.

For now, we check the input args to **ticat**, if the grammar is wrong, **ticat** will report error.
To achive full tail editing,
we need a way to let users putting wrong things and then we give back the suggestions.
