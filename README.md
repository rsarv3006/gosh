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

Create a Go file for shell customization at:
1. `./config.go` (current directory, takes precedence)  
2. `~/.config/gosh/config.go` (home directory, fallback)

```go
// config.go
package main

import (
    "fmt"
    "os"
    "strings"
)

// Runs on shell startup - Go-powered initialization
func init() {
    fmt.Println("Welcome to gosh!")
    os.Setenv("GOSH_USER", os.Getenv("USER"))
}

// Custom functions that persist throughout the shell session
func hello(name string) {
    fmt.Printf("Hello %s! Welcome to gosh!\n", name)
}

func info() {
    fmt.Printf("Config loaded successfully!\n")
    fmt.Printf("User: %s\n", os.Getenv("GOSH_USER"))
}

// Go-powered utilities - things you can't do in bash!
func smartLs() {
    // Custom ls with Go logic, filtering, sorting, etc.
}

func gitSummary() {
    // Git status parsing with Go packages
}

// Custom prompt extension (when implemented)
func CustomPrompt() string {
    return fmt.Sprintf("gosh[%s]$ ", 
        strings.TrimPrefix(os.Getenv("PWD"), os.Getenv("HOME")))
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

## Architecture

### The Core Loop

```
1. Display prompt (shows current directory with ~ substitution)
2. Read line of input (using readline with multiline support)
3. Parse and route input (smart heuristics + PATH checking)
4. Execute (yaegi eval, process spawn, or built-in)
5. Handle output and errors
6. Update state (working directory, Go interpreter state)
7. Repeat
```

### Core Components

#### 1. Router

Takes input and decides: Go eval, process spawn, or built-in:

- Built-ins checked first (`cd`, `exit`)
- Go syntax markers (`var`, `const`, `func`, `type`, `struct`, `interface`, `import`, `for`, `range`, `if`, `switch`)
- Assignment or closures (`:=`, `=`, `{`)
- Function calls with string literals → Go
- Common Go functions (`fmt.Println`, `len`, `cap`, `make`, `append`, `copy`)
- Command substitution syntax `$(command)` → Go
- Fallback to PATH check for commands with parentheses

#### 2. Go Evaluator

**Embedded yaegi interpreter:**

```go
func NewGoEvaluator() *GoEvaluator {
    i := interp.New(interp.Options{
        GoPath: os.Getenv("GOPATH"),
    })
    i.Use(stdlib.Symbols)  // All standard library
    
    // Pre-import common packages for convenience
    i.Eval(`
import (
    "fmt"
    "os"
    "strings"
    "strconv"
    "path/filepath"
)`)
    return &GoEvaluator{interp: i}
}
```

**State persistence:**
- Variables defined persist automatically
- Functions defined persist automatically
- Imports persist automatically
- No parsing, no filtering, no hoping

#### 3. Process Spawner

Resolves executables in PATH and spawns processes with proper stdio handling:

```go
func (p *ProcessSpawner) ExecuteInteractive(command string, args []string) ExecutionResult {
    cmd := exec.Command(command, args...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Dir = p.state.WorkingDirectory
    cmd.Env = p.state.EnvironmentVars()

    err := cmd.Start()
    // Track current process for signal handling
    p.state.CurrentProcess = cmd.Process
    defer func() { p.state.CurrentProcess = nil }()
    
    err = cmd.Wait()
    return ExecutionResult{...}
}
```

#### 4. Built-in Handler

- `cd` - with path expansion
- `exit` - sets shouldExit flag
- Easy to extend

#### 5. State Management

```go
type ShellState struct {
    WorkingDirectory string
    Environment      map[string]string
    ShouldExit       bool
    ExitCode         int
    CurrentProcess   *os.Process  // For signal handling
}
```

## Why gosh vs Other Solutions

| Aspect           | Other REPLs        | gosh                               |
| ---------------- | ----------------- | ---------------------------------- |
| Startup time     | 10-12 seconds     | ~10ms                              |
| Architecture     | PTY + external    | Embedded interpreter               |
| State management | Parse REPL output | Native Go values                   |
| Complexity       | PTY parsing       | Simple API calls                   |
| Persistence      | Hope REPL keeps   | Direct variable storage            |
| Platform         | OS-specific      | Cross-platform                     |

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

- [x] Hybrid environment system (standard shell configs + Go extensions) ✅
- [x] Config file support (config.go) ✅
- [x] Tab completion for commands and file paths ✅
- [x] Color system with theme support ✅
- [x] Comprehensive test coverage ✅
- [x] Enhanced help system ✅

**🚀 Phase 3 In Progress**:

- [x] Command history navigation (up/down arrows) ✅
- [ ] Better error messages with line numbers
- [ ] Pipe support (`ls | grep foo`)
- [ ] Background jobs (`long_command &`)
- [ ] Git integration in prompt

## Success Criteria

**✅ MVP Success:**

- [x] Starts instantly (< 100ms)
- [x] Can run basic commands (`ls`, `git status`, etc.)
- [x] Can write Go code with persistent state
- [x] Doesn't crash on Ctrl+C
- [x] Can write multiline Go code (functions, if statements, loops)
- [x] Can capture command output with `$(command)`
- [x] Can `cd` around properly

**🎯 Daily Driver Success:**

- [x] Want to use it instead of zsh ✅
- [x] Tab completion works well enough ✅
- [x] Command history doesn't suck ✅
- [x] Configurable with Go code ✅
- [x] Rarely have to drop back to another shell ✅
- [x] Feels snappy and responsive ✅

## Known yaegi Limitations

1. **CGo**: Can't interpret CGo code
2. **Generics**: Limited support (improving)
3. **Unsafe**: Some unsafe operations restricted

**These don't matter for a shell REPL - we're doing basic scripting, not systems programming.**

## Repository Structure

```
gosh/
├── main.go              # Entry point
├── repl.go              # Core REPL loop with multiline support
├── router.go            # Smart routing logic with command substitution detection
├── evaluator.go         # yaegi wrapper with command substitution processing
├── spawner.go           # Command execution with signal handling
├── builtins.go          # Built-in commands
├── state.go             # State management with process tracking
├── types.go             # Shared types
├── go.mod               # Dependencies (yaegi + readline)
├── README.md            # This file
└── gosh-design-doc.md   # Original design documentation
```

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
