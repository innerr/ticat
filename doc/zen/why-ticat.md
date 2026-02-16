# Why ticat?

## The problem

Modern software development involves many tools, scripts, and workflows. Teams often have:
- Numerous shell scripts for various tasks
- Configuration files scattered across projects
- Custom tools that are hard to share and maintain
- Workflows that are difficult to document and reproduce

Traditional solutions have limitations:
- **Shell scripts**: Hard to compose, no standardized interfaces
- **Makefiles**: Focused on building, not general automation
- **Package managers**: Designed for system packages, not custom tools
- **Custom frameworks**: High learning curve, hard to share

## The solution

**ticat** provides a lightweight platform for command-line components that:

### Self-sufficient component model

Modules declare their dependencies through environment operations:
- `read`: Required values that must be provided
- `write`: Values the module provides to others
- `may-read`: Optional values that enhance functionality
- `may-write`: Values that might be written conditionally

This enables automatic dependency checking and assembly.

### Low-cost ad-hoc assembling

Combine modules into workflows with simple syntax:
```bash
$> ticat module1 : module2 : module3
```

No configuration files needed. No boilerplate code. Just compose and run.

### Low-cost component development

Create a module in minutes:
1. Write a script in any language
2. Add a `.ticat` meta file with help text and dependencies
3. Done - it's now a ticat module

### User-centered distribution

Share modules via git repositories:
```bash
$> ticat hub.add user/module-repo
```

Users discover and use your modules immediately. No registry, no publishing process.

## What makes ticat unique

These features are rare or non-existent in other tools:

1. **Environment-based dependency resolution** - Automatic checking and assembly
2. **Universal composition** - Any module works with any other module
3. **Zero-boilerplate modules** - Just code and a simple meta file
4. **Git-native distribution** - No central registry needed
5. **Language agnostic** - Use Python, Bash, Go, or any executable

## Cost comparison

| Task | Traditional | ticat |
|------|------------|-------|
| Create a reusable tool | Write script + documentation + wrapper | Write script + one meta file |
| Share with team | Set up server/registry + publish | Push to git repository |
| Use shared tool | Download + configure + integrate | `hub.add` + run |
| Compose workflows | Write glue code | Use `:` separator |
| Document dependencies | Manual documentation | Automatic checking |

## When to use ticat

**ticat** is ideal for:

- Development teams with shared tooling needs
- DevOps workflows with multiple tools
- Testing and benchmarking automation
- Cluster management and deployment
- Any scenario requiring composable command-line tools

**ticat** might not be the best fit for:

- Simple one-off scripts
- GUI-based workflows
- Real-time or streaming processing
- Extremely performance-critical operations

## The bottom line

**ticat** significantly reduces the cost of creating, sharing, and using command-line tools. It's not about doing things others can't - it's about doing them with less effort and more flexibility.
