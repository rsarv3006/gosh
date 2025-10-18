# Getting Started

Welcome to gosh! This guide will help you get comfortable with the hybrid shell experience that combines Go programming with traditional shell commands.

## Your First Session

Start gosh by simply running:

```bash
gosh
```

You'll see a welcome message and the gosh prompt:

```
gosh 0.2.3 - Go shell with yaegi (BUILT: 2023-10-17 15:30:45)
Type 'exit' to quit, try some Go code or shell commands!

gosh>
```

## Basic Usage

### Traditional Shell Commands

All your favorite shell commands work exactly as you'd expect:

```bash
gosh> ls
README.md    main.go    go.mod    go.sum

gosh> pwd
/Users/username/projects/gosh

gosh> git status
On branch main
nothing to commit, working tree clean
```

### Go Code in the Shell

Now try writing some Go code:

```bash
gosh> name := "gosh"
gosh> fmt.Printf("Hello %s!\n", name)
Hello gosh!

gosh> x := 42
gosh> y := x * 2
gosh> fmt.Printf("x=%d, y=%d\n", x, y)
x=42, y=84
```

### Functions and Control Flow

Write functions and use control structures with proper multiline support:

```bash
gosh> func greet(name string) {
...     fmt.Printf("Hello, %s! Welcome to gosh!\n", name)
... }
gosh> greet("developer")
Hello, developer! Welcome to gosh!

gosh> for i := 0; i < 3; i++ {
...     fmt.Printf("Count: %d\n", i)
... }
Count: 0
Count: 1
Count: 2
```

### Command Substitution

The game-changing feature: capture command output into Go variables:

```bash
gosh> files := $(ls)
gosh> fmt.Printf("Found %d files\n", len(strings.Split(files, "\n")))
Found 6 files

gosh> currentDir := $(pwd)
gosh> fmt.Printf("Current directory: %s", currentDir)
Current directory: /Users/username/projects/gosh
```

## Pre-imported Packages

gosh automatically imports common packages, so you can use them without explicit imports:

```bash
gosh> files, _ := filepath.Glob("*.go")
gosh> fmt.Println(files)
[main.go repl.go router.go ...]

gosh> currentDir, _ := os.Getwd()
gosh> fmt.Println(currentDir)
/Users/username/projects/gosh

gosh> timestamp := time.Now().Format("2006-01-02")
gosh> fmt.Println(timestamp)
2023-10-17
```

## Built-in Commands

gosh provides essential built-in commands:

```bash
gosh> help
Available built-ins:
- cd <path>          Change directory
- pwd                Print working directory
- exit               Exit gosh
- init               Create example config
- help               Show this help

gosh> cd /tmp
gosh> pwd
/tmp

gosh> exit
```

## Navigation and History

Use arrow keys for history navigation:

```bash
gosh> # Press up arrow to see previous commands
gosh> ls -la          # Previously typed
gosh> name := "test" # Previously typed
```

## Error Handling

Errors show line numbers and helpful messages:

```bash
gosh> fmt.Println(undefined_var)
Error: undefined: undefined_var on line 1
```

## Combining Go and Shell

The real power comes from mixing both approaches:

```bash
# Get git status and process it in Go
gosh> gitStatus := $(git status --porcelain)
gosh> if gitStatus == "" {
...     fmt.Println("✅ Working directory is clean")
... } else {
...     fmt.Println("⚠️ Working directory has changes")
... }

gosh> lines := strings.Split(gitStatus, "\n")
gosh> fmt.Printf("Found %d changed files\n", len(lines)-1)
```

## Next Steps

Now that you understand the basics:

1. **[Installation Guide](install.md)** - Install gosh on your system
2. **[Configuration](config.md)** - Set up custom functions and shortcuts
3. **[User Guide](guide.md)** - Advanced features and workflows
4. **[CLI Reference](reference.md)** - Complete command and API reference

## Pro Tips

### 1. Use Variables for Common Commands

```bash
gosh> lsCmd := "ls --color=auto"
gosh> result := $(lsCmd)
gosh> fmt.Println(result)
```

### 2. Create Quick Helpers

```bash
gosh> func quickGit() {
...     status, _ := shellapi.GitStatus()
...     fmt.Println(status)
... }
gosh> quickGit()
```

### 3. Chain Operations

```bash
gosh> files := $(ls *.go)
gosh> for _, file := range strings.Split(files, "\n") {
...     if file != "" {
...         fmt.Printf("Processing: %s\n", file)
...     }
... }
```

---

Ready to dive deeper? Check out our comprehensive [User Guide](guide.md) for advanced workflows and customization options!
