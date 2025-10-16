//go:build darwin || linux

package main

import (
	"testing"
)

// Test that Go keywords are properly routed to InputTypeGo
// This test would have caught the routing bug where "func" was executed as a shell command
func TestRouter_GoKeywordRouting(t *testing.T) {
	state := NewShellState()
	builtins := NewBuiltinHandler(state)
	router := NewRouter(builtins, state)

	tests := []struct {
		name     string
		input    string
		expected InputType
	}{
		{
			name:     "function definition",
			input:    "func add(a int, b int) int { return a + b }",
			expected: InputTypeGo,
		},
		{
			name:     "variable declaration",
			input:    "x := 42",
			expected: InputTypeGo,
		},
		{
			name:     "const declaration",
			input:    "const pi = 3.14159",
			expected: InputTypeGo,
		},
		{
			name:     "type definition",
			input:    "type Person struct { Name string }",
			expected: InputTypeGo,
		},
		{
			name:     "if statement",
			input:    "if x > 0 { fmt.Println(\"positive\") }",
			expected: InputTypeGo,
		},
		{
			name:     "for loop",
			input:    "for i := 0; i < 10; i++ { fmt.Println(i) }",
			expected: InputTypeGo,
		},
		{
			name:     "switch statement",
			input:    "switch x { case 1: fmt.Println(\"one\") }",
			expected: InputTypeGo,
		},
		{
			name:     "import statement",
			input:    "import \"time\"",
			expected: InputTypeGo,
		},
		{
			name:     "go statement",
			input:    "go fmt.Println(\"hello\")",
			expected: InputTypeGo,
		},
		{
			name:     "defer statement",
			input:    "defer fmt.Println(\"done\")",
			expected: InputTypeGo,
		},
		{
			name:     "return statement",
			input:    "return 42",
			expected: InputTypeGo,
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

// Test builtins are properly routed
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

// Test command substitution routing
func TestRouter_CommandSubstitutionRouting(t *testing.T) {
	state := NewShellState()
	builtins := NewBuiltinHandler(state)
	router := NewRouter(builtins, state)

	tests := []struct {
		name     string
		input    string
		expected InputType
	}{
		{
			name:     "simple command substitution",
			input:    "files := $(ls)",
			expected: InputTypeGo,
		},
		{
			name:     "command substitution with args",
			input:    "output := $(echo hello world)",
			expected: InputTypeGo,
		},
		{
			name:     "command substitution in expression",
			input:    "fmt.Println($(date))",
			expected: InputTypeGo,
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

// Test shell command routing for actual commands
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

// Test edge cases that could cause routing confusion
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
		{
			name:     "func as shell command (should be Go code)",
			input:    "func add(a int, b int) int { return a + b }",
			expected: InputTypeGo, // Critical test - this was the bug!
		},
		{
			name:     "go as shell command keyword",
			input:    "go build .",
			expected: InputTypeCommand, // 'go' as command, not Go keyword
		},
		{
			name:     "go as Go statement",
			input:    "go func() { fmt.Println(\"async\") }()",
			expected: InputTypeGo, // 'go' as Go keyword
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

// Test that the regression is fixed - specifically the bug where 'func' was executed as a shell command
func TestRouter_FuncKeywordRegression(t *testing.T) {
	state := NewShellState()
	builtins := NewBuiltinHandler(state)
	router := NewRouter(builtins, state)

	// This is the exact input that caused the bug
	input := "func add(a int, b int) int {\nreturn a + b\n}"
	inputType, command, args := router.Route(input)

	// Should be routed as Go code, not as shell command
	if inputType != InputTypeGo {
		t.Errorf("Regression: func definition routed as %v instead of InputTypeGo", inputType)
	}

	// Command should be the original input since it's Go code
	if command != input {
		t.Errorf("Expected command to be original input for Go code, got %q", command)
	}

	// Args should be nil since it's Go code
	if args != nil {
		t.Errorf("Expected nil args for Go code, got %v", args)
	}
}
