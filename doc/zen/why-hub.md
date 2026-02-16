# Why use git repositories to distribute components?

Most component platforms build their own central service. Why does **ticat** just use git repositories?

## The goal: Build a better community

The answer is simple: **to build a healthier, more inclusive community**.

## Official-centered vs. User-centered

### Traditional model (Official-centered)

Most platforms are official-centered:
- The authority publishes components
- Users pull what they need
- Publishing user-owned components is difficult

Some platforms provide mirror tools and allow users to establish their own centers. But:
- Tools remain under authority control
- Access settings are controlled centrally
- Publishing still requires approval or process

### ticat model (User-centered)

With git repositories as the publishing mechanism, **ticat** creates a user-centered model:

- Users decide what to add to their local system
- Anyone can become a publisher at zero cost
- Just fork, edit, and push

```bash
# Share your modules instantly
$> ticat hub.add your-username/your-modules
```

## The layered ecosystem

A user-centered model enables a **layered community**:

### Entry level
- Developers don't need to be experts to contribute
- Can write "crappy" code at the beginning
- Still become publishers sharing their work
- Can assemble with professional modules immediately

### Growth path
- Once developers create good pieces, adapting work to higher-level publishers is easy
- Natural progression from beginner to core contributor
- Recognition through actual usage, not approval processes

### Community evolution
- Groups provide easy-to-use environments with runnable flows
- Smooth path from beginner to core coding
- "Better than now" rule: work doesn't need to be perfect, just improved

## Benefits of git-based distribution

### For publishers
- **Zero infrastructure** - No servers to maintain
- **Version control** - History, branches, releases
- **Access control** - Use git's built-in permissions
- **Familiar workflow** - Standard git operations

### For users
- **Trust** - Know exactly where code comes from
- **Transparency** - Can review code before using
- **Choice** - Pick from multiple sources
- **Offline** - Works without central service

### For teams
- **Private repos** - Internal distribution
- **Forking** - Customize without permission
- **Pull requests** - Contribute back easily

## Comparison

| Model | Publishing Cost | Access Control | Trust Model |
|-------|----------------|----------------|-------------|
| Central registry | High | Authority-controlled | Platform trust |
| Mirror system | Medium | Mixed | Complex |
| **Git repos (ticat)** | **Zero** | **Owner-controlled** | **Repository trust** |

## The "better than now" principle

We recommend the "better than now" rule:
- Work doesn't need to be good
- It just needs to be better than the current state

This philosophy enables:
- Gradual improvement
- Community evolution
- Lower barriers to contribution
- More diverse ecosystem

## Practical examples

### Individual developer
```bash
# Create a repo for your scripts
git init my-ticat-modules
cd my-ticat-modules
# Add modules...
git push

# Others can use them
ticat hub.add you/my-ticat-modules
```

### Team
```bash
# Team shares common tools
ticat hub.add myteam/dev-tools

# Individual can extend
ticat hub.add myteam/dev-tools my-personal-extensions
```

### Open source project
```bash
# Project provides integration tools
ticat hub.add project/official-tools

# Community provides alternatives
ticat hub.add community/enhanced-tools
```

## Conclusion

Git-based distribution creates an active, layered ecosystem where:
- Anyone can publish
- Quality emerges through usage
- Community evolves naturally
- No central authority controls innovation
