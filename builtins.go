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
	case "cd", "exit", "pwd":
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
