# gosh - Go Shell

> A hybrid shell that combines Go's interpreter (via yaegi) with traditional command execution.

## Overview

**gosh** is a hybrid shell that combines Go's interpreter (via yaegi) with traditional command execution. Cross-platform capable, designed as a daily driver, with instant startup and no PTY complexity.

### Core Philosophy

- **Go-first**: The shell is fundamentally a Go interpreter with shell conveniences
- **No shell syntax parsing**: We're not reimplementing bash. Bare commands work, everything else is Go
- **Hybrid approach**: `ls` just works, but `files := $(ls)` is also valid
- **Instant startup**: No waiting for REPLs to initialize
- **Daily driver quality**: Stable, fast, and pleasant to use

## Key Features

- **Instant startup**: No waiting for REPLs to initialize (looking at you, Swift)
- **Multiline Go code**: Write functions, if statements, for loops with proper continuation prompts
- **Command substitution**: `$(command)` syntax captures command output into Go strings
- **Go REPL**: Write Go code directly in your shell with persistent state
- **Traditional commands**: Just works - `ls`, `git status`, etc.
- **Hybrid mode**: Mix Go code and shell commands seamlessly
- **Built-ins**: `cd`, `exit`, `pwd`, `help`, `init` with path expansion
- **Signal handling**: Proper Ctrl+C behavior for interrupting processes
- **Cross-platform**: macOS & Linux support (Windows users can use PowerShell)
- **Configuration support**: Custom Go functions in `~/.config/gosh/config.go`

## Quick Example

```bash
# Regular shell commands work
gosh> ls -la
gosh> git status
gosh> cd ~/projects

# Go code just works
gosh> x := 42
gosh> fmt.Println(x * 2)
84

gosh> func add(a, b int) int {
...     return a + b
... }
gosh> fmt.Println(add(5, 3))
8

# Command substitution - game changing feature!
gosh> files := $(ls)
gosh> fmt.Println(strings.Split(files, "\n")[0])
README.md

# Mix and match
gosh> pwd
/Users/you/gosh
gosh> name := "gosh"
gosh> fmt.Printf("Welcome to %s\n", name)
Welcome to gosh
```

## Status

**✅ MVP Complete**:

- ✅ Basic REPL loop with readline
- ✅ Multiline Go code support (essential for Go!)
- ✅ Go evaluation with yaegi and state persistence
- ✅ Command execution with proper signal handling
- ✅ Built-ins (cd, exit, pwd, help, init)
- ✅ Command substitution `$(command)` syntax
- ✅ Hybrid environment strategy (standard shell + Go extensions)
- ✅ Smart routing between Go code and shell commands
- ✅ Proper Ctrl+C interrupt handling

**🎯 Phase 2 Complete**:

- ✅ Hybrid environment system (standard shell configs + Go extensions)
- ✅ Config file support (config.go)
- ✅ Tab completion for commands and file paths
- ✅ Color system with theme support
- ✅ Comprehensive test coverage
- ✅ Enhanced help system

**🚀 Phase 3 Complete**:

- ✅ Command history navigation (up/down arrows)
- ✅ Better error messages with line numbers
- ✅ Git integration in prompt

**🔧 Phase 4 Complete - Working Shellapi Functions**:

✅ **Real Command Execution (v0.2.2)**
gosh v0.2.2 features shellapi functions that execute real commands via Go's `os/exec`, providing actual command output and persistent directory changes.

## 🚀 Future Plans

### ✅ Critical Bug Fixes - COMPLETED in v0.2.4:

- [x] **Fix directory changes in functions** - Fixed critical bug where only the last `cd` operation persisted within a function. Multi-step workflows now work correctly with immediate sequential directory changes and proper prompt updates.

### Project-Specific Configuration - TODO:

- [ ] **Project config loading** - `loadProject("config.local")` or similar
- [ ] **Project-specific functions** - Define per-project Go functions
- [ ] **Environment files support** - Load `.env.local` or project configuration
- [ ] **Project detection** - Automatically detect project type and load appropriate config
- [ ] **Mix local and global** - Combine global setup with project-specific overrides

### Developer Experience - TODO:

- [ ] **Enhanced command history** - Better history search, filtering, and persistence across sessions
- [ ] **Improved tab completion** - More intelligent completion for commands, file paths, and Go identifiers
- [ ] **Go function autocomplete improvement** - Currently basic tab completion for commands and file paths, need intelligent Go function completion
- [ ] **Go intellisense implementation** - Code completion, type hints, function signatures for Go code in the REPL

### Shellapi Enhancements - TODO:

- [ ] **Interactive program handling** - Better support for commands like `vim`, `nano` that wait for user input
- [ ] **Signal forwarding** - Proper Ctrl+C handling for long-running shellapi processes  
- [ ] **Output streaming** - Real-time output capture for long-running commands
- [ ] **Background processes** - Support for launching background processes via shellapi
- [ ] LSP integration for Go code editing in the shell
- [ ] Syntax highlighting for Go code input
- [ ] Documentation lookup (`go doc` integration)

### Sharp Edges / Known Bugs
- [ ] gosh doesn't handle piped input well

## Get Started

Ready to dive in? Check out our [Installation Guide](install.md) to get gosh running on your system, then head to the [Getting Started](getting-started.md) guide to learn the basics.

For advanced usage and configuration, see the [User Guide](guide.md) and [CLI Reference](reference.md).

---

**Note**: gosh was built with passion for creating a better shell experience that combines the power of Go with the familiarity of traditional shell commands. Welcome to the future of shells! 🚀
