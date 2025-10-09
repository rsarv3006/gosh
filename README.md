# gosh - Go Shell

> A hybrid shell that combines Go's interpreter (via yaegi) with traditional command execution.

## Features

- **Instant startup**: No waiting for REPLs to initialize (looking at you, Swift)
- **Go REPL**: Write Go code directly in your shell with persistent state
- **Traditional commands**: Just works - `ls`, `git status`, etc.
- **Hybrid mode**: Mix Go code and shell commands seamlessly
- **Built-ins**: `cd`, `exit`, `pwd` with path expansion
- **macOS & Linux**: Windows users can use PowerShell

## Quick Start

```bash
# Install
go install github.com/yourusername/gosh@latest

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

gosh> for i := 0; i < 3; i++ {
...     fmt.Println("Hello", i)
... }
Hello 0
Hello 1
Hello 2

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
```

## How It Works

- **yaegi**: Embedded Go interpreter, no external processes
- **Smart routing**: Detects Go syntax vs shell commands
- **State persistence**: Variables, functions, imports all persist
- **PATH resolution**: Finds executables just like your normal shell

## Building

```bash
git clone https://github.com/yourusername/gosh
cd gosh
go build
./gosh
```

## Architecture

```
Input → Router → [Go Evaluator | Process Spawner | Builtins] → Output
                      ↓                ↓                ↓
                    yaegi          exec.Command      built-in
```

## Why?

Because waiting 12 seconds for Swift's REPL to start is unacceptable. Go + yaegi = instant startup with full language features.

## Status

**MVP Complete** (Phase 1):

- ✅ Basic REPL loop
- ✅ Go evaluation with yaegi
- ✅ Command execution
- ✅ Built-ins (cd, exit, pwd)
- ✅ Path expansion
- ✅ Smart routing

**TODO** (Phase 2):

- [ ] Command history
- [ ] Tab completion
- [ ] Better error messages
- [ ] `$(command)` output capture

## License

MIT

## Contributing

PRs welcome! This is a fun project to learn Go and build something useful.
