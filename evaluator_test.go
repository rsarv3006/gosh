//go:build darwin || linux

package main

import (
	"testing"
	"os"
	"strings"
)

func TestNewGoEvaluator_Creation(t *testing.T) {
	eval := NewGoEvaluator()
	
	if eval == nil {
		t.Fatal("NewGoEvaluator returned nil")
	}
	
	if eval.interp == nil {
		t.Fatal("GoEvaluator should have interpreter")
	}
	
	// Should have standard library symbols loaded
	if eval.originalOut != os.Stdout {
		t.Error("originalOut should point to os.Stdout")
	}
	
	if eval.originalErr != os.Stderr {
		t.Error("originalErr should point to os.Stderr")
	}
}

func TestGoEvaluator_Eval_SimpleAssignment(t *testing.T) {
	eval := NewGoEvaluator()

	// Test simple variable assignment
	result := eval.Eval("x := 42")
	
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
	
	// Variable should persist in the interpreter
	result2 := eval.Eval("x")
	expected := "42"
	if result2.Output != expected {
		t.Errorf("Expected %q, got %q", expected, result2.Output)
	}
}

func TestGoEvaluator_Eval_Arithmetic(t *testing.T) {
	eval := NewGoEvaluator()

	// Test arithmetic operations
	result := eval.Eval("result := 5 + 3")
	
	if result.Output != "8" {
		t.Errorf("Expected 8, got %q", result.Output)
	}
	
	// Test another arithmetic operation
	result = eval.Eval("y := result * 2")
	
	if result.Output != "16" {
		t.Errorf("Expected 16, got %q", result.Output)
	}
}

func TestGoEvaluator_Eval_FunctionDeclaration(t *testing.T) {
	eval := NewGoEvaluator()

	// Should handle multiline function declarations
	code := `func add(a, b int) int { 
		return a + b 
	}`
	
	result := eval.Eval(code)
	
	if result.Error != nil {
		t.Errorf("Expected no error for function declaration, got: %v", result.Error)
	}
	
	// Function should be available for use
	result2 := eval.Eval("add(3, 4)")
	if result2.Output != "7" {
		t.Errorf("Expected 7, got %q", result2.Output)
	}
}

func TestGoEvaluator_Eval_MultilineCode(t *testing.T) {
	eval := NewGoEvaluator()

	// Test Go code with control structures
	code := `for i := 0; i < 3; i++ { 
		fmt.Println(i) 
	}`
	
	result := eval.Eval(code)
	
	if result.Error != nil {
		t.Errorf("Expected no error for for loop, got: %v", result.Error)
	}
	
	// Should print 0, 1, 2 on separate lines
	expectedLines := []string{"0", "1", "2"}
	for _, line := range expectedLines {
		if !strings.Contains(result.Output, line) {
			t.Errorf("Output should contain %s", line)
		}
	}
}

func TestGoEvaluator_Eval_Strings(t *testing.T) {
	eval := NewGoEvaluator()

	// Test string operations
	result := eval.Eval(`name := "world"`)
	
	if result.Output != "world" {
		t.Errorf("Expected 'world', got %q", result.Output)
	}
	
	// Test string concatenation
	result = eval.Eval(`greeting := "Hello, " + name`)
	if result.Output != "Hello, world" {
		t.Errorf("Expected 'Hello, world', got %q", result.Output)
	}
}

func TestGoEvaluator_Eval_PrintStatements(t *testing.T) {
	eval := NewGoEvaluator()

	// Test that print statements don't return values
	result := eval.Eval(`fmt.Println("test")`)
	
	if result.Error != nil {
		t.Errorf("Expected no error for print statement, got: %v", result.Error)
	}
	
	// Empty output because print goes to stdout, not captured
	if result.Output != "" {
		t.Errorf("Expected empty output for print statement, got %q", result.Output)
	}
	
	// Test that Printf also doesn't return values
	result = eval.Eval(`fmt.Printf("Hello, %s!", name)`)
	// This will fail since name is not yet defined, but should not crash
	_ = result.Error
}

func TestGoEvaluator_Eval_PrintStatements_Printf(t *testing.T) {
	eval := NewGoEvaluator()

	// Test Printf with valid string literal
	result := eval.Eval(`fmt.Printf("Hello, %s!", "world")`)
	
	if result.Output != "Hello, world!" {
		t.Errorf("Expected 'Hello, world!', got %q", result.Output)
	}
	
	// Printf should return no error for valid formatting
	if result.Error != nil {
		t.Errorf("Expected no error for valid Printf, got: %v", result.Error)
	}
}

func TestGoEvaluator_Eval_Imports(t *testing.T) {
	eval := NewGoEvaluator()

	// Test pre-imported packages are available
	result := eval.Eval("length := len(\"hello\")")
	
	if result.Output != "5" {
		t.Errorf("Expected 5, got %q", result.Output)
	}
	
	// Test other pre-imported packages
	result = eval.Eval("files, _ := os.ReadDir(\".\")")
	// We expect this to work or fail gracefully - the directory may not exist
	// so we'll just check that it doesn't crash
	_ = result.Error
}

func TestGoEvaluator_Eval_CommandSubstitution(t *testing.T) {
	eval := NewGoEvaluator()

	// Test basic command substitution
	result := eval.Eval(`files := $(ls)`)
	
	if result.Error != nil {
		t.Errorf("Expected no error for command substitution, got: %v", result.Error)
	}
	
	// Should capture command output as string
	expectedPrefix := "files := "
	if !strings.HasPrefix(result.Output, expectedPrefix) {
		t.Errorf("Expected output to start with %q, got %q", expectedPrefix, result.Output)
	}
}

func TestGoEvaluator_Eval_ComplexCode(t *testing.T) {
	eval := NewGoEvaluator()

	// Test type declaration
	result := eval.Eval("type Person struct { Name string; Age int }")
	
	if result.Error != nil {
		t.Errorf("Expected no error for type declaration, got: %v", result.Error)
	}
}

func TestGoEvaluator_Eval_ErrorHandling(t *testing.T) {
	eval := NewGoEvaluator()

	// Test syntax error
	result := eval.Eval("invalid go syntax !!!")
	
	if result.Error == nil {
		t.Error("Expected error for invalid syntax")
	}
	
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code for syntax error")
	}
}

func TestGoEvaluator_Eval_EmptyInput(t *testing.T) {
	eval := NewGoEvaluator()

	// Empty input should not error
	result := eval.Eval("")
	
	// Empty input might return 0 exit code and nil error
	// Let's check it doesn't crash
	_ = result.ExitCode
	_ = result.Error
}

func TestGoEvaluator_StatuPersistence(t *testing.T) {
	eval := NewGoEvaluator()

	// Test variable persistence across multiple calls
	// First evaluation
	result1 := eval.Eval("counter := 0")
	
	if result1.ExitCode != 0 {
		t.Errorf("Expected success for first assignment")
	}
	
	// Second evaluation
	result2 := eval.Eval("counter += 1")
	
	if result2.Output != "1" {
		t.Errorf("Expected '1', got %q", result2.Output)
	}
	
	// Third evaluation
	result3 := eval.Eval("counter += 1")
	
	if result3.Output != "2" {
		t.Errorf("Expected '2', got %q", result3.Output)
	}
}


