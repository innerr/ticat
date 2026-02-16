# Module Development Zone

Welcome to the **ticat** module development documentation. This section covers everything you need to know to create, share, and maintain modules and flows.

## Quick Start

- [Quick-start guide](../quick-start-mod.md) - Create your first module in minutes
- [Examples: write modules in different languages](https://github.com/innerr/examples.ticat) - Python, Bash, Go, and more
- [How modules work together (with graphics)](../concept-graphics.md) - Visual explanation of module interactions

## ticat Specifications

This section contains the technical specifications for **ticat**. Note that repositories providing modules and flows may have their own specifications.

### Core Components

- [Hub: list/add/disable/enable/purge](./hub.md) - Repository and directory management
- [Command sequence](./seq.md) - How commands execute in sequences
- [Command tree](./cmds.md) - Command organization and structure
- [Environment: list/get/set/save](./env.md) - Environment variable management

### Advanced Topics

- [Abbreviations of commands, env-keys and flows](./abbr.md) - Alias and abbreviation system
- [Flow: list/save/edit](./flow.md) - Flow management and persistence
- [Display control in executing](./display.md) - Output formatting and control
- [Help info commands](./help.md) - Help system and documentation

### Module Development

- [Local store directory](./local-store.md) - Where ticat stores data
- [Repository tree](./repo-tree.md) - Repository structure and sub-repos
- [Module: environment and args](./mod-interact.md) - How modules interact with ticat
- [Module: meta file](./mod-meta.md) - Module metadata and configuration

## Learning Path for Module Developers

1. **Beginner**
   - Read the [quick-start guide](../quick-start-mod.md)
   - Try the [examples repository](https://github.com/innerr/examples.ticat)
   - Understand [module meta files](./mod-meta.md)

2. **Intermediate**
   - Learn about [environment interactions](./mod-interact.md)
   - Study [command sequences](./seq.md)
   - Explore [repository structure](./repo-tree.md)

3. **Advanced**
   - Master [environment operations](./env.md)
   - Understand [display control](./display.md)
   - Create complex [flows](./flow.md)

## Module Types

ticat supports several module types:

| Type | Extension | Description |
|------|-----------|-------------|
| Builtin | N/A | Compiled into ticat binary |
| Python | `.py` | Python scripts |
| Bash | `.bash`, `.sh` | Shell scripts |
| Go | `.go` | Go source files |
| Executable | Any | Binary executables |
| Directory | N/A | Directory with modules |

## Contributing Modules

To share your modules with the community:

1. Create a git repository with your modules
2. Add proper [meta files](./mod-meta.md) with help strings and tags
3. Share the repository address: `ticat hub.add <your-repo>`

See the [Hub specification](./hub.md) for more details.
