# Zen: the choices we made

## Why so many abbrs and aliases?

Two reasons:
* Reduce memorizing works
* To have a long command name

## Reduce memorizing works

With the ability of flexable ad-hot feature assembling,
**ticat** users also have overwhelming infos.

We did all things we can to reduce this pressure:
* Full search supporting for all things, recommend users don't look after things, just search.
* Various way to display infos, some are focus and detailed, some are large range and essential.
* ...

Abbrs supporting is one of them,
module developers are suggested to setup possible aliases, or even misspellings.
In that, users don't need to pay attention to memorize commands and arg-names,
they could just have a roughly guess.

For example, for command `tpcc.run`, it has a arg named `terminal` by a popular implementation.
If this arg have aliases `term` `thread` `threads`, that users can hardly make mistakes.

## To have a long command name

Some commands have relatively complicated meanings, for that, a long command name is a good choice.

But long name is unfriendly in use, it need more momorizing and more typing.

With abbrs supporting, these commands could have a long and meaningful name, yet still easy to use.

Realname will display no matter where an abbr is displaying, to show the command meaning.
