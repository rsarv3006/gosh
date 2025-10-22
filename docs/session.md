# Session file

The `session` command opens a temporary Go file that mirrors the current REPL session. It is useful for quickly inspecting or editing the current state (functions, variables, and recent statements) in your editor.

Usage

- `session` — open the session file in $EDITOR (if set) or the system opener (open on macOS, xdg-open on Linux).
- `session --print` or `session -p` — print the path to the session file.
 - `session cleanup` or `session --cleanup` — remove old temporary session/workspace directories (keeps 5 most recent). Use this to free disk space when needed.

What the file contains

- `package main` and common imports (fmt)
- Function definitions are emitted at package level
- Executable statements are placed inside `func session() { ... }`

Notes

- The session file is rewritten after each executed Go input so editors show the current REPL state.
- gosh prunes old session/workspace temp directories from the system temp dir (keeps recent sessions and removes entries older than 24 hours) to avoid filling up /tmp.

Cleanup behavior

- By default gosh cleans up stale session/workspace temp directories at LSP startup, removing entries older than 24 hours while keeping a few recent sessions.
- The `session cleanup` command forces immediate pruning: it keeps the 5 most recent session/workspace directories and removes the rest.
