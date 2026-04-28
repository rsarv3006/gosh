//go:build darwin || linux

package main

import (
	"testing"
)

func TestIsComplete(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "simple statement",
			input:    "fmt.Println(\"hello\")",
			expected: true,
		},
		{
			name:     "unclosed brace",
			input:    "if x > 0 {",
			expected: false,
		},
		{
			name:     "closed brace",
			input:    "if x > 0 { fmt.Println(x) }",
			expected: true,
		},
		{
			name:     "unclosed parenthesis",
			input:    "fmt.Println(",
			expected: false,
		},
		{
			name:     "closed parenthesis",
			input:    "fmt.Println(\"hello\")",
			expected: true,
		},
		{
			name:     "unclosed bracket",
			input:    "arr[",
			expected: false,
		},
		{
			name:     "closed bracket",
			input:    "arr[0]",
			expected: true,
		},
		{
			name:     "ends with comma",
			input:    "1, 2, 3,",
			expected: false,
		},
		{
			name:     "ends with plus",
			input:    "x +",
			expected: false,
		},
		{
			name:     "multiline function",
			input:    "func add(a int, b int) int {\n    return a + b\n}",
			expected: true,
		},
		{
			name:     "incomplete function",
			input:    "func add(a int, b int) int {",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isComplete(tt.input)
			if result != tt.expected {
				t.Errorf("isComplete(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLooksLikePathCompletion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "ends with slash",
			input:    "cd /usr/local/",
			expected: true,
		},
		{
			name:     "path with tilde",
			input:    "~/Documents/",
			expected: true,
		},
		{
			name:     "no path",
			input:    "fmt.Println",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "just slash",
			input:    "/",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := looksLikePathCompletion(tt.input)
			if result != tt.expected {
				t.Errorf("looksLikePathCompletion(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBlockModeConstants(t *testing.T) {
	if ModeShell != 0 {
		t.Errorf("ModeShell should be 0, got %v", ModeShell)
	}
	if ModeGo != 1 {
		t.Errorf("ModeGo should be 1, got %v", ModeGo)
	}
}
