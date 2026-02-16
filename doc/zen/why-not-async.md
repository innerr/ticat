# Why not support async/concurrent executing?

**ticat** executes commands sequentially, one at a time. Why doesn't it support async or concurrent execution?

## Historical context

We built an earlier version called `ti.sh` that **did** support async/concurrent execution using a `go` keyword:

```bash
# Old ti.sh syntax (not in ticat)
$> ti.sh go task1 : go task2 : go task3
# All three run concurrently
```

## What we learned

In practice, async/concurrent execution was **rarely used**:
- Most workflows are inherently sequential
- Dependencies between steps require ordering
- Error handling becomes complex
- Results become harder to interpret

## The decision

We decided not to support async/concurrent execution in **ticat** for these reasons:

1. **Limited demand**: Few real-world use cases
2. **Added complexity**: Significant implementation overhead
3. **User confusion**: Harder to understand execution flow
4. **Debugging difficulty**: Concurrent bugs are notoriously hard

## Current workaround

Components can implement their own async execution:

```bash
# Module starts background process
$> ticat bench.async-run

# Do other work
$> ticat other-work

# Wait for background process
$> ticat bench.wait-async
```

This approach:
- Keeps ticat simple
- Gives modules full control
- Allows custom async patterns
- Makes behavior explicit

## When we might add it

We'll consider adding async support if:
- Strong user demand emerges
- Clear use cases develop
- A clean design presents itself

For now, sequential execution provides:
- **Predictability**: Know exactly what runs when
- **Simplicity**: Easy to understand and debug
- **Reliability**: No race conditions or deadlocks

## Comparison

| Feature | Sequential (ticat) | Concurrent |
|---------|-------------------|------------|
| Execution | One at a time | Multiple at once |
| Debugging | Easy | Difficult |
| Error handling | Straightforward | Complex |
| State management | Simple | Complex |
| Determinism | High | Low |
| Use cases | Most workflows | Parallel processing |

## Future considerations

If async support is added, it might look like:

```bash
# Hypothetical future syntax
$> ticat async.start : task1 : task2 : task3 : async.wait

# Or parallel blocks
$> ticat parallel {task1 : task2} : {task3 : task4}
```

But this is speculative and not currently planned.

## Summary

- Async was tried in `ti.sh` but rarely used
- Sequential execution covers most workflows
- Modules can implement their own async patterns
- Will reconsider if demand emerges
