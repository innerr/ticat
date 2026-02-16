# Why use CLI as component platform?

## The question

Most component platforms use RPC as the component API, then use configuration files to assemble pieces, or perhaps a web-based UI. Why build a CLI-based component platform?

## The answer: Cost is everything

In engineering, **cost is the primary consideration**. (It took me more than 10 years to realize this - in my early days, I was fixated on "quality".)

Since **ticat** is designed for engineers, using CLI as the platform:
- Does not significantly compromise usability
- Greatly reduces the cost of writing components

## Lower barriers, healthier ecosystem

By lowering the barrier (cost) of component creation, small solutions can move from:
- Manual executions
- Ad-hoc scripts
- One-off commands

...into a unified, shareable platform.

This creates a healthy and growing ecosystem where:
- Simple utilities become reusable modules
- Team-specific workflows become shareable flows
- Knowledge becomes codified and transferable

## Benefits of CLI-based components

### For developers
- **No framework to learn** - If you can write a script, you can write a module
- **Language agnostic** - Use Bash, Python, Go, or any executable
- **Easy debugging** - Run modules directly during development
- **Familiar tools** - Use your existing CLI skills

### For users
- **Composable** - Chain commands naturally with `:`
- **Scriptable** - Easy to automate and integrate
- **Transparent** - See exactly what's running
- **Accessible** - Works over SSH, in containers, everywhere

### For teams
- **Version controllable** - Modules are just files
- **Easy to share** - Push to git, done
- **No infrastructure** - No servers, registries, or services to maintain

## Comparison with alternatives

| Platform Type | Setup Cost | Learning Curve | Sharing Ease |
|--------------|-----------|----------------|--------------|
| RPC-based | High | High | Medium |
| Web UI | Medium | Low | Low |
| Config-based | Medium | Medium | Medium |
| **CLI (ticat)** | **Low** | **Low** | **High** |

## When CLI is the right choice

**ticat** is ideal when:
- Your team lives in the terminal
- You need to automate workflows
- You want easy sharing without infrastructure
- Your tools are already CLI-based

Consider alternatives when:
- Non-technical users need a GUI
- Real-time collaboration is required
- Your components are inherently visual
