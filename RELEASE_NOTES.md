# gosh MVP 0.0.1 Release Notes

## 🚀 Major Release: Shellapi Integration!

**gosh v0.2.0** introduces a comprehensive shellapi integration that transforms gosh from a basic hybrid shell into a powerful extensible development environment with 100+ shell functions.

## 🎉 First Release!

**gosh v0.0.1** is a hybrid shell that combines Go's interpreter (via yaegi) with traditional command execution. This is the MVP release that provides a solid foundation for daily driver usage.

## What's in MVP 0.0.1

### ✅ Core Features
- **Instant startup** (~10ms) - no waiting for REPLs to initialize
- **Go REPL** - write Go code directly with persistent state
- **Shell commands** - `ls`, `git status`, `echo` just work
- **Multiline support** - functions, if statements, for loops with proper continuation
- **Command substitution** - `files := $(ls)` captures command output into Go strings
- **Smart routing** - automatically detects Go vs shell commands

### ✅ Daily Driver Features  
- **Built-in commands** - `cd`, `pwd`, `exit`, `help`
- **Tab completion** - for commands and file paths
- **Command history** - up/down arrow navigation
- **Config support** - `config.go` for shell customization
- **Color system** - with theme support
- **Signal handling** - proper Ctrl+C behavior

### ✅ Cross-platform
- **macOS & Linux** - Windows users should use PowerShell

## Quick Start

```bash
# Install
go install github.com/rsarv3006/gosh@latest

# Run
gosh
```

## Example Usage

```bash
# Mix shell commands and Go code
gosh> pwd
/Users/you/gosh
gosh> x := 42
gosh> fmt.Println(x * 2)
84

# Command substitution - game changer!
gosh> files := $(ls)
gosh> fmt.Println(len(strings.Split(files, "\n")))
12

# Multiline functions
gosh> func add(a, b int) int {
...     return a + b
... }
gosh> fmt.Println(add(5, 3))
8
```

## Configuration

Create `config.go` in your working directory:

```go
package main
import "fmt"

func init() {
    fmt.Println("Welcome to my custom gosh!")
}

func hello(name string) {
    fmt.Printf("Hello %s!\n", name)
}
```

## What's Next?

Future releases will add:
- Pipe support (`ls | grep foo`)
- Background jobs (`long_command &`)  
- Better error messages with line numbers
- Git integration in prompt

## Known Limitations

- No pipe support yet (Phase 3 feature)
- Limited CGo support (yaegi limitation)
- Generics support improving (yaegi limitation)

## Feedback

This is an MVP release! Please file issues on GitHub for:
- Bugs and crashes
- Feature requests
- Usability issues

---

## 🚀 Major Release: Shellapi Integration!

**gosh v0.2.0** introduces a comprehensive shellapi integration that transforms gosh from a basic hybrid shell into a powerful extensible development environment with 100+ shell functions.

### ✅ Shellapi Integration - 100+ Functions Available

#### 📦 Comprehensive Function Library
- **📁 File Operations**: `Ls()`, `Cat()`, `Find()`, `Grep()`, `Touch()`, `MakeDir()`
- **🔀 Git Operations**: `GitStatus()`, `GitLog()`, `QuickCommit()`, `GitPull()`, `GitPush()`
- **🖥️ Development Tools**: `GoBuild()`, `GoTest()`, `NpmInstall()`, `DockerPs()`, `KubectlPods()`
- **🖥️ System Commands**: `Uptime()`, `Whoami()`, `Date()`, `Pwd()`, `EnvVar()`
- **🎨 Color Functions**: `Success()`, `Error()`, `Warning()`, `Bold()`, `Underline()`
- **🏗️ Project Utilities**: `MakeTarget()`, `BuildAndTest()`, `CreateProjectDir()`

#### 🛠️ Dual Usage Patterns

**Manual Wrappers (Recommended):**
```bash
~/dev/gosh > gs()           # Git status with automatic command substitution
~/dev/gosh > ok("Done!")    # Green success message
~/dev/gosh > build()        # Build project
~/dev/gosh > ls()           # Colorful file listing
```

**Direct Shellapi Access (Advanced):**
```bash
~/dev/gosh > shellapi.GitStatus()  # Direct shellapi call
~/dev/gosh > shellapi.Successmsg() # Direct color function
```

#### 🔧 Smart Architecture
- **Manual Wrapper Pattern**: Users control which functions to expose
- **Command Substitution**: Automatic `$(command)` processing in both paths
- **Import-Based Detection**: Shellapi functions only available when imported
- **Clean Namespace**: No automatic function injection cluttering REPL

#### 🚀 Enhanced Init Command
```bash
~/dev/gosh > init
Created /Users/rjs/.config/gosh/go.mod
Created /Users/rjs/.config/gosh/config.go
✅ gosh config directory initialized at /Users/rjs/.config/gosh
```

- Creates clean config with manual wrapper template
- Ready-to-use wrapper functions (`gs()`, `ok()`, `build()`, `test()`)
- Proper gosh_lib dependency setup

### 🔧 Under the Hood Improvements

#### **Enhanced Command Substitution Processing**
- Works for both manual wrapper functions and direct shellapi calls
- Seamless integration of shell command output into Go strings
- Color output preservation and display

#### **Robust Error Handling**
- Graceful fallbacks when shellapi functions fail
- Clean error messages for invalid function calls
- Panic recovery for yaegi interpreter crashes

#### **Import Detection System**
- Automatically detects gosh_lib imports in user configs
- Conditional shellapi bridge creation when needed
- Clean separation between user code and shell functions

### ⚡ Breaking Changes

- **Config Template Change**: New configs use manual wrapper pattern instead of examples
- **Function Availability**: Some old example functions replaced with proper shellapi calls
- **Init Behavior**: Enhanced init with proper setup instructions
- **Namespace Changes**: Cleanlier REPL without automatic function pollution

### 📖 Usage Examples

**Daily Driver Workflow:**
```bash
~/dev/gosh > init              # First time setup
~/dev/gosh > gs()               # Check git status
~/dev/gosh > build()           # Build current project
~/dev/gosh > ok("✅ Built!")   # Success message
~/dev/gosh > test()            # Run tests
```

**Advanced Usage:**
```bash
~/dev/gosh > files := $(shellapi.LsColor())
~/dev/gosh > fmt.Println(shellapi.Success("Files listed!"))
~/dev/gosh > if len($(shellapi.GitStatus())) == 0 {
...     fmt.Println(shellapi.Warning("Clean working directory"))
... }
```

### ✅ Backward Compatibility

- All existing v0.1.x features continue to work
- Existing user configs will be updated automatically if needed
- Shell commands and Go REPL unchanged
- Built-in commands (`cd`, `pwd`, `exit`, `help`) unchanged

---

**Ready for production use!** The shellapi integration makes gosh a truly powerful development environment while maintaining the simplicity of the core hybrid shell concept.

Happy hacking! 🐹
