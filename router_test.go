//go:build darwin || linux

package main

import (
	"testing"
)

// Test that builtins are properly routed in shell mode
func TestRouter_BuiltinRouting(t *testing.T) {
	state := NewShellState()
	builtins := NewBuiltinHandler(state)
	router := NewRouter(builtins, state)

	tests := []struct {
		name     string
		input    string
		expected InputType
	}{
		{
			name:     "cd builtin",
			input:    "cd /tmp",
			expected: InputTypeBuiltin,
		},
		{
			name:     "exit builtin",
			input:    "exit",
			expected: InputTypeBuiltin,
		},
		{
			name:     "pwd builtin",
			input:    "pwd",
			expected: InputTypeBuiltin,
		},
		{
			name:     "help builtin",
			input:    "help",
			expected: InputTypeBuiltin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputType, _, _ := router.Route(tt.input)
			if inputType != tt.expected {
				t.Errorf("Route(%q) = %v, want %v", tt.input, inputType, tt.expected)
			}
		})
	}
}

// Test shell command routing
func TestRouter_ShellCommandRouting(t *testing.T) {
	state := NewShellState()
	builtins := NewBuiltinHandler(state)
	router := NewRouter(builtins, state)

	tests := []struct {
		name     string
		input    string
		expected InputType
	}{
		{
			name:     "ls command",
			input:    "ls -la",
			expected: InputTypeCommand,
		},
		{
			name:     "git command",
			input:    "git status",
			expected: InputTypeCommand,
		},
		{
			name:     "echo command",
			input:    "echo hello",
			expected: InputTypeCommand,
		},
		{
			name:     "cat command",
			input:    "cat file.txt",
			expected: InputTypeCommand,
		},
		{
			name:     "go build command (shell mode)",
			input:    "go build .",
			expected: InputTypeCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputType, _, _ := router.Route(tt.input)
			if inputType != tt.expected {
				t.Errorf("Route(%q) = %v, want %v", tt.input, inputType, tt.expected)
			}
		})
	}
}

// Test edge cases
func TestRouter_EdgeCaseRouting(t *testing.T) {
	state := NewShellState()
	builtins := NewBuiltinHandler(state)
	router := NewRouter(builtins, state)

	tests := []struct {
		name     string
		input    string
		expected InputType
	}{
		{
			name:     "empty input",
			input:    "",
			expected: InputTypeCommand, // Default fallback
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: InputTypeCommand, // Default fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputType, _, _ := router.Route(tt.input)
			if inputType != tt.expected {
				t.Errorf("Route(%q) = %v, want %v", tt.input, inputType, tt.expected)
			}
		})
	}
}

// Note: In the new architecture, Go code routing is not tested here because
// mode is explicit (:go/:sh commands). Go code is sent directly to the
// evaluator when in Go mode, not routed through this router.
