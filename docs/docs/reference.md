# CLI Reference

Complete reference for gosh command-line options, built-in commands, and shellapi functions.

## Command Line Options

### gosh

Start an interactive gosh session:

```bash
gosh
```

### Options

- `-v, --version` - Show version information
- `-h, --help` - Show help message
- `-c '<command>'` - Execute single command and exit

```bash
# Show version
gosh --version

# Show help
gosh --help

# Execute Go code
gosh -c 'fmt.Println("Hello world")'

# Execute shell command
gosh -c 'ls -la'

# Mixed command
gosh -c 'files := $(ls); fmt.Printf("Found %d files\n", len(strings.Split(files, "\n")))'
```

## Built-in Commands

### cd <path>

Change to the specified directory.

```bash
gosh> cd ~/projects
gosh> cd /tmp
gosh> cd ..        # Parent directory
gosh> cd ../sibling # Sibling directory
```

### pwd

Print the current working directory.

```bash
gosh> pwd
/Users/username/projects/gosh
```

### exit

Exit gosh and return to the previous shell.

```bash
gosh> exit
```

### help

Display help information for built-in commands.

```bash
gosh> help
Available built-ins:
- cd <path>          Change directory
- pwd                Print working directory  
- exit               Exit gosh
- init               Create example config
- help               Show this help
```

### init

Create an example configuration file at `~/.config/gosh/config.go`.

```bash
gosh> init
✅ Created example config: ~/.config/gosh/config.go
```

## Go REPL Features

### Variable Assignment

```bash
gosh> x := 42
gosh> name := "gosh"
gosh> files := $(ls)
```

### Function Definition

```bash
gosh> func add(a, b int) int {
...     return a + b
... }
```

### Control Structures

```bash
gosh> for i := 0; i < 3; i++ {
...     fmt.Println(i)
... }

gosh> if true {
...     fmt.Println("it's true")
... }
```

### Import Support

Common packages are pre-imported automatically:

- `fmt`
- `os`
- `strings` 
- `time`
- `filepath`
- `io/ioutil`
- `encoding/json`

```bash
gosh> fmt.Println("Hello")
gosh> os.Getenv("HOME")
gosh> strings.Split("a,b,c", ",")
gosh> time.Now()
```

## Command Substitution

### Syntax

```bash
variable := $(command)
```

### Examples

```bash
# Capture file list
files := $(ls)

# Capture git status
status := $(git status --porcelain)

# Capture file count
count := $(ls | wc -l)

# Use in expressions
if $(git status) == "" {
    fmt.Println("Working directory is clean")
}

# Complex commands
timestamp := $(date +%Y%m%d_%H%M%S)
backupFile := $(echo $HOME/.ssh/backup_${timestamp}.tar)
```

## Shellapi Functions Reference

### 📁 File Operations

| Function | Description | Purpose |
|----------|-------------|---------|
| `Ls()` | Basic file listing | `ls` |
| `LsColor()` | Colorized file listing | `ls --color=auto` |
| `LsSortBySize()` | Files sorted by size | `ls -S` |
| `Tree()` | Directory tree structure | `tree` |
| `Find(pattern)` | Find files by pattern | `find . -name pattern` |
| `Grep(pattern, args...)` | Search in files | `grep pattern file` |
| `Cat(file)` | Display file contents | `cat file` |
| `Head(n, file)` | First N lines of file | `head -n file` |
| `Tail(n, file)` | Last N lines of file | `tail -n file` |
| `Touch(filename)` | Create/update file timestamp | `touch filename` |
| `MakeDir(dir)` | Create directory | `mkdir dir` |
| `RemoveFile(file)` | Delete file | `rm file` |
| `RemoveDir(dir)` | Delete directory | `rm -rf dir` |
| `FileExists(path)` | Check if file exists | `test -f` |
| `IsDirectory(path)` | Check if path is directory | `test -d` |

### 🔧 Git Operations  

| Function | Description | Purpose |
|----------|-------------|---------|
| `GitStatus()` | Repository status | `git status` |
| `GitLog()` | Recent commits | `git log --oneline` |
| `GitBranch()` | Current branch name | `git branch --show-current` |
| `GitDiff()` | Git diff output | `git diff` |
| `GitAdd(files...)` | Add files to staging | `git add files` |
| `GitCommit(msg)` | Create commit | `git commit -m msg` |
| `QuickCommit(msg)` | Add all and commit | `git add . && git commit` |
| `GitPull()` | Pull latest changes | `git pull` |
| `GitPush()` | Push to remote | `git push` |
| `GitStashList()` | List stashes | `git stash list` |
| `GitStashPush(msg)` | Create stash | `git stash push -m msg` |

### 🛠 Development Tools

| Function | Description | Purpose |
|----------|-------------|---------|
| `GoBuild()` | Build Go project | `go build` |
| `GoRun()` | Run Go project | `go run .` |
| `GoTest()` | Run tests | `go test -v` |
| `GoTestRun()` | Build and test | `go test -v -run ./...` |
| `GoInstall(pkg)` | Install Go package | `go install pkg` |
| `GoFmt()` | Format Go code | `go fmt` |
| `GoGet()` | Get dependencies | `go get` |
| `GoTidy()` | Clean dependencies | `go mod tidy` |
| `GoVet()` | Run go vet | `go vet` |
| `NpmInstall(pkg)` | Install npm package | `npm install pkg` |
| `NpmRun(script)` | Run npm script | `npm run script` |
| `PipInstall(pkg)` | Install Python package | `pip install pkg` |
| `DockerPs()` | List containers | `docker ps` |
| `DockerImages()` | List images | `docker images` |
| `DockerLogs(container)` | Container logs | `docker logs container` |
| `DockerStop(container)` | Stop container | `docker stop container` |
| `DockerRm(container)` | Remove container | `docker rm container` |

### 💻 System Information

| Function | Description | Purpose |
|----------|-------------|---------|
| `Uptime()` | System uptime | `uptime` |
| `Whoami()` | Current user | `whoami` |
| `Date()` | Current date/time | `date` |
| `Hostname()` | System hostname | `hostname` |
| `OS()` | Kernel name | `uname -s` |
| `Arch()` | System architecture | `uname -m` |
| `Pwd()` | Working directory | `pwd` |
| `Df()` | Disk usage | `df -h` |
| `Free()` | Memory usage | `free -h` |
| `Ps()` | Running processes | `ps aux` |
| `Kill(pid)` | Terminate process | `kill pid` |

### 🎨 Formatting & Display

| Function | Description | Output |
|----------|-------------|--------|
| `Red(str)` | Red text | Text in red |
| `Green(str)` | Green text | Text in green |
| `Blue(str)` | Blue text | Text in blue |
| `Yellow(str)` | Yellow text | Text in yellow |
| `Purple(str)` | Magenta text | Text in magenta |
| `Cyan(str)` | Cyan text | Text in cyan |
| `Bold(str)` | Bold text | **Text** |
| `Underline(str)` | Underlined text | <u>Text</u> |
| `Italic(str)` | Italic text | *Text* |
| `Success(str)` | Green checkmark | ✓ Success |
| `Error(str)` | Red cross mark | ✗ Error |
| `Warning(str)` | Yellow triangle | ⚠ Warning |
| `Info(str)` | Blue info circle | ℹ Info |
| `SuccessMsg(label, msg)` | Success message with label | ✓ Label: Message |
| `ErrorMsg(msg)` | Error message with formatting | ✗ Error message |
| `WarnMsg(label, msg)` | Warning with label | ⚠ Label: Message |
| `InfoMsg(label, msg)` | Info message with label | ℹ Label: Message |

#### Color Usage Examples

```bash
# In config.go
func greetUser(name string) {
    greeting := shellapi.Green(fmt.Sprintf("Hello %s!", name))
    fmt.Println(greeting)
}

func showStatus(msg string) {
    status := shellapi.InfoMsg("Status", msg)
    fmt.Println(status)
}

func reportError(err string) {
    error := shellapi.ErrorMsg("Error", err)
    fmt.Println(error)
}
```

## gosh_lib - Complete Function Library

### Installation

Add gosh_lib to your Go module:

```bash
go get github.com/rsarv3006/gosh_lib/shellapi
```

### Import in Config

```go
// ~/.config/gosh/config.go
package main

import (
    "fmt"
    "github.com/rsarv3006/gosh_lib/shellapi"
)

func init() {
    fmt.Println("🚀 gosh with shellapi loaded!")
}
```

### RunShell Function - Core Engine

The `shellapi.RunShell` function executes real shell commands and handles special cases like directory changes:

```go
// Basic command execution
result, err := shellapi.RunShell("command", "arg1", "arg2")

// Directory changes that persist
func goProject() string {
    result, err := shellapi.RunShell("cd", "/path/to/project")
    if err != nil {
        return "CD ERROR: " + err.Error()
    }
    return result  // Directory actually changes!
}

// System commands
msg, err := shellapi.RunShell("uptime")
files, err := shellapi.RunShell("ls", "-la")

# Multi-argument commands with spaces
result, err := shellapi.RunShell("git", "commit", "-m", "Fixed bug in user login")
```

### Directory Change Integration

Directory changes work seamlessly with the gosh shell:

```go
func navigationExample() string {
    // This actually changes directories in the shell session!
    if result, err := shellapi.RunShell("cd", "/tmp"); err == nil {
        return "Switched to temp directory"
    }
    return "Failed to change directory"
}

// Usage in gosh:
// gosh> navigationExample()
// Switched to temp directory  
// gosh> pwd
// /tmp   # Directory actually changed!
```

### Complete Function Categories

#### Project Helpers

```go
MakeTarget(target)       # Run make target
MakeBuild()              # make build
MakeClean()             # make clean  
MakeTest()              # make test
RunTests()              # Build and test
BuildAndTest()          # Build and test combined
CreateProjectDir(name)  # Create project directory
TouchFile(filename)     # Touch file
JoinPaths(parts...)     # Join paths safely
Basename(path)          # Get filename only
Dirname(path)           # Get directory only
ExpandUserHome(path)     # Expand ~ to home
GetScriptDir()          # Get script directory
```

#### Text Processing

```go
Clean(str)              # Clean string
Trim(str)               # Trim whitespace
Upper(str)              # Uppercase
Lower(str)              # Lowercase
Title(str)              # Title case
Replace(str, old, new)  # Replace text
Contains(str, sub)      # Check substring
HasPrefix(str, prefix)  # Check prefix
HasSuffix(str, suffix)  # Check suffix
```

#### Environment Variables

```go
EnvVar(name)            # Get environment variable
Getenv(name, fallback)  # Env var with fallback
ExportEnv(name, value)  # Set environment variable
UnsetEnv(name)          # Unset environment variable
ListEnv()               # List all environment variables
```

#### Path Operations

```go
PathExists(path)        # Check if path exists
FileExists(path)        # Check if file exists
IsFile(path)            # Check if path is file
IsDir(path)             # Check if path is directory
Readable(path)          # Check if readable
Writable(path)          # Check if writable
Executable(path)        # Check if executable
FileSize(path)          # Get file size
FileModTime(path)       # Get modification time
```

## Error Handling

### Shellapi Error Handling

All shellapi functions return `(string, error)` tuples:

```go
func safeOperation() string {
    result, err := shellapi.GitStatus()
    if err != nil {
        return shellapi.ErrorMsg("Git Error", err.Error())
    }
    return shellapi.SuccessMsg("Git Status", result)
}

func directoryOperation() string {
    result, err := shellapi.RunShell("cd", "/some/path")
    if err != nil {
        return shellapi.ErrorMsg("Directory Error", err.Error())
    }
    return shellapi.SuccessMsg("Directory", "Changed successfully")
}
```

### Go Code Error Handling

In the REPL, errors show line numbers:

```bash
gosh> undefined_function
Error: undefined: undefined_function on line 1

gosh> fmt.Println(undefined_var)
Error: undefined: undefined_var on line 1
```

## Signal Handling

### Ctrl+C Behavior

- **Interrupt Go code**: Stops current Go execution
- **Interrupt shell commands**: Properly terminates subprocesses
- **Preserves session**: Continues interactive shell

### Signal Propagation

```bash
gosh> # Long-running command
gosh> for i := 0; i < 100; i++ {
...     fmt.Println(i)
...     time.Sleep(1 * time.Second)
... }

# Press Ctrl+C to interrupt
^C
gosh> # Shell continues running
```

## Performance Considerations

### Command Execution

- Commands use Go's `os/exec` for real execution
- Output is fully captured as strings
- Large outputs are truncated at 40,000 characters

### Memory Usage

- Go state persists across shell session
- Variables and functions remain defined
- Use `clear()` function to reset state (if available)

### Startup Time

- Instant startup - no REPL initialization waiting
- Config file loading is asynchronous
- Shell session ready immediately

### Best Practices

1. **Use shellapi functions**: Pre-built optimized functions
2. **Cache expensive operations**: Store results of costly commands
3. **Handle errors**: Always check error returns
4. **Use color functions**: Better readability
5. **Group related functions**: Consistent naming patterns

---

Need more examples? Check out our [Configuration Examples](config.md) for real-world configurations!
