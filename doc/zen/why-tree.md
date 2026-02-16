# Why are commands and environment key-values in tree form?

Most command-line tools organize commands in 2-3 levels of subcommands:

```bash
$> git remote add
$> docker container run
$> kubectl get pods
```

**ticat** uses a tree structure too, but with differences:
- Paths are joined with `.` (e.g., `cluster.start`)
- Unlimited depth supported

## Why tree structure?

### Namespace management

**ticat** is a platform that can have huge numbers of commands from different authors. The tree structure provides a namespace mechanism to avoid naming conflicts.

**Example**: Multiple teams can provide `deploy` commands:
```bash
team-alpha.deploy     # Team Alpha's deployment
team-beta.deploy      # Team Beta's deployment
k8s.deploy           # Kubernetes deployment
aws.deploy           # AWS deployment
```

### Logical organization

Branch paths represent concepts:
- Authors provide tools in their sub-tree
- Users explore specific branches for related tools

```bash
# Database branch
database.mysql.backup
database.mysql.restore
database.postgres.backup
database.postgres.restore

# Cloud provider branch
cloud.aws.deploy
cloud.gcp.deploy
cloud.azure.deploy
```

## Commands vs. Environment

Both use tree structure for the same reasons:

### Commands
```
hub.add
hub.list
hub.update
hub.disable
```

### Environment keys
```
cluster.host
cluster.port
cluster.name
cluster.region
```

## How users should navigate

**Looking through the tree is not recommended.**

Instead, use search:

```bash
# Don't browse:
$> ticat cmds.tree database

# Do search:
$> ticat / backup mysql
$> ticat / @ready deploy
```

Search is more efficient because:
- You don't need to know the structure
- Results span multiple branches
- Tags and keywords work together

## Tree depth

**ticat** supports unlimited depth, enabling sophisticated organization:

```bash
# Deep nesting for complex systems
company.project.team.module.action
tiup.cluster.tidb.tikv.start
benchmark.tpch.scale.factor.run
```

This would be unwieldy with traditional subcommand parsing, but works naturally with `.` separators.

## Benefits of the tree approach

| Benefit | Description |
|---------|-------------|
| **Namespacing** | No conflicts between authors |
| **Organization** | Logical grouping of related commands |
| **Discoverability** | Branch exploration reveals related tools |
| **Flexibility** | Any depth, any structure |
| **Consistency** | Same model for commands and env keys |

## Practical example

A team managing a TiDB cluster might organize commands:

```bash
# Cluster management
tidb.cluster.start
tidb.cluster.stop
tidb.cluster.restart
tidb.cluster.scale

# Backup operations
tidb.backup.full
tidb.backup.incremental
tidb.backup.restore

# Benchmark operations
tidb.bench.tpcc
tidb.bench.tpch
tidb.bench.sysbench
```

Users can:
```bash
# Explore a branch
$> ticat tidb.bench:-

# Search for specific functionality
$> ticat / backup @ready

# Use abbreviations
$> ticat t.b.tpcc
```

## Conclusion

The tree structure provides:
- **Scalability**: Handle thousands of commands
- **Organization**: Clear ownership and grouping
- **Conflict prevention**: Namespaces for different authors
- **Discoverability**: Branches reveal related functionality

Combined with powerful search, users get the best of both worlds.
