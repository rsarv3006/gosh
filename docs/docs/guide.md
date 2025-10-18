# User Guide

Master the advanced features of gosh with this comprehensive user guide covering configuration, shellapi functions, and productivity workflows.

## Configuration Strategy

gosh uses a **dual-layer configuration approach** that gives you the best of both worlds: standard shell compatibility plus Go-powered extensions.

### Layer 1: Standard Shell Environment (`env.go`)

**Automatic Standard Config Loading:**

gosh loads regular shell configs when run as login shell, supporting:
- `.bash_profile`, `.zprofile`, `.profile`, `.bash_login`, `.login`
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

### Layer 2: Go-Powered Extensions (`config.go`)

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

## Working with Shellapi Functions

gosh v0.2.2 features **working shellapi functions** that execute real commands via Go's `os/exec`, providing actual command output and persistent directory changes.

### Development Workflow Functions

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

// Functions that actually execute real commands
func build() string {
	result, err := shellapi.GoBuild()
	if err != nil {
		return "BUILD ERROR: " + err.Error()
	}
	return "BUILD SUCCESS: " + result
}

func test() string {
	result, _ := shellapi.GoTest()
	return result
}

func run() string {
	result, _ := shellapi.GoRun()
	return result
}

func gs() string {
	result, err := shellapi.GitStatus()
	if err != nil {
		return "GIT ERROR: " + err.Error()
	}
	return "GIT STATUS:\n" + result
}
```

### Directory Navigation Functions

Directory changes actually persist in the shell session:

```go
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
	return result
}

func goProjects() string {
	result, err := shellapi.RunShell("cd", "~/projects")
	if err != nil {
		return "ERROR: " + err.Error()
	}
	return result
}
```

### Usage Examples

```bash
# These actually execute real commands!
gosh> build()    # Executes real go build command
gosh> test()     # Executes real go test command  
gosh> run()      # Executes real go run . command
gosh> gs()       # Executes real git status with full output

# These actually change directories!
gosh> goGosh()   # Changes to ~/dev/gosh - directory persists!
gosh> goConfig() # Changes to ~/.config/gosh/ - directory persists!
gosh> pwd
/Users/rjs/dev/gosh  # Directory actually changed
```

## Advanced Command Usage

### RunShell Command Engine

The `shellapi.RunShell` function is the core command execution engine:

```go
// Basic command execution
result, err := shellapi.RunShell("command", "arg1", "arg2")

// Development tools
build := shellapi.RunShell("go", "build")
tests := shellapi.RunShell("go", "test", "./...")

// System commands
uptime := shellapi.RunShell("uptime")
ps := shellapi.RunShell("ps", "aux")

// Directory changes with persistence
result, err := shellapi.RunShell("cd", "/path/to/project")
if err != nil {
    return "CD ERROR: " + err.Error()  
}
return result  // Directory actually changes
```

### Error Handling Patterns

```go
func buildAndTest() string {
    // Build first
    build, buildErr := shellapi.GoBuild()
    if buildErr != nil {
        return "BUILD FAILED: " + buildErr.Error()
    }
    
    // Then test
    test, testErr := shellapi.GoTest()
    if testErr != nil {
        return "TESTS FAILED: " + testErr.Error()
    }
    
    return "✅ Build successful\n" + test
}

func deployTo(prodEnv string) string {
    msg := "Deploying to " + prodEnv + "...\n"
    
    // Check git status
    status, _ := shellapi.GitStatus()
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

## Project-Specific Workflow

### Multi-Environment Development

```go
// Environment-based project management
func goEnv(env string) string {
    var path string
    switch env {
    case "dev":
        path = "~/projects/myapp-dev"
    case "staging":
        path = "~/projects/myapp-staging"
    case "prod":
        path = "~/projects/myapp-prod"
    default:
        return "Unknown environment: " + env
    }
    
    result, err := shellapi.RunShell("cd", path)
    if err != nil {
        return "ERROR: " + err.Error()
    }
    return result
}

// Usage:
// gosh> goEnv("dev")  # Goes to dev environment
// gosh> goEnv("prod") # Goes to prod environment
```

### Database Operations

```go
func dbConnect(env string) string {
    var dbUrl string
    switch env {
    case "dev":
        dbUrl = "postgresql://user:pass@localhost/devdb"
    case "prod":
        dbUrl = "postgresql://user:pass@prod-server/proddb"
    }
    
    result, err := shellapi.RunShell("psql", dbUrl)
    return "Connected to " + env + " database\n" + result
}

func dbMigrate(env string) string {
    msg := "Migrating " + env + " database...\n"
    
    // Run migrations
    migrate, err := shellapi.RunShell("npm", "run", "db:migrate")
    if err != nil {
        return "MIGRATION ERROR: " + err.Error()
    }
    
    return msg + migrate
}
```

## Productivity Tips

### Git Workflow Automation

```go
func quickCommit(msg string) string {
    status, _ := shellapi.GitStatus()
    if status == "" {
        return "Nothing to commit"
    }
    
    // Add all changes
    shellapi.GitAdd(".")
    
    // Create commit
    commit, err := shellapi.GitCommit(msg)
    if err != nil {
        return "COMMIT ERROR: " + err.Error()
    }
    
    return "✅ Committed: " + msg
}

func syncBranch() string {
    msg := "Syncing branch with remote...\n"
    
    // Pull latest
    pull, err := shellapi.GitPull()
    if err != nil {
        return "PULL ERROR: " + err.Error()
    }
    
    // Push local changes
    push, err := shellapi.GitPush()
    if err != nil {
        return "PUSH ERROR: " + err.Error()
    }
    
    return msg + "✅ Branch synced"
}
```

### File Management

```go
func projectInfo() string {
    info := shellapi.Bold("Project Information:\n")
    
    // Git status
    if gitStatus, err := shellapi.GitStatus(); err == nil {
        info += shellapi.Bold("Git: ") + gitStatus + "\n"
    }
    
    # File count
    if files, _ := shellapi.Ls(); files != "" {
        fileCount := len(strings.Split(strings.TrimSpace(files), "\n"))
        info += shellapi.Bold("Files: ") + fmt.Sprintf("%d files\n", fileCount)
    }
    
    # Directory size
    if du, _ := shellapi.RunShell("du", "-sh", "."); du != "" {
        parts := strings.Split(strings.TrimSpace(du), "\t")
        if len(parts) > 0 {
            info += shellapi.Bold("Size: ") + parts[0] + "\n"
        }
    }
    
    return info
}

func cleanProject() string {
    msg := "Cleaning project...\n"
    
    # Remove temporary files
    if _, err := shellapi.RunShell("find", ".", "-name", "*.tmp", "-delete"); err == nil {
        msg += shellapi.Success("✓ Removed temporary files\n")
    }
    
    # Remove build artifacts
    if _, err := shellapi.RunShell("rm", "-rf", "dist/", "build/"); err == nil {
        msg += shellapi.Success("✓ Removed build artifacts\n")
    }
    
    return msg
}
```

## Color and Formatting Functions

Use the color functions to make your output more readable:

```go
func projectStatus() string {
    status, _ := shellapi.GitStatus()
    
    if status == "" {
        return shellapi.Success("✓ Working tree is clean")
    }
    
    return shellapi.Warning("⚠ Working tree has changes:\n") + 
           shellapi.Highlight(strings.TrimSpace(status))
}

func buildReport() string {
    result, err := shellapi.GoBuild()
    
    if err != nil {
        return shellapi.Error("✗ Build failed:") + "\n" + err.Error()
    }
    
    return shellapi.Success("✓ Build successful") + "\n" + result
}

func infoMessage() string {
    return shellapi.InfoMsg("Info", "This is an informational message")
}

func debugMessage() string {
    return shellapi.DebugMsg("Debug", "Detailed diagnostics information")
}
```

## Advanced Patterns

### Lazy Loading Commands

```go
var (
    projectCache map[string]string
    cacheLoaded  bool
)

func getProjectInfo(project string) string {
    if !cacheLoaded {
        loadProjectCache()
        cacheLoaded = true
    }
    
    if info, exists := projectCache[project]; exists {
        return info
    }
    
    return "Project not found: " + project
}

func loadProjectCache() {
    projectCache = make(map[string]string)
    // Load project information lazily
    // ... implementation
}
```

### Interactive Workflows

```go
func interactiveDeploy() string {
    // Check if we're in a git repo
    if _, err := shellapi.GitStatus(); err != nil {
        return "Not in a git repository"
    }
    
    fmt.Println(shellapi.Question("Deploy to which environment? (dev/staging/prod)"))
    // In real implementation, you'd read user input here
    
    // Example: deploy to staging
    return deployTo("staging")
}
```

## Best Practices

### Performance Tips

1. **Cache expensive operations**: Store results of expensive commands
2. **Use shellapi functions**: Pre-built functions are optimized
3. **Validate inputs**: Check parameters before executing commands
4. **Handle errors**: Always check for errors and provide feedback

### Error Handling

- Always check `err != nil` before using results
- Return meaningful error messages
- Provide context about what went wrong
- Use color functions to highlight errors

### Function Naming

- Use descriptive names: `goGosh()`, `buildAndTest()`, `projectInfo()`
- Follow Go naming conventions
- Group related functions with prefixes: `go*` for navigation, `db*` for database

---

Ready to explore specific configurations? Check out our [Configuration Examples](config.md) or the complete [CLI Reference](reference.md)!
