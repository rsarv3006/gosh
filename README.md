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
- **Built-ins**: `cd`, `exit`, `pwd`, `help`, `init` with path expansion
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
# Install the latest release
go install github.com/rsarv3006/gosh@latest

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

### Shellapi Functions - Working Command Execution

**🎉 NEW: Working shellapi functions that actually execute commands!**

gosh now supports shellapi functions that execute real commands (not just command substitution). These functions use Go's `os/exec` internally and work with full error handling.

```go
// ~/.config/gosh/config.go
package main

import (
    "fmt"
    "github.com/rsarv3006/gosh_lib/shellapi"  // Import shellapi functions
)

func init() {
    fmt.Println("🚀 gosh config loaded! Command execution system enabled!")
}

// Development helper functions that actually work
func build() string {
    result, err := shellapi.GoBuild()
    if err != nil {
        return "BUILD ERROR: " + err.Error()
    }
    return "BUILD SUCCESS: " + result
}

func test() string {
    result, _ := shellapi.GoTest()
    return result  // Returns actual test output
}

func run() string {
    result, _ := shellapi.GoRun()
    return result  // Returns actual program output
}

func gs() string {
    result, err := shellapi.GitStatus()
    if err != nil {
        return "GIT ERROR: " + err.Error()
    }
    return "GIT STATUS:\n" + result
}

// Directory changing functions that actually change directories!
func goGosh() string {
    result, err := shellapi.RunShell("cd", "/Users/rjs/dev/gosh")
    if err != nil {
        return "CD ERROR: " + err.Error()
    }
    return result  // Returns CD marker for processing
}

func goConfig() string {
    result, err := shellapi.RunShell("cd", "/Users/rjs/.config/gosh/")
    if err != nil {
        return "CD ERROR: " + err.Error()
    }
    return result  // Returns CD marker for processing
}
```

**Usage Examples:**

```bash
# These actually execute real commands now!
gosh> build()           # Executes real go build command
gosh> test()            # Executes real go test command  
gosh> run()             # Executes real go run . command
gosh> gs()              # Executes real git status with full output

# These actually change directories!
gosh> goGosh()          # Changes to ~/dev/gosh - directory persists!
gosh> goConfig()        # Changes to ~/.config/gosh/ - directory persists!

# Direct shellapi access also works
gosh> shellapi.GoBuild()
gosh> shellapi.GitStatus()
gosh> shellapi.RunShell("ls", "-la")
```

## 🔧 RunShell Command Documentation

The `shellapi.RunShell` function is the core command execution engine that powers all shellapi functions. It executes real shell commands using Go's `os/exec` and provides flexible command execution with proper output capture.

### **Basic Syntax**
```go
result, err := shellapi.RunShell("command", "arg1", "arg2", ...)
```

### **Command Types**

#### **Standard Shell Commands**
```go
// List files with colors
files, err := shellapi.RunShell("ls", "--color=auto")

// Get current directory
pwd, err := shellapi.RunShell("pwd")

// Show git status
status, err := shellapi.RunShell("git", "status")
```

#### **Directory Changes (Special Handling)**
```go
// CD commands return markers that actually change directories
result, err := shellapi.RunShell("cd", "/path/to/project")
if err != nil {
    return "CD ERROR: " + err.Error()
}
return result  // Returns @GOSH_INTERNAL_CD:/path/to/project marker
```

**How CD Integration Works:**
1. `RunShell("cd", path)` returns a CD marker (`@GOSH_INTERNAL_CD:/path`)
2. The evaluator detects this marker 
3. The shell's working directory is actually changed
4. The change persists across the entire shell session

#### **Development Scripts**
```go
// Build project  
build := shellapi.RunShell("go", "build")

// Run tests  
tests := shellapi.RunShell("go", "test", "./...")

// Start application
app := shellapi.RunShell("go", "run", ".")
```

#### **System Commands**
```go
// Get system info
uptime := shellapi.RunShell("uptime")
date := shellapi.RunShell("date")

// Process management  
ps := shellapi.RunShell("ps", "aux")

// Network tools
ping := shellapi.RunShell("ping", "-c", "3", "google.com")
```

### **Error Handling Patterns**

#### **Simple Command Execution**
```go
func getStatus() string {
    result, err := shellapi.RunShell("git", "status")
    if err != nil {
        return "ERROR: " + err.Error()
    }
    return result
}
```

#### **Directory Change with Error Handling**
```go
func goToProject() string {
    result, err := shellapi.RunShell("cd", "/Users/rjs/projects/myapp")
    if err != nil {
        return "CD ERROR: " + err.Error()  
    }
    return result  // Silent success - directory actually changed
}
```

#### **Multi-step Operations**
```go
func buildAndTest() string {
    // Build first
    build, buildErr := shellapi.RunShell("go", "build")
    if buildErr != nil {
        return "BUILD FAILED: " + buildErr.Error()
    }
    
    // Then test
    test, testErr := shellapi.RunShell("go", "test")
    if testErr != nil {
        return "TESTS FAILED: " + testErr.Error()
    }
    
    return "✅ Build successful\n" + test
}
```

### **Integration Examples**

#### **Custom Shell Functions**
```go
func deployTo(prodEnv string) string {
    msg := "Deploying to " + prodEnv + "...\n"
    
    // Check git status
    status, _ := shellapi.RunShell("git", "status")
    if status != "" {
        msg += "⚠️  Working tree not clean:\n" + status
        return msg
    }
    
    // Deploy
    result, err := shellapi.RunShell("ansible-playbook", "deploy.yml", "-e", "env="+prodEnv)
    if err != nil {
        return "DEPLOY ERROR: " + err.Error()
    }
    
    return msg + "✅ Deployment complete"
}
```

#### **Interactive Tools Support**
```go
func openEditor() string {
    // Opens vim - will wait for user input
    result, err := shellapi.RunShell("nvim", "config.yaml")
    return result  // ( vim blocks until user quits )
}
```

### **Best Practices**

#### **✅ Do:**
- Use proper error handling for all commands
- Check `err != nil` before using results
- Return error messages for better user feedback
- Use CD markers for directory changes
- Handle empty/missing output appropriately

#### **❌ Avoid:**
- Running long-running interactive programs in wrapper functions
- Assuming commands will always succeed  
- Ignoring error return values
- Hard-coding absolute paths in public functions

### **Performance Notes**
- Commands execute in real shell environment (not command substitution)
- Output is fully captured and returned as strings
- Directory changes persist across shell sessions
- Error messages are captured from stderr and stderr streams

**Key Benefits:**
- ✅ Real command execution via Go's `os/exec`
- ✅ Proper stderr/stdout capture
- ✅ Cross-platform compatibility
- ✅ Persistent directory state
- ✅ Full error reporting

**Key Benefits:**
- ✅ **Real Command Execution**: Functions execute actual commands, not just command substitution strings
- ✅ **Proper Error Handling**: Full error reporting with actual command errors
- ✅ **Directory Persistence**: CD commands actually change directories in the shell session
- ✅ **Full Output**: Commands return their actual output, success/failure status
- ✅ **Working with All Commands**: Works with `git`, `go`, Docker, npm, etc.

**Available Working Functions:**

- **Development Tools**: `GoBuild()`, `GoTest()`, `GoRun()` - execute real Go commands
- **Git Operations**: `GitStatus()` - shows actual git status  
- **Shell Commands**: `RunShell(cmd, args...)` - execute any shell command
- **Directory Changes**: `RunShell("cd", path)` - actually changes directories
- **Color Functions**: `Success()`, `Warning()`, `Error()` - format text with colors
- **File Operations**: `LsColor()` - colorful file listings

**How Directory Changes Work:**

The `RunShell("cd", path)` function returns a special CD marker (`@GOSH_INTERNAL_CD:/path`) that the evaluator detects and processes to actually change the shell's working directory.

```go
func goProject() string {
    result, err := shellapi.RunShell("cd", "/path/to/project")
    if err != nil {
        return "ERROR: " + err.Error()
    }
    return result  // Silent success - directory actually changes!
}
```

**Requirements for CD Functions:**

- Function must return a `string`
- Must return the result from `shellapi.RunShell("cd", path)`
- Successful CD returns no output (silent success)
- Errors are properly reported
```

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
- ✅ Built-ins (cd, exit, pwd, help, init)
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

**🚀 Phase 3 Complete**:

- ✅ Command history navigation (up/down arrows)
- ✅ Better error messages with line numbers
- ✅ Git integration in prompt

**🔧 Phase 4 Complete - Working Shellapi Functions**:

✅ **Real Command Execution (v0.2.2)**
gosh v0.2.2 features shellapi functions that execute real commands via Go's `os/exec`, providing actual command output and persistent directory changes.

**🎉 Phase 5 Complete - Sequential Directory Operations (v0.2.4)**:

✅ **Fixed Critical Bug** - Directory changes within functions now work correctly. Previously, only the last `cd` operation would persist within a function, breaking expected shell workflow patterns. This fix enables multi-step shell workflows with immediate directory changes and proper prompt updates.

```go
// This now works perfectly (v0.2.4+)
func deployWorkflow() {
    goConfig()    // Changes to config directory immediately
    loadConfig()  // Operations happen in config directory
    goProject()    // Changes to project directory immediately  
    build()        // Operations happen in project directory
    goDeploy()     // Changes to deploy location immediately
    copyFiles()   // Operations happen in deploy directory
}
```

```go
// ~/.config/gosh/config.go
package main

import (
    "fmt"
    "github.com/rsarv3006/gosh_lib/shellapi"
)

func init() {
    fmt.Println("🚀 gosh config loaded! Command execution system enabled!")
}

// Functions that actually work:
func build() string {
    result, err := shellapi.GoBuild()
    if err != nil {
        return "BUILD ERROR: " + err.Error()
    }
    return "BUILD SUCCESS: " + result  // Real go build execution
}

func test() string {
    result, _ := shellapi.GoTest()
    return result  // Returns actual test output
}

func goGosh() string {
    result, _ := shellapi.RunShell("cd", "/Users/rjs/dev/gosh")
    return result  // Directory actually changes and persists!
}

func gs() string {
    result, err := shellapi.GitStatus()
    if err != nil {
        return "GIT ERROR: " + err.Error()
    }
    return "GIT STATUS:\n" + result  // Real git status with colors
}

// Usage examples:
// gosh> build()        # Executes real go build with feedback
// gosh> test()         # Executes real go test with output
// gosh> goGosh()      # Actually changes directory (persists!)
// gosh> shellapi.GoBuild()  # Direct access also works
```

**Working Features:**
- **🔧 Development Commands** - `GoBuild()`, `GoTest()`, `GoRun()` execute real Go commands
- **📁 File Operations** - `LsColor()`, file operations with actual filesystem access
- **🔀 Git Tools** - `GitStatus()` shows real git repository status  
- **🖥️ Shell Commands** - `RunShell(cmd, args...)` executes any shell command
- **🎨 Color Functions** - `Success()`, `Warning()`, `Error()` format output with colors
- **📂 Directory Changes** - `RunShell("cd", path)` actually changes directories

**Key Benefits:**
- ✅ **Real Command Execution** - Functions use Go's `os/exec` for actual command execution
- ✅ **Directory Persistence** - CD commands maintain state across shell sessions  
- ✅ **Proper Error Handling** - Real command errors are captured and returned
- ✅ **Full Output Capture** - Commands return their actual output and status
- ✅ **Working Integration** - Direct shellapi access works alongside wrapper functions

**Usage Examples:**
```bash
gosh> build()        # Executes real go build with feedback
gosh> test()         # Executes real go test with output
gosh> goGosh()      # Actually changes directory (persists!)
gosh> shellapi.GoBuild()  # Direct access also works
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

### Shellapi Enhancements - TODO:

- [ ] **Interactive program handling** - Better support for commands like `vim`, `nano` that wait for user input
- [ ] **Signal forwarding** - Proper Ctrl+C handling for long-running shellapi processes  
- [ ] **Output streaming** - Real-time output capture for long-running commands
- [ ] **Background processes** - Support for launching background processes via shellapi
- [ ] LSP integration for Go code editing in the shell
- [ ] Syntax highlighting for Go code input
- [ ] Documentation lookup (`go doc` integration)

### Sharp Edges
- [ ] gosh doesn't handle piped input well

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
