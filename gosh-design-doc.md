# gosh Design Document

> A Go shell with yaegi - because waiting 12 seconds for a REPL is unacceptable

## Overview

**gosh** is a hybrid shell that combines Go's interpreter (via yaegi) with traditional command execution. Cross-platform capable, designed as a daily driver, with instant startup and no PTY complexity.

## Core Philosophy

- **Go-first**: The shell is fundamentally a Go interpreter with shell conveniences
- **No shell syntax parsing**: We're not reimplementing bash. Bare commands work, everything else is Go
- **Hybrid approach**: `ls` just works, but `files := $(ls)` is also valid
- **Instant startup**: No waiting for REPLs to initialize
- **Daily driver quality**: Stable, fast, and pleasant to use

## Why gosh > shwift

| Aspect           | shwift                        | gosh                                         |
| ---------------- | ----------------------------- | -------------------------------------------- |
| Startup time     | 12 seconds                    | ~10ms                                        |
| Architecture     | PTY + external process        | Embedded interpreter                         |
| State management | Parse REPL output             | Native Go values                             |
| Complexity       | openpty(), filtering, parsing | Simple API calls                             |
| Persistence      | Hope REPL keeps state         | Direct variable storage                      |
| Platform         | macOS only                    | Cross-platform (though we can stay Mac-only) |

## Architecture

### The Core Loop

```
1. Display prompt (shows current directory with ~ substitution)
2. Read line of input (using bufio.Scanner or readline library)
3. Parse and route input (smart heuristics + PATH checking)
4. Execute (yaegi eval, process spawn, or built-in)
5. Handle output and errors
6. Update state (working directory, Go interpreter state)
7. Repeat
```

### Core Components

#### 1. Router

**Same as shwift but adapted for Go syntax:**

- Takes input and decides: Go eval, process spawn, or built-in
- Smart heuristics for routing:
  - Built-ins checked first (`cd`, `exit`)
  - Go syntax markers (`var`, `const`, `func`, `type`, `struct`, `interface`, `import`, `for`, `range`, `if`, `switch`)
  - Assignment or closures (`:=`, `=`, `{`)
  - Function calls with string literals → Go
  - Common Go functions (`fmt.Println`, `len`, `cap`, `make`, `append`, `copy`)
  - Fallback to PATH check for commands with parentheses
- Simple quote-aware argument parser

#### 2. Go Evaluator

**The Game Changer:**

```go
type GoEvaluator struct {
    interp *interp.Interpreter
    // That's literally it. No PTY, no process, no parsing.
}

func NewGoEvaluator() *GoEvaluator {
    i := interp.New(interp.Options{
        GoPath: os.Getenv("GOPATH"),
    })
    i.Use(stdlib.Symbols)  // All standard library
    i.Use(unsafe.Symbols)  // If we need it

    return &GoEvaluator{interp: i}
}

func (g *GoEvaluator) Eval(code string) (string, error) {
    // Capture stdout
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w

    _, err := g.interp.Eval(code)

    w.Close()
    os.Stdout = old

    var buf bytes.Buffer
    io.Copy(&buf, r)

    return buf.String(), err
}
```

**State persistence:**

- Variables defined persist automatically
- Functions defined persist automatically
- Imports persist automatically
- No parsing, no filtering, no hoping

**Pre-loading shell utilities:**

```go
// Inject helper functions into the interpreter
g.interp.Eval(`
import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// Helper for command output capture
func $(cmd string) string {
    // This would call back into gosh's process spawner
    return ""
}
`)
```

#### 3. Process Spawner

**Same as shwift:**

- Resolves executables in PATH
- PATH checking exposed for router use
- Spawns processes with proper stdio handling
- Returns exit codes

```go
type ProcessSpawner struct {
    state *ShellState
}

func (p *ProcessSpawner) Execute(command string, args []string) (int, error) {
    cmd := exec.Command(command, args...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Dir = p.state.WorkingDirectory
    cmd.Env = p.state.EnvironmentVars()

    err := cmd.Run()
    return cmd.ProcessState.ExitCode(), err
}

func FindInPath(command string) (string, bool) {
    path, err := exec.LookPath(command)
    return path, err == nil
}
```

#### 4. Built-in Handler

**Same as shwift:**

- `cd` - with path expansion
- `exit` - sets shouldExit flag
- Easy to extend

#### 5. State Management

**Simpler than shwift:**

```go
type ShellState struct {
    WorkingDirectory string
    Environment      map[string]string
    ShouldExit       bool
    ExitCode         int
}

func NewShellState() *ShellState {
    wd, _ := os.Getwd()
    env := make(map[string]string)
    for _, e := range os.Environ() {
        pair := strings.SplitN(e, "=", 2)
        if len(pair) == 2 {
            env[pair[0]] = pair[1]
        }
    }

    return &ShellState{
        WorkingDirectory: wd,
        Environment:      env,
    }
}
```

### Signal Handling

**Standard Go signal handling:**

```go
import (
    "os"
    "os/signal"
    "syscall"
)

func setupSignals(state *ShellState) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        for sig := range sigChan {
            switch sig {
            case os.Interrupt:
                // Ctrl+C - kill foreground process if any
                fmt.Println() // newline after ^C
            case syscall.SIGTERM:
                state.ShouldExit = true
            }
        }
    }()
}
```

## Implementation Phases

### Phase 1: MVP

- [ ] Basic REPL loop with prompt
- [ ] yaegi integration with state persistence
- [ ] Bare command execution (no pipes, no redirects)
- [ ] Built-in `cd` and `exit`
- [ ] Ctrl+C/D handling
- [ ] Basic error messages
- [ ] Path expansion (tilde, env vars)
- [ ] Smart routing

### Phase 2: Daily Driver Basics

- [ ] Command history (readline library or custom)
- [ ] Tab completion (at least for files/dirs)
- [ ] Better error messages
- [ ] `$(command)` syntax for output capture
- [ ] Improved signal handling for foreground processes

### Phase 3: Nice to Have

- [ ] Pipe support (`ls | grep foo`)
- [ ] Redirect support (`ls > file.txt`)
- [ ] Background jobs (`long_command &`)
- [ ] Job control (fg, bg, jobs)
- [ ] Config file (.goshrc?)
- [ ] Prompt customization
- [ ] Aliases

### Phase 4: Polish

- [ ] Syntax highlighting
- [ ] Better tab completion (command-aware)
- [ ] Multiline input support
- [ ] Pretty output formatting
- [ ] Preload common packages for speed

## Technical Decisions

### Why yaegi?

- **Embedded**: It's a library, not a separate process
- **Fast**: Interprets efficiently, no compilation wait
- **Complete**: Full Go language support (with some limitations)
- **Maintained**: Used in production by Traefik
- **No CGo**: Pure Go, easy to cross-compile

### Why Go?

- **Fast compilation**: If we need to compile plugins/extensions
- **Great stdlib**: `os/exec`, `path/filepath`, `bufio`, etc.
- **Simple concurrency**: Goroutines for async operations
- **Single binary**: Easy distribution
- **Cross-platform**: Works on macOS and Linux (Windows users can use PowerShell)

### yaegi Integration

**No complexity, just this:**

```go
import (
    "github.com/traefik/yaegi/interp"
    "github.com/traefik/yaegi/stdlib"
)

interpreter := interp.New(interp.Options{})
interpreter.Use(stdlib.Symbols)

// That's it. Now evaluate:
result, err := interpreter.Eval(code)
```

**State persists automatically:**

```go
// First command
interpreter.Eval(`x := 42`)

// Later command - x is still there!
interpreter.Eval(`fmt.Println(x)`)  // Prints 42
```

**No PTY, no parsing, no filtering needed.**

### Input Handling Options

**Option 1: Simple (MVP)**

```go
scanner := bufio.NewScanner(os.Stdin)
for scanner.Scan() {
    input := scanner.Text()
    // process input
}
```

**Option 2: Readline library (Phase 2)**

```go
import "github.com/chzyer/readline"

rl, _ := readline.New("> ")
defer rl.Close()

for {
    line, err := rl.Readline()
    // Auto history and editing
}
```

### Command Output Capture

Inject a helper function:

```go
// During startup, eval this:
interpreter.Eval(`
func $(cmd string) string {
    // This calls back to gosh's process spawner
    // Returns command output as string
}
`)

// Then users can do:
// files := $(ls -la)
```

We'll need to register a Go function that yaegi can call.

## Repository Structure

```
gosh/
├── main.go              # Entry point
├── repl.go              # Core REPL loop
├── router.go            # Smart routing logic
├── evaluator.go         # yaegi wrapper
├── spawner.go           # Command execution
├── builtins.go          # Built-in commands
├── state.go             # State management
├── types.go             # Shared types
├── go.mod               # Dependencies
└── README.md
```

## Dependencies

```go
module github.com/yourusername/gosh

go 1.21

require (
    github.com/traefik/yaegi v0.15.1
    github.com/chzyer/readline v1.5.1  // Phase 2
)
```

## Success Criteria

**MVP Success:**

- [ ] Starts instantly (< 100ms)
- [ ] Can run basic commands (`ls`, `git status`, etc.)
- [ ] Can write Go code with persistent state
- [ ] Doesn't crash on Ctrl+C
- [ ] Can `cd` around

**Daily Driver Success:**

- [ ] Want to use it instead of zsh
- [ ] Tab completion works well enough
- [ ] Command history doesn't suck
- [ ] Rarely have to drop back to another shell
- [ ] Feels snappy and responsive

## Known yaegi Limitations

1. **Reflection**: Some advanced reflection doesn't work
2. **CGo**: Can't interpret CGo code
3. **Generics**: Limited support (improving)
4. **Unsafe**: Some unsafe operations restricted

**These don't matter for a shell REPL. We're doing basic scripting, not systems programming.**

## Example Session

```
gosh> ls
file1.txt  file2.go  dir/

gosh> files := $(ls)

gosh> fmt.Println(strings.Split(files, "\n"))
[file1.txt file2.go dir/]

gosh> for _, f := range strings.Split(files, "\n") {
...     if strings.HasSuffix(f, ".go") {
...         fmt.Println("Go file:", f)
...     }
... }
Go file: file2.go

gosh> import "path/filepath"

gosh> filepath.Ext("test.go")
".go"

gosh> cd ..

gosh> pwd
/Users/you/parent

gosh> exit
```

**All instant. No waiting. No PTY nonsense.**

## Next Steps

1. Set up Go module and basic project structure
2. Implement basic REPL loop with yaegi
3. Add router for command vs Go detection
4. Implement process spawner
5. Add built-ins (cd, exit)
6. Test extensively
7. Add readline for history/completion (Phase 2)

## The Bottom Line

shwift's biggest problem was architectural: PTY + external REPL = complexity and latency. gosh solves this by embedding the interpreter. Everything becomes simpler, faster, and more reliable.

**No more 12-second startup. No more PTY parsing. Just a fast, clean hybrid shell.**

---
