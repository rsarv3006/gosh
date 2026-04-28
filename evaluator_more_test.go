//go:build darwin || linux

package main

import (
	"strings"
	"testing"
)

func TestEvalSimple(t *testing.T) {
	eval := NewGoEvaluator()

	// Test simple expression
	result := eval.Eval("42")

	if result.Error != nil {
		t.Errorf("Eval failed: %v", result.Error)
	}

	if result.Output != "42" {
		t.Errorf("Expected '42', got %q", result.Output)
	}
}

func TestEvalString(t *testing.T) {
	eval := NewGoEvaluator()

	result := eval.Eval(`"hello"`)

	if result.Error != nil {
		t.Errorf("Eval failed: %v", result.Error)
	}

	if !strings.Contains(result.Output, "hello") {
		t.Errorf("Expected 'hello' in output, got %q", result.Output)
	}
}

func TestEvalVariable(t *testing.T) {
	eval := NewGoEvaluator()

	// Define a variable
	eval.Eval("x := 42")

	// Use it
	result := eval.Eval("x")

	if result.Error != nil {
		t.Errorf("Eval failed: %v", result.Error)
	}

	if result.Output != "42" {
		t.Errorf("Expected '42', got %q", result.Output)
	}
}

func TestEvalWithRecovery(t *testing.T) {
	eval := NewGoEvaluator()

	// Test that recovery works for invalid code
	result := eval.EvalWithRecovery("this is invalid code")

	if result.Error == nil {
		t.Error("Expected error for invalid code")
	}
}

func TestEvalCommandSubstitution(t *testing.T) {
	state := NewShellState()
	eval := NewGoEvaluator()
	spawner := NewProcessSpawner(state)
	eval.SetupWithShell(state, spawner)

	// Test command substitution
	result := eval.Eval("$(echo hello)")

	if result.Error != nil {
		t.Errorf("Eval failed: %v", result.Error)
	}

	if !strings.Contains(result.Output, "hello") {
		t.Errorf("Expected 'hello' in output, got %q", result.Output)
	}
}

func TestLoadConfig(t *testing.T) {
	eval := NewGoEvaluator()

	// Test loading non-existent config (should not error)
	err := eval.LoadConfig()

	// LoadConfig doesn't return error for missing files
	if err != nil {
		t.Errorf("LoadConfig should not error for missing config: %v", err)
	}
}

func TestProcessCommandSubstitutions(t *testing.T) {
	state := NewShellState()
	eval := NewGoEvaluator()
	spawner := NewProcessSpawner(state)
	eval.SetupWithShell(state, spawner)

	// Test processing command substitutions
	code := "msg := $(echo hello world)"
	processed := eval.processCommandSubstitutions(code)

	if strings.Contains(processed, "$(") {
		t.Errorf("Command substitution not processed: %s", processed)
	}

	if !strings.Contains(processed, "hello world") {
		t.Errorf("Expected 'hello world' in processed code: %s", processed)
	}
}

func TestStripImports(t *testing.T) {
	eval := NewGoEvaluator()

	// Test stripping shellapi imports
	code := `import "fmt"
import "github.com/rsarv3006/gosh_lib/shellapi"
func main() {
	fmt.Println("hello")
}`

	stripped := eval.stripImports(code)

	// Should strip shellapi import
	if strings.Contains(stripped, "shellapi") {
		t.Errorf("shellapi import not stripped: %s", stripped)
	}

	// Should keep other imports
	if !strings.Contains(stripped, "fmt") {
		t.Errorf("fmt import was removed: %s", stripped)
	}

	if !strings.Contains(stripped, "fmt.Println") {
		t.Errorf("Code content missing after strip: %s", stripped)
	}
}
