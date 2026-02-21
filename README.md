# ticat

A lightweight command-line component platform for building powerful and flexible CLI applications

[![Go Version](https://img.shields.io/badge/Go-1.16%2B-blue)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)
[![CI](https://github.com/innerr/ticat/actions/workflows/ci.yml/badge.svg)](https://github.com/innerr/ticat/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/innerr/ticat)](https://goreportcard.com/report/github.com/innerr/ticat)

**ticat** (Tiny Component Assembly Tool) is a modular CLI framework that enables you to:
- Build CLI applications with built-in composability and workflow automation
- Assemble complex workflows from simple, composable modules
- Manage configurations through a shared environment system

Unlike traditional CLI frameworks like Cobra, ticat provides workflow automation, command composition, and environment management out of the box.

## Why Choose ticat Over Cobra (or others alike)?

| Feature | Cobra | ticat |
|---------|-------|-------|
| Command composition | Manual implementation | Built-in (`:` operator) |
| Configuration management | External libraries needed | Native environment system |
| Workflow saving/loading | Not supported | `flow.save` / `flow.load` |
| Debugging | Manual | Built-in (`+`, `-`, `dbg.step`) |
| Abbreviations | Manual setup | Automatic |

## Using ticat as a CLI Framework

### Minimal Entry Point

Building a CLI with ticat is straightforward. Create a minimal `main.go`:

```go
package main

import (
    "os"
    "github.com/innerr/ticat/pkg/ticat"
    "your-project/pkg/integrate/yourmod"
)

func main() {
    tc := ticat.NewTiCat()
    yourmod.Integrate(tc)
    tc.RunCli(os.Args[1:]...)
}
```

### Integration Function

Define your integration function to set up commands and default environment:

```go
package yourmod

import (
    "github.com/innerr/ticat/pkg/core/model"
    "github.com/innerr/ticat/pkg/ticat"
    "your-project/pkg/mods/yourmod"
)

func Integrate(tc *ticat.TiCat) {
    // Register your commands
    yourmod.RegisterCmds(tc.Cmds)

    // Set default environment values
    defEnv := tc.Env.GetLayer(model.EnvLayerDefault)
    defEnv.Set("your.config.timeout", "30s")
    defEnv.SetBool("your.config.verbose", false)
}
```

### Register Commands

Define your command tree with arguments and environment operations:

```go
package yourmod

import "github.com/innerr/ticat/pkg/core/model"

func RegisterCmds(cmds *model.CmdTree) {
    // Create command branch with abbreviations
    app := cmds.AddSub("myapp", "app").RegEmptyCmd("myapp toolbox").Owner()

    // Add a command with arguments and environment bindings
    deploy := app.AddSub("deploy", "d").RegPowerCmd(
        DeployCmd,
        "deploy to environment").
        AddArg("env", "dev", "e").
        AddArg("version", "", "v").
        AddArg("dry-run", "false").
        AddArg2Env("myapp.deploy.env", "env").
        AddArg2Env("myapp.deploy.version", "version").
        AddEnvOp("myapp.deploy.env", model.EnvOpTypeRead).
        AddEnvOp("myapp.deploy.version", model.EnvOpTypeRead).
        AddEnvOp("myapp.deploy.status", model.EnvOpTypeWrite)

    // Add sub-commands
    deploy.AddSub("rollback", "rb", "undo").RegPowerCmd(
        DeployRollback,
        "rollback last deployment")
}

func DeployCmd(cli *model.Cli, env *model.Env, args *model.Args) error {
    // Read from environment
    targetEnv := env.Get("myapp.deploy.env")
    version := env.Get("myapp.deploy.version")

    // Your deployment logic here
    fmt.Printf("Deploying %s to %s\n", version, targetEnv)

    // Write back to environment (accessible to subsequent commands)
    env.Set("myapp.deploy.status", "success")
    return nil
}
```

## Key Features

### 1. Command Composition

Users can chain commands without writing additional code:

```bash
# Chain multiple operations
myapp build : test : deploy

# Set environment inline
myapp {env=prod} deploy

# Save workflow for reuse
myapp build : test : deploy : flow.save release

# Use saved workflow
myapp release
```

### 2. Environment Management

Persistent configuration without extra code:

```bash
# Set and persist values
myapp {db.host=localhost} {db.port=5432} env.save

# Inspect current environment
myapp env

# Values persist across sessions
myapp deploy  # db.host and db.port are already set
```

### 3. Built-in Debugging

Powerful debugging without instrumentation:

```bash
# Preview execution flow (brief)
myapp release :-

# Preview execution flow (detailed)
myapp release :+

# Step-by-step execution with confirmation
myapp break : release

# Inspect running or finished executions
myapp sessions
```

### 4. Automatic Abbreviations

Commands automatically support abbreviations - no manual setup needed:

```bash
# These are all equivalent
myapp deploy --env prod --version v1.2.3
myapp d -e prod -v v1.2.3
myapp app deploy env=prod version=v1.2.3
```

## Advanced Features

### Environment Operations

Commands declare their environment usage for validation:

```go
cmd.AddEnvOp("db.host", model.EnvOpTypeRead).      // Must be set before execution
    AddEnvOp("db.port", model.EnvOpTypeMayRead).   // Optional read
    AddEnvOp("result.code", model.EnvOpTypeWrite)  // Written by this command
```

ticat validates environment access at runtime and reports issues:

```bash
$ myapp deploy
-------=<unsatisfied env read>=-------

<FATAL> 'db.host'
       - read by:
            [myapp.deploy]
       - but not provided
```

### Flow Templates

Create parameterized workflows:

```bash
# All parameters of `build`/`test`/`deploy` will automatically become parameters of `release`
myapp build : test : deploy : flow.save release
```

### Breakpoint Debugging

Pause execution at specific points:

```bash
# Add breakpoint before deploy
myapp break.at deploy : release env=prod

# Resume execution
(Choose `quit` `continue` `step over` `step into` in the interacting mode)
```

## Cheat Sheet

**Essential syntax:**
- Use `:` to concatenate commands (sequential execution)
- Use `{key=value}` to modify environment key-values

**Command inspection:**
- `cmd :=` - show command usage
- `cmd :-` - show brief flow structure
- `cmd :+` - show detailed flow with env operations

**Search commands:**
- `ticat <str> :/` - global search
- `ticat <branch> :~` - browse command branch tree

**Frequently used:**
- `flow.save <name>` - save flow (abbr: `f.+`)
- `env.save` - save environment (abbr: `e.+`)
- `break` - set breakpoint

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Community

- GitHub Issues: [https://github.com/innerr/ticat/issues](https://github.com/innerr/ticat/issues)
- Pull Requests: [https://github.com/innerr/ticat/pulls](https://github.com/innerr/ticat/pulls)
