# Zen: the design philosophy of ticat

This section explains the reasoning behind **ticat**'s design decisions. Understanding these principles will help you use **ticat** more effectively and appreciate its unique approach to command-line component assembly.

## Design Principles

**ticat** was designed with these core principles in mind:

1. **Simplicity over complexity** - Easy to learn, easy to use
2. **Composability** - Small pieces that work together seamlessly
3. **Shareability** - Low-friction sharing of tools and workflows
4. **Flexibility** - Adapt to any use case without forcing patterns

## Topics

- [Why ticat](./why-ticat.md) - The motivation behind creating a new platform
- [Why use CLI as component platform](./why-cli.md) - Benefits of command-line interfaces for components
- [Why not use unix pipe](./why-not-pipe.md) - Design differences from traditional unix pipes
- [Why the usage seems weird, especially `+` and `-`](./why-weird.md) - Explanation of unusual syntax choices
- [Why use tags](./why-tags.md) - The tag system for command organization
- [Why so many abbreviations and aliases](./why-abbrs.md) - Balancing brevity with clarity
- [Why commands and environment key-values are in tree form](./why-tree.md) - Hierarchical organization benefits
- [Why use git repositories to distribute components](./why-hub.md) - Leveraging git for package management
- [Why not support async/concurrent executing](./why-not-async.md) - Simplicity trade-offs

## The ticat Philosophy

### For Users

- **Memorize nothing, just search** - All features are discoverable
- **Compose, don't code** - Build workflows by combining existing modules
- **Share freely** - Easy distribution via git repositories

### For Developers

- **Write once, use everywhere** - Modules work across different contexts
- **Loose coupling through environment** - Components communicate via shared key-values
- **Language agnostic** - Write modules in any language

### For Teams

- **Standardized interfaces** - Consistent module interaction patterns
- **Version control integration** - Leverage git for collaboration
- **Extensible without modification** - Add new capabilities without changing existing code

## Learning Path

1. Start with [Why ticat](./why-ticat.md) to understand the motivation
2. Read [Why the usage seems weird](./why-weird.md) if the syntax feels unusual
3. Explore specific topics based on your interests or questions
