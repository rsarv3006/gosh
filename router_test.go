//go:build darwin || linux

package main

import (
	"testing"
)

func TestRouter_Route_Builtins(t *testing.T) {
	builtins := NewBuiltinHandler(NewShellState())
	router := NewRouter(builtins)

	tests := []struct {
		name     string
		input    string
		expected InputType
		command  string
		args     []string
	}{
		{
			name:     "Route cd builtin",
			input:    "cd",
			expected: InputTypeBuiltin,
			command:  "cd",
			args:     []string{},
		},
		{
			name:     "Route cd with argument",
			input:    "cd /tmp",
			expected: InputTypeBuiltin,
			command:  "cd",
			args:     []string{"/tmp"},
		},
		{
			name:     "Route pwd builtin",
			input:    "pwd",
			expected: InputTypeBuiltin,
			command:  "pwd",
			args:     []string{},
		},
		{
			name:     "Route pwd with flag",
			input:    "pwd -L",
			expected: InputTypeBuiltin,
			command:  "pwd",
			args:     []string{"-L"},
		},
		{
			name:     "Route exit builtin",
			input:    "exit",
			expected: InputTypeBuiltin,
			command:  "exit",
			args:     []string{},
		},
		{
			name:     "Route exit with code",
			input:    "exit 1",
			expected: InputTypeBuiltin,
			command:  "exit",
			args:     []string{"1"},
		},
		{
			name:     "Route help builtin",
			input:    "help",
			expected: InputTypeBuiltin,
			command:  "help",
			args:     []string{},
		},
		{
			name:     "Route help with topic",
			input:    "help cd",
			expected: InputTypeBuiltin,
			command:  "help",
			args:     []string{"cd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputType, command, args := router.Route(tt.input)
			
			if inputType != tt.expected {
				t.Errorf("expected inputType %v, got %v", tt.expected, inputType)
			}
			
			if command != tt.command {
				t.Errorf("expected command '%s', got '%s'", tt.command, command)
			}
			
			if len(args) != len(tt.args) {
				t.Errorf("expected %d args, got %d", len(tt.args), len(args))
				return
			}
			
			for i, expectedArg := range tt.args {
				if args[i] != expectedArg {
					t.Errorf("expected arg %d to be '%s', got '%s'", i, expectedArg, args[i])
				}
			}
		})
	}
}

func TestRouter_Route_GoCode(t *testing.T) {
	builtins := NewBuiltinHandler(NewShellState())
	router := NewRouter(builtins)

	tests := []struct {
		name     string
		input    string
		expected InputType
	}{
		{
			name:     "Variable declaration",
			input:    "x := 42",
			expected: InputTypeGo,
		},
		{
			name:     "Function declaration",
			input:    "func add(a, b int) int { return a + b }",
			expected: InputTypeGo,
		},
		{
			name:     "For loop",
			input:    "for i := 0; i < 3; i++ { fmt.Println(i) }",
			expected: InputTypeGo,
		},
		{
			name:     "If statement",
			input:    "if x > 0 { fmt.Println(\"positive\") }",
			expected: InputTypeGo,
		},
		{
			name:     "Function call",
			input:    "fmt.Println(\"hello\")",
			expected: InputTypeGo,
		},
		{
			name:     "Assignment",
			input:    "x = 100",
			expected: InputTypeGo,
		},
		{
			name:     "Import statement",
			input:    "import \"time\"",
			expected: InputTypeGo,
		},
		{
			name:     "Type declaration",
			input:    "type Person struct { Name string }",
			expected: InputTypeGo,
		},
		{
			name:     "Const declaration",
			input:    "const Pi = 3.14",
			expected: InputTypeGo,
		},
		{
			name:     "Select statement",
			input:    "select { case <-time.After(time.Second): break }",
			expected: InputTypeGo,
		},
		{
			name:     "Switch statement",
			input:    "switch x { case 1: fmt.Println(\"one\") }",
			expected: InputTypeGo,
		},
		{
			name:     "Go statement",
			input:    "go func() { fmt.Println(\"hello\") }()",
			expected: InputTypeGo,
		},
		{
			name:     "Defer statement",
			input:    "defer fmt.Println(\"goodbye\")",
			expected: InputTypeGo,
		},
		{
			name:     "Return statement",
			input:    "return x + y",
			expected: InputTypeGo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputType, _, _ := router.Route(tt.input)
			
			if inputType != tt.expected {
				t.Errorf("expected inputType %v, got %v", tt.expected, inputType)
			}
		})
	}
}

func TestRouter_Route_CommandSubstitution(t *testing.T) {
	builtins := NewBuiltinHandler(NewShellState())
	router := NewRouter(builtins)

	tests := []struct {
		name     string
		input    string
		expected InputType
	}{
		{
			name:     "Simple command substitution",
			input:    "files := $(ls)",
			expected: InputTypeGo,
		},
		{
			name:     "Command substitution in function call",
			input:    "fmt.Println($(whoami))",
			expected: InputTypeGo,
		},
		{
			name:     "Nested command substitution",
			input:    "result := $(echo \"$(date)\")",
			expected: InputTypeGo,
		},
		{
			name:     "Command substitution with quotes",
			input:    "user := $(grep \"admin\" /etc/passwd)",
			expected: InputTypeGo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputType, _, _ := router.Route(tt.input)
			
			if inputType != tt.expected {
				t.Errorf("expected inputType %v, got %v", tt.expected, inputType)
			}
		})
	}
}

func TestRouter_Route_ShellCommand(t *testing.T) {
	builtins := NewBuiltinHandler(NewShellState())
	router := NewRouter(builtins)

	tests := []struct {
		name     string
		input    string
		expected InputType
		command  string
		args     []string
	}{
		{
			name:     "Simple ls command",
			input:    "ls",
			expected: InputTypeCommand,
			command:  "ls",
			args:     []string{},
		},
		{
			name:     "ls with arguments",
			input:    "ls -la",
			expected: InputTypeCommand,
			command:  "ls",
			args:     []string{"-la"},
		},
		{
			name:     "git command",
			input:    "git status",
			expected: InputTypeCommand,
			command:  "git",
			args:     []string{"status"},
		},
		{
			name:     "Complex command with paths",
			input:    "git add /path/to/file.txt",
			expected: InputTypeCommand,
			command:  "git",
			args:     []string{"add", "/path/to/file.txt"},
		},
		{
			name:     "Single command with flags",
			input:    "docker run -it ubuntu bash",
			expected: InputTypeCommand,
			command:  "docker",
			args:     []string{"run", "-it", "ubuntu", "bash"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputType, command, args := router.Route(tt.input)
			
			if inputType != tt.expected {
				t.Errorf("expected inputType %v, got %v", tt.expected, inputType)
			}
			
			if command != tt.command {
				t.Errorf("expected command '%s', got '%s'", tt.command, command)
			}
			
			if len(args) != len(tt.args) {
				t.Errorf("expected %d args, got %d", len(tt.args), len(args))
				return
			}
			
			for i, expectedArg := range tt.args {
				if args[i] != expectedArg {
					t.Errorf("expected arg %d to be '%s', got '%s'", i, expectedArg, args[i])
				}
			}
		})
	}
}

func TestRouter_Route_EdgeCases(t *testing.T) {
	builtins := NewBuiltinHandler(NewShellState())
	router := NewRouter(builtins)

	tests := []struct {
		name     string
		input    string
		expected InputType
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: InputTypeCommand, // Default to command
		},
		{
			name:     "Whitespace only",
			input:    "   ",
			expected: InputTypeCommand, // Default to command
		},
		{
			name:     "Single word that's not builtin or Go keyword",
			input:    "unknowncommand",
			expected: InputTypeCommand,
		},
		{
			name:     "Command that looks like Go",
			input:    "echo package main",
			expected: InputTypeCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputType, _, _ := router.Route(tt.input)
			
			if inputType != tt.expected {
				t.Errorf("expected inputType %v, got %v", tt.expected, inputType)
			}
		})
	}
}
