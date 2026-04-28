//go:build darwin || linux

package main

import (
	"strings"
	"testing"
)

func TestGoEvaluator_InjectVariable(t *testing.T) {
	eval := NewGoEvaluator()

	// Inject a simple string variable
	err := eval.InjectVariable("testVar", "hello")
	if err != nil {
		t.Errorf("Failed to inject variable: %v", err)
	}

	// Verify it works by evaluating it
	result := eval.Eval("testVar")
	if result.Error != nil {
		t.Errorf("Failed to use injected variable: %v", result.Error)
	}

	if result.Output != "hello" {
		t.Errorf("Expected 'hello', got %q", result.Output)
	}
}

func TestGoEvaluator_InjectVariableSlice(t *testing.T) {
	eval := NewGoEvaluator()

	// Inject a slice variable (like captured shell output)
	lines := []string{"line1", "line2", "line3"}
	err := eval.InjectVariable("capturedOutput", lines)
	if err != nil {
		t.Errorf("Failed to inject slice variable: %v", err)
	}

	// Verify by using len()
	result := eval.Eval("len(capturedOutput)")
	if result.Error != nil {
		t.Errorf("Failed to use injected slice: %v", result.Error)
	}

	if result.Output != "3" {
		t.Errorf("Expected '3', got %q", result.Output)
	}
}

func TestProcessCommandSubstitution(t *testing.T) {
	state := NewShellState()
	eval := NewGoEvaluator()
	spawner := NewProcessSpawner(state)
	eval.SetupWithShell(state, spawner)

	// Test command substitution processing
	code := "files := $(echo hello world)"
	processed := eval.processCommandSubstitutions(code)

	// The $(echo hello world) should be replaced with a string literal
	if strings.Contains(processed, "$(") {
		t.Errorf("Command substitution not processed: %s", processed)
	}

	// Should contain the output as a string literal
	if !strings.Contains(processed, "hello world") {
		t.Errorf("Expected 'hello world' in processed code: %s", processed)
	}
}

func TestProcessCommandSubstitutionForDisplay(t *testing.T) {
	state := NewShellState()
	eval := NewGoEvaluator()
	spawner := NewProcessSpawner(state)
	eval.SetupWithShell(state, spawner)

	// Test command substitution for display
	code := "$(echo hello)"
	result := eval.processCommandSubstitutionsForDisplay(code)

	// Should return the raw output
	if !strings.Contains(result, "hello") {
		t.Errorf("Expected 'hello' in result: %s", result)
	}
}

func TestFormatResult(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "string result",
			input:    `"hello"`,
			expected: "hello",
		},
		{
			name:     "integer result",
			input:    "42",
			expected: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eval := NewGoEvaluator()
			result := eval.Eval(tt.input)
			if result.Error != nil {
				t.Errorf("Eval failed: %v", result.Error)
			}
			// Note: formatResult is not exported, so we test via Eval
		})
	}
}

func TestEvaluatorMin(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{0, 0, 0},
		{-1, 1, -1},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := evaluatorMin(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("evaluatorMin(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestInputTypeConstants(t *testing.T) {
	if InputTypeGo != 0 {
		t.Errorf("InputTypeGo should be 0")
	}
	if InputTypeCommand != 1 {
		t.Errorf("InputTypeCommand should be 1")
	}
	if InputTypeBuiltin != 2 {
		t.Errorf("InputTypeBuiltin should be 2")
	}
	if InputTypeModeSwitch != 3 {
		t.Errorf("InputTypeModeSwitch should be 3")
	}
}

func TestExecutionResult(t *testing.T) {
	result := ExecutionResult{
		Output:   "test output",
		ExitCode: 1,
		Error:    nil,
	}

	if result.Output != "test output" {
		t.Errorf("Expected 'test output', got %q", result.Output)
	}
	if result.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", result.ExitCode)
	}
}
