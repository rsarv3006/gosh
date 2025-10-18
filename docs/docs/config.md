# Configuration Examples

Discover practical configuration examples for common workflows, development environments, and productivity enhancers.

## Basic Setup

### Directory Navigation Patterns

When creating navigation functions, there are two approaches depending on your needs:

#### Simple Print Functions (for information only)

```go
func showCurrentDir() string {
	if pwd, err := shellapi.Pwd(); err == nil {
		return "Current directory: " + pwd
	}
	return "Error getting current directory"
}

func showProject() string {
	return fmt.Sprintf("Would navigate to: %s", "~/projects")
}
```

#### Directory-Changing Functions (single directory change per function)

```go
func goConfig() string {
	// This actually changes the shell directory
	result, err := shellapi.RunShell("cd", "/Users/rjs/.config/gosh/")
	if err != nil {
		return "CD ERROR: " + err.Error()
	}
	return result
}

func goGosh() string {
	// This actually changes the shell directory
	result, err := shellapi.RunShell("cd", "/Users/rjs/dev/gosh")
	if err != nil {
		return "CD ERROR: " + err.Error()
	}
	return result
}
```

#### Sequential Directory Operations

```go
func navigateAndWork() string {
	msg := " Navigation sequence:\n"
	
	// Go to config directory
	if result, err := shellapi.RunShell("cd", "/Users/rjs/.config/gosh/"); err == nil {
		msg += "✓ Moved to config directory\n"
		// Do work here if needed
		msg += "✓ Config directory operations complete\n"
	} else {
		return "✗ Failed to navigate to config: " + err.Error()
	}
	
	// Now go to gosh project
	if result, err := shellapi.RunShell("cd", "/Users/rjs/dev/gosh"); err == nil {
		msg += "✓ Moved to gosh project directory\n"
	} else {
		return "✗ Failed to navigate to gosh: " + err.Error()
	}
	
	return msg + "✓ Navigation sequence complete"
}

func showDirectoryInfo() {
	msg := "Working directory information:"
	if pwd, err := shellapi.Pwd(); err == nil {
		msg += fmt.Sprintf("\nCurrent: %s", pwd)
	}
	
	if files, err := shellapi.Ls(); err == nil {
		fileCount := len(strings.Split(strings.TrimSpace(files), "\n"))
		msg += fmt.Sprintf("\nFiles: %d", fileCount)
	}
	
	fmt.Println(msg)
}
```

### Minimal Configuration

Create `~/.config/gosh/config.go`:

```go
package main

import (
	"fmt"
	"os"
)

func init() {
	fmt.Println("🚀 gosh ready!")
}

// Basic info function
func info() {
	fmt.Printf("gosh v%s\n", "main".GetVersion())
	fmt.Printf("User: %s\n", os.Getenv("USER"))
	fmt.Printf("Home: %s\n", os.Getenv("HOME"))
}
```

## Development Workflows

### Go Developer Configuration

```go
package main

import (
	"fmt"
	"os"
	"github.com/rsarv3006/gosh_lib/shellapi"
)

func init() {
	fmt.Println("🔧 Go developer environment loaded!")
}

// Project navigation
func goProject() string {
	return cdPath("~/go/src/github.com/yourname")
}

func goSite() string {
	return cdPath("~/go/src/github.com/yourname/website")
}

// Development workflow
func buildAndTest() string {
	msg := "Building and testing...\n"
	
	// Build
	if build, err := shellapi.GoBuild(); err != nil {
		return shellapi.ErrorMsg("Build Failed", err.Error())
	} else {
		msg += build
	}
	
	// Test
	if test, err := shellapi.GoTest(); err != nil {
		return shellapi.ErrorMsg("Tests Failed", err.Error())
	} else {
		msg += test
	}
	
	return shellapi.SuccessMsg("Complete", "Build and test successful")
}

// Coverage and linting
func check() string {
	msg := "Running checks...\n"
	
	// Test coverage
	coverage, err := shellapi.RunShell("go", "test", "-cover")
	if err != nil {
		msg += shellapi.ErrorMsg("Coverage", err.Error())
	} else {
		msg += shellapi.SuccessMsg("Coverage", "Tests passed")
	}
	
	// Vet
	vet, err := shellapi.GoVet()
	if err != nil {
		msg += shellapi.ErrorMsg("Vet", err.Error())
	} else {
		msg += shellapi.SuccessMsg("Vet", "Code passes analysis")
	}
	
	return msg
}

// Quick git operations
func quickCommit(msg string) string {
	// Check status
	status, _ := shellapi.GitStatus()
	if status == "" {
		return "Nothing to commit"
	}
	
	// Add all
	shellapi.GitAdd(".")
	
	// Commit
	if _, err := shellapi.GitCommit(msg); err != nil {
		return shellapi.Error("Commit failed: " + err.Error())
	}
	
	return shellapi.Success("✓ Committed: " + msg)
}

// Helper function for directory changes
func cdPath(path string) string {
	result, err := shellapi.RunShell("cd", path)
	if err != nil {
		return shellapi.ErrorMsg("CD Error", err.Error())
	}
	return result
}
```

### Web Developer Configuration

```go
package main

import (
	"fmt"
	"github.com/rsarv3006/gosh_lib/shellapi"
)

func init() {
	fmt.Println("🌐 Web developer environment loaded!")
}

// Project workspaces
func goReact() string {
	return cdPath("~/projects/react-app")
}

func goVue() string {
	return cdPath("~/projects/vue-app")
}

func goNext() string {
	return cdPath("~/projects/nextjs-app")
}

// Package management
func devSetup() string {
	msg := "Setting up development environment...\n"
	
	// npm install
	if install, err := shellapi.NpmInstall(); err != nil {
		msg += shellapi.ErrorMsg("npm install", err.Error())
	} else {
		msg += shellapi.SuccessMsg("Dependencies", "npm install completed")
	}
	
	return msg
}

// Development servers
func startDev() string {
	return shellapi.NpmRun("dev")
}

func startBuild() string {
	return shellapi.NpmRun("build")
}

func startTest() string {
	return shellapi.NpmRun("test")
}

// Docker workflow
func dockerDev() string {
	msg := "Starting development containers...\n"
	
	// Build and run dev containers
	if _, err := shellapi.RunShell("docker-compose", "up", "-d", "--build"); err != nil {
		return shellapi.ErrorMsg("Docker", err.Error())
	}
	
	return shellapi.SuccessMsg("Docker", "Development containers started")
}

func dockerStop() string {
	if _, err := shellapi.RunShell("docker-compose", "down"); err != nil {
		return shellapi.ErrorMsg("Docker", err.Error())
	}
	return shellapi.Success("🛑 Docker containers stopped")
}
```

### DevOps/Cloud Engineer Configuration

```go
package main

import (
	"fmt"
	"github.com/rsarv3006/gosh_lib/shellapi"
)

func init() {
	fmt.Println("☁️ DevOps environment loaded!")
}

// Kubernetes operations
func k8sPods() string {
	pods, err := shellapi.KubectlPods()
	if err != nil {
		return shellapi.ErrorMsg("K8s", err.Error())
	}
	return shellapi.InfoMsg("Pods", pods)
}

func k8sLogs(app string) string {
	logs, err := shellapi.KubectlPods()
	if err != nil {
		return shellapi.ErrorMsg("K8s", err.Error())
	}
	
	// Get first pod in deployment
	podName := firstPod(app)
	if podName == "" {
		return shellapi.WarningMsg("Pod", "No pod found for " + app)
	}
	
	podLogs, err := shellapi.KubectlLogs(podName)
	if err != nil {
		return shellapi.ErrorMsg("Logs", err.Error())
	}
	
	return shellapi.InfoMsg(app+" Logs", podLogs)
}

// Terraform workflows
func tfInit() string {
	result, err := shellapi.RunShell("terraform", "init")
	if err != nil {
		return shellapi.ErrorMsg("Terraform", err.Error())
	}
	return shellapi.SuccessMsg("Terraform", "Initialized")
}

func tfPlan() string {
	result, err := shellapi.RunShell("terraform", "plan")
	if err != nil {
		return shellapi.ErrorMsg("Terraform", err.Error())
	}
	return shellapi.InfoMsg("Terraform Plan", result)
}

func tfApply() string {
	fmt.Println(shellapi.Warning("⚠️ This will apply changes to infrastructure"))
	// In real implementation, add confirmation prompt
	
	result, err := shellapi.RunShell("terraform", "apply", "-auto-approve")
	if err != nil {
		return shellapi.ErrorMsg("Terraform", err.Error())
	}
	return shellapi.SuccessMsg("Terraform", "Applied changes")
}

// AWS CLI helpers
func awsRegions() string {
	regions, err := shellapi.RunShell("aws", "ec2", "describe-regions", "--output", "table")
	if err != nil {
		return shellapi.ErrorMsg("AWS", err.Error())
	}
	return shellapi.InfoMsg("AWS Regions", regions)
}

func awsS3List() string {
	buckets, err := shellapi.RunShell("aws", "s3", "ls")
	if err != nil {
		return shellapi.ErrorMsg("AWS S3", err.Error())
	}
	return shellapi.InfoMsg("S3 Buckets", buckets)
}

// Helper functions
func firstPod(app string) string {
	pods, err := shellapi.RunShell("kubectl", "get", "pods", "-l", "app="+app, "-o", "jsonpath='{.items[0].metadata.name}'")
	if err != nil {
		return ""
	}
	// Remove quotes and return
	return strings.ReplaceAll(pods, "'", "")
}
```

## Project Management

### Multi-Project Workspace

```go
package main

import (
	"fmt"
	"github.com/rsarv3006/gosh_lib/shellapi"
)

func init() {
	fmt.Println("📁 Multi-project workspace loaded!")
}

// Project categories
func goClient() string {
	return cdPath("~/projects/client-portal")
}

func goApi() string {
	return cdPath("~/projects/api-server")
}

func goAdmin() string {
	return cdPath("~/projects/admin-dashboard")
}

// Environment switching
func envDev() string {
	os.Setenv("NODE_ENV", "development")
	os.Setenv("API_URL", "http://localhost:3000")
	return shellapi.SuccessMsg("Environment", "Switched to development")
}

func envStaging() string {
	os.Setenv("NODE_ENV", "staging")
	os.Setenv("API_URL", "https://staging-api.example.com")
	return shellapi.SuccessMsg("Environment", "Switched to staging")
}

func envProd() string {
	os.Setenv("NODE_ENV", "production")
	os.Setenv("API_URL", "https://api.example.com")
	return shellapi.SuccessMsg("Environment", "Switched to production")
}

// Project health check
func projectHealth() string {
	msg := "Project Health Report:\n"
	
	// Git status
	if gitStatus, err := shellapi.GitStatus(); err == nil {
		if gitStatus == "" {
			msg += shellapi.Success("✓ Git: Working tree clean\n")
		} else {
			msg += shellapi.Warning("⚠ Git: Uncommitted changes\n")
		}
	}
	
	// Dependencies
	if _, err := shellapi.NpmRun("test:dependencies"); err != nil {
		msg += shellapi.Error("✗ Dependencies: Issues detected\n")
	} else {
		msg += shellapi.Success("✓ Dependencies: OK\n")
	}
	
	// Build status
	if _, err := shellapi.NpmRun("build:test"); err != nil {
		msg += shellapi.Error("✗ Build: Failed\n")
	} else {
		msg += shellapi.Success("✓ Build: OK\n")
	}
	
	return msg
}

// Database operations
func dbConnect() string {
	// Load environment
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		return shellapi.ErrorMsg("Database", "DATABASE_URL not set")
	}
	
	// Test connection
	if _, err := shellapi.RunShell("pg_isready", "-d", dbUrl); err != nil {
		return shellapi.ErrorMsg("Database", "Connection failed")
	}
	
	return shellapi.SuccessMsg("Database", "Connected")
}

func dbMigrate() string {
	msg := "Running database migrations...\n"
	
	if migrate, err := shellapi.RunShell("npm", "run", "db:migrate"); err != nil {
		return shellapi.ErrorMsg("Migration", err.Error())
	}
	
	return shellapi.SuccessMsg("Migration", "Completed") + "\n" + migrate
}

func dbSeed() string {
	msg := "Seeding database...\n"
	
	if seed, err := shellapi.RunShell("npm", "run", "db:seed"); err != nil {
		return shellapi.ErrorMsg("Seeding", err.Error())
	}
	
	return shellapi.SuccessMsg("Seeding", "Completed") + "\n" + seed
}
```

## Productivity Enhancers

### System Information and Monitoring

```go
package main

import (
	"fmt"
	"github.com/rsarv3006/gosh_lib/shellapi"
)

func init() {
	fmt.Println("📊 System monitoring loaded!")
}

// System info
func sysInfo() string {
	info := shellapi.Bold("System Information:\n")
	
	if uptime, err := shellapi.Uptime(); err == nil {
		info += shellapi.Bold("Uptime: ") + uptime + "\n"
	}
	
	if user, err := shellapi.Whoami(); err == nil {
		info += shellapi.Bold("User: ") + user + "\n"
	}
	
	if hostname, err := shellapi.Hostname(); err == nil {
		info += shellapi.Bold("Host: ") + hostname + "\n"
	}
	
	if arch, err := shellapi.Arch(); err == nil {
		info += shellapi.Bold("Arch: ") + arch + "\n"
	}
	
	return info
}

// Memory and disk usage
func sysUsage() string {
	msg := shellapi.Bold("System Usage:\n")
	
	if mem, err := shellapi.Free(); err == nil {
		msg += shellapi.Bold("Memory:\n") + mem + "\n"
	}
	
	if disk, err := shellapi.Df(); err == nil {
		msg += shellapi.Bold("Disk:\n") + disk + "\n"
	}
	
	return msg
}

// Process monitoring
func psGrep(pattern string) string {
	if pattern == "" {
		return "Usage: psGrep <pattern>"
	}
	
	result, err := shellapi.RunShell("ps", "aux", "|", "grep", pattern)
	if err != nil {
		return shellapi.ErrorMsg("Process search", err.Error())
	}
	
	return shellapi.InfoMsg("Processes matching "+pattern, result)
}

func portsUsed() string {
	result, err := shellapi.RunShell("lsof", "-i", "-n", "-P")
	if err != nil {
		return shellapi.ErrorMsg("Ports", err.Error())
	}
	
	return shellapi.InfoMsg("Used Ports", result)
}
```

### File Management Tools

```go
package main

import (
	"fmt"
	"github.com/rsarv3006/gosh_lib/shellapi"
)

func init() {
	fmt.Println("📂 File management tools loaded!")
}

// Quick file operations
func findLarge() string {
	result, err := shellapi.RunShell("find", ".", "-type", "f", "-size", "+10M", "-exec", "ls", "-lh", "{}", ";")
	if err != nil {
		return shellapi.ErrorMsg("Find", err.Error())
	}
	return shellapi.InfoMsg("Large files (>10MB)", result)
}

func findRecent(days int) string {
	if days <= 0 {
		days = 7 // default to 7 days
	}
	
	result, err := shellapi.RunShell("find", ".", "-type", "f", "-mtime", fmt.Sprintf("-%d", days), "-exec", "ls", "-lt", "{}", ";")
	if err != nil {
		return shellapi.ErrorMsg("Find", err.Error())
	}
	
	return shellapi.InfoMsg(fmt.Sprintf("Recent files (%d days)", days), result)
}

func cleanTemp() string {
	msg := "Cleaning temporary files...\n"
	
	// Remove common temp patterns
	patterns := []string{"*~", "*.tmp", "*.swp", "*.bak"}
	
	for _, pattern := range patterns {
		if result, err := shellapi.RunShell("find", ".", "-name", pattern, "-delete"); err == nil {
			if result == "" {
				msg += shellapi.Success("✓ Cleaned " + pattern + " files\n")
			}
		}
	}
	
	return shellapi.SuccessMsg("Cleanup", "Completed")
}

// Project structure tools
func treeFiles() string {
	result, err := shellapi.Tree()
	if err != nil {
		return shellapi.ErrorMsg("Tree", err.Error())
	}
	return result
}

func countByType(fileType string) string {
	if fileType == "" {
		return "Usage: countByType <extension>"
	}
	
	if !strings.HasPrefix(fileType, ".") {
		fileType = "." + fileType
	}
	
	result, err := shellapi.RunShell("find", ".", "-name", "*"+fileType, "|", "wc", "-l")
	if err != nil {
		return shellapi.ErrorMsg("Count", err.Error())
	}
	
	count := strings.TrimSpace(result)
	return shellapi.InfoMsg(fileType+" files", count)
}
```

## Environment-Specific Configs

### Container Development

```go
package main

import (
	"github.com/rsarv3006/gosh_lib/shellapi"
)

func init() {
	fmt.Println("🐳 Container development loaded!")
}

// Docker utilities
func dockerClean() string {
	msg := "Cleaning Docker resources...\n"
	
	// Remove stopped containers
	if _, err := shellapi.RunShell("docker", "container", "prune", "-f"); err == nil {
		msg += shellapi.Success("✓ Removed stopped containers\n")
	}
	
	// Remove unused images
	if _, err := shellapi.RunShell("docker", "image", "prune", "-f"); err == nil {
		msg += shellapi.Success("✓ Removed unused images\n")
	}
	
	// Remove unused volumes
	if _, err := shellapi.RunShell("docker", "volume", "prune", "-f"); err == nil {
		msg += shellapi.Success("✓ Removed unused volumes\n")
	}
	
	return shellapi.SuccessMsg("Docker", "Cleanup completed")
}

func dockerLogs(container string) string {
	if logs, err := shellapi.DockerLogs(container); err == nil {
		return shellapi.InfoMsg(container+" logs", logs)
	} else {
		return shellapi.ErrorMsg("Docker", err.Error())
	}
}

func dockerStats() string {
	result, err := shellapi.RunShell("docker", "stats", "--no-stream")
	if err != nil {
		return shellapi.ErrorMsg("Docker", err.Error())
	}
	return shellapi.InfoMsg("Container Stats", result)
}
```

## Usage Examples

After creating your `config.go`, access functions like this:

```bash
# Development workflow
gosh> buildAndTest()    # Execute Go build and test
gosh> quickCommit("fix bug")  # Quick commit with message
gosh> projectHealth()   # Check overall project status

# Environment switching
gosh> envStaging()      # Switch to staging environment
gosh> goApi()          # Navigate to API project
gosh> dbMigrate()      # Run database migrations

# System utilities
gosh> sysInfo()        # Show system information
gosh> cleanTemp()      # Clean temporary files
gosh> findRecent(7)    # Find files changed in last 7 days

# Docker operations
gosh> dockerClean()    # Clean Docker resources
gosh> dockerLogs("web") # Show logs for web container
```

## Best Practices

1. **Start simple**: Begin with basic functions and expand as needed
2. **Group related functions**: Use prefixes like `go*`, `db*`, `docker*`
3. **Add error handling**: Always check for errors and provide feedback
4. **Use color functions**: Make output more readable with `Success()`, `Error()`, etc.
5. **Document your functions**: Add comments explaining what each function does
6. **Test your configs**: Reload gosh and test functions before relying on them

## Troubleshooting Common Issues

### sequential Directory Operations

**Problem**: Multiple CD operations in one function didn't work in versions prior to v0.2.4.

**Solution**: Fixed in v0.2.4 with immediate directory changes and proper shell state synchronization.

```go
// ✅ This now works properly (v0.2.4+)
func workingExample() {
    goConfig()     // Changes to config directory
    fmt.Println("HELLO")  // Prints while in config
    goGosh()       // Changes to gosh directory
    goConfig()     // Changes back to config  
    fmt.Println("DONE")   // Prints while back in config
}
```

**How it works**: Directory changes happen immediately and persist throughout the function execution.

**Examples**:

```go
// ✅ Solution 1: Single purpose functions
func thing() string {
    result := navigateSequentially()  // One function that handles all navigation
    fmt.Println("HELLO")
    return result
}

func navigateSequentially() string {
    msg := "🔄 Starting navigation...\n"
    
    if _, err := shellapi.RunShell("cd", "/Users/rjs/.config/gosh/"); err == nil {
        msg += "✓ Visited config directory\n"
        // Print any config-specific info here
        
        if _, err := shellapi.RunShell("cd", "/Users/rjs/dev/gosh"); err == nil {
            msg += "✓ Returned to gosh directory\n"
        } else {
            return "✗ Failed to return to gosh"
        }
    } else {
        return "✗ Failed to visit config"
    }
    
    return msg + "✅ Navigation complete"
}

// ✅ Solution 2: Work outside functions
gosh> goConfig()    # Go to config
gosh> fmt.Println("Working in config directory")
gosh> goGosh()      # Go to gosh
gosh> fmt.Println("Working in gosh directory")

// ✅ Solution 3: Create utility functions that return information
func getConfigPath() string {
    return "/Users/rjs/.config/gosh"
}

func getGoshPath() string {
    return "/Users/rjs/dev/gosh"
}

func exploreConfig() string {
    msg := "📁 Exploring config directory...\n"
    
    if _, err := shellapi.RunShell("cd", getConfigPath()); err == nil {
        if files, err := shellapi.Ls(); err == nil {
            msg += "Files in config:\n" + files
        }
    } else {
        return "❌ Cannot access config directory"
    }
    
    return msg
}
```

### Sequential Directory Operations

**How it works (v0.2.4+)**: Functions execute in sequence with directory changes taking effect immediately, allowing sequential operations in different directories with proper prompt updates.

**Key implementation**: Directory changes update both the OS working directory and the shell state in real-time, ensuring that subsequent operations and the prompt reflect the current location.

```go
// ❌ Multiple directory changes don't stack
func example1() {
    result1, _ := shellapi.RunShell("cd", "/tmp")     # Changes at end
    result2, _ := shellapi.RunShell("cd", "/home")   # Overwrites above
    result3, _ := shellapi.RunShell("cd", "/etc")     # Final directory
    # After this function, you'll be in /etc
}

// ✅ Each function does one directory change
func goTmp() { shellapi.RunShell("cd", "/tmp") }
func goHome() { shellapi.RunShell("cd", "/home") }
func goEtc() { shellapi.RunShell("cd", "/etc") }

# Usage:
gosh> goTmp()     # Now in /tmp
gosh> goHome()    # Now in /home  
gosh> goEtc()     # Now in /etc
```

---

Ready for the complete reference? Check out our [CLI Reference](reference.md) for all available commands and shellapi functions!
