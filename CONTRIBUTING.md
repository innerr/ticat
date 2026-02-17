# Contributing to ticat

Thank you for your interest in contributing to **ticat**! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How to Contribute](#how-to-contribute)
- [Development Setup](#development-setup)
- [Coding Standards](#coding-standards)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment. Please be considerate of others and follow standard open-source community guidelines.

## How to Contribute

### Types of Contributions

- **Bug fixes**: Fix issues in existing code
- **New features**: Add new functionality
- **Documentation**: Improve or translate documentation
- **Modules**: Create and share useful modules
- **Examples**: Add examples and tutorials

### Getting Started

1. Fork the repository
2. Clone your fork locally
3. Create a feature branch
4. Make your changes
5. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.16 or later
- Git
- Make (optional, for build commands)

### Building from Source

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/ticat
cd ticat

# Build the project
make

# Run tests
make test

# Add to PATH for development
export PATH="$PWD/bin:$PATH"
```

### Project Structure

```
ticat/
├── bin/                    # Compiled binaries
├── doc/                    # Documentation
│   ├── drafts/            # Draft documents
│   ├── spec/              # Technical specifications
│   ├── usage/             # User guides
│   └── zen/               # Design philosophy
├── pkg/                    # Source code
│   ├── cli/               # CLI parsing
│   ├── core/              # Core functionality
│   ├── main/              # Entry point
│   ├── mods/              # Built-in modules
│   ├── ticat/             # Main package
│   └── utils/             # Utilities
├── Makefile               # Build commands
├── README.md              # Project overview
└── go.mod                 # Go module definition
```

## Coding Standards

### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Run `gofmt` before committing
- Add comments for exported functions and types
- Write tests for new functionality

### Documentation

- Use clear, concise language
- Include code examples where appropriate
- Update the README if adding user-facing features
- Add entries to relevant spec documents

### Commit Messages

Format:
```
<type>: <description>

[optional body]
[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Example:
```
feat: add support for Python 3.11 modules

This adds detection and execution support for Python 3.11
when running .py module files.

Closes #123
```

## Submitting Changes

### Pull Request Process

1. **Create a branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**:
   - Write clean, documented code
   - Add tests if applicable
   - Update documentation

3. **Test your changes**:
   ```bash
   make test
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: your feature description"
   ```

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Open a Pull Request**:
   - Go to the original repository
   - Click "New Pull Request"
   - Select your branch
   - Fill in the PR template

### Pull Request Guidelines

- Keep PRs focused on a single change
- Include tests for new functionality
- Update documentation
- Reference related issues
- Be responsive to code review feedback
- **All CI checks must pass before merging**
  - Tests must pass on all platforms (Ubuntu, macOS)
  - Code must pass linting (golangci-lint, gofmt)
  - Build must succeed
- **Branch protection rules apply**
  - Direct pushes to `main` branch are not allowed
  - All changes must go through pull requests
  - Required status checks must pass before merging

## Reporting Issues

### Before Reporting

- Check existing issues to avoid duplicates
- Try the latest version
- Gather relevant information

### Issue Template

```markdown
**Description**
A clear description of the issue.

**Steps to Reproduce**
1. Run command X
2. Execute flow Y
3. See error

**Expected Behavior**
What you expected to happen.

**Actual Behavior**
What actually happened.

**Environment**
- OS: [e.g., macOS 12.0]
- ticat version: [e.g., 0.1.0]
- Go version: [e.g., 1.18]

**Additional Context**
Any other relevant information.
```

## Creating Modules

If you want to contribute modules rather than core code:

1. Create a separate repository for your modules
2. Follow the [module development guide](doc/quick-start-mod.md)
3. Add proper meta files with help strings and tags
4. Share by telling users to run:
   ```bash
   ticat hub.add YOUR_USERNAME/YOUR_MODULE_REPO
   ```

## Questions?

- Open a GitHub issue for bugs or feature requests
- Check the [documentation](doc/) for usage questions
- Review [existing discussions](https://github.com/innerr/ticat/discussions) if available

Thank you for contributing to **ticat**!
