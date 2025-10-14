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

## Features

- **Instant startup**: No waiting for REPLs to initialize (looking at you, Swift)
- **Multiline Go code**: Write functions, if statements, for loops with proper continuation prompts
- **Command substitution**: `$(command)` syntax captures command output into Go strings
- **Go REPL**: Write Go code directly in your shell with persistent state
- **Traditional commands**: Just works - `ls`, `git status`, etc.
- **Hybrid mode**: Mix Go code and shell commands seamlessly
- **Built-ins**: `cd`, `exit`, `pwd` with path expansion
- **Signal handling**: Proper Ctrl+C behavior for interrupting processes
- **macOS & Linux**: Windows users can use PowerShell

## Quick Start

### Homebrew (Recommended)

```bash
# Install via homebrew tap
brew install rsarv3006/gosh/gosh

# Add to system shells (optional, to use as login shell)
echo '/opt/homebrew/bin/gosh' | sudo tee -a /etc/shells

# Set as default shell (optional)
chsh -s /opt/homebrew/bin/gosh

# Run
gosh
```

### Go Install

```bash
# Install the current MVP release
go install github.com/rsarv3006/gosh@v0.0.1

# Run
gosh
```

## Usage

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

gosh> for i := 0; i < 3; i++ {
...     fmt.Println("Hello", i)
... }
Hello 0
Hello 1
Hello 2

# Command substitution - game changing feature!
gosh> files := $(ls)
gosh> fmt.Println(strings.Split(files, "\n")[0])
README.md

# Common packages pre-imported
gosh> files, _ := filepath.Glob("*.go")
gosh> fmt.Println(files)
[main.go repl.go router.go ...]

# Mix and match
gosh> pwd
/Users/you/gosh
gosh> name := "gosh"
gosh> fmt.Printf("Welcome to %s\n", name)
Welcome to gosh

# Configure your shell with config.go
gosh> hello("user")
Hello user! Welcome to gosh with config support!
gosh> info()
Config loaded successfully!
User: rjs
```

### Configuration - Hybrid Environment Strategy

gosh uses a **dual-layer configuration approach** that gives you the best of both worlds: standard shell compatibility plus Go-powered extensions.

#### Layer 1: Standard Shell Environment (`env.go`)

**Automatic Standard Config Loading:**

- Loads regular shell configs when run as login shell
- Supports: `.bash_profile`, `.zprofile`, `.profile`, `.bash_login`, `.login`
- Full POSIX environment inheritance
- Shell variable expansion `$HOME`, `$PATH`, `$GOPATH`

**Example .bash_profile**

```bash
# Your existing shell configs just work!
export PATH="/opt/homebrew/bin:$PATH"
export GOPATH="$HOME/go"
export EDITOR="vim"
export JAVA_HOME="/usr/local/opt/openjdk"
```

#### Layer 2: Go-Powered Extensions (`config.go`)

Create a Go file for your global shell customization at:

`~/.config/gosh/config.go`

This single global config loads every time gosh starts, providing consistent shell behavior across all projects.

```go
// ~/.config/gosh/config.go
package main

import (
	"fmt"
	"os"
)

func init() {
	// Global environment setup
	os.Setenv("GOPATH", os.Getenv("HOME") + "/go")
	os.Setenv("EDITOR", "vim")
	fmt.Println("gosh global config loaded!")
}

// Global functions available in any gosh session
func info() {
	fmt.Printf("gosh %s - GOPATH: %s, EDITOR: %s\n",
		"main".GetVersion(), os.Getenv("GOPATH"), os.Getenv("EDITOR"))
}

func clean(a string) string {
	return strings.TrimSpace(a)
}
```

#### Environment Layer Benefits

**Standard Shell Features:**

- ✅ `export VAR=value` syntax (no learning curve)
- ✅ Supports your existing `.bash_profile` / `.zprofile`
- ✅ Traditional environment variable management
- ✅ Shell variable expansion in commands: `echo $HOME`

**Go Extension Features:**

- ✅ Full Go language for custom shell functions
- ✅ Access to all Go packages and types
- ✅ Better error handling and debugging
- ✅ Cross-platform compatibility
- ✅ IDE support with LSP and autocomplete
- ✅ Persistent state across shell session

#### Usage Examples

```bash
# Standard shell commands work exactly as expected
gosh> echo $HOME
/Users/rjs
gosh> echo $GOPATH
/Users/rjs/go
gosh> go install github.com/kubernetes/kompose@latest
# ✅ Works because GOPATH is properly set and shell variables expand

# Go-powered extensions are available too
gosh> hello("world")
Hello world! Welcome to gosh!
gosh> gitSummary()
[GIT STATUS WITH CUSTOM GO LOGIC]

# Mix standard shell and Go code seamlessly
gosh> files := $(ls)  # Command substitution
gosh> fmt.Println("Found", len(strings.Split(files, "\n")), "files")
Found 12 files
```

**The hybrid approach means you get:**

- Zero learning curve for basic shell usage
- Standard POSIX environment behavior
- Your existing shell configs work automatically
- Go superpowers when you need them
- No custom environment syntax to learn

## For Technical Details

📖 **See [ARCHITECTURE.md](ARCHITECTURE.md)** for complete technical documentation including:

- Core components and data flow
- Command substitution implementation
- Error handling and signal management
- Testing strategies and performance considerations

## Building

```bash
git clone https://github.com/rsarv3006/gosh
cd gosh
go build
./gosh
```

## Status

**✅ MVP Complete**:

- ✅ Basic REPL loop with readline
- ✅ Multiline Go code support (essential for Go!)
- ✅ Go evaluation with yaegi and state persistence
- ✅ Command execution with proper signal handling
- ✅ Built-ins (cd, exit, pwd, help)
- ✅ Command substitution `$(command)` syntax
- ✅ Hybrid environment strategy (standard shell + Go extensions)
- ✅ Smart routing between Go code and shell commands
- ✅ Proper Ctrl+C interrupt handling
- ✅ Clean architecture with separated concerns

**🎯 Phase 2 Complete**:

- ✅ Hybrid environment system (standard shell configs + Go extensions)
- ✅ Config file support (config.go)
- ✅ Tab completion for commands and file paths
- ✅ Color system with theme support
- ✅ Comprehensive test coverage
- ✅ Enhanced help system

**🚀 Phase 3 In Progress**:

- ✅ Command history navigation (up/down arrows)
- [ ] Better error messages with line numbers
- [ ] Git integration in prompt

**🔧 Phase 4 Planned**:

**goshlib - Shell-Friendly Go Library**

Phase 4 introduces an optional utility library that provides shell-like convenience functions while maintaining Go's power and flexibility. Functions can be used both in gosh and as standalone Go imports.

```go
// One-liner friendly functions (optional via config)
gosh> result := run("ls -la")           // Auto-trim output
gosh> files := lsDir(".", "*.go")       // List files by pattern
gosh> found := grepFile("README", "gosh") // Search file content
gosh> write("out.txt", "hello")         // Write file
gosh> content := read("out.txt")         // Read file with auto-trim

// Channel sugar for concurrent operations
gosh> ch := channel()                   // Create buffered channel
gosh> async(send, ch, "hello")          // Background send
gosh> println(recv(ch))                 // Receive value

// String processing utilities
gosh> cleaned := clean(text)            // Trim whitespace (custom example)
gosh> parts := split(text, "\n")        // Split by delimiter
gosh> first := head(parts, 3)           // Get first N items
```

**Key Benefits:**
- **Consistent Interface**: Same functions work in gosh and standalone Go programs
- **Gradual Migration**: Start with one-liners, import library for formal scripts
- **Optional Loading**: Purists get clean Go by default, enable via `EnableUtils = true` in config
- **Explicit Power**: All utilities are visible Go code, no hidden magic
- **Learning Bridge**: Makes Go patterns accessible to shell scripters

**Configuration:**
```go
// ~/.config/gosh/config.go
var EnableUtils = true  // Set to true to load shell-friendly utility functions
```

## 🚀 Future Plans

### Project-Specific Configuration - TODO:

- [ ] **Project config loading** - `loadProject("config.local")` or similar
- [ ] **Project-specific functions** - Define per-project Go functions
- [ ] **Environment files support** - Load `.env.local` or project configuration
- [ ] **Project detection** - Automatically detect project type and load appropriate config
- [ ] **Mix local and global** - Combine global setup with project-specific overrides

### Developer Experience - TODO:

- [ ] **Go function autocomplete improvement** - Currently basic tab completion for commands and file paths, need intelligent Go function completion
- [ ] **Go intellisense implementation** - Code completion, type hints, function signatures for Go code in the REPL
- [ ] LSP integration for Go code editing in the shell
- [ ] Syntax highlighting for Go code input
- [ ] Documentation lookup (`go doc` integration)

## Known yaegi Limitations

1. **CGo**: Can't interpret CGo code
2. **Generics**: Limited support (improving)
3. **Unsafe**: Some unsafe operations restricted

**These don't matter for a shell REPL - we're doing basic scripting, not systems programming.**

## Dependencies

```go
module github.com/rsarv3006/gosh

go 1.21

require (
    github.com/chzyer/readline v1.5.1    # Multiline input & history
    github.com/traefik/yaegi v0.15.1     # Go interpreter
)
```

## License

MIT

## Contributing

PRs welcome! This is a fun project to learn Go and build something useful. The architecture is intentionally simple - everything has clear responsibilities and the code is easy to follow.
