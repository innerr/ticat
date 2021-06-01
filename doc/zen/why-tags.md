# Zen: the choices we made

## Why use tags, and it looks like very informal, just some words in help string.

Using tags to declair "what we are" is a regular thing,
we use some conventional tag to connect module authors and users,
to let them know "how to use" and "how to tell people how to use".

As a platform (even is a small one) searching is the most important thing,
**ticat** can do a excellent job in searching,
any properties in a command could match by keywords:
command name, help string, arg names, env ops, anyting.

Since we have full text indexing, a common word in help string should be enough as a tag.
We recommend adding prefix `@` could improve the searching accuracy.
So tags will looks like `@ready` `@selftest` embedded in help string.
