//go:build darwin || linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type BuiltinHandler struct {
	state *ShellState
}

func NewBuiltinHandler(state *ShellState) *BuiltinHandler {
	return &BuiltinHandler{state: state}
}

func (b *BuiltinHandler) IsBuiltin(command string) bool {
	switch command {
	case "cd", "exit", "help", "init", "session":
		return true
	default:
		return false
	}
}

func (b *BuiltinHandler) Execute(command string, args []string) ExecutionResult {
	switch command {
	case "cd":
		return b.cd(args)
	case "exit":
		return b.exit(args)
	case "help":
		return b.help(args)
	case "init":
		return b.initConfig(args)
	case "session":
		return b.session(args)
	default:
		return ExecutionResult{
			Output:   fmt.Sprintf("Unknown builtin: %s", command),
			ExitCode: 1,
			Error:    fmt.Errorf("unknown builtin: %s", command),
		}
	}
}

func (b *BuiltinHandler) cd(args []string) ExecutionResult {
	target := b.state.Environment["HOME"]

	if len(args) > 0 {
		target = args[0]
	}

	// Expand path
	expanded := b.state.ExpandPath(target)

	// Check if directory exists
	info, err := os.Stat(expanded)
	if err != nil {
		return ExecutionResult{
			Output:   fmt.Sprintf("cd: %s: %v", target, err),
			ExitCode: 1,
			Error:    err,
		}
	}

	if !info.IsDir() {
		return ExecutionResult{
			Output:   fmt.Sprintf("cd: %s: not a directory", target),
			ExitCode: 1,
			Error:    fmt.Errorf("not a directory"),
		}
	}

	// Change directory
	if err := os.Chdir(expanded); err != nil {
		return ExecutionResult{
			Output:   fmt.Sprintf("cd: %v", err),
			ExitCode: 1,
			Error:    err,
		}
	}

	b.state.WorkingDirectory = expanded

	return ExecutionResult{
		Output:   "",
		ExitCode: 0,
		Error:    nil,
	}
}

func (b *BuiltinHandler) exit(args []string) ExecutionResult {
	b.state.ShouldExit = true
	b.state.ExitCode = 0

	if len(args) > 0 {
		// Try to parse exit code
		var code int
		_, err := fmt.Sscanf(args[0], "%d", &code)
		if err == nil {
			b.state.ExitCode = code
		}
	}

	return ExecutionResult{
		Output:   "",
		ExitCode: 0, // Always return success for exit command itself
		Error:    nil,
	}
}

func (b *BuiltinHandler) help(args []string) ExecutionResult {
	if len(args) == 0 {
		// General help
		return ExecutionResult{
			Output: "gosh - Go Shell with yaegi interpreter\n\n" +
				"COMMANDS:\n" +
				"  cd [DIR]          Change directory to DIR (or home if no DIR)\n" +
				"  exit [CODE]        Exit shell with optional exit code\n" +
				"  help [COMMAND]    Show help for COMMAND, or this general help\n" +
				"  init               Initialize ~/.config/gosh with shellapi config\n\n" +
				"CONFIGURATION:\n" +
				"  config.go          Go configuration file executed on startup\n" +
				"    - Checked in current directory first\n" +
				"    - Falls back to ~/.config/gosh/config.go\n" +
				"    - Full Go syntax with IDE support (LSP, treesitter)\n" +
				"    - Define functions, set environment, import packages\n" +
				"    - Functions persist and are available in the shell\n\n" +
				"GO CODE:\n" +
				"  Write Go code directly:\n" +
				"    x := 42\n" +
				"    fmt.Println(x)\n" +
				"    func add(a, b int) int { return a + b }\n\n" +
				"  Pre-imported packages: fmt, os, strings, strconv, path/filepath\n\n" +
				"  Multiline code supported with continuation prompts (...)\n\n" +
				"COMMAND SUBSTITUTION:\n" +
				"  $(command) captures command output into a Go string:\n" +
				"    files := $(ls)\n" +
				"    result := $(curl -s https://example.com)\n\n" +
				"SHELLAPI (v0.2.1+):\n" +
				"  Advanced shell functions via manual wrapper system:\n" +
				"    gs()              # Git status with colors\n" +
				"    ok('message')     # Green success message\n" +
				"    warn('message')   # Yellow warning message\n" +
				"    err('message')    # Red error message\n" +
				"    build()           # Go build project\n" +
				"    shellapi.GitStatus()  # Direct access also works\n" +
				"    shellapi.Success('text') # Direct access with colors\n" +
				"  Try 'help shellapi' for more information\n\n" +
				"ROUTING:\n" +
				"  - Built-in commands are executed first\n" +
				"  - Go syntax (assignments, functions, loops) is evaluated with yaegi\n" +
				"  - Everything else is executed as shell commands\n\n" +
				"For more information, visit: https://github.com/rsarv3006/gosh",
			ExitCode: 0,
			Error:    nil,
		}
	}

	// Help for specific commands
	command := args[0]

	if command == "cd" {
		return ExecutionResult{
			Output: "cd - Change Directory\n\n" +
				"USAGE:\n" +
				"    cd [DIRECTORY]\n\n" +
				"DESCRIPTION:\n" +
				"    Change the current working directory to DIRECTORY.\n" +
				"    If no DIRECTORY is specified, change to the user's home directory.\n\n" +
				"EXAMPLES:\n" +
				"    cd                    # Change to home directory\n" +
				"    cd ~/projects        # Change to projects directory\n" +
				"    cd /usr/local        # Change to absolute path\n" +
				"    cd ..               # Change to parent directory",
			ExitCode: 0, Error: nil,
		}
	}

	if command == "exit" {
		return ExecutionResult{
			Output: "exit - Exit Shell\n\n" +
				"USAGE:\n" +
				"    exit [EXIT_CODE]\n\n" +
				"DESCRIPTION:\n" +
				"    Exit the shell with an optional exit code.\n\n" +
				"EXAMPLES:\n" +
				"    exit          # Exit with code 0\n" +
				"    exit 1        # Exit with code 1 (error)\n" +
				"    exit 127      # Exit with code 127 (command not found)",
			ExitCode: 0, Error: nil,
		}
	}

	if command == "help" {
		return ExecutionResult{
			Output: "help - Show Help\n\n" +
				"USAGE:\n" +
				"    help [COMMAND]\n\n" +
				"DESCRIPTION:\n" +
				"    Show help information for COMMAND, or general help if no COMMAND specified.\n\n" +
				"EXAMPLES:\n" +
				"    help          # Show this general help\n" +
				"    help cd       # Show help for cd command\n" +
				"    help init     # Show help for init command\n" +
				"    help shellapi # Show help for shellapi functions\n" +
				"    help go       # Show help for Go code execution",
			ExitCode: 0, Error: nil,
		}
	}

	// Check for init help
	if command == "init" {
		return ExecutionResult{
			Output: "init - Initialize gosh Configuration\n\n" +
				"USAGE:\n" +
				"    init\n\n" +
				"DESCRIPTION:\n" +
				"    Initialize ~/.config/gosh directory with shellapi configuration.\n" +
				"    Creates go.mod file and template config.go with manual wrapper examples.\n\n" +
				"CREATES:\n" +
				"    ~/.config/gosh/                      - Configuration directory\n" +
				"    ~/.config/gosh/go.mod                 - Go module file\n" +
				"    ~/.config/gosh/config.go              - Template config with examples\n\n" +
				"TEMPLATE INCLUDES:\n" +
				"    • shellapi import for advanced functions\n" +
				"    • Working shellapi examples (build, test, run, gs)\n" +
				"    • Directory navigation functions (goGosh, goHome, goConfig)\n" +
				"    • Real command execution with error handling\n" +
				"    • Color output functions (ok, warn, err)\n" +
				"    • Persistent directory changes\n\n" +
				"AFTER INIT:\n" +
				"    1. Restart gosh to load the new configuration\n" +
				"    2. Try: build()       # Execute real go build\n" +
				"    3. Try: test()        # Execute real go test\n" +
				"    4. Try: goGosh()      # Navigate to project directory\n" +
				"    5. Try: gs()          # Real git status with colors\n" +
				"    6. Try: ok('Done!')  # Green success message\n\n" +
				"NOTE:\n" +
				"    The config provides shellapi functions that execute real commands via Go's os/exec.\n" +
				"    Directory changes persist across shell sessions. Functions work reliably!",
			ExitCode: 0, Error: nil,
		}
	}

	// Check for shellapi help
	if command == "shellapi" {
		return ExecutionResult{
			Output: "shellapi - Shell Function Library (v0.2.1+)\n\n" +
				"OVERVIEW:\n" +
				"    shellapi provides 100+ shell-friendly functions organized\n" +
				"    into categories: development tools, file operations, git,\n" +
				"    system commands, colors, and project utilities.\n\n" +
				"MANUAL WRAPPER PATTERN:\n" +
				"    Instead of direct access, create manual wrapper functions:\n\n" +
				"EXAMPLE WRAPPER CONFIG:\n" +
				"    import \"github.com/rsarv3006/gosh_lib/shellapi\"\n\n" +
				"    func gs() string {\n" +
				"        result, _ := shellapi.GitStatus()\n" +
				"        return result  // Command substitution processed\n" +
				"    }\n\n" +
				"    func ok(msg string) string {\n" +
				"        return shellapi.Success(msg)\n" +
				"    }\n\n" +
				"DUAL ACCESS:\n" +
				"    • Manual wrappers: gs(), ok(), build(), warn(), err()\n" +
				"    • Direct access: shellapi.GitStatus(), shellapi.Success()\n" +
				"    • Both patterns process command substitutions automatically\n\n" +
				"AVAILABLE CATEGORIES:\n" +
				"    🔧 Development: GoBuild(), GoTest(), NpmInstall(), DockerPs()\n" +
				"    📁 File Ops:    Ls(), Cat(), Find(), Grep(), Touch()\n" +
				"    🔀 Git:         GitStatus(), GitLog(), QuickCommit(), GitPull()\n" +
				"    🖥️  System:      Uptime(), Date(), Pwd(), EnvVar()\n" +
				"    🎨 Colors:      Success(), Error(), Warning(), Bold()\n" +
				"    🏗️  Project:     MakeTarget(), BuildAndTest(), CreateProjectDir()\n\n" +
				"COLOR EXAMPLES:\n" +
				"    shellapi.Success(\"Build passed!\")   # Green text\n" +
				"    shellapi.Warning(\"Caution\")        # Yellow text\n" +
				"    shellapi.Error(\"Failed!\")          # Red text\n\n" +
				"SETUP:\n" +
				"    1. Run 'init' to create config with examples\n" +
				"    2. Or manually create ~/.config/gosh/config.go\n" +
				"    3. Import shellapi and define your wrappers\n\n" +
				"For more information: https://github.com/rsarv3006/gosh_lib",
			ExitCode: 0, Error: nil,
		}
	}

	// Check for config help
	if command == "config" || command == "config.go" {
		return ExecutionResult{
			Output: "Configuration - config.go\n\n" +
				"USAGE:\n" +
				"    Create a config.go file in current directory or ~/.config/gosh/\n\n" +
				"DESCRIPTION:\n" +
				"    config.go is a regular Go file executed when gosh starts.\n" +
				"    It provides full Go syntax with IDE support (LSP, treesitter, autocomplete).\n" +
				"    Functions and variables defined in config.go persist and are available\n" +
				"    throughout the shell session.\n\n" +
				"FILE LOCATIONS:\n" +
				"    1. ./config.go                    (current directory, takes precedence)\n" +
				"    2. ~/.config/gosh/config.go      (home directory, fallback)\n\n" +
				"EXAMPLE config.go:\n" +
				"    package main\n\n" +
				"    import (\n" +
				"        \"fmt\"\n" +
				"        \"os\"\n" +
				"    )\n\n" +
				"    // Runs on shell startup\n" +
				"    func init() {\n" +
				"        fmt.Println(\"Loading custom config...\")\n" +
				"        os.Setenv(\"EDITOR\", \"vim\")\n" +
				"    }\n\n" +
				"    // Available throughout the shell session\n" +
				"    func hello(name string) {\n" +
				"        fmt.Printf(\"Hello %s!\\n\", name)\n" +
				"    }\n\n" +
				"    // Custom prompt example (when implemented)\n" +
				"    func CustomPrompt() string {\n" +
				"        return fmt.Sprintf(\"gosh[%s]$ \", \n" +
				"            strings.TrimPrefix(os.Getenv(\"PWD\"), os.Getenv(\"HOME\")))\n" +
				"    }\n\n" +
				"FEATURES:\n" +
				"    • Full Go syntax support\n" +
				"    • IDE editing with LSP and syntax highlighting\n" +
				"    • Pre-imported packages available (fmt, os, strings, etc.)\n" +
				"    • Additional imports handled automatically\n" +
				"    • Functions persist in shell REPL\n" +
				"    • Environment variables set during startup\n\n" +
				"NOTES:\n" +
				"    • Common packages (fmt, os, strings, etc.) are already imported\n" +
				"    • Additional imports are stripped from config.go before evaluation\n" +
				"    • Use init() for startup configuration",
			ExitCode: 0, Error: nil,
		}
	}

	// Help for session builtin
	if command == "session" {
		return ExecutionResult{
			Output: "session - Open or print the current REPL session file\n\n" +
				"USAGE:\n" +
				"    session          # Open session file in $EDITOR or system opener\n" +
				"    session --print  # Print path to session file\n\n" +
				"DESCRIPTION:\n" +
				"    The session command opens a temporary Go file that mirrors the REPL\n" +
				"    state. Function definitions are placed at package level and other\n" +
				"    executable statements are placed inside func session(). This file\n" +
				"    is updated after each executed Go input so editors show the current\n" +
				"    REPL state.\n",
			ExitCode: 0, Error: nil,
		}
	}

	// Check for Go features
	if command == "substitution" || command == "command" || command == "go" || command == "golang" || command == "yaegi" {
		if command == "substitution" || command == "command" {
			return ExecutionResult{
				Output: "Command Substitution\n\n" +
					"SYNTAX:\n" +
					"    $(command)\n\n" +
					"DESCRIPTION:\n" +
					"    Execute SHELL command and capture its output as a Go string literal.\n" +
					"    This enables seamless integration between shell commands and Go code.\n\n" +
					"EXAMPLES:\n" +
					"    files := $(ls)                    # Capture ls output\n" +
					"    result := $(curl -s api.example) # Capture curl response\n" +
					"    user := $(whoami)                 # Capture user name\n\n" +
					"NOTES:\n" +
					"    - Command output is automatically escaped for Go string literals\n" +
					"    - Works in assignments, function calls, anywhere Go expects a string\n" +
					"    - Shell commands are executed with the shell's PATH and environment",
				ExitCode: 0, Error: nil,
			}
		}

		if command == "go" || command == "golang" || command == "yaegi" {
			return ExecutionResult{
				Output: "Go Code Execution\n\n" +
					"DESCRIPTION:\n" +
					"    Write and execute Go code directly in the shell with full language support.\n\n" +
					"FEATURES:\n" +
					"    • Persistent variables and functions\n" +
					"    • Pre-imported packages: fmt, os, strings, strconv, path/filepath\n" +
					"    • Multiline support with continuation prompts\n" +
					"    • Full Go language features (except CGo, limited generics)\n\n" +
					"EXAMPLES:\n" +
					"    x := 42\n" +
					"    fmt.Println(x*2)\n\n" +
					"    func add(a, b int) int { return a + b }\n" +
					"    fmt.Println(add(5, 3))\n\n" +
					"    for i := 0; i < 3; i++ {\n" +
					"        fmt.Println(\"iteration\", i)\n" +
					"    }\n\n" +
					"NOTES:\n" +
					"    • Code is executed by yaegi interpreter\n" +
					"    • State persists across commands\n" +
					"    • No compilation required",
				ExitCode: 0, Error: nil,
			}
		}
	}

	// Check if it's a shell command
	if path, found := FindInPath(command, b.state.Environment["PATH"]); found {
		return ExecutionResult{
			Output: fmt.Sprintf("%s - External Command\n\n"+
				"This is an external command. Use \"man %s\" for detailed documentation,\n"+
				"or run it with --help or -h for usage information.\n\n"+
				"Location: %s", command, command, path),
			ExitCode: 0, Error: nil,
		}
	}

	return ExecutionResult{
		Output:   fmt.Sprintf("No help available for '%s'", command),
		ExitCode: 1,
		Error:    fmt.Errorf("no help available"),
	}
}

// session prints or opens the LSP session file used by the REPL
func (b *BuiltinHandler) session(args []string) ExecutionResult {
	// Parse flags: --print/-p and cleanup
	printOnly := false
	cleanupOnly := false
	for _, a := range args {
		if a == "--print" || a == "-p" {
			printOnly = true
		}
		if a == "cleanup" || a == "--cleanup" {
			cleanupOnly = true
		}
	}

	sessionPath := b.state.SessionFilePath
	if sessionPath == "" {
		// If LSP not initialized, create a fallback temp file path
		tmpDir := os.TempDir()
		sessionPath = filepath.Join(tmpDir, "gosh-session.go")
	}

	if printOnly {
		return ExecutionResult{Output: sessionPath, ExitCode: 0, Error: nil}
	}

	if cleanupOnly {
		// Perform cleanup of old session/workspace temp dirs. Keep 5 most recent.
		if err := CleanOldSessionDirs(os.TempDir(), 0, 5); err != nil {
			return ExecutionResult{Output: fmt.Sprintf("Cleanup failed: %v", err), ExitCode: 1, Error: err}
		}
		return ExecutionResult{Output: "Old session directories cleaned (kept 5 most recent)", ExitCode: 0, Error: nil}
	}

	// Ensure the session file exists on disk so system openers / editors can open it.
	// gopls uses a virtual file path inside a temp dir and may never create a physical
	// file. Creating a minimal session file avoids `open` failing with exit status 1.
	dir := filepath.Dir(sessionPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return ExecutionResult{Output: fmt.Sprintf("Failed to create session dir: %v", err), ExitCode: 1, Error: err}
	}
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		initial := "package main\n\nimport \"fmt\"\n\nfunc session() {\n}\n"
		if err := os.WriteFile(sessionPath, []byte(initial), 0644); err != nil {
			return ExecutionResult{Output: fmt.Sprintf("Failed to create session file: %v", err), ExitCode: 1, Error: err}
		}
	}

	// Try to open with user's EDITOR
	editor := strings.TrimSpace(b.state.Environment["EDITOR"])
	if editor != "" {
		parts := strings.Fields(editor)
		cmd := exec.Command(parts[0], append(parts[1:], sessionPath)...)
		cmd.Env = b.state.EnvironmentSlice()
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return ExecutionResult{Output: fmt.Sprintf("Failed to open editor %s: %v", editor, err), ExitCode: 1, Error: err}
		}
		return ExecutionResult{Output: "", ExitCode: 0, Error: nil}
	}

	// Fallback: try to use system opener (macOS: open, Linux: xdg-open)
	if _, err := exec.LookPath("open"); err == nil {
		cmd := exec.Command("open", sessionPath)
		cmd.Env = b.state.EnvironmentSlice()
		if err := cmd.Run(); err != nil {
			return ExecutionResult{Output: fmt.Sprintf("Failed to open session file with open: %v", err), ExitCode: 1, Error: err}
		}
		return ExecutionResult{Output: "", ExitCode: 0, Error: nil}
	}

	if _, err := exec.LookPath("xdg-open"); err == nil {
		cmd := exec.Command("xdg-open", sessionPath)
		cmd.Env = b.state.EnvironmentSlice()
		if err := cmd.Run(); err != nil {
			return ExecutionResult{Output: fmt.Sprintf("Failed to open session file with xdg-open: %v", err), ExitCode: 1, Error: err}
		}
		return ExecutionResult{Output: "", ExitCode: 0, Error: nil}
	}

	// As a last resort, just print the path
	return ExecutionResult{Output: sessionPath, ExitCode: 0, Error: nil}
}

// initConfig creates the .config/gosh directory with go.mod and template config.go
func (b *BuiltinHandler) initConfig(args []string) ExecutionResult {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return ExecutionResult{
			Output:   "Cannot determine home directory",
			ExitCode: 1,
			Error:    fmt.Errorf("HOME environment variable not set"),
		}
	}

	configDir := filepath.Join(homeDir, ".config", "gosh")

	// Create .config/gosh directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return ExecutionResult{
			Output:   fmt.Sprintf("Failed to create config directory: %v", err),
			ExitCode: 1,
			Error:    err,
		}
	}

	// Create go.mod if it doesn't exist
	goModPath := filepath.Join(configDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		// Use the published package with v0.1.0
		goModContent := `module user-config

go 1.21

// Import gosh_lib for rich shell functions  
require github.com/rsarv3006/gosh_lib v0.2.0
`
		if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
			return ExecutionResult{
				Output:   fmt.Sprintf("Failed to create go.mod: %v", err),
				ExitCode: 1,
				Error:    err,
			}
		}
		debugf("Created %s\n", goModPath)
	} else {
		debugf("go.mod already exists at %s\n", goModPath)
	}

	// Create config.go if it doesn't exist
	configPath := filepath.Join(configDir, "config.go")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configContent := `package main

import (
	"fmt"
	"github.com/rsarv3006/gosh_lib/shellapi"
)

func init() {
	fmt.Println("🚀 gosh config loaded! Command execution system enabled!")
}

// ==============================================================================
// WORKING SHELLAPI FUNCTIONS (v0.2.2+)
// ==============================================================================
// These functions execute real commands via Go's os/exec and return results.

// Development helper functions that actually work
func build() string {
	result, err := shellapi.GoBuild()
	if err != nil {
		return "BUILD ERROR: " + err.Error()
	}
	return "BUILD SUCCESS: " + result  // Executes real go build command
}

func test() string {
	result, _ := shellapi.GoTest()
	return result  // Executes real go test with full output
}

func run() string {
	result, _ := shellapi.GoRun()
	return result  // Executes real go run . application
}

func gs() string {
	result, err := shellapi.GitStatus()
	if err != nil {
		return "GIT ERROR: " + err.Error()
	}
	return "GIT STATUS:\n" + result  // Executes real git status with colors
}

// ==============================================================================
// DIRECTORY NAVIGATION FUNCTIONS
// ==============================================================================
// These functions change directories and persist the change across shell sessions.

// goGosh() navigates to the gosh development directory
//
// IMPORTANT: This function MUST return the shellapi.RunShell result as-is for CD to work!
// The shellapi.RunShell("cd", path) returns a special marker: @GOSH_INTERNAL_CD:/path
// When this marker is returned, the gosh evaluator detects it and actually changes the shell's
// working directory. If you try to format or modify the result, the CD functionality breaks!
//
// ✅ DO:     return result  // Return CD marker unmodified for directory change
// ❌ DON'T:  return "Changed to " + result  // This breaks the CD marker system
// ❌ DON'T:  return fmt.Sprintf("CD: %s", result)  // This breaks the CD marker system
//
// After calling goGosh(), the directory change persists for the entire shell session.
// This enables creating navigation functions like goHome(), goProjects(), etc.
func goGosh() string {
	result, _ := shellapi.RunShell("cd", "/Users/rjs/dev/gosh")
	return result // 🚨 CRITICAL: Return CD marker unmodified for directory change to work!
}

// Navigate to home config directory
func goConfig() string {
	result, _ := shellapi.RunShell("cd", "/Users/rjs/.config/gosh/")
	return result  // Navigate to config directory
}

// Navigate to home directory  
func goHome() string {
	result, _ := shellapi.RunShell("cd", "~")
	return result  // Navigate to home directory
}

// ==============================================================================
// UTILITY FUNCTIONS
// ==============================================================================

// Simple welcome function
func hello() string {
	return "Hello from gosh!"
}

// Success message with green color
func ok(msg string) string {
	return shellapi.Success(msg)  // Green colored text
}

// Warning message with yellow color  
func warn(msg string) string {
	return shellapi.Warning(msg)  // Yellow colored text
}

// Error message with red color
func err(msg string) string {
	return shellapi.Error(msg)  // Red colored text
}
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			return ExecutionResult{
				Output:   fmt.Sprintf("Failed to create config.go: %v", err),
				ExitCode: 1,
				Error:    err,
			}
		}
		debugf("Created %s\n", configPath)
	} else {
		debugf("config.go already exists at %s\n", configPath)
	}

	// Note: Skip go mod tidy for now since v0.1.0 checksum isn't published yet
	debugln("📝 Config files created successfully!")
	debugln("💡 Run 'cd ~/.config/gosh && go mod tidy' manually if needed")
	return ExecutionResult{
		Output:   fmt.Sprintf("✅ gosh config directory initialized at %s", configDir),
		ExitCode: 0,
		Error:    nil,
	}
}
