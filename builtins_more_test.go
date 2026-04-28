//go:build darwin || linux

package main

import (
	"testing"
)

func TestBuiltinPwd(t *testing.T) {
	state := NewShellState()
	builtins := NewBuiltinHandler(state)

	result := builtins.pwd(nil)

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if result.Output == "" {
		t.Error("Expected non-empty output for pwd")
	}

	// Output should be the working directory
	if result.Output != state.WorkingDirectory {
		t.Errorf("Expected %q, got %q", state.WorkingDirectory, result.Output)
	}
}

func TestBuiltinSession(t *testing.T) {
	state := NewShellState()
	builtins := NewBuiltinHandler(state)

	// Test without args - this will try to open editor and fail in test environment
	result := builtins.session(nil)

	// In test environment, editor likely fails, so just check it returns something
	if result.Output == "" && result.Error == nil {
		t.Error("Expected either output or error for session command")
	}
}

func TestBuiltinIsBuiltin(t *testing.T) {
	state := NewShellState()
	builtins := NewBuiltinHandler(state)

	tests := []struct {
		command  string
		expected bool
	}{
		{"cd", true},
		{"exit", true},
		{"help", true},
		{"init", true},
		{"pwd", true},
		{"session", true},
		{"ls", false},
		{"echo", false},
		{"git", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := builtins.IsBuiltin(tt.command)
			if result != tt.expected {
				t.Errorf("IsBuiltin(%q) = %v, want %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestBuiltinExecute(t *testing.T) {
	state := NewShellState()
	builtins := NewBuiltinHandler(state)

	tests := []struct {
		name     string
		command  string
		args     []string
		expected int
	}{
		{
			name:     "cd without args",
			command:  "cd",
			args:     nil,
			expected: 0,
		},
		{
			name:     "exit",
			command:  "exit",
			args:     nil,
			expected: 0,
		},
		{
			name:     "help",
			command:  "help",
			args:     nil,
			expected: 0,
		},
		{
			name:     "pwd",
			command:  "pwd",
			args:     nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builtins.Execute(tt.command, tt.args)
			if result.ExitCode != tt.expected {
				t.Errorf("Execute(%q) exit code = %d, want %d", tt.command, result.ExitCode, tt.expected)
			}
		})
	}
}

func TestBuiltinCdErrors(t *testing.T) {
	state := NewShellState()
	builtins := NewBuiltinHandler(state)

	// Test cd to non-existent directory
	result := builtins.cd([]string{"/nonexistent/path"})

	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code for cd to non-existent path")
	}

	if result.Error == nil {
		t.Error("Expected error for cd to non-existent path")
	}
}
