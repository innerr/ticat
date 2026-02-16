# Troubleshooting Guide

This guide helps you resolve common issues with **ticat**.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Command Execution Issues](#command-execution-issues)
- [Environment Issues](#environment-issues)
- [Hub and Repository Issues](#hub-and-repository-issues)
- [Flow Issues](#flow-issues)
- [Module Issues](#module-issues)
- [Performance Issues](#performance-issues)

## Installation Issues

### "command not found: ticat"

**Problem**: The `ticat` command isn't recognized.

**Solutions**:
1. Verify ticat is in your PATH:
   ```bash
   which ticat
   ```

2. Add to PATH temporarily:
   ```bash
   export PATH="/path/to/ticat/bin:$PATH"
   ```

3. Add to PATH permanently (add to `~/.bashrc` or `~/.zshrc`):
   ```bash
   echo 'export PATH="/path/to/ticat/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc
   ```

### Build fails with Go errors

**Problem**: `make` fails with Go compilation errors.

**Solutions**:
1. Check Go version (1.16+ required):
   ```bash
   go version
   ```

2. Update Go if needed:
   ```bash
   # macOS with Homebrew
   brew upgrade go
   
   # Linux
   # Download from https://golang.org/dl/
   ```

3. Clean and rebuild:
   ```bash
   make clean
   make
   ```

### Permission denied errors

**Problem**: Cannot execute ticat binary.

**Solution**: Make the binary executable:
```bash
chmod +x bin/ticat
```

## Command Execution Issues

### "unsatisfied env read" error

**Problem**: Command fails with environment read errors.

**Example**:
```
-------=<unsatisfied env read>=-------

<FATAL> 'cluster.port'
       - read by:
            [bench.load]
            [bench.run]
       - but not provided
```

**Solution**: Provide the required environment value:
```bash
# One-time
$> ticat {cluster.port=4000} bench

# Persistent
$> ticat {cluster.port=4000} env.save
$> ticat bench
```

### Command not found

**Problem**: A command you expect to exist isn't found.

**Solutions**:
1. Check if the repository is added:
   ```bash
   ticat hub
   ```

2. Search for the command:
   ```bash
   ticat / command-name
   ```

3. Check if the repository is enabled:
   ```bash
   ticat hub enable repo-name
   ```

### Command runs but does nothing

**Problem**: Command executes but produces no visible output.

**Solutions**:
1. Check if the command is in quiet mode:
   ```bash
   ticat env.list quiet
   ```

2. Run with verbose output:
   ```bash
   ticat verb : your-command
   ```

3. Verify command execution with step-by-step mode:
   ```bash
   ticat dbg.step.on : your-command
   ```

## Environment Issues

### Environment values not persisting

**Problem**: Saved environment values disappear between sessions.

**Solutions**:
1. Verify save location:
   ```bash
   ticat env.flat sys.paths.data
   ```

2. Check file permissions:
   ```bash
   ls -la ~/.ticat/
   ```

3. Verify the save worked:
   ```bash
   ticat {key=value} env.save
   ticat env.list key
   ```

### Cannot set environment value

**Problem**: Environment value doesn't change when using `{key=value}`.

**Solutions**:
1. Check syntax (no spaces around `=`):
   ```bash
   # Correct
   ticat {key=value} command
   
   # Also correct
   ticat {key = value} command
   ```

2. Verify the value is being read:
   ```bash
   ticat {key=value} env.list key
   ```

### Environment layers confusion

**Problem**: Values in one layer aren't visible to commands.

**Solution**: Understand environment layers:
- **Command layer** (`: {key=value}`): Only for that command
- **Session layer** (`{key=value}`): For entire sequence
- **Persisted layer**: Saved values from `env.save`
- **Default layer**: Hard-coded defaults

```bash
# Command layer - only affects second command
ticat cmd1 : {key=value} cmd2

# Session layer - affects all commands
ticat {key=value} cmd1 : cmd2
```

## Hub and Repository Issues

### "hub.add" fails

**Problem**: Cannot add a repository to hub.

**Solutions**:
1. Check git is installed:
   ```bash
   git --version
   ```

2. Verify the repository exists:
   ```bash
   git ls-remote https://github.com/user/repo
   ```

3. Check network connectivity

4. Try with full URL:
   ```bash
   ticat hub.add https://github.com/user/repo.git
   ```

### Repository modules not loading

**Problem**: Added repository but modules aren't available.

**Solutions**:
1. Check if repository is enabled:
   ```bash
   ticat hub
   # Look for [on] or [off] status
   ```

2. Enable the repository:
   ```bash
   ticat hub.enable repo-name
   ```

3. Update the repository:
   ```bash
   ticat hub.update
   ```

### Cannot purge repository

**Problem**: `hub.purge` command fails.

**Solution**: Disable first, then purge:
```bash
ticat hub.disable repo-name
ticat hub.purge repo-name
```

## Flow Issues

### Flow not found

**Problem**: Saved flow cannot be executed.

**Solutions**:
1. List saved flows:
   ```bash
   ticat flow
   ```

2. Check flow file location:
   ```bash
   ticat env.flat sys.paths.flows
   ```

3. Verify the flow file exists in that directory

### Flow execution order wrong

**Problem**: Commands in flow execute in unexpected order.

**Solution**: Priority commands run first. Check for priority commands:
```bash
ticat flow-name :+
```

Look for `(priority)` in command types.

### Cannot save flow

**Problem**: `flow.save` command fails.

**Solutions**:
1. Check write permissions:
   ```bash
   ls -la $(ticat env.flat sys.paths.flows | cut -d= -f2)
   ```

2. Verify the path exists:
   ```bash
   ticat env.flat sys.paths.flows
   mkdir -p ~/.ticat/flows
   ```

## Module Issues

### Module not executable

**Problem**: Module file exists but won't execute.

**Solutions**:
1. Check file permissions:
   ```bash
   chmod +x module-file
   ```

2. Verify shebang line (for scripts):
   ```bash
   #!/bin/bash
   # or
   #!/usr/bin/env python3
   ```

### Module meta file not recognized

**Problem**: `.ticat` meta file isn't being read.

**Solutions**:
1. Verify naming: Meta file must be `<module-name>.<ext>.ticat`
   ```
   my-module.bash       # Module
   my-module.bash.ticat # Meta file
   ```

2. Check meta file format:
   ```
   help = description of module
   
   [args]
   arg1 = default-value
   
   [env]
   key1 = read
   key2 = write
   ```

### Python module import errors

**Problem**: Python module fails with import errors.

**Solutions**:
1. Check Python version compatibility
2. Install required packages:
   ```bash
   pip install -r requirements.txt
   ```
3. Verify Python path in shebang

## Performance Issues

### Slow startup

**Problem**: ticat takes long to start.

**Solutions**:
1. Reduce number of repositories:
   ```bash
   ticat hub
   # Disable unused repos
   ticat hub.disable unused-repo
   ```

2. Check for slow network (repository updates)

### Slow command execution

**Problem**: Commands execute slowly.

**Solutions**:
1. Use quiet mode for batch operations:
   ```bash
   ticat quiet : your-command
   ```

2. Check if step-by-step is enabled:
   ```bash
   ticat env.list step
   ticat dbg.step.off : env.save
   ```

## Debugging Tips

### Enable verbose output

```bash
# Maximum verbosity
ticat verb : your-command

# Moderate verbosity
ticat v.+ : your-command
```

### Step-by-step execution

```bash
# Confirm each step
ticat dbg.step.on : your-flow
```

### Inspect command details

```bash
# Brief info
ticat command-name :-

# Detailed info
ticat command-name :+
```

### Check environment state

```bash
# All environment values
ticat env.tree

# Filter specific values
ticat env.list keyword
```

## Getting Help

If you can't resolve an issue:

1. **Search existing issues**: [GitHub Issues](https://github.com/innerr/ticat/issues)
2. **Check documentation**: [doc/](doc/)
3. **Open a new issue**: Include:
   - Error messages (full output)
   - Your environment (OS, ticat version)
   - Steps to reproduce
   - What you expected vs. what happened

## Common Error Messages

| Error | Meaning | Solution |
|-------|---------|----------|
| `FATAL` | Required value missing | Provide the environment value |
| `risk` | Optional value missing | Usually safe to ignore |
| `command not found` | Command doesn't exist | Check spelling, add repo |
| `permission denied` | Cannot access file | Check permissions |
| `unsatisfied env read` | Reading before writing | Provide value or add provider |
