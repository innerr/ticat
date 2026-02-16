# Why not just use unix-pipe?

**ticat** uses a unix-pipe-like style, but it's fundamentally different. Here's why.

## Execution order difference

### Unix pipe
All commands in a pipeline launch simultaneously:
```bash
$> cmd1 | cmd2 | cmd3
# All three start at the same time
# Data flows through buffers
```

### ticat sequence
Commands execute one at a time, sequentially:
```bash
$> ticat cmd1 : cmd2 : cmd3
# cmd1 runs first, completes
# Then cmd2 runs, completes
# Then cmd3 runs
```

**Why?** Workflow control requires sequential execution with state management between steps.

## Unix-pipe limitations

### Anonymous and unmanaged

Unix pipes are anonymous - hard to define and enforce a specific protocol.

**Alternative**: Named pipes (`mkfifo`) could solve this, but:
- Resource recycling is problematic
- Complex error handling
- Not suitable for workflow management

### Single data stream

There's only one pipe between commands, meaning only one type of data can pass through.

To pass multiple values, you'd need to:
- Define complex protocols
- Serialize/deserialize data
- Handle encoding/decoding in every component

This makes component development significantly harder.

## ticat environment: The better solution

### Named key-values

Environment key-values can be considered as named pipes with benefits:
- Each key has a meaningful name
- Values can have defined formats
- Multiple values flow through the system

### Automatic management

- **No recycling issues** - ticat handles cleanup automatically
- **Session model** - Clean separation between executions
- **Type awareness** - Keys can be validated

### Dependency checking

Before execution, ticat checks environment operations:

```bash
$> ticat bench :+
-------=<unsatisfied env read>=-------

<FATAL> 'cluster.port'
       - read by:
            [bench.load]
            [bench.run]
       - but not provided
```

Commands must register what they read or write:
- `read` - Must be provided
- `write` - Will be created
- `may-read` - Optional input
- `may-write` - Optional output

This makes ticat a **managed named-pipe-like system**.

## Call stack support

A ticat flow can call other flows, forming call stacks:

```
┌───────────────────┐
│ stack-level: [3]  │             06-02 04:51:45
├───────────────────┴────────────────────────────┐
│    bench.scale = 4                             │
│    cluster.port = 4000                         │
├────────────────────────────────────────────────┤
│    local.build                                 │
│ >> cluster.local                               │
│        port = ''                               │
│        host = 127.0.0.1                        │
│    cluster.restart                             │
│    ben(=bench).load                            │
│    ben(=bench).run                             │
└────────────────────────────────────────────────┘
```

Unix pipes cannot support this - there's no concept of nested calls or return values.

## Comparison

| Feature | Unix Pipe | ticat |
|---------|-----------|-------|
| Execution | Parallel | Sequential |
| Data format | Single stream | Named key-values |
| Protocol | Manual | Automatic checking |
| Nesting | No | Yes (call stacks) |
| Error handling | Exit codes | Rich error context |
| State | None | Environment |
| Debugging | Difficult | Built-in support |

## When to use each

**Use unix-pipe when**:
- Processing data streams
- Parallel execution needed
- Simple transformations

**Use ticat when**:
- Orchestrating workflows
- Managing state between steps
- Building composable tools
- Need dependency checking
