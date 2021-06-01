# Zen: the choices we made

## Why not support async/concurrent executing?

We built a immature version of **ticat** before call `ti.sh`,
it do support async/concurrent executing with a keyword `go`.

But in practices we found out that these were rarely used,
it has limited demands.

So we decide not to support async/concurrent executing in **ticat** for now.
Once we need them we will provide them.

Components can do async executing by themselves now,
an example:
```
$> ticat bench.async-run : ... : ... : bench.wait-async
```
