# gosh Architecture

> Technical documentation for the block-based hybrid shell redesign.

## Overview

gosh is a Go-first shell REPL that treats shell commands as first-class citizens alongside Go code. This document describes the redesigned block-based architecture that replaces the heuristic context detection approach.

## Core Design Principles

- **Explicit over implicit**: Mode is declared, never guessed
- **Blocks as the unit of execution**: Not lines
- **Shell by default**: Zero friction for normal terminal use
- **Go when you want it**: Explicit mode switch, no magic detection
- **History is block-aware**: Up arrow returns the whole block, not just the last line

---

## The Block Model

### The Problem with Line-Oriented Shells

The previous architecture used `IsGoContext()` — a heuristic function that tried to guess whether the current input was Go or shell. This caused:

- Misclassification at edge cases
- Fragile pattern matching that needed constant maintenance
- History that only remembered the last line of multiline input
- LSP and autocomplete that didn't know what context they were in

### Blocks as the Solution

A **block** is the atomic unit of gosh. Every input — whether one line or twenty — belongs to a block. Blocks have an explicit type: `shell` or `go`.

```
$ ls -la -> dirs          ← shell block, output captured into `dirs`
───────────────────
drwxr-xr-x  5 rjs
drwxr-xr-x  3 rjs
───────────────────

go> for _, d := range dirs {
...     fmt.Println(d.Name)
... }
───────────────────
rjs
rjs
───────────────────
```

### Mode Switching

- Default mode is **shell** — gosh behaves like a normal terminal
- `:go` switches to Go mode, prompt changes to `go>`
- `:sh` switches back to shell mode, prompt changes to `$`
- Mode persists until explicitly switched

No heuristics. No guessing. The prompt tells you where you are.

### Continuation Prompts

Multiline Go is handled by syntax, not by Shift+Enter:

```
go> for _, d := range dirs {
...     fmt.Println(d.Name)
... }
```

An open `{`, `(`, or `[` automatically triggers continuation with `...` prompt. Execution happens when the block is syntactically complete. This mirrors how Python's REPL handles multiline input and removes the need for special key bindings.

---

## Variable Capture with `->`

Shell output can be captured into Go variables using the `->` syntax:

```
$ ls -la -> dirs
$ ps aux -> processes
```

- `-> varname` on a shell command captures stdout into `varname` as a `[]string` (split on newlines)
- The variable is immediately available in any subsequent Go block
- Shell commands **without** `->` print normally and don't pollute the Go namespace
- Multiple captures accumulate in session state — `dirs` and `processes` both exist simultaneously

### Type of Captured Variables

Captured variables are `[]string` by default. This means in Go blocks:

```
go> for _, d := range dirs {
...     fmt.Println(d)
... }
```

Just works. No parsing needed for the common case.

---

## Input Layer — Bubbletea/Bubbles

### Why Replace readline

`chzyer/readline` is line-oriented. It has no concept of blocks, which means:

- Multiline history stores lines, not blocks
- No path to block-aware history navigation
- Rendering separators and block boundaries is fighting the library
- LSP integration has no surface to attach to

### Bubbletea Model

The input layer is replaced with [Bubbletea](https://github.com/charmbracelet/bubbletea) + the `bubbles/textarea` component.

Key properties:

- `textarea` natively handles multiline input
- The Elm architecture (Model/Update/View) makes block state explicit and testable
- Rendering is owned by gosh — separators, prompts, and output formatting are first-class
- Path toward syntax highlighting (in TODO) via lipgloss styling

### Block-Aware History

History is stored as blocks, not lines:

```go
type HistoryBlock struct {
    Mode    BlockMode  // Shell or Go
    Input   string     // Full block content (may be multiline)
    Output  string     // Output produced
    Capture string     // Variable name if -> was used, empty otherwise
}
```

Up/down arrow navigation returns whole blocks. When a block is recalled into the textarea, the full multiline content is restored and editable before re-execution.

History is persisted to `~/.gosh_history` as newline-delimited JSON blocks.

---

## Data Flow

```
User Input (bubbletea textarea)
        │
        ▼
  Block Finalizer
  (Enter pressed, braces balanced)
        │
        ├── Mode: Shell ──► Shell Executor ──► stdout/stderr
        │                        │
        │                   -> capture? ──► Session State (varname = []string)
        │
        └── Mode: Go ────► Yaegi Interpreter
                                 │
                            Session State (all captured vars injected as globals)
                                 │
                            stdout/stderr
        │
        ▼
  Output Renderer
  (separator, prompt, next block)
        │
        ▼
  History Store
  (append HistoryBlock)
```

---

## Session State

Session state is the shared context between shell and Go blocks:

```go
type SessionState struct {
    CapturedVars map[string][]string  // Shell captures available to Go
    GoInterp     *interp.Interpreter  // Yaegi instance with persistent state
    WorkingDir   string               // Current directory
    Mode         BlockMode            // Current input mode
    History      []HistoryBlock       // Block history
}
```

Yaegi's interpreter is long-lived across the session. Captured variables from shell blocks are injected into the interpreter's scope so they're available as normal Go variables.

---

## Rendering

### Prompts

| Mode         | Prompt |
| ------------ | ------ |
| Shell        | `$`    |
| Go           | `go>`  |
| Continuation | `...`  |

### Output Separators

Lightweight separators appear only when a block produces output:

```
$ ls -la -> dirs
───────────────────
drwxr-xr-x  5 rjs
───────────────────
```

Blocks with no output produce no separator — the prompt returns immediately, matching normal shell behavior.

Separator character: `─` (U+2500). Width: terminal width.

### Error Rendering

Errors are rendered inline with the block output, not as separate UI elements:

```
go> fmt.Println(missing)
───────────────────
error: undefined: missing
───────────────────
```

---

## Configuration

Configuration is unchanged from the current architecture. Two layers:

**Layer 1: `~/.bash_profile` / `~/.zprofile`**
Standard shell environment loaded on startup. PATH, exports, etc.

**Layer 2: `~/.config/gosh/config.go`**
Go file interpreted by Yaegi on startup. Custom functions, aliases as Go functions, environment setup in Go.

The config.go approach is a genuine strength of gosh — your shell functions are real Go code with types, error handling, and IDE support.

---

## Dependencies

```go
require (
    github.com/charmbracelet/bubbletea  // Input model and rendering
    github.com/charmbracelet/bubbles    // textarea component
    github.com/charmbracelet/lipgloss   // Terminal styling
    github.com/traefik/yaegi           // Go interpreter
)
```

`chzyer/readline` is removed.

---

## Migration from v0.x

The external interface is largely unchanged:

- `$()` command substitution syntax still works (backward compat)
- `-> varname` is the new preferred capture syntax
- `:go` / `:sh` replace implicit context detection
- `config.go` and `shellapi` are unchanged
- Homebrew install path unchanged

The `IsGoContext()` function and associated heuristics are deleted entirely.

---

## Future Work

- **Syntax highlighting**: lipgloss + treesitter for Go blocks in the textarea
- **LSP integration**: With explicit block types, Go blocks have a clear surface for gopls attachment
- **Named blocks**: Save a block with a name, recall by name from history
- **`.gosh` files**: Serialized sessions that can be replayed
- **Project detection**: Auto-load a `config.go` from project root alongside global config
- **Streaming output**: Real-time output for long-running shell commands
- **Background processes**: `&` syntax for non-blocking shell execution
