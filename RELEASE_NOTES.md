# gosh MVP 0.0.1 Release Notes

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

Happy hacking! 🐹
