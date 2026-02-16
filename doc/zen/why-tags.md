# Why use tags?

Tags in **ticat** may seem informal - they're just words in help strings with `@` prefixes. Why this approach?

## The purpose of tags

Tags connect module authors with users:
- **Authors**: Declare "what this module is for"
- **Users**: Discover "how to find what they need"

## Why not a formal tag system?

**ticat** has powerful full-text search capabilities. Any property in a command can match keywords:
- Command name
- Help string
- Argument names
- Environment operations
- Everything

Since we have full-text indexing, a common word in a help string is effectively a tag.

## The `@` prefix convention

We recommend adding the `@` prefix to improve search accuracy:

```bash
# In a module's help string:
help = Run benchmark tests @ready @benchmark @performance

# Users can search:
$> ticat / @ready
$> ticat / @benchmark
```

Benefits of the `@` prefix:
- Clearly distinguishes tags from regular words
- Improves search precision
- Easy to spot when reading help
- No special syntax required

## Conventional tags

**ticat** has some conventional tags with specific meanings:

| Tag | Meaning | Example |
|-----|---------|---------|
| `@ready` | Ready-to-go, works out of the box | `@ready @production` |
| `@selftest` | For testing the repository | `@selftest @ci` |
| `@scanner` | Scans for issues | `@scanner @jitter` |
| `@benchmark` | Performance testing | `@benchmark @tpcc` |

Repository authors should explain their custom tags in their README.

## Best practices

### For module authors

```bash
# Good: Clear tags in help string
help = Deploy cluster to kubernetes @ready @deploy @k8s @cluster

# Good: Multiple relevant tags
help = Run TPCC benchmark @benchmark @tpcc @performance @ready

# Avoid: Too many tags
help = Do stuff @ready @test @dev @prod @fast @slow @v1 @v2
```

### For users

```bash
# Find ready-to-go commands from a specific repo
$> ticat / repo-name @ready

# Combine tags for precise results
$> ticat / @benchmark @tpcc

# Mix tags with keywords
$> ticat / @ready mysql
```

## Comparison with formal tag systems

| Aspect | Formal System | ticat Tags |
|--------|--------------|------------|
| Definition | Separate metadata | In help string |
| Registration | Required | Optional |
| Validation | Enforced | Convention |
| Flexibility | Limited | Unlimited |
| Maintenance | High overhead | Zero overhead |

## Why this works

1. **Simplicity**: No special files or registration
2. **Flexibility**: Any tag, any time
3. **Discoverability**: Tags appear in search results
4. **Evolution**: Tags can change as needs evolve
5. **Convention over configuration**: `@` prefix is enough structure
