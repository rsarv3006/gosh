//go:build darwin || linux

package main

import (
	"fmt"
	"os"
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
	case "cd", "exit", "pwd", "help":
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
	case "pwd":
		return b.pwd(args)
	case "help":
		return b.help(args)
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
		ExitCode: b.state.ExitCode,
		Error:    nil,
	}
}

func (b *BuiltinHandler) pwd(args []string) ExecutionResult {
	output := b.state.WorkingDirectory
	if len(args) > 0 && args[0] == "-L" {
		// Logical pwd (with symlinks)
		if wd, err := os.Getwd(); err == nil {
			output = wd
		}
	}

	return ExecutionResult{
		Output:   strings.TrimSpace(output),
		ExitCode: 0,
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
				"  pwd               Print current working directory\n" +
				"  exit [CODE]        Exit shell with optional exit code\n" +
				"  help [COMMAND]    Show help for COMMAND, or this general help\n\n" +
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
	
	if command == "pwd" {
		return ExecutionResult{
			Output: "pwd - Print Working Directory\n\n" +
				"USAGE:\n" +
				"    pwd [-L]\n\n" +
				"DESCRIPTION:\n" +
				"    Print the full pathname of the current working directory.\n\n" +
				"OPTIONS:\n" +
				"    -L    Print logical path (resolve symlinks)",
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
				"    help go       # Show help for Go code execution",
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
	if path, found := FindInPath(command); found {
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
