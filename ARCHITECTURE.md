# Gosh Architecture

This document describes the internal architecture and components of the gosh shell.

## Core Loop

```
1. Display prompt (shows current directory with ~ substitution)
2. Read line of input (using readline with multiline support)
3. Parse and route input (smart heuristics + PATH checking)
4. Execute (yaegi eval, process spawn, or built-in)
5. Handle output and errors
6. Update state (working directory, Go interpreter state)
7. Repeat
```

## Architecture Diagram

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   REPL Loop      │    │     Router      │    │  Evaluator       │
│   (repl.go)      │────▶│   (router.go)   │────▶│ (evaluator.go)   │
│                 │    │                 │    │                 │
│ - Prompt display │    │ - Input parsing  │    │ - yaegi interp. │
│ - Readline input │    │ - Syntax routing │    │ - Command subst.│
│ - Loop control   │    │ - Path checking  │    │ - Output capture │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                       │                       │
        ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Shell State     │    │  Process Spawn  │    │  Built-in Handler│
│   (state.go)     │    │   (spawner.go)  │    │   (builtins.go)  │
│                 │    │                 │    │                 │
│ - Working dir    │    │ - Exec commands │    │ - cd, pwd, exit│
│ - Environment   │    │ - Signal handling│    │ - Built-in ops   │
│ - Process state  │    │ - Stdio routing   │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                 │
                                 ▼
                    ┌─────────────────────────┐
                    │     OS Resources       │
                    │ (system calls, files)  │
                    └─────────────────────────┘
```

## Core Components

### 1. Router (`router.go`)
Takes input and decides: Go eval, process spawn, or built-in:

- **Built-ins checked first** (`cd`, `exit`, `pwd`, `help`)
- **Go syntax markers** (`var`, `const`, `func`, `type`, `struct`, `interface`, `import`, `for`, `range`, `if`, `switch`)
- **Assignment or closures** (`:=`, `=`, `{`)
- **Function calls with string literals** → Go
- **Common Go functions** (`fmt.Println`, `len`, `cap`, `make`, `append`, `copy`)
- **Command substitution syntax** `$(command)` → Go
- **Fallback to PATH check** for commands with parentheses

### 2. Go Evaluator (`evaluator.go`)
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

### 3. Process Spawner (`spawner.go`)
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

### 4. Built-in Handler (`builtins.go`)
- `cd` - with path expansion
- `exit` - sets shouldExit flag
- `pwd` - shows current working directory
- `help` - shows help information
- Easy to extend with new built-ins

### 5. State Management (`state.go`)
```go
type ShellState struct {
    WorkingDirectory string
    Environment      map[string]string
    ShouldExit       bool
    ExitCode         int
    CurrentProcess   *os.Process  // For signal handling
}
```

## Command Substitution

One of the most powerful features - `$(command)` syntax:

### How it works:
1. **Parse input** for `$(command)` patterns
2. **Execute command** in subprocess
3. **Capture output** as Go string literal
4. **Replace** `$(command)` with literal
5. **Evaluate** resulting Go code

### Example flow:
```go
// Input: files := $(ls)
// Step 1: Parse pattern $(ls)
// Step 2: Run "ls" command → output: "README.md\ngo.mod\ngo.sum\n"
// Step 3: Replace with literal: files := "README.md\ngo.mod\ngo.sum\n"
// Step 4: Evaluate Go eval
```

## Environment Management

### Hybrid Environment Strategy

gosh combines standard shell environment with Go-powered extensions:

#### Standard Shell Setup (`env.go`)

**Automatic Standard Config Loading:**
- Loads regular shell configs when run as login shell
- Supports: `.bash_profile`, `.zprofile`, `.profile`, `.bash_login`, `.login`
- Full POSIX environment inheritance
- Shell variable expansion `$HOME`, `$PATH`, `$GOPATH`

#### Go-Powered Extensions (Global Config Only)

**Single global config:** `~/.config/gosh/config.go`

```go
package main

import "os"

func init() {
    // Global environment setup
    os.Setenv("GOPATH", os.Getenv("HOME") + "/go")
    os.Setenv("EDITOR", "vim")
}

func globalTest() {
    println("Global function available everywhere!")
}
```

**Benefits:**
- ✅ One source of truth
- ✅ Consistent behavior across all directories
- ✅ No yaegi auto-loading headaches
- ✅ Simple and maintainable

## Signal Handling

Proper handling of Ctrl+C and other signals:

```go
// When Ctrl+C pressed:
// 1. Check if current process running
// 2. If so: forward signal to subprocess
// 3. If not: handle in gosh (interrupt Go eval)
// 4. Never crash, always clean up
```

## Error Handling

### Multiple Error Sources:
1. **Go evaluation errors** (yaegi syntax errors)
2. **Process spawn errors** (command not found, permissions)
3. **Built-in errors** (cd failures, etc.)

### Error Display:
- **Clean messages** for common errors
- **Detailed stack traces** for Go syntax errors
- **Exit codes** preserved for script compatibility

## Testing Strategy

### Current Test Coverage:
- Unit tests for core components
- Integration tests for shell functionality
- Regression testing for key features

### Test Categories:
- **Parser tests** - input routing and command detection
- **Eval tests** - Go code evaluation and command substitution
- **Built-in tests** - shell built-in commands
- **Environment tests** - config loading and variable expansion
- **Integration tests** - complete shell usage scenarios

## Future Extensibility

### Areas for Growth:
1. **More built-in commands** - history management, aliases
2. **Plugin system** - load extensions at runtime
3. **Project-specific configuration** - per-project Go configs
4. **Performance optimization** - faster startup, better resource usage
5. **Cross-platform enhancements** - Windows compatibility improvements

## Dependencies

```go
module github.com/rsarv3006/gosh

go 1.21

require (
    github.com/chzyer/readline v1.5.1    # Multiline input & history
    github.com/traefik/yaegi v0.15.1     # Go interpreter
    github.com/charmbracelet/lipgloss v1.1.0 # Color support
)
```

## Repository Structure

```
gosh/
├── main.go              # Entry point
├── repl.go              # Core REPL loop with multiline support
├── router.go            # Smart routing logic
├── evaluator.go         # yaegi wrapper with command substitution
├── spawner.go           # Command execution with signal handling
├── builtins.go          # Built-in commands
├── state.go             # State management
├── types.go             # Shared types
├── env.go               # Environment management  
├── colors.go            # Color system and display
├── completer.go        # Tab completion engine
├── config.go            # Configuration loading (removed - global only)
├── go.mod               # Dependencies
├── README.md            # User documentation
├── ARCHITECTURE.md      # This file - technical documentation
└── gosh                  # Compiled binary
```
