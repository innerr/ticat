---
title: ticat — Agent context
---

## What this repo is

ticat (Tiny Component Assembly Tool) is a modular CLI framework for building
composable command-line applications with workflow automation and environment
management.

## High-level modules

| Path              | Role                                    |
| ----------------- | --------------------------------------- |
| `pkg/core/model/` | Core data structures (Cmd, Env, Flow)   |
| `pkg/core/parser/`| Command line parsing                     |
| `pkg/core/execute/`| Execution engine and interactive mode   |
| `pkg/cli/display/`| Terminal output, colors, help rendering |
| `pkg/mods/builtin/`| Built-in commands (env, flow, dbg, etc.)|
| `pkg/mods/persist/`| Persistence (flow files, meta, hub)     |
| `pkg/ticat/`      | Main TiCat assembly and entry point     |
| `pkg/utils/`      | Shared utilities                        |
| `pkg/version/`    | Version information                     |

## Commands

```bash
# Build
make ticat
# Or: go build -o bin/ticat ./pkg/main

# Run tests
make unit-test
# Or: go test -p 3 ./pkg/...

# Run single test
go test -run TestName ./pkg/path/to/package/

# Lint
make lint
# Or: golangci-lint run ./pkg/...
```

## Global conventions

- Commands are composable using `:` operator (e.g., `cmd1 : cmd2 : cmd3`)
- Environment is passed between commands and can be modified with `{key=value}`
- Commands declare environment operations (read/write/mayread) for validation
- Flow templates allow saving/loading workflows
- Abbreviations are auto-generated for commands

## Go style

- Format with `gofmt` and `goimports` (enforced by golangci-lint)
- Use `self` as receiver name (e.g., `func (self *MyStruct) Method()`)
- Imports use three groups: stdlib, external, internal (separated by blank lines)
- `PascalCase` for exported names, `camelCase` for unexported names
- No global variables without explicit permission

## Key concepts

| Term        | Description                                           |
| ----------- | ----------------------------------------------------- |
| `CmdTree`   | Hierarchical command tree with abbreviations          |
| `Env`       | Layered environment (default, meta, session, runtime) |
| `Flow`      | Sequence of commands to execute                       |
| `ParsedCmd` | A parsed command with args and env modifications      |
| `Cli`       | Global CLI context with display and executor          |

## Where to look

| Task                      | File                                 |
| ------------------------- | ------------------------------------ |
| Add new command           | `pkg/mods/builtin/` or your mod      |
| Command tree structure    | `pkg/core/model/cmds.go`             |
| Environment handling      | `pkg/core/model/env_val.go`          |
| Flow parsing              | `pkg/core/parser/`                   |
| Execution logic           | `pkg/core/execute/executor.go`       |
| Display/output formatting | `pkg/cli/display/`                   |
| Built-in commands         | `pkg/mods/builtin/builtin.go`        |
| Flow save/load            | `pkg/mods/builtin/flow.go`           |
| Persistence               | `pkg/mods/persist/`                  |

## Explicitly absent

- No `.cursor/rules/`, `.cursorrules`, or `.github/copilot-instructions.md`
