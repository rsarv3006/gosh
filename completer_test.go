//go:build darwin || linux

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoshCompleter_completeCommands_InvalidInput(t *testing.T) {
	// Create a mock evaluator for testing
	evaluator := NewGoEvaluator()
	
	c := NewGoshCompleter(evaluator)

	// Empty input should return no matches
	result, length := c.Do([]rune(""), 0)
	
	if len(result) != 0 {
		t.Errorf("Expected no matches for empty input")
	}
	
	if length != 0 {
		t.Errorf("Expected length 0 for empty input, got %d", length)
	}
}

func TestGoshCompleter_completeCommands_NonexistentCommand(t *testing.T) {
	evaluator := NewGoEvaluator()
	
	c := NewGoshCompleter(evaluator)
	
	// Non-existent command should return no matches
	result, _ := c.Do([]rune("nonexistent_command"), 18)
	
	if len(result) != 0 {
		t.Errorf("Expected no matches for non-existent command")
	}
}

func TestGoshCompleter_completeCommands_ExactMatch(t *testing.T) {
	evaluator := NewGoEvaluator()
	
	c := NewGoshCompleter(evaluator)
	
	// Exact command should return empty suffix
	result, _ := c.Do([]rune("pwd"), 3)
	
	if len(result) < 1 {
		t.Errorf("Expected at least 1 match for exact match, got %d", len(result))
	}
	
	// Should find the exact 'pwd' command with empty suffix
	found := false
	for _, match := range result {
		if string(match) == "" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected empty suffix for exact match among %d results: %v", len(result), result)
	}
}

func TestGoshCompleter_completeCommands_WithMultiple(t *testing.T) {
	evaluator := NewGoEvaluator()
	
	c := NewGoshCompleter(evaluator)
	
	// Partial that matches multiple commands
	result, _ := c.Do([]rune("c"), 1)
	
	// Should return multiple matches for 'c' (cd, cat, const, plus any PATH executables starting with 'c')
	expectedMatches := 3 // Minimum expected (cd, cat, const), will be more with PATH executables
	if len(result) < expectedMatches {
		t.Errorf("Expected at least %d matches for 'c', got %d", expectedMatches, len(result))
	}
}

func TestGoshCompleter_Do_CommandCompletion(t *testing.T) {
	evaluator := NewGoEvaluator()
	c := NewGoshCompleterForTesting(evaluator)

	tests := []struct {
		name           string
		line           string
		pos            int
		expectedLength int
		expectedMatch  string
	}{
		{
			name:           "Complete 'wh' to 'whoami'",
			line:           "wh",
			pos:            2,
			expectedLength: 2,
			expectedMatch:  "oami", // Should be suffix only
		},
		{
			name:           "Complete 'ls' to 'ls'",
			line:           "ls",
			pos:            2,
			expectedLength: 2,
			expectedMatch:  "", // Should be exact match, no suffix
		},
		{
			name:           "Complete 'c' (multiple matches)",
			line:           "c",
			pos:            1,
			expectedLength: 1,
			expectedMatch:  "", // Multiple matches: cd, cat, const
		},
		{
			name:           "Complete 'ex' to 'exit'",
			line:           "ex",
			pos:            2,
			expectedLength: 2,
			expectedMatch:  "it", // Should include exit builtin
		},
		{
			name:           "Complete 'x' (multiple matches)",
			line:           "x",
			pos:            1,
			expectedLength: 1,
			expectedMatch:  "", // Multiple matches, exact completion not performed
		},
		{
			name:           "Complete 'nonexistent'",
			line:           "nonexistent",
			pos:            11,
			expectedLength: 11,
			expectedMatch:  "", // No matches
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, length := c.Do([]rune(tt.line), tt.pos)
			
			if length != tt.expectedLength {
				t.Errorf("expected length %d, got %d", tt.expectedLength, length)
			}
			
			if tt.expectedMatch != "" {
				if len(matches) < 1 {
					t.Errorf("expected at least 1 match, got %d", len(matches))
					return
				}
				// Check that expected match is among the results
				found := false
				for _, match := range matches {
					if string(match) == tt.expectedMatch {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected match '%s' not found in results: %v", tt.expectedMatch, matches)
				}
			}
		})
	}
}

func TestGoshCompleter_Do_ArgumentCompletion_Help(t *testing.T) {
	evaluator := NewGoEvaluator()
	c := NewGoshCompleterForTesting(evaluator)

	tests := []struct {
		name           string
		line           string
		pos            int
		expectedMatch  string
	}{
		{
			name:          "Complete help topic 'c' to 'cd'",
			line:          "help c",
			pos:           6,
			expectedMatch: "", // Multiple matches: cd, command
		},
		{
			name:          "Complete help topic 'g'",
			line:          "help g",
			pos:           6,
			expectedMatch: "", // Multiple matches: go, golang
		},
		{
			name:          "Complete help topic 'subs' to 'substitution'",
			line:          "help subs",
			pos:           9,
			expectedMatch: "titution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, length := c.Do([]rune(tt.line), tt.pos)
			
			if tt.expectedMatch != "" {
				if len(matches) != 1 {
					t.Errorf("expected 1 match, got %d", len(matches))
					return
				}
				if string(matches[0]) != tt.expectedMatch {
					t.Errorf("expected match '%s', got '%s'", tt.expectedMatch, string(matches[0]))
				}
			} else {
				// When expectedMatch is empty, we expect multiple matches
				if len(matches) == 0 {
					t.Errorf("expected multiple matches, got %d", len(matches))
				}
			}
			
			// Verify position is within bounds to prevent panic
			if tt.pos > len(tt.line) {
				t.Errorf("test error: pos %d is longer than line length %d", tt.pos, len(tt.line))
				return
			}
			
			// Verify length matches the partial word being completed
			// Find the word being completed (from last space to cursor)
			lastSpace := strings.LastIndex(tt.line[:tt.pos], " ")
			if lastSpace == -1 {
				lastSpace = 0
			} else {
				lastSpace++ // Skip the space
			}
			partialWord := tt.line[lastSpace:tt.pos]
			expectedLength := len(partialWord)
			if length != expectedLength {
				t.Errorf("expected length %d (for partial word '%s'), got %d", expectedLength, partialWord, length)
			}
		})
	}
}

func TestGoshCompleter_Do_FileCompletion(t *testing.T) {
	evaluator := NewGoEvaluator()
	c := NewGoshCompleterForTesting(evaluator)

	// Create a temporary directory with test files
	tempDir := t.TempDir()
	
	testFiles := []string{"file1.txt", "file2.go", "directory", "test.txt"}
	for _, name := range testFiles {
		if name == "directory" {
			os.MkdirAll(filepath.Join(tempDir, name), 0755)
		} else {
			CreateTestFile(t, filepath.Join(tempDir, name), "content")
		}
	}

	// Change to temp directory for file completion tests
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	tests := []struct {
		name           string
		line           string
		pos            int
		expectMatches bool  // true if we expect some matches, false for no matches
	}{
		{
			name:          "Complete 'f' in current directory",
			line:          "ls f",
			pos:           4,
			expectMatches: true, // Should find file1.txt, file2.go
		},
		{
			name:          "Complete 'file1' in current directory",
			line:          "cat file1",
			pos:           9,
			expectMatches: true, // Should find file1.txt
		},
		{
			name:          "Complete 'dir' to 'directory/'",
			line:          "cd dir",
			pos:           6,
			expectMatches: true, // Should find directory/
		},
		{
			name:          "Complete with no matches",
			line:          "ls xyz",
			pos:           6,
			expectMatches: false, // Should find no matches
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, length := c.Do([]rune(tt.line), tt.pos)
			
			if tt.expectMatches {
				if len(matches) == 0 {
					t.Errorf("expected some matches, got %d", len(matches))
				}
			} else {
				if len(matches) > 0 {
					t.Errorf("expected no matches, got %d", len(matches))
				}
			}
			
			// Verify position is within bounds to prevent panic
			if tt.pos > len(tt.line) {
				t.Errorf("test error: pos %d is longer than line length %d", tt.pos, len(tt.line))
				return
			}
			
			// Verify length matches the partial word being completed
			// Find the word being completed (from last space to cursor)
			lastSpace := strings.LastIndex(tt.line[:tt.pos], " ")
			if lastSpace == -1 {
				lastSpace = 0
			} else {
				lastSpace++ // Skip the space
			}
			partialWord := tt.line[lastSpace:tt.pos]
			expectedLength := len(partialWord)
			if length != expectedLength {
				t.Errorf("expected length %d (for partial word '%s'), got %d", expectedLength, partialWord, length)
			}
		})
	}
}

func TestGoshCompleter_completeCommands(t *testing.T) {
	evaluator := NewGoEvaluator()
	c := NewGoshCompleterForTesting(evaluator)

	tests := []struct {
		partial         string
		expectedBuiltins []string // Should find these builtin completions
		shouldFindAtLeast int     // Minimum number of results expected
	}{
		{
			partial:         "wh",
			expectedBuiltins: []string{}, // No builtins start with "wh"
			shouldFindAtLeast: 1,        // But should find whoami or other PATH commands
		},
		{
			partial:         "cd", 
			expectedBuiltins: []string{""}, // exact match with builtin
			shouldFindAtLeast: 1,            
		},
		{
			partial:         "c",
			expectedBuiltins: []string{"d", "at"}, // cd, cat builtins
			shouldFindAtLeast: 2,                // At least cd, cat (const is Go keyword, not command)
		},
		{
			partial:         "e",
			expectedBuiltins: []string{"xit"},    // exit builtin
			shouldFindAtLeast: 1,                 // At least exit
		},
		{
			partial:         "nonexistent",
			expectedBuiltins: []string{},        // Should find nothing
			shouldFindAtLeast: 0,                
		},
	}

	for _, tt := range tests {
		t.Run("completeCommands_"+tt.partial, func(t *testing.T) {
			result := c.completeCommands(tt.partial)
			
			if len(result) < tt.shouldFindAtLeast {
				t.Errorf("expected at least %d matches for '%s', got %d", tt.shouldFindAtLeast, tt.partial, len(result))
				return
			}
			
			// Convert results to strings for easier comparison
			resultStrings := make([]string, len(result))
			for i, match := range result {
				resultStrings[i] = string(match)
			}
			
			// Check that expected builtins are present
			for _, expectedBuiltin := range tt.expectedBuiltins {
				found := false
				for _, actualResult := range resultStrings {
					if actualResult == expectedBuiltin {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected builtin suffix '%s' not found in results: %v", expectedBuiltin, resultStrings)
				}
			}
		})
	}
}

func TestGoshCompleter_completeFiles(t *testing.T) {
	evaluator := NewGoEvaluator()
	c := NewGoshCompleterForTesting(evaluator)

	// Create a temporary directory with test files
	tempDir := t.TempDir()
	
	testFiles := []string{"file1.txt", "file2.go", "directory", "test.txt"}
	for _, name := range testFiles {
		if name == "directory" {
			os.MkdirAll(filepath.Join(tempDir, name), 0755)
		} else {
			CreateTestFile(t, filepath.Join(tempDir, name), "content")
		}
	}

	// Change to temp directory for file completion tests  
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	tests := []struct {
		partial  string
		dirsOnly bool
		expected []string
	}{
		{
			partial:  "file",
			dirsOnly: false,
			expected: []string{"1.txt", "2.go"}, // suffixes for file1.txt, file2.go
		},
		{
			partial:  "dir",
			dirsOnly: true,
			expected: []string{"ectory/"}, // directory/ - should include slash
		},
		{
			partial:  "test",
			dirsOnly: false,
			expected: []string{".txt"}, // .txt for test.txt
		},
		{
			partial:  "nonexistent",
			dirsOnly: false,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run("completeFiles_"+tt.partial, func(t *testing.T) {
			result := c.completeFiles(tt.partial, tt.dirsOnly)
			
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d matches, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, match := range result {
				expected := tt.expected[i]
				if string(match) != expected {
					t.Errorf("expected match '%s', got '%s'", expected, string(match))
				}
			}
		})
	}
}

// Test local executable completion (regression test for local executables not completing)
func TestGoshCompleter_completeCommands_LocalExecutables(t *testing.T) {
	evaluator := NewGoEvaluator()
	completer := NewGoshCompleterForTesting(evaluator)
	
	// Create a temporary directory with a test executable
	tempDir := t.TempDir()
	
	// Save current directory and change to temp dir
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	os.Chdir(tempDir)
	
	// Create test executable files
	testExecs := []string{"mytool", "test-app", "local-script"}
	for _, name := range testExecs {
		createTestExecutable(t, tempDir, name)
	}
	
	tests := []struct {
		name     string
		partial  string
		expected []string // Should find these completions
	}{
		{
			name:     "complete mytool",
			partial:  "mytoo",
			expected: []string{"l"},
		},
		{
			name:     "complete test-app", 
			partial:  "test-",
			expected: []string{"app"},
		},
		{
			name:     "complete from scratch",
			partial:  "",
			expected: []string{}, // Empty partial should return empty list
		},
		{
			name:     "no match",
			partial:  "xyz",
			expected: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := completer.completeCommands(tt.partial)
			
			// Convert matches to strings for comparison
			matchStrings := make([]string, len(matches))
			for i, match := range matches {
				matchStrings[i] = string(match)
			}
			
			// Check if all expected results are present
			for _, expectedSuffix := range tt.expected {
				found := false
				for _, actualSuffix := range matchStrings {
					if actualSuffix == expectedSuffix {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected completion suffix %q not found in %v", expectedSuffix, matchStrings)
				}
			}
		})
	}
}

// createTestExecutable creates an executable file for testing
func createTestExecutable(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	content := "#!/bin/bash\necho 'test executable'\n"
	err := os.WriteFile(path, []byte(content), 0755)
	if err != nil {
		t.Fatalf("Failed to create test executable %s: %v", name, err)
	}
}

// TestLocalExecutableCompletionWithDotSlash tests the ./prefix completion behavior
func TestLocalExecutableCompletionWithDotSlash(t *testing.T) {
	evaluator := NewGoEvaluator()
	completer := NewGoshCompleterForTesting(evaluator)
	
	// Create a temporary directory with test executables
	tempDir := t.TempDir()
	
	// Save current directory and change to temp dir
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	os.Chdir(tempDir)
	
	// Create test executables
	createTestExecutable(t, tempDir, "gosh")
	createTestExecutable(t, tempDir, "go-build")
	createTestExecutable(t, tempDir, "go.test")
	
	tests := []struct {
		name     string
		input    string
		expectedSuffixes []string // Should find these completions
	}{
		{
			name:     "./go to ./gosh",
			input:    "./go",
			expectedSuffixes: []string{"sh"}, // gosh -> go + sh
		},
		{
			name:     "./go to multiple",
			input:    "./go",
			expectedSuffixes: []string{"sh", "-build", ".test"}, // gosh, go-build, go.test
		},
		{
			name:     "./go exact match doesn't exist",
			input:    "./go",
			expectedSuffixes: []string{"sh", "-build", ".test"}, // Should find gosh, go-build, go.test
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, length := completer.Do([]rune(tt.input), len(tt.input))
			
			if length != len(tt.input) {
				t.Errorf("expected length %d, got %d", len(tt.input), length)
			}
			
			// Check that expected suffixes are present
			resultStrings := make([]string, len(matches))
			for i, match := range matches {
				resultStrings[i] = string(match)
			}
			
			for _, expectedSuffix := range tt.expectedSuffixes {
				found := false
				for _, actualSuffix := range resultStrings {
					if actualSuffix == expectedSuffix {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected completion suffix '%s' not found in results: %v", expectedSuffix, resultStrings)
				}
			}
		})
	}
}

// Helper function to create test files
func CreateTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", path, err)
	}
}
